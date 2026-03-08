// Package database handles MySQL connection pooling and routing
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

// Router routes database requests to the appropriate cluster
type Router struct {
	mu        sync.RWMutex
	pool      *Pool
	databases config.DatabasesConfig
}

// NewRouter creates a new database router
func NewRouter(pool *Pool, databases config.DatabasesConfig) *Router {
	return &Router{
		pool:      pool,
		databases: databases,
	}
}

// GetDB returns the database connection for a specific database
func (r *Router) GetDB(database string) (*sql.DB, error) {
	r.mu.RLock()
	dbConfig, ok := r.databases[database]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown database: %s", database)
	}

	return r.pool.Get(dbConfig.Cluster)
}

// GetCluster returns the cluster name for a database
func (r *Router) GetCluster(database string) (string, error) {
	r.mu.RLock()
	dbConfig, ok := r.databases[database]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("unknown database: %s", database)
	}

	return dbConfig.Cluster, nil
}

// UpdateConfig updates the router configuration
func (r *Router) UpdateConfig(databases config.DatabasesConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.databases = databases
}

// ListDatabases returns all configured databases
func (r *Router) ListDatabases() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	databases := make([]string, 0, len(r.databases))
	for name := range r.databases {
		databases = append(databases, name)
	}
	return databases
}

// Query executes a query on the specified database
func (r *Router) Query(ctx context.Context, database, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := r.GetDB(database)
	if err != nil {
		return nil, err
	}

	return db.QueryContext(ctx, query, args...)
}

// Exec executes a statement on the specified database
func (r *Router) Exec(ctx context.Context, database, query string, args ...interface{}) (sql.Result, error) {
	db, err := r.GetDB(database)
	if err != nil {
		return nil, err
	}

	return db.ExecContext(ctx, query, args...)
}
