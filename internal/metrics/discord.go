package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// DiscordAPIMetrics provides methods for recording Discord API-related metrics
type DiscordAPIMetrics struct{}

// NewDiscordAPIMetrics creates a new Discord API metrics instance
func NewDiscordAPIMetrics() *DiscordAPIMetrics {
	return &DiscordAPIMetrics{}
}

// RecordAPIRequest records metrics for a Discord API request
func (d *DiscordAPIMetrics) RecordAPIRequest(endpoint, method string, statusCode int, duration time.Duration) {
	status := strconv.Itoa(statusCode)

	// Record request count
	DiscordAPIRequests.WithLabelValues(endpoint, method, status).Inc()

	// Record latency only for successful requests to avoid skewing metrics
	if statusCode >= 200 && statusCode < 400 {
		DiscordAPILatency.WithLabelValues(endpoint, method).Observe(duration.Seconds())
	}
}

// DiscordAPITimer provides a timer for measuring Discord API request duration
type DiscordAPITimer struct {
	metrics   *DiscordAPIMetrics
	endpoint  string
	method    string
	startTime time.Time
}

// NewDiscordAPITimer creates a new timer for Discord API requests
func (d *DiscordAPIMetrics) NewDiscordAPITimer(endpoint, method string) *DiscordAPITimer {
	return &DiscordAPITimer{
		metrics:   d,
		endpoint:  endpoint,
		method:    method,
		startTime: time.Now(),
	}
}

// Finish completes the timer and records the API request metrics
func (t *DiscordAPITimer) Finish(statusCode int) {
	duration := time.Since(t.startTime)
	t.metrics.RecordAPIRequest(t.endpoint, t.method, statusCode, duration)
}

// HTTPMiddleware creates an HTTP middleware for Discord API metrics collection
func (d *DiscordAPIMetrics) HTTPMiddleware(next http.RoundTripper) http.RoundTripper {
	return &discordAPITransport{
		next:    next,
		metrics: d,
	}
}

// discordAPITransport wraps HTTP transport to collect metrics
type discordAPITransport struct {
	next    http.RoundTripper
	metrics *DiscordAPIMetrics
}

// RoundTrip implements http.RoundTripper interface
func (t *discordAPITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract endpoint from URL path
	endpoint := extractDiscordEndpoint(req.URL.Path)

	// Start timer
	timer := t.metrics.NewDiscordAPITimer(endpoint, req.Method)

	// Execute request
	resp, err := t.next.RoundTrip(req)

	// Record metrics
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	if err != nil && statusCode == 0 {
		statusCode = 500 // Use 500 for network errors
	}

	timer.Finish(statusCode)

	return resp, err
}

// extractDiscordEndpoint extracts a clean endpoint identifier from Discord API paths
func extractDiscordEndpoint(path string) string {
	// Clean up dynamic parts of Discord API paths for better metric grouping
	// Examples:
	// /api/v10/guilds/123456789/channels -> /guilds/{guild.id}/channels
	// /api/v10/channels/123456789/messages -> /channels/{channel.id}/messages

	// Simple implementation - in production you might want more sophisticated parsing
	if len(path) == 0 {
		return "unknown"
	}

	// Remove /api/v10 prefix if present
	if len(path) > 8 && path[:8] == "/api/v10" {
		path = path[8:]
	}

	// For now, return the path as-is but limit length to avoid cardinality explosion
	if len(path) > 50 {
		path = path[:50]
	}

	return path
}
