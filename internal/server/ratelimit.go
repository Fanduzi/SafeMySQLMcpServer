// Package server handles HTTP server setup
package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond rate.Limit
	Burst             int
	CleanupInterval   time.Duration
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,  // 10 requests per second
		Burst:             20,  // Allow burst of 20 requests
		CleanupInterval:   time.Minute,
	}
}

// IPRateLimiter tracks rate limiters per IP address
type IPRateLimiter struct {
	ips    map[string]*ipEntry
	mu     sync.RWMutex
	config RateLimiterConfig
	ctx    context.Context
	cancel context.CancelFunc
}

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(config RateLimiterConfig) *IPRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	limiter := &IPRateLimiter{
		ips:    make(map[string]*ipEntry),
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Start cleanup goroutine
	go limiter.cleanupLoop()

	return limiter
}

// Close gracefully stops the cleanup goroutine
func (l *IPRateLimiter) Close() {
	l.cancel()
}

// GetLimiter returns the rate limiter for an IP address
func (l *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.ips[ip]
	if !exists {
		limiter := rate.NewLimiter(l.config.RequestsPerSecond, l.config.Burst)
		l.ips[ip] = &ipEntry{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupLoop periodically removes old entries
func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

// cleanup removes entries that haven't been used recently
func (l *IPRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	threshold := time.Now().Add(-5 * time.Minute)
	for ip, entry := range l.ips {
		if entry.lastSeen.Before(threshold) {
			delete(l.ips, ip)
		}
	}
}

// rateLimitMiddleware creates a rate limiting middleware
func (s *Server) rateLimitMiddleware(limiter *IPRateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRealIP(r)
		if !limiter.GetLimiter(ip).Allow() {
			log.Printf("Rate limit exceeded for IP: %s", ip)
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getRealIP extracts the real IP address from the request
// It checks common headers used by proxies
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may contain multiple IPs, use the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP header (used by nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if isValidIP(ip) {
			return ip
		}
	}

	// Fall back to RemoteAddr (may contain port)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isValidIP validates if a string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
