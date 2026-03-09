// Package mcp provides unit tests for MCP tools.
// input: test cases for tool handlers
// output: test coverage for validation and error handling
// pos: test layer, validates MCP tool behavior
// note: if this file changes, update header and internal/mcp/README.md
package mcp

import (
	"testing"

	"github.com/fan/safe-mysql-mcp/internal/validation"
)

// TestValidateQueryInput tests input validation for query tool
func TestValidateQueryInput(t *testing.T) {
	tests := []struct {
		name    string
		input   QueryInput
		wantErr bool
	}{
		{
			name:    "empty database",
			input:   QueryInput{Database: "", SQL: "SELECT 1"},
			wantErr: true,
		},
		{
			name:    "empty SQL",
			input:   QueryInput{Database: "testdb", SQL: ""},
			wantErr: true,
		},
		{
			name:    "invalid database name",
			input:   QueryInput{Database: "test-db", SQL: "SELECT 1"},
			wantErr: true,
		},
		{
			name:    "valid input",
			input:   QueryInput{Database: "testdb", SQL: "SELECT 1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Always validate both fields
			err := validation.ValidateDatabaseName(tt.input.Database)
			if err == nil {
				err = validation.ValidateSQL(tt.input.SQL)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestValidateListTablesInput tests input validation for list_tables tool
func TestValidateListTablesInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ListTablesInput
		wantErr bool
	}{
		{
			name:    "empty database",
			input:   ListTablesInput{Database: ""},
			wantErr: true,
		},
		{
			name:    "invalid database name with special chars",
			input:   ListTablesInput{Database: "test;drop"},
			wantErr: true,
		},
		{
			name:    "valid database name",
			input:   ListTablesInput{Database: "mydb"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateDatabaseName(tt.input.Database)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestValidateDescribeTableInput tests input validation for describe_table tool
func TestValidateDescribeTableInput(t *testing.T) {
	tests := []struct {
		name    string
		input   DescribeTableInput
		wantErr bool
	}{
		{
			name:    "empty database",
			input:   DescribeTableInput{Database: "", Table: "users"},
			wantErr: true,
		},
		{
			name:    "empty table",
			input:   DescribeTableInput{Database: "mydb", Table: ""},
			wantErr: true,
		},
		{
			name:    "invalid table name",
			input:   DescribeTableInput{Database: "mydb", Table: "user-table"},
			wantErr: true,
		},
		{
			name:    "valid input",
			input:   DescribeTableInput{Database: "mydb", Table: "users"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateDatabaseName(tt.input.Database)
			if err == nil {
				err = validation.ValidateTableName(tt.input.Table)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestValidateSearchTablesInput tests input validation for search_tables tool
func TestValidateSearchTablesInput(t *testing.T) {
	tests := []struct {
		name    string
		input   SearchTablesInput
		wantErr bool
	}{
		{
			name:    "empty pattern",
			input:   SearchTablesInput{TablePattern: ""},
			wantErr: true,
		},
		{
			name:    "valid pattern",
			input:   SearchTablesInput{TablePattern: "user%"},
			wantErr: false,
		},
		{
			name:    "pattern too long",
			input:   SearchTablesInput{TablePattern: string(make([]byte, 300))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateSearchPattern(tt.input.TablePattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestValidateExplainInput tests input validation for explain tool
func TestValidateExplainInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ExplainInput
		wantErr bool
	}{
		{
			name:    "empty database",
			input:   ExplainInput{Database: "", SQL: "SELECT 1"},
			wantErr: true,
		},
		{
			name:    "empty SQL",
			input:   ExplainInput{Database: "mydb", SQL: ""},
			wantErr: true,
		},
		{
			name:    "invalid database name",
			input:   ExplainInput{Database: "my-db", SQL: "SELECT 1"},
			wantErr: true,
		},
		{
			name:    "valid input",
			input:   ExplainInput{Database: "mydb", SQL: "SELECT * FROM users"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateDatabaseName(tt.input.Database)
			if err == nil {
				err = validation.ValidateSQL(tt.input.SQL)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestValidateShowCreateTableInput tests input validation for show_create_table tool
func TestValidateShowCreateTableInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ShowCreateTableInput
		wantErr bool
	}{
		{
			name:    "empty database",
			input:   ShowCreateTableInput{Database: "", Table: "users"},
			wantErr: true,
		},
		{
			name:    "empty table",
			input:   ShowCreateTableInput{Database: "mydb", Table: ""},
			wantErr: true,
		},
		{
			name:    "both invalid",
			input:   ShowCreateTableInput{Database: "my-db", Table: "user-table"},
			wantErr: true,
		},
		{
			name:    "valid input",
			input:   ShowCreateTableInput{Database: "mydb", Table: "users"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateDatabaseName(tt.input.Database)
			if err == nil {
				err = validation.ValidateTableName(tt.input.Table)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestQueryResultTypes tests QueryResult struct
func TestQueryResultTypes(t *testing.T) {
	result := &QueryResult{
		Columns:      []string{"id", "name"},
		Rows:         [][]interface{}{{1, "test"}, {2, "test2"}},
		RowsAffected: 2,
		LastInsertID: 0,
		Warnings:     []string{"warning1"},
	}

	if len(result.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(result.Columns))
	}
	if len(result.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result.Rows))
	}
	if result.RowsAffected != 2 {
		t.Errorf("expected 2 rows affected, got %d", result.RowsAffected)
	}
}

// TestListDatabasesResult tests ListDatabasesResult struct
func TestListDatabasesResult(t *testing.T) {
	result := &ListDatabasesResult{
		Databases: []string{"db1", "db2", "db3"},
	}

	if len(result.Databases) != 3 {
		t.Errorf("expected 3 databases, got %d", len(result.Databases))
	}
}

// TestListTablesResult tests ListTablesResult struct
func TestListTablesResult(t *testing.T) {
	result := &ListTablesResult{
		Tables: []TableInfo{
			{Name: "users", Comment: "User table"},
			{Name: "orders", Comment: ""},
		},
		Truncated: true,
	}

	if len(result.Tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(result.Tables))
	}
	if !result.Truncated {
		t.Error("expected Truncated to be true")
	}
}

// TestDescribeTableResult tests DescribeTableResult struct
func TestDescribeTableResult(t *testing.T) {
	result := &DescribeTableResult{
		Columns: []ColumnInfo{
			{Field: "id", Type: "int", Null: "NO", Key: "PRI", Extra: "auto_increment"},
			{Field: "name", Type: "varchar(255)", Null: "YES", Key: "", Default: "NULL"},
		},
	}

	if len(result.Columns) != 2 {
		t.Errorf("expected 2 columns, got %d", len(result.Columns))
	}
	if result.Columns[0].Field != "id" {
		t.Errorf("expected first column to be 'id', got %s", result.Columns[0].Field)
	}
}

// TestSearchTablesResult tests SearchTablesResult struct
func TestSearchTablesResult(t *testing.T) {
	result := &SearchTablesResult{
		Matches: []TableMatch{
			{Database: "db1", Table: "users"},
			{Database: "db2", Table: "users_backup"},
		},
		Truncated: false,
	}

	if len(result.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result.Matches))
	}
	if result.Truncated {
		t.Error("expected Truncated to be false")
	}
}

// TestExplainResult tests ExplainResult struct
func TestExplainResult(t *testing.T) {
	result := &ExplainResult{
		Columns: []string{"id", "select_type", "table"},
		Rows:    [][]interface{}{{1, "SIMPLE", "users"}},
	}

	if len(result.Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(result.Columns))
	}
	if len(result.Rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(result.Rows))
	}
}

// TestShowCreateTableResult tests ShowCreateTableResult struct
func TestShowCreateTableResult(t *testing.T) {
	result := &ShowCreateTableResult{
		CreateStatement: "CREATE TABLE `users` (`id` int PRIMARY KEY)",
	}

	if result.CreateStatement == "" {
		t.Error("expected non-empty create statement")
	}
}

// TestTableInfo tests TableInfo struct
func TestTableInfo(t *testing.T) {
	info := TableInfo{
		Name:    "users",
		Comment: "User accounts table",
	}

	if info.Name != "users" {
		t.Errorf("expected name 'users', got %s", info.Name)
	}
	if info.Comment != "User accounts table" {
		t.Errorf("expected comment 'User accounts table', got %s", info.Comment)
	}
}

// TestColumnInfo tests ColumnInfo struct
func TestColumnInfo(t *testing.T) {
	info := ColumnInfo{
		Field:   "id",
		Type:    "int(11)",
		Null:    "NO",
		Key:     "PRI",
		Default: "",
		Extra:   "auto_increment",
		Comment: "Primary key",
	}

	if info.Field != "id" {
		t.Errorf("expected field 'id', got %s", info.Field)
	}
	if info.Type != "int(11)" {
		t.Errorf("expected type 'int(11)', got %s", info.Type)
	}
}

// TestTableMatch tests TableMatch struct
func TestTableMatch(t *testing.T) {
	match := TableMatch{
		Database: "production",
		Table:    "orders",
	}

	if match.Database != "production" {
		t.Errorf("expected database 'production', got %s", match.Database)
	}
	if match.Table != "orders" {
		t.Errorf("expected table 'orders', got %s", match.Table)
	}
}

// TestDefaultConstants tests default values
func TestDefaultConstants(t *testing.T) {
	if defaultQueryTimeout <= 0 {
		t.Error("defaultQueryTimeout should be positive")
	}
	if defaultMaxRows <= 0 {
		t.Error("defaultMaxRows should be positive")
	}
	if defaultMaxTables <= 0 {
		t.Error("defaultMaxTables should be positive")
	}
}
