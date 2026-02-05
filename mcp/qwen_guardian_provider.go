package mcp

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// QwenGuardianProvider implements a specialized provider for Qwen with browser automation
// 这是千问劫匪计划的专用实现
type QwenGuardianProvider struct {
	*GuardianClient
}

// NewQwenGuardianProvider creates a new Qwen Guardian provider for the 劫匪计划
func NewQwenGuardianProvider() BrowserAIProvider {
	config := DefaultConfig()
	config.Provider = "qwen-browser"
	config.BaseURL = "https://tongyi.aliyun.com/qwen/"
	config.Model = "qwen-browser-automation"

	client := NewGuardianClientFromConfig(*config)
	provider := &QwenGuardianProvider{
		GuardianClient: client,
	}

	return provider
}

// GetServiceName returns the service name
func (p *QwenGuardianProvider) GetServiceName() string {
	return "qwen"
}

// GetDefaultURL returns the default URL for Qwen
func (p *QwenGuardianProvider) GetDefaultURL() string {
	return "https://tongyi.aliyun.com/qwen/"
}

// GetInputSelectors returns Qwen-specific input selectors for the 劫匪计划
func (p *QwenGuardianProvider) GetInputSelectors() []string {
	return []string{
		"textarea[placeholder='向千问提问']", // 您提供的具体占位符文本
		"textarea[placeholder*='提问'], textarea[placeholder*='Question']",
		"textarea[aria-label*='输入'], textarea[aria-label*='input']",
		"div[contenteditable='true'][data-slate-editor]",
		"span[data-slate-placeholder='true'][contenteditable='false']",
		".chat-input-box textarea",
		"#chat-input",
		"[data-testid='chat-input']",
	}
}

// GetSubmitSelectors returns Qwen-specific submit selectors for the 劫匪计划
func (p *QwenGuardianProvider) GetSubmitSelectors() []string {
	return []string{
		"div.operateBtn-JsB9e2",                        // Qwen提交按钮容器
		"div.operateBtn-JsB9e2 span[data-role='icon']", // Qwen图标按钮
		"div.operateBtn-JsB9e2 svg",                    // Qwen SVG图标按钮（您提供的代码）
		"button[type='submit']",
		"button[aria-label*='发送'], button[title*='发送']",
		".send-button, #send-button",
		"button:enabled:not([disabled])",
	}
}

// GetResponseSelectors returns Qwen-specific response selectors
func (p *QwenGuardianProvider) GetResponseSelectors() []string {
	return []string{
		"div.chat-message-content",
		"div.markdown-body",
		"pre",
		"code",
		".response-text",
		"[data-testid='response-container']",
	}
}

// OpenLongLivedBrowser implements browser automation for Qwen in the 劫匪计划
func (p *QwenGuardianProvider) OpenLongLivedBrowser(targetURL string) error {
	if targetURL == "" {
		targetURL = p.GetDefaultURL()
	}

	// 验证URL格式
	if _, err := url.Parse(targetURL); err != nil {
		return fmt.Errorf("invalid target URL '%s': %w", targetURL, err)
	}

	p.logger.Printf("🚀 [劫匪计划] Opening Qwen browser for target: %s", targetURL)

	// Create browser context with options - 每个实例使用独立的用户数据目录，确保每个交易员有独立的登录状态
	timestamp := time.Now().UnixNano()
	uniqueID := fmt.Sprintf("%d_%p", timestamp, p)
	userDir := fmt.Sprintf("%s_%s", GuardianBrowserDataDir, uniqueID)
	p.logger.Printf("🚨 [QWEN PROVIDER DEBUG] Qwen browser with uniqueID: '%s'", uniqueID)
	p.logger.Printf("🚨 [QWEN PROVIDER DEBUG] Qwen user data directory: %s", userDir)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"), // 确保窗口可见
		chromedp.Flag("disable-renderer-backgrounding", "true"),          // 防止后台渲染
		chromedp.Flag("disable-background-timer-throttling", "true"),     // 防止定时器节流
		chromedp.Flag("disable-background-networking", "false"),          // 允许后台网络活动
		chromedp.Flag("window-size", "1200,900"),
		chromedp.Flag("user-data-dir", userDir), // 每个实例使用独立的用户数据目录
		chromedp.Flag("profile-directory", "Default"),
		chromedp.ExecPath("C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Navigate to Qwen
	p.logger.Println("🧭 [劫匪计划] Navigating to Qwen...")
	err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to Qwen: %w", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	// Find and interact with input controls
	return p.performQwenAutomation(ctx)
}

// performQwenAutomation handles the specific Qwen automation logic for the 劫匪计划
func (p *QwenGuardianProvider) performQwenAutomation(ctx context.Context) error {
	p.logger.Println("🔍 [劫匪计划] Starting Qwen automation...")

	// Step 1: Find input controls
	inputSelectors := p.GetInputSelectors()
	var foundInputSelector string
	inputFound := false

	for _, selector := range inputSelectors {
		p.logger.Printf("🔍 [劫匪计划] Trying input selector: %s", selector)

		var nodes []*cdp.Node
		err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes, chromedp.ByQuery))
		if err == nil && len(nodes) > 0 {
			// Check if element is visible and enabled using JavaScript evaluation
			var visible bool
			err = chromedp.Run(ctx, chromedp.EvaluateAsDevTools(fmt.Sprintf(`
				(function() {
					var element = document.querySelector('%s');
					if (!element) return false;
					var style = window.getComputedStyle(element);
					return element.offsetWidth > 0 && 
					       element.offsetHeight > 0 && 
					       style.visibility !== 'hidden' && 
					       style.display !== 'none' &&
					       !element.disabled;
				})()
			`, selector), &visible))

			if err == nil && visible {
				p.logger.Printf("✅ [劫匪计划] Found visible input control: %s", selector)
				foundInputSelector = selector
				inputFound = true
				break
			}
		}
	}

	if !inputFound {
		p.logger.Printf("⚠️ [劫匪计划] Could not find any input controls with selectors: %v", inputSelectors)
		return fmt.Errorf("no input controls found")
	}

	// Step 2: Input text (using placeholder prompt for testing)
	testMessage := "这是一个测试消息用于验证千问劫匪计划"
	p.logger.Printf("📝 [劫匪计划] Attempting to input text: %s", testMessage)

	// Try different input methods
	inputMethods := []struct {
		name string
		fn   func() error
	}{
		{
			name: "SendKeys method",
			fn: func() error {
				return chromedp.SendKeys(foundInputSelector, testMessage).Do(ctx)
			},
		},
		{
			name: "Focus + SendKeys method",
			fn: func() error {
				err := chromedp.Focus(foundInputSelector).Do(ctx)
				if err != nil {
					return err
				}
				return chromedp.SendKeys(foundInputSelector, testMessage).Do(ctx)
			},
		},
	}

	inputSuccess := false
	for _, method := range inputMethods {
		p.logger.Printf("🔄 [劫匪计划] Trying input method: %s", method.name)
		err := method.fn()
		if err == nil {
			p.logger.Printf("✅ [劫匪计划] Text input successful with method: %s", method.name)
			inputSuccess = true
			break
		} else {
			p.logger.Printf("❌ [劫匪计划] Text input failed with method %s: %v", method.name, err)
		}
	}

	if !inputSuccess {
		p.logger.Println("❌ [劫匪计划] All input methods failed")
		return fmt.Errorf("failed to input text")
	}

	// Step 3: Find submit button
	submitSelectors := p.GetSubmitSelectors()
	var foundButtonSelector string
	buttonFound := false

	for _, selector := range submitSelectors {
		p.logger.Printf("🔍 [劫匪计划] Trying submit button selector: %s", selector)

		var nodes []*cdp.Node
		err := chromedp.Run(ctx, chromedp.Nodes(selector, &nodes, chromedp.ByQuery))
		if err == nil && len(nodes) > 0 {
			// Check if button is clickable
			var clickable bool
			err = chromedp.Run(ctx, chromedp.EvaluateAsDevTools(fmt.Sprintf(`
				(function() {
					var element = document.querySelector('%s');
					if (!element) return false;
					var rect = element.getBoundingClientRect();
					return rect.width > 0 && rect.height > 0 && !element.disabled;
				})()
			`, selector), &clickable))

			if err == nil && clickable {
				p.logger.Printf("✅ [劫匪计划] Found clickable submit button: %s", selector)
				buttonFound = true
				foundButtonSelector = selector
				break
			} else {
				p.logger.Printf("⚠️ [劫匪计划] Button found but not clickable: %s", selector)
			}
		} else {
			p.logger.Printf("❌ [劫匪计划] Button selector %s not found or error: %v", selector, err)
		}
	}

	if !buttonFound {
		p.logger.Printf("⚠️ [劫匪计划] Could not find any clickable submit buttons with selectors: %v", submitSelectors)
	} else {
		p.logger.Printf("✅ [劫匪计划] Successfully found submit button: %s", foundButtonSelector)
	}

	// Step 4: Click submit button (if found)
	if buttonFound && foundButtonSelector != "" {
		p.logger.Println("🖱️ [劫匪计划] Attempting to click submit button...")

		err := chromedp.Click(foundButtonSelector).Do(ctx)
		if err != nil {
			p.logger.Printf("⚠️ [劫匪计划] Failed to click submit button: %v", err)
		} else {
			p.logger.Printf("✅ [劫匪计划] Successfully clicked submit button: %s", foundButtonSelector)
		}
	} else {
		p.logger.Println("⏭️ [劫匪计划] Skipping button click - no submit button found")
	}

	// Keep browser open for manual interaction
	p.logger.Println("✅ [劫匪计划] Qwen automation completed - browser will remain open")

	// Wait indefinitely to keep browser open
	<-ctx.Done()

	return nil
}
