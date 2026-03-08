package security

import (
	"strings"
	"testing"

	"github.com/fan/safe-mysql-mcp/internal/config"
	_ "github.com/pingcap/tidb/parser/test_driver" // Required for TiDB parser
)

func TestRewriter_Rewrite(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		rules       *config.SecurityRules
		sql         string
		wantChanged bool
		wantContain string
	}{
		{
			name: "UPDATE without WHERE adds LIMIT",
			rules: &config.SecurityRules{
				AutoLimit: 1000,
			},
			sql:         "UPDATE users SET name = 'test'",
			wantChanged: true,
			wantContain: "LIMIT 1000",
		},
		{
			name: "UPDATE with WHERE no change",
			rules: &config.SecurityRules{
				AutoLimit: 1000,
			},
			sql:         "UPDATE users SET name = 'test' WHERE id = 1",
			wantChanged: false,
		},
		{
			name: "DELETE without WHERE adds LIMIT",
			rules: &config.SecurityRules{
				AutoLimit: 500,
			},
			sql:         "DELETE FROM users",
			wantChanged: true,
			wantContain: "LIMIT 500",
		},
		{
			name: "DELETE with WHERE no change",
			rules: &config.SecurityRules{
				AutoLimit: 500,
			},
			sql:         "DELETE FROM users WHERE id = 1",
			wantChanged: false,
		},
		{
			name: "SELECT with large LIMIT gets capped",
			rules: &config.SecurityRules{
				MaxLimit: 1000,
			},
			sql:         "SELECT * FROM users LIMIT 10000",
			wantChanged: true,
			wantContain: "LIMIT 1000",
		},
		{
			name: "SELECT with small LIMIT no change",
			rules: &config.SecurityRules{
				MaxLimit: 1000,
			},
			sql:         "SELECT * FROM users LIMIT 100",
			wantChanged: false,
		},
		{
			name: "SELECT without LIMIT no change",
			rules: &config.SecurityRules{
				MaxLimit: 1000,
			},
			sql:         "SELECT * FROM users",
			wantChanged: false,
		},
		{
			name: "INSERT no change",
			rules: &config.SecurityRules{
				AutoLimit: 1000,
			},
			sql:         "INSERT INTO users (id) VALUES (1)",
			wantChanged: false,
		},
		{
			name:        "nil rules no change",
			rules:       nil,
			sql:         "UPDATE users SET name = 'test'",
			wantChanged: false,
		},
		{
			name: "UPDATE already has LIMIT",
			rules: &config.SecurityRules{
				AutoLimit: 1000,
			},
			sql:         "UPDATE users SET name = 'test' LIMIT 10",
			wantChanged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriter := NewRewriter(tt.rules)
			parsed, err := parser.Parse(tt.sql)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			result := rewriter.Rewrite(parsed)
			if result.Changed != tt.wantChanged {
				t.Errorf("Rewrite() Changed = %v, want %v (SQL: %s)", result.Changed, tt.wantChanged, result.SQL)
			}
			if tt.wantContain != "" && result.Changed {
				if !strings.Contains(result.SQL, tt.wantContain) {
					t.Errorf("Rewrite() SQL = %q, want to contain %q", result.SQL, tt.wantContain)
				}
			}
		})
	}
}

func TestRewriter_addLimit(t *testing.T) {
	rewriter := &Rewriter{}

	tests := []struct {
		name     string
		sql      string
		limit    int
		expected string
	}{
		{
			name:     "simple update",
			sql:      "UPDATE users SET name = 'test'",
			limit:    100,
			expected: "UPDATE users SET name = 'test' LIMIT 100",
		},
		{
			name:     "delete with semicolon",
			sql:      "DELETE FROM users;",
			limit:    50,
			expected: "DELETE FROM users LIMIT 50",
		},
		{
			name:     "with trailing whitespace",
			sql:      "UPDATE users SET name = 'test'   ",
			limit:    200,
			expected: "UPDATE users SET name = 'test' LIMIT 200",
		},
		{
			name:     "zero limit returns original",
			sql:      "UPDATE users SET name = 'test'",
			limit:    0,
			expected: "UPDATE users SET name = 'test'",
		},
		{
			name:     "negative limit returns original",
			sql:      "UPDATE users SET name = 'test'",
			limit:    -1,
			expected: "UPDATE users SET name = 'test'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.addLimit(tt.sql, tt.limit)
			if result != tt.expected {
				t.Errorf("addLimit() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRewriter_capLimit(t *testing.T) {
	rewriter := &Rewriter{}

	tests := []struct {
		name     string
		sql      string
		maxLimit int
		expected string
	}{
		{
			name:     "limit exceeds max",
			sql:      "SELECT * FROM users LIMIT 10000",
			maxLimit: 1000,
			expected: "SELECT * FROM users LIMIT 1000",
		},
		{
			name:     "limit within max",
			sql:      "SELECT * FROM users LIMIT 100",
			maxLimit: 1000,
			expected: "SELECT * FROM users LIMIT 100",
		},
		{
			name:     "no limit clause",
			sql:      "SELECT * FROM users",
			maxLimit: 1000,
			expected: "SELECT * FROM users",
		},
		{
			name:     "zero max limit returns original",
			sql:      "SELECT * FROM users LIMIT 10000",
			maxLimit: 0,
			expected: "SELECT * FROM users LIMIT 10000",
		},
		{
			name:     "negative max limit returns original",
			sql:      "SELECT * FROM users LIMIT 10000",
			maxLimit: -1,
			expected: "SELECT * FROM users LIMIT 10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rewriter.capLimit(tt.sql, tt.maxLimit)
			if result != tt.expected {
				t.Errorf("capLimit() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRewriter_UpdateRules(t *testing.T) {
	initialRules := &config.SecurityRules{
		AutoLimit: 100,
	}
	newRules := &config.SecurityRules{
		AutoLimit: 500,
	}

	rewriter := NewRewriter(initialRules)
	parser := NewParser()

	// Initially should add LIMIT 100
	parsed, _ := parser.Parse("UPDATE users SET name = 'test'")
	result := rewriter.Rewrite(parsed)
	if !strings.Contains(result.SQL, "LIMIT 100") {
		t.Errorf("Expected LIMIT 100, got: %s", result.SQL)
	}

	// Update rules
	rewriter.UpdateRules(newRules)

	// Now should add LIMIT 500
	parsed, _ = parser.Parse("UPDATE users SET name = 'test'")
	result = rewriter.Rewrite(parsed)
	if !strings.Contains(result.SQL, "LIMIT 500") {
		t.Errorf("Expected LIMIT 500, got: %s", result.SQL)
	}
}
