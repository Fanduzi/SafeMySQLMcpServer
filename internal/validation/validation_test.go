package validation

import (
	"strings"
	"testing"
)

func TestValidateDatabaseName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"valid simple", "mydb", false},
		{"valid with underscore", "my_database", false},
		{"valid with numbers", "db123", false},
		{"valid starting underscore", "_private", false},
		{"too long", strings.Repeat("a", 65), true},
		{"max length", strings.Repeat("a", 64), false},
		{"starts with number", "1db", true},
		{"contains hyphen", "my-db", true},
		{"contains space", "my db", true},
		{"contains semicolon", "my;db", true},
		{"contains dot", "my.db", true},
		{"contains special chars", "my@db", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDatabaseName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"valid simple", "users", false},
		{"valid with underscore", "user_accounts", false},
		{"valid with numbers", "users123", false},
		{"too long", strings.Repeat("a", 65), true},
		{"starts with number", "1users", true},
		{"contains hyphen", "user-accounts", true},
		{"contains space", "user accounts", true},
		{"contains semicolon", "users;drop", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTableName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTableName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSQL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"valid select", "SELECT * FROM users", false},
		{"valid insert", "INSERT INTO users (id) VALUES (1)", false},
		{"valid update", "UPDATE users SET name = 'test' WHERE id = 1", false},
		{"valid delete", "DELETE FROM users WHERE id = 1", false},
		{"max length", strings.Repeat("a", 100000), false},
		{"too long", strings.Repeat("a", 100001), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSQL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSearchPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"valid simple", "user", false},
		{"valid with underscore", "user_name", false},
		{"max length", strings.Repeat("a", 256), false},
		{"too long", strings.Repeat("a", 257), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchPattern(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchPattern(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with%percent", `with\%percent`},
		{"with_underscore", `with\_underscore`},
		{"with\\backslash", `with\\backslash`},
		{"combined%_test", `combined\%\_test`},
		{"100%", `100\%`},
		{"user_name", `user\_name`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := EscapeLikePattern(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeLikePattern(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "`users`"},
		{"user`name", "`user``name`"},
		{"table_name", "`table_name`"},
		{"table``name", "`table````name`"},
		{"", "``"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := QuoteIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateAndQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    string
	}{
		{"valid simple", "users", false, "`users`"},
		{"valid with underscore", "user_accounts", false, "`user_accounts`"},
		{"invalid starts with number", "1users", true, ""},
		{"invalid contains hyphen", "user-name", true, ""},
		{"empty", "", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateAndQuoteIdentifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndQuoteIdentifier(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.want {
				t.Errorf("ValidateAndQuoteIdentifier(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

// SQL Injection prevention tests
func TestSQLInjectionPrevention(t *testing.T) {
	maliciousInputs := []string{
		"users; DROP TABLE users;--",
		"users/**/UNION/**/SELECT",
		"users' OR '1'='1",
		"users\" OR \"1\"=\"1",
		"users--; ",
		"users/*comment*/",
		"users\x00null",
	}

	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			err := ValidateTableName(input)
			if err == nil {
				t.Errorf("ValidateTableName(%q) should reject malicious input", input)
			}
		})
	}
}
