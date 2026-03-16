// Package metrics tests Prometheus metric registration and recording.
// input: metric operations and HTTP scrape requests
// output: assertions on exported Prometheus metrics
// pos: test layer for internal metrics behavior
// note: if this file changes, update header and internal/metrics/README.md
package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	instance = nil

	m := Init("testapp")
	if m == nil {
		t.Fatal("Init returned nil")
	}
	if m.RequestsTotal == nil {
		t.Error("RequestsTotal not initialized")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration not initialized")
	}
	if m.QueriesTotal == nil {
		t.Error("QueriesTotal not initialized")
	}
	if m.MCPCallsTotal == nil {
		t.Error("MCPCallsTotal not initialized")
	}
}

func TestGet(t *testing.T) {
	instance = nil

	m := Get()
	if m == nil {
		t.Fatal("Get returned nil")
	}
}

func TestHandler(t *testing.T) {
	instance = nil
	m := Init("testapp")

	handler := m.Handler()
	if handler == nil {
		t.Fatal("Handler returned nil")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Handler returned status %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRecordRequest(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordRequest("GET", "/api/query", http.StatusOK, 100*time.Millisecond)

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_http_requests_total") {
		t.Error("expected http_requests_total metric in output")
	}
	if !strings.Contains(body, `status="200"`) {
		t.Fatalf("expected status label 200 in metrics output, got: %s", body)
	}
}

func TestRecordQuery(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordQuery("mydb", "SELECT", 50*time.Millisecond, 100)

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_db_queries_total") {
		t.Error("expected db_queries_total metric in output")
	}
	if !strings.Contains(body, "testapp_db_query_rows") {
		t.Error("expected db_query_rows metric in output")
	}
}

func TestRecordSecurityViolation(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordSecurityViolation("sql_injection")

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_security_violations_total") {
		t.Error("expected security_violations_total metric in output")
	}
}

func TestRecordBlockedQuery(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordBlockedQuery("mydb", "unsafe_operation")

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_security_blocked_queries_total") {
		t.Error("expected security_blocked_queries_total metric in output")
	}
}

func TestRecordRateLimitExceeded(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordRateLimitExceeded("192.168.1.1")

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_rate_limit_exceeded_total") {
		t.Error("expected rate_limit_exceeded_total metric in output")
	}
}

func TestRecordAuthAttempt(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordAuthAttempt("jwt", true)
	m.RecordAuthAttempt("jwt", false)

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_auth_attempts_total") {
		t.Error("expected auth_attempts_total metric in output")
	}
	if !strings.Contains(body, "testapp_auth_failures_total") {
		t.Error("expected auth_failures_total metric in output")
	}
}

func TestRecordMCPCall(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.RecordMCPCall("query", 50*time.Millisecond, nil)
	m.RecordMCPCall("query", 10*time.Millisecond, errors.New("test error"))

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_mcp_calls_total") {
		t.Error("expected mcp_calls_total metric in output")
	}
	if !strings.Contains(body, "testapp_mcp_errors_total") {
		t.Error("expected mcp_errors_total metric in output")
	}
}

func TestUpdateDBConnections(t *testing.T) {
	instance = nil
	m := Init("testapp")

	m.UpdateDBConnections("primary", 5, 3)

	handler := m.Handler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "testapp_db_connections_active") {
		t.Error("expected db_connections_active metric in output")
	}
	if !strings.Contains(body, "testapp_db_connections_idle") {
		t.Error("expected db_connections_idle metric in output")
	}
}
