// Package audit handles audit logging with rotation.
// input: audit.Entry (user, sql, status, duration), config.AuditConfig
// output: JSON log files with rotation, compression
// pos: observability layer, records all SQL operations for compliance
// note: if this file changes, update header and internal/audit/README.md
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Entry represents a single audit log entry
type Entry struct {
	Timestamp    time.Time `json:"timestamp"`
	UserID       string    `json:"user_id"`
	UserEmail    string    `json:"user_email"`
	Database     string    `json:"database"`
	SQL          string    `json:"sql"`
	SQLType      string    `json:"sql_type"`
	Status       string    `json:"status"`
	BlockReason  string    `json:"block_reason,omitempty"`
	RowsAffected int64     `json:"rows_affected"`
	DurationMs   int64     `json:"duration_ms"`
	SQLLength    int       `json:"sql_length"`
}

// Logger handles audit logging
type Logger struct {
	mu        sync.Mutex
	writer    *lumberjack.Logger
	maxSQLLen int
	enabled   bool
}

// NewLogger creates a new audit logger
func NewLogger(cfg *config.AuditConfig) (*Logger, error) {
	if !cfg.Enabled {
		return &Logger{enabled: false}, nil
	}

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	writer := &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}

	return &Logger{
		writer:    writer,
		maxSQLLen: cfg.MaxSQLLength,
		enabled:   true,
	}, nil
}

// Log writes an audit entry to the log file
func (l *Logger) Log(entry Entry) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Set timestamp if not set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	// Truncate SQL if too long
	entry.SQLLength = len(entry.SQL)
	if l.maxSQLLen > 0 && len(entry.SQL) > l.maxSQLLen {
		entry.SQL = entry.SQL[:l.maxSQLLen] + "...(truncated)"
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	// Write to log file
	data = append(data, '\n')
	if _, err := l.writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit entry: %v\n", err)
	}
}

// Close closes the audit logger
func (l *Logger) Close() error {
	if l.writer == nil {
		return nil
	}
	return l.writer.Close()
}

// UpdateConfig updates the logger configuration
func (l *Logger) UpdateConfig(cfg *config.AuditConfig) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !cfg.Enabled {
		l.enabled = false
		if l.writer != nil {
			_ = l.writer.Close()
			l.writer = nil
		}
		return nil
	}

	// Create new writer with updated config
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	l.writer = &lumberjack.Logger{
		Filename:   cfg.LogFile,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	l.maxSQLLen = cfg.MaxSQLLength
	l.enabled = true

	return nil
}

// TruncateSQL truncates a SQL string to the maximum length
func (l *Logger) TruncateSQL(sql string) string {
	if l.maxSQLLen <= 0 || len(sql) <= l.maxSQLLen {
		return sql
	}
	return sql[:l.maxSQLLen]
}
