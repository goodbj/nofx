package mcp

import (
	"log"
	"net/http"
	"time"
)

// Logger interface (abstract dependency)
// Uses Printf-style method names for easy integration with mainstream logging libraries like logrus, zap, etc.
// Default uses global logger package (see mcp/config.go)
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

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

	// Dependency injection
	Logger     Logger
	HTTPClient *http.Client
}

// GuardianClient represents a client for Guardian AI service
type GuardianClient struct {
	Name           string
	ModelID        string
	APIKey         string
	BaseURL        string
	Model          string
	MaxTokens      int
	Temperature    float64
	ProviderConfig Config
	SystemPrompt   string
	logger         *log.Logger
	DisplayEnabled bool // 是否启用显示
}

// Request represents a request to the AI service
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const (
	GUARDIAN_BROWSER_TIMEOUT_SECONDS = 999 // 改为999秒，确保AI有足够时间完成输出

	GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS = 999
	GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS   = 999
	GUARDIAN_AUTO_KEEP_OPEN_SECONDS         = 999 // 改为999秒，确保窗口长时间存活
)
