package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

// AIClient interface
type AIClient interface {
	SetAPIKey(apiKey string, customURL string, customModel string)
	CallWithMessages(systemPrompt, userPrompt string) (string, error)
	CallWithRequest(req *Request) (string, error)
	GetProvider() string
	GetModel() string
	GetBaseURL() string
	GetAPIKey() string
	SetDynamicConfig(baseURL string)
	GetDisplayEnabled() bool
	SetDisplayEnabled(enabled bool)
	SetTimeout(timeout time.Duration)
	String() string
	GetTokenizer() interface{}
	GetMaxTokenCount() int
	CalculateTokenCount(text string) int
	GetRemainingTokenCount() int
	SetLogger(logger *logrus.Logger)
}

const (
	ProviderGuardian       = "guardian"
	DefaultGuardianBaseURL = "https://chat.deepseek.com"
	DefaultGuardianModel   = "guardian-browser-automation"
)

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
		ModelID:        "guardian-browser-automation",
		APIKey:         config.APIKey,
		BaseURL:        config.BaseURL,
		Model:          config.Model,
		MaxTokens:      config.MaxTokens,
		Temperature:    config.Temperature,
		ProviderConfig: config,
		SystemPrompt:   "", // Config结构体中没有SystemPrompt字段，暂时设为空
		logger:         log.Default(),
		DisplayEnabled: true, // 默认启用显示
	}

	return client
}

// NewGuardianClient creates Guardian client (backward compatible)
func NewGuardianClient_() AIClient {
	return NewGuardianClientWithOptions()
}

// NewGuardianClient creates Guardian client (backward compatible)
func NewGuardianClient__() AIClient {
	return NewGuardianClientWithOptions()
}

// NewGuardianClient creates Guardian client (backward compatible)
func NewGuardianClient() AIClient {
	return NewGuardianClientWithOptions(
		WithProvider(ProviderGuardian),
		WithModel(DefaultGuardianModel),
		WithBaseURL(DefaultGuardianBaseURL),
	)
}

// NewGuardianClientForBrowser creates a GuardianClient for browser automation
func NewGuardianClientForBrowser() AIClient {
	return NewGuardianClientWithOptions(
		WithProvider(ProviderGuardian),
		WithModel(DefaultGuardianModel),
		WithBaseURL(DefaultGuardianBaseURL),
	)
}

// NewGuardianClientWithTarget creates a GuardianClient with specific target configuration
func NewGuardianClientWithTarget(targetURL, inputSelector, submitSelector, responseSelector string) AIClient {
	return NewGuardianClientWithOptions(
		WithProvider(ProviderGuardian),
		WithModel(DefaultGuardianModel),
		WithBaseURL(targetURL),
	)
}

// ClientOption defines a function to configure a client
type ClientOption func(*Config)

// WithProvider sets the provider
func WithProvider(provider string) ClientOption {
	return func(c *Config) {
		c.Provider = provider
	}
}

// WithModel sets the model
func WithModel(model string) ClientOption {
	return func(c *Config) {
		c.Model = model
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// NewGuardianClientWithOptions creates a GuardianClient with options
func NewGuardianClientWithOptions(opts ...ClientOption) AIClient {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	client := NewGuardianClientFromConfig(*cfg)

	return client
}

// SetAPIKey 设置API密钥和自定义URL、模型名称
func (gc *GuardianClient) SetAPIKey(apiKey string, customURL string, customModel string) {
	gc.APIKey = apiKey
	if customURL != "" {
		gc.BaseURL = customURL
	}
	if customModel != "" {
		gc.Model = customModel
	}
	// 更新ProviderConfig
	gc.ProviderConfig.APIKey = apiKey
	if customURL != "" {
		gc.ProviderConfig.BaseURL = customURL
	}
}

// CallWithMessages 调用AI服务的主要方法
func (gc *GuardianClient) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	return gc.call(systemPrompt, userPrompt)
}

// CallWithRequest 使用Request对象调用AI服务
func (gc *GuardianClient) CallWithRequest(req *Request) (string, error) {
	// 对于浏览器自动化客户端，我们简化处理
	// 分离系统提示和其他消息
	var systemPrompt string
	var userPrompt string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			userPrompt += msg.Content + "\n"
		}
	}

	return gc.call(systemPrompt, userPrompt)
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

// SetTimeout 设置超时时间
func (gc *GuardianClient) SetTimeout(timeout time.Duration) {
	gc.ProviderConfig.Timeout = timeout
}

// String returns string representation of the client
func (gc *GuardianClient) String() string {
	return fmt.Sprintf("[Provider: %s, Model: %s]", gc.ProviderConfig.Provider, gc.ProviderConfig.Model)
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

const (
	GUARDIAN_BROWSER_TIMEOUT_SECONDS = 999 // 改为999秒，确保AI有足够时间完成输出

	GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS = 999
	GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS   = 999
	GUARDIAN_AUTO_KEEP_OPEN_SECONDS         = 999 // 改为999秒，确保窗口长时间存活
)

// IsDeepSeekService checks if the current service is DeepSeek
func (gc *GuardianClient) IsDeepSeekService() bool {
	return strings.Contains(strings.ToLower(gc.ProviderConfig.Provider), "deepseek") ||
		strings.Contains(strings.ToLower(gc.BaseURL), "deepseek")
}

// ConfigureForDeepSeek configures the client specifically for DeepSeek service
func (gc *GuardianClient) ConfigureForDeepSeek() {
	gc.ProviderConfig.Provider = "deepseek"
	gc.BaseURL = "https://chat.deepseek.com"
	gc.Model = "deepseek-chat"
	gc.logger.Println("🔧 Guardian Client configured for DeepSeek service")
}

// OpenLongLivedBrowser 打开长生命周期的浏览器窗口
func (gc *GuardianClient) OpenLongLivedBrowser(targetURL string) error {
	// 启动浏览器自动化流程，但保持浏览器长时间打开
	_, err := gc.performBrowserAutomationWithKeepAlive(targetURL)
	return err
}

// performBrowserAutomationWithKeepAlive 类似于performBrowserAutomation，但保持浏览器长时间打开
func (gc *GuardianClient) performBrowserAutomationWithKeepAlive(targetURL string) (string, error) {
	// 设置Chrome选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),             // 非无头模式以便观察
		chromedp.Flag("disable-web-security", false), // 启用网络安全以支持正常网站功能
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false),                    // 启用GPU加速
		chromedp.Flag("blink-settings", "imagesEnabled=true"),  // 启用图片加载以支持验证码等功能
		chromedp.Flag("enable-automation", false),              // 防止被网站检测为自动化
		chromedp.Flag("exclude-switches", "enable-automation"), // 排除自动化开关
		chromedp.Flag("disable-extensions", false),             // 启用扩展
		chromedp.Flag("disable-plugins-discovery", false),      // 启用插件发现
		chromedp.Flag("incognito", false),                      // 不使用隐身模式
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"), // 设置正常用户代理
		chromedp.Flag("disable-blink-features", "AutomationControlled"),                                                                                // 禁用自动化控制特征
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("max_old_space_size", "4096"),
		chromedp.Flag("no-first-run", "true"),
		chromedp.Flag("no-default-browser-check", "true"),
		chromedp.Flag("window-size", "1000,850"),                  // 设置浏览器窗口尺寸为1000x850
		chromedp.Flag("user-data-dir", "./guardian_browser_data"), // 设置用户数据目录以保存登录状态
		chromedp.Flag("profile-directory", "Default"),             // 使用默认配置文件
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel() // 临时注释掉，保持浏览器窗口打开

	// 创建chrome实例上下文
	ctx, _ := chromedp.NewContext(allocCtx)
	// defer cancel() // 临时注释掉，保持浏览器窗口打开

	// 设置较长时间的超时
	timeout := time.Duration(GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS) * time.Second
	ctx, _ = context.WithTimeout(ctx, timeout)
	// defer cancel() // 临时注释掉，保持浏览器窗口打开

	gc.logger.Printf("⏰ Long-lived browser automation started, timeout: %v", timeout)

	// 访问目标URL
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	gc.logger.Printf("🌐 Opening long-lived browser window for URL: %s", targetURL)

	// 导航到目标页面
	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second), // 等待页面加载
	); err != nil {
		return "", fmt.Errorf("failed to navigate to target URL: %w", err)
	}

	gc.logger.Printf("✅ Successfully opened long-lived browser window for URL: %s", targetURL)

	// 保持浏览器打开指定时间
	time.Sleep(time.Duration(GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS) * time.Second)

	return "Browser window opened successfully", nil
}

// buildMCPRequestBody 构建MCP请求体
func (gc *GuardianClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// 对于浏览器自动化，不需要构建HTTP请求体
	return nil
}

// buildRequestBodyFromRequest 从Request构建请求体
func (gc *GuardianClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	// 对于浏览器自动化，不需要构建HTTP请求体
	return nil
}

// buildUrl 构建URL
func (gc *GuardianClient) buildUrl() string {
	// 返回配置的基本URL
	return gc.ProviderConfig.BaseURL
}

// buildRequest 构建HTTP请求
func (gc *GuardianClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	// 对于浏览器自动化，我们不使用标准HTTP请求
	return nil, fmt.Errorf("GuardianClient does not use standard HTTP requests")
}

// setAuthHeader 设置认证头
func (gc *GuardianClient) setAuthHeader(reqHeaders http.Header) {
	// 对于浏览器自动化，认证通过浏览器会话处理
}

// marshalRequestBody 序列化请求体
func (gc *GuardianClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	// 对于浏览器自动化，不需要序列化请求体
	return nil, nil
}

// parseMCPResponse 解析MCP响应
func (gc *GuardianClient) parseMCPResponse(body []byte) (string, error) {
	// 对于浏览器自动化，响应由浏览器自动化流程处理
	return "", nil
}

// isRetryableError 是否可重试错误
func (gc *GuardianClient) isRetryableError(err error) bool {
	// 对于浏览器自动化，所有错误都可以重试
	return true
}
