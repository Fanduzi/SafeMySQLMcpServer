package auth

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	secret := "test-secret-key-min-32-characters!!"
	validator := NewValidator(secret)

	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidator_Validate(t *testing.T) {
	secret := "test-secret-key-min-32-characters!!"
	validator := NewValidator(secret)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			token:   "not-a-valid-token",
			wantErr: true,
		},
		{
			name:    "random string",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := validator.Validate(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && claims == nil {
				t.Error("Validate() returned nil claims")
			}
		})
	}
}

func TestValidator_GenerateAndValidate(t *testing.T) {
	secret := "test-secret-key-min-32-characters!!"
	validator := NewValidator(secret)

	// Generate a valid token
	token, err := validator.GenerateToken("user123", "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	// Validate the token
	claims, err := validator.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Claims.UserID = %v, want user123", claims.UserID)
	}

	if claims.UserEmail != "test@example.com" {
		t.Errorf("Claims.UserEmail = %v, want test@example.com", claims.UserEmail)
	}
}

func TestValidator_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-min-32-characters!!"
	validator := NewValidator(secret)

	// Generate an already expired token
	token, err := validator.GenerateToken("user123", "test@example.com", -time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Validate should fail
	_, err = validator.Validate(token)
	if err == nil {
		t.Error("Validate() should fail for expired token")
	}
	if err != ErrTokenExpired {
		t.Errorf("Validate() error = %v, want ErrTokenExpired", err)
	}
}

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid secret",
			secret:  "this-is-a-valid-secret-with-32-chars!",
			wantErr: false,
		},
		{
			name:    "exactly 32 chars",
			secret:  "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "31 chars - too short",
			secret:  "1234567890123456789012345678901",
			wantErr: true,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJWTSecret(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWTSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateToken_TooShortSecret(t *testing.T) {
	// This tests the GenerateToken function - note that it doesn't validate secret length
	// Secret validation happens at NewValidatorFromEnv level
	shortSecret := "short"
	validator := NewValidator(shortSecret)

	token, err := validator.GenerateToken("user123", "test@example.com", time.Hour)
	if err != nil {
		t.Errorf("GenerateToken() with short secret error = %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}
}

func TestNewValidatorFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		configValue string
		wantErr     bool
	}{
		{
			name:        "from env variable",
			envValue:    "env-secret-key-with-at-least-32-chars!",
			configValue: "",
			wantErr:     false,
		},
		{
			name:        "from config",
			envValue:    "",
			configValue: "config-secret-key-with-at-least-32-chars!",
			wantErr:     false,
		},
		{
			name:        "both empty - error",
			envValue:    "",
			configValue: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("JWT_SECRET", tt.envValue)
				defer os.Unsetenv("JWT_SECRET")
			} else {
				os.Unsetenv("JWT_SECRET")
			}

			validator, err := NewValidatorFromEnv(tt.configValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewValidatorFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && validator == nil {
				t.Error("NewValidatorFromEnv() returned nil validator")
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer mytoken123",
			want:       "mytoken123",
		},
		{
			name:       "lowercase bearer",
			authHeader: "bearer mytoken123",
			want:       "mytoken123",
		},
		{
			name:       "no bearer prefix",
			authHeader: "mytoken123",
			want:       "",
		},
		{
			name:       "empty header",
			authHeader: "",
			want:       "",
		},
		{
			name:       "wrong auth type",
			authHeader: "Basic mytoken123",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractToken(tt.authHeader)
			if got != tt.want {
				t.Errorf("ExtractToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextWithUser(t *testing.T) {
	ctx := ContextWithUser(context.Background(), "user123", "test@example.com")

	if GetUserID(ctx) != "user123" {
		t.Errorf("GetUserID() = %v, want user123", GetUserID(ctx))
	}

	if GetUserEmail(ctx) != "test@example.com" {
		t.Errorf("GetUserEmail() = %v, want test@example.com", GetUserEmail(ctx))
	}
}
