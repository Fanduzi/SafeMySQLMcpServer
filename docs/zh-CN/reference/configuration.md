# 配置参考

SafeMySQLMcpServer 配置完整参考。

## 主配置文件 (config.yaml)

### 结构

```yaml
server:
  # 服务器配置
  host: 0.0.0.0
  port: 8080
  jwt_secret: ${JWT_SECRET}  # 环境变量

clusters:
  # MySQL 集群连接
  primary:
    host: mysql.example.com
    port: 3306
    username: ${DB_USER}
    password: ${DB_PASSWORD}

databases:
  # 数据库路由
  mydb:
    cluster: primary

security:
  # 安全配置文件路径
  config_file: config/security.yaml

audit:
  # 审计日志配置
  enabled: true
  log_file: logs/audit.log

rate_limit:
  # 限流配置
  enabled: true
  requests_per_minute: 100
```

## Server 配置
tab: Server 配置

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| host | string | 0.0.0.0 | 服务器监听地址 |
| port | int | 8080 | 服务器监听端口 |
| jwt_secret | string | - | JWT 签名密钥（必需） |
| log_level | string | info | 日志级别（debug, info, warn, error） |

### 环境变量

| 变量 | 用途 |
|----------|----------|
| `${VAR_NAME}` | 在配置中被替换 |
| `${JWT_SECRET}` | JWT 签名密钥 |

## Clusters 配置
tab: Clusters 配置

### 结构

```yaml
clusters:
  primary:
    host: mysql.example.com
    port: 3306
    username: root
    password: secret
    max_open_conns: 50
    max_idle_conns: 25
    conn_max_lifetime: 5m
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| host | string | 必需 | MySQL 主机 |
| port | int | 3306 | MySQL 端口 |
| username | string | 必需 | MySQL 用户名 |
| password | string | "" | MySQL 密码 |
| max_open_conns | int | 50 | 最大打开连接数 |
| max_idle_conns | int | 25 | 最大空闲连接数 |
| conn_max_lifetime | duration | 5m | 连接最大生命周期 |
| conn_max_idle_time | duration | 1m | 连接最大空闲时间 |

### 多集群配置

```yaml
clusters:
  dev-primary:
    host: dev-mysql.example.com
    port: 3306

  dev-replica:
    host: dev-mysql-replica.example.com
    port: 3306

  prod-primary:
    host: prod-mysql.example.com
    port: 3306
```

## Databases 配置
tab: Databases 配置

### 结构

```yaml
databases:
  user_db:
    cluster: primary
  order_db:
    cluster: primary
  analytics_db:
    cluster: replica
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| cluster | string | 必需 | 目标集群名称 |

### 路由逻辑

```
请求 (database: user_db)
         │
         ▼
    databases.yaml
         │
         ▼
    cluster: primary
         │
         ▼
    clusters.yaml
         │
         ▼
    MySQL 连接
```

## Security 配置
tab: Security 配置

### 结构 (security.yaml)

```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE

  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE

  blocked:
    - DROP
    - TRUNCATE

  auto_limit: 1000
  max_limit: 10000
  query_timeout: 30s
  max_rows: 10000
  max_sql_length: 100000
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| allowed_dml | []string | [SELECT, INSERT, UPDATE, DELETE] | 允许的 DML 操作 |
| allowed_ddl | []string | [CREATE_TABLE, CREATE_INDEX, ALTER_TABLE] | 允许的 DDL 操作 |
| blocked | []string | [DROP, TRUNCATE] | 始终阻止的操作 |
| auto_limit | int | 1000 | 不安全操作的自动 LIMIT |
| max_limit | int | 10000 | 最大允许的 LIMIT |
| query_timeout | duration | 30s | 查询执行超时 |
| max_rows | int | 10000 | 最大返回行数 |
| max_sql_length | int | 100000 | SQL 最大长度 |

### DML 值

| 值 | 说明 |
|-------|-------------|
| SELECT | 查询数据 |
| INSERT | 插入数据 |
| UPDATE | 更新数据 |
| DELETE | 删除数据 |

### DDL 值

| 值 | 说明 |
|-------|-------------|
| CREATE_TABLE | 创建表 |
| CREATE_INDEX | 创建索引 |
| ALTER_TABLE | 修改表 |
| DROP_TABLE | 删除表 |

## Audit 配置
tab: Audit 配置

### 结构

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

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| enabled | bool | true | 启用审计日志 |
| log_file | string | logs/audit.log | 日志文件路径 |
| max_sql_length | int | 2000 | N 字符后截断 SQL |
| max_size_mb | int | 100 | 达到 N MB 时轮转 |
| max_backups | int | 10 | 保留 N 个备份文件 |
| max_age_days | int | 30 | N 天后删除备份 |
| compress | bool | true | 压缩轮转日志 |

## Rate Limit 配置
tab: Rate Limit 配置

### 结构

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100
  burst: 20
```

### 字段

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|-------------|
| enabled | bool | true | 启用限流 |
| requests_per_minute | int | 100 | 每 IP 每分钟最大请求数 |
| burst | int | 20 | 突发允许量 |

### 响应头

限流响应头包含在响应中：

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1646824600
```
