# Development Setup

This guide helps you set up your development environment for SafeMySQLMcpServer.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Build and run |
| MySQL | 5.7+ | Database for testing |
| Docker | Latest | Containerized testing (optional) |
| golangci-lint | Latest | Linting |

## Initial Setup
tab: Initial Setup

### 1. Clone the Repository

```bash
git clone https://github.com/YOUR_USERNAME/SafeMySQLMcpServer.git
cd SafeMySQLMcpServer
```

### 2. Install Go Dependencies

```bash
go mod download
go mod verify
```

### 3. Install Development Tools

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security scanner
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Vulnerability checker
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 4. Set up Local Database

```bash
# Using Docker (recommended)
docker run -d --name mysql-test \
  -e MYSQL_ROOT_PASSWORD=testpassword \
  -e MYSQL_DATABASE=testdb \
  -p 3306:3306 \
  mysql:8.0

# Wait for MySQL to be ready
docker exec mysql-test mysqladmin ping -h localhost -ptestpassword
```

## IDE Setup
tab: IDE Setup

### VS Code

Recommended extensions:

| Extension | Purpose |
|-----------|---------|
| Go | Official Go extension |
| gopls | Language server |
| Go Night Switches | Syntax highlighting |

### Configuration

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

## Running Tests
tab: Running Tests

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
go test ./... -race -cover

# Run specific package
go test ./internal/security/... -v
```

### Integration Tests

```bash
# Set environment variables
export MYSQL_HOST=127.0.0.1
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=testpassword
export MYSQL_DATABASE=testdb

# Run integration tests
go test ./... -v -run Integration
```

### Test with Docker Compose

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run tests
go test ./... -v

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Building
tab: Building

### Local Build

```bash
# Build all binaries
make build

# Binaries are in ./bin/
ls bin/
# safe-mysql-mcp
# mysql-mcp-token
```

### Docker Build

```bash
# Build Docker image
docker build -t safemysql:local .

# Run locally
docker run -p 8080:8080 safemysql:local
```

## Debugging
tab: Debugging

### Enable Debug Logging

```yaml
# config/config.yaml
server:
  log_level: debug
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Connection refused | Check MySQL is running and accessible |
| Authentication failed | Verify JWT secret matches |
| SQL blocked | Check security.yaml rules |
| Config parse error | Validate YAML syntax |
