# Database Module

MySQL connection pool management and database routing.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Database Layer                           │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                      Router                              ││
│  │   database → cluster mapping                            ││
│  │                                                          ││
│  │   user_db     → primary                                  ││
│  │   order_db    → primary                                  ││
│  │   analytics_db → replica                                 ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                       Pool                               ││
│  │   cluster → *sql.DB (connection pool)                   ││
│  │                                                          ││
│  │   primary:  50 open, 25 idle                            ││
│  │   replica:  50 open, 25 idle                            ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                   MySQL Clusters                         ││
│  │   primary:3306    replica:3306                          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| pool.go | Connection pool with reference counting | ~120 |
| router.go | Database-to-cluster routing | ~80 |
| doc.go | Package documentation | ~20 |
| pool_test.go | Pool unit tests | ~100 |
| router_test.go | Router unit tests | ~80 |
| integration_test.go | Integration tests (requires MySQL) | ~150 |

## Test Coverage
```
Unit Tests: ~85%
- Pool creation and connection management
- Router database-to-cluster mapping
- Error handling for unknown databases

Integration Tests: (requires MySQL)
- Actual query execution
- Connection pool behavior
- Multi-cluster routing
```

## Exports

### Pool
```go
type Pool struct {
    mu      sync.RWMutex
    conns   map[string]*sql.DB  // cluster → connection
    config  config.ClustersConfig
}

func NewPool(clusters config.ClustersConfig) (*Pool, error)
func (p *Pool) Get(cluster string) (*sql.DB, error)
func (p *Pool) Close() error
func (p *Pool) Stats() map[string]PoolStats
```

### Router
```go
type Router struct {
    pool      *Pool
    databases map[string]string  // database → cluster
}

func NewRouter(pool *Pool, databases config.DatabasesConfig) *Router
func (r *Router) GetDB(database string) (*sql.DB, error)
func (r *Router) Query(ctx context.Context, db, sql string, args ...any) (*sql.Rows, error)
func (r *Router) Exec(ctx context.Context, db, sql string, args ...any) (sql.Result, error)
```

## Connection Pool Settings
| Setting | Default | Description |
|---------|---------|-------------|
| max_open_conns | 50 | Max open connections per cluster |
| max_idle_conns | 25 | Max idle connections per cluster |
| conn_max_lifetime | 5m | Connection max lifetime |
| conn_max_idle_time | 1m | Connection max idle time |

## Routing Example
```yaml
# config.yaml
clusters:
  primary:
    host: mysql-primary.example.com
    port: 3306
  replica:
    host: mysql-replica.example.com
    port: 3306

databases:
  user_db:
    cluster: primary      # Writes go to primary
  analytics_db:
    cluster: replica      # Reads go to replica
```

## Dependencies
```
Upstream:
  └── internal/config  → ClustersConfig, DatabasesConfig

Downstream:
  └── internal/mcp     → Router for SQL execution

External:
  └── database/sql     → Standard library
  └── github.com/go-sql-driver/mysql  → MySQL driver
```

## Update Rule
If pool/router behavior changes, update:
1. This file
2. Relevant .go file
3. *_test.go
4. docs/reference/configuration.md
