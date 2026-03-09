# Config Module

Configuration loading with environment variable expansion and hot reload support.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Configuration Flow                        │
│                                                              │
│  config.yaml          security.yaml                         │
│       │                    │                                 │
│       ▼                    ▼                                 │
│  ┌─────────────┐    ┌─────────────┐                        │
│  │    Load()   │    │LoadSecurity │                        │
│  │ ${VAR} expand│    │   ()        │                        │
│  └─────────────┘    └─────────────┘                        │
│       │                    │                                 │
│       ▼                    ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              ReloadableConfig                            ││
│  │   • Thread-safe access                                  ││
│  │   • Atomic updates                                      ││
│  │   • Hot reload support                                  ││
│  └─────────────────────────────────────────────────────────┘│
│       │                    │                                 │
│       ▼                    ▼                                 │
│  ┌─────────────┐    ┌─────────────┐                        │
│  │  Watcher    │    │  Validation │                        │
│  │ (fsnotify)  │    │  (defaults) │                        │
│  └─────────────┘    └─────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| config.go | Config structs, Load, LoadSecurity, defaults | ~200 |
| validation.go | Configuration validation helpers | ~80 |
| watcher.go | File watcher for hot reload | ~100 |
| config_test.go | Config loading tests | ~100 |
| validation_test.go | Validation tests | ~60 |
| watcher_test.go | Watcher tests | ~50 |

## Test Coverage
```
Coverage: ~85%
- YAML parsing
- Environment variable expansion
- Default value application
- Validation rules
- Hot reload mechanism
```

## Exports

### Configuration Types
```go
type Config struct {
    Server    ServerConfig              `yaml:"server"`
    Clusters  map[string]ClusterConfig  `yaml:"clusters"`
    Databases map[string]DatabaseConfig `yaml:"databases"`
    Security  SecurityFileConfig        `yaml:"security"`
    Audit     AuditConfig               `yaml:"audit"`
    RateLimit RateLimitConfig           `yaml:"rate_limit"`
}

type SecurityConfig struct {
    AllowedDML   []string `yaml:"allowed_dml"`
    AllowedDDL   []string `yaml:"allowed_ddl"`
    Blocked      []string `yaml:"blocked"`
    AutoLimit    int      `yaml:"auto_limit"`
    MaxLimit     int      `yaml:"max_limit"`
    QueryTimeout string   `yaml:"query_timeout"`
    MaxRows      int      `yaml:"max_rows"`
    MaxSQLLength int      `yaml:"max_sql_length"`
}
```

### Loading Functions
```go
func Load(path string) (*Config, error)
func LoadSecurity(path string) (*SecurityConfig, error)
```

### Reloadable Config
```go
type ReloadableConfig struct {
    // Thread-safe config access
}

func NewReloadableConfig(cfg *Config, security *SecurityConfig) *ReloadableConfig
func (r *ReloadableConfig) Get() *Config
func (r *ReloadableConfig) GetSecurity() *SecurityConfig
func (r *ReloadableConfig) Update(cfg *Config, security *SecurityConfig)
```

### Watcher
```go
type Watcher struct {
    // File system watcher
}

func NewWatcher(paths []string, onChange func()) (*Watcher, error)
func (w *Watcher) Close() error
```

## Environment Variables
| Variable | Usage |
|----------|-------|
| `${JWT_SECRET}` | JWT signing secret |
| `${DB_USER}` | Database username |
| `${DB_PASSWORD}` | Database password |
| `${ANY_VAR}` | Substituted in config |

## Default Values
| Setting | Default | Description |
|---------|---------|-------------|
| server.host | 0.0.0.0 | Listen address |
| server.port | 8080 | Listen port |
| server.log_level | info | Log level |
| max_open_conns | 50 | Max DB connections |
| max_idle_conns | 25 | Max idle connections |
| query_timeout | 30s | Query timeout |
| max_rows | 10000 | Max rows returned |

## Dependencies
```
Upstream: None (root module)

Downstream:
  ├── internal/server   → ServerConfig
  ├── internal/database → ClustersConfig, DatabasesConfig
  ├── internal/security → SecurityConfig
  ├── internal/audit    → AuditConfig
  └── internal/mcp      → ReloadableConfig

External:
  ├── gopkg.in/yaml.v3  → YAML parsing
  └── github.com/fsnotify/fsnotify → File watching
```

## Update Rule
If config structure changes, update:
1. This file
2. config.go (structs)
3. validation.go (rules)
4. docs/reference/configuration.md
