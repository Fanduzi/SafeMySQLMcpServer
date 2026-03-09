# Config Module

Configuration loading with environment variable expansion and hot reload support.

## Files
| File | Responsibility |
|------|---------------|
| config.go | Config structs, Load, LoadSecurity, defaults |
| validation.go | Configuration validation helpers |
| watcher.go | File watcher for hot reload |
| config_test.go | Unit tests |

## Exports
- `Config` - Main configuration struct
- `Load(path string) (*Config, error)` - Load main config
- `LoadSecurity(path string) (*SecurityConfig, error)` - Load security config
- `ReloadableConfig` - Thread-safe config with hot reload
- `Watcher` - File system watcher

## Dependencies
- Upstream: None
- Downstream:
  - `internal/server` - Server config
  - `internal/database` - Database config
  - `internal/security` - Security rules
  - `internal/audit` - Audit config

## Update Rule
If config structure changes, update this file in the same change.
