package mcp

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

const (
	ProviderGuardian       = "guardian"
	DefaultGuardianBaseURL = "https://chat.deepseek.com"
	DefaultGuardianModel   = "guardian-browser-automation"
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
		ModelID:        "guardian-browser-automation",
		APIKey:         config.APIKey,
		BaseURL:        config.BaseURL,
		Model:          config.Model,
		MaxTokens:      config.MaxTokens,
		Temperature:    config.Temperature,
		ProviderConfig: config,
		SystemPrompt:   "", // Config结构体中没有SystemPrompt字段，暂时设为空
		logger:         log.Default(),
		httpClient:     config.HTTPClient,
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
	// 初始化响应变量
	var response string

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

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel() // 恢复取消函数，确保浏览器在完成任务后关闭

	// 创建chrome实例上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel() // 恢复取消函数，确保浏览器在完成任务后关闭

	// 设置超时
	timeout := time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS) * time.Second
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel() // 恢复取消函数，确保浏览器在完成任务后关闭

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
	case strings.Contains(serviceType, "chat"):
	case targetURL == "https://chat.deepseek.com" || strings.Contains(targetURL, "deepseek"):
		// 如果是DeepSeek或包含chat的服务，或目标URL包含deepseek
		inputSelectors = []string{
			"textarea[placeholder='给 DeepSeek 发送消息 ']",
			"textarea._27c9245.ds-scroll-area.d96f2d2a",
			"textarea._27c9245.ds-scroll-area.d96f2d2a[placeholder='给 DeepSeek 发送消息 ']",
			"textarea[placeholder='Send a message']",
			"div.public-DraftEditor-content",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}
		submitSelectors = []string{
			"div._7436101.ds-icon-button.ds-icon-button--l.ds-icon-button--sizing-container[role='button'][aria-disabled='false']", // DeepSeek特有按钮样式 - 启用状态
			"div._7436101.ds-icon-button[role='button']:not([aria-disabled='true'])",                                               // DeepSeek特有按钮样式 - 启用状态
			"div.ds-icon-button[role='button']:not([aria-disabled='true'])",                                                        // 启用状态的按钮
			"button._3quh._30yy._2t_",
			"button[type='submit']",
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
			"button._2t_",
			"button[type='button'][aria-label='Send Message']",
			"button:enabled:not([disabled])",
		}
		responseSelectors = []string{
			"div.ds-flex._0a3d93b", // 主要的AI输出完成标识容器
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
			"div.ds-flex._0a3d93b", // 主要的AI输出完成标识容器
		}
	case strings.Contains(serviceType, "other"):
		// 预留其他服务类型
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
			"div.ds-flex._0a3d93b", // 主要的AI输出完成标识容器
		}
	}

	gc.logger.Printf("🔍 Attempting to find and fill input field with selectors: %v", inputSelectors)

	// 尝试填充输入框
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			gc.logger.Println("📝 Starting to find and fill input field...")
			gc.logger.Printf("🔍 Trying input selectors: %v", inputSelectors)

			inputFound := false // 标志变量，跟踪是否成功找到并输入

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

					// 使用chromedp.Focus和SetValue方法快速填充文本并激活控件
					// 1. 选中控件并聚焦
					// 2. 使用SetValue设置值
					// 3. 触发必要的事件以激活提交按钮
					textLength := len(strings.ReplaceAll(prompt, "\n", " ")) // 使用原始长度计算用于日志
					gc.logger.Printf("🚀 Fast input mode: Setting %d characters using chromedp.Focus + SetValue...", textLength)

					// 首先使用chromedp.Focus方法聚焦元素
					err = chromedp.Focus(selector).Do(ctx)
					if err != nil {
						gc.logger.Printf("❌ Focus activation failed with chromedp.Focus: %v", err)
						// 直接跳过此选择器
						gc.logger.Printf("⏭️ Skipping this selector due to focus activation failure")
						continue // 继续尝试下一个选择器
					} else {
						gc.logger.Printf("✅ Control focused successfully with chromedp.Focus, now using optimized chunked input approach...")

						// 首先清空输入框
						err = chromedp.Clear(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Could not clear input field, proceeding anyway: %v", err)
						}

						// 使用SendKeys方法批量输入内容，模拟快速用户输入
						// 根据文本大小调整块大小和延迟，以优化输入速度
						textToInput := prompt
						var chunkSize int
						var delay time.Duration

						// 根据文本长度动态调整块大小和延迟
						if len(textToInput) > 100000 { // 超大文本 (>100K)
							chunkSize = 5000 // 大块大小以提高效率
							delay = 0        // 无延迟，最大化速度
						} else if len(textToInput) > 50000 { // 大文本 (50K-100K)
							chunkSize = 2000 // 较大块大小
							delay = 0        // 无延迟
						} else if len(textToInput) > 10000 { // 中等文本 (10K-50K)
							chunkSize = 1000 // 中等块大小
							delay = 0        // 无延迟
						} else { // 普通大小文本
							chunkSize = 500              // 较小块大小保证稳定性
							delay = 1 * time.Millisecond // 很小的延迟
						}

						gc.logger.Printf("📦 Using chunk size: %d and delay: %v for text of length: %d", chunkSize, delay, len(textToInput))

						// 首先等待元素可见，增加等待时间以确保页面加载完成
						err = chromedp.WaitVisible(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Selector not immediately visible: %v", err)
							// 如果元素不可见，等待一段时间后再次尝试
							err = chromedp.Sleep(1 * time.Second).Do(ctx)
							if err != nil {
								gc.logger.Printf("⚠️ Sleep failed: %v", err)
							}
							// 再次尝试等待元素可见
							err = chromedp.WaitVisible(selector).Do(ctx)
							if err != nil {
								gc.logger.Printf("❌ Selector still not visible after waiting: %v", err)
								// 直接跳过此选择器
								gc.logger.Printf("⏭️ Skipping this selector due to visibility failure")
								continue // 继续尝试下一个选择器
							}
						}

						gc.logger.Printf("✅ Selector is visible, now focusing and setting value...")

						// 使用chromedp.Focus方法聚焦元素
						err = chromedp.Focus(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Focus activation failed with chromedp.Focus: %v", err)
							// 即使聚焦失败也继续，因为SetValue可以直接设置值
						}

						// 首先清空输入框
						err = chromedp.Clear(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Could not clear input field, proceeding anyway: %v", err)
						}

						// 首先点击输入框以获得焦点
						err = chromedp.Click(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("❌ Click input failed: %v", err)
							// 直接跳过此选择器
							gc.logger.Printf("⏭️ Skipping this selector due to click failure")
							continue // 继续尝试下一个选择器
						}

						// 短暂等待确保焦点设置完成
						time.Sleep(100 * time.Millisecond)

						// 首先聚焦输入框
						err = chromedp.Focus(selector).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Focus failed: %v", err)
						}

						// 先设置完整内容
						err = chromedp.SetValue(selector, prompt).Do(ctx)
						if err != nil {
							gc.logger.Printf("❌ SetValue failed: %v", err)
							// 直接跳过此选择器
							gc.logger.Printf("⏭️ Skipping this selector due to SetValue failure")
							continue // 继续尝试下一个选择器
						}

						gc.logger.Printf("✅ Full content set with %d characters using SetValue approach", len(prompt))

						// 根据您的发现，现在模拟输入一个字符来激活按钮
						// 这个关键操作会触发前端框架的状态更新
						err = chromedp.SendKeys(selector, " ").Do(ctx) // 发送一个空格字符
						if err != nil {
							gc.logger.Printf("⚠️ SendKeys activation character failed: %v", err)
							// 即使发送激活字符失败，也要继续，因为主要内容已经设置了
						} else {
							gc.logger.Printf("✅ Activation character sent to trigger button state")
						}

						// 然后删除这个额外的空格字符，恢复原始内容
						// 通过JavaScript模拟Backspace键
						backspaceScript := fmt.Sprintf(`
							(function() {
								var element = document.querySelector('%s');
								if (element) {
									// 获取当前值并去掉最后一个字符（空格）
									var currentValue = element.value;
									if (currentValue.length > 0) {
										element.value = currentValue.substring(0, currentValue.length - 1);
										
										// 触发Backspace相关的事件
										var keydownEvent = new KeyboardEvent('keydown', {
											bubbles: true,
											cancelable: true,
											key: 'Backspace',
											code: 'Backspace',
											keyCode: 8,
											which: 8
										});
										element.dispatchEvent(keydownEvent);
										
										var inputEvent = new InputEvent('input', {
											bubbles: true,
											cancelable: true,
											inputType: 'deleteContentBackward',
											data: null
										});
										element.dispatchEvent(inputEvent);
										
										var keyupEvent = new KeyboardEvent('keyup', {
											bubbles: true,
											cancelable: true,
											key: 'Backspace',
											code: 'Backspace',
											keyCode: 8,
											which: 8
										});
										element.dispatchEvent(keyupEvent);
										
										return true;
									}
								}
								return false;
							})();
						`, selector)

						var backspaceResult bool
						err = chromedp.Evaluate(backspaceScript, &backspaceResult).Do(ctx)
						if err != nil || !backspaceResult {
							gc.logger.Printf("⚠️ JavaScript Backspace simulation failed: %v, success: %v", err, backspaceResult)
						} else {
							gc.logger.Printf("✅ Removed activation character via JavaScript")
						}

						// 再次触发事件以确保状态更新
						ensureStateUpdateScript := fmt.Sprintf(`
							(function() {
								try {
									var element = document.querySelector('%s');
									if (element) {
										// 触发关键事件来激活前端框架状态
										var events = ['input', 'change', 'propertychange'];
										events.forEach(function(eventName) {
											var event = new Event(eventName, { bubbles: true, cancelable: true });
											element.dispatchEvent(event);
										});
																
										return true;
									} else {
										return false;
									}
								} catch (e) {
									console.error('State update failed:', e);
									return false;
								}
							})();
						`, selector)

						var stateUpdated bool
						err = chromedp.Evaluate(ensureStateUpdateScript, &stateUpdated).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ State update script failed: %v", err)
						} else if !stateUpdated {
							gc.logger.Printf("⚠️ State update script returned false")
						} else {
							gc.logger.Printf("✅ State updated successfully after activation sequence")
						}

						gc.logger.Printf("✅ Successfully filled input field with %d characters total using activation sequence approach", len(prompt))

						// 等待操作完成
						time.Sleep(200 * time.Millisecond)

						gc.logger.Printf("✅ Successfully filled input field with %d characters total using JavaScript paste simulation approach", len(prompt))

						// 增加延迟时间，确保页面有充分时间处理输入并激活提交按钮
						// 对于超大文本，增加更多的等待时间
						baseDelay := 500 * time.Millisecond
						if textLength > 50000 {
							baseDelay = 1000 * time.Millisecond // 1秒延迟用于超大文本
						} else if textLength > 10000 {
							baseDelay = 800 * time.Millisecond // 0.8秒延迟用于大文本
						} else {
							baseDelay = 600 * time.Millisecond // 0.6秒延迟用于普通文本
						}
						time.Sleep(baseDelay)

						// 再次检查并确保提交按钮被激活
						// 在等待后再次尝试触发事件来激活提交按钮
						finalActivationScript := fmt.Sprintf(`
							(function() {
								try {
									var element = document.querySelector('%s');
									if (element) {
										// 再次触发关键事件以确保提交按钮激活
										var finalEvents = ['input', 'change', 'keyup'];
										finalEvents.forEach(function(eventName) {
											var event = new Event(eventName, { bubbles: true, cancelable: true });
											element.dispatchEvent(event);
										});
										return true;
									} else {
										return false;
									}
								} catch (e) {
									return false;
								}
							})();
						`, selector)

						var finalActivationSuccess bool
						err = chromedp.Evaluate(finalActivationScript, &finalActivationSuccess).Do(ctx)
						if err != nil {
							gc.logger.Printf("⚠️ Final activation check failed: %v", err)
						} else if !finalActivationSuccess {
							gc.logger.Printf("⚠️ Final activation check returned false")
						} else {
							gc.logger.Printf("✅ Final activation check successful")
						}

						// 标记输入成功
						inputFound = true

						// 输入完成，跳出选择器循环，继续执行提交按钮逻辑
						break // 跳出选择器循环，继续执行提交按钮逻辑
					}
				} else {
					gc.logger.Printf("❌ Input selector %s not found or not visible: %v", selector, err)
				}
			}

			// 只有在所有选择器都失败的情况下才返回错误
			if !inputFound {
				gc.logger.Printf("❌ No input element found with any of the attempted selectors: %v", inputSelectors)
				return fmt.Errorf("no input element found with any of the attempted selectors: %v", inputSelectors)
			}

			// 如果找到了输入元素并成功输入，则不返回错误
			return nil
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fill input field: %w", err)
	}

	gc.logger.Printf("✅ Successfully filled input field")

	// 点击提交按钮
	gc.logger.Printf("👆 Attempting to click submit button with selectors: %v", submitSelectors)

	// 处理换行符，确保长度比较使用的是实际输入到文本框的内容长度
	safePrompt := strings.ReplaceAll(prompt, "\n", " ")

	// 等待输入框内容长度与处理后提示词长度一致后点击提交按钮
	gc.logger.Printf("⏳ Waiting for input content to match processed prompt length (%d characters)", len(safePrompt))

	// 使用优化的chromedp操作来等待输入完成并点击提交按钮
	waitErr := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 根据文本长度动态调整最大等待时间
			var maxWaitDuration time.Duration
			textLength := len(safePrompt)
			if textLength > 100000 { // 超大文本
				maxWaitDuration = 15 * time.Second // 给予足够的等待时间给超大文本
			} else if textLength > 50000 { // 大文本
				maxWaitDuration = 10 * time.Second // 给予足够的等待时间
			} else if textLength > 10000 { // 中等文本
				maxWaitDuration = 7 * time.Second // 给予足够的等待时间
			} else { // 小文本
				maxWaitDuration = 5 * time.Second // 标准等待时间
			}

			maxWait := time.Now().Add(maxWaitDuration)
			gc.logger.Printf("⏳ Waiting for input completion with max timeout: %v (text length: %d)", maxWaitDuration, textLength)

			for time.Now().Before(maxWait) {
				// 尝试找到输入框并检查其内容长度
				var inputLength int
				foundInput := false

				for _, inputSelector := range inputSelectors {
					var inputValue string
					err := chromedp.Value(inputSelector, &inputValue).Do(ctx)
					if err == nil && len(inputValue) > 0 {
						inputLength = len(inputValue)
						foundInput = true
						gc.logger.Printf("📊 Input box length check: %d/%d characters", inputLength, len(safePrompt))

						// 根据文本大小调整阈值
						expectedLength := len(safePrompt)
						var thresholdPercent int
						if textLength > 100000 {
							thresholdPercent = 85 // 超大文本需要85%才算完成（更快响应）
						} else if textLength > 50000 {
							thresholdPercent = 85 // 大文本需要85%
						} else if textLength > 10000 {
							thresholdPercent = 80 // 中等文本需要80%
						} else {
							thresholdPercent = 95 // 小文本需要95%（更精确）
						}

						if inputLength >= expectedLength*thresholdPercent/100 { // 达到阈值即可
							gc.logger.Printf("✅ Input length (%d) reached threshold (%d%%) of expected (%d), attempting to click submit button", inputLength, thresholdPercent, expectedLength)

							// 尝试点击提交按钮
							for _, submitSelector := range submitSelectors {
								gc.logger.Printf("👆 Clicking submit button with selector: %s", submitSelector)
								err := chromedp.Click(submitSelector).Do(ctx)
								if err == nil {
									gc.logger.Printf("✅ Successfully clicked submit button: %s", submitSelector)
									return nil // 成功点击，退出ActionFunc
								} else {
									gc.logger.Printf("⚠️ Failed to click submit button %s: %v, trying next selector", submitSelector, err)
								}
							}
							// 如果所有提交按钮都点击失败，继续等待
						}
						break
					}
				}

				if !foundInput {
					gc.logger.Println("⚠️ Could not find input box, continuing to wait...")
				}

				// 根据文本大小调整检查频率 - 现在输入更快，可以更频繁地检查
				var checkInterval time.Duration
				if textLength > 100000 {
					checkInterval = 200 * time.Millisecond // 超大文本使用适中的检查频率
				} else {
					checkInterval = 100 * time.Millisecond // 更频繁的检查，因为输入更快了
				}

				time.Sleep(checkInterval)
			}
			return fmt.Errorf("timeout waiting for input length to match processed prompt length")
		}),
	)

	if waitErr != nil {
		gc.logger.Printf("⚠️ Wait for input completion failed or timeout: %v", waitErr)
		// 即使等待失败，我们也尝试点击提交按钮，以防万一
		gc.logger.Println("🔄 Proceeding to try clicking submit button anyway...")
		// 尝试点击提交按钮
		clickErr := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				for _, submitSelector := range submitSelectors {
					gc.logger.Printf("👆 Clicking submit button with selector: %s", submitSelector)
					err := chromedp.Click(submitSelector).Do(ctx)
					if err == nil {
						gc.logger.Printf("✅ Successfully clicked submit button: %s", submitSelector)
						return nil
					} else {
						gc.logger.Printf("⚠️ Failed to click submit button %s: %v, trying next selector", submitSelector, err)
					}
				}
				return fmt.Errorf("failed to click any submit button")
			}),
		)
		if clickErr != nil {
			gc.logger.Printf("⚠️ All submit button attempts failed: %v", clickErr)
		}
	}

	// 提交按钮点击已在上面的逻辑中处理
	gc.logger.Println("✅ Submit button processing completed")

	// 在等待AI响应之前，先输出当前页面的DOM结构用于调试
	gc.logger.Println("🔍 Setting up DOM snapshot for debugging...")
	// 等待1秒让页面更新后再输出DOM
	time.Sleep(1 * time.Second)

	// 添加标志来跟踪是否已经通过优化方法提取了AI内容
	var aiContentExtractedSuccessfully bool = false

	// 使用新的等待和验证机制检测AI输出完成
	gc.logger.Println("⏳ Waiting for AI to complete output using optimized detection mechanism...")
	validationCtx, validationCancel := context.WithTimeout(ctx, time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS)*time.Second)
	extractedContent, err := gc.waitForAndValidateOutput(validationCtx)
	validationCancel() // 确保取消上下文以释放资源
	// 无论waitForAndValidateOutput是否返回错误，只要有内容被提取，我们都认为AI内容已成功处理
	if extractedContent != "" {
		gc.logger.Println("✅ AI output completed successfully detected and content extracted")
		// 设置标志，表示AI内容已成功提取
		aiContentExtractedSuccessfully = true
		response = extractedContent
		gc.logger.Printf("✅ Using extracted content as response, length: %d", len(response))
	} else if err != nil {
		gc.logger.Printf("⚠️ Error waiting for AI completion: %v", err)
		// 继续执行，即使等待出现问题
	} else {
		gc.logger.Println("✅ AI output completed successfully detected (with empty content)")
		// 即使内容为空，也表示AI处理过程已完成
		aiContentExtractedSuccessfully = true
		gc.logger.Printf("⚠️ AI content was extracted but became empty after cleaning, still marking as extracted")
	}

	// 重新启用：保存网页到文件功能
	// var pageHTML string
	// err = chromedp.Run(ctx,
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		// 获取完整的页面HTML内容
	// 		err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// 将完整的HTML内容保存到文件
	// 		os.MkdirAll("temp", 0755) // 确保temp目录存在
	// 		filename := fmt.Sprintf("temp/dom_snapshot_%d.html", time.Now().Unix())
	// 		err = os.WriteFile(filename, []byte(pageHTML), 0644)
	// 		if err == nil {
	// 			gc.logger.Printf("📄 Full DOM snapshot saved to %s, size: %d bytes", filename, len(pageHTML))
	// 		} else {
	// 			gc.logger.Printf("⚠️ Error saving DOM snapshot to file: %v", err)
	// 			// 如果保存文件失败，至少输出部分HTML内容到日志
	// 			maxLength := 2000
	// 			if len(pageHTML) < maxLength {
	// 				maxLength = len(pageHTML)
	// 			}
	// 			gc.logger.Printf("📄 DOM Snapshot (first %d chars):\n%s", maxLength, pageHTML[:maxLength])
	// 		}
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	gc.logger.Printf("⚠️ Error getting page HTML: %v", err)
	// }

	// 启用：增强的DOM抓取和调试功能
	// 获取页面上所有ds-theme元素的详细信息
	// var themeElements []map[string]interface{}
	// err = chromedp.Run(ctx,
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		err := chromedp.Evaluate(`(() => {
	// 			const elements = document.querySelectorAll('div.ds-theme');
	// 			return Array.from(elements).map(el => ({
	// 				tagName: el.tagName,
	// 				className: el.className,
	// 				style: el.style.cssText,
	// 				innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
	// 				attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
	// 					acc[attr.name] = attr.value;
	// 					return acc;
	// 				}, {})
	// 			}));
	// 		})()`, &themeElements).Do(ctx)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// 将ds-theme元素详情也保存到文件
	// 		os.MkdirAll("temp", 0755) // 确保temp目录存在
	// 		themeFilename := fmt.Sprintf("temp/ds_theme_elements_%d.json", time.Now().Unix())
	// 		themeJSON, jsonErr := json.MarshalIndent(themeElements, "", "  ")
	// 		if jsonErr == nil {
	// 			err = os.WriteFile(themeFilename, themeJSON, 0644)
	// 			if err == nil {
	// 				gc.logger.Printf("🔍 ds-theme Elements Detail saved to %s", themeFilename)
	// 			} else {
	// 				gc.logger.Printf("⚠️ Error saving ds-theme elements to file: %v", err)
	// 				gc.logger.Printf("🔍 ds-theme Elements Detail: %+v", themeElements)
	// 			}
	// 		} else {
	// 			gc.logger.Printf("⚠️ Error marshaling ds-theme elements to JSON: %v", jsonErr)
	// 			gc.logger.Printf("🔍 ds-theme Elements Detail: %+v", themeElements)
	// 		}
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	gc.logger.Printf("⚠️ Error getting ds-theme elements detail: %v", err)
	// }

	// 也获取页面上所有可能的AI完成标志元素
	// var aiCompletionElements []map[string]interface{}
	// err = chromedp.Run(ctx,
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		err := chromedp.Evaluate(`(() => {
	// 			const selectors = ['div.ds-flex._0a3d93b', '.ds-floating-position-wrapper', '[class*="_0a3d93b"]', '[class*="ds-flex"]'];
	// 			let allElements = [];
	// 			selectors.forEach(selector => {
	// 				try {
	// 					const elements = document.querySelectorAll(selector);
	// 					Array.from(elements).forEach(el => {
	// 						allElements.push({
	// 							selector: selector,
	// 							tagName: el.tagName,
	// 							className: el.className,
	// 							style: el.style.cssText,
	// 							innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
	// 							attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
	// 								acc[attr.name] = attr.value;
	// 								return acc;
	// 							}, {})
	// 						});
	// 					});
	// 				} catch(e) {
	// 					// 忽略选择器错误
	// 				}
	// 			});
	// 			return allElements;
	// 		})()`, &aiCompletionElements).Do(ctx)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// 将AI完成标志元素详情也保存到文件
	// 		os.MkdirAll("temp", 0755) // 确保temp目录存在
	// 		aiCompletionFilename := fmt.Sprintf("temp/ai_completion_elements_%d.json", time.Now().Unix())
	// 		aiCompletionJSON, jsonErr := json.MarshalIndent(aiCompletionElements, "", "  ")
	// 		if jsonErr == nil {
	// 			err = os.WriteFile(aiCompletionFilename, aiCompletionJSON, 0644)
	// 			if err == nil {
	// 				gc.logger.Printf("🔍 Potential AI Completion Elements saved to %s", aiCompletionFilename)
	// 			} else {
	// 				gc.logger.Printf("⚠️ Error saving AI completion elements to file: %v", err)
	// 				gc.logger.Printf("🔍 Potential AI Completion Elements: %+v", aiCompletionElements)
	// 			}
	// 		} else {
	// 			gc.logger.Printf("⚠° Error marshaling AI completion elements to JSON: %v", jsonErr)
	// 			gc.logger.Printf("🔍 Potential AI Completion Elements: %+v", aiCompletionElements)
	// 		}
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	gc.logger.Printf("⚠️ Error getting potential AI completion elements: %v", err)
	// }

	// 在现有AI完成标志元素抓取后添加增强的DOM抓取功能
	// 增加更多DOM抓取方法，特别是针对可能动态加载的元素
	// var allPageElements []map[string]interface{}
	// err = chromedp.Run(ctx,
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		err := chromedp.Evaluate(`(() => {
	// 			// 获取所有可能的按钮相关元素
	// 			const buttonSelectors = [
	// 				'div.ds-icon-button',
	// 				'button',
	// 				'[role="button"]',
	// 				'.ds-flex',
	// 				'[class*="button"]',
	// 				'[class*="icon"]',
	// 				'[class*="action"]',
	// 				'div[tabindex]',
	// 				'[class*="position-wrapper"]',
	// 				'[class*="floating"]'
	// 			];
	// 			let allElements = [];
	// 			buttonSelectors.forEach(selector => {
	// 				try {
	// 					const elements = document.querySelectorAll(selector);
	// 					Array.from(elements).forEach(el => {
	// 						// 获取元素的rect信息以了解其位置
	// 						const rect = el.getBoundingClientRect();
	// 						allElements.push({
	// 							selector: selector,
	// 							tagName: el.tagName,
	// 							className: el.className,
	// 							style: el.style.cssText,
	// 							innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
	// 							attributes: Object.fromEntries(Array.from(el.attributes).map(attr => [attr.name, attr.value])),
	// 							isVisible: !!(rect.width && rect.height),
	// 							isInViewport: rect.top >= 0 && rect.left >= 0 && rect.bottom <= window.innerHeight && rect.right <= window.innerWidth,
	// 							rect: { top: rect.top, left: rect.left, bottom: rect.bottom, right: rect.right, width: rect.width, height: rect.height }
	// 						});
	// 					});
	// 				} catch(e) {
	// 					// 忽略选择器错误
	// 				}
	// 			});
	// 			return allElements;
	// 		})()`, &allPageElements).Do(ctx)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// 将所有页面元素详情也保存到文件
	// 		os.MkdirAll("temp", 0755) // 确保temp目录存在
	// 		allElementsFilename := fmt.Sprintf("temp/all_page_elements_%d.json", time.Now().Unix())
	// 		allElementsJSON, jsonErr := json.MarshalIndent(allPageElements, "", "  ")
	// 		if jsonErr == nil {
	// 			err = os.WriteFile(allElementsFilename, allElementsJSON, 0644)
	// 			if err == nil {
	// 				gc.logger.Printf("🔍 All Page Elements saved to %s", allElementsFilename)
	// 			} else {
	// 				gc.logger.Printf("⚠️ Error saving all page elements to file: %v", err)
	// 				gc.logger.Printf("🔍 All Page Elements count: %d", len(allPageElements))
	// 			}
	// 		} else {
	// 			gc.logger.Printf("⚠️ Error marshaling all page elements to JSON: %v", jsonErr)
	// 			gc.logger.Printf("🔍 All Page Elements count: %d", len(allPageElements))
	// 		}
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	gc.logger.Printf("⚠️ Error getting all page elements: %v", err)
	// }

	// 再次获取完整的DOM树结构
	// var domTree map[string]interface{}
	// err = chromedp.Run(ctx,
	// 	chromedp.ActionFunc(func(ctx context.Context) error {
	// 		err := chromedp.Evaluate(`(function() {
	// 			// 递归获取DOM树结构
	// 			function getDOMTree(node) {
	// 				var result = {
	// 					tagName: node.tagName || '#text',
	// 					className: node.className || '',
	// 					attributes: node.nodeType === 1 ? Array.prototype.reduce.call(node.attributes || [], function(acc, attr) {
	// 						acc[attr.name] = attr.value;
	// 						return acc;
	// 					}, {}) : {},
	// 					textContent: node.nodeType === 3 ? (node.textContent || '').trim().substring(0, 100) : '',
	// 					children: []
	// 				};

	// 				if (node.childNodes) {
	// 					for (var i = 0; i < node.childNodes.length; i++) {
	// 						var child = node.childNodes[i];
	// 						if (child.nodeType === 1 || child.nodeType === 3) { // Element or Text node
	// 							result.children.push(getDOMTree(child));
	// 						}
	// 					}
	// 				}
	// 				return result;
	// 			}

	// 			return getDOMTree(document.documentElement);
	// 		})()`, &domTree).Do(ctx)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// 将DOM树结构保存到文件
	// 		os.MkdirAll("temp", 0755) // 确保temp目录存在
	// 		domTreeFilename := fmt.Sprintf("temp/dom_tree_%d.json", time.Now().Unix())
	// 		domTreeJSON, jsonErr := json.MarshalIndent(domTree, "", "  ")
	// 		if jsonErr == nil {
	// 			err = os.WriteFile(domTreeFilename, domTreeJSON, 0644)
	// 			if err == nil {
	// 				gc.logger.Printf("🔍 Full DOM Tree saved to %s", domTreeFilename)
	// 			} else {
	// 				gc.logger.Printf("⚠️ Error saving DOM tree to file: %v", err)
	// 			}
	// 		} else {
	// 			gc.logger.Printf("⚠️ Error marshaling DOM tree to JSON: %v", jsonErr)
	// 		}
	// 		return nil
	// 	}),
	// )
	// if err != nil {
	// 	gc.logger.Printf("⚠️ Error getting DOM tree: %v", err)
	// }

	// 检查是否已经通过优化方法成功提取了AI内容
	if aiContentExtractedSuccessfully {
		gc.logger.Println("✅ AI content already extracted via optimized method, skipping additional waiting loop")
	} else {
		// 仅当内容未通过优化方法提取时才进入额外等待循环
		// 延长窗口存活时间，确保AI有足够时间完成输出
		gc.logger.Printf("⏳ Waiting for AI response with selectors: %v, timeout: %d seconds", responseSelectors, GUARDIAN_BROWSER_TIMEOUT_SECONDS)

		// 等待响应出现
		waitCtx, cancel := context.WithTimeout(ctx, time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS)*time.Second) // 使用配置的超时时间
		defer cancel()

	Loop:
		for {
			select {
			case <-waitCtx.Done():
				gc.logger.Println("⏰ Timeout waiting for AI response")
				break Loop
			default:
				// 首先尝试获取响应
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
					gc.logger.Println("✅ Response received")

					// 检测AI是否完成输出
					completionCtx, completionCancel := context.WithTimeout(ctx, 10*time.Second) // 只等待10秒检测完成状态
					defer completionCancel()

					for {
						select {
						case <-completionCtx.Done():
							gc.logger.Println("⏰ Finished checking AI completion status")
							break Loop
						default:
							// 检测AI是否完成输出的多种方法 - 使用新的检测机制
							// 检查特定关键词是否出现在页面中
							keywordFound := false
							err := chromedp.EvaluateAsDevTools(
								`(function() {
									var html = document.documentElement.outerHTML;
									// 检测AI输出完成 - 仅使用ds-flex _0a3d93b作为唯一条件
									return html.indexOf('ds-flex _0a3d93b') !== -1;
								})();`, &keywordFound).Do(ctx)

							if err == nil && keywordFound {
								gc.logger.Println("✅ Specific keyword found, indicating AI processing completed")
								break Loop
							}

							time.Sleep(1 * time.Second) // 等待一秒后再次检查
						}
					}
				}
				// 继续循环等待
				time.Sleep(1 * time.Second)
			}
		}
	}

	// 确保响应内容已经被清理
	if response != "" {
		response = gc.cleanHTMLContent(response)
		gc.logger.Printf("✅ Final response content cleaned, length: %d", len(response))
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
	} else {
		// 如果已经有响应内容，确保它经过了HTML清理
		response = gc.cleanHTMLContent(response)
		gc.logger.Printf("✅ Cleaned response content after extraction, length: %d", len(response))
	}

	gc.logger.Printf("✅ Successfully retrieved AI response, length: %d", len(response))

	// 计算总耗时
	duration := time.Since(startTime)
	gc.logger.Printf("⏱️ Browser automation completed in %v", duration)

	// 如果有提取到的内容，打印内容摘要到日志
	if response != "" {
		contentLength := len(response)
		var preview string
		if contentLength <= 40 { // 如果内容少于等于40个字符，直接显示
			preview = response
		} else {
			// 显示前20个字符 + " ... " + 后20个字符
			startPart := response[:20]
			endPart := response[contentLength-20:]
			preview = fmt.Sprintf("%s ... %s", startPart, endPart)
		}

		// 对于超大文本，额外显示更多信息
		if contentLength > 10000 {
			gc.logger.Printf("📢 Large content detected (length: %d) - Preview: %s", contentLength, preview)
		} else {
			gc.logger.Printf("📋 Extracted content preview (length: %d): %s", contentLength, preview)
		}
	}

	// 根据配置决定是否保持浏览器打开
	// 由于Config结构体中没有KeepAlive字段，暂时移除该条件
	// gc.logger.Printf("😴 Keeping browser alive for %d seconds as configured", GUARDIAN_AUTO_KEEP_OPEN_SECONDS)
	// time.Sleep(time.Duration(GUARDIAN_AUTO_KEEP_OPEN_SECONDS) * time.Second)

	// 根据配置决定是否保持浏览器打开
	// 现在正常关闭浏览器上下文以释放资源
	gc.logger.Printf("🔚 Browser automation completed, closing browser context")

	// 如果有提取到的内容，打印内容摘要到日志（在关闭浏览器前）
	if response != "" {
		contentLength := len(response)
		var preview string
		if contentLength <= 40 { // 如果内容少于等于40个字符，直接显示
			preview = response
		} else {
			// 显示前20个字符 + " ... " + 后20个字符
			var startPart, endPart string
			if len(response) >= 20 {
				startPart = response[:20]
				endPart = response[contentLength-20:]
			} else {
				startPart = response
				endPart = response
			}
			preview = fmt.Sprintf("%s ... %s", startPart, endPart)
		}

		// 对于超大文本，额外显示更多信息
		if contentLength > 10000 {
			gc.logger.Printf("📢 Large content detected (length: %d) - Preview: %s", contentLength, preview)
		} else {
			gc.logger.Printf("📋 Extracted content preview (length: %d): %s", contentLength, preview)
		}
	}

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

// 实现 AIClient 接口

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

// SetTimeout 设置超时时间
func (gc *GuardianClient) SetTimeout(timeout time.Duration) {
	gc.ProviderConfig.Timeout = timeout
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

// String returns string representation of the client
func (gc *GuardianClient) String() string {
	return fmt.Sprintf("[Provider: %s, Model: %s]", gc.ProviderConfig.Provider, gc.ProviderConfig.Model)
}

func (gc *GuardianClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// 对于浏览器自动化，不需要构建HTTP请求体
	return nil
}

func (gc *GuardianClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	// 对于浏览器自动化，不需要构建HTTP请求体
	return nil
}

func (gc *GuardianClient) buildUrl() string {
	// 返回配置的基本URL
	return gc.ProviderConfig.BaseURL
}

func (gc *GuardianClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	// 对于浏览器自动化，我们不使用标准HTTP请求
	return nil, fmt.Errorf("GuardianClient does not use standard HTTP requests")
}

func (gc *GuardianClient) setAuthHeader(reqHeaders http.Header) {
	// 对于浏览器自动化，认证通过浏览器会话处理
}

func (gc *GuardianClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	// 对于浏览器自动化，不需要序列化请求体
	return nil, nil
}

func (gc *GuardianClient) parseMCPResponse(body []byte) (string, error) {
	// 对于浏览器自动化，响应由浏览器自动化流程处理
	return "", nil
}

func (gc *GuardianClient) isRetryableError(err error) bool {
	// 对于浏览器自动化，所有错误都可以重试
	return true
}

// SetDynamicConfig 动态设置配置
func (gc *GuardianClient) SetDynamicConfig(baseURL string) {
	gc.BaseURL = baseURL
	gc.ProviderConfig.BaseURL = baseURL
}

// 在提交后等待指定时间并轮询检测AI输出完成信号，并提取AI输出内容
func (gc *GuardianClient) waitForAndValidateOutput(ctx context.Context) (string, error) {
	// 从环境变量获取配置值
	initialWaitSeconds := getEnvInt("GUARDIAN_INITIAL_WAIT_SECONDS", 60)
	minPollSeconds := getEnvInt("GUARDIAN_MIN_POLL_SECONDS", 2)
	maxPollSeconds := getEnvInt("GUARDIAN_MAX_POLL_SECONDS", 5)

	// 前 initialWaitSeconds 秒等待AI输入完成，不做检测
	gc.logger.Printf("⏳ Waiting %d seconds for AI to start output...", initialWaitSeconds)
	time.Sleep(time.Duration(initialWaitSeconds) * time.Second)

	gc.logger.Println("🔍 Starting to poll for AI completion signal...")

	// 开始轮询检测，直到找到 ds-flex _0a3d93b 字符串
	for {
		select {
		case <-ctx.Done():
			gc.logger.Println("⏰ Context cancelled, stopping polling for AI completion")
			return "", ctx.Err()
		default:
			// 获取当前页面的HTML内容
			var pageHTML string
			err := chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
					return err
				}),
			)
			if err != nil {
				gc.logger.Printf("⚠️ Error getting page HTML: %v", err)
				// 继续尝试，不中断轮询
			} else {
				// 检查实时获取的HTML是否包含 ds-flex _0a3d93b 标识
				if gc.checkAICompletionInHTML(pageHTML) {
					gc.logger.Println("✅ AI completion detected in live page: 'ds-flex _0a3d93b' found")
					// 提取AI输出内容
					extractedContent, extractErr := gc.extractAIOutputContent(ctx)
					if extractErr != nil {
						gc.logger.Printf("⚠️ Error extracting AI output content: %v", extractErr)
						// 如果提取失败，仍然返回成功，但内容为空
						return "", nil
					}
					gc.logger.Printf("✅ Successfully extracted AI output content, length: %d", len(extractedContent))

					// 清理HTML内容，移除多余的HTML标签
					cleanedContent := gc.cleanHTMLContent(extractedContent)
					gc.logger.Printf("✅ Cleaned AI output content, length: %d", len(cleanedContent))

					return cleanedContent, nil // 找到关键字并提取内容，返回成功
				} else {
					// 添加调试日志，显示当前HTML的一些信息
					gc.logger.Printf("🔍 Still waiting... Live HTML length: %d characters", len(pageHTML))
					// 查找页面中可能的相关元素
					if strings.Contains(pageHTML, "ds-flex") {
						gc.logger.Println("💡 Found 'ds-flex' in live page, checking for completion markers...")
					}
					if strings.Contains(pageHTML, "_0a3d93b") {
						gc.logger.Println("💡 Found '_0a3d93b' in live page, checking for completion markers...")
					}
				}
			}

			// 同时尝试从内存中获取HTML内容进行检测
			// 之前已经通过chromedp获取了页面HTML内容，直接使用内存中的内容进行检测
			// 获取当前页面的HTML内容用于检测
			var currentPageHTML string
			htmlErr := chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					err := chromedp.OuterHTML("html", &currentPageHTML).Do(ctx)
					return err
				}),
			)
			if htmlErr == nil && currentPageHTML != "" {
				// 检测内存中的HTML内容
				if gc.checkAICompletionInHTML(currentPageHTML) {
					gc.logger.Println("✅ AI completion detected in current page HTML: 'ds-flex _0a3d93b' found")
					// 提取AI输出内容
					extractedContent, extractErr := gc.extractAIOutputContent(ctx)
					if extractErr != nil {
						gc.logger.Printf("⚠️ Error extracting AI output content: %v", extractErr)
						// 如果提取失败，仍然返回成功，但内容为空
						return "", nil
					}
					gc.logger.Printf("✅ Successfully extracted AI output content from current page, length: %d", len(extractedContent))

					// 清理HTML内容，移除多余的HTML标签
					cleanedContent := gc.cleanHTMLContent(extractedContent)
					gc.logger.Printf("✅ Cleaned AI output content, length: %d", len(cleanedContent))

					return cleanedContent, nil // 找到关键字并提取内容，返回成功
				}
			}

			// 随机等待 minPollSeconds 到 maxPollSeconds 秒
			randWait := time.Duration(minPollSeconds+rand.Intn(maxPollSeconds-minPollSeconds+1)) * time.Second
			gc.logger.Printf("⏳ Polling wait: checking again in %.0f seconds", randWait.Seconds())
			time.Sleep(randWait)
		}
	}
}

// 从临时文件中检测AI输出完成标识
func (gc *GuardianClient) detectAICompletionFromFile(filePath string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	pageHTML := string(content)

	// 检查是否包含 ds-flex _0a3d93b 标识 - 这是唯一的AI完成检测条件
	dsFlexPattern1 := strings.Contains(pageHTML, "ds-flex _0a3d93b")                                             // 精确匹配
	dsFlexPattern2 := strings.Contains(pageHTML, "ds-flex") && strings.Contains(pageHTML, "_0a3d93b")            // 分开匹配
	dsFlexPattern3 := regexp.MustCompile(`ds-flex[.\s\S]*?_0a3d93b|ds-flex[^>]*?_0a3d93b`).MatchString(pageHTML) // 正则匹配

	result := dsFlexPattern1 || dsFlexPattern2 || dsFlexPattern3
	if result {
		gc.logger.Println("✅ AI completion detected from file: 'ds-flex _0a3d93b' found in file content")
	} else {
		gc.logger.Printf("❌ AI completion NOT detected from file: 'ds-flex _0a3d93b' NOT found in file content (file: %s)", filePath)
	}

	return result, nil
}

// 检查HTML内容中是否包含AI完成标识
func (gc *GuardianClient) checkAICompletionInHTML(pageHTML string) bool {
	// 检查是否包含 ds-flex _0a3d93b 标识 - 这是唯一的AI完成检测条件
	// 使用多种方式检测，因为页面结构可能有细微变化
	dsFlexPattern1 := strings.Contains(pageHTML, "ds-flex _0a3d93b")                                             // 精确匹配
	dsFlexPattern2 := strings.Contains(pageHTML, "ds-flex") && strings.Contains(pageHTML, "_0a3d93b")            // 分开匹配
	dsFlexPattern3 := regexp.MustCompile(`ds-flex[.\s\S]*?_0a3d93b|ds-flex[^>]*?_0a3d93b`).MatchString(pageHTML) // 正则匹配

	return dsFlexPattern1 || dsFlexPattern2 || dsFlexPattern3
}

// 从包含 ds-flex _0a3d93b 的HTML容器中提取AI输出内容
func (gc *GuardianClient) extractAIOutputContent(ctx context.Context) (string, error) {
	var aiContent string

	// 尝试使用选择器来定位AI输出内容 - 仅使用ds-flex _0a3d93b相关选择器
	selectors := []string{
		"div.ds-flex._0a3d93b",      // 包含 ds-flex _0a3d93b 的容器
		"div.ds-flex._0a3d93b div",  // 容器内的内容
		"div.ds-flex._0a3d93b span", // 容器内的文本节点
		"div.ds-flex._0a3d93b *",    // 容器内的任意元素
	}

	for _, selector := range selectors {
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 尝试获取元素的文本内容
				var content string
				err := chromedp.Text(selector, &content).Do(ctx)
				if err == nil && strings.TrimSpace(content) != "" {
					aiContent = strings.TrimSpace(content)
					gc.logger.Printf("✅ Extracted content from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果Text获取失败，尝试获取innerHTML
				var innerHTML string
				err = chromedp.InnerHTML(selector, &innerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(innerHTML) != "" {
					aiContent = strings.TrimSpace(innerHTML)
					gc.logger.Printf("✅ Extracted innerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果上述方法都失败，尝试OuterHTML
				var outerHTML string
				err = chromedp.OuterHTML(selector, &outerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(outerHTML) != "" {
					aiContent = strings.TrimSpace(outerHTML)
					gc.logger.Printf("✅ Extracted outerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				return fmt.Errorf("no content found with selector: %s", selector)
			}),
		)

		if err == nil && aiContent != "" {
			return aiContent, nil
		}
	}

	// 如果常规选择器都失败，尝试使用JavaScript直接查找AI输出元素及其内容
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			// 仅使用ds-flex _0a3d93b相关的选择器来查找AI输出内容
			const selectors = [
				'div.ds-flex._0a3d93b',           // 主要标识
				'div.ds-flex._0a3d93b div',       // 容器内的内容
				'div.ds-flex._0a3d93b span',      // 容器内的文本节点
				'div.ds-flex._0a3d93b *'          // 容器内的任意元素
			];
			
			for (let selector of selectors) {
				const elements = document.querySelectorAll(selector);
				if (elements.length > 0) {
					// 获取第一个匹配元素的内容
					const element = elements[0];
					// 首先尝试获取文本内容
					if (element.textContent && element.textContent.trim()) {
						return element.textContent.trim();
					}
					// 如果文本内容为空，返回内部HTML
					return element.innerHTML || element.outerHTML || '';
				}
			}
			return '';
		})()`, &aiContent),
	)

	if err != nil {
		return "", err
	}

	if aiContent != "" {
		gc.logger.Printf("✅ Extracted content using JavaScript evaluation, length: %d", len(aiContent))
		return aiContent, nil
	}

	return "", fmt.Errorf("no AI output content found in ds-flex _0a3d93b container")
}

// 清理HTML内容，移除多余的HTML标签，保留文本内容
func (gc *GuardianClient) cleanHTMLContent(htmlContent string) string {
	// 首先尝试使用正则表达式移除HTML标签
	cleaned := htmlContent

	// 移除HTML标签（保留标签间的内容）
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	// 替换多个空白字符为单个空格
	space := regexp.MustCompile(`\s+`)
	cleaned = space.ReplaceAllString(cleaned, " ")

	// 移除常见的转义字符
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")

	return strings.TrimSpace(cleaned)
}

const (
	GUARDIAN_BROWSER_TIMEOUT_SECONDS        = 990 // 默认 4分50秒，确保AI有足够时间完成输出 主要浏览器自动化操作的总超时时间
	GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS = 999 //手动浏览器操作模式下的超时时间 支持人工干预或手动登录场景
	GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS   = 999 //长时间浏览器操作的超时时间
	GUARDIAN_AUTO_KEEP_OPEN_SECONDS         = 999 // 改为999秒，确保窗口长时间存活 浏览器自动化完成后保持窗口打开的时间
)

// findSafeUTF8Boundary 找到安全的UTF-8字符边界
func findSafeUTF8Boundary(text string, start, proposedEnd int) int {
	if proposedEnd >= len(text) || start >= len(text) || start >= proposedEnd {
		return proposedEnd
	}

	// 从提议的结束位置向前查找，直到找到有效的UTF-8字符边界
	// UTF-8字节模式：
	// 0xxxxxxx - ASCII字符 (0x00-0x7F)
	// 10xxxxxx - 多字节字符的延续字节 (0x80-0xBF)
	// 11xxxxxx - 多字节字符的开始字节 (0xC0-0xFF)
	for proposedEnd > start {
		// 获取该位置的字节
		b := text[proposedEnd]
		// 检查是否为有效的UTF-8字符开始 (不是延续字节)
		if b < 0x80 || b >= 0xC0 { // ASCII (0xxxxxxx) 或 多字节开始 (11xxxxxx)
			return proposedEnd
		}
		// 如果是延续字节 (10xxxxxx)，继续向前
		proposedEnd--
	}

	// 如果找不到合适的位置，返回起始位置
	return start
}
