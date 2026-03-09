# Testing Guide

This guide explains the testing approach and best practices for SafeMySQLMcpServer.

## Test Structure

```
SafeMySQLMcpServer/
├── internal/
│   ├── security/
│   │   ├── checker.go
│   │   ├── checker_test.go
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── database/
│   │   ├── pool.go
│   │   ├── pool_test.go
│   │   ├── router.go
│   │   └── router_test.go
│   └── ...
└── tests/
    ├── integration/
    │   └── integration_test.go
    └── e2e/
        └── e2e_test.go
```

## Test Categories
tab: Test Categories

### Unit Tests
Test individual functions and components in isolation.

```go
func TestValidateDatabaseName(t *testing.T) {
    tests := []struct {
        name string
        input string
        wantErr bool
    }{
        {"valid", "mydb", false},
        {"invalid with dash", "my-db", true},
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

### Integration Tests
Test multiple components working together.

```go
func TestIntegrationQuery(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Requires real database connection
    db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
    // ...
}
```

### Security Tests
Test security boundaries and SQL injection prevention.

```go
func TestSQLInjectionPrevention(t *testing.T) {
    maliciousInputs := []string{
        "users; DROP TABLE users--",
        "users' OR '1'='1",
        "users; INSERT INTO admin VALUES...",
    }

    for _, input := range maliciousInputs {
        _, err := validation.ValidateDatabaseName(input)
        if err == nil {
            t.Errorf("expected error for malicious input: %s", input)
        }
    }
}
```

## Running Tests
tab: Running Tests

### All Tests

```bash
# Run all tests
go test ./... -v

# Run with race detection
go test ./... -race

# Run with coverage
go test ./... -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Specific Packages

```bash
# Test security package
go test ./internal/security/... -v

# Test database package
go test ./internal/database/... -v
```

### Short vs Full

```bash
# Skip slow tests
go test ./... -short

# Run all including slow tests
go test ./...
```

## Test Coverage
tab: Test Coverage

### Coverage Goals

| Package | Target | Current |
|---------|--------|---------|
| internal/validation | 80%+ | 100% |
| internal/security | 80%+ | 87% |
| internal/auth | 80%+ | 90% |
| internal/config | 80%+ | 83% |
| internal/database | 60%+ | 32% |
| internal/server | 60%+ | 38% |

### Coverage Report

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Mocking
tab: Mocking

### Database Mocking

```go
type MockDB struct {
    QueryFunc func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    if m.QueryFunc != nil {
        return m.QueryFunc(ctx, query, args...)
    }
    return nil, errors.New("not implemented")
}
```

### Config Mocking

```go
type MockConfig struct {
    GetFunc func() *config.Config
}

func (m *MockConfig) Get() *config.Config {
    if m.GetFunc != nil {
        return m.GetFunc()
    }
    return &config.Config{}
}
```

## Best Practices
tab: Best Practices

### 1. Table-Driven Tests
Use table-driven tests for multiple cases.

```go
// GOOD
tests := []struct {
    name string
    input string
    want string
    err bool
}{
    {"case 1", "input1", "output1", false},
    {"case 2", "input2", "", true},
}

// BAD
func TestCase1(t *testing.T) { /* ... */ }
func TestCase2(t *testing.T) { /* ... */ }
```

### 2. Test Error Cases
Always test error handling.

```go
func TestQuery_Error(t *testing.T) {
    _, err := handler.Query(ctx, "nonexistent_db", "SELECT 1")
    if err == nil {
        t.Error("expected error for nonexistent database")
    }
}
```

### 3. Use t.Parallel for Parallel Tests

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // This test runs in parallel with other tests
}
```

### 4. Clean Up Resources

```go
func TestWithCleanup(t *testing.T) {
    db, err := setupTestDB(t)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close() // Always clean up
}
```
