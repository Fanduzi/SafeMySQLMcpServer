# Validation Module

Input validation utilities for database names, table names, and SQL patterns.

## Files
| File | Responsibility |
|------|---------------|
| validator.go | Validation functions |
| doc.go | Package documentation |
| *_test.go | Unit tests |

## Exports
- `ValidateDatabaseName(name string) error` - Validate database name
- `ValidateTableName(name string) error` - Validate table name
- `ValidateSQL(sql string) error` - Validate SQL statement
- `ValidateSearchPattern(pattern string) error` - Validate search pattern
- `QuoteIdentifier(name string) string` - Quote SQL identifier safely
- `EscapeLikePattern(pattern string) string` - Escape LIKE special chars

## Validation Rules
### Database Name
- 1-64 characters
- Alphanumeric and underscore only
- Cannot start with number

### Table Name
- 1-64 characters
- Alphanumeric and underscore only
- Cannot start with number

### SQL
- Cannot be empty
- Max length: 1MB

### Search Pattern
- Cannot be empty
- Max length: 100 characters

## Dependencies
- Upstream: None
- Downstream: `internal/mcp` - Validates all inputs

## Update Rule
If validation rules change, update this file in the same change.
