// Package database handles MySQL connection pooling and routing.
// input: config.ClustersConfig (host, port, credentials)
// output: *sql.DB connections, Query/Exec helpers
// pos: data layer, manages MySQL connections for all queries
// note: if this file changes, update header and internal/database/README.md
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

// managedDB wraps a sql.DB with reference counting
type managedDB struct {
	db       *sql.DB
	refCount int32
	closing  int32
}

// Pool manages database connections for multiple clusters
type Pool struct {
	mu       sync.RWMutex
	clusters map[string]*managedDB
	configs  config.ClustersConfig
}

// NewPool creates a new connection pool
func NewPool(clusters config.ClustersConfig) (*Pool, error) {
	p := &Pool{
		clusters: make(map[string]*managedDB),
		configs:  clusters,
	}

	// Initialize connections for all clusters
	for name, cfg := range clusters {
		db, err := p.connect(cfg)
		if err != nil {
			_ = p.Close()
			return nil, fmt.Errorf("connect to cluster %s: %w", name, err)
		}
		p.clusters[name] = &managedDB{db: db}
	}

	return p, nil
}

// connect creates a new database connection
func (p *Pool) connect(cfg config.ClusterConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// Get returns the database connection for a cluster
func (p *Pool) Get(cluster string) (*sql.DB, error) {
	p.mu.RLock()
	mdb, ok := p.clusters[cluster]
	p.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown cluster: %s", cluster)
	}

	// Check if closing
	if atomic.LoadInt32(&mdb.closing) == 1 {
		return nil, fmt.Errorf("cluster %s is being closed", cluster)
	}

	// Increment reference count
	atomic.AddInt32(&mdb.refCount, 1)

	return mdb.db, nil
}

// release decrements the reference count
func (p *Pool) release(db *sql.DB) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, mdb := range p.clusters {
		if mdb.db == db {
			atomic.AddInt32(&mdb.refCount, -1)
		}
	}
}

// UpdateConfig updates the pool configuration with graceful connection handling
func (p *Pool) UpdateConfig(clusters config.ClustersConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Track clusters to close
	toClose := make(map[string]bool)
	for name := range p.clusters {
		if _, ok := clusters[name]; !ok {
			toClose[name] = true
		}
	}

	// Mark all existing connections for graceful close
	for name, mdb := range p.clusters {
		if toClose[name] {
			// Mark for closing - no new references allowed
			atomic.StoreInt32(&mdb.closing, 1)
		}
	}

	// Create or update connections
	for name, cfg := range clusters {
		mdb, exists := p.clusters[name]

		if exists {
			// Check if config changed
			oldCfg := p.configs[name]
			if oldCfg.Host != cfg.Host || oldCfg.Port != cfg.Port ||
				oldCfg.Username != cfg.Username || oldCfg.Password != cfg.Password {
				// Config changed, need to reconnect
				atomic.StoreInt32(&mdb.closing, 1)
				toClose[name] = true
			}
		} else {
			// New cluster
			db, err := p.connect(cfg)
			if err != nil {
				log.Printf("Failed to connect to cluster %s: %v", name, err)
				continue
			}
			p.clusters[name] = &managedDB{db: db}
		}
	}

	// Wait for connections to be released (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		p.mu.Lock()
		for name, mdb := range p.clusters {
			if atomic.LoadInt32(&mdb.refCount) == 0 && atomic.LoadInt32(&mdb.closing) == 1 {
				if err := mdb.db.Close(); err != nil {
					log.Printf("Error closing cluster %s: %v", name, err)
				}
				delete(p.clusters, name)
			}
		}
		p.mu.Unlock()

		// Check if all marked connections are closed
		allClosed := true
		for name := range toClose {
			if _, ok := p.clusters[name]; ok {
				allClosed = false
				break
			}
		}
		if allClosed {
			break
		}

		select {
		case <-ctx.Done():
			log.Printf("Timeout waiting for connections to close, forcing close")
			p.mu.Lock()
			for name := range toClose {
				if mdb, ok := p.clusters[name]; ok {
					if err := mdb.db.Close(); err != nil {
						log.Printf("Error force-closing cluster %s: %v", name, err)
					}
					delete(p.clusters, name)
				}
			}
			p.mu.Unlock()
			break
		case <-time.After(100 * time.Millisecond):
		}
	}

	p.configs = clusters
	return nil
}

// Close closes all database connections
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for name, mdb := range p.clusters {
		// Force close regardless of reference count
		if err := mdb.db.Close(); err != nil {
			lastErr = fmt.Errorf("close cluster %s: %w", name, err)
		}
	}

	p.clusters = make(map[string]*managedDB)
	return lastErr
}
