# Audit Logging

SafeMySQLMcpServer logs all SQL operations for compliance and debugging.

## Log Format
tab: Log Format

### JSON Structure

```json
{
  "timestamp": "2026-03-09T10:30:00.123Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "client_ip": "192.168.1.100",
  "database": "mydb",
  "sql": "SELECT * FROM users WHERE id = 123",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 1,
  "duration_ms": 15,
  "block_reason": ""
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| timestamp | string | ISO 8601 timestamp |
| user_id | string | User identifier from JWT |
| user_email | string | User email from JWT |
| client_ip | string | Client IP address |
| database | string | Target database name |
| sql | string | SQL statement (truncated if too long) |
| sql_type | string | SQL type (SELECT, INSERT, etc.) |
| status | string | success, error, blocked |
| rows_affected | number | Rows affected/returned |
| duration_ms | number | Execution duration in ms |
| block_reason | string | Reason if blocked |

## Log Rotation
tab: Log Rotation

### Configuration

```yaml
audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000      # Truncate SQL after 2000 chars
  max_size_mb: 100          # Rotate at 100MB
  max_backups: 10           # Keep 10 backup files
  max_age_days: 30          # Delete backups after 30 days
  compress: true            # Compress backups with gzip
```

### Rotation Behavior

1. When log reaches `max_size_mb`, it is rotated
2. Old logs are renamed with timestamp: `audit-2026-03-09T10.30.00.log`
3. If `compress: true`, rotated logs are gzipped
4. Old backups beyond `max_backups` or `max_age_days` are deleted

### Manual Rotation

```bash
# Force rotation (send USR1 signal)
kill -USR1 <pid>

# Or using logrotate
logrotate -f /etc/logrotate.d/safemysql
```

## Log Analysis
tab: Log Analysis

### Query by User

```bash
# Find all queries by user
jq 'select(.user_id == "admin")' logs/audit.log
```

### Find Blocked Queries

```bash
# Find all blocked queries
jq 'select(.status == "blocked")' logs/audit.log
```

### Find Slow Queries

```bash
# Find queries taking >1 second
jq 'select(.duration_ms > 1000)' logs/audit.log
```

### Query by Time Range

```bash
# Queries in specific time range
jq 'select(.timestamp >= "2026-03-09T10:00:00" and .timestamp < "2026-03-09T11:00:00")' logs/audit.log
```

### Statistics

```bash
# Query count by user
jq -r '.user_id' logs/audit.log | sort | uniq -c | sort -rn

# Average query duration by database
jq -s 'group_by(.database) | .[] | {database: .[0].database, avg_duration: (map(.duration_ms) | add) / length}'' logs/audit.log
```

## Compliance
tab: Compliance

### Data Retention

Configure retention based on compliance requirements:

| Regulation | Recommended Retention |
|------------|----------------------|
| SOX | 7 years |
| GDPR | As needed, minimize data |
| HIPAA | 6 years |
| PCI-DSS | 1 year |

### PII Considerations

The audit log may contain PII:

| Field | PII | Recommendation |
|-------|-----|----------------|
| user_email | Yes | Consider hashing |
| client_ip | Maybe | Consider masking |
| sql | Maybe | Review for sensitive data |

### Log Security

```bash
# Set restrictive permissions
chmod 640 logs/audit.log
chown safemysql:safemysql logs/audit.log

# Ensure log directory is secure
chmod 750 logs/
```

## Log Shipping
tab: Log Shipping

### Filebeat Configuration

```yaml
# /etc/filebeat/filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /opt/safemysql/logs/audit.log
    json.keys_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "safemysql-audit"
```

### Fluentd Configuration

```xml
<!-- /etc/fluentd/fluent.conf -->
<source>
  @type tail
  path /opt/safemysql/logs/audit.log
  tag safemysql.audit
  <parse>
    @type json
  </parse>
</source>

<match safemysql.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
</match>
```
