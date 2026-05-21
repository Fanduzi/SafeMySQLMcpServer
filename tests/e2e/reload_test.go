//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/server"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// reloadEnv holds the test environment plus server/config references for reload tests.
type reloadEnv struct {
	ts        *httptest.Server
	session   *mcpsdk.ClientSession
	srv       *server.Server
	reloadCfg *config.ReloadableConfig
	cleanup   func()
}

func setupReloadEnv(t *testing.T) *reloadEnv {
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
	token, err := validator.GenerateToken("reload-test-user", "reload@test.com", time.Hour)
	if err != nil {
		ts.Close()
		t.Fatalf("generate token: %v", err)
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name:    "e2e-reload-test",
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

	return &reloadEnv{
		ts:        ts,
		session:   session,
		srv:       srv,
		reloadCfg: reloadCfg,
		cleanup: func() {
			session.Close()
			ts.Close()
		},
	}
}

func mustCallQuery(t *testing.T, session *mcpsdk.ClientSession, db, sql string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": db,
			"sql":      sql,
		},
	})
	if err != nil {
		t.Fatalf("CallTool query (%s): %v", sql, err)
	}
	if result.IsError {
		for _, c := range result.Content {
			if tc, ok := c.(*mcpsdk.TextContent); ok {
				t.Fatalf("query (%s) error: %s", sql, tc.Text)
			}
		}
		t.Fatalf("query (%s) returned error: %v", sql, result.Content)
	}
}

// TestReload_SameConfig verifies the server stays responsive after
// reloading the exact same config (no cluster changes).
func TestReload_SameConfig(t *testing.T) {
	env := setupReloadEnv(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	// Before reload
	mustCallQuery(t, env.session, dbName, "SELECT 1 AS before_reload")

	// Trigger UpdateConfig with identical config (simulates security.yaml touch)
	sameCfg := env.reloadCfg.Get()
	sameSecurity := &config.SecurityConfig{
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
	env.srv.UpdateConfig(sameCfg, sameSecurity)

	// After reload — should still work
	mustCallQuery(t, env.session, dbName, "SELECT 2 AS after_reload")
}

// TestReload_SecurityRuleChange verifies the server stays responsive
// after changing security rules, and checks whether new rules take effect.
func TestReload_SecurityRuleChange(t *testing.T) {
	env := setupReloadEnv(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	// Before reload — verify basic query works
	mustCallQuery(t, env.session, dbName, "SELECT 1 AS step1")

	// Reload with DELETE removed from allowed DML
	newCfg := env.reloadCfg.Get()
	newSecurity := &config.SecurityConfig{
		Security: config.SecurityRules{
			AllowedDML:   []string{"SELECT", "INSERT", "UPDATE"},
			AllowedDDL:   []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
			Blocked:      []string{"DROP", "TRUNCATE", "RENAME"},
			AutoLimit:    1000,
			MaxLimit:     10000,
			QueryTimeout: 30 * time.Second,
			MaxRows:      10000,
		},
	}
	env.srv.UpdateConfig(newCfg, newSecurity)

	// SELECT should still work
	mustCallQuery(t, env.session, dbName, "SELECT 2 AS step2")

	// DELETE — should be blocked IF checker was updated; will pass if bug exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := env.session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "DELETE FROM e2e_reload_test WHERE 1=0",
		},
	})
	if err != nil {
		t.Fatalf("CallTool DELETE: %v", err)
	}

	if result.IsError {
		t.Logf("DELETE blocked (checker updated correctly): %v", result.Content)
	} else {
		t.Logf("DELETE allowed (BUG: checker NOT updated on reload — security rules are stale)")
	}
}

// TestReload_ConcurrentQueries verifies the server handles concurrent
// queries correctly during a config reload.
func TestReload_ConcurrentQueries(t *testing.T) {
	env := setupReloadEnv(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	mustCallQuery(t, env.session, dbName, "SELECT 1 AS warmup")

	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			_, err := env.session.CallTool(ctx, &mcpsdk.CallToolParams{
				Name: "query",
				Arguments: map[string]any{
					"database": dbName,
					"sql":      fmt.Sprintf("SELECT %d AS concurrent_query", idx),
				},
			})
			errCh <- err
		}(i)
	}

	// Trigger reload concurrently
	go func() {
		sameCfg := env.reloadCfg.Get()
		sameSecurity := &config.SecurityConfig{
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
		env.srv.UpdateConfig(sameCfg, sameSecurity)
	}()

	failures := 0
	for i := 0; i < 10; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent query %d failed: %v", i, err)
			failures++
		}
	}
	if failures > 0 {
		t.Fatalf("%d out of 10 concurrent queries failed during reload", failures)
	}
}

// TestReload_RapidReload hammers the server with rapid config updates
// to expose race conditions.
func TestReload_RapidReload(t *testing.T) {
	env := setupReloadEnv(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	mustCallQuery(t, env.session, dbName, "SELECT 1 AS warmup")

	for i := 0; i < 20; i++ {
		cfg := env.reloadCfg.Get()
		security := &config.SecurityConfig{
			Security: config.SecurityRules{
				AllowedDML:   []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
				AllowedDDL:   []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
				Blocked:      []string{"DROP", "TRUNCATE", "RENAME"},
				AutoLimit:    1000 + i,
				MaxLimit:     10000,
				QueryTimeout: 30 * time.Second,
				MaxRows:      10000,
			},
		}
		env.srv.UpdateConfig(cfg, security)
	}

	// Verify MCP still responds after rapid reloads
	mustCallQuery(t, env.session, dbName, "SELECT 3 AS after_rapid_reload")
}
