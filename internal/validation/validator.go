// Package validation provides input validation utilities
package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validation constants
const (
	MaxIdentifierLength = 64
	MaxSQLLength        = 100000 // 100KB
	MaxPatternLength    = 256
)

var (
	// ErrEmptyDatabase indicates database name is empty
	ErrEmptyDatabase = errors.New("database name is required")
	// ErrEmptyTable indicates table name is empty
	ErrEmptyTable = errors.New("table name is required")
	// ErrEmptySQL indicates SQL is empty
	ErrEmptySQL = errors.New("SQL is required")
	// ErrIdentifierTooLong indicates identifier exceeds max length
	ErrIdentifierTooLong = errors.New("identifier too long")
	// ErrInvalidIdentifier indicates identifier contains invalid characters
	ErrInvalidIdentifier = errors.New("invalid identifier: contains disallowed characters")
	// ErrSQLTooLong indicates SQL exceeds max length
	ErrSQLTooLong = errors.New("SQL too long")
	// ErrPatternTooLong indicates search pattern exceeds max length
	ErrPatternTooLong = errors.New("search pattern too long")

	// MySQL identifier pattern: starts with letter or underscore, followed by letters, digits, or underscores
	identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// ValidateDatabaseName validates a database name
func ValidateDatabaseName(name string) error {
	if name == "" {
		return ErrEmptyDatabase
	}
	if len(name) > MaxIdentifierLength {
		return fmt.Errorf("%w: max %d characters, got %d", ErrIdentifierTooLong, MaxIdentifierLength, len(name))
	}
	if !identifierRegex.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidIdentifier, name)
	}
	return nil
}

// ValidateTableName validates a table name
func ValidateTableName(name string) error {
	if name == "" {
		return ErrEmptyTable
	}
	if len(name) > MaxIdentifierLength {
		return fmt.Errorf("%w: max %d characters, got %d", ErrIdentifierTooLong, MaxIdentifierLength, len(name))
	}
	if !identifierRegex.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidIdentifier, name)
	}
	return nil
}

// ValidateSQL validates SQL statement length
func ValidateSQL(sql string) error {
	if sql == "" {
		return ErrEmptySQL
	}
	if len(sql) > MaxSQLLength {
		return fmt.Errorf("%w: max %d bytes, got %d", ErrSQLTooLong, MaxSQLLength, len(sql))
	}
	return nil
}

// ValidateSearchPattern validates and sanitizes a search pattern
func ValidateSearchPattern(pattern string) error {
	if pattern == "" {
		return errors.New("search pattern is required")
	}
	if len(pattern) > MaxPatternLength {
		return ErrPatternTooLong
	}
	return nil
}

// EscapeLikePattern escapes special LIKE pattern characters
func EscapeLikePattern(pattern string) string {
	// Escape LIKE special characters: %, _, \
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(pattern)
}

// QuoteIdentifier safely quotes a MySQL identifier
func QuoteIdentifier(name string) string {
	// Escape backticks by doubling them
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// ValidateAndQuoteIdentifier validates an identifier and returns a safely quoted version
func ValidateAndQuoteIdentifier(name string) (string, error) {
	if err := ValidateDatabaseName(name); err != nil {
		return "", err
	}
	return QuoteIdentifier(name), nil
}
