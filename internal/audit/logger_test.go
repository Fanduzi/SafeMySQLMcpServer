package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

func TestNewLogger_Disabled(t *testing.T) {
	cfg := &config.AuditConfig{
		Enabled: false,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	if logger.enabled {
		t.Error("Logger should be disabled")
	}
}

func TestNewLogger_Enabled(t *testing.T) {
	tmpDir, err := os.CreateTemp("", "audit*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	_ = tmpDir.Close()
	_ = os.Remove(tmpDir.Name())
	_ = os.MkdirAll(tmpDir.Name(), 0755)
	defer func() { _ = os.RemoveAll(tmpDir.Name()) }()

	cfg := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit.log"),
		MaxSQLLength: 1000,
		MaxSizeMB:    10,
		MaxBackups:   5,
		MaxAgeDays:   30,
		Compress:     false,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer func() { _ = logger.Close() }()

	if !logger.enabled {
		t.Error("Logger should be enabled")
	}
}

func TestLogger_Log(t *testing.T) {
	tmpDir, err := os.CreateTemp("", "audit*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	_ = tmpDir.Close()
	_ = os.Remove(tmpDir.Name())
	_ = os.MkdirAll(tmpDir.Name(), 0755)
	defer func() { _ = os.RemoveAll(tmpDir.Name()) }()

	cfg := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit.log"),
		MaxSQLLength: 1000,
		MaxSizeMB:    10,
		MaxBackups:   5,
		MaxAgeDays:   30,
		Compress:     false,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer func() { _ = logger.Close() }()

	entry := Entry{
		UserID:     "user123",
		UserEmail:  "test@example.com",
		Database:   "mydb",
		SQL:        "SELECT * FROM users",
		SQLType:    "SELECT",
		Status:     "success",
		DurationMs: 50,
	}

	logger.Log(entry)

	// Verify log file was created and contains expected content
	data, err := os.ReadFile(cfg.LogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "user123") {
		t.Error("Log should contain user_id")
	}
	if !strings.Contains(content, "SELECT * FROM users") {
		t.Error("Log should contain SQL")
	}
	if !strings.Contains(content, "success") {
		t.Error("Log should contain status")
	}
}

func TestLogger_LogDisabled(t *testing.T) {
	cfg := &config.AuditConfig{
		Enabled: false,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	entry := Entry{
		UserID: "user123",
		SQL:    "SELECT * FROM users",
	}

	// Should not panic or error when logging is disabled
	logger.Log(entry)
}

func TestLogger_TruncateSQL(t *testing.T) {
	tests := []struct {
		name     string
		maxLen   int
		sql      string
		expected string
	}{
		{
			name:     "no truncation needed",
			maxLen:   100,
			sql:      "SELECT 1",
			expected: "SELECT 1",
		},
		{
			name:     "truncation needed",
			maxLen:   10,
			sql:      "SELECT * FROM users WHERE id = 1",
			expected: "SELECT * F",
		},
		{
			name:     "zero max length",
			maxLen:   0,
			sql:      "SELECT 1",
			expected: "SELECT 1",
		},
		{
			name:     "exact length",
			maxLen:   8,
			sql:      "SELECT 1",
			expected: "SELECT 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &Logger{maxSQLLen: tt.maxLen}
			got := logger.TruncateSQL(tt.sql)
			if got != tt.expected {
				t.Errorf("TruncateSQL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLogger_LogWithTimestamp(t *testing.T) {
	tmpDir, err := os.CreateTemp("", "audit*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	_ = tmpDir.Close()
	_ = os.Remove(tmpDir.Name())
	_ = os.MkdirAll(tmpDir.Name(), 0755)
	defer func() { _ = os.RemoveAll(tmpDir.Name()) }()

	cfg := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit.log"),
		MaxSQLLength: 1000,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer func() { _ = logger.Close() }()

	// Entry without timestamp - should be set automatically
	entry := Entry{
		UserID: "user123",
		SQL:    "SELECT 1",
	}
	logger.Log(entry)

	// Entry with timestamp - should be preserved
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	entry2 := Entry{
		Timestamp: fixedTime,
		UserID:    "user456",
		SQL:       "SELECT 2",
	}
	logger.Log(entry2)

	data, err := os.ReadFile(cfg.LogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "2024-01-01") {
		t.Error("Log should contain the fixed timestamp")
	}
}

func TestLogger_LogTruncation(t *testing.T) {
	tmpDir, err := os.CreateTemp("", "audit*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	tmpDir.Close()
	os.Remove(tmpDir.Name())
	os.MkdirAll(tmpDir.Name(), 0755)
	defer os.RemoveAll(tmpDir.Name())

	cfg := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit.log"),
		MaxSQLLength: 20,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer func() { _ = logger.Close() }()

	longSQL := "SELECT * FROM users WHERE id = 1 AND name = 'test'"
	entry := Entry{
		UserID: "user123",
		SQL:    longSQL,
	}
	logger.Log(entry)

	data, err := os.ReadFile(cfg.LogFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "truncated") {
		t.Error("Log should indicate SQL was truncated")
	}
}

func TestLogger_Close(t *testing.T) {
	// Test closing disabled logger
	logger := &Logger{enabled: false}
	if err := logger.Close(); err != nil {
		t.Errorf("Close() on disabled logger should not error: %v", err)
	}

	// Test closing logger with nil writer
	logger = &Logger{enabled: true, writer: nil}
	if err := logger.Close(); err != nil {
		t.Errorf("Close() with nil writer should not error: %v", err)
	}
}

func TestLogger_UpdateConfig(t *testing.T) {
	tmpDir, err := os.CreateTemp("", "audit*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	tmpDir.Close()
	os.Remove(tmpDir.Name())
	os.MkdirAll(tmpDir.Name(), 0755)
	defer os.RemoveAll(tmpDir.Name())

	cfg := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit.log"),
		MaxSQLLength: 1000,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	// Update to disabled
	newCfg := &config.AuditConfig{
		Enabled: false,
	}
	if err := logger.UpdateConfig(newCfg); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
	if logger.enabled {
		t.Error("Logger should be disabled after update")
	}

	// Re-enable with new config
	newCfg2 := &config.AuditConfig{
		Enabled:      true,
		LogFile:      filepath.Join(tmpDir.Name(), "audit2.log"),
		MaxSQLLength: 500,
	}
	if err := logger.UpdateConfig(newCfg2); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
	if !logger.enabled {
		t.Error("Logger should be enabled after update")
	}

	logger.Close()
}
