// Package config handles configuration loading and management
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the main configuration structure
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Clusters  ClustersConfig  `yaml:"clusters"`
	Databases DatabasesConfig `yaml:"databases"`
	Security  SecurityRef     `yaml:"security"`
	Audit     AuditConfig     `yaml:"audit"`
}

// ServerConfig holds server settings
type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	JWTSecret string `yaml:"jwt_secret"`
}

// ClustersConfig maps cluster names to their configurations
type ClustersConfig map[string]ClusterConfig

// ClusterConfig holds database cluster connection settings
type ClusterConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// DatabasesConfig maps database names to their cluster assignments
type DatabasesConfig map[string]DatabaseConfig

// DatabaseConfig holds database routing information
type DatabaseConfig struct {
	Cluster string `yaml:"cluster"`
}

// SecurityRef references the security configuration file
type SecurityRef struct {
	ConfigFile string `yaml:"config_file"`
}

// SecurityConfig holds security rules
type SecurityConfig struct {
	Security SecurityRules `yaml:"security"`
}

// SecurityRules defines allowed and blocked operations
type SecurityRules struct {
	AllowedDML   []string      `yaml:"allowed_dml"`
	AllowedDDL   []string      `yaml:"allowed_ddl"`
	Blocked      []string      `yaml:"blocked"`
	AutoLimit    int           `yaml:"auto_limit"`
	MaxLimit     int           `yaml:"max_limit"`
	QueryTimeout time.Duration `yaml:"query_timeout"`
	MaxRows      int           `yaml:"max_rows"`
}

// AuditConfig holds audit logging settings
type AuditConfig struct {
	Enabled      bool   `yaml:"enabled"`
	LogFile      string `yaml:"log_file"`
	MaxSQLLength int    `yaml:"max_sql_length"`
	MaxSizeMB    int    `yaml:"max_size_mb"`
	MaxBackups   int    `yaml:"max_backups"`
	MaxAgeDays   int    `yaml:"max_age_days"`
	Compress     bool   `yaml:"compress"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Expand environment variables
	expanded := expandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Set defaults
	setDefaults(&cfg)

	return &cfg, nil
}

// LoadSecurity loads security configuration from a YAML file
func LoadSecurity(path string) (*SecurityConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read security config file: %w", err)
	}

	var cfg SecurityConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse security config file: %w", err)
	}

	return &cfg, nil
}

// expandEnv expands environment variables in the format ${VAR} or $VAR
func expandEnv(s string) string {
	// Match ${VAR} pattern
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1]
		return os.Getenv(varName)
	})
}

// setDefaults sets default values for optional configuration fields
func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	for name, cluster := range cfg.Clusters {
		if cluster.Port == 0 {
			cluster.Port = 3306
			cfg.Clusters[name] = cluster
		}
		if cluster.MaxOpenConns == 0 {
			cluster.MaxOpenConns = 10
			cfg.Clusters[name] = cluster
		}
		if cluster.MaxIdleConns == 0 {
			cluster.MaxIdleConns = 5
			cfg.Clusters[name] = cluster
		}
		if cluster.ConnMaxLifetime == 0 {
			cluster.ConnMaxLifetime = 5 * time.Minute
			cfg.Clusters[name] = cluster
		}
	}

	if cfg.Audit.MaxSQLLength == 0 {
		cfg.Audit.MaxSQLLength = 2000
	}
	if cfg.Audit.MaxSizeMB == 0 {
		cfg.Audit.MaxSizeMB = 100
	}
	if cfg.Audit.MaxBackups == 0 {
		cfg.Audit.MaxBackups = 10
	}
	if cfg.Audit.MaxAgeDays == 0 {
		cfg.Audit.MaxAgeDays = 30
	}
}

// GetDSN returns the MySQL DSN for a cluster
func (c ClusterConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=Local&charset=utf8mb4",
		c.Username, c.Password, c.Host, c.Port)
}

// IsBlocked checks if an operation is in the blocked list
func (s *SecurityRules) IsBlocked(op string) bool {
	opUpper := strings.ToUpper(op)
	for _, blocked := range s.Blocked {
		if strings.ToUpper(blocked) == opUpper {
			return true
		}
	}
	return false
}

// IsDMLAllowed checks if a DML operation is allowed
func (s *SecurityRules) IsDMLAllowed(op string) bool {
	opUpper := strings.ToUpper(op)
	for _, allowed := range s.AllowedDML {
		if strings.ToUpper(allowed) == opUpper {
			return true
		}
	}
	return false
}

// IsDDLAllowed checks if a DDL operation is allowed
func (s *SecurityRules) IsDDLAllowed(op string) bool {
	opUpper := strings.ToUpper(op)
	for _, allowed := range s.AllowedDDL {
		if strings.ToUpper(allowed) == opUpper {
			return true
		}
	}
	return false
}
