package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{Port: 8080},
				Clusters: ClustersConfig{
					"primary": ClusterConfig{
						Host:         "localhost",
						Port:         3306,
						Username:     "root",
						MaxOpenConns: 10,
						MaxIdleConns: 5,
						ConnMaxLifetime: 5 * time.Minute,
					},
				},
				Audit: AuditConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "invalid port - negative",
			config: &Config{
				Server: ServerConfig{Port: -1},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{Port: 70000},
			},
			wantErr: true,
		},
		{
			name: "audit enabled without log file",
			config: &Config{
				Server: ServerConfig{Port: 8080},
				Audit: AuditConfig{Enabled: true, LogFile: ""},
			},
			wantErr: true,
		},
		{
			name: "audit enabled with invalid max size",
			config: &Config{
				Server: ServerConfig{Port: 8080},
				Audit: AuditConfig{Enabled: true, LogFile: "/tmp/audit.log", MaxSizeMB: 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClusterConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ClusterConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ClusterConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "root",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: ClusterConfig{
				Host:     "",
				Port:     3306,
				Username: "root",
			},
			wantErr: true,
		},
		{
			name: "invalid port - zero",
			config: ClusterConfig{
				Host:     "localhost",
				Port:     0,
				Username: "root",
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: ClusterConfig{
				Host:     "localhost",
				Port:     70000,
				Username: "root",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			config: ClusterConfig{
				Host:     "localhost",
				Port:     3306,
				Username: "",
			},
			wantErr: true,
		},
		{
			name: "invalid max open conns",
			config: ClusterConfig{
				Host:         "localhost",
				Port:         3306,
				Username:     "root",
				MaxOpenConns: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid max idle conns",
			config: ClusterConfig{
				Host:         "localhost",
				Port:         3306,
				Username:     "root",
				MaxOpenConns: 10,
				MaxIdleConns: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid conn max lifetime",
			config: ClusterConfig{
				Host:            "localhost",
				Port:            3306,
				Username:        "root",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 500 * time.Millisecond,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecurityRules_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  SecurityRules
		wantErr bool
	}{
		{
			name: "valid config",
			config: SecurityRules{
				AllowedDML:     []string{"SELECT", "INSERT"},
				AllowedDDL:     []string{"CREATE_TABLE"},
				QueryTimeout:   30 * time.Second,
				MaxRows:        10000,
				AutoLimit:      1000,
			},
			wantErr: false,
		},
		{
			name: "conflicting DML rules",
			config: SecurityRules{
				AllowedDML: []string{"SELECT"},
				Blocked:    []string{"SELECT"},
			},
			wantErr: true,
		},
		{
			name: "conflicting DDL rules",
			config: SecurityRules{
				AllowedDDL: []string{"CREATE_TABLE"},
				Blocked:     []string{"CREATE_TABLE"},
			},
			wantErr: true,
		},
		{
			name: "query timeout too long",
			config: SecurityRules{
				QueryTimeout: 10 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "max rows too large",
			config: SecurityRules{
				MaxRows: 2000000,
			},
			wantErr: true,
		},
		{
			name: "auto limit too large",
			config: SecurityRules{
				AutoLimit: 50000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			 err := tt.config.Validate()
		 if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
