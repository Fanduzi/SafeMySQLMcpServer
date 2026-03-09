# 身份认证

SafeMySQLMcpServer 使用 JWT (JSON Web Token) 进行身份认证。

## 概述

```
客户端                    服务器
   │                        │
   │  1. 请求 + Token    │
   │ ─────────────────────>│
   │                        │
   │  2. 验证 Token     │
   │                        │
   │  3. 返回响应      │
   │ <─────────────────────│
```

## Token 生成
tab: Token 生成

### 使用 Token CLI 工具

```bash
# 基本用法
./bin/mysql-mcp-token --user admin --email admin@example.com

# 自定义过期时间（默认: 24h）
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d

# 使用环境变量设置密钥
export JWT_SECRET="your-secret-key"
./bin/mysql-mcp-token --user admin --email admin@example.com
```

### CLI 选项

| 参数 | 必需 | 默认值 | 说明 |
|------|----------|---------|-------------|
| `--user` | 是 | - | 用户标识符 |
| `--email` | 是 | - | 用户邮箱 |
| `--expire` | 否 | 24h | Token 过期时间 |
| `--secret` | 否 | - | JWT 密钥（或使用 JWT_SECRET 环境变量） |

### Token 输出

Token 输出到 stdout。 可在 Authorization header 中使用：

```bash
# 复制 token 到剪贴板 (macOS)
./bin/mysql-mcp-token --user admin --email admin@example.com | pbcopy

# 在 curl 中使用
curl -H "Authorization: Bearer $(./bin/mysql-mcp-token --user admin --email admin@example.com)" ...
```

## 安全最佳实践
tab: 安全最佳实践

### 1. 不要记录 Token

```bash
# 错误: Token 会保留在 shell 历史中
TOKEN=$(./bin/mysql-mcp-token --user admin --email admin@example.com)

# 正确: 直接使用，./bin/mysql-mcp-token --user admin --email admin@example.com | pbcopy
```

### 2. 使用强密钥

JWT 密钥至少需要 32 个字符：

```bash
# 生成强密钥
openssl rand -base64 32
```

### 3. 定期轮换密钥

定期轮换 JWT 密钥：

1. 生成新密钥
2. 更新配置
3. 重启服务
4. 生成新 Token
5. 旧 Token 会自然过期

### 4. 使用环境变量

永远不要在配置文件中硬编码密钥：

```yaml
# 正确
server:
  jwt_secret: ${JWT_SECRET}

# 错误
server:
  jwt_secret: "hardcoded-secret-do-not-do-this"
```

## Token 验证
tab: Token 验证

服务器在每个 MCP 请求时验证 token：

```go
// 服务端验证
claims, err := validator.Validate(token)
if err != nil {
    return http.Error(401, "invalid token")
}
```

### Token Claims

| Claim | 说明 |
|-------|-------------|
| `sub` | 用户 ID |
| `email` | 用户邮箱 |
| `iat` | 签发时间戳 |
| `exp` | 过期时间戳 |
### 错误响应
| 状态码 | 错误 | 说明 |
|-------------|-------|-------------|
| 401 | missing authorization token | 未提供 token |
| 401 | invalid token | Token 无效或已过期 |
| 401 | token has expired | Token 已过期 |
