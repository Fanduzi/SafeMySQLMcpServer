# Audit Module

Audit logging with rotation and compression.

## Files
| File | Responsibility |
|------|---------------|
| logger.go | Audit logger with log rotation |
| logger_test.go | Unit tests |

## Exports
- `Logger` - Audit logger
- `NewLogger(cfg *config.AuditConfig) (*Logger, error)` - Create logger
- `Log(entry Entry)` - Write audit entry
- `Close() error` - Close logger
- `Entry` - Audit log entry struct

## Entry Fields
| Field | Type | Description |
|-------|------|-------------|
| Timestamp | time.Time | When operation occurred |
| UserID | string | User from JWT |
| UserEmail | string | Email from JWT |
| Database | string | Target database |
| SQL | string | SQL statement |
| SQLType | string | SELECT/INSERT/UPDATE/DELETE/etc |
| Status | string | success/blocked/error |
| BlockReason | string | Why blocked (if blocked) |
| RowsAffected | int64 | Rows affected |
| DurationMs | int64 | Duration in milliseconds |

## Features
- JSON format logging
- Log rotation by size
- Configurable retention
- Optional compression
- SQL truncation for long queries

## Dependencies
- Upstream: `internal/config` - Audit config
- Downstream: `internal/mcp` - Logs all operations

## Update Rule
If audit format changes, update this file in the same change.
