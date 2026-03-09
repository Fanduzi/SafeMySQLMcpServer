# MCP Module

Model Context Protocol implementation using official Go SDK (`github.com/modelcontextprotocol/go-sdk`).

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP SDK Server                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              StreamableHTTPHandler                       │ │
│  │              (HTTP transport layer)                      │ │
│  └─────────────────────────────────────────────────────────┘ │
│                            │                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Tool Registry (7 tools)                     │ │
│  │  query | list_databases | list_tables | describe_table  │ │
│  │  show_create_table | explain | search_tables            │ │
│  └─────────────────────────────────────────────────────────┘ │
│                            │                                 │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Handler (business logic)                    │ │
│  │  • Input validation                                      │ │
│  │  • SQL parsing & security check                         │ │
│  │  • Database execution                                    │ │
│  │  • Audit logging                                         │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| server.go | MCP SDK server setup, tool registration | ~50 |
| tools.go | MCP tool implementations, input/output types | ~300 |
| tools_test.go | Unit tests for input validation | ~200 |

## Test Coverage
```
Coverage: ~85%
- Input validation for all 7 MCP tools
- Result type marshaling (JSON)
- Handler creation with dependencies
- Error handling paths
```

## Exports

### Server Creation
```go
// Create MCP server with implementation info
func NewMCPServer() *mcp.Server

// Register all tools to server
func RegisterTools(server *mcp.Server, h *Handler)
```

### Handler
```go
type Handler struct {
    router   *database.Router
    parser   *security.Parser
    checker  *security.Checker
    rewriter *security.Rewriter
    audit    *audit.Logger
    config   *config.ReloadableConfig
}

func NewHandler(router, parser, checker, rewriter, audit, config) *Handler
```

### Input Types (with jsonschema tags)
```go
type QueryInput struct {
    Database string `json:"database" jsonschema:"required,description=Target database name"`
    SQL      string `json:"sql" jsonschema:"required,description=SQL statement to execute"`
}

type ListTablesInput struct {
    Database string `json:"database" jsonschema:"required,description=Database name"`
}
// ... etc
```

## MCP Tools
| Tool | Input | Output | Description |
|------|-------|--------|-------------|
| query | database, sql | columns, rows, rows_affected | Execute SQL query |
| list_databases | - | databases[] | List configured databases |
| list_tables | database | tables[] | List tables in database |
| describe_table | database, table | columns[] | Get table structure |
| show_create_table | database, table | create_statement | Get CREATE TABLE |
| explain | database, sql | columns, rows | Get execution plan |
| search_tables | table_pattern | matches[] | Search tables by pattern |

## Dependencies
```
Upstream:
  ├── internal/database  → Router for SQL execution
  ├── internal/security  → Parser, Checker, Rewriter
  ├── internal/audit     → Logger for audit trail
  └── internal/config    → ReloadableConfig

Downstream:
  └── internal/server    → HTTP handler wraps MCP handler

External:
  └── github.com/modelcontextprotocol/go-sdk
```

## Update Rule
If MCP tools change, update:
1. This file (tool list)
2. `tools.go` (implementation)
3. `tools_test.go` (tests)
4. `docs/reference/mcp-tools.md` (user docs)
