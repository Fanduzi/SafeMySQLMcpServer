package security

import (
	"testing"

	"github.com/fan/safe-mysql-mcp/internal/config"
	_ "github.com/pingcap/tidb/parser/test_driver" // Required for TiDB parser
)

func TestChecker_Check(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name       string
		rules      *config.SecurityRules
		sql        string
		wantAllow  bool
		wantReason string
	}{
		{
			name: "SELECT allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
			},
			sql:       "SELECT * FROM users",
			wantAllow: true,
		},
		{
			name: "INSERT not allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
			},
			sql:        "INSERT INTO users (id) VALUES (1)",
			wantAllow:  false,
			wantReason: "DML operation INSERT is not allowed",
		},
		{
			name: "UPDATE without WHERE triggers auto-rewrite",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT", "UPDATE"},
			},
			sql:       "UPDATE users SET name = 'test'",
			wantAllow: true,
		},
		{
			name: "DELETE without WHERE triggers auto-rewrite",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT", "DELETE"},
			},
			sql:       "DELETE FROM users",
			wantAllow: true,
		},
		{
			name: "DROP is blocked",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
				Blocked:    []string{"DROP"},
			},
			sql:        "DROP TABLE users",
			wantAllow:  false,
			wantReason: "operation DROP is blocked",
		},
		{
			name: "TRUNCATE is blocked",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
				Blocked:    []string{"TRUNCATE"},
			},
			sql:        "TRUNCATE TABLE users",
			wantAllow:  false,
			wantReason: "operation TRUNCATE is blocked",
		},
		{
			name: "SHOW allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
			},
			sql:       "SHOW TABLES",
			wantAllow: true,
		},
		{
			name: "nil rules allow all",
			rules: nil,
			sql:       "SELECT * FROM users",
			wantAllow: true,
		},
		{
			name: "empty SQL not allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
			},
			sql:        "",
			wantAllow:  false,
			wantReason: "empty or invalid SQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewChecker(tt.rules)
			parsed, err := parser.Parse(tt.sql)
			if err != nil && tt.sql != "" {
				t.Fatalf("Parse error: %v", err)
			}

			result := checker.Check(parsed)
			if result.Allowed != tt.wantAllow {
				t.Errorf("Check() Allowed = %v, want %v", result.Allowed, tt.wantAllow)
			}
			if tt.wantReason != "" && result.Reason != tt.wantReason {
				t.Errorf("Check() Reason = %v, want %v", result.Reason, tt.wantReason)
			}
		})
	}
}

func TestChecker_checkDML(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		rules     *config.SecurityRules
		sql       string
		wantAllow bool
		wantAuto  bool
	}{
		{
			name: "SELECT allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
			},
			sql:       "SELECT * FROM users",
			wantAllow: true,
			wantAuto:  false,
		},
		{
			name: "INSERT allowed",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT", "INSERT"},
			},
			sql:       "INSERT INTO users (id) VALUES (1)",
			wantAllow: true,
			wantAuto:  false,
		},
		{
			name: "UPDATE without WHERE needs rewrite",
			rules: &config.SecurityRules{
				AllowedDML: []string{"UPDATE"},
			},
			sql:       "UPDATE users SET name = 'test'",
			wantAllow: true,
			wantAuto:  true,
		},
		{
			name: "UPDATE with WHERE is fine",
			rules: &config.SecurityRules{
				AllowedDML: []string{"UPDATE"},
			},
			sql:       "UPDATE users SET name = 'test' WHERE id = 1",
			wantAllow: true,
			wantAuto:  false,
		},
		{
			name: "DELETE without WHERE needs rewrite",
			rules: &config.SecurityRules{
				AllowedDML: []string{"DELETE"},
			},
			sql:       "DELETE FROM users",
			wantAllow: true,
			wantAuto:  true,
		},
		{
			name: "DELETE with WHERE is fine",
			rules: &config.SecurityRules{
				AllowedDML: []string{"DELETE"},
			},
			sql:       "DELETE FROM users WHERE id = 1",
			wantAllow: true,
			wantAuto:  false,
		},
		{
			name: "UPDATE not in allowed list",
			rules: &config.SecurityRules{
				AllowedDML: []string{"SELECT"},
			},
			sql:       "UPDATE users SET name = 'test'",
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewChecker(tt.rules)
			parsed, err := parser.Parse(tt.sql)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			result := checker.checkDML(parsed)
			if result.Allowed != tt.wantAllow {
				t.Errorf("checkDML() Allowed = %v, want %v", result.Allowed, tt.wantAllow)
			}
			if result.AutoRewrite != tt.wantAuto {
				t.Errorf("checkDML() AutoRewrite = %v, want %v", result.AutoRewrite, tt.wantAuto)
			}
		})
	}
}

func TestChecker_UpdateRules(t *testing.T) {
	initialRules := &config.SecurityRules{
		AllowedDML: []string{"SELECT"},
	}
	newRules := &config.SecurityRules{
		AllowedDML: []string{"SELECT", "INSERT"},
	}

	checker := NewChecker(initialRules)
	parser := NewParser()

	// Initially INSERT should not be allowed
	parsed, _ := parser.Parse("INSERT INTO users (id) VALUES (1)")
	result := checker.Check(parsed)
	if result.Allowed {
		t.Error("INSERT should not be allowed with initial rules")
	}

	// Update rules
	checker.UpdateRules(newRules)

	// Now INSERT should be allowed
	result = checker.Check(parsed)
	if !result.Allowed {
		t.Error("INSERT should be allowed after rules update")
	}
}
