// Package security provides SQL parsing and security validation for the MySQL MCP server.
//
// This package is responsible for:
//   - Parsing SQL statements using TiDB parser
//   - Validating SQL against security rules
//   - Rewriting SQL for safety (e.g., adding LIMIT)
//   - Blocking dangerous operations
//
// # Security Model
//
// The security model operates on parsed SQL statements. Each SQL statement
// is first parsed into an AST, The AST is then checked against
// security rules to determine if the operation is allowed.
//
// # SQL Types
//
// The parser recognizes the following SQL types:
//   - SELECT: Read operations
//   - INSERT: Create operations
//   - UPDATE: Update operations
//   - DELETE: Delete operations
//   - CREATE_TABLE: Table creation
//   - CREATE_INDEX: Index creation
//   - ALTER_TABLE: Table modification
//   - DROP: Object deletion
//   - TRUNCATE: Table truncation
//   - RENAME: Object renaming
//   - SHOW: Schema inspection
//   - EXPLAIN: Query analysis
//
// # Security Rules
//
// Security rules control:
//   - Which DML operations are allowed (SELECT, INSERT, UPDATE, DELETE)
//   - Which DDL operations are allowed (CREATE, ALTER, etc.)
//   - Which operations are explicitly blocked (DROP, TRUNCATE, etc.)
//   - Maximum row limits for queries
//   - Query timeout values
//   - Whether auto-LIMIT should be added to queries without WHERE
//
// # Example Usage
//
//	parser := security.NewParser()
//	parsed, err := parser.Parse("SELECT * FROM users")
//	if err != nil {
//	    // handle parse error
//	}
//
//	rules := &config.SecurityRules{
//	    AllowedDML: []string{"SELECT", "INSERT"},
//	}
//	checker := security.NewChecker(rules)
//	result := checker.Check(parsed)
//	if !result.Allowed {
//	    // SQL blocked
//	}
//
// # SQL Rewriting
//
// The rewriter can automatically modify SQL for safety:
//   - Adding LIMIT to UPDATE/DELETE without WHERE
//   - Capping LIMIT values in SELECT statements
//
//	rewriter := security.NewRewriter(rules)
//	rewriteResult := rewriter.Rewrite(parsed)
//	if rewriteResult.Changed {
//	    // SQL was modified
//	    safeSQL := rewriteResult.SQL
//	}
package security
