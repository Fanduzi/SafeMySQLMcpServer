// Package security handles SQL parsing and security checks.
// input: raw SQL string from user
// output: ParsedSQL with type, tables, columns
// pos: security layer, parses SQL using TiDB parser AST
// note: if this file changes, update header and internal/security/README.md
package security

import (
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
)

// ParsedSQL represents a parsed SQL statement
type ParsedSQL struct {
	Original string
	Type     SQLType
	Tables   []string
	HasWhere bool
	HasLimit bool
	Limit    int
	AST      ast.StmtNode
}

// SQLType represents the type of SQL statement
type SQLType string

const (
	SQLTypeSelect      SQLType = "SELECT"
	SQLTypeInsert      SQLType = "INSERT"
	SQLTypeUpdate      SQLType = "UPDATE"
	SQLTypeDelete      SQLType = "DELETE"
	SQLTypeCreateTable SQLType = "CREATE_TABLE"
	SQLTypeCreateIndex SQLType = "CREATE_INDEX"
	SQLTypeAlterTable  SQLType = "ALTER_TABLE"
	SQLTypeDrop        SQLType = "DROP"
	SQLTypeTruncate    SQLType = "TRUNCATE"
	SQLTypeRename      SQLType = "RENAME"
	SQLTypeShow        SQLType = "SHOW"
	SQLTypeExplain     SQLType = "EXPLAIN"
	SQLTypeOther       SQLType = "OTHER"
)

// Parser parses SQL statements
type Parser struct {
	parser *parser.Parser
}

// NewParser creates a new SQL parser
func NewParser() *Parser {
	return &Parser{
		parser: parser.New(),
	}
}

// Parse parses a SQL statement
func (p *Parser) Parse(sql string) (*ParsedSQL, error) {
	stmts, _, err := p.parser.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	if len(stmts) == 0 {
		return nil, nil
	}

	return p.parseStmt(stmts[0], sql)
}

// parseStmt parses a single statement
func (p *Parser) parseStmt(stmt ast.StmtNode, original string) (*ParsedSQL, error) {
	result := &ParsedSQL{
		Original: original,
		AST:      stmt,
	}

	switch s := stmt.(type) {
	case *ast.SelectStmt:
		result.Type = SQLTypeSelect
		result.HasWhere = s.Where != nil
		if s.Limit != nil {
			result.HasLimit = true
			// Try to get the limit value from the count expression
			if s.Limit.Count != nil {
				result.Limit = getLimitValue(s.Limit.Count)
			}
		}
		result.Tables = extractTablesFromTableRefs(s.From)

	case *ast.InsertStmt:
		result.Type = SQLTypeInsert
		result.Tables = extractTablesFromTableRefs(s.Table)

	case *ast.UpdateStmt:
		result.Type = SQLTypeUpdate
		result.HasWhere = s.Where != nil
		if s.Limit != nil {
			result.HasLimit = true
		}
		result.Tables = extractTablesFromTableRefs(s.TableRefs)

	case *ast.DeleteStmt:
		result.Type = SQLTypeDelete
		result.HasWhere = s.Where != nil
		if s.Limit != nil {
			result.HasLimit = true
		}
		result.Tables = extractTablesFromTableRefs(s.TableRefs)

	case *ast.CreateTableStmt:
		result.Type = SQLTypeCreateTable
		if s.Table != nil {
			result.Tables = []string{s.Table.Name.String()}
		}

	case *ast.CreateIndexStmt:
		result.Type = SQLTypeCreateIndex
		if s.Table != nil {
			result.Tables = []string{s.Table.Name.String()}
		}

	case *ast.AlterTableStmt:
		result.Type = SQLTypeAlterTable
		if s.Table != nil {
			result.Tables = []string{s.Table.Name.String()}
		}

	case *ast.DropTableStmt:
		result.Type = SQLTypeDrop
		for _, t := range s.Tables {
			result.Tables = append(result.Tables, t.Name.String())
		}

	case *ast.TruncateTableStmt:
		result.Type = SQLTypeTruncate
		if s.Table != nil {
			result.Tables = []string{s.Table.Name.String()}
		}

	case *ast.RenameTableStmt:
		result.Type = SQLTypeRename
		for _, t := range s.TableToTables {
			if t.OldTable != nil {
				result.Tables = append(result.Tables, t.OldTable.Name.String())
			}
		}

	case *ast.ShowStmt:
		result.Type = SQLTypeShow

	default:
		result.Type = SQLTypeOther
	}

	return result, nil
}

// getLimitValue extracts the limit value from an expression
func getLimitValue(expr ast.ExprNode) int {
	if expr == nil {
		return 0
	}

	// The tidb parser uses ast.ValueExpr as an interface
	// Try to get the value using reflection-like approach
	if val, ok := expr.(interface{ GetValue() interface{} }); ok {
		if v := val.GetValue(); v != nil {
			switch n := v.(type) {
			case int64:
				return int(n)
			case int:
				return n
			case float64:
				return int(n)
			}
		}
	}

	return 0
}

// extractTablesFromTableRefs extracts table names from a TableRefsClause
func extractTablesFromTableRefs(node *ast.TableRefsClause) []string {
	if node == nil {
		return nil
	}

	return extractTablesFromNode(node.TableRefs)
}

// extractTablesFromNode extracts table names from a ResultSetNode
func extractTablesFromNode(node ast.ResultSetNode) []string {
	if node == nil {
		return nil
	}

	var tables []string

	switch n := node.(type) {
	case *ast.TableSource:
		if t, ok := n.Source.(*ast.TableName); ok {
			tables = append(tables, t.Name.String())
		}

	case *ast.Join:
		tables = append(tables, extractTablesFromNode(n.Left)...)
		tables = append(tables, extractTablesFromNode(n.Right)...)

	case *ast.TableName:
		tables = append(tables, n.Name.String())
	}

	return tables
}

// IsDML checks if the SQL type is a DML operation
func (t SQLType) IsDML() bool {
	switch t {
	case SQLTypeSelect, SQLTypeInsert, SQLTypeUpdate, SQLTypeDelete:
		return true
	default:
		return false
	}
}

// IsDDL checks if the SQL type is a DDL operation
func (t SQLType) IsDDL() bool {
	switch t {
	case SQLTypeCreateTable, SQLTypeCreateIndex, SQLTypeAlterTable,
		SQLTypeDrop, SQLTypeTruncate, SQLTypeRename:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (t SQLType) String() string {
	return string(t)
}

// GetLimit returns the limit value if set
func (p *ParsedSQL) GetLimit() (int, bool) {
	if p.HasLimit && p.Limit > 0 {
		return p.Limit, true
	}
	return 0, false
}
