# Server Entry Point

Application entry point that initializes and starts the SafeMySQLMcpServer.

## Files
| File | Responsibility |
|------|---------------|
| main.go | Parse flags, load config, create and start server |
| main_test.go | Unit tests for config loading and signal handling |

## Test Coverage
- Config file validation
- Signal handling for graceful shutdown
- Config flag parsing

## Exports
- `main()` - Application entry point

## Dependencies
- Upstream: None (entry point)
- Downstream:
  - `internal/config` - Configuration loading
  - `internal/server` - HTTP server
  - `internal/auth` - JWT validation
  - `internal/database` - Connection pool

## Update Rule
If server initialization changes, update this file in the same change.
