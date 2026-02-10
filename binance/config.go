package binance

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config Binance API configuration
type Config struct {
	APIKey         string        `json:"api_key"`
	SecretKey      string        `json:"secret_key"`
	BaseURL        string        `json:"base_url"`
	ProxyURL       string        `json:"proxy_url,omitempty"`
	Testnet        bool          `json:"testnet"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
	RateLimit      int           `json:"rate_limit"`       // requests per minute
	RateLimitBurst int           `json:"rate_limit_burst"` // burst requests allowed
}

// ConfigOption configuration option function
type ConfigOption func(*Config)

// NewConfig creates new Binance configuration with default values
func NewConfig(apiKey, secretKey string, opts ...ConfigOption) *Config {
	config := &Config{
		APIKey:         apiKey,
		SecretKey:      secretKey,
		BaseURL:        "https://fapi.binance.com", // Default production URL
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		RateLimit:      1200, // 1200 requests per minute (20 per second)
		RateLimitBurst: 10,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Auto-detect testnet based on URL if not explicitly set
	if !config.Testnet {
		config.Testnet = strings.Contains(config.BaseURL, "testnet")
	}

	// Set testnet URL if needed
	if config.Testnet && !strings.Contains(config.BaseURL, "testnet") {
		config.BaseURL = "https://testnet.binancefuture.com"
	}

	return config
}

// WithBaseURL sets custom base URL
func WithBaseURL(url string) ConfigOption {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithProxyURL sets proxy URL
func WithProxyURL(proxyURL string) ConfigOption {
	return func(c *Config) {
		c.ProxyURL = proxyURL
	}
}

// WithTestnet enables testnet mode
func WithTestnet(testnet bool) ConfigOption {
	return func(c *Config) {
		c.Testnet = testnet
		if testnet {
			c.BaseURL = "https://testnet.binancefuture.com"
		} else {
			c.BaseURL = "https://fapi.binance.com"
		}
	}
}

// WithTimeout sets request timeout
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetries sets retry configuration
func WithRetries(maxRetries int, delay time.Duration) ConfigOption {
	return func(c *Config) {
		c.MaxRetries = maxRetries
		c.RetryDelay = delay
	}
}

// WithRateLimit sets rate limit configuration
func WithRateLimit(requestsPerMinute, burst int) ConfigOption {
	return func(c *Config) {
		c.RateLimit = requestsPerMinute
		c.RateLimitBurst = burst
	}
}

// FromEnvironment creates config from environment variables
func FromEnvironment() *Config {
	return NewConfig(
		os.Getenv("BINANCE_API_KEY"),
		os.Getenv("BINANCE_SECRET_KEY"),
		WithBaseURL(os.Getenv("BINANCE_BASE_URL")),
		WithProxyURL(os.Getenv("BINANCE_PROXY_URL")),
		WithTestnet(os.Getenv("BINANCE_TESTNET") == "true"),
		WithTimeout(parseDuration(os.Getenv("BINANCE_TIMEOUT"), 30*time.Second)),
		WithRetries(
			parseInt(os.Getenv("BINANCE_MAX_RETRIES"), 3),
			parseDuration(os.Getenv("BINANCE_RETRY_DELAY"), 1*time.Second),
		),
	)
}

// Validate validates configuration
func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret key is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if c.RateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}
	return nil
}

// GetEndpoint returns full endpoint URL for given path
func (c *Config) GetEndpoint(path string) string {
	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(c.BaseURL, "/"), path)
}

// IsTestnet returns whether this is testnet configuration
func (c *Config) IsTestnet() bool {
	return c.Testnet
}

// GetRateLimit returns rate limit configuration
func (c *Config) GetRateLimit() (requestsPerMinute, burst int) {
	return c.RateLimit, c.RateLimitBurst
}

// Helper functions for parsing environment variables
func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if s == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return d
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return defaultValue
	}
	return i
}
