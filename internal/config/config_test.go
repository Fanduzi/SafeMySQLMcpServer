package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
server:
  host: "0.0.0.0"
  port: 8080

clusters:
  primary:
    host: localhost
    port: 3306
    username: root
    password: test123

databases:
  mydb:
    cluster: primary

security:
  config_file: security.yaml

audit:
  enabled: true
  log_file: /tmp/audit.log
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify defaults are applied
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if len(cfg.Clusters) == 0 {
		t.Error("Expected clusters to be loaded")
	}

	if cfg.Clusters["primary"].MaxOpenConns == 0 {
		t.Error("Expected MaxOpenConns default to be set")
	}
}

func TestLoadSecurity(t *testing.T) {
	content := `
security:
  allowed_dml:
    - SELECT
    - INSERT
    - UPDATE
    - DELETE
  allowed_ddl:
    - CREATE_TABLE
    - CREATE_INDEX
  blocked:
    - DROP
    - TRUNCATE
  auto_limit: 1000
  max_limit: 10000
`
	tmpFile, err := os.CreateTemp("", "security*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	cfg, err := LoadSecurity(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSecurity() error = %v", err)
	}

	if len(cfg.Security.AllowedDML) != 4 {
		t.Errorf("AllowedDML count = %d, want 4", len(cfg.Security.AllowedDML))
	}

	if !cfg.Security.IsDMLAllowed("SELECT") {
		t.Error("SELECT should be allowed")
	}

	if !cfg.Security.IsBlocked("DROP") {
		t.Error("DROP should be blocked")
	}
}

func TestSecurityRules(t *testing.T) {
	rules := &SecurityRules{
		AllowedDML: []string{"SELECT", "INSERT"},
		AllowedDDL: []string{"CREATE_TABLE"},
		Blocked:    []string{"DROP", "TRUNCATE"},
	}

	// Test IsDMLAllowed
	if !rules.IsDMLAllowed("SELECT") {
		t.Error("SELECT should be allowed")
	}
	if !rules.IsDMLAllowed("select") { // case insensitive
		t.Error("select (lowercase) should be allowed")
	}
	if rules.IsDMLAllowed("DELETE") {
		t.Error("DELETE should not be allowed")
	}

	// Test IsDDLAllowed
	if !rules.IsDDLAllowed("CREATE_TABLE") {
		t.Error("CREATE_TABLE should be allowed")
	}
	if rules.IsDDLAllowed("ALTER_TABLE") {
		t.Error("ALTER_TABLE should not be allowed")
	}

	// Test IsBlocked
	if !rules.IsBlocked("DROP") {
		t.Error("DROP should be blocked")
	}
	if !rules.IsBlocked("drop") { // case insensitive
		t.Error("drop (lowercase) should be blocked")
	}
	if rules.IsBlocked("SELECT") {
		t.Error("SELECT should not be blocked")
	}
}

func TestClusterConfig_GetDSN(t *testing.T) {
	cfg := ClusterConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "secret",
	}

	dsn := cfg.GetDSN()
	expected := "root:secret@tcp(localhost:3306)/?parseTime=true&loc=Local&charset=utf8mb4"
	if dsn != expected {
		t.Errorf("GetDSN() = %s, want %s", dsn, expected)
	}
}

func TestExpandEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	// Test expansion
	input := "password: ${TEST_VAR}"
	result := expandEnv(input)
	expected := "password: test_value"
	if result != expected {
		t.Errorf("expandEnv() = %s, want %s", result, expected)
	}

	// Test with non-existent variable
	input2 := "value: ${NON_EXISTENT}"
	result2 := expandEnv(input2)
	if result2 != "value: " {
		t.Errorf("expandEnv() = %s, want 'value: '", result2)
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &Config{
		Clusters: ClustersConfig{
			"test": ClusterConfig{},
		},
	}

	setDefaults(cfg)

	// Check server default
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port default = %d, want 8080", cfg.Server.Port)
	}

	// Check cluster defaults
	testCluster := cfg.Clusters["test"]
	if testCluster.Port != 3306 {
		t.Errorf("Cluster.Port default = %d, want 3306", testCluster.Port)
	}
	if testCluster.MaxOpenConns != 10 {
		t.Errorf("Cluster.MaxOpenConns default = %d, want 10", testCluster.MaxOpenConns)
	}
	if testCluster.MaxIdleConns != 5 {
		t.Errorf("Cluster.MaxIdleConns default = %d, want 5", testCluster.MaxIdleConns)
	}
	if testCluster.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("Cluster.ConnMaxLifetime default = %v, want 5m", testCluster.ConnMaxLifetime)
	}

	// Check audit defaults
	if cfg.Audit.MaxSQLLength != 2000 {
		t.Errorf("Audit.MaxSQLLength default = %d, want 2000", cfg.Audit.MaxSQLLength)
	}
}
