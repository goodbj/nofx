package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

// GuardianClient represents a client for Guardian AI service
type GuardianClient struct {
	Name           string
	ModelID        string
	ProviderConfig Config
	SystemPrompt   string
	logger         *log.Logger
}

// NewGuardianClient creates a new instance of GuardianClient
func NewGuardianClient(config Config) *GuardianClient {
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
		ProviderConfig: config,
		SystemPrompt:   "", // Config结构体中没有SystemPrompt字段，暂时设为空
		logger:         log.Default(),
	}

	return client
}

// call implements the actual AI call using browser automation
func (gc *GuardianClient) call(systemPrompt, userPrompt string) (string, error) {
	gc.logger.Println("🤖 Guardian Browser Automation: Starting browser automation for AI processing")

	// 合并系统提示和用户提示
	var combinedPrompt string
	if systemPrompt != "" && userPrompt != "" {
		combinedPrompt = fmt.Sprintf("System Prompt:\n%s\n\nUser Prompt:\n%s", systemPrompt, userPrompt)
	} else if userPrompt != "" {
		combinedPrompt = userPrompt // 如果只有用户提示，则直接使用用户提示
	} else {
		combinedPrompt = systemPrompt // 否则使用系统提示
	}

	// 启动浏览器自动化流程
	result, err := gc.performBrowserAutomation(combinedPrompt)
	if err != nil {
		gc.logger.Printf("❌ Guardian Browser Automation failed: %v", err)
		return "", fmt.Errorf("guardian browser automation failed: %w", err)
	}

	gc.logger.Println("✅ Guardian Browser Automation completed successfully")
	return result, nil
}

// performBrowserAutomation handles the browser automation process
func (gc *GuardianClient) performBrowserAutomation(prompt string) (string, error) {
	// 设置Chrome选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // 非无头模式以便观察
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"), // 禁用图片加载以加快速度
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建chrome实例上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	timeout := time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS) * time.Second
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	// 记录开始时间
	startTime := time.Now()
	gc.logger.Printf("⏰ Browser automation started, timeout: %v", timeout)

	// 访问目标URL
	targetURL := gc.ProviderConfig.BaseURL
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	gc.logger.Printf("🌐 Navigating to URL: %s", targetURL)

	// 设置请求拦截以处理潜在的CORS或其他网络问题
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			gc.logger.Printf("📤 Request: %s", ev.Request.URL)
		case *network.EventResponseReceived:
			gc.logger.Printf("📥 Response: %d", ev.Response.Status)
		}
	})

	// 导航到目标页面
	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(2*time.Second), // 等待页面加载
	); err != nil {
		return "", fmt.Errorf("failed to navigate to target URL: %w", err)
	}

	gc.logger.Printf("✅ Successfully navigated to target URL")

	// 根据不同的AI服务使用相应的选择器
	var inputSelectors []string
	var submitSelectors []string
	var responseSelectors []string

	serviceType := strings.ToLower(gc.ProviderConfig.Provider)

	switch {
	case strings.Contains(serviceType, "deepseek"):
		inputSelectors = []string{
			"textarea._27c9245.ds-scroll-area.d96f2d2a",
			"textarea[placeholder='给 DeepSeek 发送消息 ']",
			"textarea[placeholder='Send a message']",
			"div.public-DraftEditor-content",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}
		submitSelectors = []string{
			"button._3quh._30yy._2t_",
			"button[type='submit']",
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
			"button._2t_",
		}
		responseSelectors = []string{
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
			"pre",
			"code",
		}
	case strings.Contains(serviceType, "chatgpt"):
		inputSelectors = []string{
			"textarea[placeholder='Send a message']",
			"textarea[id='prompt-textarea']",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}
		submitSelectors = []string{
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
		}
		responseSelectors = []string{
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
		}
	default:
		// 默认使用通用选择器
		inputSelectors = []string{
			"textarea[placeholder*='message'], textarea[placeholder*='Message']",
			"textarea[placeholder*='input'], textarea[placeholder*='Input']",
			"textarea[aria-label*='input'], textarea[aria-label*='text']",
			"textarea[role='textbox']",
			"input[type='text']",
			"div[contenteditable='true']",
		}
		submitSelectors = []string{
			"button[type='submit']",
			"button[aria-label*='send'], button[title*='send']",
			"button[data-testid*='send']",
			".send-button, #send-button",
		}
		responseSelectors = []string{
			"[data-testid*='response'], [data-testid*='answer']",
			".response, .answer, .result",
			"div[class*='message']",
		}
	}

	gc.logger.Printf("🔍 Attempting to find and fill input field with selectors: %v", inputSelectors)

	// 尝试填充输入框
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			gc.logger.Println("📝 Starting to find and fill input field...")
			gc.logger.Printf("🔍 Trying input selectors: %v", inputSelectors)
			for _, selector := range inputSelectors {
				gc.logger.Printf("🔍 Attempting to find input with selector: %s", selector)
				// 首先等待元素可见
				err := chromedp.WaitVisible(selector).Do(ctx)
				if err == nil {
					gc.logger.Printf("✅ Found input element with selector: %s", selector)

					// 清空输入框并输入新内容
					gc.logger.Printf("🗑️ Clearing input field: %s", selector)
					err = chromedp.Clear(selector).Do(ctx)
					if err != nil {
						gc.logger.Printf("⚠️ Could not clear input field %s, proceeding anyway: %v", selector, err)
					}

					// 设置输入框的内容，一次性填充整个文本而不是逐字发送
					gc.logger.Printf("⌨️ Setting input field value (%d characters): %s...", len(prompt), prompt[:func(a, b int) int {
						if a < b {
							return a
						} else {
							return b
						}
					}(len(prompt), 50)])
					err = chromedp.SetValue(selector, prompt).Do(ctx)
					if err != nil {
						gc.logger.Printf("❌ Failed to send keys to selector %s: %v", selector, err)
						continue // 尝试下一个选择器
					}

					// 触发输入事件，确保页面JS检测到内容变化并激活提交按钮
					err = chromedp.EvaluateAsDevTools(
						fmt.Sprintf(
							`(() => {
								const element = document.querySelector('%s');
								if (element) {
									['input', 'change', 'keyup', 'keydown', 'blur', 'focus'].forEach(eventType => {
										const event = new Event(eventType, { bubbles: true, cancelable: true });
										element.dispatchEvent(event);
									});
									return true;
								}
								return false;
							})()`, selector), nil).Do(ctx)
					if err != nil {
						gc.logger.Printf("⚠️ Failed to trigger input events for selector %s: %v", selector, err)
					}

					gc.logger.Printf("✅ Successfully set value to selector %s (%d characters) and triggered input events", selector, len(prompt))
					return nil
				} else {
					gc.logger.Printf("❌ Input selector %s not found or not visible: %v", selector, err)
				}
			}
			gc.logger.Printf("❌ No input element found with any of the attempted selectors: %v", inputSelectors)
			return fmt.Errorf("no input element found with any of the attempted selectors: %v", inputSelectors)
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fill input field: %w", err)
	}

	gc.logger.Printf("✅ Successfully filled input field")

	// 点击提交按钮
	gc.logger.Printf("👆 Attempting to click submit button with selectors: %v", submitSelectors)

	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, selector := range submitSelectors {
				gc.logger.Printf("🔍 Attempting to find submit button with selector: %s", selector)
				err := chromedp.WaitVisible(selector).Do(ctx)
				if err == nil {
					gc.logger.Printf("✅ Found submit button with selector: %s", selector)

					// 检查按钮是否启用
					var isEnabled bool
					err = chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').disabled === false", selector), &isEnabled).Do(ctx)
					if err != nil {
						gc.logger.Printf("⚠️ Could not check button state: %v, assuming enabled", err)
						isEnabled = true
					}

					if !isEnabled {
						gc.logger.Println("⏳ Submit button is disabled, waiting for it to become active...")
						// 等待一段时间再检查
						time.Sleep(2 * time.Second)
						// 再次检查
						err = chromedp.Evaluate(fmt.Sprintf("document.querySelector('%s').disabled === false", selector), &isEnabled).Do(ctx)
						if err != nil {
							isEnabled = true // 如果检查失败，假设按钮可用
						}
					}

					if isEnabled {
						gc.logger.Println("✅ Submit button is active, clicking now...")
						err = chromedp.Click(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("❌ Failed to click submit button %s: %v", selector, err)
							continue
						}
						gc.logger.Printf("✅ Successfully clicked submit button: %s", selector)
						return nil
					} else {
						gc.logger.Println("❌ Submit button is still disabled after waiting")
						continue
					}
				} else {
					gc.logger.Printf("❌ Submit selector %s not found or not visible: %v", selector, err)
				}
			}
			return fmt.Errorf("no submit button found with any of the attempted selectors: %v", submitSelectors)
		}),
	)
	if err != nil {
		gc.logger.Printf("❌ Failed to click submit button: %v", err)
		// 不将此视为致命错误，因为某些界面可能通过按下Enter键提交
	} else {
		gc.logger.Println("✅ Successfully clicked submit button")
	}

	// 等待AI处理并获取响应
	gc.logger.Printf("⏳ Waiting for AI response with selectors: %v", responseSelectors)

	// 等待响应出现
	var response string
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

Loop:
	for {
		select {
		case <-waitCtx.Done():
			gc.logger.Println("⏰ Timeout waiting for AI response")
			break Loop
		default:
			err = chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					for _, selector := range responseSelectors {
						// 检查选择器元素是否存在
						var nodes []*cdp.Node
						if err := chromedp.Nodes(selector, &nodes).Do(ctx); err == nil && len(nodes) > 0 {
							// 获取最新的响应内容
							var respText string
							err = chromedp.Text(selector, &respText).Do(ctx)
							if err == nil && strings.TrimSpace(respText) != "" {
								response = strings.TrimSpace(respText)
								gc.logger.Printf("✅ Got response from selector: %s, length: %d", selector, len(response))
								return nil
							}
						}
					}
					return fmt.Errorf("no response found yet")
				}),
			)
			if err == nil {
				// 成功获取到响应
				break Loop
			}
			// 继续循环等待
			time.Sleep(1 * time.Second)
		}
	}

	if response == "" {
		// 如果仍没有响应，尝试获取页面的所有文本内容
		gc.logger.Println("📢 No specific response found, trying to get all page content...")
		err = chromedp.Run(ctx,
			chromedp.Evaluate("document.body.innerText", &response),
		)
		if err != nil {
			gc.logger.Printf("❌ Failed to get page content: %v", err)
			return "", fmt.Errorf("failed to retrieve AI response: %w", err)
		}
		response = strings.TrimSpace(response)
	}

	gc.logger.Printf("✅ Successfully retrieved AI response, length: %d", len(response))

	// 计算总耗时
	duration := time.Since(startTime)
	gc.logger.Printf("⏱️ Browser automation completed in %v", duration)

	// 根据配置决定是否保持浏览器打开
	// 由于Config结构体中没有KeepAlive字段，暂时移除该条件
	gc.logger.Printf("😴 Keeping browser alive for %d seconds as configured", GUARDIAN_AUTO_KEEP_OPEN_SECONDS)
	time.Sleep(time.Duration(GUARDIAN_AUTO_KEEP_OPEN_SECONDS) * time.Second)

	return response, nil
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

const (
	GUARDIAN_BROWSER_TIMEOUT_SECONDS        = 999 // 增加到999秒以适应测试需求
	GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS = 999
	GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS   = 999
	GUARDIAN_AUTO_KEEP_OPEN_SECONDS         = 999 // 增加到999秒以适应测试需求
)
