# Server Entry Point

Application entry point that initializes and starts the SafeMySQLMcpServer.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Startup Sequence                          │
│                                                              │
│  1. Parse Flags                                             │
│     └── -config path (default: config/config.yaml)          │
│                                                              │
│  2. Load Configuration                                       │
│     └── config.yaml + security.yaml + env vars              │
│                                                              │
│  3. Initialize Components                                    │
│     ├── Auth Validator (JWT)                                │
│     ├── Database Pool (MySQL connections)                   │
│     ├── Security Layer (Parser, Checker, Rewriter)          │
│     ├── Audit Logger                                        │
│     └── Metrics Collector                                   │
│                                                              │
│  4. Create MCP Server                                        │
│     └── Register 7 MCP tools                                │
│                                                              │
│  5. Start HTTP Server                                        │
│     └── Listen on :8080                                     │
│                                                              │
│  6. Watch for Signals                                        │
│     ├── SIGINT/SIGTERM → Graceful shutdown                  │
│     └── SIGHUP → Hot reload config                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| main.go | Parse flags, load config, start server | ~100 |

## Command Line
```bash
./bin/safe-mysql-mcp -config config/config.yaml

# Options:
#   -config string
#       Path to config file (default "config/config.yaml")
```

## Environment Variables
| Variable | Required | Description |
|----------|----------|-------------|
| `JWT_SECRET` | Yes | JWT signing secret (min 32 chars) |
| `DB_USER` | For clusters | Database username |
| `DB_PASSWORD` | For clusters | Database password |

## Startup Flow
```
1. main()
   │
   ├── flag.Parse() ──────────────────────┐
   │                                       │
   ├── config.Load(path)                   │
   │   └── Expand ${VAR} from environment  │
   │                                       │
   ├── auth.NewValidatorFromEnv()          │
   │   └── Read JWT_SECRET                 │
   │                                       │
   ├── database.NewPool(clusters)          │
   │   └── Connect to MySQL clusters       │
   │                                       │
   ├── security.New*()                     │
   │   └── Parser, Checker, Rewriter       │
   │                                       │
   ├── audit.NewLogger(config)             │
   │   └── Open audit log file             │
   │                                       │
   ├── server.New(config)                  │
   │   ├── Create MCP server               │
   │   └── Setup HTTP handlers             │
   │                                       │
   └── server.Start()                      │
       └── Listen on :8080                 │
```

## Graceful Shutdown
```
Signal (SIGINT/SIGTERM)
    │
    ├── Stop accepting new requests
    │
    ├── Wait for in-flight requests (max 30s)
    │
    ├── Close database connections
    │
    └── Exit 0
```

## Dependencies
```
Upstream: None (entry point)

Downstream:
  ├── internal/config   → Configuration loading
  ├── internal/server   → HTTP server
  ├── internal/auth     → JWT validation
  ├── internal/database → Connection pool
  ├── internal/security → SQL validation
  ├── internal/audit    → Audit logging
  ├── internal/metrics  → Prometheus metrics
  └── internal/mcp      → MCP server
```

## Update Rule
If server initialization changes, update:
1. This file
2. main.go
3. docs/guide/getting-started.md
