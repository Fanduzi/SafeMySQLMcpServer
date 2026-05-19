//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestHealthCheck verifies the /health endpoint returns 200 without auth.
func TestHealthCheck(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	resp, err := http.Get(env.ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("body = %q, want %q", body, "OK")
	}
}

// TestMCP_NoAuth verifies that /mcp rejects requests without a JWT token.
func TestMCP_NoAuth(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	// Use MCP client WITHOUT auth transport
	client := http.Client{}
	resp, err := client.Post(env.ts.URL+"/mcp", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

// TestMCP_InvalidToken verifies that /mcp rejects requests with an invalid JWT.
func TestMCP_InvalidToken(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	req, _ := http.NewRequest("POST", env.ts.URL+"/mcp", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

// TestMCP_ListDatabases verifies list_databases returns the configured database.
func TestMCP_ListDatabases(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      "list_databases",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool list_databases: %v", err)
	}

	if result.IsError {
		t.Fatalf("list_databases returned error: %v", result.Content)
	}

	found := false
	for _, content := range result.Content {
		if text, ok := content.(*mcpsdk.TextContent); ok {
			if strings.Contains(text.Text, envOr("MYSQL_DATABASE", "testdb")) {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("list_databases result doesn't contain test database: %v", result.Content)
	}
}

// TestMCP_Query_Select verifies a SELECT query through the full chain.
func TestMCP_Query_Select(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	// Ensure a test table exists
	setupTestTable(t, dbName)

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "SELECT 1 AS value",
		},
	})
	if err != nil {
		t.Fatalf("CallTool query: %v", err)
	}

	if result.IsError {
		t.Fatalf("query returned error: %v", result.Content)
	}
}

// TestMCP_ListTables verifies list_tables returns tables from INFORMATION_SCHEMA.
func TestMCP_ListTables(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")
	setupTestTable(t, dbName)

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "list_tables",
		Arguments: map[string]any{
			"database": dbName,
		},
	})
	if err != nil {
		t.Fatalf("CallTool list_tables: %v", err)
	}

	if result.IsError {
		t.Fatalf("list_tables returned error: %v", result.Content)
	}
}

// TestMCP_DescribeTable verifies describe_table returns column info.
func TestMCP_DescribeTable(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")
	setupTestTable(t, dbName)

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "describe_table",
		Arguments: map[string]any{
			"database": dbName,
			"table":    "e2e_test",
		},
	})
	if err != nil {
		t.Fatalf("CallTool describe_table: %v", err)
	}

	if result.IsError {
		t.Fatalf("describe_table returned error: %v", result.Content)
	}
}

// TestMCP_Query_SecurityBlock verifies DROP TABLE is blocked by security rules.
func TestMCP_Query_SecurityBlock(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "query",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "DROP TABLE e2e_test",
		},
	})
	if err != nil {
		t.Fatalf("CallTool query: %v", err)
	}

	if !result.IsError {
		t.Fatal("DROP TABLE should be blocked by security rules")
	}

	blocked := false
	for _, content := range result.Content {
		if text, ok := content.(*mcpsdk.TextContent); ok {
			if strings.Contains(strings.ToLower(text.Text), "blocked") {
				blocked = true
			}
		}
	}
	if !blocked {
		t.Errorf("error message should mention 'blocked', got: %v", result.Content)
	}
}

// TestMCP_Explain verifies EXPLAIN works for SELECT queries.
func TestMCP_Explain(t *testing.T) {
	env := setupE2E(t)
	defer env.cleanup()

	dbName := envOr("MYSQL_DATABASE", "testdb")
	setupTestTable(t, dbName)

	result, err := env.session.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: "explain",
		Arguments: map[string]any{
			"database": dbName,
			"sql":      "SELECT * FROM e2e_test",
		},
	})
	if err != nil {
		t.Fatalf("CallTool explain: %v", err)
	}

	if result.IsError {
		t.Fatalf("explain returned error: %v", result.Content)
	}
}

// setupTestTable creates a test table if it doesn't exist.
func setupTestTable(t *testing.T, dbName string) {
	t.Helper()

	db, err := sql.Open("mysql", envDSN())
	if err != nil {
		t.Fatalf("connect MySQL: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("USE `%s`", dbName))
	if err != nil {
		t.Fatalf("USE %s: %v", dbName, err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS e2e_test (
			id   INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE e2e_test: %v", err)
	}
}
