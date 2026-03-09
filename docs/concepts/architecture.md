# Architecture Overview

SafeMySQLMcpServer is built with a layered architecture for security and maintainability.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AI Client (Claude Code)                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              HTTP Server (:8080)                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ JWT Auth    │→ │ Rate Limit  │→ │ Metrics     │→ │ MCP Handler         │ │
│  │ Middleware  │  │ Middleware  │  │ Middleware  │  │ (7 MCP Tools)       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│                                                              │               │
│  Endpoints:                                                  │               │
│  • POST /mcp     MCP JSON-RPC                                │               │
│  • GET /health   Health check                                │               │
│  • GET /metrics  Prometheus metrics                          │               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Security Layer                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ SQL Parser  │→ │ Security    │→ │ SQL Rewriter│→ │ Input Validator     │ │
│  │ (vitess)    │  │ Checker     │  │ (auto-LIMIT)│  │ (identifier regex)  │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│                                                              │               │
│  Security Rules:                                             │               │
│  • DML/DDL allowlist                                         │               │
│  • Blocked operations (DROP, TRUNCATE)                       │               │
│  • Query timeout & row limits                                │               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
┌───────────────────────┐ ┌───────────────────────┐ ┌───────────────────────┐
│    Database Router    │ │    Audit Logger       │ │   Prometheus Metrics  │
│  ┌─────────────────┐  │ │  ┌─────────────────┐  │ │  ┌─────────────────┐  │
│  │ Connection Pool │  │ │  │ JSON Logs       │  │ │  │ HTTP Metrics    │  │
│  │ (per cluster)   │  │ │  │ Rotation        │  │ │  │ DB Metrics      │  │
│  └─────────────────┘  │ │  │ Compression     │  │ │  │ Security Stats  │  │
│                       │ │  └─────────────────┘  │ │  └─────────────────┘  │
│  Clusters:            │ │                       │ │                       │
│  • dev-cluster-1      │ │  Logs:                │ │  Exposed:             │
│  • dev-cluster-2      │ │  • User identity      │ │  • /metrics endpoint  │
│  • ...                │ │  • SQL statement      │ │  • scrape interval    │
└───────────────────────┘ │  • Status/Duration    │ └───────────────────────┘
          │               └───────────────────────┘
          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MySQL Clusters                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐              │
│  │ dev-cluster-1   │  │ dev-cluster-2   │  │ ...             │              │
│  │ :3306           │  │ :3306           │  │                 │              │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module Structure

```
safe-mysql-mcp/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── auth/            # JWT authentication
│   ├── audit/           # Audit logging
│   ├── config/          # Configuration management
│   ├── constants/       # Shared constants
│   ├── database/        # Connection pool & routing
│   ├── mcp/             # MCP protocol implementation
│   ├── metrics/         # Prometheus metrics
│   ├── security/        # SQL security layer
│   ├── server/          # HTTP server
│   └── validation/      # Input validation
├── pkg/
│   └── token/           # Token generation CLI
├── docs/                # Documentation
├── examples/            # Example configurations
└── config/              # Default configurations
```

## Key Components

| Component | Package | Responsibility |
|-----------|---------|---------------|
| HTTP Server | `internal/server` | HTTP handling, middleware |
| MCP Handler | `internal/mcp` | MCP protocol, tool execution |
| Security Layer | `internal/security` | SQL parsing, validation, rewriting |
| Database Router | `internal/database` | Connection pooling, query routing |
| Audit Logger | `internal/audit` | Operation logging |
| Configuration | `internal/config` | Config loading, hot reload |
| Authentication | `internal/auth` | JWT validation |
| Metrics | `internal/metrics` | Prometheus metrics |
| Validation | `internal/validation` | Input sanitization |

## Design Principles

tab: Design Principles

### 1. Defense in Depth
Multiple security layers protect against SQL injection:
- Input validation at boundaries
- SQL parsing and analysis
- Security rule checking
- Automatic rewriting for safety

### 2. Least Privilege
Each component has minimal required permissions:
- MCP tools can only access configured databases
- SQL operations limited by allowlist
- Query results limited by max rows

### 3. Audit Everything
All operations are logged:
- User identity from JWT
- SQL statement (truncated)
- Execution status
- Duration and rows affected

### 4. Graceful Degradation
Failures are handled gracefully:
- Invalid tokens return 401
- Blocked SQL returns error message
- Timeouts return appropriate error
- Connection errors are logged and reported
