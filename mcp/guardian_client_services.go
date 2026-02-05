package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// NewGuardianClientWithService creates a GuardianClient configured for a specific service
func NewGuardianClientWithService(serviceType string) AIClient {
	client := NewGuardianClientWithOptions(
		WithProvider(ProviderGuardian),
		WithModel(DefaultGuardianModel),
	)

	gc, ok := client.(*GuardianClient)
	if !ok {
		return client
	}

	// Configure for the specific service
	switch strings.ToLower(serviceType) {
	case "deepseek":
		gc.ConfigureForDeepSeek()
	case "chatgpt":
		gc.ProviderConfig.Provider = "chatgpt"
		gc.BaseURL = "https://chat.openai.com"
		gc.Model = "chatgpt"
	case "claude":
		gc.ProviderConfig.Provider = "claude"
		gc.BaseURL = "https://claude.ai/chat"
		gc.Model = "claude"
	case "qwen":
		gc.ProviderConfig.Provider = "qwen"
		gc.BaseURL = "https://tongyi.aliyun.com/qwen/"
		gc.Model = "qwen"
		// Add more services as needed
	}

	return gc
}

// NewGuardianClientForBrowserWithService creates a GuardianClient for browser automation configured for a specific service
func NewGuardianClientForBrowserWithService(serviceType string) AIClient {
	client := NewGuardianClientForBrowser()

	gc, ok := client.(*GuardianClient)
	if !ok {
		return client
	}

	// Configure for the specific service
	switch strings.ToLower(serviceType) {
	case "deepseek":
		gc.ConfigureForDeepSeek()
	case "chatgpt":
		gc.ProviderConfig.Provider = "chatgpt"
		gc.BaseURL = "https://chat.openai.com"
		gc.Model = "chatgpt"
	case "claude":
		gc.ProviderConfig.Provider = "claude"
		gc.BaseURL = "https://claude.ai/chat"
		gc.Model = "claude"
	case "qwen":
		gc.ProviderConfig.Provider = "qwen"
		gc.BaseURL = "https://tongyi.aliyun.com/qwen/"
		gc.Model = "qwen"
		// Add more services as needed
	}

	return gc
}

// NewGuardianClient creates Guardian client (backward compatible)
func NewGuardianClient_() AIClient {
	return NewGuardianClientWithOptions()
}

// NewGuardianClient creates Guardian client (backward compatible)
func NewGuardianClient__() AIClient {
	return NewGuardianClientWithOptions()
}

// Legacy NewGuardianClient for backward compatibility
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

// NewGuardianClientWithOptions creates a GuardianClient with options
func NewGuardianClientWithOptions(opts ...ClientOption) AIClient {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	client := NewGuardianClientFromConfig(*cfg)

	return client
}

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

// CheckLoginStatus 检查目标AI服务的登录状态
func (gc *GuardianClient) CheckLoginStatus(targetURL string) (bool, error) {
	// 验证URL格式
	if targetURL == "" {
		targetURL = DefaultGuardianBaseURL
		gc.logger.Printf("⚠️ targetURL is empty, using default: %s", targetURL)
	}
	if _, err := url.Parse(targetURL); err != nil {
		return false, fmt.Errorf("invalid target URL '%s': %w", targetURL, err)
	}
	// 设置Chrome选项，启用用户数据目录以保持登录状态
	uniqueID := gc.TraderID
	if uniqueID == "" {
		// 如果没有交易员ID，则使用时间戳和实例地址作为后备
		timestamp := time.Now().UnixNano()
		uniqueID = fmt.Sprintf("%d_%p", timestamp, gc)
	}

	// 记录调试信息
	userDir := fmt.Sprintf("%s_%s", GuardianBrowserDataDir, uniqueID)
	gc.logger.Printf("🚨 [GUARDIAN LOGIN DEBUG] CheckLoginStatus with TraderID: '%s', uniqueID: '%s'", gc.TraderID, uniqueID)
	gc.logger.Printf("🚨 [GUARDIAN LOGIN DEBUG] Login check user data directory: %s", userDir)

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
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"), // 确保窗口可见
		chromedp.Flag("disable-renderer-backgrounding", "true"),          // 防止后台渲染
		chromedp.Flag("disable-background-timer-throttling", "true"),     // 防止定时器节流
		chromedp.Flag("disable-background-networking", "false"),          // 允许后台网络活动
		chromedp.Flag("window-size", "1000,850"),                         // 设置浏览器窗口尺寸为1000x850
		chromedp.Flag("user-data-dir", userDir),                          // 每个交易员使用独立的用户数据目录
		chromedp.Flag("profile-directory", "Default"),                    // 使用默认配置文件
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建chrome实例上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置较短的超时时间用于登录检查
	ctx, timeoutCancel := context.WithTimeout(ctx, 10*time.Second)
	defer timeoutCancel()

	// 访问目标URL
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	gc.logger.Printf("🌐 Checking login status at URL: %s", targetURL)

	// 导航到目标页面并检查登录状态
	var isLoggedIn bool
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(3*time.Second), // 等待页面加载
		chromedp.Evaluate(`(() => {
			// 检查是否存在登录相关的元素（如登录按钮、用户名显示等）
			const loginIndicators = [
				'button:contains("Login")',
				'button:contains("Sign in")',
				'a:contains("Login")',
				'a:contains("Sign in")',
				'[href*="login"]',
				'[href*="signin"]',
				'#login',
				'.login-form',
				'.auth-form'
			];
			
			// 检查是否已登录（通常表现为用户头像、用户名等元素）
			const loggedInIndicators = [
				'[data-testid="user-menu"]',
				'.user-avatar',
				'.username',
				'.account-menu',
				'#avatar',
				'.profile-image',
				'[title="Logout"]',
				'[aria-label="User menu"]'
			];
			
			// 检查登录指示器
			for (let selector of loggedInIndicators) {
				if (document.querySelector(selector)) {
					return true; // 发现已登录指示器
				}
			}
			
			// 检查未登录指示器
			for (let selector of loginIndicators) {
				if (document.querySelector(selector)) {
					return false; // 发现未登录指示器
				}
			}
			
			// 如果没有明确的登录/未登录指示器，检查是否有输入框（通常未登录页面不会有主要输入框）
			const inputBox = document.querySelector('textarea[placeholder*="message"], textarea[placeholder*="input"], input[type="text"]');
			if (inputBox) {
				// 如果有输入框，可能是已登录状态
				return true;
			}
			
			// 默认返回假定未登录
			return false;
		})()`, &isLoggedIn),
	)

	if err != nil {
		// 如果评估失败，假设未登录
		gc.logger.Printf("⚠️ Error checking login status: %v, assuming not logged in", err)
		return false, nil
	}

	if isLoggedIn {
		gc.logger.Println("✅ User appears to be logged in")
	} else {
		gc.logger.Println("❌ User appears to be not logged in")
	}

	return isLoggedIn, nil
}
