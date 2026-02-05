package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

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

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel() // 取消浏览器分配器

	// 创建chrome实例上下文
	ctx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel() // 取消浏览器实例

	// 设置超时 - 使用强制关闭时间作为主超时，确保浏览器不会运行超过4分50秒
	forceCloseTimeoutSeconds := getEnvInt("GUARDIAN_FORCE_CLOSE_TIMEOUT_SECONDS", 290)
	timeout := time.Duration(forceCloseTimeoutSeconds) * time.Second
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel() // 取消带超时的上下文

	// 记录实际使用的超时时间
	gc.logger.Printf("⏰ Browser automation started with force close timeout: %v", timeout)

	// 记录开始时间
	startTime := time.Now()
	gc.logger.Printf("⏰ Browser automation started, timeout: %v", timeout)

	// 启动强制关闭定时器：4分50秒后强制关闭浏览器，防止影响下一轮查询
	forceCloseTimer := time.AfterFunc(time.Duration(forceCloseTimeoutSeconds)*time.Second, func() {
		gc.logger.Printf("🚨 Force closing browser after %d seconds (safety timeout)", forceCloseTimeoutSeconds)
		// 注意：这里我们不能直接调用cancel函数，因为我们需要通过其他方式通知函数终止
		// 我们将在后续检查中检测这个超时
	})
	defer forceCloseTimer.Stop() // 确保在函数正常结束时停止定时器

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
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
			"pre",
			"code",
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
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
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

						// 首先点击输入框以获得焦点
						err = chromedp.Click(selector).Do(ctx)
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
											var inputEvent = new Event('input', { bubbles: true, cancelable: true });
											element.dispatchEvent(inputEvent);
											var changeEvent = new Event('change', { bubbles: true, cancelable: true });
											element.dispatchEvent(changeEvent);
											return true;
									}
										return false;
									} catch(e) {
											console.error('Error in state update script:', e);
											return false;
									}
									})();
							`, selector)

						var ensureResult bool
						err = chromedp.Evaluate(ensureStateUpdateScript, &ensureResult).Do(ctx)
						if err != nil || !ensureResult {
							gc.logger.Printf("⚠️ State update script failed: %v, success: %v", err, ensureResult)
						} else {
							gc.logger.Printf("✅ State update script executed successfully")
						}

						// 短暂延迟，确保页面响应
						time.Sleep(200 * time.Millisecond)

						// 标记输入成功
						inputFound = true

						// 输入完成，跳出选择器循环，继续执行提交按钮逻辑
						break // 跳出选择器循环，继续执行提交按钮逻辑
					}

					// 短暂延迟，确保页面响应
					time.Sleep(200 * time.Millisecond)

					// 标记输入成功
					inputFound = true

					// 输入完成，跳出选择器循环，继续执行提交按钮逻辑
					break // 跳出选择器循环，继续执行提交按钮逻辑
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

	// 使用新的chromedp操作来等待输入完成并点击提交按钮
	waitErr := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			maxWait := time.Now().Add(10 * time.Second) // 最大等待10秒
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

						// 如果输入框内容长度与处理后的提示词长度一致，可以点击提交按钮
						if inputLength == len(safePrompt) {
							gc.logger.Printf("✅ Input length matches processed prompt length (%d), attempting to click submit button", len(safePrompt))

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

				time.Sleep(1 * time.Second) // 等待1秒后再次检查
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

	// 使用新的等待和验证机制检测AI输出完成
	gc.logger.Println("⏳ Waiting for AI to complete output using optimized detection mechanism...")
	// 使用较短的超时时间，确保不会超过强制关闭时间
	validationCtx, validationCancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second)
	extractedContent, err := gc.waitForAndValidateOutput(validationCtx)
	validationCancel() // 确保取消上下文以释放资源
	if err != nil {
		gc.logger.Printf("⚠️ Error waiting for AI completion: %v", err)
		// 继续执行，即使等待出现问题
	} else {
		gc.logger.Println("✅ AI output completed successfully detected")
		// 如果成功提取了内容，使用提取的内容作为响应
		if extractedContent != "" {
			response = extractedContent
			gc.logger.Printf("✅ Using extracted content as response, length: %d", len(response))
		}

		// 在AI输出完成后关闭浏览器
		gc.logger.Println("✅ Closing browser after AI output completion")

		// 显式取消浏览器上下文以确保浏览器被关闭
		// browserCancel应该通过defer语句自动调用，但为了确保，这里记录
		gc.logger.Println("💡 Browser context will be cancelled via defer statement when function exits")
	}

	// 获取完整的页面HTML内容但不保存到文件
	var pageHTML string
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 获取完整的页面HTML内容
			err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
			if err != nil {
				return err
			}
			// 仅在内存中使用HTML内容，不保存到文件
			gc.logger.Printf("📄 Full DOM snapshot acquired, size: %d bytes", len(pageHTML))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Error getting page HTML: %v", err)
	}

	// 启用：增强的DOM抓取和调试功能
	// 获取页面上所有ds-theme元素的详细信息（不保存到文件）
	var themeElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				const elements = document.querySelectorAll('div.ds-theme');
				return Array.from(elements).map(el => ({
					tagName: el.tagName,
					className: el.className,
					style: el.style.cssText,
					innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
					attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
						acc[attr.name] = attr.value;
						return acc;
					}, {})
				}));
			})()`, &themeElements).Do(ctx)
			if err != nil {
				return err
			}

			// 仅在内存中使用主题元素信息，不保存到文件
			gc.logger.Printf("🔍 ds-theme Elements Detail acquired, count: %d", len(themeElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Error getting ds-theme elements detail: %v", err)
	}

	// 也获取页面上所有可能的AI完成标志元素（不保存到文件）
	var aiCompletionElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				const selectors = ['div.ds-flex._0a3d93b', '.ds-floating-position-wrapper', '[class*="_0a3d93b"]', '[class*="ds-flex"]'];
				let allElements = [];
				selectors.forEach(selector => {
					try {
						const elements = document.querySelectorAll(selector);
						Array.from(elements).forEach(el => {
							allElements.push({
								selector: selector,
								tagName: el.tagName,
								className: el.className,
								style: el.style.cssText,
								innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
								attributes: Array.prototype.reduce.call(el.attributes, function(acc, attr) {
									acc[attr.name] = attr.value;
									return acc;
								}, {})
							});
						});
					} catch(e) {
						// 忽略选择器错误
					}
				});
				return allElements;
			})()`, &aiCompletionElements).Do(ctx)
			if err != nil {
				return err
			}

			// 仅在内存中使用AI完成标志元素信息，不保存到文件
			gc.logger.Printf("🔍 Potential AI Completion Elements acquired, count: %d", len(aiCompletionElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Error getting potential AI completion elements: %v", err)
	}

	// 在现有AI完成标志元素抓取后添加增强的DOM抓取功能
	// 增加更多DOM抓取方法，特别是针对可能动态加载的元素（不保存到文件）
	var allPageElements []map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(() => {
				// 获取所有可能的按钮相关元素
				const buttonSelectors = [
					'div.ds-icon-button',
					'button',
					'[role="button"]',
					'.ds-flex',
					'[class*="button"]',
					'[class*="icon"]',
					'[class*="action"]',
					'div[tabindex]',
					'[class*="position-wrapper"]',
					'[class*="floating"]'
				];
				let allElements = [];
				buttonSelectors.forEach(selector => {
					try {
						const elements = document.querySelectorAll(selector);
						Array.from(elements).forEach(el => {
							// 获取元素的rect信息以了解其位置
							const rect = el.getBoundingClientRect();
							allElements.push({
								selector: selector,
								tagName: el.tagName,
								className: el.className,
								style: el.style.cssText,
								innerHTML: el.innerHTML ? el.innerHTML.substring(0, 500) : '',
								attributes: Object.fromEntries(Array.from(el.attributes).map(attr => [attr.name, attr.value])),
								isVisible: !!(rect.width && rect.height),
								isInViewport: rect.top >= 0 && rect.left >= 0 && rect.bottom <= window.innerHeight && rect.right <= window.innerWidth,
								rect: { top: rect.top, left: rect.left, bottom: rect.bottom, right: rect.right, width: rect.width, height: rect.height }
							});
						});
					} catch(e) {
						// 忽略选择器错误
					}
				});
				return allElements;
			})()`, &allPageElements).Do(ctx)
			if err != nil {
				return err
			}

			// 仅在内存中使用所有页面元素信息，不保存到文件
			gc.logger.Printf("🔍 All Page Elements acquired, count: %d", len(allPageElements))
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Error getting all page elements: %v", err)
	}

	// 再次获取完整的DOM树结构
	var domTree map[string]interface{}
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`(function() {
				// 递归获取DOM树结构
				function getDOMTree(node) {
					var result = {
						tagName: node.tagName || '#text',
						className: node.className || '',
						attributes: node.nodeType === 1 ? Array.prototype.reduce.call(node.attributes || [], function(acc, attr) {
							acc[attr.name] = attr.value;
							return acc;
						}, {}) : {},
						textContent: node.nodeType === 3 ? (node.textContent || '').trim().substring(0, 100) : '',
						children: []
					};

					if (node.childNodes) {
						for (var i = 0; i < node.childNodes.length; i++) {
							var child = node.childNodes[i];
							if (child.nodeType === 1 || child.nodeType === 3) { // Element or Text node
								result.children.push(getDOMTree(child));
							}
						}
					}
					return result;
				}

				return getDOMTree(document.documentElement);
			})()`, &domTree).Do(ctx)
			if err != nil {
				return err
			}

			// 仅在内存中使用DOM树信息，不保存到文件
			gc.logger.Printf("🔍 Full DOM Tree acquired")
			return nil
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ Error getting DOM tree: %v", err)
	}

	// 检查是否已经检测到AI输出完成，如果是，则跳过主等待循环
	if extractedContent != "" {
		gc.logger.Println("⏭️ Skipping main wait loop as AI output has already been detected and captured")
		// 直接跳到函数末尾，避免进入主等待循环
		// 不再等待主循环，直接执行清理并返回
		gc.logger.Println("✅ AI processing completed, closing browser immediately")
		return response, nil
	}

	// 延长窗口存活时间，确保AI有足够时间完成输出
	gc.logger.Printf("⏳ Keeping browser window alive for configured timeout: %d seconds", forceCloseTimeoutSeconds)
	// 不需要额外的sleep，因为整体超时已经在配置中设置

	// 使用配置的超时时间等待AI处理并获取响应
	gc.logger.Printf("⏳ Waiting for AI response with selectors: %v, timeout: %d seconds", responseSelectors, forceCloseTimeoutSeconds)

	// 等待响应出现
	waitCtx, cancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second) // 使用强制关闭超时时间
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
				// 成功获取到响应，现在等待AI处理完成的标志：提交按钮变为可用状态 或 复制按钮出现
				gc.logger.Println("✅ Response received, waiting for AI processing to complete (checking submit button or copy button)...")

				// 检查特定关键词是否出现（AI完成的标志）
				// 检测是否出现<div class="ds-flex _0a3d93b" style="align-items: center; gap: 10px;"><div class="ds-flex
				//这是一个匹配是否完成输出的判断
				/*
					completionCtx, cancel := context.WithTimeout(ctx, time.Duration(GUARDIAN_BROWSER_TIMEOUT_SECONDS)*time.Second)
					defer cancel()

						for {
							select {
							case <-completionCtx.Done():
								gc.logger.Println("⏰ Timeout waiting for AI processing to complete")
								break Loop // 即使没有明确完成标志，我们也已有响应，所以退出主循环
							default:
								// 检查特定关键词是否出现在页面中
								keywordFound := false
								err = chromedp.EvaluateAsDevTools(
									`(function() {
											var html = document.documentElement.outerHTML;
											return html.indexOf('div class=\"ds-flex _0a3d93b\" style=\"align-items: center; gap: 10px;\"') !== -1 &&
												html.indexOf('<div class=\"ds-flex') !== -1;
										})();`, &keywordFound).Do(ctx)

								if err == nil && keywordFound {
									gc.logger.Println("✅ Specific keyword found, indicating AI processing completed")
									break Loop
								}

								time.Sleep(1 * time.Second) // 等待一秒后再次检查
							}
						}
				*/
				// 使用多种策略检测AI是否完成输出
				completionCtx, cancel := context.WithTimeout(ctx, time.Duration(forceCloseTimeoutSeconds)*time.Second)
				defer cancel()

				for {
					select {
					case <-completionCtx.Done():
						gc.logger.Println("⏰ Timeout waiting for AI processing to complete")
						break Loop // 即使没有明确完成标志，我们也已有响应，所以退出主循环
					default:
						// 检测AI是否完成输出的多种方法 - 使用新的检测机制
						// 使用新增的AI完成检测方法
						err := gc.ProcessAIOutputCompletion(ctx)
						if err == nil {
							gc.logger.Println("✅ AI processing completed using enhanced detection")
							break Loop
						} else {
							gc.logger.Printf("⏳ AI processing still in progress: %v", err)
						}

						time.Sleep(1 * time.Second) // 等待一秒后再次检查
					}
				}

				// 暂时注释掉所有AI完成检测条件，以便重新寻找有效的检测条件
				/*
					for {
						select {
						case <-completionCtx.Done():
							gc.logger.Println("⏰ Timeout waiting for AI processing to complete")
							break Loop // 即使没有明确完成标志，我们也已有响应，所以退出主循环
						default:
							// 检查提交按钮是否变为向上箭头且禁用状态（AI完成的标志之一）
							// 根据您提供的信息，AI正在输出时按钮是方形图标+可点击（aria-disabled="false"）
							// AI输出完毕后按钮变成向上箭头+禁用（aria-disabled="true")
							buttonChanged := false
							for _, submitSel := range submitSelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element) {
												// 检查按钮是否变为禁用状态（aria-disabled="true"），这表明AI已完成
												var isAriaDisabled = element.hasAttribute('aria-disabled') && element.getAttribute('aria-disabled') === 'true';
												return isAriaDisabled;
											}
											return false; // 元素不存在认为按钮不可用
										})();`, submitSel), &buttonChanged).Do(ctx)

								if err == nil && buttonChanged {
									gc.logger.Println("⚠️⚠️ WARNING: 已经检测到输出完成，准备复制 ⚠️⚠️")
									break Loop
								}
							}

							// 同时检查复制按钮是否出现（AI完成的另一个标志）
							copyButtonExists := false
							copySelectors := []string{
								"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child", // 第一个按钮 - 复制
								"div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']",                                  // DeepSeek复制按钮 - 启用状态
								"div.db183363.ds-icon-button[role='button'][aria-disabled='false']",                                                                                     // DeepSeek复制按钮 - 启用状态（备选）
							}

							for _, copySelector := range copySelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element !== null && element.hasAttribute('aria-disabled') && element.getAttribute('aria-disabled') === 'false') {
												// 检查按钮是否可见且在视口中
												var rect = element.getBoundingClientRect();
												var isVisible = rect.top >= 0 && rect.left >= 0 &&
																rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) &&
																rect.right <= (window.innerWidth || document.documentElement.clientWidth);

												// 检查按钮是否有尺寸（不为0）
												var hasDimensions = rect.width > 0 && rect.height > 0;

												return isVisible && hasDimensions;
											}
											return false;
										})();`, copySelector), &copyButtonExists).Do(ctx)

								if err == nil && copyButtonExists {
									gc.logger.Println("✅ Copy button detected as enabled and visible, indicating AI processing is complete")

									// 立即尝试点击复制按钮 - 优先点击第一个（复制按钮）
									gc.logger.Println("📋 Immediately attempting to click copy button...")

									// 使用更精确的选择器来点击第一个复制按钮
									firstCopySelector := "div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child"
									clickErr := chromedp.Click(firstCopySelector).Do(ctx)
									if clickErr == nil {
										gc.logger.Printf("✅ Successfully clicked first copy button: %s", firstCopySelector)
									} else {
										gc.logger.Printf("⚠️ Failed to click first copy button %s: %v, trying original selector", firstCopySelector, clickErr)
										// 如果第一个选择器失败，尝试原来的选择器
										clickErr2 := chromedp.Click(copySelector).Do(ctx)
										if clickErr2 == nil {
											gc.logger.Printf("✅ Successfully clicked copy button with original selector: %s", copySelector)
										} else {
											gc.logger.Printf("⚠️ Failed to click copy button %s: %v", copySelector, clickErr2)
										}
									}

									break Loop
								}
							}

							// 检查特定完成文本是否出现（最高优先级）
							completionTextExists := false
							completionTextSelectors := []string{
								"div.dbe8cf4a", // AI生成完成标记文本
							}

							for _, textSelector := range completionTextSelectors {
								err = chromedp.EvaluateAsDevTools(
									fmt.Sprintf(
										`(function() {
											var element = document.querySelector('%s');
											if (element) {
												var text = element.textContent || element.innerText;
												return text && text.includes('本回答由 AI 生成');
											}
											return false;
										})();`, textSelector), &completionTextExists).Do(ctx)

								if err == nil && completionTextExists {
									gc.logger.Println("✅ AI completion text detected: '本回答由 AI 生成'")

									// 随机等待1-2秒
									randomWait := time.Duration(1000+rand.Intn(1000)) * time.Millisecond
									gc.logger.Printf("⏳ Waiting %.1f seconds before copying...", randomWait.Seconds())
									time.Sleep(randomWait)

									// 立即尝试点击复制按钮
									gc.logger.Println("📋 Immediately attempting to click copy button after detecting completion text...")

									// 查找并点击可用的复制按钮 - 使用您提供的实际按钮定位点，优先点击第一个（复制按钮）
									copySelectors := []string{
										"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:first-child",  // 第一个按钮 - 复制
										"div.ds-flex._0a3d93b div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']:nth-child(1)", // 第一个按钮 - 复制
										"div.ds-icon-button__hover-bg:first-child", // 第一个按钮的实际位置
										"div.ds-icon-button__hover-bg",             // 复制按钮的实际位置
										"div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container[role='button'][aria-disabled='false']",
										"div.db183363.ds-icon-button[role='button'][aria-disabled='false']",
									}

									clicked := false
									for _, copySelector := range copySelectors {
										// 首先等待按钮可见
										err := chromedp.WaitVisible(copySelector).Do(ctx)
										if err == nil {
											clickErr := chromedp.Click(copySelector).Do(ctx)
											if clickErr == nil {
												gc.logger.Printf("✅ Successfully clicked copy button after detecting completion text: %s", copySelector)
												clicked = true
												break
											} else {
												gc.logger.Printf("⚠️ Failed to click copy button %s: %v", copySelector, clickErr)
											}
										} else {
											gc.logger.Printf("⚠️ Copy button %s not visible: %v", copySelector, err)
										}
									}

									if !clicked {
										gc.logger.Println("⚠️ Failed to click any copy button after detecting completion text")
									}

									break Loop
								}
							}

							time.Sleep(1 * time.Second) // 等待一秒后再次检查
						}
					}
				*/

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

	// 清理最终响应内容
	cleanedResponse := gc.cleanAndDecodeContent(response)
	gc.logger.Printf("🧹 Final response content cleaned, length: %d", len(cleanedResponse))

	// 根据配置决定是否保持浏览器打开
	// 由于Config结构体中没有KeepAlive字段，暂时移除该条件
	// gc.logger.Printf("😴 Keeping browser alive for %d seconds as configured", GUARDIAN_AUTO_KEEP_OPEN_SECONDS)
	// time.Sleep(time.Duration(GUARDIAN_AUTO_KEEP_OPEN_SECONDS) * time.Second)

	// 在AI处理完成后立即关闭浏览器，不再等待长时间延迟
	gc.logger.Println("✅ AI processing completed, closing browser immediately")

	// 不再等待配置的延迟时间，直接返回
	return cleanedResponse, nil
}

// OpenLongLivedBrowser 打开长生命周期的浏览器窗口，用于首次登录设置
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
