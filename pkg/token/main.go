// Package main provides a CLI tool for generating JWT tokens.
// input: CLI flags (user, email, expire, secret), JWT_SECRET env
// output: JWT token string to stdout
// pos: CLI utility, used by admins to generate tokens for API access
// note: if this file changes, update header and pkg/token/README.md
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
)

func main() {
	// Parse command line arguments
	user := flag.String("user", "", "User ID (required)")
	email := flag.String("email", "", "User email (required)")
	expire := flag.Duration("expire", 24*time.Hour, "Token expiration duration (e.g., 24h, 7d, 365d)")
	secret := flag.String("secret", "", "JWT secret (required, or set JWT_SECRET env var)")
	flag.Parse()

	// Validate required arguments
	if *user == "" {
		fmt.Fprintln(os.Stderr, "Error: --user is required")
		flag.Usage()
		os.Exit(1)
	}

	if *email == "" {
		fmt.Fprintln(os.Stderr, "Error: --email is required")
		flag.Usage()
		os.Exit(1)
	}

	// Get secret from flag or environment
	jwtSecret := *secret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: --secret or JWT_SECRET env var is required")
		os.Exit(1)
	}

	// Create validator (which can also generate tokens)
	validator := auth.NewValidator(jwtSecret)

	// Generate token
	token, err := validator.GenerateToken(*user, *email, *expire)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// Output token
	fmt.Println(token)
}
