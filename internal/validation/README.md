# Validation Module

Input validation utilities for database names, table names, and SQL patterns.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Input Validation                          │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                 Validation Pipeline                      ││
│  │                                                          ││
│  │   User Input ──▶ ValidateDatabaseName ──▶ Valid/Error   ││
│  │                 ValidateTableName     ──▶ Valid/Error   ││
│  │                 ValidateSQL          ──▶ Valid/Error    ││
│  │                 ValidateSearchPattern ──▶ Valid/Error  ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                 SQL Safety Utilities                     ││
│  │                                                          ││
│  │   QuoteIdentifier   → Backtick escaping for identifiers ││
│  │   EscapeLikePattern → Escape % and _ in LIKE patterns   ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| validator.go | Validation functions | ~80 |
| doc.go | Package documentation | ~20 |
| validation_test.go | Unit tests | ~100 |

## Test Coverage
```
Coverage: ~95%
- Valid/invalid database names
- Valid/invalid table names
- Empty and long SQL
- Search pattern edge cases
- Identifier quoting
- LIKE pattern escaping
```

## Exports

### Validation Functions
```go
func ValidateDatabaseName(name string) error
func ValidateTableName(name string) error
func ValidateSQL(sql string) error
func ValidateSearchPattern(pattern string) error
```

### SQL Safety Utilities
```go
func QuoteIdentifier(name string) string
func EscapeLikePattern(pattern string) string
```

## Validation Rules

### Database Name
| Rule | Description |
|------|-------------|
| Length | 1-64 characters |
| Characters | `[a-zA-Z_][a-zA-Z0-9_]*` |
| No numbers at start | `123db` → ❌ |

| Example | Valid |
|---------|-------|
| `mydb` | ✅ |
| `my_db` | ✅ |
| `MyDb123` | ✅ |
| `my-db` | ❌ |
| `my db` | ❌ |
| `123db` | ❌ |

### Table Name
| Rule | Description |
|------|-------------|
| Length | 1-64 characters |
| Characters | `[a-zA-Z_][a-zA-Z0-9_]*` |
| No numbers at start | `123table` → ❌ |

### SQL Statement
| Rule | Description |
|------|-------------|
| Not empty | Must contain non-whitespace |
| Max length | 100,000 characters (configurable) |

### Search Pattern
| Rule | Description |
|------|-------------|
| Not empty | Must contain non-whitespace |
| Max length | 100 characters |
| LIKE chars | `%` and `_` are automatically escaped |

## Usage Examples

### Validate and Quote
```go
// Validate database name
if err := validation.ValidateDatabaseName(input); err != nil {
    return err
}

// Safely quote for SQL
quoted := validation.QuoteIdentifier(tableName)
// "users" → "`users`"
```

### LIKE Pattern
```go
// User searches for "user%"
pattern := validation.EscapeLikePattern("user%")
// → "user\%" (literal match, not wildcard)
```

## Dependencies
```
Upstream: None

Downstream:
  └── internal/mcp  → Validates all tool inputs

External:
  └── regexp  → Standard library
```

## Update Rule
If validation rules change, update:
1. This file
2. validator.go
3. validation_test.go
4. docs/reference/error-codes.md
