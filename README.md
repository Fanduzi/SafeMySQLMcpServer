# SafeMySQLMcpServer

A secure MySQL MCP (Model Context Protocol) Server that allows AI tools like Claude Code to safely operate databases in development and testing environments.

## Features

- **MCP Protocol Support**: Full MCP protocol implementation with HTTP transport
- **JWT Authentication**: Secure token-based authentication
- **SQL Security**:
  - SQL parsing and validation
  - Fine-grained DML/DDL control
  - Auto LIMIT for dangerous operations
  - Query timeout and row limits
- **Audit Logging**: Complete audit trail with JSON format and log rotation
- **Hot Configuration Reload**: Update settings without restart
- **Prometheus Metrics**: Comprehensive observability support
- **Docker Support**: Production-ready containerization
- **CI/CD**: GitHub Actions pipeline with automated testing and security scans

## Quick Start

### Using Docker (Recommended)

```bash
# Start with docker-compose (includes MySQL)
docker-compose up -d

# Generate a token
docker exec safemysql-app /app/token -user admin -email admin@example.com -secret your-jwt-secret

# Check health
curl http://localhost:8080/health
```

### Manual Setup

#### 1. Build

```bash
make build
```

#### 2. Configure

Copy and edit configuration files:

```bash
cp config/config.yaml.example config/config.yaml
cp config/security.yaml.example config/security.yaml
```

Set environment variables:

```bash
export JWT_SECRET=your-secret-key-min-32-characters
export DEV_DB_USER=your-db-user
export DEV_DB_PASSWORD=your-db-password
```

#### 3. Generate Token

```bash
./bin/mysql-mcp-token --user zhangsan --email zhangsan@company.com --expire 365d
```

#### 4. Run Server

```bash
./bin/safe-mysql-mcp -config config/config.yaml
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `POST /mcp` | MCP JSON-RPC endpoint (requires auth) |
| `GET /health` | Health check endpoint |
| `GET /metrics` | Prometheus metrics endpoint |

## MCP Tools

| Tool | Description |
|------|-------------|
| `query` | Execute SQL query on specified database |
| `list_databases` | List all available databases |
| `list_tables` | List tables in specified database |
| `describe_table` | Get table structure |
| `show_create_table` | Get CREATE TABLE statement |
| `explain` | Get SQL execution plan |
| `search_tables` | Search tables by name across databases |

## Configuration

### Main Configuration (config/config.yaml)

```yaml
server:
  host: 0.0.0.0
  port: 8080
  jwt_secret: ${JWT_SECRET}

clusters:
  dev-cluster-1:
    host: mysql.example.com
    port: 3306
    username: ${DEV_DB_USER}
    password: ${DEV_DB_PASSWORD}

databases:
  user_db:
    cluster: dev-cluster-1

audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true
```

### Security Configuration (config/security.yaml)

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
```

## Security Features

### SQL Validation

- All SQL statements are parsed and validated
- Dangerous operations (DROP, TRUNCATE, etc.) are blocked
- UPDATE/DELETE without WHERE clause get auto LIMIT

### Audit Trail

All operations are logged with:

- User identity (from JWT)
- SQL statement (truncated if too long)
- Execution status
- Rows affected
- Duration

### Hot Reload

Configuration changes are automatically applied:

- Database connections
- Security rules
- Audit settings

## Prometheus Metrics

Access metrics at `/metrics` endpoint. Available metrics:

### HTTP Metrics
- `safemysql_http_requests_total` - Total HTTP requests
- `safemysql_http_request_duration_seconds` - Request duration histogram
- `safemysql_http_requests_active` - Active requests gauge

### Database Metrics
- `safemysql_db_queries_total` - Total database queries
- `safemysql_db_query_duration_seconds` - Query duration histogram
- `safemysql_db_query_rows` - Rows returned histogram
- `safemysql_db_connections_active` - Active connections
- `safemysql_db_connections_idle` - Idle connections

### Security Metrics
- `safemysql_security_violations_total` - Security violations
- `safemysql_security_blocked_queries_total` - Blocked queries
- `safemysql_security_sql_injection_attempts_total` - SQL injection attempts

### Rate Limit & Auth
- `safemysql_rate_limit_exceeded_total` - Rate limit exceeded
- `safemysql_auth_attempts_total` - Authentication attempts
- `safemysql_auth_failures_total` - Authentication failures

### MCP Metrics
- `safemysql_mcp_calls_total` - MCP tool calls
- `safemysql_mcp_call_duration_seconds` - MCP call duration
- `safemysql_mcp_errors_total` - MCP errors

## Development

### Prerequisites

- Go 1.22+
- MySQL 5.7+ (or use Docker)
- Docker (optional, for containerized development)

### Run Tests

```bash
# Run all tests with coverage
make test

# Run with race detection
go test ./... -race -cover

# Run specific package tests
go test ./internal/security/... -v
```

### Lint

```bash
make lint
# or
golangci-lint run
```

### Docker Development

```bash
# Build and run with MySQL
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

### CI/CD

The project includes GitHub Actions workflows for:

- **Build & Test**: Runs on Go 1.22, 1.23, 1.24
- **Lint**: golangci-lint with 25+ linters
- **Security Scan**: Gosec vulnerability scanner

## API Documentation

See [docs/openapi.yaml](docs/openapi.yaml) for the complete OpenAPI specification.

## Architecture

SafeMySQLMcpServer provides secure MySQL access through MCP protocol with SQL injection prevention and audit logging.

### Request Flow

```
Client → JWT Auth → Rate Limit → MCP Handler → SQL Validator → MySQL
                                                    ↓
                                              Audit Logger
```

### Modules

| Module | Description | Doc |
|--------|-------------|-----|
| cmd/server | Application entry point | [README](cmd/server/README.md) |
| internal/auth | JWT authentication and token generation | [README](internal/auth/README.md) |
| internal/config | Configuration loading and hot reload | [README](internal/config/README.md) |
| internal/database | MySQL connection pool and routing | [README](internal/database/README.md) |
| internal/mcp | MCP protocol implementation and tools | [README](internal/mcp/README.md) |
| internal/metrics | Prometheus metrics collection | [README](internal/metrics/README.md) |
| internal/security | SQL parsing, validation, and rewriting | [README](internal/security/README.md) |
| internal/server | HTTP server with middleware | [README](internal/server/README.md) |
| internal/audit | Audit logging with rotation | [README](internal/audit/README.md) |
| internal/validation | Input validation utilities | [README](internal/validation/README.md) |
| internal/constants | Shared constants | [README](internal/constants/README.md) |
| pkg/token | CLI tool for generating JWT tokens | [README](pkg/token/README.md) |

## License

MIT
