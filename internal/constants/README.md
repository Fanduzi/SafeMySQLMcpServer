# Constants Module

Shared constants used across the application.

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| constants.go | Application-wide constants | ~30 |

## Exports

### Application Info
```go
const (
    Version  = "1.0.0"
    AppName  = "safe-mysql-mcp"
)
```

### Server Defaults
```go
const (
    DefaultPort       = 8080
    DefaultConfigPath = "config/config.yaml"
    DefaultLogLevel   = "info"
)
```

### Security Defaults
```go
const (
    DefaultMaxRows      = 10000
    DefaultMaxLimit     = 10000
    DefaultAutoLimit    = 1000
    DefaultQueryTimeout = 30 * time.Second
    DefaultMaxSQLLength = 100000
)
```

### Database Defaults
```go
const (
    DefaultMaxOpenConns    = 50
    DefaultMaxIdleConns    = 25
    DefaultConnMaxLifetime = 5 * time.Minute
    DefaultConnMaxIdleTime = 1 * time.Minute
)
```

### Audit Defaults
```go
const (
    DefaultMaxSQLLogLength = 2000
    DefaultMaxLogSizeMB    = 100
    DefaultMaxLogBackups   = 10
    DefaultMaxLogAgeDays   = 30
)
```

### Rate Limit Defaults
```go
const (
    DefaultRequestsPerMinute = 100
    DefaultBurst             = 20
)
```

## Usage Example
```go
import "safemysqlmcpserver/internal/constants"

func main() {
    fmt.Printf("%s v%s\n", constants.AppName, constants.Version)
    // Output: safe-mysql-mcp v1.0.0

    server := http.Server{
        Addr: fmt.Sprintf(":%d", constants.DefaultPort),
    }
}
```

## Dependencies
```
Upstream: None (root module)

Downstream:
  ├── cmd/server
  ├── internal/config
  ├── internal/server
  └── (all other modules)
```

## Update Rule
If constants change, update:
1. This file
2. constants.go
3. Check all modules using the constant
