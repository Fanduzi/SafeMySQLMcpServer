# SafeMySQL MCP Server 改进计划

## 总览

| 阶段 | 优先级 | 预计工作量 | 目标分数提升 |
|------|--------|-----------|-------------|
| Phase 1 | CRITICAL | 中 | 45 → 55 |
| Phase 2 | HIGH | 中 | 55 → 65 |
| Phase 3 | HIGH | 大 | 65 → 75 |
| Phase 4 | MEDIUM | 中 | 75 → 80 |
| Phase 5 | LOW | 小 | 80 → 85+ |

---

## Phase 1: 关键安全修复 (CRITICAL)

### 1.1 修复 SQL 注入漏洞

**问题**: `executeShowCreateTable` 和 `executeExplain` 直接拼接用户输入

**文件**: `internal/mcp/tools.go`

**修改内容**:
```go
// Before (漏洞代码)
db.QueryRowContext(ctx, "SHOW CREATE TABLE "+tableName)

// After (安全代码)
// 1. 添加标识符验证函数
func validateIdentifier(name string) error {
    // 只允许字母、数字、下划线，且不能以数字开头
    matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
    if !matched {
        return fmt.Errorf("invalid identifier: %s", name)
    }
    if len(name) > 64 {
        return fmt.Errorf("identifier too long: %s", name)
    }
    return nil
}

// 2. 使用反引号包裹标识符
func quoteIdentifier(name string) string {
    return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// 3. 在执行前验证
if err := validateIdentifier(tableName); err != nil {
    return nil, err
}
db.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE TABLE %s", quoteIdentifier(tableName)))
```

**验收标准**:
- [ ] 所有用户输入的表名/数据库名都经过验证
- [ ] 使用参数化查询或安全的标识符引用
- [ ] 添加 SQL 注入测试用例

---

### 1.2 JWT Secret 安全加固

**问题**: JWT Secret 可能在配置文件中硬编码

**文件**: `internal/config/config.go`, `internal/auth/jwt.go`

**修改内容**:
```go
// 1. 优先从环境变量读取
func (c *ServerConfig) GetJWTSecret() (string, error) {
    // 优先级: 环境变量 > 配置文件
    if secret := os.Getenv("JWT_SECRET"); secret != "" {
        return secret, nil
    }
    if c.JWTSecret != "" {
        // 警告: 配置文件中的 secret 不安全
        log.Println("WARNING: JWT secret loaded from config file, consider using JWT_SECRET env var")
        return c.JWTSecret, nil
    }
    return "", errors.New("JWT secret not configured")
}

// 2. 验证 secret 强度
func validateJWTSecret(secret string) error {
    if len(secret) < 32 {
        return fmt.Errorf("JWT secret must be at least 32 characters, got %d", len(secret))
    }
    return nil
}
```

**验收标准**:
- [ ] 支持从 `JWT_SECRET` 环境变量读取
- [ ] Secret 长度最少 32 字符
- [ ] 启动时验证 secret 强度
- [ ] 配置文件中的 secret 打印警告

---

### 1.3 输入验证增强

**问题**: 缺少对用户输入的系统性验证

**新增文件**: `internal/validation/validator.go`

**修改内容**:
```go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

const (
    maxIdentifierLength = 64
    maxSQLLength        = 100000  // 100KB
    identifierPattern   = `^[a-zA-Z_][a-zA-Z0-9_]*$`
)

// ValidateDatabaseName 验证数据库名
func ValidateDatabaseName(name string) error {
    if name == "" {
        return fmt.Errorf("database name is required")
    }
    if len(name) > maxIdentifierLength {
        return fmt.Errorf("database name too long (max %d)", maxIdentifierLength)
    }
    matched, _ := regexp.MatchString(identifierPattern, name)
    if !matched {
        return fmt.Errorf("invalid database name: %s", name)
    }
    return nil
}

// ValidateTableName 验证表名
func ValidateTableName(name string) error {
    return ValidateDatabaseName(name) // 相同规则
}

// ValidateSQL 验证 SQL 长度
func ValidateSQL(sql string) error {
    if sql == "" {
        return fmt.Errorf("SQL is required")
    }
    if len(sql) > maxSQLLength {
        return fmt.Errorf("SQL too long (max %d bytes)", maxSQLLength)
    }
    return nil
}

// ValidateSearchPattern 验证搜索模式 (转义 LIKE 特殊字符)
func ValidateSearchPattern(pattern string) string {
    // 转义 LIKE 特殊字符: %, _, \
    replacer := strings.NewReplacer(
        "\\", "\\\\",
        "%", "\\%",
        "_", "\\_",
    )
    return replacer.Replace(pattern)
}
```

**验收标准**:
- [ ] 所有工具输入都经过验证
- [ ] 验证逻辑集中管理
- [ ] 错误信息清晰明确

---

### 1.4 添加基础限流

**问题**: 没有限流保护，可被 DoS 攻击

**文件**: `internal/server/middleware.go` (新建)

**修改内容**:
```go
package server

import (
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type IPRateLimiter struct {
    ips map[string]*rate.Limiter
    mu  sync.RWMutex
    r   rate.Limit  // 每秒请求数
    b   int         // 桶容量
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        ips: make(map[string]*rate.Limiter),
        r:   r,
        b:   b,
    }
}

func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    l.mu.Lock()
    defer l.mu.Unlock()

    limiter, exists := l.ips[ip]
    if !exists {
        limiter = rate.NewLimiter(l.r, l.b)
        l.ips[ip] = limiter
    }
    return limiter
}

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
    limiter := NewIPRateLimiter(10, 20) // 10 req/s, burst 20

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getRealIP(r)
        if !limiter.GetLimiter(ip).Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**验收标准**:
- [ ] 按 IP 限流
- [ ] 限流参数可配置
- [ ] 返回 429 状态码

---

## Phase 2: 高优先级修复 (HIGH)

### 2.1 修复 Rewriter 使用 AST

**问题**: 使用字符串操作修改 SQL，容易出错

**文件**: `internal/security/rewriter.go`

**修改内容**: 使用 TiDB Parser 的 AST 修改 SQL

```go
func (r *Rewriter) Rewrite(parsed *ParsedSQL) *RewriteResult {
    if parsed == nil || parsed.AST == nil {
        return &RewriteResult{SQL: "", Changed: false}
    }

    // 直接操作 AST
    switch stmt := parsed.AST.(type) {
    case *ast.SelectStmt:
        if r.rules.MaxLimit > 0 {
            if stmt.Limit == nil || getLimitValue(stmt.Limit.Count) > r.rules.MaxLimit {
                stmt.Limit = &ast.LimitClause{
                    Count: ast.NewValueExpr(r.rules.MaxLimit, "", ""),
                }
                return &RewriteResult{
                    SQL:     restoreSQL(stmt),
                    Changed: true,
                }
            }
        }
    case *ast.UpdateStmt, *ast.DeleteStmt:
        // 类似处理
    }

    return &RewriteResult{SQL: parsed.Original, Changed: false}
}

func restoreSQL(stmt ast.StmtNode) string {
    var sb strings.Builder
    ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &sb)
    stmt.Restore(ctx)
    return sb.String()
}
```

**验收标准**:
- [ ] 所有 SQL 修改通过 AST 完成
- [ ] 删除字符串操作代码
- [ ] 添加重写测试用例

---

### 2.2 修复连接池竞态条件

**问题**: 热更新时直接关闭正在使用的连接

**文件**: `internal/database/pool.go`

**修改内容**:
```go
type Pool struct {
    mu       sync.RWMutex
    clusters map[string]*managedDB
}

type managedDB struct {
    db       *sql.DB
    refCount int32 // 原子计数
    closing  int32 // 标记正在关闭
}

func (p *Pool) GetDB(name string) (*sql.DB, error) {
    p.mu.RLock()
    defer p.mu.RUnlock()

    managed, ok := p.clusters[name]
    if !ok {
        return nil, fmt.Errorf("unknown database: %s", name)
    }

    if atomic.LoadInt32(&managed.closing) == 1 {
        return nil, fmt.Errorf("database %s is being reconnected", name)
    }

    atomic.AddInt32(&managed.refCount, 1)
    return &refCountedDB{
        db:      managed.db,
        managed: managed,
    }, nil
}

func (p *Pool) UpdateConfig(clusters config.ClustersConfig) error {
    p.mu.Lock()
    defer p.mu.Unlock()

    for name, cfg := range clusters {
        old, exists := p.clusters[name]

        if exists {
            // 标记旧连接正在关闭
            atomic.StoreInt32(&old.closing, 1)

            // 等待引用归零或超时
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            for atomic.LoadInt32(&old.refCount) > 0 {
                select {
                case <-ctx.Done():
                    log.Printf("Force closing pool %s with %d active connections", name, old.refCount)
                    goto closeOld
                case <-time.After(100 * time.Millisecond):
                }
            }
        closeOld:
            old.db.Close()
        }

        // 创建新连接
        newDB, err := createDB(cfg)
        if err != nil {
            return err
        }
        p.clusters[name] = &managedDB{db: newDB}
    }
    return nil
}
```

**验收标准**:
- [ ] 引用计数管理
- [ ] 优雅关闭等待
- [ ] 超时强制关闭

---

### 2.3 所有数据库操作添加超时

**问题**: 部分数据库操作没有超时控制

**文件**: `internal/mcp/tools.go`

**修改内容**:
```go
const defaultQueryTimeout = 30 * time.Second

func (h *Handler) executeListTables(ctx context.Context, dbName string) (*ListTablesResult, error) {
    timeout := defaultQueryTimeout
    if rules := h.config.GetSecurity(); rules != nil && rules.QueryTimeout > 0 {
        timeout = rules.QueryTimeout
    }

    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // ... 原有逻辑
}

// 对所有 execute* 方法应用相同模式
```

**验收标准**:
- [ ] 所有 DB 操作都有超时
- [ ] 超时时间可配置
- [ ] 超时错误清晰区分

---

### 2.4 添加结果集大小限制

**问题**: `list_tables` 等操作没有 LIMIT

**文件**: `internal/mcp/tools.go`

**修改内容**:
```go
func (h *Handler) executeListTables(ctx context.Context, dbName string) (*ListTablesResult, error) {
    // ... 获取 db

    maxTables := 1000
    if rules := h.config.GetSecurity(); rules != nil && rules.MaxTables > 0 {
        maxTables = rules.MaxTables
    }

    query := `
        SELECT TABLE_NAME, TABLE_COMMENT
        FROM INFORMATION_SCHEMA.TABLES
        WHERE TABLE_SCHEMA = DATABASE()
        ORDER BY TABLE_NAME
        LIMIT ?
    `

    rows, err := db.QueryContext(ctx, query, maxTables+1) // +1 用于检测是否超过限制
    // ...
}
```

**验收标准**:
- [ ] 所有列表操作有 LIMIT
- [ ] 超过限制时返回警告
- [ ] 限制值可配置

---

## Phase 3: 测试基础设施 (HIGH)

### 3.1 测试目录结构

```
internal/
├── security/
│   ├── parser_test.go
│   ├── checker_test.go
│   ├── rewriter_test.go
│   └── testdata/
│       ├── valid_sql/
│       └── invalid_sql/
├── mcp/
│   ├── tools_test.go
│   └── server_test.go
├── auth/
│   └── jwt_test.go
└── database/
    ├── pool_test.go
    └── router_test.go
```

### 3.2 安全测试优先

**文件**: `internal/security/checker_test.go`

```go
func TestSQLInjectionPrevention(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
        errMatch string
    }{
        {
            name:    "table name with semicolon",
            input:   "users; DROP TABLE users;--",
            wantErr: true,
        },
        {
            name:    "table name with comment",
            input:   "users/**/",
            wantErr: true,
        },
        {
            name:    "valid table name",
            input:   "users",
            wantErr: false,
        },
        {
            name:    "table name with valid underscore",
            input:   "user_accounts",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validation.ValidateTableName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateTableName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 3.3 测试覆盖率目标

| 包 | 目标覆盖率 | 优先级 |
|----|-----------|--------|
| internal/security | 90% | P0 |
| internal/validation | 90% | P0 |
| internal/auth | 85% | P1 |
| internal/mcp | 80% | P1 |
| internal/database | 75% | P2 |
| internal/config | 70% | P2 |

**验收标准**:
- [ ] `go test -cover ./...` 总覆盖率 >= 80%
- [ ] 关键安全路径覆盖率 >= 90%
- [ ] CI 集成覆盖率检查

---

## Phase 4: 中优先级改进 (MEDIUM)

### 4.1 安全配置验证

**文件**: `internal/config/validation.go`

```go
func (s *SecurityRules) Validate() error {
    var errs []error

    // 验证 DML/DDL 配置不冲突
    for _, dml := range s.AllowedDML {
        if s.IsBlocked(dml) {
            errs = append(errs, fmt.Errorf("DML %s is both allowed and blocked", dml))
        }
    }

    // 验证超时配置合理
    if s.QueryTimeout > 5*time.Minute {
        errs = append(errs, fmt.Errorf("query timeout too long: %v (max 5m)", s.QueryTimeout))
    }

    // 验证限制值
    if s.MaxRows > 1000000 {
        errs = append(errs, fmt.Errorf("max_rows too large: %d (max 1M)", s.MaxRows))
    }

    return errors.Join(errs...)
}
```

### 4.2 审计日志增强

**文件**: `internal/audit/logger.go`

```go
// 添加安全事件告警接口
type AlertSink interface {
    Alert(entry Entry)
}

// 添加请求 ID 用于追踪
type Entry struct {
    // ... 现有字段
    RequestID   string `json:"request_id"`
    ClientIP    string `json:"client_ip"`
}

// 敏感字段脱敏
func (e *Entry) sanitizeSQL() {
    if len(e.SQL) > 500 {
        e.SQL = e.SQL[:500] + "...[truncated]"
    }
    // 脱敏可能的敏感数据
    e.SQL = redactSensitiveData(e.SQL)
}
```

### 4.3 连接池指标

**文件**: `internal/database/metrics.go`

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    poolOpenConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "mysql_pool_open_connections",
            Help: "Number of open connections in the pool",
        },
        []string{"cluster"},
    )
    poolWaitCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mysql_pool_wait_count_total",
            Help: "Total number of connections waited for",
        },
        []string{"cluster"},
    )
)
```

---

## Phase 5: 文档与代码质量 (LOW)

### 5.1 包文档

每个包添加 `doc.go`:

```go
// Package security provides SQL parsing and security validation.
//
// The security package is responsible for:
//   - Parsing SQL statements using TiDB parser
//   - Checking SQL against security rules
//   - Rewriting SQL for safety (e.g., adding LIMIT)
//
// Example usage:
//
//	parser := security.NewParser()
//	parsed, err := parser.Parse("SELECT * FROM users")
//	if err != nil {
//	    // handle parse error
//	}
//
//	checker := security.NewChecker(rules)
//	result := checker.Check(parsed)
//	if !result.Allowed {
//	    // SQL blocked
//	}
package security
```

### 5.2 常量定义

**文件**: `internal/constants/constants.go`

```go
package constants

const (
    // MySQL identifier limits
    MaxIdentifierLength = 64

    // Query limits
    DefaultMaxRows      = 10000
    DefaultQueryTimeout = 30 * time.Second
    MaxSQLLength        = 100000 // 100KB

    // Rate limiting
    DefaultRateLimit    = 10  // requests per second
    DefaultRateBurst    = 20  // burst capacity

    // JWT
    MinJWTSecretLength  = 32
)
```

### 5.3 错误处理标准化

**文件**: `internal/errors/errors.go`

```go
package errors

import "errors"

var (
    // Validation errors
    ErrEmptyDatabase    = errors.New("database name is required")
    ErrEmptySQL         = errors.New("SQL is required")
    ErrInvalidIdentifier = errors.New("invalid identifier")

    // Security errors
    ErrSQLBlocked       = errors.New("SQL blocked by security rules")
    ErrSQLParseError    = errors.New("SQL parse error")

    // Auth errors
    ErrMissingToken     = errors.New("missing authentication token")
    ErrInvalidToken     = errors.New("invalid authentication token")

    // Rate limiting
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// Wrap creates a wrapped error with context
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
```

---

## 实施顺序

```
Week 1: Phase 1 (Critical Security)
├── 1.1 SQL 注入修复
├── 1.2 JWT Secret 加固
├── 1.3 输入验证
└── 1.4 基础限流

Week 2: Phase 2 (High Priority)
├── 2.1 Rewriter AST 重构
├── 2.2 连接池竞态修复
├── 2.3 超时控制
└── 2.4 结果集限制

Week 3-4: Phase 3 (Testing)
├── 3.1 测试框架搭建
├── 3.2 安全测试
└── 3.3 覆盖率达标

Week 5: Phase 4 (Medium Priority)
├── 4.1 配置验证
├── 4.2 审计增强
└── 4.3 指标收集

Week 6: Phase 5 (Polish)
├── 5.1 文档完善
├── 5.2 常量整理
└── 5.3 错误标准化
```

---

## 验收清单

### Phase 1 完成标准
- [ ] 无 SQL 注入漏洞 (通过安全测试验证)
- [ ] JWT Secret 必须从环境变量或安全存储获取
- [ ] 所有用户输入经过验证
- [ ] 限流生效 (压测验证)

### Phase 2 完成标准
- [ ] Rewriter 使用 AST，无字符串操作
- [ ] 热更新无竞态条件 (并发测试验证)
- [ ] 所有 DB 操作有超时
- [ ] 列表操作有 LIMIT

### Phase 3 完成标准
- [ ] 测试覆盖率 >= 80%
- [ ] 安全测试覆盖所有注入场景
- [ ] CI 集成测试

### Phase 4 完成标准
- [ ] 配置验证生效
- [ ] 审计日志完整
- [ ] Prometheus 指标可采集

### Phase 5 完成标准
- [ ] 所有包有文档
- [ ] 无魔法数字
- [ ] 错误处理一致

---

## 最终目标

完成所有阶段后:

| 指标 | 当前 | 目标 |
|------|------|------|
| 代码评分 | 45 | 85+ |
| 测试覆盖率 | 0% | 80%+ |
| 安全漏洞 | 5 CRITICAL | 0 |
| 文档完整度 | 30% | 90% |
