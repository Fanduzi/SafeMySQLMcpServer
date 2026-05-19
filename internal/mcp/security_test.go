package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/fan/safe-mysql-mcp/internal/audit"
	"github.com/fan/safe-mysql-mcp/internal/config"
	"github.com/fan/safe-mysql-mcp/internal/security"
	_ "github.com/pingcap/tidb/parser/test_driver"
)

// --- Checker-level tests: verify blocklist independent of DDL allowlist ---

func TestChecker_BlockedList_Drop(t *testing.T) {
	parser := security.NewParser()
	checker := newBlocklistTestChecker()
	parsed, err := parser.Parse("DROP TABLE users")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result := checker.Check(parsed)
	if result.Allowed {
		t.Error("DROP should be blocked by blocklist")
	}
	if !strings.Contains(result.Reason, "blocked") {
		t.Errorf("reason should mention blocked, got: %s", result.Reason)
	}
}

func TestChecker_BlockedList_Truncate(t *testing.T) {
	parser := security.NewParser()
	checker := newBlocklistTestChecker()
	parsed, err := parser.Parse("TRUNCATE TABLE users")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result := checker.Check(parsed)
	if result.Allowed {
		t.Error("TRUNCATE should be blocked by blocklist")
	}
}

func TestChecker_BlockedList_Rename(t *testing.T) {
	parser := security.NewParser()
	checker := newBlocklistTestChecker()
	parsed, err := parser.Parse("RENAME TABLE users TO accounts")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result := checker.Check(parsed)
	if result.Allowed {
		t.Error("RENAME should be blocked by blocklist")
	}
}

// newBlocklistTestChecker creates a checker where DROP is BOTH in the blocklist
// AND in the DDL allowlist. If the blocklist works, DROP is still rejected.
// This ensures the test catches blocklist bypass — not DDL allowlist catching it.
func newBlocklistTestChecker() *security.Checker {
	return security.NewChecker(&config.SecurityRules{
		AllowedDML: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		AllowedDDL: []string{"CREATE_TABLE", "ALTER_TABLE", "DROP"}, // DROP intentionally allowed in DDL
		Blocked:    []string{"DROP", "TRUNCATE", "RENAME"},          // but blocked by blocklist
		AutoLimit:  1000,
		MaxLimit:   10000,
		MaxRows:    10000,
	})
}

// --- Handler-level tests (integrated path through executeQuery) ---

func TestExecuteQuery_SecurityBlock_Drop(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeQuery(context.Background(), "testdb", "DROP TABLE users")
	if err == nil {
		t.Fatal("expected DROP to be blocked")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Errorf("error should mention blocked, got: %v", err)
	}
}

func TestExecuteQuery_SecurityBlock_Truncate(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeQuery(context.Background(), "testdb", "TRUNCATE TABLE users")
	if err == nil {
		t.Fatal("expected TRUNCATE to be blocked")
	}
}

func TestExecuteQuery_SecurityBlock_Rename(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeQuery(context.Background(), "testdb", "RENAME TABLE users TO accounts")
	if err == nil {
		t.Fatal("expected RENAME to be blocked")
	}
}

func TestExecuteQuery_ParseError(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeQuery(context.Background(), "testdb", "NOT VALID SQL !!!")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention parse, got: %v", err)
	}
}

func TestExecuteExplain_RejectsDrop(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeExplain(context.Background(), "testdb", "DROP TABLE users")
	if err == nil {
		t.Fatal("expected EXPLAIN DROP to be rejected")
	}
}

func TestExecuteExplain_RejectsCreate(t *testing.T) {
	h := newTestHandler(t)
	_, err := h.executeExplain(context.Background(), "testdb", "CREATE TABLE t (id INT)")
	if err == nil {
		t.Fatal("expected EXPLAIN CREATE to be rejected")
	}
}

func TestGetQueryTimeout_Default(t *testing.T) {
	h := newTestHandler(t)
	if timeout := h.getQueryTimeout(); timeout <= 0 {
		t.Errorf("getQueryTimeout() = %v, want positive", timeout)
	}
}

func TestGetMaxRows_Default(t *testing.T) {
	h := newTestHandler(t)
	if rows := h.getMaxRows(); rows <= 0 {
		t.Errorf("getMaxRows() = %d, want positive", rows)
	}
}

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	parser := security.NewParser()
	checker := security.NewChecker(&config.SecurityRules{
		AllowedDML: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		AllowedDDL: []string{"CREATE_TABLE", "CREATE_INDEX", "ALTER_TABLE"},
		Blocked:    []string{"DROP", "TRUNCATE", "RENAME"},
		AutoLimit:  1000,
		MaxLimit:   10000,
		MaxRows:    10000,
	})
	rewriter := security.NewRewriter(&config.SecurityRules{
		AutoLimit: 1000,
		MaxLimit:  10000,
	})
	auditLogger, err := audit.NewLogger(&config.AuditConfig{Enabled: false})
	if err != nil {
		t.Fatalf("audit logger: %v", err)
	}
	reloadCfg := &config.ReloadableConfig{}
	reloadCfg.Update(&config.Config{}, &config.SecurityConfig{
		Security: config.SecurityRules{QueryTimeout: 30e9, MaxRows: 10000},
	})

	return NewHandler(nil, parser, checker, rewriter, auditLogger, reloadCfg)
}
