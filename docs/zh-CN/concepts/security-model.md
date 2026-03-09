# 安全模型

SafeMySQLMcpServer 实现多层安全防护，防止 SQL 注入和未授权访问。

## 安全层

```
┌─────────────────────────────────────────────────────────────────┐
│ 第 1 层: 身份认证 (JWT)                                         │
│ - 验证 JWT token                                               │
│ - 提取用户身份                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 第 2 层: 输入验证                                               │
│ - 验证数据库名 (正则: ^[a-zA-Z_][a-zA-Z0-9_]*$)              │
│ - 验证表名 (正则: ^[a-zA-Z_][a-zA-Z0-9_]*$)                   │
│ - 验证 SQL 长度 (最大: 100,000 字符)                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 第 3 层: SQL 解析                                                │
│ - 将 SQL 解析为 AST (抽象语法树)                                 │
│ - 识别 SQL 类型 (SELECT, INSERT, UPDATE, DELETE, DDL)           │
│ - 提取受影响的表和列                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 第 4 层: 安全检查                                                 │
│ - 检查 DML 允许列表                                               │
│ - 检查 DDL 允许列表                                               │
│ - 检查阻止列表                                                   │
│ - 检测危险模式                                                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 第 5 层: SQL 重写                                                  │
│ - 为无 WHERE 的 UPDATE/DELETE 添加 LIMIT                          │
│ - 为所有查询添加超时                                              │
│ - 限制结果集大小                                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 第 6 层: 执行                                                     │
│ - 使用 prepared statement 执行                                    │
│ - 强制查询超时                                                    │
│ - 限制返回行数                                                    │
└─────────────────────────────────────────────────────────────────┘
```

## SQL 注入防护
tab: SQL 注入防护

### 标识符验证
所有标识符（数据库名、表名）使用严格正则验证:

```go
// 仅允许字母数字和下划线
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
```

### Prepared Statements
用户输入永远不会拼接到 SQL 中:

```go
// 错误: SQL 注入漏洞
query := fmt.Sprintf("SELECT * FROM %s", tableName)

// 正确: 参数化查询
query := "SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = ?"
rows, _ := db.Query(query, tableName)
```

### LIKE 模式转义
搜索模式被转义以防止 LIKE 注入:

```go
func EscapeLikePattern(pattern string) string {
    r := strings.NewReplacer(
        "%", "\\%",
        "_", "\\_",
        "\\", "\\\\",
    )
    return r.Replace(pattern)
}
```

### 引用标识符
当需要动态标识符时，正确引用:

```go
func QuoteIdentifier(name string) string {
    return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
```

## 访问控制
tab: 访问控制

### JWT 身份认证
所有请求必须包含有效的 JWT token:

```http
POST /mcp HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

### 数据库路由
用户只能访问 `databases` 部分配置的数据库:

```yaml
databases:
  user_db:
    cluster: dev-cluster-1
  order_db:
    cluster: dev-cluster-1
  # 用户无法访问未在此列出的数据库
```

### 操作允许列表
只有明确允许的操作可以执行:

```yaml
security:
  allowed_dml: [SELECT]  # 仅允许 SELECT
  allowed_ddl: []     # 不允许 DDL
```

## 速率限制
tab: 速率限制

### 基于 IP 的速率限制
每个 IP 地址有请求限制:

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100  # 每个 IP 每分钟最多 100 请求
  burst: 20              # 允许突发到 20 请求
```

### 速率限制头
响应包含速率限制信息:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1646824600
```

## 审计追踪
tab: 审计追踪

### 记录的信息
每个 SQL 操作都被记录:

| 字段 | 说明 | 示例 |
|-------|-------------|---------|
| timestamp | 操作发生时间 | 2026-03-09T10:30:00Z |
| user_id | JWT 中的用户 | admin |
| user_email | JWT 中的邮箱 | admin@example.com |
| database | 目标数据库 | mydb |
| sql | SQL 语句（截断） | SELECT * FROM users... |
| sql_type | SQL 类型 | SELECT |
| status | 结果状态 | success, error, blocked |
| rows_affected | 影响行数 | 10 |
| duration_ms | 执行时间 | 15 |

### 日志格式
JSON 格式便于解析:

```json
{
  "timestamp": "2026-03-09T10:30:00.123Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "database": "mydb",
  "sql": "SELECT * FROM users LIMIT 10",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 10,
  "duration_ms": 15,
  "client_ip": "192.168.1.100"
}
```
