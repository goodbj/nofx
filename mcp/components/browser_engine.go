package components

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"nofx/mcp" // Import mcp to access GuardianBrowserDataDir constant

	"github.com/chromedp/chromedp"
)

// BrowserAutomationEngine handles browser automation tasks
type BrowserAutomationEngine struct {
	logger         *log.Logger
	displayEnabled bool
	timeout        time.Duration
}

// NewBrowserAutomationEngine creates a new browser automation engine
func NewBrowserAutomationEngine(logger *log.Logger, displayEnabled bool, timeout time.Duration) *BrowserAutomationEngine {
	return &BrowserAutomationEngine{
		logger:         logger,
		displayEnabled: displayEnabled,
		timeout:        timeout,
	}
}

// NavigateToURL navigates to the specified URL
func (bae *BrowserAutomationEngine) NavigateToURL(ctx context.Context, targetURL string) error {
	// 验证URL格式
	if targetURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if _, err := url.Parse(targetURL); err != nil {
		return fmt.Errorf("invalid URL '%s': %w", targetURL, err)
	}

	bae.logger.Printf("Navigating to URL: %s", targetURL)

	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second), // 等待页面加载
	); err != nil {
		return err
	}

	bae.logger.Printf("Successfully navigated to URL: %s", targetURL)
	return nil
}

// SetupChromeOptions sets up Chrome options for the browser
func (bae *BrowserAutomationEngine) SetupChromeOptions() []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !bae.displayEnabled), // 根据设置决定是否无头模式
		chromedp.Flag("disable-web-security", false),   // 启用网络安全以支持正常网站功能
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", !bae.displayEnabled),      // 如果启用显示，则启用GPU加速
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
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"),                                                // 确保窗口可见
		chromedp.Flag("disable-renderer-backgrounding", "true"),                                                         // 防止后台渲染
		chromedp.Flag("disable-background-timer-throttling", "true"),                                                    // 防止定时器节流
		chromedp.Flag("disable-background-networking", "false"),                                                         // 允许后台网络活动
		chromedp.Flag("window-size", "1000,850"),                                                                        // 设置浏览器窗口尺寸为1000x850
		chromedp.Flag("user-data-dir", fmt.Sprintf("%s_%d_%p", mcp.GuardianBrowserDataDir, time.Now().UnixNano(), bae)), // 每个实例使用独立的用户数据目录
		chromedp.Flag("profile-directory", "Default"),                                                                   // 使用默认配置文件
	)

	return opts
}

// GetContext creates a new Chrome context with the configured options
func (bae *BrowserAutomationEngine) GetContext(parentCtx context.Context) (context.Context, context.CancelFunc, error) {
	opts := bae.SetupChromeOptions()
	allocCtx, cancel := chromedp.NewExecAllocator(parentCtx, opts...)

	// 创建chrome实例上下文
	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	// 设置超时
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, bae.timeout)

	// 创建组合的cancel函数
	combinedCancel := func() {
		timeoutCancel()
		ctxCancel()
		cancel()
	}

	return timeoutCtx, combinedCancel, nil
}
