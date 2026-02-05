package mcp

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	ProviderGuardian       = "guardian"
	DefaultGuardianBaseURL = "https://chat.deepseek.com"
	DefaultGuardianModel   = "guardian-ai"
)

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
	httpClient     *http.Client
	DisplayEnabled bool // 是否启用显示
}

// NewGuardianClient creates a new instance of GuardianClient
func NewGuardianClientFromConfig(config Config) *GuardianClient {
	// Set default values if not configured
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	// Config结构体中没有TimeoutSeconds字段，使用timeout.Duration类型的Timeout字段
	if config.Timeout <= 0 {
		config.Timeout = time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS) * time.Second
	}

	client := &GuardianClient{
		Name:           "Guardian",
		ModelID:        "guardian-ai",
		APIKey:         config.APIKey,
		BaseURL:        config.BaseURL,
		Model:          config.Model,
		MaxTokens:      config.MaxTokens,
		Temperature:    config.Temperature,
		ProviderConfig: config,
		SystemPrompt:   "", // Config结构体中没有SystemPrompt字段，暂时设为空
		logger:         log.Default(),
		httpClient:     config.HTTPClient,
		DisplayEnabled: true, // 默认启用显示
	}

	return client
}

// GetTokenizer returns nil since we're using browser automation
func (gc *GuardianClient) GetTokenizer() interface{} {
	return nil
}

// GetMaxTokenCount returns the max token count
func (gc *GuardianClient) GetMaxTokenCount() int {
	return gc.ProviderConfig.MaxTokens
}

// CalculateTokenCount returns approximate token count
func (gc *GuardianClient) CalculateTokenCount(text string) int {
	// Rough estimation: 1 token ≈ 4 characters
	return len(text) / 4
}

// GetRemainingTokenCount returns remaining token count
func (gc *GuardianClient) GetRemainingTokenCount() int {
	return gc.ProviderConfig.MaxTokens
}

// SetLogger sets the logger for the client
func (gc *GuardianClient) SetLogger(logger *logrus.Logger) {
	gc.logger.SetOutput(logger.Out)
}

// SetTimeout 设置超时时间
func (gc *GuardianClient) SetTimeout(timeout time.Duration) {
	gc.ProviderConfig.Timeout = timeout
}

// String returns string representation of the client
func (gc *GuardianClient) String() string {
	return fmt.Sprintf("[Provider: %s, Model: %s]", gc.ProviderConfig.Provider, gc.ProviderConfig.Model)
}

// GetProvider returns the provider name
func (gc *GuardianClient) GetProvider() string {
	return gc.ProviderConfig.Provider
}

// GetModel returns the model name
func (gc *GuardianClient) GetModel() string {
	return gc.Model
}

// GetBaseURL returns the base URL
func (gc *GuardianClient) GetBaseURL() string {
	return gc.BaseURL
}

// GetAPIKey returns the API key
func (gc *GuardianClient) GetAPIKey() string {
	return gc.APIKey
}

// SetDynamicConfig dynamically sets the base URL
func (gc *GuardianClient) SetDynamicConfig(baseURL string) {
	gc.BaseURL = baseURL
	gc.ProviderConfig.BaseURL = baseURL
}

// GetDisplayEnabled returns whether display is enabled
func (gc *GuardianClient) GetDisplayEnabled() bool {
	return gc.DisplayEnabled
}

// SetDisplayEnabled sets whether display is enabled
func (gc *GuardianClient) SetDisplayEnabled(enabled bool) {
	gc.DisplayEnabled = enabled
}

const (
	GUARDIAN_BROWSER_TIMEOUT_SECONDS = 999 // 改为999秒，确保AI有足够时间完成输出

	GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS = 999
	GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS   = 999
	GUARDIAN_AUTO_KEEP_OPEN_SECONDS         = 999 // 改为999秒，确保窗口长时间存活
)
