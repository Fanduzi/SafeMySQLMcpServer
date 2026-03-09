# Server Module

HTTP server with authentication, rate limiting, metrics, and MCP handler.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Server                             │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                  Middleware Stack                        ││
│  │                                                          ││
│  │  Request ──▶ Metrics ──▶ RateLimit ──▶ Auth ──▶ Handler ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Endpoints                             ││
│  │                                                          ││
│  │  POST /mcp      MCP JSON-RPC (auth required)            ││
│  │  GET  /health   Health check (no auth)                  ││
│  │  GET  /metrics  Prometheus metrics (no auth)            ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                MCP SDK Handler                           ││
│  │           (StreamableHTTPHandler)                       ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| http.go | HTTP server setup, middleware, routing | ~200 |
| ratelimit.go | IP-based rate limiting | ~80 |
| http_test.go | HTTP handler tests | ~100 |
| ratelimit_test.go | Rate limiter tests | ~80 |

## Test Coverage
```
Coverage: ~80%
- Endpoint routing
- Middleware execution order
- JWT authentication
- Rate limiting logic
- Graceful shutdown
```

## Exports

### Server
```go
type Server struct {
    config    *config.ReloadableConfig
    http      *http.Server
    validator *auth.Validator
    pool      *database.Pool
    router    *database.Router
    parser    *security.Parser
    checker   *security.Checker
    rewriter  *security.Rewriter
    audit     *audit.Logger
    metrics   *metrics.Collector
    rateLimit *rateLimiter
    mcpServer *mcp.Server
}

func New(cfg *config.ReloadableConfig) (*Server, error)
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
func (s *Server) UpdateConfig(cfg *config.Config, security *config.SecurityConfig)
```

## Endpoints
| Path | Method | Auth | Handler |
|------|--------|------|---------|
| `/mcp` | POST | Required | MCP JSON-RPC |
| `/health` | GET | None | Health check |
| `/metrics` | GET | None | Prometheus |

## Middleware Stack
```
Order: outer → inner

1. Metrics Middleware
   - Record request start time
   - Track active requests
   - Record duration on completion

2. Rate Limit Middleware
   - Per-IP rate limiting
   - Default: 100 requests/minute
   - Returns 429 when exceeded

3. Auth Middleware (for /mcp only)
   - Extract Bearer token
   - Validate JWT
   - Add user info to context
```

## Response Headers
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1646824600
```

## Dependencies
```
Upstream:
  ├── internal/config   → ReloadableConfig
  ├── internal/auth     → Validator
  ├── internal/metrics  → Collector
  └── internal/mcp      → Server, Handler

Downstream:
  └── (none - entry point)

External:
  └── net/http  → Standard library
```

## Update Rule
If server/middleware changes, update:
1. This file
2. Relevant .go file
3. *_test.go
4. docs/admin/troubleshooting.md (if error handling changes)
