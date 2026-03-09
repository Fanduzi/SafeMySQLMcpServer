# Configuration Reference

Complete reference for SafeMySQLMcpServer configuration.

## Main Configuration (config.yaml)

### Structure

```yaml
server:
  # Server configuration
  host: 0.0.0.0
  port: 8080
  jwt_secret: ${JWT_SECRET}  # Environment variable

clusters:
  # MySQL cluster connections
  primary:
    host: mysql.example.com
    port: 3306
    username: ${DB_USER}
    password: ${DB_PASSWORD}

databases:
  # Database routing
  mydb:
    cluster: primary

security:
  # Security configuration file path
  config_file: config/security.yaml

audit:
  # Audit logging configuration
  enabled: true
  log_file: logs/audit.log

rate_limit:
  # Rate limiting configuration
  enabled: true
  requests_per_minute: 100
```

## Server Configuration
tab: Server Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| host | string | 0.0.0.0 | Server listen address |
| port | int | 8080 | Server listen port |
| jwt_secret | string | - | JWT signing secret (required) |
| log_level | string | info | Log level (debug, info, warn, error) |

### Environment Variables

| Variable | Used For |
|----------|----------|
| `${VAR_NAME}` | Substituted in configuration |
| `${JWT_SECRET}` | JWT signing secret |

## Clusters Configuration
tab: Clusters Configuration

### Structure

```yaml
clusters:
  primary:
    host: mysql.example.com
    port: 3306
    username: root
    password: secret
    max_open_conns: 50
    max_idle_conns: 25
    conn_max_lifetime: 5m
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| host | string | Required | MySQL host |
| port | int | 3306 | MySQL port |
| username | string | Required | MySQL username |
| password | string | "" | MySQL password |
| max_open_conns | int | 50 | Maximum open connections |
| max_idle_conns | int | 25 | Maximum idle connections |
| conn_max_lifetime | duration | 5m | Connection max lifetime |
| conn_max_idle_time | duration | 1m | Connection max idle time |

### Multiple Clusters

```yaml
clusters:
  dev-primary:
    host: dev-mysql.example.com
    port: 3306

  dev-replica:
    host: dev-mysql-replica.example.com
    port: 3306

  prod-primary:
    host: prod-mysql.example.com
    port: 3306
```

## Databases Configuration
tab: Databases Configuration

### Structure

```yaml
databases:
  user_db:
    cluster: primary
  order_db:
    cluster: primary
  analytics_db:
    cluster: replica
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| cluster | string | Required | Target cluster name |

### Routing

```
Request (database: user_db)
         │
         ▼
    databases.yaml
         │
         ▼
    cluster: primary
         │
         ▼
    clusters.yaml
         │
         ▼
    MySQL Connection
```

## Security Configuration
tab: Security Configuration

### Structure (security.yaml)

```yaml
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE

  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
    - ALTER_TABLE

  blocked:
    - DROP
    - TRUNCATE

  auto_limit: 1000
  max_limit: 10000
  query_timeout: 30s
  max_rows: 10000
  max_sql_length: 100000
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| allowed_dml | []string | [SELECT, INSERT, UPDATE, DELETE] | Allowed DML operations |
| allowed_ddl | []string | [CREATE_TABLE, CREATE_INDEX, ALTER_TABLE] | Allowed DDL operations |
| blocked | []string | [DROP, TRUNCATE] | Always blocked operations |
| auto_limit | int | 1000 | Auto LIMIT for unsafe operations |
| max_limit | int | 10000 | Maximum allowed LIMIT |
| query_timeout | duration | 30s | Query execution timeout |
| max_rows | int | 10000 | Maximum rows returned |
| max_sql_length | int | 100000 | Maximum SQL length |

### DML Values

| Value | Description |
|-------|-------------|
| SELECT | Query data |
| INSERT | Insert data |
| UPDATE | Update data |
| DELETE | Delete data |

### DDL Values

| Value | Description |
|-------|-------------|
| CREATE_TABLE | Create tables |
| CREATE_INDEX | Create indexes |
| ALTER_TABLE | Modify tables |
| DROP_TABLE | Drop tables |

## Audit Configuration
tab: Audit Configuration

### Structure

```yaml
audit:
  enabled: true
  log_file: logs/audit.log
  max_sql_length: 2000
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| enabled | bool | true | Enable audit logging |
| log_file | string | logs/audit.log | Log file path |
| max_sql_length | int | 2000 | Truncate SQL after N chars |
| max_size_mb | int | 100 | Rotate at N MB |
| max_backups | int | 10 | Keep N backup files |
| max_age_days | int | 30 | Delete backups after N days |
| compress | bool | true | Compress rotated logs |

## Rate Limit Configuration
tab: Rate Limit Configuration

### Structure

```yaml
rate_limit:
  enabled: true
  requests_per_minute: 100
  burst: 20
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| enabled | bool | true | Enable rate limiting |
| requests_per_minute | int | 100 | Max requests per minute per IP |
| burst | int | 20 | Burst allowance |

### Headers

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1646824600
```
