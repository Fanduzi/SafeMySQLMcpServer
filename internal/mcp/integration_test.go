//go:build integration

package mcp

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/database"
	"github.com/fan/safe-mysql-mcp/internal/security"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		fmt.Sscanf(v, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return defaultVal
}

// testEnv holds shared test infrastructure for MCP integration tests.
type testEnv struct {
	router   *database.Router
	pool     *database.Pool
	handler  *Handler
	reloadCfg *config.ReloadableConfig
	ctx      context.Context
	dbName   string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("MYSQL_HOST not set, skipping integration test")
	}

	dbName := os.Getenv("MYSQL_DATABASE")
	if dbName == "" {
		dbName = "testdb"
	}

	clusters := config.ClustersConfig{
		"primary": {
			Host:            os.Getenv("MYSQL_HOST"),
			Port:            getEnvInt("MYSQL_PORT", 3306),
			Username:        os.Getenv("MYSQL_USER"),
			Password:        os.Getenv("MYSQL_PASSWORD"),
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}

	pool, err := database.NewPool(clusters)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	databases := config.DatabasesConfig{
		dbName: {Cluster: "primary"},
	}
	router := database.NewRouter(pool, databases)

	parser := security.NewParser()
	checker := security.NewChecker(&config.SecurityRules{
		AllowedDML: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		AllowedDDL: []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
		Blocked:    []string{"DROP", "TRUNCATE"},
		AutoLimit:  1000,
		MaxLimit:   10000,
		MaxRows:    10000,
	})
	rewriter := security.NewRewriter(&config.SecurityRules{
		AutoLimit: 1000,
		MaxLimit:  10000,
	})

	auditLogger, err := audit.NewLogger(&config.AuditConfig{Enabled: false})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	reloadCfg := &config.ReloadableConfig{}
	reloadCfg.Update(&config.Config{
		Clusters:  clusters,
		Databases: databases,
	}, &config.SecurityConfig{
		Security: config.SecurityRules{
			QueryTimeout: 30 * time.Second,
			MaxRows:      10000,
		},
	})

	handler := NewHandler(router, parser, checker, rewriter, auditLogger, reloadCfg)

	te := &testEnv{
		router:    router,
		pool:      pool,
		handler:   handler,
		reloadCfg: reloadCfg,
		ctx:       context.Background(),
		dbName:    dbName,
	}

	t.Cleanup(func() {
		te.dropTestTables(t)
		pool.Close()
	})

	te.createTestTables(t)

	return te
}

func (te *testEnv) createTestTables(t *testing.T) {
	t.Helper()

	_, err := te.router.Exec(te.ctx, te.dbName, `
		CREATE TABLE IF NOT EXISTS mcp_integ_users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("create mcp_integ_users: %v", err)
	}

	_, err = te.router.Exec(te.ctx, te.dbName, `
		CREATE TABLE IF NOT EXISTS mcp_integ_orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			status ENUM('pending','completed','cancelled') DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("create mcp_integ_orders: %v", err)
	}

	_, err = te.router.Exec(te.ctx, te.dbName,
		"INSERT IGNORE INTO mcp_integ_users (id, name, email) VALUES (1, 'Alice', 'alice@integ.com'), (2, 'Bob', 'bob@integ.com')")
	if err != nil {
		t.Fatalf("insert test data: %v", err)
	}
}

func (te *testEnv) dropTestTables(t *testing.T) {
	t.Helper()
	db, err := te.router.GetDB(te.dbName)
	if err != nil {
		return
	}
	_, _ = db.ExecContext(te.ctx, "DROP TABLE IF EXISTS mcp_integ_orders")
	_, _ = db.ExecContext(te.ctx, "DROP TABLE IF EXISTS mcp_integ_users")
}

// --- Tests ---

func TestIntegrationListTables(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeListTables(te.ctx, te.dbName)
	if err != nil {
		t.Fatalf("executeListTables() error = %v", err)
	}

	names := make([]string, 0, len(result.Tables))
	for _, tbl := range result.Tables {
		names = append(names, tbl.Name)
	}

	sort.Strings(names)

	foundUsers, foundOrders := false, false
	for _, n := range names {
		if n == "mcp_integ_users" {
			foundUsers = true
		}
		if n == "mcp_integ_orders" {
			foundOrders = true
		}
	}
	if !foundUsers {
		t.Errorf("expected mcp_integ_users in tables, got %v", names)
	}
	if !foundOrders {
		t.Errorf("expected mcp_integ_orders in tables, got %v", names)
	}
}

// Regression test: list_tables must return tables from the correct database.
// Before the fix, DATABASE() returned NULL because connections were never switched.
func TestIntegrationListTables_DatabaseRegression(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeListTables(te.ctx, te.dbName)
	if err != nil {
		t.Fatalf("executeListTables() error = %v", err)
	}

	if len(result.Tables) == 0 {
		t.Fatal("expected at least one table, got zero — DATABASE() regression?")
	}

	for _, tbl := range result.Tables {
		if tbl.Name == "" {
			t.Error("got empty table name — possible DATABASE() NULL regression")
		}
	}
}

func TestIntegrationDescribeTable(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeDescribeTable(te.ctx, te.dbName, "mcp_integ_users")
	if err != nil {
		t.Fatalf("executeDescribeTable() error = %v", err)
	}

	if len(result.Columns) == 0 {
		t.Fatal("expected columns, got none")
	}

	colNames := make([]string, 0, len(result.Columns))
	for _, col := range result.Columns {
		colNames = append(colNames, col.Field)
	}

	for _, ec := range []string{"id", "name", "email", "created_at"} {
		found := false
		for _, cn := range colNames {
			if cn == ec {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected column %s, not found in %v", ec, colNames)
		}
	}
}

func TestIntegrationQuery_Select(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeQuery(te.ctx, te.dbName,
		"SELECT id, name, email FROM mcp_integ_users WHERE name = 'Alice'")
	if err != nil {
		t.Fatalf("executeQuery() error = %v", err)
	}

	if len(result.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d: %v", len(result.Columns), result.Columns)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	row := result.Rows[0]
	if row[1] != "Alice" {
		t.Errorf("name = %v, want Alice", row[1])
	}
}

func TestIntegrationQuery_InsertUpdateDelete(t *testing.T) {
	te := newTestEnv(t)

	// INSERT
	result, err := te.handler.executeQuery(te.ctx, te.dbName,
		"INSERT INTO mcp_integ_users (name, email) VALUES ('Charlie', 'charlie@integ.com')")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("insert rows_affected = %d, want 1", result.RowsAffected)
	}

	// SELECT to verify
	selResult, err := te.handler.executeQuery(te.ctx, te.dbName,
		"SELECT name FROM mcp_integ_users WHERE email = 'charlie@integ.com'")
	if err != nil {
		t.Fatalf("select after insert: %v", err)
	}
	if len(selResult.Rows) != 1 {
		t.Fatalf("expected 1 row after insert, got %d", len(selResult.Rows))
	}

	// UPDATE
	updResult, err := te.handler.executeQuery(te.ctx, te.dbName,
		"UPDATE mcp_integ_users SET name = 'Charles' WHERE email = 'charlie@integ.com'")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updResult.RowsAffected != 1 {
		t.Errorf("update rows_affected = %d, want 1", updResult.RowsAffected)
	}

	// DELETE
	delResult, err := te.handler.executeQuery(te.ctx, te.dbName,
		"DELETE FROM mcp_integ_users WHERE email = 'charlie@integ.com'")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if delResult.RowsAffected != 1 {
		t.Errorf("delete rows_affected = %d, want 1", delResult.RowsAffected)
	}
}

func TestIntegrationQuery_SecurityBlock(t *testing.T) {
	te := newTestEnv(t)

	_, err := te.handler.executeQuery(te.ctx, te.dbName,
		"DROP TABLE mcp_integ_users")
	if err == nil {
		t.Fatal("expected DROP to be blocked, got nil error")
	}
}

func TestIntegrationShowCreateTable(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeShowCreateTable(te.ctx, te.dbName, "mcp_integ_users")
	if err != nil {
		t.Fatalf("executeShowCreateTable() error = %v", err)
	}

	if result.CreateStatement == "" {
		t.Fatal("expected non-empty CREATE TABLE statement")
	}
}

func TestIntegrationExplain(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeExplain(te.ctx, te.dbName,
		"SELECT * FROM mcp_integ_users WHERE id = 1")
	if err != nil {
		t.Fatalf("executeExplain() error = %v", err)
	}

	if len(result.Columns) == 0 {
		t.Fatal("expected explain columns, got none")
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected explain rows, got none")
	}
}

func TestIntegrationExplain_RejectsDrop(t *testing.T) {
	te := newTestEnv(t)

	_, err := te.handler.executeExplain(te.ctx, te.dbName,
		"DROP TABLE mcp_integ_users")
	if err == nil {
		t.Fatal("expected EXPLAIN DROP to be rejected")
	}
}

func TestIntegrationSearchTables(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeSearchTables(te.ctx, "mcp_integ")
	if err != nil {
		t.Fatalf("executeSearchTables() error = %v", err)
	}

	if len(result.Matches) == 0 {
		t.Fatal("expected at least one match, got none")
	}

	foundUsers, foundOrders := false, false
	for _, m := range result.Matches {
		if m.Table == "mcp_integ_users" {
			foundUsers = true
		}
		if m.Table == "mcp_integ_orders" {
			foundOrders = true
		}
	}
	if !foundUsers {
		t.Errorf("expected mcp_integ_users in search results, got %v", result.Matches)
	}
	if !foundOrders {
		t.Errorf("expected mcp_integ_orders in search results, got %v", result.Matches)
	}
}

func TestIntegrationListDatabases(t *testing.T) {
	te := newTestEnv(t)

	result, err := te.handler.executeListDatabases(te.ctx)
	if err != nil {
		t.Fatalf("executeListDatabases() error = %v", err)
	}

	if len(result.Databases) != 1 || result.Databases[0] != te.dbName {
		t.Errorf("expected [%s], got %v", te.dbName, result.Databases)
	}
}
