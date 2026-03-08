package database

import (
	"testing"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

func TestNewRouter(t *testing.T) {
	// Create a mock pool
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	databases := config.DatabasesConfig{
		"mydb":    {Cluster: "primary"},
		"testdb":  {Cluster: "primary"},
		"admindb": {Cluster: "admin"},
	}

	router := NewRouter(pool, databases)
	if router == nil {
		t.Fatal("NewRouter returned nil")
	}
}

func TestRouter_GetCluster(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	databases := config.DatabasesConfig{
		"mydb":   {Cluster: "primary"},
		"testdb": {Cluster: "replica"},
	}

	router := NewRouter(pool, databases)

	tests := []struct {
		name      string
		database  string
		want      string
		wantError bool
	}{
		{
			name:     "existing database",
			database: "mydb",
			want:     "primary",
		},
		{
			name:     "another database",
			database: "testdb",
			want:     "replica",
		},
		{
			name:      "unknown database",
			database:  "unknowndb",
			wantError: true,
		},
		{
			name:      "empty database name",
			database:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := router.GetCluster(tt.database)
			if tt.wantError {
				if err == nil {
					t.Error("GetCluster() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetCluster() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_ListDatabases(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	databases := config.DatabasesConfig{
		"mydb":   {Cluster: "primary"},
		"testdb": {Cluster: "replica"},
		"admindb": {Cluster: "admin"},
	}

	router := NewRouter(pool, databases)

	list := router.ListDatabases()
	if len(list) != 3 {
		t.Errorf("ListDatabases() returned %d databases, want 3", len(list))
	}

	// Check all databases are present
	found := make(map[string]bool)
	for _, db := range list {
		found[db] = true
	}

	for db := range databases {
		if !found[db] {
			t.Errorf("ListDatabases() missing database %s", db)
		}
	}
}

func TestRouter_UpdateConfig(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	databases := config.DatabasesConfig{
		"mydb": {Cluster: "primary"},
	}

	router := NewRouter(pool, databases)

	// Update config
	newDatabases := config.DatabasesConfig{
		"newdb":   {Cluster: "newcluster"},
		"another": {Cluster: "newcluster"},
	}

	router.UpdateConfig(newDatabases)

	// Verify update
	cluster, err := router.GetCluster("newdb")
	if err != nil {
		t.Errorf("GetCluster() after update error = %v", err)
	}
	if cluster != "newcluster" {
		t.Errorf("GetCluster() after update = %v, want newcluster", cluster)
	}

	// Old database should not exist
	_, err = router.GetCluster("mydb")
	if err == nil {
		t.Error("GetCluster() for old database should return error")
	}
}

func TestRouter_EmptyDatabases(t *testing.T) {
	pool := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  make(config.ClustersConfig),
	}

	databases := config.DatabasesConfig{}

	router := NewRouter(pool, databases)

	list := router.ListDatabases()
	if len(list) != 0 {
		t.Errorf("ListDatabases() returned %d databases, want 0", len(list))
	}

	_, err := router.GetCluster("anydb")
	if err == nil {
		t.Error("GetCluster() should return error for empty databases")
	}
}
