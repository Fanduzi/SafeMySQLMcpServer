// Package database handles MySQL connection pooling and routing.
// input: database name, SQL queries, Pool connections
// output: Query results, Exec results, database routing
// pos: data router layer, maps logical db names to physical clusters
// note: if this file changes, update header and internal/database/README.md
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

// useDB switches the connection to the specified database
func useDB(ctx context.Context, conn *sql.Conn, database string) error {
	_, err := conn.ExecContext(ctx, fmt.Sprintf("USE `%s`", database))
	if err != nil {
		return fmt.Errorf("switch to database %s: %w", database, err)
	}
	return nil
}

// Query executes a query on the specified database
func (r *Router) Query(ctx context.Context, database, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := r.GetDB(database)
	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}

	if err := useDB(ctx, conn, database); err != nil {
		conn.Close()
		return nil, err
	}

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		conn.Close()
		return nil, err
	}
	// rows.Close() releases the conn back to pool
	return rows, nil
}

// Exec executes a statement on the specified database
func (r *Router) Exec(ctx context.Context, database, query string, args ...interface{}) (sql.Result, error) {
	db, err := r.GetDB(database)
	if err != nil {
		return nil, err
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()

	if err := useDB(ctx, conn, database); err != nil {
		return nil, err
	}

	return conn.ExecContext(ctx, query, args...)
}
