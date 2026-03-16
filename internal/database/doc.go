// Package database handles MySQL connection pooling and routing for the MySQL MCP server.
//
// This package manages:
//   - Connection pooling for multiple MySQL clusters
//   - Database routing based on configuration
//   - Query execution with timeout and context support
//   - Graceful connection management during configuration updates
//
// # Connection Pool
//
// The Pool type manages database connections for multiple MySQL clusters.
// Each connection is wrapped in a managedDB that tracks:
//   - Reference count for active queries
//   - Closing state for graceful shutdown
//
// # Graceful Reconnection
//
// When configuration changes (e.g., password rotation), the pool:
//  1. Marks existing connections for graceful close
//  2. Waits for active queries to complete (with timeout)
//  3. Closes connections when reference count reaches zero
//  4. Force closes after timeout (30 seconds)
//
// # Example Usage
//
//	pool, err := database.NewPool(config.Clusters)
//	if err != nil {
//	    // handle error
//	}
//	defer pool.Close()
//
//	// Get connection for a specific cluster
//	db, err := pool.Get("primary")
//	if err != nil {
//	    // handle error
//	}
//
//	// Execute query
//	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
//
// # Configuration Updates
//
// The pool supports hot-reloading of configuration:
//
//	err := pool.UpdateConfig(newConfig)
//	if err != nil {
//	    // handle error
//	}
//
// The will gracefully migrate connections without interrupting active queries.
package database
