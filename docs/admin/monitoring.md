# Monitoring

SafeMySQLMcpServer exposes Prometheus metrics for observability.

## Metrics Endpoint

```
GET /metrics
```

Returns Prometheus-formatted metrics.

## Available Metrics
tab: Available Metrics

### HTTP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_http_requests_total` | Counter | Total HTTP requests by method, path, status |
| `safemysql_http_request_duration_seconds` | Histogram | Request duration by method, path |
| `safemysql_http_requests_active` | Gauge | Currently active requests |

### Database Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_db_queries_total` | Counter | Total database queries by database, type |
| `safemysql_db_query_duration_seconds` | Histogram | Query duration by database |
| `safemysql_db_query_rows` | Histogram | Rows returned by query |
| `safemysql_db_connections_active` | Gauge | Active database connections |
| `safemysql_db_connections_idle` | Gauge | Idle database connections |

### Security Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_security_violations_total` | Counter | Security rule violations |
| `safemysql_security_blocked_queries_total` | Counter | Blocked queries by reason |
| `safemysql_security_sql_injection_attempts_total` | Counter | SQL injection attempts detected |

### Rate Limit Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_rate_limit_exceeded_total` | Counter | Rate limit exceeded by IP |

### Authentication Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_auth_attempts_total` | Counter | Authentication attempts by type |
| `safemysql_auth_failures_total` | Counter | Authentication failures by type |

### MCP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `safemysql_mcp_calls_total` | Counter | MCP tool calls by tool name |
| `safemysql_mcp_call_duration_seconds` | Histogram | MCP call duration by tool |
| `safemysql_mcp_errors_total` | Counter | MCP errors by tool |

## Prometheus Configuration
tab: Prometheus Configuration

### prometheus.yml

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'safemysql'
    static_configs:
      - targets: ['safemysql:8080']
    metrics_path: /metrics
```

### Example Queries

```promql
# Request rate
rate(safemysql_http_requests_total[5m])

# Average request duration
rate(safemysql_http_request_duration_seconds_sum[5m]) / rate(safemysql_http_request_duration_seconds_count[5m])

# Error rate
rate(safemysql_http_requests_total{status=~"5.."}[5m])

# Database query rate by type
rate(safemysql_db_queries_total[5m]) by (type)

# Blocked queries rate
rate(safemysql_security_blocked_queries_total[5m])
```

## Grafana Dashboard
tab: Grafana Dashboard

### Example Dashboard JSON

```json
{
  "dashboard": {
    "title": "SafeMySQLMcpServer",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [{
          "expr": "rate(safemysql_http_requests_total[5m])"
        }]
      },
      {
        "title": "Query Duration",
        "type": "graph",
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(safemysql_db_query_duration_seconds_bucket[5m]))"
        }]
      },
      {
        "title": "Blocked Queries",
        "type": "graph",
        "targets": [{
          "expr": "rate(safemysql_security_blocked_queries_total[5m])"
        }]
      }
    ]
  }
}
```

## Alerting Rules
tab: Alerting Rules

### Prometheus Alert Rules

```yaml
groups:
  - name: safemysql
    rules:
      - alert: HighErrorRate
        expr: rate(safemysql_http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"

      - alert: BlockedQueries
        expr: rate(safemysql_security_blocked_queries_total[5m]) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SQL queries are being blocked"

      - alert: SQLInjectionAttempts
        expr: rate(safemysql_security_sql_injection_attempts_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SQL injection attempts detected"
```
