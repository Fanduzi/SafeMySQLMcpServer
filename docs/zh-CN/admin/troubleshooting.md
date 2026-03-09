# 故障排查

常见问题及其解决方案。

## 连接问题
tab: 连接问题

### MySQL 连接被拒绝

**症状:**
```
Error: dial tcp 127.0.0.1:3306: connect: connection refused
```

**解决方案:**

| 检查项 | 命令 |
|-------|---------|
| MySQL 是否运行? | `docker ps` 或 `systemctl status mysql` |
| 主机/端口是否正确? | 检查 `config.yaml` 的 clusters 部分 |
| 网络是否可达? | `telnet mysql-host 3306` |
| 防火墙是否阻止? | 检查 iptables/防火墙规则 |

### 连接池耗尽

**症状:**
```
Error: connection pool exhausted
```

**解决方案:**

```yaml
# 在 config.yaml 中增加连接池大小
clusters:
  primary:
    max_open_conns: 50  # 从默认值增加
    max_idle_conns: 25
```

### 认证失败

**症状:**
```
Error: invalid token: token has expired
```

**解决方案:**

```bash
# 生成新的 token
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d
```

## SQL 问题
tab: SQL 问题

### 查询被阻止

**症状:**
```
Error: SQL blocked: operation not allowed
```

**解决方案:**

1. 检查 `security.yaml` 的允许操作列表
2. 如果合适，将操作添加到允许列表

```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT  # 如果 INSERT 被阻止，添加此项
```

### 查询超时

**症状:**
```
Error: context deadline exceeded
```

**解决方案:**

```yaml
# 在 security.yaml 中增加超时时间
security:
  query_timeout: 60s  # 从 30s 增加
```

### SQL 解析错误

**症状:**
```
Error: SQL parse error: syntax error at position X
```

**解决方案:**

1. 检查 SQL 语法
2. 确保是有效的 MySQL 语法
3. 检查是否有不支持的 MySQL 特性

### 结果集过大

**症状:**
```
Error: result set too large
```

**解决方案:**

```yaml
# 增加最大行数
security:
  max_rows: 50000  # 从 10000 增加
```

或在查询中添加 LIMIT。

## 性能问题
tab: 性能问题

### 慢查询

**症状:**
- 查询耗时 >5 秒
- CPU 使用率高

**解决方案:**

1. 从审计日志分析慢查询:

```bash
jq 'select(.duration_ms > 5000)' logs/audit.log
```

2. 为频繁查询的列添加索引

3. 使用 EXPLAIN 分析:

```bash
# 通过 MCP 工具
"解释这个查询: SELECT * FROM users WHERE email = 'test@example.com'"
```

### 内存使用过高

**症状:**
- OOM 错误
- 容器重启

**解决方案:**

1. 减少 max_rows:

```yaml
security:
  max_rows: 5000  # 从 10000 减少
```

2. 减少连接池:

```yaml
clusters:
  primary:
    max_open_conns: 20  # 从 50 减少
```

3. 添加内存限制 (Docker):

```yaml
services:
  safemysql:
    deploy:
      resources:
        limits:
          memory: 256M
```

## 配置问题
tab: 配置问题

### 配置文件未找到

**症状:**
```
Error: open config/config.yaml: no such file or directory
```

**解决方案:**

```bash
# 检查文件是否存在
ls -la config/config.yaml

# 检查命令中的路径
./bin/safe-mysql-mcp -config config/config.yaml
```

### YAML 格式错误

**症状:**
```
Error: yaml: line 15: could not find expected ':'
```

**解决方案:**

```bash
# 验证 YAML 格式
python -c "import yaml; yaml.safe_load(open('config/config.yaml'))"

# 或使用 yamllint
yamllint config/config.yaml
```

### 环境变量未设置

**症状:**
```
Error: JWT secret not configured
```

**解决方案:**

```bash
# 设置环境变量
export JWT_SECRET="your-secret-key-min-32-characters-long"

# 或在 docker-compose 中
environment:
  JWT_SECRET: ${JWT_SECRET}
```

## 调试模式
tab: 调试模式

### 启用调试日志

```yaml
# config.yaml
server:
  log_level: debug
```

### 查看日志

```bash
# Docker 日志
docker-compose logs -f safemysql

# Systemd 日志
journalctl -u safemysql -f

# 直接日志文件
tail -f /var/log/safemysql/server.log
```

### 常见日志模式

| 日志信息 | 问题 |
|-------------|-------|
| "connection refused" | MySQL 未运行 |
| "invalid token" | JWT 认证失败 |
| "SQL blocked" | 安全规则违规 |
| "context deadline" | 查询超时 |
| "pool exhausted" | 连接数过多 |
