# Security Model

SafeMySQLMcpServer implements multiple security layers to protect against SQL injection and unauthorized access.

## Security Layers

```
┌─────────────────────────────────────────────────────────────────┐
│ Layer 1: Authentication (JWT)                                   │
│ - Verify JWT token                                              │
│ - Extract user identity                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 2: Input Validation                                       │
│ - Validate database names (regex: ^[a-zA-Z_][a-zA-Z0-9_]*$)     │
│ - Validate table names (regex: ^[a-zA-Z_][a-zA-Z0-9_]*$)        │
│ - Validate SQL length (max: 100,000 chars)                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 3: SQL Parsing                                            │
│ - Parse SQL into AST (Abstract Syntax Tree)                     │
│ - Identify SQL type (SELECT, INSERT, UPDATE, DELETE, DDL)       │
│ - Extract affected tables and columns                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 4: Security Checking                                      │
│ - Check against DML allowlist                                   │
│ - Check against DDL allowlist                                   │
│ - Check blocked operations list                                 │
│ - Detect dangerous patterns                                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 5: SQL Rewriting                                          │
│ - Add LIMIT to UPDATE/DELETE without WHERE                      │
│ - Add timeout to all queries                                    │
│ - Limit result set size                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Layer 6: Execution                                              │
│ - Execute with prepared statements                              │
│ - Enforce query timeout                                         │
│ - Limit returned rows                                           │
└─────────────────────────────────────────────────────────────────┘
```

## SQL Injection Prevention
tab: SQL Injection Prevention

### Identifier Validation
All identifiers (database names, table names) are validated with strict regex:

```go
// Only alphanumeric and underscore
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
```

### Prepared Statements
User input is never concatenated into SQL:

```go
// BAD: SQL injection vulnerable
query := fmt.Sprintf("SELECT * FROM %s", tableName)

// GOOD: Parameterized query
query := "SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = ?"
rows, _ := db.Query(query, tableName)
```

### LIKE Pattern Escaping
Search patterns are escaped to prevent LIKE injection:

```go
func EscapeLikePattern(pattern string) string {
    r := strings.NewReplacer(
        "%", "\\%",
        "_", "\\_",
        "\\", "\\\\",
    )
    return r.Replace(pattern)
}
```

### Quoted Identifiers
When dynamic identifiers are needed, they are properly quoted:

```go
func QuoteIdentifier(name string) string {
    return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
```

## Access Control
tab: Access Control

### JWT Authentication
All requests must include a valid JWT token:

```http
POST /mcp HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

### Database Routing
Users can only access databases configured in `databases` section:

```yaml
databases:
  user_db:
    cluster: dev-cluster-1
  order_db:
    cluster: dev-cluster-1
  # Users CANNOT access databases not listed here
```

### Operation Allowlist
Only explicitly allowed operations can be executed:

```yaml
security:
  allowed_dml: [SELECT]  # Only SELECT allowed
  allowed_ddl: []    # No DDL allowed
```

## Rate Limiting
tab: Rate Limiting

### IP-Based Rate Limiting
Each IP address has a request limit:

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100  # Max 100 requests per minute per IP
  burst: 20              # Allow bursts up to 20 requests
```

### Rate Limit Headers
Responses include rate limit information:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1646824600
```

## Audit Trail
tab: Audit Trail

### Logged Information
Every SQL operation is logged:

| Field | Description | Example |
|-------|-------------|---------|
| timestamp | When the operation occurred | 2026-03-09T10:30:00Z |
| user_id | User from JWT | admin |
| user_email | Email from JWT | admin@example.com |
| database | Target database | mydb |
| sql | SQL statement (truncated) | SELECT * FROM users... |
| sql_type | Type of SQL | SELECT |
| status | Result status | success, error, blocked |
| rows_affected | Rows affected | 10 |
| duration_ms | Execution time | 15 |

### Log Format
JSON format for easy parsing:

```json
{
  "timestamp": "2026-03-09T10:30:00.123Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "database": "mydb",
  "sql": "SELECT * FROM users LIMIT 10",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 10,
  "duration_ms": 15,
  "client_ip": "192.168.1.100"
}
```
