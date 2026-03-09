# 错误码参考

本文档列出 SafeMySQLMcpServer 返回的所有错误码和消息。

## HTTP 状态码

| 状态码 | 名称 | 说明 |
|------|------|-------------|
| 200 | OK | 请求成功 |
| 400 | Bad Request | 请求格式无效 |
| 401 | Unauthorized | 缺少或无效的 token |
| 429 | Too Many Requests | 超过限流阈值 |
| 500 | Internal Server Error | 服务器错误 |

## 认证错误
tab: 认证错误

### 缺少 Token

**HTTP 状态码:** 401

```json
{
  "error": "missing authorization token"
}
```

**原因:** 未提供 Authorization 头。

**解决方案:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" ...
```

### 无效 Token

**HTTP 状态码:** 401

```json
{
  "error": "invalid token: signature is invalid"
}
```

**原因:** Token 未使用正确的密钥签名。

**解决方案:** 使用正确的 JWT 密钥生成新 token。

### Token 已过期

**HTTP 状态码:** 401

```json
{
  "error": "invalid token: token has expired"
}
```

**原因:** Token 已过有效期。

**解决方案:** 生成新 token。

## 验证错误
tab: 验证错误

### 无效数据库名

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Validation error: invalid database name"}],
  "isError": true
}
```

**原因:** 数据库名包含无效字符。

**有效格式:** `^[a-zA-Z_][a-zA-Z0-9_]*$`

| 输入 | 有效 |
|-------|-------|
| mydb | ✅ |
| my_db | ✅ |
| MyDb123 | ✅ |
| my-db | ❌ |
| my db | ❌ |
| 123db | ❌ |

### 无效表名

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Validation error: invalid table name"}],
  "isError": true
}
```

**原因:** 表名包含无效字符。

**有效格式:** `^[a-zA-Z_][a-zA-Z0-9_]*$`

### 空 SQL

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Validation error: SQL statement cannot be empty"}],
  "isError": true
}
```

**原因:** SQL 参数为空或仅包含空白字符。

### SQL 过长

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Validation error: SQL statement exceeds maximum length"}],
  "isError": true
}
```

**原因:** SQL 超过 `max_sql_length`（默认: 100,000 字符）。

## 安全错误
tab: 安全错误

### 操作不允许

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "SQL blocked: INSERT operation not allowed"}],
  "isError": true
}
```

**原因:** 操作不在 `allowed_dml` 或 `allowed_ddl` 中。

**解决方案:** 更新 `security.yaml` 允许该操作。

### 被阻止的操作

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "SQL blocked: DROP operation is blocked"}],
  "isError": true
}
```

**原因:** 操作在 `blocked` 列表中。

**解决方案:** 从阻止列表移除（生产环境不推荐）。

### SQL 解析错误

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "SQL parse error: syntax error at position 10"}],
  "isError": true
}
```

**原因:** SQL 语法无效。

**解决方案:** 修正 SQL 语法。

## 数据库错误
tab: 数据库错误

### 未知数据库

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Error: unknown database: mydb"}],
  "isError": true
}
```

**原因:** 数据库未在路由中配置。

**解决方案:** 在配置的 `databases` 部分添加数据库。

### 连接失败

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Error: connection refused"}],
  "isError": true
}
```

**原因:** MySQL 服务器不可达。

**解决方案:** 检查 MySQL 是否运行且网络可达。

### 查询超时

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Error: context deadline exceeded"}],
  "isError": true
}
```

**原因:** 查询执行超过 `query_timeout`。

**解决方案:** 优化查询或增加超时时间。

### 结果集过大

**MCP 错误:**

```json
{
  "content": [{"type": "text", "text": "Error: result set too large"}],
  "isError": true
}
```

**原因:** 查询返回超过 `max_rows` 行。

**解决方案:** 在查询中添加 LIMIT 或增加 `max_rows`。

## 限流错误
tab: 限流错误

### 超过限流阈值

**HTTP 状态码:** 429

```json
{
  "error": "rate limit exceeded"
}
```

**原因:** 同一 IP 请求数过多。

**响应头:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1646824600
Retry-After: 30
```

**解决方案:** 等待限流重置或提高限制。

## 错误处理最佳实践
tab: 最佳实践

### 开发者

1. **始终检查 isError 字段**:
```json
{
  "isError": true,
  "content": [{"type": "text", "text": "错误消息"}]
}
```

2. **解析错误消息**:
```go
if strings.Contains(errMsg, "blocked") {
    // 安全阻止
}
```

3. **处理限流**:
```go
if resp.StatusCode == 429 {
    retryAfter := resp.Header.Get("Retry-After")
    time.Sleep(time.Duration(retryAfter) * time.Second)
}
```

### 运维人员

1. **检查审计日志**获取错误详情
2. **监控** `safemysql_security_blocked_queries_total` 了解被阻止的查询
3. **告警**高错误率
