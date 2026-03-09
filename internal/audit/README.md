# Audit Module

Audit logging with rotation and compression for compliance.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Audit Logging                             │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Entry Source                          ││
│  │                                                          ││
│  │   MCP Tool Call ──▶ Security Check ──▶ Database Query   ││
│  │         │                │                   │           ││
│  │         └────────────────┼───────────────────┘           ││
│  │                          ▼                               ││
│  │                    Entry{}                               ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                            │                                 │
│                            ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                     Logger                               ││
│  │                                                          ││
│  │   ┌───────────┐  ┌───────────┐  ┌───────────┐          ││
│  │   │ Truncate  │→ │  Format   │→ │  Write    │          ││
│  │   │   SQL     │  │   JSON    │  │   File    │          ││
│  │   └───────────┘  └───────────┘  └───────────┘          ││
│  │                          │                               ││
│  │                          ▼                               ││
│  │   ┌───────────────────────────────────────────┐        ││
│  │   │         Log Rotation (lumberjack)          │        ││
│  │   │  Size limit → Rotate → Compress → Age out │        ││
│  │   └───────────────────────────────────────────┘        ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Files
| File | Responsibility | Lines |
|------|---------------|-------|
| logger.go | Audit logger with log rotation | ~80 |
| logger_test.go | Unit tests | ~60 |

## Test Coverage
```
Coverage: ~85%
- Entry logging
- SQL truncation
- JSON formatting
- File rotation
```

## Exports

### Logger
```go
type Logger struct {
    file       *lumberjack.Logger
    maxSQLLen  int
}

func NewLogger(cfg *config.AuditConfig) (*Logger, error)
func (l *Logger) Log(entry Entry)
func (l *Logger) Close() error
```

### Entry
```go
type Entry struct {
    Timestamp    time.Time `json:"timestamp"`
    UserID       string    `json:"user_id"`
    UserEmail    string    `json:"user_email"`
    ClientIP     string    `json:"client_ip"`
    Database     string    `json:"database"`
    SQL          string    `json:"sql"`
    SQLType      string    `json:"sql_type"`
    Status       string    `json:"status"`        // success, error, blocked
    BlockReason  string    `json:"block_reason"`
    RowsAffected int64     `json:"rows_affected"`
    DurationMs   int64     `json:"duration_ms"`
}
```

## Log Format (JSON)
```json
{
  "timestamp": "2026-03-09T10:30:00.123Z",
  "user_id": "admin",
  "user_email": "admin@example.com",
  "client_ip": "192.168.1.100",
  "database": "mydb",
  "sql": "SELECT * FROM users WHERE id = 123",
  "sql_type": "SELECT",
  "status": "success",
  "rows_affected": 1,
  "duration_ms": 15,
  "block_reason": ""
}
```

## Configuration
| Setting | Default | Description |
|---------|---------|-------------|
| enabled | true | Enable audit logging |
| log_file | logs/audit.log | Log file path |
| max_sql_length | 2000 | Truncate SQL after N chars |
| max_size_mb | 100 | Rotate at N MB |
| max_backups | 10 | Keep N backup files |
| max_age_days | 30 | Delete backups after N days |
| compress | true | Compress rotated logs |

## Log Rotation
```
1. Log reaches max_size_mb
2. Current file renamed: audit-2026-03-09T10.30.00.log
3. New audit.log created
4. If compress: gzip the backup
5. Delete files older than max_age_days
6. Keep only max_backups files
```

## Compliance
| Regulation | Retention | Notes |
|------------|-----------|-------|
| SOX | 7 years | Financial systems |
| GDPR | As needed | Data minimization |
| HIPAA | 6 years | Healthcare |
| PCI-DSS | 1 year | Payment cards |

## Dependencies
```
Upstream:
  └── internal/config  → AuditConfig

Downstream:
  └── internal/mcp     → Logs all operations

External:
  └── gopkg.in/natefinch/lumberjack.v2  → Log rotation
```

## Update Rule
If audit format changes, update:
1. This file
2. logger.go
3. logger_test.go
4. docs/admin/audit-logging.md
