# Server Module

HTTP server with authentication, rate limiting, and metrics middleware.

## Files
| File | Responsibility |
|------|---------------|
| http.go | HTTP server setup, middleware, routing |
| ratelimit.go | IP-based rate limiting |
| *_test.go | Unit tests |

## Exports
- `Server` - HTTP server
- `New(cfg *config.ReloadableConfig) (*Server, error)` - Create server
- `Start() error` - Start HTTP server
- `Shutdown(ctx context.Context) error` - Graceful shutdown
- `UpdateConfig(cfg *config.Config, security *config.SecurityConfig)` - Hot reload

## Endpoints
| Path | Handler | Auth |
|------|---------|------|
| `/mcp` | MCP JSON-RPC | Required |
| `/health` | Health check | None |
| `/metrics` | Prometheus metrics | None |

## Middleware Stack (outer → inner)
1. Metrics - Record request duration
2. Rate Limit - Per-IP rate limiting
3. Auth - JWT validation

## Dependencies
- Upstream:
  - `internal/config` - Server config
  - `internal/auth` - JWT validation
  - `internal/metrics` - Prometheus metrics
- Downstream:
  - `internal/mcp` - MCP handlers
  - `internal/database` - Connection pool

## Update Rule
If server/middleware changes, update this file in the same change.
