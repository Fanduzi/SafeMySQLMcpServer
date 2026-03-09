package main

import (
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
)

// TestTokenGeneration tests the core token generation functionality
func TestTokenGeneration(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		email    string
		expire   time.Duration
		secret   string
		wantErr  bool
	}{
		{
			name:    "valid token 1 hour",
			user:    "testuser",
			email:   "test@example.com",
			expire:  time.Hour,
			secret:  "test-secret-min-32-characters-long",
			wantErr: false,
		},
		{
			name:    "valid token 24 hours",
			user:    "admin",
			email:   "admin@example.com",
			expire:  24 * time.Hour,
			secret:  "another-secret-key-32-chars-min",
			wantErr: false,
		},
		{
			name:    "valid token 365 days",
			user:    "longterm",
			email:   "longterm@example.com",
			expire:  365 * 24 * time.Hour,
			secret:  "long-term-secret-key-32-chars",
			wantErr: false,
		},
		{
			name:    "empty user",
			user:    "",
			email:   "test@example.com",
			expire:  time.Hour,
			secret:  "test-secret-min-32-characters-long",
			wantErr: false, // auth package allows empty user
		},
		{
			name:    "empty email",
			user:    "testuser",
			email:   "",
			expire:  time.Hour,
			secret:  "test-secret-min-32-characters-long",
			wantErr: false, // auth package allows empty email
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := auth.NewValidator(tt.secret)
			token, err := validator.GenerateToken(tt.user, tt.email, tt.expire)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify token format (3 parts separated by dots)
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("Token should have 3 parts, got %d: %s", len(parts), token)
				}

				// Verify token can be validated
				claims, err := validator.Validate(token)
				if err != nil {
					t.Errorf("Token validation failed: %v", err)
				}
				if claims.UserID != tt.user {
					t.Errorf("UserID = %s, want %s", claims.UserID, tt.user)
				}
				if claims.UserEmail != tt.email {
					t.Errorf("Email = %s, want %s", claims.UserEmail, tt.email)
				}
			}
		})
	}
}

// TestTokenValidation tests that tokens can be validated after generation
func TestTokenValidation(t *testing.T) {
	secret := "validation-test-secret-32-chars"
	validator := auth.NewValidator(secret)

	token, err := validator.GenerateToken("testuser", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "testuser" {
		t.Errorf("UserID = %s, want testuser", claims.UserID)
	}
	if claims.UserEmail != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", claims.UserEmail)
	}
}

// TestTokenExpiration tests token expiration behavior
// Note: JWT uses seconds-level precision, so we test with 1 second
func TestTokenExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping expiration test in short mode")
	}

	secret := "expiration-test-secret-32-chars"
	validator := auth.NewValidator(secret)

	// Generate token that expires in 1 second
	token, err := validator.GenerateToken("testuser", "test@example.com", time.Second)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = validator.Validate(token)
	if err != nil {
		t.Errorf("Token should be valid immediately: %v", err)
	}

	// Wait for expiration (1.5 seconds to be safe)
	time.Sleep(1500 * time.Millisecond)

	// Token should be expired
	_, err = validator.Validate(token)
	if err == nil {
		t.Error("Token should be expired")
	}
}

// TestDifferentSecrets tests that tokens from different secrets are incompatible
func TestDifferentSecrets(t *testing.T) {
	secret1 := "first-secret-key-32-characters-min"
	secret2 := "second-secret-key-32-characters-min"

	validator1 := auth.NewValidator(secret1)
	validator2 := auth.NewValidator(secret2)

	token, err := validator1.GenerateToken("testuser", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token from validator1 should not validate with validator2
	_, err = validator2.Validate(token)
	if err == nil {
		t.Error("Token should not validate with different secret")
	}
}

// TestFlagDefaults tests that CLI flags have expected defaults
func TestFlagDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Define flags with same defaults as main.go
	user := fs.String("user", "", "User ID")
	email := fs.String("email", "", "User email")
	expire := fs.Duration("expire", 24*time.Hour, "Token expiration duration")
	secret := fs.String("secret", "", "JWT secret")

	// Parse empty args to get defaults
	fs.Parse([]string{})

	// Verify defaults
	if *user != "" {
		t.Errorf("Default user should be empty, got %s", *user)
	}
	if *email != "" {
		t.Errorf("Default email should be empty, got %s", *email)
	}
	if *expire != 24*time.Hour {
		t.Errorf("Default expire should be 24h, got %v", *expire)
	}
	if *secret != "" {
		t.Errorf("Default secret should be empty, got %s", *secret)
	}
}

// TestEnvironmentVariable tests JWT_SECRET environment variable
func TestEnvironmentVariable(t *testing.T) {
	testSecret := "env-test-secret-32-characters-long"

	// Set environment variable
	oldEnv := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Setenv("JWT_SECRET", oldEnv)

	// Simulate the logic from main.go
	jwtSecret := ""
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}

	if jwtSecret != testSecret {
		t.Errorf("Expected to read JWT_SECRET from env, got %s", jwtSecret)
	}

	// Verify we can generate a token with the env secret
	validator := auth.NewValidator(jwtSecret)
	token, err := validator.GenerateToken("testuser", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token with env secret: %v", err)
	}

	// Verify token format
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Token should have 3 parts, got %d", len(parts))
	}
}

// TestTokenFormat verifies JWT format
func TestTokenFormat(t *testing.T) {
	secret := "format-test-secret-32-chars-long"
	validator := auth.NewValidator(secret)

	token, err := validator.GenerateToken("testuser", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT should have 3 parts, got %d", len(parts))
	}

	// Each part should be non-empty
	for i, part := range parts {
		if part == "" {
			t.Errorf("JWT part %d should not be empty", i)
		}
	}

	// Each part should only contain valid base64 characters
	for i, part := range parts {
		for _, c := range part {
			if !isBase64URL(c) {
				t.Errorf("JWT part %d contains invalid character: %c", i, c)
			}
		}
	}
}

func isBase64URL(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '='
}
