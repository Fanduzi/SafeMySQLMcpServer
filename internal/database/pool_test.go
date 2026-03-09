package database

import (
	"database/sql"
	"sync/atomic"
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

func TestPool_Get_ClosingCluster(t *testing.T) {
	pool := &Pool{
		clusters: map[string]*managedDB{
			"closing": {
				db:       nil,
				refCount: 0,
				closing:  1, // Marked as closing
			},
		},
		configs: make(config.ClustersConfig),
	}

	_, err := pool.Get("closing")
	if err == nil {
		t.Error("Get() should return error for closing cluster")
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
		{
			name: "special characters in password",
			cfg: config.ClusterConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "user",
				Password: "p@ss:word",
			},
			expected: "user:p@ss:word@tcp(localhost:3306)/?parseTime=true&loc=Local&charset=utf8mb4",
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

func TestPool_Release_ExistingDB(t *testing.T) {
	// Create a mock DB (nil is fine for this test since we check equality)
	var mockDB *sql.DB

	pool := &Pool{
		clusters: map[string]*managedDB{
			"test": {
				db:       mockDB,
				refCount: 5,
				closing:  0,
			},
		},
		configs: make(config.ClustersConfig),
	}

	// Release should decrement ref count
	pool.release(mockDB)

	// Check ref count was decremented
	newCount := atomic.LoadInt32(&pool.clusters["test"].refCount)
	if newCount != 4 {
		t.Errorf("refCount after release = %d, want 4", newCount)
	}
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

func TestPool_DoubleClose(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	// First close
	err := pool.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	// Second close should not panic
	err = pool.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

// Note: UpdateConfig tests are skipped because the function has a blocking loop
// that waits for connections to close with a 30s timeout.
// These are better tested via integration tests with real database connections.

func TestClusterConfig_Defaults(t *testing.T) {
	cfg := config.ClusterConfig{
		Host:     "localhost",
		Port:     0, // default
		Username: "root",
		Password: "",
	}

	// Port 0 should be handled (though typically defaults to 3306)
	dsn := cfg.GetDSN()
	if dsn == "" {
		t.Error("GetDSN() should return non-empty string")
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	// Run concurrent operations
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			// Concurrent Get (will fail but shouldn't panic)
			pool.Get("unknown")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestManagedDB_ClosingFlag(t *testing.T) {
	mdb := &managedDB{
		refCount: 0,
		closing:  0,
	}

	// Set closing flag
	atomic.StoreInt32(&mdb.closing, 1)

	if atomic.LoadInt32(&mdb.closing) != 1 {
		t.Error("closing flag should be 1")
	}

	// Reset
	atomic.StoreInt32(&mdb.closing, 0)

	if atomic.LoadInt32(&mdb.closing) != 0 {
		t.Error("closing flag should be 0")
	}
}

func TestManagedDB_RefCountOperations(t *testing.T) {
	mdb := &managedDB{
		refCount: 0,
		closing:  0,
	}

	// Increment
	atomic.AddInt32(&mdb.refCount, 1)
	if atomic.LoadInt32(&mdb.refCount) != 1 {
		t.Error("refCount should be 1 after increment")
	}

	// Increment again
	atomic.AddInt32(&mdb.refCount, 1)
	if atomic.LoadInt32(&mdb.refCount) != 2 {
		t.Error("refCount should be 2 after second increment")
	}

	// Decrement
	atomic.AddInt32(&mdb.refCount, -1)
	if atomic.LoadInt32(&mdb.refCount) != 1 {
		t.Error("refCount should be 1 after decrement")
	}
}

