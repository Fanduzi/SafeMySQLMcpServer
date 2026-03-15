// Package metrics provides Prometheus metrics for the MCP server.
// input: HTTP requests, DB queries, auth events, MCP calls
// output: prometheus.Counter, prometheus.Histogram, prometheus.Gauge
// pos: observability layer, exposes /metrics endpoint for Prometheus
// note: if this file changes, update header and internal/metrics/README.md
package metrics

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// Request metrics
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestsActive  prometheus.Gauge

	// Query metrics
	QueriesTotal    *prometheus.CounterVec
	QueryDuration   *prometheus.HistogramVec
	QueriesActive   prometheus.Gauge
	QueryRows       *prometheus.HistogramVec

	// Security metrics
	SQLInjectionAttempts *prometheus.CounterVec
	BlockedQueries       *prometheus.CounterVec
	SecurityViolations   *prometheus.CounterVec

	// Rate limiter metrics
	RateLimitExceeded *prometheus.CounterVec

	// Auth metrics
	AuthAttempts *prometheus.CounterVec
	AuthFailures *prometheus.CounterVec

	// Connection metrics
	DBConnectionsActive *prometheus.GaugeVec
	DBConnectionsIdle   *prometheus.GaugeVec
	DBQueriesTotal      *prometheus.CounterVec
	DBQueryErrors       *prometheus.CounterVec

	// MCP metrics
	MCPCallsTotal   *prometheus.CounterVec
	MCPCallDuration *prometheus.HistogramVec
	MCPErrors       *prometheus.CounterVec

	registry *prometheus.Registry
}

var (
	instance *Metrics
	mu       sync.Mutex
)

// Init initializes the metrics instance
func Init(namespace string) *Metrics {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance
	}

	instance = &Metrics{
		registry: prometheus.NewRegistry(),
	}

	// Request metrics
	instance.RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	instance.RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	instance.RequestsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_active",
			Help:      "Number of active HTTP requests",
		},
	)

	// Query metrics
	instance.QueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"database", "operation"},
	)

	instance.QueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"database", "operation"},
	)

	instance.QueriesActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_queries_active",
			Help:      "Number of active database queries",
		},
	)

	instance.QueryRows = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "db_query_rows",
			Help:      "Number of rows returned by queries",
			Buckets:   []float64{0, 1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"database"},
	)

	// Security metrics
	instance.SQLInjectionAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "security_sql_injection_attempts_total",
			Help:      "Total number of SQL injection attempts detected",
		},
		[]string{"database"},
	)

	instance.BlockedQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "security_blocked_queries_total",
			Help:      "Total number of blocked queries",
		},
		[]string{"database", "reason"},
	)

	instance.SecurityViolations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "security_violations_total",
			Help:      "Total number of security violations",
		},
		[]string{"type"},
	)

	// Rate limiter metrics
	instance.RateLimitExceeded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "rate_limit_exceeded_total",
			Help:      "Total number of rate limit exceeded events",
		},
		[]string{"ip"},
	)

	// Auth metrics
	instance.AuthAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_attempts_total",
			Help:      "Total number of authentication attempts",
		},
		[]string{"method"},
	)

	instance.AuthFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_failures_total",
			Help:      "Total number of authentication failures",
		},
		[]string{"method", "reason"},
	)

	// Connection metrics
	instance.DBConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_connections_active",
			Help:      "Number of active database connections",
		},
		[]string{"cluster"},
	)

	instance.DBConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_connections_idle",
			Help:      "Number of idle database connections",
		},
		[]string{"cluster"},
	)

	instance.DBQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_connection_queries_total",
			Help:      "Total queries per database connection",
		},
		[]string{"cluster"},
	)

	instance.DBQueryErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "db_query_errors_total",
			Help:      "Total database query errors",
		},
		[]string{"cluster", "error_type"},
	)

	// MCP metrics
	instance.MCPCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "mcp_calls_total",
			Help:      "Total number of MCP tool calls",
		},
		[]string{"tool"},
	)

	instance.MCPCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "mcp_call_duration_seconds",
			Help:      "MCP tool call duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"tool"},
	)

	instance.MCPErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "mcp_errors_total",
			Help:      "Total MCP tool call errors",
		},
		[]string{"tool", "error_type"},
	)

	// Register all metrics
	instance.registry.MustRegister(
		instance.RequestsTotal,
		instance.RequestDuration,
		instance.RequestsActive,
		instance.QueriesTotal,
		instance.QueryDuration,
		instance.QueriesActive,
		instance.QueryRows,
		instance.SQLInjectionAttempts,
		instance.BlockedQueries,
		instance.SecurityViolations,
		instance.RateLimitExceeded,
		instance.AuthAttempts,
		instance.AuthFailures,
		instance.DBConnectionsActive,
		instance.DBConnectionsIdle,
		instance.DBQueriesTotal,
		instance.DBQueryErrors,
		instance.MCPCallsTotal,
		instance.MCPCallDuration,
		instance.MCPErrors,
	)

	return instance
}

// Get returns the metrics instance
func Get() *Metrics {
	if instance == nil {
		log.Println("Warning: metrics not initialized, using default instance")
		return Init("safemysql")
	}
	return instance
}

// Handler returns the Prometheus HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordRequest records an HTTP request
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
	m.RequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordQuery records a database query
func (m *Metrics) RecordQuery(database, operation string, duration time.Duration, rows int64) {
	m.QueriesTotal.WithLabelValues(database, operation).Inc()
	m.QueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
	m.QueryRows.WithLabelValues(database).Observe(float64(rows))
}

// RecordSecurityViolation records a security violation
func (m *Metrics) RecordSecurityViolation(violationType string) {
	m.SecurityViolations.WithLabelValues(violationType).Inc()
}

// RecordBlockedQuery records a blocked query
func (m *Metrics) RecordBlockedQuery(database, reason string) {
	m.BlockedQueries.WithLabelValues(database, reason).Inc()
}

// RecordRateLimitExceeded records a rate limit event
func (m *Metrics) RecordRateLimitExceeded(ip string) {
	m.RateLimitExceeded.WithLabelValues(ip).Inc()
}

// RecordAuthAttempt records an authentication attempt
func (m *Metrics) RecordAuthAttempt(method string, success bool) {
	m.AuthAttempts.WithLabelValues(method).Inc()
	if !success {
		m.AuthFailures.WithLabelValues(method, "invalid").Inc()
	}
}

// RecordMCPCall records an MCP tool call
func (m *Metrics) RecordMCPCall(tool string, duration time.Duration, err error) {
	m.MCPCallsTotal.WithLabelValues(tool).Inc()
	m.MCPCallDuration.WithLabelValues(tool).Observe(duration.Seconds())
	if err != nil {
		m.MCPErrors.WithLabelValues(tool, "execution").Inc()
	}
}

// UpdateDBConnections updates connection metrics
func (m *Metrics) UpdateDBConnections(cluster string, active, idle int) {
	m.DBConnectionsActive.WithLabelValues(cluster).Set(float64(active))
	m.DBConnectionsIdle.WithLabelValues(cluster).Set(float64(idle))
}
