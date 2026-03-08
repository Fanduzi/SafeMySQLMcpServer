// Package validation provides input validation utilities for the MySQL MCP server.
//
// This package provides functions to validate and sanitize all user inputs
// before they are used in SQL queries or database operations.
//
// # Validation Functions
//
// The package provides validation for:
//   - Database names
//   - Table names
//   - SQL statements
//   - Search patterns
//
// # Identifier Validation
//
// MySQL identifiers (database names, table names, column names) must:
//   - Start with a letter (a-z, A-Z) or underscore (_)
//   - Contain only letters, digits, and underscores
//   - Be at most 64 characters long
//
// # SQL Validation
//
// SQL statements are validated for:
//   - Non-empty content
//   - Maximum length (100KB by default)
//
// # Identifier Quoting
//
// Use QuoteIdentifier to safely quote MySQL identifiers:
//
//	quoted := validation.QuoteIdentifier("table_name")
//	// Result: `table_name`
//
// # LIKE Pattern Escaping
//
// Use EscapeLikePattern to escape special LIKE characters:
//
//	escaped := validation.EscapeLikePattern("100%_test")
//	// Result: 100\%\_test
//
// # Example Usage
//
//	// Validate database name
//	if err := validation.ValidateDatabaseName("my_db"); err != nil {
//	    // handle error
//	}
//
//	// Validate and quote identifier
//	quoted, err := validation.ValidateAndQuoteIdentifier("table_name")
//	if err != nil {
//	    // handle error
//	}
//	// Use quoted identifier in SQL
//
//	// Escape LIKE pattern
//	pattern := validation.EscapeLikePattern(userInput)
//	query := fmt.Sprintf("SELECT * FROM users WHERE name LIKE '%%%s%%'", pattern)
package validation
