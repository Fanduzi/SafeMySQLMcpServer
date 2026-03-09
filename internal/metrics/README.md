# Metrics Module

Prometheus metrics collection and HTTP endpoint for observability.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Prometheus Metrics                        │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                  Metric Types                            ││
│  │                                                          ││
│  │   Counter    → Cumulative count (requests_total)        ││
│  │   Gauge      → Point-in-time value (connections_active) ││
│  │   Histogram  → Distribution (request_duration)          ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   Categories                             ││
│  │                                                          ││
│  │   HTTP      DB       Security    MCP      Auth          ││
│  │   ─────     ───      ────────    ───      ────          ││
│  │   requests  queries  violations  calls    attempts      ││
│  │   duration  duration blocked     duration failures      ││
│  │   active    rows     injection   errors                 ││
│  │            connections                                    ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   /metrics                               ││
│  │              Prometheus Text Format                      ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| metrics.go | Prometheus metrics definitions | ~150 |
| metrics_test.go | Unit tests | ~60 |

## Test Coverage
```
Coverage: ~80%
- Metric registration
- Counter increment
- Histogram observation
- HTTP handler
```

## Exports

### Metrics Container
```go
type Metrics struct {
    // HTTP metrics
    HTTPRequestsTotal   *prometheus.CounterVec
    HTTPRequestDuration *prometheus.HistogramVec
    HTTPRequestsActive  prometheus.Gauge

    // Database metrics
    DBQueriesTotal        *prometheus.CounterVec
    DBQueryDuration       *prometheus.HistogramVec
    DBQueryRows           *prometheus.HistogramVec
    DBConnectionsActive   *prometheus.GaugeVec
    DBConnectionsIdle      *prometheus.GaugeVec

    // Security metrics
    SecurityViolationsTotal       *prometheus.CounterVec
    SecurityBlockedQueriesTotal   *prometheus.CounterVec
    SecuritySQLInjectionAttempts  prometheus.Counter

    // MCP metrics
    MCPCallsTotal     *prometheus.CounterVec
    MCPCallDuration   *prometheus.HistogramVec
    MCPErrorsTotal    *prometheus.CounterVec

    // Auth metrics
    AuthAttemptsTotal  *prometheus.CounterVec
    AuthFailuresTotal  *prometheus.CounterVec

    // Rate limit metrics
    RateLimitExceededTotal prometheus.Counter
}

func Init(namespace string) *Metrics
func Get() *Metrics
func Handler() http.Handler
```

## Metric Categories
| Category | Prefix | Metrics |
|----------|--------|---------|
| HTTP | `safemysql_http_` | requests_total, request_duration, requests_active |
| DB | `safemysql_db_` | queries_total, query_duration, query_rows, connections_* |
| Security | `safemysql_security_` | violations_total, blocked_queries_total, sql_injection_attempts |
| Auth | `safemysql_auth_` | attempts_total, failures_total |
| MCP | `safemysql_mcp_` | calls_total, call_duration, errors_total |
| Rate Limit | `safemysql_rate_limit_` | exceeded_total |

## Histogram Buckets
```go
// Duration buckets (seconds)
DurationBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// Row count buckets
RowBuckets = []float64{1, 10, 100, 1000, 10000}
```

## Example Metrics
```
# HTTP
safemysql_http_requests_total{method="POST",path="/mcp",status="200"} 1234
safemysql_http_request_duration_seconds_bucket{method="POST",path="/mcp",le="0.1"} 1000

# Database
safemysql_db_queries_total{database="mydb",type="SELECT"} 5678
safemysql_db_connections_active{cluster="primary"} 15

# Security
safemysql_security_blocked_queries_total{reason="DROP_NOT_ALLOWED"} 3

# MCP
safemysql_mcp_calls_total{tool="query"} 1234
```

## Dependencies
```
Upstream: None

Downstream:
  └── internal/server  → Metrics middleware

External:
  └── github.com/prometheus/client_golang  → Prometheus library
```

## Update Rule
If metrics change, update:
1. This file
2. metrics.go
3. docs/reference/metrics.md
