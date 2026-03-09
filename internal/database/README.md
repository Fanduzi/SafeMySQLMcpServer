# Database Module

MySQL connection pool management and database routing.

## Files
| File | Responsibility |
|------|---------------|
| pool.go | Connection pool with reference counting |
| router.go | Database-to-cluster routing |
| doc.go | Package documentation |
| pool_test.go | Unit tests for pool |
| router_test.go | Unit tests for router |
| integration_test.go | Integration tests (requires MySQL) |

## Exports
- `Pool` - Connection pool manager
- `NewPool(clusters config.ClustersConfig) (*Pool, error)` - Create pool
- `Get(cluster string) (*sql.DB, error)` - Get database connection
- `Close() error` - Close all connections
- `Router` - Database router
- `NewRouter(pool *Pool, databases config.DatabasesConfig) *Router` - Create router
- `GetDB(database string) (*sql.DB, error)` - Get DB by name
- `Query(ctx, db, sql, args...) (*sql.Rows, error)` - Execute query
- `Exec(ctx, db, sql, args...) (sql.Result, error)` - Execute statement

## Dependencies
- Upstream: `internal/config` - Cluster/database config
- Downstream: `internal/mcp` - Executes SQL queries

## Update Rule
If pool/router behavior changes, update this file in the same change.
