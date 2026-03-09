# 快速开始

本指南帮助你安装和运行 SafeMySQLMcpServer。

## 安装方式

### 方式 1: Docker（推荐）

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer

# 使用 docker-compose 启动（包含 MySQL）
docker-compose up -d

# 检查服务健康状态
curl http://localhost:8080/health
```

### 方式 2: 二进制安装

```bash
# 从源码构建
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer
make build

# 二进制文件位于 ./bin/safe-mysql-mcp
```

### 方式 3: Go Install

```bash
go install github.com/YOUR_USERNAME/safe-mysql-mcp/cmd/server@latest
```

## 配置

### 1. 创建配置文件

```bash
# 复制示例配置
cp -r examples/config config

# 编辑配置
vim config/config.yaml
vim config/security.yaml
```

### 2. 设置环境变量

```bash
# 必需: JWT secret（至少 32 个字符）
export JWT_SECRET="your-secret-key-min-32-characters-long"

# 数据库凭证
export DEV_DB_USER="your_db_user"
export DEV_DB_PASSWORD="your_db_password"
```

### 3. 生成认证 Token

```bash
# 为用户生成 token
./bin/mysql-mcp-token --user admin --email admin@example.com

# 或指定过期时间
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d
```

## 启动服务器

```bash
./bin/safe-mysql-mcp -config config/config.yaml
```

### 验证服务运行

```bash
# 健康检查
curl http://localhost:8080/health
# 预期返回: OK

# 指标端点
curl http://localhost:8080/metrics
# 预期返回: Prometheus 指标输出
```

### 测试认证

```bash
# 将 YOUR_TOKEN 替换为之前生成的 token
curl -X POST http://localhost:8080/mcp \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## 连接 Claude Code

添加到 Claude Code 配置中:

```json
{
  "mcpServers": {
    "mysql-dev": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

详细说明请参考 [集成 Claude Code](./claude-code.md)。

## 下一步

- [配置安全规则](./security-config.md)
- [设置审计日志](../admin/audit-logging.md)
- [使用 Prometheus 监控](../admin/monitoring.md)
