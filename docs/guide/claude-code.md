# Integration with Claude Code

This guide explains how to connect Claude Code to your MySQL database via SafeMySQLMcpServer.

## Prerequisites

- Claude Code installed
- SafeMySQLMcpServer running and accessible
- Valid JWT token generated

## Configuration
tab: Configuration

### 1. Add MCP Server to Claude Code

Edit your Claude Code settings file:

#### macOS/Linux
```bash
vim ~/.claude/settings.json
```

#### Windows
```bash
notepad %USERPROFILE%\.claude\settings.json
```

### 2. Add Server Configuration

```json
{
  "mcpServers": {
    "mysql-dev": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_JWT_TOKEN_HERE"
      }
    }
  }
}
```

Replace `YOUR_JWT_TOKEN_HERE` with the token generated in the [Authentication](./authentication.md) guide.

### 3. Restart Claude Code

Restart Claude Code for the configuration to take effect.

## Available Tools
tab: Available Tools

Once connected, Claude Code can use the following MCP tools:

| Tool | Description | Example Usage |
|------|-------------|----------------|
| `query` | Execute SQL query | "Query the users table" |
| `list_databases` | List all databases | "Show me available databases" |
| `list_tables` | List tables in database | "List tables in mydb" |
| `describe_table` | Get table structure | "Describe the users table in mydb" |
| `show_create_table` | Get CREATE TABLE | "Show CREATE statement for users table" |
| `explain` | Get query execution plan | "Explain SELECT * FROM users" |
| `search_tables` | Search tables by name | "Find tables containing 'user'" |

## Example Usage
tab: Example Usage

### Listing Databases

```
User: List all available databases
Claude: I'll use the list_databases tool.
[Results: mysql, information_schema, performance_schema]
```

### Querying Data

```
User: Query the first 10 users from mydb
Claude: I'll execute a SELECT query on the mydb database.
[Results: 10 users with their details]
```

### Understanding Table Structure

```
User: What columns does the orders table have?
Claude: Let me describe the orders table in mydb.
[Results: Table has id, user_id, total, status, created_at columns]
```

### Analyzing Query Performance

```
User: Is this query using an index: SELECT * FROM orders WHERE user_id = 123
Claude: Let me run EXPLAIN on that query.
[Results: Query uses idx_user_id index, type=ref]
```

## Security Considerations
tab: Security Considerations

### What Claude Code CAN Do
- Execute allowed DML operations (SELECT, INSERT, UPDATE, DELETE)
- Execute allowed DDL operations (CREATE_TABLE, CREATE_INDEX, ALTER_TABLE)
- View table structures and metadata

### What Claude Code CANNOT Do
- Execute blocked operations (DROP, TRUNCATE, GRANT, REVOKE)
- Perform operations not in the allowlist
- Execute UPDATE/DELETE without WHERE (auto-limited)
- Access databases not configured in the routing

### Audit Trail
All operations performed by Claude Code are logged in the audit log:
```json
{
  "timestamp": "2026-03-09T10:30:00Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "database": "mydb",
  "sql": "SELECT * FROM users LIMIT 10",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 10,
  "duration_ms": 15
}
```

## Troubleshooting
tab: Troubleshooting

### Connection Refused

```
Error: Connection refused
```

**Solution**: Check if SafeMySQLMcpServer is running:
```bash
curl http://localhost:8080/health
```

### Authentication Failed

```
Error: invalid token
```

**Solution**: Generate a new token
```bash
./bin/mysql-mcp-token --user admin --email admin@example.com
```

### SQL Blocked

```
Error: SQL blocked: operation not allowed
```

**Solution**: Update security.yaml to allow the operation

### Query Timeout

```
Error: context deadline exceeded
```

**Solution**: Optimize the query or increase query_timeout in security.yaml
