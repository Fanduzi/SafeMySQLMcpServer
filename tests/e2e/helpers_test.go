//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/server"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

const testJWTSecret = "e2e-test-secret-key-at-least-32-chars"

// e2eEnv holds the test environment for E2E tests.
type e2eEnv struct {
	ts      *httptest.Server
	token   string
	session *mcpsdk.ClientSession
	cleanup func()
}

// setupE2E creates a full test environment: server + MCP client + JWT token.
func setupE2E(t *testing.T) *e2eEnv {
	t.Helper()

	mysqlHost := envOr("MYSQL_HOST", "127.0.0.1")
	mysqlPort := envIntOr("MYSQL_PORT", 3306)
	mysqlUser := envOr("MYSQL_USER", "root")
	mysqlPass := envOr("MYSQL_PASSWORD", "testpassword")
	mysqlDB := envOr("MYSQL_DATABASE", "testdb")

	t.Setenv("JWT_SECRET", testJWTSecret)

	cfg := &config.Config{
		Server: config.ServerConfig{
			JWTSecret: testJWTSecret,
		},
		Clusters: config.ClustersConfig{
			"primary": {
				Host:     mysqlHost,
				Port:     mysqlPort,
				Username: mysqlUser,
				Password: mysqlPass,
			},
		},
		Databases: config.DatabasesConfig{
			mysqlDB: {Cluster: "primary"},
		},
		Audit: config.AuditConfig{Enabled: false},
	}

	securityCfg := &config.SecurityConfig{
		Security: config.SecurityRules{
			AllowedDML:   []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
			AllowedDDL:   []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
			Blocked:      []string{"DROP", "TRUNCATE", "RENAME"},
			AutoLimit:    1000,
			MaxLimit:     10000,
			QueryTimeout: 30 * time.Second,
			MaxRows:      10000,
		},
	}

	reloadCfg := &config.ReloadableConfig{}
	reloadCfg.Update(cfg, securityCfg)

	srv, err := server.New(reloadCfg)
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())

	validator := auth.NewValidator(testJWTSecret)
	token, err := validator.GenerateToken("e2e-user", "e2e@test.com", time.Hour)
	if err != nil {
		ts.Close()
		t.Fatalf("generate token: %v", err)
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name:    "e2e-test-client",
		Version: "1.0.0",
	}, nil)

	httpClient := &http.Client{
		Transport: &authTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}

	session, err := client.Connect(context.Background(), &mcpsdk.StreamableClientTransport{
		Endpoint:             ts.URL + "/mcp",
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		ts.Close()
		t.Fatalf("connect MCP client: %v", err)
	}

	cleanup := func() {
		session.Close()
		ts.Close()
	}

	return &e2eEnv{
		ts:      ts,
		token:   token,
		session: session,
		cleanup: cleanup,
	}
}

// authTransport injects Authorization header into every request.
type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=Local&charset=utf8mb4",
		envOr("MYSQL_USER", "root"),
		envOr("MYSQL_PASSWORD", "testpassword"),
		envOr("MYSQL_HOST", "127.0.0.1"),
		envIntOr("MYSQL_PORT", 3306),
	)
}
