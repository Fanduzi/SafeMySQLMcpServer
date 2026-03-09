# 测试指南

本指南说明 SafeMySQLMcpServer 的测试方法和最佳实践。

## 测试结构

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

## 测试分类
tab: 测试分类

### 单元测试
独立测试各个函数和组件。

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

### 集成测试
测试多个组件协同工作。

```go
func TestIntegrationQuery(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // 需要真实的数据库连接
    db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
    // ...
}
```

### 安全测试
测试安全边界和 SQL 注入防护。

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

## 运行测试
tab: 运行测试

### 所有测试

```bash
# 运行所有测试
go test ./... -v

# 带竞态检测运行
go test ./... -race

# 带覆盖率运行
go test ./... -coverprofile=coverage.out

# 查看覆盖率
go tool cover -html=coverage.out
```

### 特定包

```bash
# 测试 security 包
go test ./internal/security/... -v

# 测试 database 包
go test ./internal/database/... -v
```

### 短测试 vs 完整测试

```bash
# 跳过慢速测试
go test ./... -short

# 运行所有测试（包括慢速测试）
go test ./...
```

## 测试覆盖率
tab: 测试覆盖率

### 覆盖率目标

| 包 | 目标 | 当前 |
|---------|--------|---------|
| internal/validation | 80%+ | 100% |
| internal/security | 80%+ | 87% |
| internal/auth | 80%+ | 90% |
| internal/config | 80%+ | 83% |
| internal/database | 60%+ | 32% |
| internal/server | 60%+ | 38% |

### 覆盖率报告

```bash
# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Mocking
tab: Mocking

### 数据库 Mock

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

### 配置 Mock

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

## 最佳实践
tab: 最佳实践

### 1. 表驱动测试
使用表驱动测试覆盖多种情况。

```go
// 好的做法
tests := []struct {
    name string
    input string
    want string
    err bool
}{
    {"case 1", "input1", "output1", false},
    {"case 2", "input2", "", true},
}

// 不好的做法
func TestCase1(t *testing.T) { /* ... */ }
func TestCase2(t *testing.T) { /* ... */ }
```

### 2. 测试错误情况
始终测试错误处理。

```go
func TestQuery_Error(t *testing.T) {
    _, err := handler.Query(ctx, "nonexistent_db", "SELECT 1")
    if err == nil {
        t.Error("expected error for nonexistent database")
    }
}
```

### 3. 使用 t.Parallel 并行测试

```go
func TestParallel(t *testing.T) {
    t.Parallel()
    // 此测试与其他测试并行运行
}
```

### 4. 清理资源

```go
func TestWithCleanup(t *testing.T) {
    db, err := setupTestDB(t)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close() // 始终清理
}
```
