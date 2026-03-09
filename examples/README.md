# Examples

This directory contains example configurations to help you get started with SafeMySQLMcpServer.

## Contents

| File | Description |
|------|-------------|
| [config/config.yaml](config/config.yaml) | Example main configuration |
| [config/security.yaml](config/security.yaml) | Example security rules |
| [config/.env.example](config/.env.example) | Environment variables template |

## Quick Start

1. Copy the example files to your config directory:
   ```bash
   cp examples/config/config.yaml config/
   cp examples/config/security.yaml config/
   cp examples/config/.env.example .env
   ```

2. Edit `.env` and set your secrets:
   ```bash
   # Required
   JWT_SECRET=your-secret-key-min-32-characters-long

   # Database credentials
   DEV_DB_USER=your_db_user
   DEV_DB_PASSWORD=your_db_password
   ```

3. Update `config/config.yaml` with your database hosts and cluster names.

4. Generate a token:
   ```bash
   ./bin/mysql-mcp-token --user admin --email admin@example.com
   ```

5. Start the server:
   ```bash
   ./bin/safe-mysql-mcp -config config/config.yaml
   ```

## Configuration Reference

### Main Configuration (config.yaml)

| Section | Description |
|---------|-------------|
| `server` | HTTP server settings (host, port, JWT secret) |
| `clusters` | MySQL cluster connections |
| `databases` | Database to cluster routing |
| `security` | Security config file path |
| `audit` | Audit logging settings |
| `rate_limit` | Rate limiting configuration |

### Security Configuration (security.yaml)

| Setting | Description | Example |
|---------|-------------|---------|
| `allowed_dml` | Allowed DML operations | SELECT, INSERT, UPDATE, DELETE |
| `allowed_ddl` | Allowed DDL operations | CREATE_TABLE, CREATE_INDEX |
| `blocked` | Always blocked operations | DROP, TRUNCATE |
| `auto_limit` | Auto LIMIT for unsafe operations | 1000 |
| `query_timeout` | Query timeout | 30s |
| `max_rows` | Maximum rows returned | 10000 |

## Docker Example

```bash
# Start with docker-compose (uses examples/config by default)
docker-compose up -d

# Check health
curl http://localhost:8080/health

# Generate token
docker exec safemysql-app /app/token -user admin -email admin@example.com
```
