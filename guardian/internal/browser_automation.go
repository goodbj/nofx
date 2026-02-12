package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nofx/config"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserAutomation 浏览器自动化控制器
type BrowserAutomation struct {
	config  *config.BrowserConfig
	ctx     context.Context
	cancel  context.CancelFunc
	browser context.Context // chromedp浏览器上下文
	isReady bool
}

// NewBrowserAutomation 创建浏览器自动化实例
func NewBrowserAutomation(browserConfig *config.BrowserConfig) (*BrowserAutomation, error) {
	ctx, cancel := context.WithCancel(context.Background())

	ba := &BrowserAutomation{
		config:  browserConfig,
		ctx:     ctx,
		cancel:  cancel,
		isReady: false,
	}

	// 初始化浏览器
	err := ba.InitBrowser()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	return ba, nil
}

// InitBrowser 初始化浏览器
func (ba *BrowserAutomation) InitBrowser() error {
	log.Println("🌐 Initializing browser automation...")

	// 设置chromedp选项
	options := []chromedp.ExecAllocatorOption{}

	if ba.config.ExecutablePath != "" {
		options = append(options, chromedp.ExecPath(ba.config.ExecutablePath))
	}

	// 添加规避检测的选项
	options = append(options,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-plugins-discovery", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("allow-running-insecure-content", true),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-features", "TranslateUI,VizDisplayCompositor"),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-plugins", true),
		chromedp.Flag("disable-image-animation-resampling", true),
		chromedp.Flag("disable-session-crashed-bubble", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-field-trial-config", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-ipc-flooding-protection", false),
		chromedp.Flag("remote-debugging-port", "9222"),
	)

	if ba.config.Headless {
		options = append(options, chromedp.Headless)
	}

	if ba.config.UserDataDir != "" {
		options = append(options, chromedp.UserDataDir(ba.config.UserDataDir))
	}

	if ba.config.DisableImages {
		options = append(options, chromedp.Flag("disable-images", true))
	}

	if ba.config.DisableJS {
		options = append(options, chromedp.Flag("disable-javascript", true))
	}

	// 添加额外参数
	for _, arg := range ba.config.AdditionalArgs {
		options = append(options, chromedp.Flag(strings.Split(arg, "=")[0], strings.Split(arg, "=")[1]))
	}

	// 创建分配器上下文
	allocatorCtx, browserCancel := chromedp.NewExecAllocator(context.Background(), options...)
	ba.browser, ba.cancel = chromedp.NewContext(allocatorCtx)
	ba.cancel = browserCancel // Override with allocator cancel to close browser properly

	// 启动浏览器
	if err := chromedp.Run(ba.browser); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}

	ba.isReady = true
	log.Println("✅ Browser automation ready")
	// 注入JavaScript来隐藏webdriver属性
	go func() {
		chromedp.Run(ba.browser,
			chromedp.Evaluate(
				`(function(){
					Object.defineProperty(navigator, 'webdriver', {
						get: () => undefined,
				});
				window.chrome = {
					runtime: {}
				};
				Object.defineProperty(navigator, 'plugins', {
					get: () => [1, 2, 3, 4, 5],
				});
				Object.defineProperty(navigator, 'languages', {
					get: () => ['zh-CN', 'zh', 'en'],
				});
			})())`, nil,
			),
		)
	}()
	return nil
}

// ProcessPrompt 处理提示词，发送到浏览器AI服务并获取响应
func (ba *BrowserAutomation) ProcessPrompt(systemPrompt, userPrompt string) (string, error) {
	if !ba.isReady {
		return "", fmt.Errorf("browser automation not ready")
	}

	log.Println("💬 Processing prompt in browser AI service...")

	// 导航到AI服务页面
	err := ba.navigateToPage()
	if err != nil {
		return "", fmt.Errorf("failed to navigate to AI page: %w", err)
	}

	// 清空输入框（如果存在之前的对话）
	err = ba.clearChat()
	if err != nil {
		log.Printf("⚠️ Could not clear chat: %v", err)
		// 继续执行，不清空可能也无妨
	}

	// 发送提示词到AI服务
	err = ba.sendPrompt(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to send prompt: %w", err)
	}

	// 等待AI响应
	response, err := ba.waitForResponse()
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	log.Printf("✅ AI response received, length: %d", len(response))
	return response, nil
}

// navigateToPage 导航到AI服务页面
func (ba *BrowserAutomation) navigateToPage() error {
	log.Printf("🧭 Navigating to AI service page...")

	// 导航到指定的AI服务页面
	err := chromedp.Run(ba.browser,
		chromedp.Navigate("https://chat.deepseek.com"),                               // 默认使用DeepSeek，可以通过配置更改
		chromedp.WaitVisible("textarea[data-testid='chat-input']", chromedp.ByQuery), // 等待输入框可见
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to page: %w", err)
	}

	log.Println("✅ Navigation completed")
	return nil
}

// clearChat 清空聊天记录
func (ba *BrowserAutomation) clearChat() error {
	log.Println("🧹 Clearing chat history...")

	// 尝试找到并点击清除按钮
	// 注意：这取决于具体网站的DOM结构，需要根据实际情况调整
	err := chromedp.Run(ba.browser,
		chromedp.Click("button[aria-label='Clear conversation']", chromedp.ByQuery),
	)

	// 如果清除按钮不存在，忽略错误
	if err != nil {
		log.Printf("⚠️ Could not find clear button, skipping: %v", err)
		// 不返回错误，因为这通常是可选操作
	}

	log.Println("✅ Chat cleared or clear button not found")
	return nil
}

// sendPrompt 发送提示词到AI服务
func (ba *BrowserAutomation) sendPrompt(systemPrompt, userPrompt string) error {
	log.Printf("📤 Sending prompt to AI service, user prompt length: %d", len(userPrompt))

	// 将提示词发送到输入框并点击发送
	err := chromedp.Run(ba.browser,
		chromedp.Clear("textarea[data-testid='chat-input']", chromedp.ByQuery),
		chromedp.SetValue("textarea[data-testid='chat-input']", userPrompt, chromedp.ByQuery),
		chromedp.Click("button[data-testid='chat-send-button']", chromedp.ByQuery),
	)

	if err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}

	log.Println("✅ Prompt sent to AI service")
	return nil
}

// waitForResponse 等待AI服务响应
func (ba *BrowserAutomation) waitForResponse() (string, error) {
	log.Println("⏳ Waiting for AI response...")

	var responseText string
	timeout := time.After(60 * time.Second) // 60秒超时
	tick := time.Tick(2 * time.Second)      // 每2秒检查一次

	// 等待AI响应，直到超时
	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for AI response")
		case <-tick:
			// 尝试获取AI响应
			err := chromedp.Run(ba.browser,
				chromedp.Text("div[data-testid='chat-response']", &responseText, chromedp.ByQuery),
			)

			if err == nil && responseText != "" {
				log.Println("✅ Response received from AI service")
				// 尝试解析响应为JSON格式，如果不是JSON则返回原始文本
				var parsed interface{}
				err = json.Unmarshal([]byte(responseText), &parsed)
				if err != nil {
					// 如果不是有效JSON，包装成适当的响应格式
					wrappedResponse := map[string]interface{}{
						"raw_response": responseText,
						"decisions":    []interface{}{},
					}
					wrappedBytes, _ := json.Marshal(wrappedResponse)
					return string(wrappedBytes), nil
				}
				return responseText, nil
			}
		}
	}
}

// Close 关闭浏览器自动化
func (ba *BrowserAutomation) Close() {
	log.Println("🛑 Closing browser automation...")

	if ba.cancel != nil {
		ba.cancel()
	}

	// 在实际实现中，这里会关闭浏览器实例
	log.Println("✅ Browser automation closed")
}
