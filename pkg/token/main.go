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
	"strconv"
	"strings"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/auth"
)

// parseDuration extends time.ParseDuration to support "d" (days) suffix.
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.ParseFloat(daysStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days * 24 * float64(time.Hour)), nil
	}
	return time.ParseDuration(s)
}

func main() {
	user := flag.String("user", "", "User ID (required)")
	email := flag.String("email", "", "User email (required)")
	expireStr := flag.String("expire", "24h", "Token expiration duration (e.g., 24h, 7d, 365d)")
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

	expire, err := parseDuration(*expireStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid --expire value: %v\n", err)
		os.Exit(1)
	}
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: --secret or JWT_SECRET env var is required")
		os.Exit(1)
	}

	validator := auth.NewValidator(jwtSecret)

	token, err := validator.GenerateToken(*user, *email, expire)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// Output token
	fmt.Println(token)
}
