# Metrics Reference

Complete reference for Prometheus metrics exposed by SafeMySQLMcpServer.

## Endpoint

```
GET /metrics
```

Returns Prometheus text-format metrics.

## HTTP Metrics
tab: HTTP Metrics

### safemysql_http_requests_total

**Type:** Counter

Total HTTP requests by method, path, and status.

```
safemysql_http_requests_total{method="POST",path="/mcp",status="200"} 1234
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| method | GET, POST | HTTP method |
| path | /mcp, /health, /metrics | Endpoint path |
| status | 200, 401, 500, ... | HTTP status code |

### safemysql_http_request_duration_seconds

**Type:** Histogram

HTTP request duration in seconds.

```
safemysql_http_request_duration_seconds_bucket{method="POST",path="/mcp",le="0.1"} 1000
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| method | GET, POST | HTTP method |
| path | /mcp, /health, /metrics | Endpoint path |

**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### safemysql_http_requests_active

**Type:** Gauge

Currently active HTTP requests.

```
safemysql_http_requests_active 5
```

## Database Metrics
tab: Database Metrics

### safemysql_db_queries_total

**Type:** Counter

Total database queries by database and type.

```
safemysql_db_queries_total{database="mydb",type="SELECT"} 5678
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| database | Database name | Target database |
| type | SELECT, INSERT, UPDATE, DELETE | SQL type |

### safemysql_db_query_duration_seconds

**Type:** Histogram

Database query duration in seconds.

```
safemysql_db_query_duration_seconds_bucket{database="mydb",le="0.01"} 4500
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| database | Database name | Target database |

**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### safemysql_db_query_rows

**Type:** Histogram

Rows returned by queries.

```
safemysql_db_query_rows_bucket{database="mydb",le="100"} 3000
```

**Buckets:** 1, 10, 100, 1000, 10000

### safemysql_db_connections_active

**Type:** Gauge

Active database connections.

```
safemysql_db_connections_active{cluster="primary"} 15
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| cluster | Cluster name | Database cluster |

### safemysql_db_connections_idle

**Type:** Gauge

Idle database connections.

```
safemysql_db_connections_idle{cluster="primary"} 10
```

## Security Metrics
tab: Security Metrics

### safemysql_security_violations_total

**Type:** Counter

Security rule violations.

```
safemysql_security_violations_total{type="blocked"} 5
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| type | blocked, rate_limit | Violation type |

### safemysql_security_blocked_queries_total

**Type:** Counter

Blocked queries by reason.

```
safemysql_security_blocked_queries_total{reason="DROP_NOT_ALLOWED"} 3
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| reason | See Blocked Reasons | Block reason |

**Blocked Reasons:**
| Reason | Description |
|--------|-------------|
| DML_NOT_ALLOWED | DML operation not in allowlist |
| DDL_NOT_ALLOWED | DDL operation not in allowlist |
| OPERATION_BLOCKED | Operation in blocked list |
| AUTO_LIMIT | Auto-LIMIT applied |
| SQL_PARSE_ERROR | SQL could not be parsed |

### safemysql_security_sql_injection_attempts_total

**Type:** Counter

SQL injection attempts detected.

```
safemysql_security_sql_injection_attempts_total 0
```

## MCP Metrics
tab: MCP Metrics

### safemysql_mcp_calls_total

**Type:** Counter

MCP tool calls by tool name.

```
safemysql_mcp_calls_total{tool="query"} 1234
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| tool | query, list_databases, list_tables, ... | Tool name |

### safemysql_mcp_call_duration_seconds

**Type:** Histogram

MCP tool call duration in seconds.

```
safemysql_mcp_call_duration_seconds_bucket{tool="query",le="0.1"} 1000
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| tool | query, list_databases, ... | Tool name |

**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### safemysql_mcp_errors_total

**Type:** Counter

MCP errors by tool.

```
safemysql_mcp_errors_total{tool="query"} 5
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| tool | query, list_databases, ... | Tool name |

## Rate Limit Metrics
tab: Rate Limit Metrics

### safemysql_rate_limit_exceeded_total

**Type:** Counter

Rate limit exceeded by IP.

```
safemysql_rate_limit_exceeded_total 10
```

## Authentication Metrics
tab: Authentication Metrics

### safemysql_auth_attempts_total

**Type:** Counter

Authentication attempts by type.

```
safemysql_auth_attempts_total{type="jwt",success="true"} 1234
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| type | jwt | Auth type |
| success | true, false | Success status |

### safemysql_auth_failures_total

**Type:** Counter

Authentication failures by type.

```
safemysql_auth_failures_total{type="jwt"} 5
```

**Labels:**
| Label | Values | Description |
|-------|--------|-------------|
| type | jwt, expired, invalid | Failure type |

## Example PromQL Queries
tab: Example PromQL Queries

### Request Rate

```promql
# Requests per second
rate(safemysql_http_requests_total[5m])
```

### Error Rate

```promql
# Error rate percentage
sum(rate(safemysql_http_requests_total{status=~"5.."}[5m]))
/
sum(rate(safemysql_http_requests_total[5m]))
* 100
```

### 95th Percentile Latency

```promql
# 95th percentile request duration
histogram_quantile(0.95, rate(safemysql_http_request_duration_seconds_bucket[5m]))
```

### Top Databases by Query Count

```promql
# Top 5 databases by query count
topk(5, sum by (database) (rate(safemysql_db_queries_total[5m])))
```

### Blocked Query Rate

```promql
# Blocked queries per second
rate(safemysql_security_blocked_queries_total[5m])
```
