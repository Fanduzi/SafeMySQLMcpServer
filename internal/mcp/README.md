# MCP Module

Model Context Protocol implementation for MySQL operations.

## Files
| File | Responsibility |
|------|---------------|
| server.go | MCP SDK server setup |
| tools.go | MCP tool implementations (query, list_databases, etc.) |
| tools_test.go | Unit tests for input validation and result types |

## Test Coverage
- Input validation for all 7 MCP tools
- Result type marshaling
- Handler creation with nil dependencies

## Exports
- `NewMCPServer() *mcp.Server` - Create MCP server
- `RegisterTools(server *mcp.Server, h *Handler)` - Register all tools
- `Handler` - MCP tool handler
- `NewHandler(router, parser, checker, rewriter, audit, config) *Handler` - Create handler

## MCP Tools
| Tool | Description |
|------|-------------|
| query | Execute SQL query |
| list_databases | List available databases |
| list_tables | List tables in database |
| describe_table | Get table structure |
| show_create_table | Get CREATE TABLE statement |
| explain | Get query execution plan |
| search_tables | Search tables by pattern |

## Dependencies
- Upstream:
  - `internal/database` - Database operations
  - `internal/security` - SQL validation
  - `internal/audit` - Audit logging
  - `internal/config` - Configuration
- Downstream: `internal/server` - HTTP handler

## Update Rule
If MCP tools change, update this file in the same change.
