// Package server handles HTTP server setup
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
		pool.Close()
		return nil, fmt.Errorf("create audit logger: %w", err)
	}

	// Create MCP handler
	handler := mcp.NewHandler(router, parser, checker, rewriter, auditLogger, cfg)

	// Create MCP SDK server
	mcpServer := mcp.NewMCPServer()

	// Create rate limiter
	rateLimiter := NewIPRateLimiter(DefaultRateLimiterConfig())

	return &Server{
		cfg:         cfg,
		validator:   validator,
		handler:     handler,
		audit:       auditLogger,
		pool:        pool,
		mcpServer:   mcpServer,
		rateLimiter: rateLimiter,
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
	mux.Handle("/mcp", rateLimitedMCP)

	// Health check endpoint with rate limiting (no auth required)
	healthHandler := http.HandlerFunc(s.handleHealth)
	rateLimitedHealth := s.rateLimitMiddleware(s.rateLimiter, healthHandler)
	mux.Handle("/health", rateLimitedHealth)

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
	w.Write([]byte("OK"))
}

// authMiddleware validates JWT tokens and adds user info to context
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		token := auth.ExtractToken(authHeader)

		if token == "" {
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := s.validator.Validate(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

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
