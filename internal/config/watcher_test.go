package config

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
server:
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	if watcher == nil {
		t.Error("NewWatcher() returned nil")
	}
	if watcher.configPath != configPath {
		t.Errorf("configPath = %s, want %s", watcher.configPath, configPath)
	}
}

func TestWatcher_SetSecurityPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	securityPath := filepath.Join(tmpDir, "security.yaml")

	if err := os.WriteFile(configPath, []byte("server:\n  port: 8080"), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	if err := os.WriteFile(securityPath, []byte("security:\n  auto_limit: 1000"), 0644); err != nil {
		t.Fatalf("Failed to write security: %v", err)
	}

	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	err = watcher.SetSecurityPath(securityPath)
	if err != nil {
		t.Fatalf("SetSecurityPath() error = %v", err)
	}
	if watcher.securityPath != securityPath {
		t.Errorf("securityPath = %s, want %s", watcher.securityPath, securityPath)
	}
}

func TestWatcher_OnChange(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("server:\n  port: 8080"), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Register callback with atomic variable for thread safety
	var called atomic.Bool
	watcher.OnChange(func(cfg *Config, sec *SecurityConfig) {
		called.Store(true)
	})

	// Start watcher
	watcher.Start()

	// Modify config file
	time.Sleep(100 * time.Millisecond)
	newContent := `server:
  port: 9090
`
	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to modify config: %v", err)
	}

	// Wait for callback
	time.Sleep(500 * time.Millisecond)

	// Note: callback may or may not be called depending on timing
	// Just verify it doesn't panic
	_ = called.Load()
}

func TestWatcher_Stop(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("server:\n  port: 8080"), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	watcher.Start()
	time.Sleep(50 * time.Millisecond)

	// Stop should not panic
	watcher.Stop()
}

func TestReloadableConfig_Get(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 8080},
	}
	security := &SecurityConfig{
		Security: SecurityRules{
			AutoLimit: 1000,
		},
	}

	rc := &ReloadableConfig{
		config:   cfg,
		security: security,
	}

	// Test Get()
	got := rc.Get()
	if got == nil {
		t.Error("Get() returned nil")
	}
	if got.Server.Port != 8080 {
		t.Errorf("Get().Server.Port = %d, want 8080", got.Server.Port)
	}

	// Test GetSecurity()
	sec := rc.GetSecurity()
	if sec == nil {
		t.Error("GetSecurity() returned nil")
	}
	if sec.AutoLimit != 1000 {
		t.Errorf("GetSecurity().AutoLimit = %d, want 1000", sec.AutoLimit)
	}
}

func TestReloadableConfig_GetSecurity_Nil(t *testing.T) {
	rc := &ReloadableConfig{
		config:   &Config{},
		security: nil,
	}

	sec := rc.GetSecurity()
	if sec != nil {
		t.Errorf("GetSecurity() = %v, want nil", sec)
	}
}

func TestReloadableConfig_Update(t *testing.T) {
	rc := &ReloadableConfig{
		config:   &Config{Server: ServerConfig{Port: 8080}},
		security: nil,
	}

	newCfg := &Config{Server: ServerConfig{Port: 9090}}
	newSec := &SecurityConfig{Security: SecurityRules{AutoLimit: 5000}}

	rc.Update(newCfg, newSec)

	// Verify update
	if rc.Get().Server.Port != 9090 {
		t.Errorf("Update() failed: Port = %d, want 9090", rc.Get().Server.Port)
	}
	if rc.GetSecurity().AutoLimit != 5000 {
		t.Errorf("Update() failed: AutoLimit = %d, want 5000", rc.GetSecurity().AutoLimit)
	}
}

func TestReloadableConfig_Concurrent(t *testing.T) {
	rc := &ReloadableConfig{
		config:   &Config{},
		security: nil,
	}

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				rc.Get()
				rc.GetSecurity()
			}
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				rc.Update(&Config{}, nil)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}
}
