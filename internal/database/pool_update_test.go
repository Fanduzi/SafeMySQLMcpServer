package database

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fan/safe-mysql-mcp/internal/config"
)

// newClosableMock creates a sqlmock DB that expects Close().
func newClosableMock(t *testing.T) *sql.DB {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectClose()
	return db
}

func TestPool_UpdateConfig_RemoveCluster(t *testing.T) {
	db := newClosableMock(t)

	pool := &Pool{
		clusters: map[string]*managedDB{
			"old": {db: db, refCount: 0, closing: 0},
		},
		configs: config.ClustersConfig{
			"old": {Host: "localhost", Port: 3306},
		},
	}

	if err := pool.UpdateConfig(config.ClustersConfig{}); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	if _, exists := pool.clusters["old"]; exists {
		t.Error("old cluster should be removed")
	}
}

func TestPool_UpdateConfig_ChangedCredentials(t *testing.T) {
	db := newClosableMock(t)

	pool := &Pool{
		clusters: map[string]*managedDB{
			"primary": {db: db, refCount: 0, closing: 0},
		},
		configs: config.ClustersConfig{
			"primary": {Host: "localhost", Port: 3306, Username: "old", Password: "old"},
		},
	}

	if err := pool.UpdateConfig(config.ClustersConfig{
		"primary": {Host: "localhost", Port: 3306, Username: "new", Password: "new"},
	}); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	if _, exists := pool.clusters["primary"]; exists {
		t.Error("old cluster with changed credentials should be removed")
	}
}

func TestPool_UpdateConfig_NoChange(t *testing.T) {
	db := newClosableMock(t)

	pool := &Pool{
		clusters: map[string]*managedDB{
			"primary": {db: db, refCount: 0, closing: 0},
		},
		configs: config.ClustersConfig{
			"primary": {Host: "localhost", Port: 3306, Username: "root", Password: "pass"},
		},
	}

	if err := pool.UpdateConfig(config.ClustersConfig{
		"primary": {Host: "localhost", Port: 3306, Username: "root", Password: "pass"},
	}); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	if _, exists := pool.clusters["primary"]; !exists {
		t.Error("unchanged cluster should still exist")
	}
}

func TestPool_UpdateConfig_DeadlockRegression(t *testing.T) {
	db := newClosableMock(t)

	pool := &Pool{
		clusters: map[string]*managedDB{
			"primary": {db: db, refCount: 0, closing: 0},
		},
		configs: config.ClustersConfig{
			"primary": {Host: "localhost", Port: 3306, Username: "root", Password: "pass"},
		},
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Simulate concurrent Get/Release cycles through the real code path.
	// pool.release() holds p.mu.Lock(), which is exactly the path that
	// deadlocked with the old UpdateConfig — so this goroutine exercises
	// the lock contention that the regression test must verify.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					gotDB, err := pool.Get("primary")
					if err == nil {
						pool.release(gotDB)
					}
				}
			}
		}()
	}

	done := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		done <- pool.UpdateConfig(config.ClustersConfig{
			"primary": {Host: "newhost", Port: 3306, Username: "root", Password: "pass"},
		})
	}()

	if err := <-done; err != nil {
		t.Errorf("UpdateConfig() error = %v", err)
	}

	close(stop)
	wg.Wait()
}

func TestPool_UpdateConfig_EmptyToEmpty(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	if err := pool.UpdateConfig(config.ClustersConfig{}); err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}
}
