package mcp

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"nofx/logger"
)

// Config client configuration (centralized management of all configurations)
type Config struct {
	// Provider configuration
	Provider string
	APIKey   string
	BaseURL  string
	Model    string

	// Behavior configuration
	MaxTokens   int
	Temperature float64
	UseFullURL  bool

	// Retry configuration
	MaxRetries      int
	RetryWaitBase   time.Duration
	RetryableErrors []string

	// Timeout configuration
	Timeout time.Duration

	// Context compression configuration (上下文压缩配置)
	EnableContextCompression bool // 是否启用上下文压缩
	ModelContextSize         int  // 模型上下文大小（字符数/tokens）

	// Browser display configuration
	DisplayEnabled bool // 是否启用浏览器显示

	// Trader identification for browser data isolation
	TraderID string // 交易员ID，用于浏览器数据隔离

	// Dependency injection
	Logger     Logger
	HTTPClient *http.Client
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		// Default values
		MaxTokens:       getEnvInt("AI_MAX_TOKENS", 2000),
		Temperature:     MCPClientTemperature,
		MaxRetries:      MaxRetryTimes,
		RetryWaitBase:   2 * time.Second,
		Timeout:         DefaultTimeout,
		RetryableErrors: retryableErrors,

		// Context compression configuration
		EnableContextCompression: getEnvBool("MCP_ENABLE_CONTEXT_COMPRESSION", true), // 默认开启压缩
		ModelContextSize:         getEnvInt("MCP_MODEL_CONTEXT_SIZE", 0),             // 0表示自动检测

		// Browser display configuration
		DisplayEnabled: getEnvBool("GUARDIAN_DISPLAY_ENABLED", true), // 从环境变量读取显示设置

		// Default dependencies (use global logger)
		Logger:     logger.NewMCPLogger(),
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}
}

// getEnvInt reads integer from environment variable, returns default value if failed
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// getEnvString reads string from environment variable, returns default value if empty
func getEnvString(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvBool reads boolean from environment variable, returns default value if failed
func getEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if val == "true" || val == "1" || val == "yes" {
			return true
		}
		if val == "false" || val == "0" || val == "no" {
			return false
		}
	}
	return defaultValue
}
