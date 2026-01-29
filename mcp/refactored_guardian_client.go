package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"nofx/mcp/components"
)

// RefactoredGuardianClient represents a refactored client for Guardian AI service using component-based design
type RefactoredGuardianClient struct {
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

	// Components for better separation of concerns (遵循单一职责原则)
	BrowserEngine components.BrowserEngine
	DOMInspector  components.DOMInspectorInterface
}

// NewRefactoredGuardianClientFromConfig creates a new instance of RefactoredGuardianClient with proper component injection
func NewRefactoredGuardianClientFromConfig(config Config) *RefactoredGuardianClient {
	// Set default values if not configured
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	if config.Timeout <= 0 {
		config.Timeout = time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS) * time.Second
	}

	// Create logger
	logger := log.Default()

	// Create and inject components (依赖注入)
	browserEngine := components.NewBrowserAutomationEngine(
		logger,
		false, // Use default display setting since Config doesn't have DisplayEnabled
		time.Duration(config.Timeout),
	)
	domInspector := components.NewDOMInspector(logger)

	client := &RefactoredGuardianClient{
		Name:           "RefactoredGuardian",
		ModelID:        "guardian-browser-automation-refactored",
		APIKey:         config.APIKey,
		BaseURL:        config.BaseURL,
		Model:          config.Model,
		MaxTokens:      config.MaxTokens,
		Temperature:    config.Temperature,
		ProviderConfig: config,
		SystemPrompt:   "", // Use default since Config doesn't have SystemPrompt
		logger:         logger,
		httpClient:     &http.Client{Timeout: config.Timeout},
		DisplayEnabled: false, // Use default since Config doesn't have DisplayEnabled
		BrowserEngine:  browserEngine,
		DOMInspector:   domInspector,
	}

	return client
}

// PerformBrowserAutomation executes browser automation tasks using injected components
func (rgc *RefactoredGuardianClient) PerformBrowserAutomation(ctx context.Context, url string, selectors []string) (string, error) {
	// 使用注入的BrowserEngine组件进行页面导航
	if err := rgc.BrowserEngine.NavigateToURL(ctx, url); err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %v", err)
	}

	// 使用注入的DOMInspector组件寻找输入框
	inputSelector, found := rgc.DOMInspector.FindInputElement(ctx, selectors)
	if !found {
		return "", fmt.Errorf("input element not found with any of the provided selectors")
	}

	// 记录找到的元素信息
	rgc.logger.Printf("Found input element with selector: %s", inputSelector)

	// 返回找到的选择器，供后续使用
	return inputSelector, nil
}

// GetTokenizer returns the tokenizer for this client
func (rgc *RefactoredGuardianClient) GetTokenizer() (interface{}, error) {
	// Placeholder implementation
	return nil, nil
}

// GetMaxTokenCount returns the maximum token count for this client
func (rgc *RefactoredGuardianClient) GetMaxTokenCount() (int, error) {
	return rgc.MaxTokens, nil
}

// CalculateTokenCount calculates the token count for the given text
func (rgc *RefactoredGuardianClient) CalculateTokenCount(text string) (int, error) {
	return len([]rune(text)), nil
}

// GetModelID returns the model ID
func (rgc *RefactoredGuardianClient) GetModelID() string {
	return rgc.ModelID
}

// GetName returns the client name
func (rgc *RefactoredGuardianClient) GetName() string {
	return rgc.Name
}

// GetProvider returns the provider name
func (rgc *RefactoredGuardianClient) GetProvider() string {
	return ProviderGuardian
}

// GetBaseURL returns the base URL
func (rgc *RefactoredGuardianClient) GetBaseURL() string {
	return rgc.BaseURL
}

// GetModel returns the model name
func (rgc *RefactoredGuardianClient) GetModel() string {
	return rgc.Model
}

// GetMaxTokens returns the max tokens
func (rgc *RefactoredGuardianClient) GetMaxTokens() int {
	return rgc.MaxTokens
}

// GetTemperature returns the temperature setting
func (rgc *RefactoredGuardianClient) GetTemperature() float64 {
	return rgc.Temperature
}

// GetAPIKey returns the API key
func (rgc *RefactoredGuardianClient) GetAPIKey() string {
	return rgc.APIKey
}

// GetSystemPrompt returns the system prompt
func (rgc *RefactoredGuardianClient) GetSystemPrompt() string {
	return rgc.SystemPrompt
}

// GetConfig returns the provider config
func (rgc *RefactoredGuardianClient) GetConfig() Config {
	return rgc.ProviderConfig
}

// GetHTTPClient returns the HTTP client
func (rgc *RefactoredGuardianClient) GetHTTPClient() *http.Client {
	return rgc.httpClient
}

// SetBaseURL sets the base URL
func (rgc *RefactoredGuardianClient) SetBaseURL(baseURL string) {
	rgc.BaseURL = baseURL
	rgc.ProviderConfig.BaseURL = baseURL
}

// CallWithMessages implements the AIClient interface
func (rgc *RefactoredGuardianClient) CallWithMessages(messages []map[string]interface{}) (string, error) {
	// For now, return an error indicating this is not implemented
	return "", fmt.Errorf("CallWithMessages not implemented for RefactoredGuardianClient")
}

// CallWithPrompt implements the AIClient interface
func (rgc *RefactoredGuardianClient) CallWithPrompt(prompt string) (string, error) {
	// For now, return an error indicating this is not implemented
	return "", fmt.Errorf("CallWithPrompt not implemented for RefactoredGuardianClient")
}

// StreamWithMessages implements the AIClient interface
func (rgc *RefactoredGuardianClient) StreamWithMessages(messages []map[string]interface{}, handler func(string)) error {
	return fmt.Errorf("StreamWithMessages not implemented for RefactoredGuardianClient")
}

// GetEncoding returns the encoding for this client
func (rgc *RefactoredGuardianClient) GetEncoding() (interface{}, error) {
	return nil, nil
}
