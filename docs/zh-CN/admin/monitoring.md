# 监控告警

SafeMySQLMcpServer 暴露 Prometheus 指标用于可观测性。

## 指标端点

```
GET /metrics
```

返回 Prometheus 文本格式的指标。

## HTTP 指标
tab: HTTP 指标

### safemysql_http_requests_total

**类型:** Counter

按 method、path 和 status 统计的 HTTP 请求总数。

```
safemysql_http_requests_total{method="POST",path="/mcp",status="200"} 1234
```

**标签:**
| 标签 | 值 | 说明 |
|-------|--------|-------------|
| method | GET, POST | HTTP 方法 |
| path | /mcp, /health, /metrics | 端点路径 |
| status | 200, 401, 500, ... | HTTP 状态码 |

### safemysql_http_request_duration_seconds

**类型:** Histogram

HTTP 请求持续时间（秒）。

```
safemysql_http_request_duration_seconds_bucket{method="POST",path="/mcp",le="0.1"} 1000
```

**标签:**
| 标签 | 值 | 说明 |
|-------|--------|-------------|
| method | GET, POST | HTTP 方法 |
| path | /mcp, /health, /metrics | 端点路径 |

**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### safemysql_http_requests_active

**类型:** Gauge

当前活跃的 HTTP 请求数。

```
safemysql_http_requests_active 5
```

## 数据库指标
tab: 数据库指标

### safemysql_db_queries_total

**类型:** Counter

按数据库和类型统计的查询总数。

```
safemysql_db_queries_total{database="mydb",type="SELECT"} 5678
```

**标签:**
| 标签 | 值 | 说明 |
|-------|--------|-------------|
| database | 数据库名 | 目标数据库 |
| type | SELECT, INSERT, UPDATE, DELETE | SQL 类型 |

### safemysql_db_query_duration_seconds

**类型:** Histogram

数据库查询持续时间（秒）。

```
safemysql_db_query_duration_seconds_bucket{database="mydb",le="0.01"} 4500
```

**Buckets:** 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10

### safemysql_db_connections_active

**类型:** Gauge

活跃的数据库连接数。

```
safemysql_db_connections_active{cluster="primary"} 15
```

## 安全指标
tab: 安全指标

### safemysql_security_blocked_queries_total

**类型:** Counter

按原因统计的被阻止查询数。

```
safemysql_security_blocked_queries_total{reason="DROP_NOT_ALLOWED"} 3
```

**标签:**
| 标签 | 值 | 说明 |
|-------|--------|-------------|
| reason | 见阻止原因 | 阻止原因 |

**阻止原因:**
| 原因 | 说明 |
|--------|-------------|
| DML_NOT_ALLOWED | DML 操作不在允许列表 |
| DDL_NOT_ALLOWED | DDL 操作不在允许列表 |
| OPERATION_BLOCKED | 操作在阻止列表中 |
| SQL_PARSE_ERROR | SQL 无法解析 |

### safemysql_security_sql_injection_attempts_total

**类型:** Counter

检测到的 SQL 注入尝试数。

```
safemysql_security_sql_injection_attempts_total 0
```

## MCP 指标
tab: MCP 指标

### safemysql_mcp_calls_total

**类型:** Counter

按工具名统计的 MCP 工具调用数。

```
safemysql_mcp_calls_total{tool="query"} 1234
```

**标签:**
| 标签 | 值 | 说明 |
|-------|--------|-------------|
| tool | query, list_databases, list_tables, ... | 工具名 |

### safemysql_mcp_call_duration_seconds

**类型:** Histogram

MCP 工具调用持续时间（秒）。

```
safemysql_mcp_call_duration_seconds_bucket{tool="query",le="0.1"} 1000
```

### safemysql_mcp_errors_total

**类型:** Counter

按工具统计的 MCP 错误数。

```
safemysql_mcp_errors_total{tool="query"} 5
```

## 告警规则
tab: 告警规则

### Prometheus 告警规则

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
          summary: "检测到高错误率"

      - alert: BlockedQueries
        expr: rate(safemysql_security_blocked_queries_total[5m]) > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SQL 查询被阻止"

      - alert: SQLInjectionAttempts
        expr: rate(safemysql_security_sql_injection_attempts_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "检测到 SQL 注入尝试"
```

## PromQL 查询示例
tab: PromQL 示例

### 请求速率

```promql
# 每秒请求数
rate(safemysql_http_requests_total[5m])
```

### 95 分位延迟

```promql
# 95 分位请求延迟
histogram_quantile(0.95, rate(safemysql_http_request_duration_seconds_bucket[5m]))
```

### 按查询数排名的数据库

```promql
# 查询数最多的前 5 个数据库
topk(5, sum by (database) (rate(safemysql_db_queries_total[5m])))
```

### 被阻止查询率

```promql
# 每秒被阻止的查询数
rate(safemysql_security_blocked_queries_total[5m])
```
