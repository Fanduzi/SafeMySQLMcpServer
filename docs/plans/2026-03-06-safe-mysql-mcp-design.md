# SafeMySQLMcpServer 设计文档

> Date: 2026-03-06
> Version: v1.0

## 1. 项目概述

### 1.1 目标

为研发提供安全的 MySQL MCP Server，让 Claude Code 等 AI 工具能安全地操作数据库。

### 1.2 架构

```
研发 MacBook
    ↓ Claude Code
    ↓ HTTP (MCP 协议)
SafeMySQLMcpServer (Go)
    ↓ MySQL 协议
多个 MySQL 集群 (dev/test 环境)
```

### 1.3 使用场景

- 开发环境 + 测试环境
- 研发日常开发调试
- AI 辅助查询和操作数据库

## 2. 核心功能模块

```
┌─────────────────────────────────────────────────────────────┐
│                    SafeMySQLMcpServer                        │
├─────────────────────────────────────────────────────────────┤
│  MCP 协议层                                                  │
│  ├── tools: query, list_databases, list_tables, ...         │
│  └── HTTP Server (Streamable HTTP Transport)                │
├─────────────────────────────────────────────────────────────┤
│  安全校验层                                                  │
│  ├── Token 认证 (JWT)                                       │
│  ├── SQL 解析 (pingcap/parser)                              │
│  ├── 安全规则检查                                           │
│  └── SQL 改写 (自动加 LIMIT)                                │
├─────────────────────────────────────────────────────────────┤
│  执行层                                                      │
│  ├── 连接池管理                                             │
│  ├── SQL 执行 (超时控制)                                     │
│  └── 结果截断 (行数限制)                                     │
├─────────────────────────────────────────────────────────────┤
│  审计层                                                      │
│  └── 完整审计日志 (JSON 文件)                               │
└─────────────────────────────────────────────────────────────┘
```

## 3. 安全规则配置

### 3.1 配置文件 (config/security.yaml)

```yaml
security:
  # 允许的 DML 操作
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE

  # 允许的 DDL 操作（细分）
  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE

  # 禁止的操作
  blocked:
    - DROP
    - TRUNCATE
    - RENAME
    - CREATE_USER
    - CREATE_DATABASE
    - CREATE_VIEW
    - CREATE_FUNCTION
    - CREATE_PROCEDURE
    - GRANT
    - REVOKE
    - SET_GLOBAL

  # DML 安全增强
  auto_limit: 1000        # 无 WHERE 的 DML 自动加 LIMIT
  max_limit: 10000        # 用户指定的 LIMIT 最大值

  # 执行限制
  query_timeout: 30s      # 单次查询超时
  max_rows: 10000         # 最大返回行数
```

### 3.2 特性

- 配置文件修改后自动热更新（fsnotify 监听）
- 支持 DDL 细粒度控制（CREATE_TABLE vs CREATE_VIEW 等）

## 4. MCP Tools 设计

### 4.1 Tool 列表

| Tool | 功能 | 参数 |
|------|------|------|
| query | 执行 SQL 查询 | database, sql |
| list_databases | 列出所有可用数据库 | 无 |
| list_tables | 列出指定数据库的表 | database |
| describe_table | 查看表结构 | database, table |
| show_create_table | 查看建表语句 | database, table |
| explain | 查看 SQL 执行计划 | database, sql |
| search_tables | 按表名搜索 | table_pattern |

### 4.2 智能匹配

- 研发说 "查 users 表" 时，调用 `search_tables` 查找
- 多个库有同名表时，返回列表让研发选择
- 实时查询 INFORMATION_SCHEMA.TABLES

### 4.3 数据库路由

研发只需指定 database，不需要知道 cluster。Server 内部维护映射关系。

## 5. 认证设计

### 5.1 JWT Token 结构

```json
{
  "sub": "zhangsan",
  "email": "zhangsan@company.com",
  "iat": 1709600000,
  "exp": 1712200000
}
```

### 5.2 Token 生成方式

**方式 1：CLI 工具**
```bash
$ mysql-mcp token create --user zhangsan --email zhangsan@company.com --expire 365d
```

**方式 2：HTTP 接口（后续 SSO 集成用）**
```
POST /auth/login
{ "username": "zhangsan", "password": "xxx" }
→ { "token": "eyJhbGciOiJIUzI1NiIs..." }
```

### 5.3 请求验证流程

```
请求 → Header: Authorization: Bearer <token>
     → 解析 JWT 获取用户身份
     → 记录到审计日志
     → 执行 SQL
```

## 6. 审计日志设计

### 6.1 日志格式

```json
{
  "timestamp": "2024-03-05T10:30:00Z",
  "user_id": "zhangsan",
  "user_email": "zhangsan@company.com",
  "database": "user_db",
  "sql": "SELECT * FROM users LIMIT 10",
  "sql_type": "SELECT",
  "status": "success",
  "block_reason": "",
  "rows_affected": 10,
  "duration_ms": 23
}
```

### 6.2 SQL 截断

- 超过 2000 字符自动截断
- 记录原始长度

### 6.3 日志轮转

```yaml
audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true
```

使用 lumberjack 库实现。

## 7. 配置文件设计

### 7.1 主配置 (config/config.yaml)

```yaml
# 服务配置
server:
  host: 0.0.0.0
  port: 8080
  jwt_secret: ${JWT_SECRET}

# 数据库连接配置
clusters:
  dev-cluster-1:
    host: dev-mysql-1.internal.company.com
    port: 3306
    username: ${DEV_DB_USER}
    password: ${DEV_DB_PASSWORD}

  dev-cluster-2:
    host: dev-mysql-2.internal.company.com
    port: 3306
    username: ${DEV_DB_USER}
    password: ${DEV_DB_PASSWORD}

# 数据库路由映射
databases:
  user_db:
    cluster: dev-cluster-1
  order_db:
    cluster: dev-cluster-1
  product_db:
    cluster: dev-cluster-2
  log_db:
    cluster: dev-cluster-2

# 安全规则（独立文件）
security:
  config_file: config/security.yaml

# 审计配置
audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true
```

### 7.2 热更新支持

| 配置项 | 热更新 | 说明 |
|--------|--------|------|
| clusters | ✅ | 新增/修改集群连接 |
| databases | ✅ | 新增/修改数据库路由 |
| security | ✅ | 安全规则 |
| audit | ✅ | 审计配置 |
| server.port | ❌ | 需要重启 |
| jwt_secret | ❌ | 不建议热更新 |

## 8. 项目结构

```
SafeMySQLMcpServer/
├── cmd/
│   └── server/
│       └── main.go              # 入口
│
├── internal/
│   ├── config/
│   │   ├── config.go            # 配置加载
│   │   └── watcher.go           # 热更新监听
│   │
│   ├── server/
│   │   └── http.go              # HTTP Server
│   │
│   ├── mcp/
│   │   ├── handler.go           # MCP 协议处理
│   │   └── tools.go             # Tool 定义
│   │
│   ├── database/
│   │   ├── pool.go              # 连接池管理
│   │   └── router.go            # 数据库路由
│   │
│   ├── security/
│   │   ├── parser.go            # SQL 解析
│   │   ├── checker.go           # 安全规则检查
│   │   └── rewriter.go          # SQL 改写（加 LIMIT）
│   │
│   ├── auth/
│   │   └── jwt.go               # JWT 验证
│   │
│   └── audit/
│       ├── logger.go            # 审计日志
│       └── rotate.go            # 日志轮转
│
├── pkg/
│   └── token/
│       └── main.go              # Token 生成 CLI 工具
│
├── config/
│   ├── config.yaml              # 主配置
│   └── security.yaml            # 安全规则
│
├── logs/
│   └── audit.log                # 审计日志
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 9. 技术栈

```
┌─────────────────────────────────────────────────────────────┐
│  语言：Go 1.21+                                              │
├─────────────────────────────────────────────────────────────┤
│  核心依赖：                                                   │
│  ├── github.com/modelcontextprotocol/go-sdk/mcp  # MCP 协议 │
│  ├── go-sql-driver/mysql                          # MySQL   │
│  ├── pingcap/parser                               # SQL 解析 │
│  ├── golang-jwt/jwt/v5                            # JWT     │
│  ├── lumberjack                                   # 日志轮转 │
│  ├── fsnotify                                     # 配置监听 │
│  └── viper                                        # 配置管理 │
├─────────────────────────────────────────────────────────────┤
│  开发工具：                                                   │
│  ├── golangci-lint         # 代码检查                        │
│  └── go test               # 单元测试                        │
└─────────────────────────────────────────────────────────────┘
```

## 9.1 MCP SDK 使用方式

使用官方 `github.com/modelcontextprotocol/go-sdk/mcp`：

```go
// 创建 MCP Server
server := mcp.NewServer(&mcp.Implementation{
    Name:    "safe-mysql-mcp",
    Version: "v1.0.0",
}, nil)

// 添加 Tool（自动生成 input/output schema）
mcp.AddTool(server, &mcp.Tool{
    Name:        "query",
    Description: "Execute SQL query on specified database",
}, handleQuery)

// 使用 StreamableHTTPHandler 提供 HTTP 服务
handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
    return server
}, nil)

http.Handle("/mcp", handler)
http.ListenAndServe(":8080", nil)
```

### 9.2 认证集成

使用 SDK 的 middleware 机制：

```go
server.AddReceivingMiddleware(func(h mcp.MethodHandler) mcp.MethodHandler {
    return func(ctx context.Context, method string, params mcp.Params) (mcp.Result, error) {
        // 从 context 获取 HTTP header 中的 JWT token
        // 验证 token 并注入用户信息到 context
        return h(ctx, method, params)
    }
})
```

### 9.3 审计集成

在 Tool Handler 中记录审计日志：

```go
func handleQuery(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, QueryResult, error) {
    start := time.Now()
    // 从 context 获取用户信息
    userID := auth.GetUserID(ctx)
    userEmail := auth.GetUserEmail(ctx)

    // 执行查询...

    // 记录审计日志
    audit.Log(audit.Entry{
        UserID:     userID,
        UserEmail:  userEmail,
        Database:   args.Database,
        SQL:        truncateSQL(args.SQL),
        Duration:   time.Since(start),
        Status:     "success",
    })

    return nil, result, nil
}
```

## 10. 部署

### 10.1 单机部署

```bash
# 编译
make build

# 运行
./bin/safe-mysql-mcp -config config/config.yaml
```

### 10.2 环境变量

```bash
export JWT_SECRET=your-secret-key
export DEV_DB_USER=dev_user
export DEV_DB_PASSWORD=dev_password
```

## 11. 后续扩展

| 功能 | 说明 |
|------|------|
| SSO 集成 | LDAP/OAuth 登录 |
| 权限管理 | 用户-数据库映射 |
| 管理后台 | 独立项目，查看审计日志、管理 Token |
