// Package server handles HTTP server setup and routing.
// input: config.ReloadableConfig, HTTP requests
// output: HTTP responses, MCP protocol handling
// pos: HTTP layer, wires middleware (auth, rate limit, metrics) to handlers
// note: if this file changes, update header and internal/server/README.md
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/database"
	"github.com/fan/safe-mysql-mcp/internal/mcp"
	"github.com/fan/safe-mysql-mcp/internal/metrics"
	"github.com/fan/safe-mysql-mcp/internal/security"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server represents the HTTP server
type Server struct {
	cfg         *config.ReloadableConfig
	validator   *auth.Validator
	handler     *mcp.Handler
	audit       *audit.Logger
	pool        *database.Pool
	httpSrv     *http.Server
	mcpServer   *mcpsdk.Server
	rateLimiter *IPRateLimiter
	metrics     *metrics.Metrics
}

// New creates a new server
func New(cfg *config.ReloadableConfig) (*Server, error) {
	config := cfg.Get()

	// Create JWT validator (supports environment variable)
	validator, err := auth.NewValidatorFromEnv(config.Server.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("create JWT validator: %w", err)
	}

	// Create connection pool
	pool, err := database.NewPool(config.Clusters)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Create router
	router := database.NewRouter(pool, config.Databases)

	// Create security components
	parser := security.NewParser()
	securityRules := cfg.GetSecurity()
	checker := security.NewChecker(securityRules)
	rewriter := security.NewRewriter(securityRules)

	// Create audit logger
	auditLogger, err := audit.NewLogger(&config.Audit)
	if err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("create audit logger: %w", err)
	}

	// Create MCP handler
	handler := mcp.NewHandler(router, parser, checker, rewriter, auditLogger, cfg)

	// Create MCP SDK server
	mcpServer := mcp.NewMCPServer()

	// Create rate limiter
	rateLimiter := NewIPRateLimiter(DefaultRateLimiterConfig())

	// Initialize metrics
	m := metrics.Init("safemysql")

	return &Server{
		cfg:         cfg,
		validator:   validator,
		handler:     handler,
		audit:       auditLogger,
		pool:        pool,
		mcpServer:   mcpServer,
		rateLimiter: rateLimiter,
		metrics:     m,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	cfg := s.cfg.Get()

	// Register tools with the MCP server
	mcp.RegisterTools(s.mcpServer, s.handler)

	// Create MCP HTTP handler using SDK
	mcpHandler := mcpsdk.NewStreamableHTTPHandler(func(r *http.Request) *mcpsdk.Server {
		return s.mcpServer
	}, nil)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register MCP endpoint with rate limiting and auth middleware
	authHandler := s.authMiddleware(mcpHandler)
	rateLimitedMCP := s.rateLimitMiddleware(s.rateLimiter, authHandler)
	metricsMCP := s.metricsMiddleware("/mcp", rateLimitedMCP)
	mux.Handle("/mcp", metricsMCP)

	// Health check endpoint with rate limiting (no auth required)
	healthHandler := http.HandlerFunc(s.handleHealth)
	rateLimitedHealth := s.rateLimitMiddleware(s.rateLimiter, healthHandler)
	metricsHealth := s.metricsMiddleware("/health", rateLimitedHealth)
	mux.Handle("/health", metricsHealth)

	// Metrics endpoint (no auth required, for Prometheus scraping)
	mux.Handle("/metrics", s.metrics.Handler())

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Starting server on %s", addr)
	return s.httpSrv.ListenAndServe()
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// metricsMiddleware records HTTP request metrics
func (s *Server) metricsMiddleware(path string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track active requests
		s.metrics.RequestsActive.Inc()
		defer s.metrics.RequestsActive.Dec()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)
		s.metrics.RequestsTotal.WithLabelValues(r.Method, path, fmt.Sprintf("%d", wrapped.status)).Inc()
		s.metrics.RequestDuration.WithLabelValues(r.Method, path).Observe(duration.Seconds())
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// authMiddleware validates JWT tokens and adds user info to context
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		token := auth.ExtractToken(authHeader)

		if token == "" {
			s.metrics.RecordAuthAttempt("jwt", false)
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := s.validator.Validate(token)
		if err != nil {
			s.metrics.RecordAuthAttempt("jwt", false)
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Record successful auth
		s.metrics.RecordAuthAttempt("jwt", true)

		// Add user info to context
		ctx := auth.ContextWithUser(r.Context(), claims.UserID, claims.UserEmail)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	var errs []error

	if s.httpSrv != nil {
		if err := s.httpSrv.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if s.rateLimiter != nil {
		s.rateLimiter.Close()
	}

	if s.pool != nil {
		if err := s.pool.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.audit != nil {
		if err := s.audit.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// UpdateConfig updates the server configuration
func (s *Server) UpdateConfig(cfg *config.Config, security *config.SecurityConfig) {
	s.cfg.Update(cfg, security)

	// Update connection pool
	if err := s.pool.UpdateConfig(cfg.Clusters); err != nil {
		log.Printf("Failed to update pool config: %v", err)
	}

	// Update audit logger
	if err := s.audit.UpdateConfig(&cfg.Audit); err != nil {
		log.Printf("Failed to update audit config: %v", err)
	}
}
