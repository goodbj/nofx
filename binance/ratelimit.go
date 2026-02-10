package binance

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter Binance API rate limiter using token bucket algorithm
type RateLimiter struct {
	// Token bucket implementation
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time

	// Configuration
	config *Config

	// Statistics
	stats *RateLimitStats

	// Mutex for thread safety
	mu sync.RWMutex
}

// RateLimitStats rate limit statistics
type RateLimitStats struct {
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	RejectedRequests   int64     `json:"rejected_requests"`
	LastErrorTime      time.Time `json:"last_error_time"`
	LastError          string    `json:"last_error"`
	LastResetTime      time.Time `json:"last_reset_time"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *Config) *RateLimiter {
	// Binance futures rate limits (as of current documentation):
	// - REQUEST_WEIGHT: 2400 per minute (40 per second)
	// - ORDERS: 1200 per minute (20 per second)
	// - RAW_REQUESTS: configurable

	// Use the most restrictive limit as our baseline
	// For safety, we'll use 20 requests per second (1200 per minute)
	maxTokens := 20.0
	refillRate := 20.0 // 20 tokens per second

	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
		config:         config,
		stats: &RateLimitStats{
			LastResetTime: time.Now(),
		},
	}
}

// refillTokens refills tokens based on time elapsed
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime).Seconds()

	// Add tokens based on elapsed time
	tokensToAdd := elapsed * rl.refillRate
	rl.tokens = min(rl.tokens+tokensToAdd, rl.maxTokens)
	rl.lastRefillTime = now
}

// Wait waits for rate limit permission
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens first
	rl.refillTokens()

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		rl.stats.TotalRequests++
		rl.stats.SuccessfulRequests++
		return nil
	}

	// Calculate wait time needed
	waitTime := time.Duration((1.0-rl.tokens)/rl.refillRate) * time.Second

	// Wait for tokens or context cancellation
	select {
	case <-time.After(waitTime):
		rl.tokens = 0 // Consume one token
		rl.stats.TotalRequests++
		rl.stats.SuccessfulRequests++
		return nil
	case <-ctx.Done():
		rl.stats.RejectedRequests++
		rl.stats.LastError = fmt.Sprintf("context cancelled while waiting for rate limit: %v", ctx.Err())
		rl.stats.LastErrorTime = time.Now()
		return ctx.Err()
	}
}

// WaitOrder waits for order rate limit permission (more restrictive)
func (rl *RateLimiter) WaitOrder(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// For orders, use a more restrictive rate (10 per second instead of 20)
	originalRate := rl.refillRate
	rl.refillRate = 10.0

	err := rl.Wait(ctx)

	// Restore original rate
	rl.refillRate = originalRate

	return err
}

// Reserve reserves rate limit tokens without blocking
func (rl *RateLimiter) Reserve() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// ReserveOrder reserves order rate limit tokens without blocking
func (rl *RateLimiter) ReserveOrder() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Use more restrictive rate for orders
	originalRate := rl.refillRate
	rl.refillRate = 10.0

	rl.refillTokens()
	result := rl.tokens >= 1.0
	if result {
		rl.tokens -= 1.0
	}

	// Restore original rate
	rl.refillRate = originalRate

	return result
}

// UpdateFromResponse updates rate limits based on API response
func (rl *RateLimiter) UpdateFromResponse(headers map[string][]string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Parse rate limit headers and adjust if necessary
	// This is a simplified implementation
	if usedWeight := getHeaderValue(headers, "X-Mbx-Used-Weight-1m"); usedWeight != "" {
		// In a real implementation, you might adjust the rate based on actual usage
		// For now, we'll just log it
	}
}

// GetStats returns current rate limit statistics
func (rl *RateLimiter) GetStats() *RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Return a copy of stats
	stats := *rl.stats
	return &stats
}

// ResetStats resets statistics
func (rl *RateLimiter) ResetStats() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.stats = &RateLimitStats{
		LastResetTime: time.Now(),
	}
}

// IsHealthy checks if rate limiter is functioning properly
func (rl *RateLimiter) IsHealthy() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// Check if we're not rejecting too many requests
	if rl.stats.TotalRequests > 0 {
		rejectionRate := float64(rl.stats.RejectedRequests) / float64(rl.stats.TotalRequests)
		return rejectionRate < 0.1 // Less than 10% rejection rate is considered healthy
	}

	return true
}

// AdaptiveRateLimiter adaptive rate limiter that adjusts based on error history
type AdaptiveRateLimiter struct {
	*RateLimiter
	errorWindow    time.Duration
	errorThreshold int
	errorCount     int
	lastErrorTime  time.Time
	mu             sync.RWMutex
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(config *Config) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		RateLimiter:    NewRateLimiter(config),
		errorWindow:    5 * time.Minute,
		errorThreshold: 5, // Slow down after 5 errors in 5 minutes
	}
}

// RecordError records an API error for adaptive rate limiting
func (arl *AdaptiveRateLimiter) RecordError() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	now := time.Now()

	// Reset count if outside error window
	if now.Sub(arl.lastErrorTime) > arl.errorWindow {
		arl.errorCount = 0
	}

	arl.errorCount++
	arl.lastErrorTime = now

	// If error threshold exceeded, temporarily reduce rate limit
	if arl.errorCount >= arl.errorThreshold {
		// Reduce rate by 50% temporarily by adjusting refill rate
		originalRate := arl.RateLimiter.refillRate
		arl.RateLimiter.refillRate = originalRate * 0.5
	}
}

// WaitWithAdaptiveBackoff waits with adaptive backoff based on error history
func (arl *AdaptiveRateLimiter) WaitWithAdaptiveBackoff(ctx context.Context) error {
	arl.mu.RLock()
	shouldBackoff := arl.errorCount >= arl.errorThreshold
	errorCount := arl.errorCount
	arl.mu.RUnlock()

	if shouldBackoff {
		// Add additional delay based on error count
		additionalDelay := time.Duration(errorCount) * time.Second
		select {
		case <-time.After(additionalDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return arl.Wait(ctx)
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func getHeaderValue(headers map[string][]string, key string) string {
	if values, exists := headers[key]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}
