# Troubleshooting

Common issues and their solutions.

## Connection Issues
tab: Connection Issues

### MySQL Connection Refused

**Symptoms:**
```
Error: dial tcp 127.0.0.1:3306: connect: connection refused
```

**Solutions:**

| Check | Command |
|-------|---------|
| MySQL running? | `docker ps` or `systemctl status mysql` |
| Correct host/port? | Check `config.yaml` clusters section |
| Network accessible? | `telnet mysql-host 3306` |
| Firewall blocking? | Check iptables/firewall rules |

### Connection Pool Exhausted

**Symptoms:**
```
Error: connection pool exhausted
```

**Solutions:**

```yaml
# Increase pool size in config.yaml
clusters:
  primary:
    max_open_conns: 50  # Increase from default
    max_idle_conns: 25
```

### Authentication Failed

**Symptoms:**
```
Error: invalid token: token has expired
```

**Solutions:**

```bash
# Generate new token with longer expiration
./bin/mysql-mcp-token --user admin --email admin@example.com --expire 365d
```

## SQL Issues
tab: SQL Issues

### Query Blocked

**Symptoms:**
```
Error: SQL blocked: operation not allowed
```

**Solutions:**

1. Check `security.yaml` for allowed operations
2. Add the operation to allowlist if appropriate

```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT  # Add if INSERT was blocked
```

### Query Timeout

**Symptoms:**
```
Error: context deadline exceeded
```

**Solutions:**

```yaml
# Increase timeout in security.yaml
security:
  query_timeout: 60s  # Increase from 30s
```

### SQL Parse Error

**Symptoms:**
```
Error: SQL parse error: syntax error at position X
```

**Solutions:**

1. Check SQL syntax
2. Ensure SQL is valid MySQL syntax
3. Check for unsupported MySQL features

### Max Rows Exceeded

**Symptoms:**
```
Error: result set too large
```

**Solutions:**

```yaml
# Increase max rows
security:
  max_rows: 50000  # Increase from 10000
```

Or add LIMIT to query.

## Performance Issues
tab: Performance Issues

### Slow Queries

**Symptoms:**
- Queries taking >5 seconds
- High CPU usage

**Solutions:**

1. Analyze slow queries from audit logs:

```bash
jq 'select(.duration_ms > 5000)' logs/audit.log
```

2. Add indexes to frequently queried columns

3. Use EXPLAIN to analyze:

```bash
# Via MCP tool
"Explain this query: SELECT * FROM users WHERE email = 'test@example.com'"
```

### High Memory Usage

**Symptoms:**
- OOM errors
- Container restarts

**Solutions:**

1. Reduce max_rows:

```yaml
security:
  max_rows: 5000  # Reduce from 10000
```

2. Reduce connection pool:

```yaml
clusters:
  primary:
    max_open_conns: 20  # Reduce from 50
```

3. Add memory limits (Docker):

```yaml
services:
  safemysql:
    deploy:
      resources:
        limits:
          memory: 256M
```

## Configuration Issues
tab: Configuration Issues

### Config File Not Found

**Symptoms:**
```
Error: open config/config.yaml: no such file or directory
```

**Solutions:**

```bash
# Check file exists
ls -la config/config.yaml

# Check path in command
./bin/safe-mysql-mcp -config config/config.yaml
```

### Invalid YAML

**Symptoms:**
```
Error: yaml: line 15: could not find expected ':'
```

**Solutions:**

```bash
# Validate YAML
python -c "import yaml; yaml.safe_load(open('config/config.yaml'))"

# Or use yamllint
yamllint config/config.yaml
```

### Environment Variable Not Set

**Symptoms:**
```
Error: JWT secret not configured
```

**Solutions:**

```bash
# Set environment variable
export JWT_SECRET="your-secret-key-min-32-characters-long"

# Or in docker-compose
environment:
  JWT_SECRET: ${JWT_SECRET}
```

## Debug Mode
tab: Debug Mode

### Enable Debug Logging

```yaml
# config.yaml
server:
  log_level: debug
```

### Check Logs

```bash
# Docker logs
docker-compose logs -f safemysql

# Systemd logs
journalctl -u safemysql -f

# Direct logs
tail -f /var/log/safemysql/server.log
```

### Common Log Patterns

| Log Message | Issue |
|-------------|-------|
| "connection refused" | MySQL not running |
| "invalid token" | JWT authentication failed |
| "SQL blocked" | Security rule violation |
| "context deadline" | Query timeout |
| "pool exhausted" | Too many connections |
