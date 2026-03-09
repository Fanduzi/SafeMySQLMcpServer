// Package security performs security checks on SQL statements.
// input: ParsedSQL from parser.go, SecurityConfig rules
// output: CheckResult (allowed/blocked, reason, autoRewrite flag)
// pos: security layer, enforces DML/DDL allowlists and blocklists
// note: if this file changes, update header and internal/security/README.md
package security

import (
	"fmt"

	"github.com/fan/safe-mysql-mcp/internal/config"
)

// Checker performs security checks on SQL statements
type Checker struct {
	rules *config.SecurityRules
}

// NewChecker creates a new security checker
func NewChecker(rules *config.SecurityRules) *Checker {
	return &Checker{
		rules: rules,
	}
}

// CheckResult represents the result of a security check
type CheckResult struct {
	Allowed     bool
	Reason      string
	AutoRewrite bool
}

// Check checks if a SQL statement is allowed
func (c *Checker) Check(parsed *ParsedSQL) *CheckResult {
	if parsed == nil {
		return &CheckResult{
			Allowed: false,
			Reason:  "empty or invalid SQL",
		}
	}

	// Check blocked operations first
	if c.isBlocked(parsed) {
		return &CheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("operation %s is blocked", parsed.Type),
		}
	}

	// Check DML operations
	if parsed.Type.IsDML() {
		return c.checkDML(parsed)
	}

	// Check DDL operations
	if parsed.Type.IsDDL() {
		return c.checkDDL(parsed)
	}

	// Allow SHOW and EXPLAIN
	if parsed.Type == SQLTypeShow || parsed.Type == SQLTypeExplain {
		return &CheckResult{Allowed: true}
	}

	// Unknown operation type
	return &CheckResult{
		Allowed: false,
		Reason:  fmt.Sprintf("unknown operation type: %s", parsed.Type),
	}
}

// isBlocked checks if the operation is in the blocked list
func (c *Checker) isBlocked(parsed *ParsedSQL) bool {
	if c.rules == nil {
		return false
	}

	// Map our SQL types to blocked operation names
	blockedOps := map[SQLType]string{
		SQLTypeDrop:     "DROP",
		SQLTypeTruncate: "TRUNCATE",
		SQLTypeRename:   "RENAME",
	}

	if op, ok := blockedOps[parsed.Type]; ok {
		return c.rules.IsBlocked(op)
	}

	return false
}

// checkDML checks if a DML operation is allowed
func (c *Checker) checkDML(parsed *ParsedSQL) *CheckResult {
	if c.rules == nil {
		return &CheckResult{Allowed: true}
	}

	// Check if DML type is allowed
	if !c.rules.IsDMLAllowed(parsed.Type.String()) {
		return &CheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("DML operation %s is not allowed", parsed.Type),
		}
	}

	// Check if rewrite is needed (no WHERE clause for UPDATE/DELETE)
	if (parsed.Type == SQLTypeUpdate || parsed.Type == SQLTypeDelete) && !parsed.HasWhere {
		return &CheckResult{
			Allowed:     true,
			AutoRewrite: true,
			Reason:      "auto LIMIT will be added for safety",
		}
	}

	return &CheckResult{Allowed: true}
}

// checkDDL checks if a DDL operation is allowed
func (c *Checker) checkDDL(parsed *ParsedSQL) *CheckResult {
	if c.rules == nil {
		return &CheckResult{Allowed: true}
	}

	if !c.rules.IsDDLAllowed(string(parsed.Type)) {
		return &CheckResult{
			Allowed: false,
			Reason:  fmt.Sprintf("DDL operation %s is not allowed", parsed.Type),
		}
	}

	return &CheckResult{Allowed: true}
}

// UpdateRules updates the security rules
func (c *Checker) UpdateRules(rules *config.SecurityRules) {
	c.rules = rules
}
