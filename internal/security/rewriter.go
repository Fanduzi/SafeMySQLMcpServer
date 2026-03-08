// Package security handles SQL parsing and security checks
package security

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

// limitRegex matches LIMIT clause in SQL statements
var limitRegex = regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)`)

// Rewriter rewrites SQL statements for safety
type Rewriter struct {
	rules *config.SecurityRules
}

// NewRewriter creates a new SQL rewriter
func NewRewriter(rules *config.SecurityRules) *Rewriter {
	return &Rewriter{
		rules: rules,
	 }
}

// RewriteResult represents the result of SQL rewriting
type RewriteResult struct {
	SQL     string
	Changed bool
}

// Rewrite rewrites a SQL statement for safety
// Note: This implementation uses string operations with safety validation.
// For production use, consider using AST-based rewriting for more robust handling.
func (r *Rewriter) Rewrite(parsed *ParsedSQL) *RewriteResult {
	if parsed == nil || r.rules == nil {
		return &RewriteResult{SQL: "", Changed: false}
	}

	sql := parsed.Original

	// Add LIMIT to UPDATE/DELETE without WHERE
	if (parsed.Type == SQLTypeUpdate || parsed.Type == SQLTypeDelete) && !parsed.HasWhere {
		if !parsed.HasLimit && r.rules.AutoLimit > 0 {
			sql = r.addLimit(sql, r.rules.AutoLimit)
			return &RewriteResult{SQL: sql, Changed: true}
		}
	}

	// Cap LIMIT for SELECT
	if parsed.Type == SQLTypeSelect && r.rules.MaxLimit > 0 && parsed.HasLimit {
		// Always check and cap the limit if needed
		// The parser may not extract the exact limit value, so we check in the SQL
		newSQL := r.capLimit(sql, r.rules.MaxLimit)
		if newSQL != sql {
			return &RewriteResult{SQL: newSQL, Changed: true}
		}
	}

	return &RewriteResult{SQL: sql, Changed: false}
}

// addLimit adds a LIMIT clause to the SQL
func (r *Rewriter) addLimit(sql string, limit int) string {
    // Validate limit is a positive integer
    if limit <= 0 {
        return sql
    }

    // Remove trailing semicolon and whitespace
    sql = strings.TrimRight(strings.TrimSpace(sql), ";")

    return fmt.Sprintf("%s LIMIT %d", sql, limit)
}

// capLimit caps the LIMIT value in a SELECT statement
func (r *Rewriter) capLimit(sql string, maxLimit int) string {
	// Validate maxLimit is a positive integer
	if maxLimit <= 0 {
		return sql
	}

	return limitRegex.ReplaceAllStringFunc(sql, func(match string) string {
		parts := strings.Fields(match)
		if len(parts) == 2 {
			var limit int
			if _, err := fmt.Sscanf(parts[1], "%d", &limit); err == nil {
				if limit > maxLimit {
					return fmt.Sprintf("LIMIT %d", maxLimit)
				}
				return match
			}
		}
		return match
	})
}

// UpdateRules updates the security rules
func (r *Rewriter) UpdateRules(rules *config.SecurityRules) {
    r.rules = rules
}
