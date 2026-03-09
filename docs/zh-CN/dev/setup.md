# 开发环境搭建

本指南帮助你搭建 SafeMySQLMcpServer 的开发环境。

## 初始设置
tab: 初始设置

### 1. 克隆仓库

```bash
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer
```

### 2. 安装 Go 依赖

```bash
go mod download
go mod verify
```

### 3. 安装开发工具

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安全扫描
go install github.com/securego/gosec/v2/cmd/gosec@latest

# 漏洞检查
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 4. 设置本地数据库

```bash
# 使用 Docker（推荐）
docker run -d --name mysql-test \
  -e MYSQL_ROOT_PASSWORD=testpassword \
  -e MYSQL_DATABASE=testdb \
  -p 3306:3306 \
  mysql:8.0

# 等待 MySQL 就绪
docker exec mysql-test mysqladmin ping -h localhost -ptestpassword
```

## IDE 配置
tab: IDE 配置

### VS Code

推荐扩展:

| 扩展 | 用途 |
|-----------|---------|
| Go | 官方 Go 扩展 |
| gopls | 语言服务器 |
| Go Night Switches | 语法高亮 |

### 配置

```json
// .vscode/settings.json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  }
  }
}
```

## 运行测试
tab: 运行测试

### 单元测试

```bash
# 运行所有测试
make test

# 带覆盖率运行
go test ./... -race -cover

# 运行特定包
go test ./internal/security/... -v
```

### 集成测试

```bash
# 设置环境变量
export MYSQL_HOST=127.0.0.1
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=testpassword
export MYSQL_DATABASE=testdb

# 运行集成测试
go test ./... -v -run Integration
```

### 使用 Docker Compose 测试

```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行测试
go test ./... -v

# 清理
docker-compose -f docker-compose.test.yml down
```

## 构建
tab: 构建

### 本地构建

```bash
# 构建所有二进制文件
make build

# 二进制文件在 ./bin/ 目录
ls bin/
# safe-mysql-mcp
# mysql-mcp-token
```

### Docker 构建

```bash
# 构建 Docker 镜像
docker build -t safemysql:local .

# 本地运行
docker run -p 8080:8080 safemysql:local
```

## 调试
tab: 调试

### 启用调试日志

```yaml
# config/config.yaml
server:
  log_level: debug
```

### 常见问题

| 问题 | 解决方案 |
|-------|----------|
| Connection refused | 检查 MySQL 是否运行且可访问 |
| Authentication failed | 验证 JWT secret 是否正确 |
| SQL blocked | 检查 security.yaml 规则 |
| Config parse error | 验证 YAML 语法 |
