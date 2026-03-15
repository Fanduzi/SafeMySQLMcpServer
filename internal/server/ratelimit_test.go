package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fan/safe-mysql-mcp/internal/metrics"
	"golang.org/x/time/rate"
)

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()
	if config.RequestsPerSecond != 10 {
		t.Errorf("RequestsPerSecond = %v, want 10", config.RequestsPerSecond)
	}
	if config.Burst != 20 {
		t.Errorf("Burst = %v, want 20", config.Burst)
	}
	if config.CleanupInterval != time.Minute {
		t.Errorf("CleanupInterval = %v, want 1m", config.CleanupInterval)
	}
}

func TestNewIPRateLimiter(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 5,
		Burst:             10,
		CleanupInterval:   time.Second,
	}
	limiter := NewIPRateLimiter(config)
	defer limiter.Close()

	if limiter == nil {
		t.Fatal("NewIPRateLimiter returned nil")
	}
}

func TestIPRateLimiter_GetLimiter(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             5,
		CleanupInterval:   time.Minute,
	}
	limiter := NewIPRateLimiter(config)
	defer limiter.Close()

	// First request for IP should create new limiter
	l1 := limiter.GetLimiter("192.168.1.1")
	if l1 == nil {
		t.Fatal("GetLimiter returned nil")
	}

	// Second request for same IP should return same limiter
	l2 := limiter.GetLimiter("192.168.1.1")
	if l1 != l2 {
		t.Error("GetLimiter should return same limiter for same IP")
	}

	// Different IP should get different limiter
	l3 := limiter.GetLimiter("192.168.1.2")
	if l1 == l3 {
		t.Error("GetLimiter should return different limiter for different IP")
	}
}

func TestIPRateLimiter_RateLimiting(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 1,
		Burst:             2,
		CleanupInterval:   time.Minute,
	}
	limiter := NewIPRateLimiter(config)
	defer limiter.Close()

	ip := "192.168.1.1"
	rl := limiter.GetLimiter(ip)

	// Should allow burst requests
	if !rl.Allow() {
		t.Error("First request should be allowed")
	}
	if !rl.Allow() {
		t.Error("Second request should be allowed (burst)")
	}

	// Third request should be limited (burst exhausted)
	if rl.Allow() {
		t.Error("Third request should be rate limited")
	}
}

func TestIPRateLimiter_Cleanup(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             5,
		CleanupInterval:   100 * time.Millisecond,
	}
	limiter := NewIPRateLimiter(config)
	defer limiter.Close()

	// Add an entry
	limiter.GetLimiter("192.168.1.1")

	// Verify entry exists
	limiter.mu.RLock()
	_, exists := limiter.ips["192.168.1.1"]
	limiter.mu.RUnlock()
	if !exists {
		t.Fatal("IP entry should exist")
	}

	// Wait for cleanup to run (entries older than 5 minutes are removed)
	// Since we just added it, it won't be removed by normal cleanup
	// Let's manually trigger cleanup with old entry
	limiter.mu.Lock()
	limiter.ips["192.168.1.2"] = &ipEntry{
		limiter:  rate.NewLimiter(10, 5),
		lastSeen: time.Now().Add(-10 * time.Minute), // Old entry
	}
	limiter.mu.Unlock()

	// Trigger cleanup
	limiter.cleanup()

	// Old entry should be removed
	limiter.mu.RLock()
	_, existsOld := limiter.ips["192.168.1.2"]
	_, existsNew := limiter.ips["192.168.1.1"]
	limiter.mu.RUnlock()

	if existsOld {
		t.Error("Old IP entry should be cleaned up")
	}
	if !existsNew {
		t.Error("Recent IP entry should still exist")
	}
}

func TestIPRateLimiter_Close(t *testing.T) {
	config := RateLimiterConfig{
		RequestsPerSecond: 10,
		Burst:             5,
		CleanupInterval:   100 * time.Millisecond,
	}
	limiter := NewIPRateLimiter(config)

	// Close should not panic
	limiter.Close()

	// Give time for goroutine to exit
	time.Sleep(50 * time.Millisecond)
}

func TestGetRealIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			want:       "192.168.1.1",
		},
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			want:       "203.0.113.1",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1, 70.41.3.18"},
			want:       "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.2"},
			want:       "203.0.113.2",
		},
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1", "X-Real-IP": "203.0.113.2"},
			want:       "203.0.113.1",
		},
		{
			name:       "Invalid X-Forwarded-For falls back",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "invalid-ip"},
			want:       "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			headers:    map[string]string{},
			want:       "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := getRealIP(req)
			if got != tt.want {
				t.Errorf("getRealIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"127.0.0.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"", false},
		{"invalid", false},
		{"256.256.256.256", false},
		{"192.168.1", false},
		{"192.168.1.1:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := isValidIP(tt.ip); got != tt.want {
				t.Errorf("isValidIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestMetricsMiddleware(t *testing.T) {
	// Initialize metrics
	m := metrics.Init("test")

	// Create a test server with metrics middleware
	s := &Server{metrics: m}

	// Create a simple handler that returns 200
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap with metrics middleware
	wrapped := s.metricsMiddleware("/test", handler)

	// Make a request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify metrics were recorded
	metricsHandler := m.Handler()
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	metricsHandler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if body == "" {
		t.Error("Expected metrics output")
	}
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	wrapped := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	// Test default status
	if wrapped.status != http.StatusOK {
		t.Errorf("Default status = %d, want %d", wrapped.status, http.StatusOK)
	}

	// Test WriteHeader
	wrapped.WriteHeader(http.StatusNotFound)
	if wrapped.status != http.StatusNotFound {
		t.Errorf("Status after WriteHeader = %d, want %d", wrapped.status, http.StatusNotFound)
	}
}
