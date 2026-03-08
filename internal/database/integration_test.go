//go:build integration
// +build integration

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

// TestIntegrationPool tests the pool with real MySQL
func TestIntegrationPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("MySQL connection not available (set MYSQL_HOST env var)")
	}

	clusters := config.ClustersConfig{
		"test": config.ClusterConfig{
			Host:            os.Getenv("MYSQL_HOST"),
			Port:            getEnvInt("MYSQL_PORT", 3306),
			Username:        os.Getenv("MYSQL_USER"),
			Password:        os.Getenv("MYSQL_PASSWORD"),
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}

	pool, err := NewPool(clusters)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	// Test Get
	db, err := pool.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Test connection is working
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	t.Logf("MySQL version: %s", version)
}

// TestIntegrationRouter tests the router with real MySQL
func TestIntegrationRouter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("MySQL connection not available")
	}

	clusters := config.ClustersConfig{
		"primary": config.ClusterConfig{
			Host:            os.Getenv("MYSQL_HOST"),
			Port:            getEnvInt("MYSQL_PORT", 3306),
			Username:        os.Getenv("MYSQL_USER"),
			Password:        os.Getenv("MYSQL_PASSWORD"),
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}

	pool, err := NewPool(clusters)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	databases := config.DatabasesConfig{
		"testdb": {Cluster: "primary"},
	}

	router := NewRouter(pool, databases)

	// Test GetDB
	db, err := router.GetDB("testdb")
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}

	// Test query
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("QueryRow() error = %v", err)
	}
	if result != 1 {
		t.Errorf("SELECT 1 = %d, want 1", result)
	}

	// Test GetCluster
	cluster, err := router.GetCluster("testdb")
	if err != nil {
		t.Fatalf("GetCluster() error = %v", err)
	}
	if cluster != "primary" {
		t.Errorf("GetCluster() = %s, want primary", cluster)
	}

	// Test ListDatabases
	list := router.ListDatabases()
	if len(list) != 1 || list[0] != "testdb" {
		t.Errorf("ListDatabases() = %v, want [testdb]", list)
	}
}

// TestIntegrationQuery tests Query and Exec methods
func TestIntegrationQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("MySQL connection not available")
	}

	clusters := config.ClustersConfig{
		"primary": config.ClusterConfig{
			Host:            os.Getenv("MYSQL_HOST"),
			Port:            getEnvInt("MYSQL_PORT", 3306),
			Username:        os.Getenv("MYSQL_USER"),
			Password:        os.Getenv("MYSQL_PASSWORD"),
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
		},
	}

	pool, err := NewPool(clusters)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	databases := config.DatabasesConfig{
		os.Getenv("MYSQL_DATABASE"): {Cluster: "primary"},
	}

	router := NewRouter(pool, databases)
	ctx := context.Background()
	dbName := os.Getenv("MYSQL_DATABASE")

	// Create test table
	_, err = router.Exec(ctx, dbName, `
		CREATE TABLE IF NOT EXISTS test_users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Create table error = %v", err)
	}

	// Clean up
	defer router.Exec(ctx, dbName, "DROP TABLE IF EXISTS test_users")

	// Test INSERT
	result, err := router.Exec(ctx, dbName, "INSERT INTO test_users (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("Insert error = %v", err)
	}
	lastID, _ := result.LastInsertId()
	t.Logf("Inserted ID: %d", lastID)

	// Test SELECT
	rows, err := router.Query(ctx, dbName, "SELECT id, name FROM test_users")
	if err != nil {
		t.Fatalf("Query error = %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatalf("Scan error = %v", err)
		}
		t.Logf("Row: id=%d, name=%s", id, name)
		count++
	}
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// Test DELETE
	_, err = router.Exec(ctx, dbName, "DELETE FROM test_users WHERE name = ?", "Alice")
	if err != nil {
		t.Fatalf("Delete error = %v", err)
	}
}

func getTestDSN() string {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		return ""
	}
	port := getEnvInt("MYSQL_PORT", 3306)
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")
	if database == "" {
		database = "testdb"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4",
		user, password, host, port, database)
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		fmt.Sscanf(v, "%d", &i)
		if i > 0 {
			return i
		}
	}
	return defaultVal
}

// TestConnectionTest tests basic MySQL connectivity
func TestConnectionTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dsn := getTestDSN()
	if dsn == "" {
		t.Skip("MySQL connection not available")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	t.Log("MySQL connection successful")
}
