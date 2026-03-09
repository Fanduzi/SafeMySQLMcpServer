# Metrics Module

Prometheus metrics collection and HTTP endpoint.

## Files
| File | Responsibility |
|------|---------------|
| metrics.go | Prometheus metrics definitions and helpers |
| metrics_test.go | Unit tests |

## Exports
- `Metrics` - Metrics container
- `Init(namespace string) *Metrics` - Initialize metrics
- `Get() *Metrics` - Get global instance
- `Handler() http.Handler` - Prometheus HTTP handler

## Metric Categories
| Category | Prefix | Metrics |
|----------|--------|---------|
| HTTP | `safemysql_http_` | requests_total, request_duration, requests_active |
| DB | `safemysql_db_` | queries_total, query_duration, query_rows, connections_* |
| Security | `safemysql_security_` | violations_total, blocked_queries_total |
| Auth | `safemysql_auth_` | attempts_total, failures_total |
| MCP | `safemysql_mcp_` | calls_total, call_duration, errors_total |
| Rate Limit | `safemysql_rate_limit_` | exceeded_total |

## Dependencies
- Upstream: None
- Downstream: `internal/server` - Uses metrics middleware

## Update Rule
If metrics change, update this file in the same change.
