# Hot Reload

SafeMySQLMcpServer supports hot configuration reload, allowing you to update settings without restarting the server.

## How It Works

```
┌───────────────────┐     ┌───────────────────┐
│  Config File      │     │  Security File    │
│  config.yaml      │     │  security.yaml   │
└───────────────────┘     └───────────────────┘
         │                        │
         │    fsnotify            │
         ▼                        ▼
┌─────────────────────────────────────────────────┐
│              Config Watcher                     │
│  - Detects file changes                       │
│  - Debounces rapid changes                    │
│  - Parses new configuration                  │
└─────────────────────────────────────────────────┘
         │
         │ OnChange callback
         ▼
┌─────────────────────────────────────────────────┐
│              Server Update                     │
│  - Updates connection pool                   │
│  - Updates security rules                    │
│  - Updates audit settings                    │
└─────────────────────────────────────────────────┘
```

## Configuration Files
tab: Configuration Files

### Main Configuration (config.yaml)

Changes that trigger reload:

| Section | Hot Reloadable | Notes |
|---------|---------------|-------|
| `server` | Partial | Port/host require restart |
| `clusters` | Yes | Connections updated gracefully |
| `databases` | Yes | Routing updated immediately |
| `audit` | Yes | Logger settings updated |
| `rate_limit` | Yes | Rate limits updated |

### Security Configuration (security.yaml)

All settings are hot reloadable:

| Setting | Hot Reloadable |
|---------|---------------|
| `allowed_dml` | Yes |
| `allowed_ddl` | Yes |
| `blocked` | Yes |
| `auto_limit` | Yes |
| `query_timeout` | Yes |
| `max_rows` | Yes |

## Update Behavior
tab: Update Behavior

### Connection Pool Updates

When cluster configuration changes:

1. New connections use new settings
2. Existing connections complete their queries
3. Old connections gracefully closed when idle

```go
func (p *Pool) UpdateConfig(clusters ClustersConfig) error {
    // Create new connections for new clusters
    // Update settings for existing clusters
    // Mark removed clusters for cleanup
}
```

### Security Rule Updates

When security configuration changes:

1. New rules apply immediately
2. In-flight queries continue with old rules
3. New queries use new rules

```go
func (rc *ReloadableConfig) Update(cfg *Config, security *SecurityConfig) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.config = cfg
    rc.security = security
}
```

## Example
tab: Example

### Adding a New Database

```yaml
# config/config.yaml
databases:
  user_db:
    cluster: dev-cluster-1
  order_db:
    cluster: dev-cluster-1
  # Add new database - hot reload will make it available immediately
  analytics_db:
    cluster: dev-cluster-1
```

```bash
# Save the file - changes are detected automatically
# No restart needed!
# New database is immediately available
```

### Tightening Security

```yaml
# config/security.yaml
security:
  allowed_dml:
    - SELECT
    # Remove INSERT/UPDATE/DELETE - hot reload applies immediately
```

```bash
# Save the file
# New queries are immediately blocked
# In-flight queries complete normally
```

## Limitations
tab: Limitations

### Requires Restart

These changes require server restart:

| Change | Reason |
|-------|--------|
| Server port | Requires new listener |
| Server host | Requires new listener |
| JWT secret | Invalidates all tokens |

### Graceful Connection Handling

When cluster settings change:
- Existing queries complete normally
- New queries use new settings
- Idle connections are updated
- Active connections are left until idle
