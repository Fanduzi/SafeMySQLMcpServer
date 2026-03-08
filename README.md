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

## Quick Start

### 1. Build

```bash
make build
```

### 2. Configure

Copy and edit configuration files:

```bash
cp config/config.yaml.example config/config.yaml
cp config/security.yaml.example config/security.yaml
```

Set environment variables:

```bash
export JWT_SECRET=your-secret-key
export DEV_DB_USER=your-db-user
export DEV_DB_PASSWORD=your-db-password
```

### 3. Generate Token

```bash
./bin/mysql-mcp-token --user zhangsan --email zhangsan@company.com --expire 365d
```

### 4. Run Server

```bash
./bin/safe-mysql-mcp -config config/config.yaml
```

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

## Development

### Prerequisites

- Go 1.21+
- MySQL 5.7+

### Run Tests

```bash
make test
```

### Lint

```bash
make lint
```

## License

MIT
