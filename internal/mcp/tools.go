// Package mcp implements MCP tools for MySQL operations.
// input: QueryInput, ListTablesInput, etc. from MCP SDK
// output: QueryResult, ListTablesResult, etc. to MCP SDK
// pos: tool layer, bridges MCP protocol to database layer
// note: if this file changes, update header and internal/mcp/README.md
package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/auth"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/database"
	"github.com/fan/safe-mysql-mcp/internal/security"
	"github.com/fan/safe-mysql-mcp/internal/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Default timeouts and limits
const (
	defaultQueryTimeout = 30 * time.Second
	defaultMaxRows      = 10000
	defaultMaxTables    = 1000
)

// QueryInput represents arguments for the query tool
type QueryInput struct {
	Database string `json:"database" jsonschema:"required,Database name"`
	SQL      string `json:"sql" jsonschema:"required,SQL statement to execute"`
}

// QueryResult represents the result of a query
type QueryResult struct {
	Columns      []string        `json:"columns,omitempty"`
	Rows         [][]interface{} `json:"rows,omitempty"`
	RowsAffected int64           `json:"rows_affected,omitempty"`
	LastInsertID int64           `json:"last_insert_id,omitempty"`
	Warnings     []string        `json:"warnings,omitempty"`
}

// ListDatabasesResult represents the result of list_databases
type ListDatabasesResult struct {
	Databases []string `json:"databases"`
}

// ListTablesInput represents arguments for list_tables
type ListTablesInput struct {
	Database string `json:"database" jsonschema:"required,Database name"`
}

// ListTablesResult represents the result of list_tables
type ListTablesResult struct {
	Tables    []TableInfo `json:"tables"`
	Truncated bool        `json:"truncated,omitempty"`
}

// TableInfo represents table information
type TableInfo struct {
	Name    string `json:"name"`
	Comment string `json:"comment,omitempty"`
}

// DescribeTableInput represents arguments for describe_table
type DescribeTableInput struct {
	Database string `json:"database" jsonschema:"required,Database name"`
	Table    string `json:"table" jsonschema:"required,Table name"`
}

// DescribeTableResult represents the result of describe_table
type DescribeTableResult struct {
	Columns []ColumnInfo `json:"columns"`
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Field   string `json:"field"`
	Type    string `json:"type"`
	Null    string `json:"null"`
	Key     string `json:"key"`
	Default string `json:"default,omitempty"`
	Extra   string `json:"extra,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// ShowCreateTableInput represents arguments for show_create_table
type ShowCreateTableInput struct {
	Database string `json:"database" jsonschema:"required,Database name"`
	Table    string `json:"table" jsonschema:"required,Table name"`
}

// ShowCreateTableResult represents the result of show_create_table
type ShowCreateTableResult struct {
	CreateStatement string `json:"create_statement"`
}

// ExplainInput represents arguments for explain
type ExplainInput struct {
	Database string `json:"database" jsonschema:"required,Database name"`
	SQL      string `json:"sql" jsonschema:"required,SQL statement to explain"`
}

// ExplainResult represents the result of explain
type ExplainResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

// SearchTablesInput represents arguments for search_tables
type SearchTablesInput struct {
	TablePattern string `json:"table_pattern" jsonschema:"required,Table name pattern to search"`
}

// SearchTablesResult represents the result of search_tables
type SearchTablesResult struct {
	Matches   []TableMatch `json:"matches"`
	Truncated bool         `json:"truncated,omitempty"`
}

// TableMatch represents a matched table
type TableMatch struct {
	Database string `json:"database"`
	Table    string `json:"table"`
}

// Handler handles MCP tool calls
type Handler struct {
	router   *database.Router
	parser   *security.Parser
	checker  *security.Checker
	rewriter *security.Rewriter
	audit    *audit.Logger
	config   *config.ReloadableConfig
}

// NewHandler creates a new MCP handler
func NewHandler(
	router *database.Router,
	parser *security.Parser,
	checker *security.Checker,
	rewriter *security.Rewriter,
	auditLogger *audit.Logger,
	cfg *config.ReloadableConfig,
) *Handler {
	return &Handler{
		router:   router,
		parser:   parser,
		checker:  checker,
		rewriter: rewriter,
		audit:    auditLogger,
		config:   cfg,
	}
}

// RegisterTools registers all MCP tools with the SDK server
func RegisterTools(server *mcp.Server, h *Handler) {
	// Query tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query",
		Description: "Execute SQL query on specified database",
	}, h.handleQueryTool)

	// List databases tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_databases",
		Description: "List all available databases",
	}, h.handleListDatabasesTool)

	// List tables tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tables",
		Description: "List tables in specified database",
	}, h.handleListTablesTool)

	// Describe table tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "describe_table",
		Description: "Get table structure",
	}, h.handleDescribeTableTool)

	// Show create table tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "show_create_table",
		Description: "Get CREATE TABLE statement",
	}, h.handleShowCreateTableTool)

	// Explain tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "explain",
		Description: "Get SQL execution plan",
	}, h.handleExplainTool)

	// Search tables tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_tables",
		Description: "Search tables by name across all databases",
	}, h.handleSearchTablesTool)
}

// handleQueryTool handles the query tool with SDK signature
func (h *Handler) handleQueryTool(ctx context.Context, req *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, *QueryResult, error) {
	// Validate input
	if err := validation.ValidateDatabaseName(input.Database); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	if err := validation.ValidateSQL(input.SQL); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeQuery(ctx, input.Database, input.SQL)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleListDatabasesTool handles the list_databases tool with SDK signature
func (h *Handler) handleListDatabasesTool(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, *ListDatabasesResult, error) {
	result, err := h.executeListDatabases(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleListTablesTool handles the list_tables tool with SDK signature
func (h *Handler) handleListTablesTool(ctx context.Context, req *mcp.CallToolRequest, input ListTablesInput) (*mcp.CallToolResult, *ListTablesResult, error) {
	// Validate input
	if err := validation.ValidateDatabaseName(input.Database); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeListTables(ctx, input.Database)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleDescribeTableTool handles the describe_table tool with SDK signature
func (h *Handler) handleDescribeTableTool(ctx context.Context, req *mcp.CallToolRequest, input DescribeTableInput) (*mcp.CallToolResult, *DescribeTableResult, error) {
	// Validate input
	if err := validation.ValidateDatabaseName(input.Database); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	if err := validation.ValidateTableName(input.Table); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeDescribeTable(ctx, input.Database, input.Table)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleShowCreateTableTool handles the show_create_table tool with SDK signature
func (h *Handler) handleShowCreateTableTool(ctx context.Context, req *mcp.CallToolRequest, input ShowCreateTableInput) (*mcp.CallToolResult, *ShowCreateTableResult, error) {
	// Validate input
	if err := validation.ValidateDatabaseName(input.Database); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	if err := validation.ValidateTableName(input.Table); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeShowCreateTable(ctx, input.Database, input.Table)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleExplainTool handles the explain tool with SDK signature
func (h *Handler) handleExplainTool(ctx context.Context, req *mcp.CallToolRequest, input ExplainInput) (*mcp.CallToolResult, *ExplainResult, error) {
	// Validate input
	if err := validation.ValidateDatabaseName(input.Database); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	if err := validation.ValidateSQL(input.SQL); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeExplain(ctx, input.Database, input.SQL)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// handleSearchTablesTool handles the search_tables tool with SDK signature
func (h *Handler) handleSearchTablesTool(ctx context.Context, req *mcp.CallToolRequest, input SearchTablesInput) (*mcp.CallToolResult, *SearchTablesResult, error) {
	// Validate input
	if err := validation.ValidateSearchPattern(input.TablePattern); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Validation error: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	result, err := h.executeSearchTables(ctx, input.TablePattern)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}, nil, nil
	}
	return nil, result, nil
}

// getQueryTimeout returns the configured query timeout
func (h *Handler) getQueryTimeout() time.Duration {
	if rules := h.config.GetSecurity(); rules != nil && rules.QueryTimeout > 0 {
		return rules.QueryTimeout
	}
	return defaultQueryTimeout
}

// getMaxRows returns the configured max rows
func (h *Handler) getMaxRows() int {
	if rules := h.config.GetSecurity(); rules != nil && rules.MaxRows > 0 {
		return rules.MaxRows
	}
	return defaultMaxRows
}

// executeQuery executes a SQL query
func (h *Handler) executeQuery(ctx context.Context, db, sqlStr string) (*QueryResult, error) {
	start := time.Now()
	userID := auth.GetUserID(ctx)
	userEmail := auth.GetUserEmail(ctx)

	result := &QueryResult{}

	// Parse SQL
	parsed, err := h.parser.Parse(sqlStr)
	if err != nil {
		h.audit.Log(audit.Entry{
			UserID:      userID,
			UserEmail:   userEmail,
			Database:    db,
			SQL:         sqlStr,
			SQLType:     "UNKNOWN",
			Status:      "error",
			BlockReason: fmt.Sprintf("parse error: %v", err),
			DurationMs:  time.Since(start).Milliseconds(),
		})
		return nil, fmt.Errorf("SQL parse error: %w", err)
	}

	// Security check
	checkResult := h.checker.Check(parsed)
	if !checkResult.Allowed {
		h.audit.Log(audit.Entry{
			UserID:      userID,
			UserEmail:   userEmail,
			Database:    db,
			SQL:         sqlStr,
			SQLType:     string(parsed.Type),
			Status:      "blocked",
			BlockReason: checkResult.Reason,
			DurationMs:  time.Since(start).Milliseconds(),
		})
		return nil, fmt.Errorf("SQL blocked: %s", checkResult.Reason)
	}

	// Rewrite if needed
	sqlToExecute := sqlStr
	if checkResult.AutoRewrite {
		rewriteResult := h.rewriter.Rewrite(parsed)
		sqlToExecute = rewriteResult.SQL
		result.Warnings = append(result.Warnings, "SQL was auto-rewritten for safety")
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	// Determine if it's a query or exec
	if parsed.Type == security.SQLTypeSelect || parsed.Type == security.SQLTypeShow || parsed.Type == security.SQLTypeExplain {
		rows, err := h.router.Query(execCtx, db, sqlToExecute)
		if err != nil {
			h.audit.Log(audit.Entry{
				UserID:      userID,
				UserEmail:   userEmail,
				Database:    db,
				SQL:         sqlStr,
				SQLType:     string(parsed.Type),
				Status:      "error",
				BlockReason: err.Error(),
				DurationMs:  time.Since(start).Milliseconds(),
			})
			return nil, err
		}
		defer func() { _ = rows.Close() }()

		// Get column names
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		result.Columns = cols

		// Fetch rows with limit
		maxRows := h.getMaxRows()

		rowCount := 0
		for rows.Next() && rowCount < maxRows {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			// Convert []byte to string for JSON serialization
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					values[i] = string(b)
				}
			}

			result.Rows = append(result.Rows, values)
			rowCount++
		}
		result.RowsAffected = int64(rowCount)

	} else {
		// Exec for INSERT, UPDATE, DELETE, DDL
		execResult, err := h.router.Exec(execCtx, db, sqlToExecute)
		if err != nil {
			h.audit.Log(audit.Entry{
				UserID:      userID,
				UserEmail:   userEmail,
				Database:    db,
				SQL:         sqlStr,
				SQLType:     string(parsed.Type),
				Status:      "error",
				BlockReason: err.Error(),
				DurationMs:  time.Since(start).Milliseconds(),
			})
			return nil, err
		}

		rowsAffected, err := execResult.RowsAffected()
		if err != nil {
			log.Printf("Warning: Could not get rows affected: %v", err)
		} else {
			result.RowsAffected = rowsAffected
		}

		lastInsertID, err := execResult.LastInsertId()
		if err != nil {
			// LastInsertId may not be supported for all queries, log but don't fail
			log.Printf("Warning: Could not get last insert ID: %v", err)
		} else {
			result.LastInsertID = lastInsertID
		}
	}

	// Log success
	h.audit.Log(audit.Entry{
		UserID:       userID,
		UserEmail:    userEmail,
		Database:     db,
		SQL:          sqlStr,
		SQLType:      string(parsed.Type),
		Status:       "success",
		RowsAffected: result.RowsAffected,
		DurationMs:   time.Since(start).Milliseconds(),
	})

	return result, nil
}

// executeListDatabases lists all available databases
func (h *Handler) executeListDatabases(ctx context.Context) (*ListDatabasesResult, error) {
	databases := h.router.ListDatabases()
	return &ListDatabasesResult{Databases: databases}, nil
}

// executeListTables lists tables in a database
func (h *Handler) executeListTables(ctx context.Context, dbName string) (*ListTablesResult, error) {
	db, err := h.router.GetDB(dbName)
	if err != nil {
		return nil, err
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	maxTables := defaultMaxTables

	query := `
		SELECT TABLE_NAME, TABLE_COMMENT
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME
		LIMIT ?
	`

	rows, err := db.QueryContext(execCtx, query, maxTables+1)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close rows for list tables: %v", err)
		}
	}()

	var tables []TableInfo
	count := 0
	truncated := false

	for rows.Next() {
		if count >= maxTables {
			truncated = true
			break
		}
		var name, comment string
		if err := rows.Scan(&name, &comment); err != nil {
			return nil, err
		}
		tables = append(tables, TableInfo{Name: name, Comment: comment})
		count++
	}

	return &ListTablesResult{Tables: tables, Truncated: truncated}, nil
}

// executeDescribeTable describes a table structure
func (h *Handler) executeDescribeTable(ctx context.Context, dbName, tableName string) (*DescribeTableResult, error) {
	db, err := h.router.GetDB(dbName)
	if err != nil {
		return nil, err
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	query := `
		SELECT
			COLUMN_NAME as Field,
			COLUMN_TYPE as Type,
			IS_NULLABLE as 'Null',
			COLUMN_KEY as 'Key',
			COLUMN_DEFAULT as 'Default',
			EXTRA as Extra,
			COLUMN_COMMENT as Comment
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.QueryContext(execCtx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close rows for describe table: %v", err)
		}
	}()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var defaultVal, comment sql.NullString
		if err := rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &defaultVal, &col.Extra, &comment); err != nil {
			return nil, err
		}
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		if comment.Valid {
			col.Comment = comment.String
		}
		columns = append(columns, col)
	}

	return &DescribeTableResult{Columns: columns}, nil
}

// executeShowCreateTable shows the CREATE TABLE statement
// FIXED: Use validated and quoted identifier to prevent SQL injection
func (h *Handler) executeShowCreateTable(ctx context.Context, dbName, tableName string) (*ShowCreateTableResult, error) {
	db, err := h.router.GetDB(dbName)
	if err != nil {
		return nil, err
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	// SAFE: Table name is already validated by handler, use quoted identifier
	quotedTable := validation.QuoteIdentifier(tableName)
	query := fmt.Sprintf("SHOW CREATE TABLE %s", quotedTable)

	var tblName, createStmt string
	err = db.QueryRowContext(execCtx, query).Scan(&tblName, &createStmt)
	if err != nil {
		return nil, err
	}

	return &ShowCreateTableResult{CreateStatement: createStmt}, nil
}

// executeExplain explains a SQL statement
// FIXED: Parse and validate the SQL before using in EXPLAIN
func (h *Handler) executeExplain(ctx context.Context, dbName, sqlStr string) (*ExplainResult, error) {
	db, err := h.router.GetDB(dbName)
	if err != nil {
		return nil, err
	}

	// Parse the SQL to ensure it's valid and safe
	parsed, err := h.parser.Parse(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SQL for EXPLAIN: %w", err)
	}

	// Only allow SELECT, UPDATE, DELETE, INSERT for EXPLAIN
	switch parsed.Type {
	case security.SQLTypeSelect, security.SQLTypeInsert, security.SQLTypeUpdate, security.SQLTypeDelete:
		// OK
	default:
		return nil, fmt.Errorf("EXPLAIN not supported for SQL type: %s", parsed.Type)
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	// SAFE: SQL has been parsed and validated, reconstruct from AST if possible
	// For now, we use the original SQL since it passed parsing
	// #nosec G202 -- sqlStr is parsed and restricted to supported statement types above.
	query := "EXPLAIN " + sqlStr

	rows, err := db.QueryContext(execCtx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close rows for explain: %v", err)
		}
	}()

	// Get column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &ExplainResult{Columns: cols}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert []byte to string
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		result.Rows = append(result.Rows, values)
	}

	return result, nil
}

// executeSearchTables searches tables by pattern
func (h *Handler) executeSearchTables(ctx context.Context, tablePattern string) (*SearchTablesResult, error) {
	databases := h.router.ListDatabases()
	var matches []TableMatch

	// SAFE: Escape LIKE special characters
	escapedPattern := validation.EscapeLikePattern(tablePattern)
	pattern := strings.ToLower(escapedPattern)

	maxMatches := h.getMaxRows()
	truncated := false

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, h.getQueryTimeout())
	defer cancel()

	for _, dbName := range databases {
		if truncated {
			break
		}

		db, err := h.router.GetDB(dbName)
		if err != nil {
			continue
		}

		query := `
			SELECT TABLE_NAME
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = ? AND LOWER(TABLE_NAME) LIKE ?
			ORDER BY TABLE_NAME
			LIMIT ?
		`

		remaining := maxMatches - len(matches)
		if remaining <= 0 {
			truncated = true
			break
		}

		if err := func() error {
			rows, err := db.QueryContext(execCtx, query, dbName, "%"+pattern+"%", remaining+1)
			if err != nil {
				return err
			}
			defer func() {
				if err := rows.Close(); err != nil {
					log.Printf("close rows for search tables: %v", err)
				}
			}()

			for rows.Next() {
				if len(matches) >= maxMatches {
					truncated = true
					break
				}
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					continue
				}
				matches = append(matches, TableMatch{
					Database: dbName,
					Table:    tableName,
				})
			}

			return rows.Err()
		}(); err != nil {
			continue
		}
	}

	return &SearchTablesResult{Matches: matches, Truncated: truncated}, nil
}

// MarshalResult marshals a result to JSON for legacy compatibility
func MarshalResult(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		return fmt.Sprintf(`{"error": "failed to marshal result: %s"}`, err.Error())
	}
	return string(data)
}
