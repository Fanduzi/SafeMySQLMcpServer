package mcp

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/database"
	"github.com/fan/safe-mysql-mcp/internal/security"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// newExecuteTestHandler creates a Handler backed by a sqlmock DB.
// Returns handler, mock, and cleanup.
func newExecuteTestHandler(t *testing.T, dbName string) (*Handler, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	pool := database.NewTestPool("primary", db)
	router := database.NewRouter(pool, config.DatabasesConfig{
		dbName: {Cluster: "primary"},
	})

	securityRules := &config.SecurityRules{
		AllowedDML:   []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		AllowedDDL:   []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
		Blocked:      []string{"DROP", "TRUNCATE", "RENAME"},
		AutoLimit:    1000,
		MaxLimit:     10000,
		QueryTimeout: 30e9,
		MaxRows:      10000,
	}

	parser := security.NewParser()
	checker := security.NewChecker(securityRules)
	rewriter := security.NewRewriter(securityRules)
	auditLogger, _ := audit.NewLogger(&config.AuditConfig{Enabled: false})

	reloadCfg := &config.ReloadableConfig{}
	reloadCfg.Update(&config.Config{}, &config.SecurityConfig{Security: *securityRules})

	handler := NewHandler(router, parser, checker, rewriter, auditLogger, reloadCfg)
	cleanup := func() { _ = db.Close() }

	return handler, mock, cleanup
}

func TestHandler_ExecuteQuery_SelectSuccess(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice").AddRow(2, "Bob"),
	)

	result, err := h.executeQuery(context.Background(), "testdb", "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("executeQuery() error = %v", err)
	}
	if result.RowsAffected != 2 {
		t.Errorf("RowsAffected = %d, want 2", result.RowsAffected)
	}
	if len(result.Columns) != 2 {
		t.Errorf("Columns = %v, want 2 columns", result.Columns)
	}
}

func TestHandler_ExecuteQuery_InsertSuccess(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := h.executeQuery(context.Background(), "testdb", "INSERT INTO users (name) VALUES ('test')")
	if err != nil {
		t.Fatalf("executeQuery() error = %v", err)
	}
	if result.RowsAffected != 1 {
		t.Errorf("RowsAffected = %d, want 1", result.RowsAffected)
	}
	if result.LastInsertID != 1 {
		t.Errorf("LastInsertID = %d, want 1", result.LastInsertID)
	}
}

func TestHandler_ExecuteQuery_SelectRowLimit(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	// Just verify the function handles maxRows correctly by checking config
	if h.getMaxRows() != 10000 {
		t.Errorf("getMaxRows() = %d, want 10000", h.getMaxRows())
	}
}

func TestHandler_ExecuteListTables_Success(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectQuery("SELECT TABLE_NAME").WithArgs("testdb", 1001).WillReturnRows(
		sqlmock.NewRows([]string{"TABLE_NAME", "TABLE_COMMENT"}).
			AddRow("users", "user table").
			AddRow("orders", ""),
	)

	result, err := h.executeListTables(context.Background(), "testdb")
	if err != nil {
		t.Fatalf("executeListTables() error = %v", err)
	}
	if len(result.Tables) != 2 {
		t.Fatalf("len(Tables) = %d, want 2", len(result.Tables))
	}
	if result.Tables[0].Name != "users" {
		t.Errorf("Tables[0].Name = %q, want %q", result.Tables[0].Name, "users")
	}
	if result.Tables[0].Comment != "user table" {
		t.Errorf("Tables[0].Comment = %q, want %q", result.Tables[0].Comment, "user table")
	}
}

func TestHandler_ExecuteListTables_UnknownDB(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	_, err := h.executeListTables(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown database")
	}
}

func TestHandler_ExecuteDescribeTable_Success(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectQuery("SELECT(.+)FROM INFORMATION_SCHEMA.COLUMNS").WithArgs("testdb", "users").WillReturnRows(
		sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra", "Comment"}).
			AddRow("id", "int(11)", "NO", "PRI", nil, "auto_increment", "primary key").
			AddRow("name", "varchar(100)", "NO", "", nil, "", ""),
	)

	result, err := h.executeDescribeTable(context.Background(), "testdb", "users")
	if err != nil {
		t.Fatalf("executeDescribeTable() error = %v", err)
	}
	if len(result.Columns) != 2 {
		t.Fatalf("len(Columns) = %d, want 2", len(result.Columns))
	}
	if result.Columns[0].Field != "id" {
		t.Errorf("Columns[0].Field = %q, want %q", result.Columns[0].Field, "id")
	}
}

func TestHandler_ExecuteShowCreateTable_Success(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectQuery("SHOW CREATE TABLE").WillReturnRows(
		sqlmock.NewRows([]string{"Table", "Create Table"}).
			AddRow("users", "CREATE TABLE `users` (`id` int(11) NOT NULL AUTO_INCREMENT)"),
	)

	result, err := h.executeShowCreateTable(context.Background(), "testdb", "users")
	if err != nil {
		t.Fatalf("executeShowCreateTable() error = %v", err)
	}
	if !contains(result.CreateStatement, "CREATE TABLE") {
		t.Errorf("CreateStatement = %q, want CREATE TABLE", result.CreateStatement)
	}
}

func TestHandler_ExecuteExplain_Success(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectExec("USE `testdb`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("EXPLAIN SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "select_type", "table", "type"}).
			AddRow(1, "SIMPLE", "users", "ALL"),
	)

	result, err := h.executeExplain(context.Background(), "testdb", "SELECT * FROM users")
	if err != nil {
		t.Fatalf("executeExplain() error = %v", err)
	}
	if len(result.Rows) != 1 {
		t.Errorf("len(Rows) = %d, want 1", len(result.Rows))
	}
}

func TestHandler_ExecuteSearchTables_Success(t *testing.T) {
	h, mock, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mock.ExpectQuery("SELECT TABLE_NAME").WithArgs("testdb", "%user%", 10001).WillReturnRows(
		sqlmock.NewRows([]string{"TABLE_NAME"}).
			AddRow("users").
			AddRow("user_logs"),
	)

	result, err := h.executeSearchTables(context.Background(), "user")
	if err != nil {
		t.Fatalf("executeSearchTables() error = %v", err)
	}
	if len(result.Matches) != 2 {
		t.Fatalf("len(Matches) = %d, want 2", len(result.Matches))
	}
	if result.Matches[0].Table != "users" {
		t.Errorf("Matches[0].Table = %q, want %q", result.Matches[0].Table, "users")
	}
}

func TestHandler_HandleQueryTool_ValidationError(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mcpResult, queryResult, err := h.handleQueryTool(
		context.Background(), nil, QueryInput{Database: "", SQL: "SELECT 1"},
	)
	if err != nil {
		t.Fatalf("handleQueryTool() error = %v", err)
	}
	if !mcpResult.IsError {
		t.Error("expected error for empty database name")
	}
	if queryResult != nil {
		t.Error("expected nil queryResult on validation error")
	}
}

func TestHandler_HandleListTablesTool_ValidationError(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mcpResult, tableResult, err := h.handleListTablesTool(
		context.Background(), nil, ListTablesInput{Database: ""},
	)
	if err != nil {
		t.Fatalf("handleListTablesTool() error = %v", err)
	}
	if !mcpResult.IsError {
		t.Error("expected error for empty database name")
	}
	if tableResult != nil {
		t.Error("expected nil tableResult on validation error")
	}
}

func TestHandler_HandleDescribeTableTool_ValidationError(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mcpResult, descResult, err := h.handleDescribeTableTool(
		context.Background(), nil, DescribeTableInput{Database: "testdb", Table: ""},
	)
	if err != nil {
		t.Fatalf("handleDescribeTableTool() error = %v", err)
	}
	if !mcpResult.IsError {
		t.Error("expected error for empty table name")
	}
	if descResult != nil {
		t.Error("expected nil descResult on validation error")
	}
}

func TestHandler_HandleExplainTool_ValidationError(t *testing.T) {
	h, _, cleanup := newExecuteTestHandler(t, "testdb")
	defer cleanup()

	mcpResult, explainResult, err := h.handleExplainTool(
		context.Background(), nil, ExplainInput{Database: "testdb", SQL: "DROP TABLE x"},
	)
	if err != nil {
		t.Fatalf("handleExplainTool() error = %v", err)
	}
	if !mcpResult.IsError {
		t.Error("expected error for blocked SQL in explain")
	}
	if explainResult != nil {
		t.Error("expected nil explainResult on validation error")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || sub == "" || searchSubstring(s, sub))
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
