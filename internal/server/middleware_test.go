package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/auth"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/database"
	"github.com/fan/safe-mysql-mcp/internal/mcp"
	"github.com/fan/safe-mysql-mcp/internal/metrics"
	"github.com/fan/safe-mysql-mcp/internal/security"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

const testSecret = "middleware-test-secret-at-least-32-chars"

func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	t.Setenv("JWT_SECRET", testSecret)

	securityRules := &config.SecurityRules{
		AllowedDML:   []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		AllowedDDL:   []string{"CREATE_TABLE"},
		Blocked:      []string{"DROP"},
		QueryTimeout: 30e9,
		MaxRows:      10000,
	}

	reloadCfg := &config.ReloadableConfig{}
	reloadCfg.Update(&config.Config{
		Server:   config.ServerConfig{JWTSecret: testSecret},
		Clusters: config.ClustersConfig{"primary": {Host: "localhost", Port: 3306}},
		Audit:    config.AuditConfig{Enabled: false},
	}, &config.SecurityConfig{Security: *securityRules})

	validator := auth.NewValidator(testSecret)
	rateLimiter := NewIPRateLimiter(DefaultRateLimiterConfig())
	m := metrics.Init("test")

	pool, _ := database.NewPool(config.ClustersConfig{})
	router := database.NewRouter(pool, config.DatabasesConfig{})
	parser := security.NewParser()
	checker := security.NewChecker(securityRules)
	rewriter := security.NewRewriter(securityRules)
	auditLogger, _ := audit.NewLogger(&config.AuditConfig{Enabled: false})

	handler := mcp.NewHandler(router, parser, checker, rewriter, auditLogger, reloadCfg)
	mcpServer := mcp.NewMCPServer()

	srv := &Server{
		cfg:         reloadCfg,
		validator:   validator,
		handler:     handler,
		audit:       auditLogger,
		pool:        pool,
		mcpServer:   mcpServer,
		rateLimiter: rateLimiter,
		metrics:     m,
	}

	token, err := validator.GenerateToken("test-user", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	return srv, token
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	srv, token := newTestServer(t)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		userID := auth.GetUserID(r.Context())
		if userID != "test-user" {
			t.Errorf("userID = %q, want %q", userID, "test-user")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := srv.authMiddleware(next)

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	srv, _ := newTestServer(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called without token")
	})

	handler := srv.authMiddleware(next)

	req := httptest.NewRequest("POST", "/mcp", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	srv, _ := newTestServer(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next should not be called with invalid token")
	})

	handler := srv.authMiddleware(next)

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	srv, _ := newTestServer(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := srv.rateLimitMiddleware(srv.rateLimiter, next)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
}

func TestHandleHealth(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	srv.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "OK")
	}
}

func TestShutdown_NoHTTPServer(t *testing.T) {
	srv, _ := newTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestHandler_RegistersTools(t *testing.T) {
	srv, _ := newTestServer(t)

	// Handler() should register tools via sync.Once and return a valid mux
	handler := srv.Handler()
	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Calling Handler() again should not panic (sync.Once protects registration)
	handler2 := srv.Handler()
	if handler2 == nil {
		t.Fatal("second Handler() call returned nil")
	}
}
