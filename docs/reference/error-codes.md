# Error Codes Reference

This document lists all error codes and messages returned by SafeMySQLMcpServer.

## HTTP Status Codes

| Code | Name | Description |
|------|------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request format |
| 401 | Unauthorized | Missing or invalid token |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |

## Authentication Errors
tab: Authentication Errors

### Missing Token

**HTTP Status:** 401

```json
{
  "error": "missing authorization token"
}
```

**Cause:** No Authorization header provided.

**Solution:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" ...
```

### Invalid Token

**HTTP Status:** 401

```json
{
  "error": "invalid token: signature is invalid"
}
```

**Cause:** Token is not signed with the correct secret.

**Solution:** Generate new token with correct JWT secret.

### Expired Token

**HTTP Status:** 401

```json
{
  "error": "invalid token: token has expired"
}
```

**Cause:** Token has passed expiration time.

**Solution:** Generate new token.

## Validation Errors
tab: Validation Errors

### Invalid Database Name

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Validation error: invalid database name"}],
  "isError": true
}
```

**Cause:** Database name contains invalid characters.

**Valid Format:** `^[a-zA-Z_][a-zA-Z0-9_]*$`

| Input | Valid |
|-------|-------|
| mydb | ✅ |
| my_db | ✅ |
| MyDb123 | ✅ |
| my-db | ❌ |
| my db | ❌ |
| 123db | ❌ |

### Invalid Table Name

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Validation error: invalid table name"}],
  "isError": true
}
```

**Cause:** Table name contains invalid characters.

**Valid Format:** `^[a-zA-Z_][a-zA-Z0-9_]*$`

### Empty SQL

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Validation error: SQL statement cannot be empty"}],
  "isError": true
}
```

**Cause:** SQL parameter is empty or whitespace only.

### SQL Too Long

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Validation error: SQL statement exceeds maximum length"}],
  "isError": true
}
```

**Cause:** SQL exceeds `max_sql_length` (default: 100,000 chars).

## Security Errors
tab: Security Errors

### Operation Not Allowed

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "SQL blocked: INSERT operation not allowed"}],
  "isError": true
}
```

**Cause:** Operation not in `allowed_dml` or `allowed_ddl`.

**Solution:** Update `security.yaml` to allow operation.

### Blocked Operation

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "SQL blocked: DROP operation is blocked"}],
  "isError": true
}
```

**Cause:** Operation in `blocked` list.

**Solution:** Remove from blocked list (not recommended for production).

### SQL Parse Error

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "SQL parse error: syntax error at position 10"}],
  "isError": true
}
```

**Cause:** SQL syntax is invalid.

**Solution:** Fix SQL syntax.

## Database Errors
tab: Database Errors

### Unknown Database

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Error: unknown database: mydb"}],
  "isError": true
}
```

**Cause:** Database not configured in routing.

**Solution:** Add database to `databases` section in config.

### Connection Failed

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Error: connection refused"}],
  "isError": true
}
```

**Cause:** MySQL server not reachable.

**Solution:** Check MySQL is running and network is accessible.

### Query Timeout

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Error: context deadline exceeded"}],
  "isError": true
}
```

**Cause:** Query execution exceeded `query_timeout`.

**Solution:** Optimize query or increase timeout.

### Result Too Large

**MCP Error:**

```json
{
  "content": [{"type": "text", "text": "Error: result set too large"}],
  "isError": true
}
```

**Cause:** Query returned more than `max_rows`.

**Solution:** Add LIMIT to query or increase `max_rows`.

## Rate Limit Errors
tab: Rate Limit Errors

### Rate Limit Exceeded

**HTTP Status:** 429

```json
{
  "error": "rate limit exceeded"
}
```

**Cause:** Too many requests from same IP.

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1646824600
Retry-After: 30
```

**Solution:** Wait for rate limit reset or increase limit.

## Error Handling Best Practices
tab: Best Practices

### For Developers

1. **Always check isError field**:
```json
{
  "isError": true,
  "content": [{"type": "text", "text": "Error message"}]
}
```

2. **Parse error messages**:
```go
if strings.Contains(errMsg, "blocked") {
    // Security blocked
}
```

3. **Handle rate limits**:
```go
if resp.StatusCode == 429 {
    retryAfter := resp.Header.Get("Retry-After")
    time.Sleep(time.Duration(retryAfter) * time.Second)
}
```

### For Operators

1. **Check audit logs** for error details
2. **Monitor** `safemysql_security_blocked_queries_total` for blocked queries
3. **Alert** on high error rates
