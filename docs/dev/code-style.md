# Code Style Guide

This guide defines coding standards for SafeMySQLMcpServer.

## General Principles

### 1. Simplicity
- Prefer simple, readable code over clever code
- Avoid premature abstraction
- Keep functions focused and small (<50 lines)

### 2. Consistency
- Follow existing patterns in the codebase
- Use consistent naming conventions
- Maintain consistent error handling

### 3. Documentation
- Document public APIs
- Explain "why", not "what"
- Keep comments up-to-date

## Go Conventions
tab: Go Conventions

### Naming

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase | `security`, `database` |
| Exported | PascalCase | `NewValidator`, `ValidateToken` |
| Unexported | camelCase | `validateInput`, `parseSQL` |
| Constants | UPPER_SNAKE or PascalCase | `MaxRows`, `QUERY_TIMEOUT` |
| Interfaces | -er or -or suffix | `Validator`, `Parser` |

### Error Handling

```go
// GOOD: Wrap errors with context
if err != nil {
    return fmt.Errorf("validate token: %w", err)
}

// BAD: Lose error context
if err != nil {
    return errors.New("validation failed")
}
```

### Context Usage

```go
// GOOD: Pass context as first parameter
func (h *Handler) Query(ctx context.Context, db, sql string) (*Result, error) {
    // ...
}

// BAD: Ignore context
func (h *Handler) Query(db, sql string) (*Result, error) {
    // ...
}
```

## File Organization
tab: File Organization

### Package Structure

```
internal/security/
├── checker.go        # Security rule checker
├── checker_test.go   # Checker tests
├── parser.go         # SQL parser
├── parser_test.go    # Parser tests
├── rewriter.go       # SQL rewriter
├── rewriter_test.go  # Rewriter tests
├── doc.go            # Package documentation
└── README.md         # Module documentation
```

### File Header Comments

Every Go file should start with a header comment:

```go
// Package security handles SQL parsing and security validation.
// input: SQL statements, security rules configuration
// output: Parsed SQL, security check results, rewritten SQL
// pos: security layer, between input validation and database execution
// note: if this file changes, update header and internal/security/README.md
package security
```

### README Structure

Each module should have a README with:

```markdown
# Module Name

Brief description of the module.

## Files
| File | Responsibility |
|------|---------------|
| file.go | What it does |

## Exports
- `Function()` - Description

## Dependencies
- Upstream: packages it depends on
- Downstream: packages that depend on it

## Update Rule
If X changes, update this file in the same change.
```

## Testing Standards
tab: Testing Standards

### Test File Naming

```
checker.go         -> checker_test.go
parser.go          -> parser_test.go
integration_test.go -> Integration suffix for integration tests
```

### Test Function Naming

```go
func TestFunctionName(t *testing.T)                    // Basic test
func TestFunctionName_Scenario(t *testing.T)          // Specific scenario
func TestIntegration_FunctionName(t *testing.T)       // Integration test
```

### Table-Driven Tests

```go
func TestValidateDatabaseName(t *testing.T) {
    tests := []struct {
        name string
        input string
        wantErr bool
    }{
        {"valid lowercase", "mydb", false},
        {"valid with underscore", "my_db", false},
        {"invalid with dash", "my-db", true},
        {"invalid with space", "my db", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validation.ValidateDatabaseName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
        })
    }
}
```

## Code Review Checklist
tab: Code Review Checklist

Before submitting a PR, verify:

- [ ] Code compiles without warnings
- [ ] Tests pass locally
- [ ] New code has tests
- [ ] Test coverage maintained or improved
- [ ] File headers are updated
- [ ] Module README is updated
- [ ] No hardcoded secrets
- [ ] Error handling is comprehensive
- [ ] Context is properly passed
- [ ] Documentation is clear and accurate
