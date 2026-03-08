// Package constants defines shared constants for the application
package constants

import (
	"time"
)

// MySQL identifier limits
const (
	// MaxIdentifierLength is the maximum length for MySQL identifiers
	MaxIdentifierLength = 64

	// MaxTableNameLength is the maximum length for table names
	MaxTableNameLength = 64
)

// Query limits
const (
	// DefaultMaxRows is the default maximum number of rows to return
	DefaultMaxRows = 10000

	// MaxRowsLimit is the absolute maximum number of rows allowed
	MaxRowsLimit = 1000000

	// DefaultQueryTimeout is the default query timeout
	DefaultQueryTimeout = 30 * time.Second

	// MaxQueryTimeout is the maximum allowed query timeout
	MaxQueryTimeout = 5 * time.Minute

	// MaxSQLLength is the maximum SQL statement length (100KB)
	MaxSQLLength = 100000
)

// Rate limiting
const (
	// DefaultRateLimit is the default requests per second
	DefaultRateLimit = 10

	// DefaultRateBurst is the default burst capacity
	DefaultRateBurst = 20

	// MaxRateLimit is the maximum allowed rate limit
	MaxRateLimit = 1000
)

// Connection pool limits
const (
	// DefaultMaxOpenConns is the default maximum open connections
	DefaultMaxOpenConns = 10

	// DefaultMaxIdleConns is the default maximum idle connections
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default connection max lifetime
	DefaultConnMaxLifetime = 5 * time.Minute
)

// Security constants
const (
	// MinJWTSecretLength is the minimum required length for JWT secret
	MinJWTSecretLength = 32

	// DefaultAutoLimit is the default LIMIT added to queries without WHERE
	DefaultAutoLimit = 1000

	// MaxAutoLimit is the maximum auto-limit value
	MaxAutoLimit = 10000
)

// Audit log limits
const (
	// DefaultMaxSQLLength is the default maximum SQL length in audit logs
	DefaultMaxSQLLength = 2000

	// DefaultMaxLogSizeMB is the default maximum log file size in MB
	DefaultMaxLogSizeMB = 100

	// DefaultMaxBackups is the default maximum number of backup files
	DefaultMaxBackups = 10

	// DefaultMaxAgeDays is the default maximum age of backup files in days
	DefaultMaxAgeDays = 30
)

// HTTP server constants
const (
	// DefaultReadTimeout is the default read timeout for HTTP server
	DefaultReadTimeout = 30 * time.Second

	// DefaultWriteTimeout is the default write timeout for HTTP server
	DefaultWriteTimeout = 60 * time.Second

	// DefaultIdleTimeout is the default idle timeout for HTTP server
	DefaultIdleTimeout = 120 * time.Second
)
