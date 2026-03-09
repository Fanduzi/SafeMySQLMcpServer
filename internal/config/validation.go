// Package config provides configuration validation helpers.
// input: Config, SecurityConfig structs
// output: validation errors, normalized config
// pos: validation layer, ensures config integrity before use
// note: if this file changes, update header and internal/config/README.md
package config

import (
	"fmt"
	"time"
)

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// Validate clusters
	for name, cluster := range c.Clusters {
		if err := cluster.Validate(); err != nil {
			return fmt.Errorf("cluster %s: %w", name, err)
	 }
	 }

	// Validate audit config
	if c.Audit.Enabled {
		if c.Audit.LogFile == "" {
			return fmt.Errorf("audit log file is required when audit is enabled")
		}
		if c.Audit.MaxSizeMB <= 0 {
			return fmt.Errorf("audit max size must be positive")
		}
	}

	return nil
}

// Validate validates the cluster configuration
func (c ClusterConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections should be positive")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections should be non-negative")
	}
	if c.ConnMaxLifetime > 0 && c.ConnMaxLifetime < time.Second {
		return fmt.Errorf("connection max lifetime should be at least 1 second")
	}

	return nil
}

// Validate validates the security rules
func (s *SecurityRules) Validate() error {
	// Check for conflicting DML/DDL rules
	for _, dml := range s.AllowedDML {
		if s.IsBlocked(dml) {
			return fmt.Errorf("DML operation %s is both allowed and blocked", dml)
		}
	}

	for _, ddl := range s.AllowedDDL {
		if s.IsBlocked(ddl) {
			return fmt.Errorf("DDL operation %s is both allowed and blocked", ddl)
		}
	}

	// Validate timeout configuration
	if s.QueryTimeout > 5*time.Minute {
		return fmt.Errorf("query timeout too long: %v (max 5m)", s.QueryTimeout)
	}

	// Validate limit values
	if s.MaxRows > 1000000 {
		return fmt.Errorf("max rows too large: %d (max 1M)", s.MaxRows)
	}

	if s.AutoLimit > 10000 {
		return fmt.Errorf("auto limit too large: %d (max 10000)", s.AutoLimit)
	}

	return nil
}
