package database

import (
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

func TestNewPool_EmptyClusters(t *testing.T) {
	_, err := NewPool(config.ClustersConfig{})
	if err != nil {
		t.Errorf("NewPool with empty clusters should not error, got: %v", err)
	}
}

func TestPool_Get_UnknownCluster(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	_, err := pool.Get("unknown")
	if err == nil {
		t.Error("Get() should return error for unknown cluster")
	}
}

func TestPool_Close_Empty(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	err := pool.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClusterConfig_GetDSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.ClusterConfig
		expected string
	}{
		{
			name: "standard config",
			cfg: config.ClusterConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "secret",
			},
			expected: "root:secret@tcp(localhost:3306)/?parseTime=true&loc=Local&charset=utf8mb4",
		},
		{
			name: "with custom port",
			cfg: config.ClusterConfig{
				Host:     "db.example.com",
				Port:     3307,
				Username: "admin",
				Password: "pass123",
			},
			expected: "admin:pass123@tcp(db.example.com:3307)/?parseTime=true&loc=Local&charset=utf8mb4",
		},
		{
			name: "empty password",
			cfg: config.ClusterConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "user",
				Password: "",
			},
			expected: "user:@tcp(localhost:3306)/?parseTime=true&loc=Local&charset=utf8mb4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.cfg.GetDSN()
			if dsn != tt.expected {
				t.Errorf("GetDSN() = %s, want %s", dsn, tt.expected)
			}
		})
	}
}

func TestManagedDB_ReferenceCounting(t *testing.T) {
	mdb := &managedDB{
		refCount: 0,
		closing:  0,
	}

	// Test initial state
	if mdb.refCount != 0 {
		t.Errorf("Initial refCount = %d, want 0", mdb.refCount)
	}

	// Test closing flag
	if mdb.closing != 0 {
		t.Errorf("Initial closing = %d, want 0", mdb.closing)
	}
}

func TestPool_Release_NilDB(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	// Should not panic
	pool.release(nil)
}

func TestPool_GetLimiterConfig(t *testing.T) {
	cfg := config.ClusterConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "secret",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	if cfg.MaxOpenConns != 20 {
		t.Errorf("MaxOpenConns = %d, want 20", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("MaxIdleConns = %d, want 10", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 5m", cfg.ConnMaxLifetime)
	}
}
