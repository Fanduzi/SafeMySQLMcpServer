// Package auth handles JWT authentication for the MCP server.
// input: JWT_SECRET env var, Authorization header, context
// output: Validator, Claims, context helpers (GetUserID, GetUserEmail)
// pos: authentication layer, validates JWT tokens and extracts user info
// note: if this file changes, update header and internal/auth/README.md
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Security constants
const (
	// MinJWTSecretLength is the minimum required length for JWT secret
	MinJWTSecretLength = 32
)

// Claims represents the JWT claims
type Claims struct {
	UserID    string `json:"sub"`
	UserEmail string `json:"email"`
	jwt.RegisteredClaims
}

// contextKey is the type for context keys
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "userID"
	// UserEmailKey is the context key for user email
	UserEmailKey contextKey = "userEmail"
	// EnvJWTSecret is the environment variable name for JWT secret
	EnvJWTSecret = "JWT_SECRET"
)

var (
	// ErrMissingToken indicates the token is missing
	ErrMissingToken = errors.New("missing authentication token")
	// ErrInvalidToken indicates the token is invalid
	ErrInvalidToken = errors.New("invalid authentication token")
	// ErrTokenExpired indicates the token has expired
	ErrTokenExpired = errors.New("token has expired")
	// ErrSecretTooShort indicates the JWT secret is too short
	ErrSecretTooShort = errors.New("JWT secret too short")
	// ErrSecretNotConfigured indicates the JWT secret is not configured
	ErrSecretNotConfigured = errors.New("JWT secret not configured")
)

// Validator handles JWT validation
type Validator struct {
	secret []byte
}

// NewValidator creates a new JWT validator
func NewValidator(secret string) *Validator {
	return &Validator{
		secret: []byte(secret),
	}
}

// NewValidatorFromEnv creates a JWT validator from environment variable or config
// Priority: environment variable > config value
func NewValidatorFromEnv(configSecret string) (*Validator, error) {
	secret, source, err := GetJWTSecret(configSecret)
	if err != nil {
		return nil, err
	}

	// Warn if secret is loaded from config file
	if source == "config" {
		log.Println("WARNING: JWT secret loaded from config file. Consider using JWT_SECRET environment variable for better security.")
	}

	if err := ValidateJWTSecret(secret); err != nil {
		return nil, err
	}

	return NewValidator(secret), nil
}

// GetJWTSecret retrieves the JWT secret from environment variable or config
// Returns the secret, the source ("env" or "config"), and any error
func GetJWTSecret(configSecret string) (string, string, error) {
	// Priority 1: Environment variable
	if secret := os.Getenv(EnvJWTSecret); secret != "" {
		return secret, "env", nil
	}

	// Priority 2: Config file
	if configSecret != "" {
		return configSecret, "config", nil
	}

	return "", "", ErrSecretNotConfigured
}

// ValidateJWTSecret validates that the JWT secret meets security requirements
func ValidateJWTSecret(secret string) error {
	if len(secret) < MinJWTSecretLength {
		return fmt.Errorf("%w: must be at least %d characters, got %d",
			ErrSecretTooShort, MinJWTSecretLength, len(secret))
	}
	return nil
}

// Validate validates a JWT token and returns the claims
func (v *Validator) Validate(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrMissingToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateToken generates a new JWT token
func (v *Validator) GenerateToken(userID, email string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID:    userID,
		UserEmail: email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(v.secret)
}

// ExtractToken extracts the token from an Authorization header
func ExtractToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// ContextWithUser creates a context with user information
func ContextWithUser(ctx context.Context, userID, userEmail string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UserEmailKey, userEmail)
	return ctx
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}
