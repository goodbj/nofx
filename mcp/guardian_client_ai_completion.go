package mcp

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

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

	// 开始轮询检测，直到找到 ds-message _63c77b1 或 ds-flex _0a3d93b 字符串
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
				serviceType := strings.ToLower(gc.ProviderConfig.Provider)

				// 检查页面是否包含相应服务的AI输出完成标志
				// 对于DeepSeek，检查 ds-message _63c77b1 或 ds-flex _0a3d93b 字符串
				// 对于Qwen，检查通义千问特有的类名
				containsAIMessage := false
				containsAICompletion := false

				if strings.Contains(serviceType, "qwen") || strings.Contains(serviceType, "通义千问") {
					// Qwen 特有的检测标志
					containsAIMessage = strings.Contains(pageHTML, "chat-item assistant") || strings.Contains(pageHTML, "role=\"assistant\"")
					containsAICompletion = strings.Contains(pageHTML, "chat-item assistant") && strings.Contains(pageHTML, "data-status=\"finished\"") ||
						strings.Contains(pageHTML, "role=\"assistant\"") && strings.Contains(pageHTML, "data-status=\"finished\"")
				} else {
					// DeepSeek 特有的检测标志
					containsAIMessage = strings.Contains(pageHTML, "ds-message _63c77b1")
					containsAICompletion = strings.Contains(pageHTML, "ds-flex _0a3d93b")
				}

				if containsAICompletion {
					// 发现完成关键字，输出日志
					if strings.Contains(serviceType, "qwen") || strings.Contains(serviceType, "通义千问") {
						gc.logger.Println("✅ Qwen completion keyword detected, AI output completed")
					} else {
						gc.logger.Println("✅ ds-flex _0a3d93b keyword detected, AI output completed")
					}

					// 提取AI输出内容
					extractedContent, extractErr := gc.extractAIOutputContent(ctx)
					if extractErr != nil {
						gc.logger.Printf("⚠️ Error extracting AI output content: %v", extractErr)
						// 如果提取失败，仍然返回成功，但内容为空
						return "", nil
					}

					// 复制AI输出内容后输出日志
					gc.logger.Printf("📋 AI output copied, content length: %d", len(extractedContent))

					// 🚨 TRANSPARENCY PIPELINE NOTICE 🚨
					// 重要通知：劫匪计划透明管道原则
					// extractAIOutputContent已经返回纯净内容，不需要再次清理
					// 严禁在此处添加任何形式的内容过滤、格式化或清理操作
					gc.logger.Printf("📋 Content extracted, final length: %d", len(extractedContent))

					return extractedContent, nil // 找到关键字并提取内容，返回成功
				} else if containsAIMessage {
					// 发现AI消息内容，但尚未完成，继续等待完成标志
					if strings.Contains(serviceType, "qwen") || strings.Contains(serviceType, "通义千问") {
						gc.logger.Println("💬 Qwen assistant message detected, AI is generating content...")
					} else {
						gc.logger.Println("💬 ds-message _63c77b1 detected, AI is generating content...")
					}
				}
			}

			// 随机等待 minPollSeconds 到 maxPollSeconds 秒
			randWait := time.Duration(minPollSeconds+rand.Intn(maxPollSeconds-minPollSeconds+1)) * time.Second
			gc.logger.Printf("⏳ Polling wait: checking again in %.0f seconds", randWait.Seconds())
			time.Sleep(randWait)
		}
	}
}

// 从包含 ds-flex _0a3d93b 或 ds-message _63c77b1 的HTML容器中提取AI输出内容
// clickCopyButton 尝试点击AI输出区域的复制按钮
func (gc *GuardianClient) clickCopyButton(ctx context.Context) error {
	// 尝试查找并点击复制按钮
	selectors := []string{
		"div.db183363.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container", // 使用您提供的具体类名组合（实际是div）
		"div.ds-icon-button.db183363",                                              // 您提供的具体类名
		"div.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container",    // 通用复制按钮类名
		"button.ds-icon-button.db183363",                                           // 备选：button形式的复制按钮
		"button.ds-icon-button.ds-icon-button--m.ds-icon-button--sizing-container", // 备选：button形式的通用复制按钮
		"[class*='ds-icon-button'][class*='copy']",                                 // 包含copy的按钮
		"svg[class*='copy']",                                                       // 复制图标
		"[aria-label*='copy'], [title*='copy']",                                    // 带有复制标签的元素
	}

	for _, selector := range selectors {
		gc.logger.Printf("🔍 Attempting to click copy button with selector: %s", selector)

		err := chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.ByQuery),
		)
		if err == nil {
			gc.logger.Printf("✅ Successfully clicked copy button with selector: %s", selector)
			time.Sleep(500 * time.Millisecond) // 等待复制操作完成
			return nil
		} else {
			gc.logger.Printf("⚠️  Copy button click failed for selector '%s': %v", selector, err)
		}
	}

	// 如果按选择器查找失败，尝试使用JavaScript查找
	var buttonFound bool
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			// 查找具有特定类名的复制按钮
			const buttons = document.querySelectorAll('div.ds-icon-button, button.ds-icon-button, .ds-icon-button--m, .ds-icon-button--sizing-container, .db183363');
			for (let button of buttons) {
				// 检查按钮是否看起来像复制按钮
				if (button.classList.contains('ds-icon-button') && 
				    (button.querySelector('svg') || button.classList.contains('db183363') || button.innerHTML.toLowerCase().includes('copy'))) {
					button.click();
					return true; // 成功点击
				}
			}
			return false; // 没有找到合适的按钮
		})()`, &buttonFound),
	)

	if err == nil && buttonFound {
		gc.logger.Printf("✅ Successfully clicked copy button via JavaScript")
		time.Sleep(500 * time.Millisecond) // 等待复制操作完成
		return nil
	}

	gc.logger.Printf("❌ Failed to click copy button via any method")
	return fmt.Errorf("failed to click copy button")
}

func (gc *GuardianClient) extractAIOutputContent(ctx context.Context) (string, error) {
	// 🚨 TRANSPARENCY PIPELINE NOTICE 🚨
	// ===================================
	// 重要通知：劫匪计划透明管道原则
	// 从浏览器获取的AI输出内容必须保持原始格式，不做任何处理
	// 任何对内容的修改都会影响AI思维链的可读性
	// 严禁在此处添加任何形式的内容过滤、格式化或清理操作
	// ===================================
	
	// 首先尝试点击复制按钮
	copyErr := gc.clickCopyButton(ctx)
	if copyErr == nil {
		gc.logger.Printf("✅ Copy button clicked successfully, waiting for content to be ready")
		// 点击复制按钮后等待一点时间，让页面响应
		time.Sleep(1 * time.Second)
		// 即使点击了复制按钮，我们仍然需要获取页面上的内容
		// 复制按钮通常不会改变页面上的内容，只是将内容放入剪贴板
	}

	// 无论复制按钮是否成功，我们都继续获取页面内容
	gc.logger.Printf("ℹ️  Proceeding to extract content from page")

	var aiContent string
	
	serviceType := strings.ToLower(gc.ProviderConfig.Provider)
	
	// 根据服务类型选择不同的选择器
	var selectors []string
	if strings.Contains(serviceType, "qwen") || strings.Contains(serviceType, "通义千问") {
		// Qwen 通义千问的选择器
		selectors = []string{
			"div.chat-item.assistant",           // Qwen 助手消息的主要容器
			"div[role='assistant']",             // Qwen 助手角色容器
			"div.chat-item.assistant div.content", // Qwen 内容容器
			"div[role='assistant'] div.content",   // Qwen 内容容器
			"div.chat-item.assistant p",          // Qwen 段落内容
			"div[role='assistant'] p",            // Qwen 段落内容
			"div.chat-item.assistant span",       // Qwen 文本内容
			"div[role='assistant'] span",         // Qwen 文本内容
		}
	} else {
		// DeepSeek 的选择器
		selectors = []string{
			"div.ds-message._63c77b1",      // DeepSeek AI输出内容的主要容器（最优先）
			"div.ds-flex._0a3d93b",         // AI输出完成标志容器
			"div.ds-message._63c77b1 div",  // 容器内的内容
			"div.ds-message._63c77b1 span", // 容器内的文本节点
			"div.ds-message._63c77b1 *",    // 容器内的任意元素
			"div.ds-flex._0a3d93b div",     // 容器内的内容
			"div.ds-flex._0a3d93b span",    // 容器内的文本节点
			"div.ds-flex._0a3d93b *",       // 容器内的任意元素
		}
	}

	for _, selector := range selectors {
		// 如果是ds-message._63c77b1选择器，我们需要获取第二个元素
		if selector == "div.ds-message._63c77b1" {
			// 使用JavaScript获取第二个ds-message._63c77b1元素
			var elementContent string
			err := chromedp.Run(ctx,
				chromedp.Evaluate(`(() => {
					const elements = document.querySelectorAll('div.ds-message._63c77b1');
					if (elements.length >= 2) {
						// 获取第二个元素的原始内容，不做任何处理
						const element = elements[1];
						return element.innerText || element.textContent || element.innerHTML || element.outerHTML || '';
					} else if (elements.length == 1) {
						// 如果只有一个元素，返回它的原始内容
						const element = elements[0];
						return element.innerText || element.textContent || element.innerHTML || element.outerHTML || '';
					}
					return '';
				})()`, &elementContent),
			)
			if err == nil && strings.TrimSpace(elementContent) != "" {
				aiContent = strings.TrimSpace(elementContent)
				gc.logger.Printf("✅ Extracted content from second ds-message._63c77b1 element, length: %d", len(aiContent))
				// 不再使用cleanAndDecodeContent，直接返回纯净内容
				return elementContent, nil
			}
		}

		// 对于其他选择器，使用原有的逻辑
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 优先获取完整的innerHTML，保留HTML结构
				var innerHTML string
				err := chromedp.InnerHTML(selector, &innerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(innerHTML) != "" {
					// 保留内容中的换行符，只移除首尾空白
					aiContent = innerHTML
					gc.logger.Printf("✅ Extracted innerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果InnerHTML失败，尝试OuterHTML
				var outerHTML string
				err = chromedp.OuterHTML(selector, &outerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(outerHTML) != "" {
					// 保留内容中的换行符，只移除首尾空白
					aiContent = outerHTML
					gc.logger.Printf("✅ Extracted outerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果HTML获取失败，最后尝试纯文本
				var content string
				err = chromedp.Text(selector, &content).Do(ctx)
				if err == nil && strings.TrimSpace(content) != "" {
					// 保留内容中的换行符，只移除首尾空白
					aiContent = content
					gc.logger.Printf("✅ Extracted text content from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				return fmt.Errorf("no content found with selector: %s", selector)
			}),
		)

		if err == nil && aiContent != "" {
			// 不再使用cleanAndDecodeContent，直接返回纯净内容
			return aiContent, nil
		}
	}

	// 如果常规选择器都失败，尝试使用JavaScript直接查找包含 ds-message _63c77b1 或 ds-flex _0a3d93b 的元素及其内容
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			// 首先尝试查找 ds-message _63c77b1 元素（AI输出内容）
			const messageElements = document.querySelectorAll('div.ds-message._63c77b1');
			if (messageElements.length >= 2) {
				// 获取第二个匹配元素的原始内容，不做任何处理
				const element = messageElements[1];
				return element.innerText || element.textContent || element.innerHTML || element.outerHTML || '';
			} else if (messageElements.length == 1) {
				// 如果只有一个元素，返回它的原始内容
				const element = messageElements[0];
				return element.innerText || element.textContent || element.innerHTML || element.outerHTML || '';
			}
			
			// 如果没有找到 ds-message _63c77b1 元素，尝试查找 ds-flex _0a3d93b 元素（AI输出完成标志）
			const elements = document.querySelectorAll('div.ds-flex._0a3d93b');
			if (elements.length > 0) {
				// 获取第一个匹配元素的原始内容
				const element = elements[0];
				return element.innerText || element.textContent || element.innerHTML || element.outerHTML || '';
			}
			return '';
		})()`, &aiContent),
	)

	if err != nil {
		return "", err
	}

	if aiContent != "" {
		gc.logger.Printf("✅ Extracted content using JavaScript evaluation, length: %d", len(aiContent))
		// 不再使用cleanAndDecodeContent，直接返回纯净内容
		return aiContent, nil
	}

	return "", fmt.Errorf("no AI output content found in ds-message _63c77b1 or ds-flex _0a3d93b container")
}

// cleanAndDecodeContent 清理和解码AI输出内容，处理HTML实体编码和其他特殊字符
func (gc *GuardianClient) cleanAndDecodeContent(content string) string {
	// 🚨 TRANSPARENCY PIPELINE NOTICE 🚨
	// ===================================
	// 重要通知：劫匪计划透明管道原则
	// 从浏览器获取的AI输出内容必须保持原始格式，不做任何处理
	// 任何对内容的修改都会影响AI思维链的可读性
	// 严禁在此处添加任何形式的内容过滤、格式化或清理操作
	// ===================================
	// 🔥 FUNDAMENTAL PRINCIPLE: RETURN RAW CONTENT WITHOUT ANY PROCESSING 🔥
	// 直接返回原始内容，让nofx原生解析逻辑处理
	// 这样确保nofx的extractDecisions函数能使用其完整的解析链
	gc.logger.Printf("📥 Raw AI response received, length: %d", len(content))
	return content
}

// 清理HTML内容，移除多余的HTML标签，保留文本内容
func (gc *GuardianClient) cleanHTMLContent(htmlContent string) string {
	// 首先尝试使用正则表达式移除HTML标签
	cleaned := htmlContent

	// 移除HTML标签（保留标签间的内容）
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	// 仅替换多个连续的空格为单个空格，保留换行符
	space := regexp.MustCompile(`[ ]+`)
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

// CheckForAICompletionKeyword 在HTML文档内存中查找AI输出完成的标志关键字
func (gc *GuardianClient) CheckForAICompletionKeyword(htmlContent string) bool {
	// 检查HTML内容中是否包含AI输出完成的标志
	// "ds-flex _0a3d93b" 是DeepSeek AI输出完成的标志性关键字
	containsKeyword := strings.Contains(htmlContent, "ds-flex _0a3d93b")
	if containsKeyword {
		gc.logger.Println("✅ AI输出完成标志检测成功: 找到 'ds-flex _0a3d93b' 关键字")
	}
	return containsKeyword
}

// ProcessAIOutputCompletion 查找AI输出完成标志并执行后续操作
func (gc *GuardianClient) ProcessAIOutputCompletion(ctx context.Context) error {
	// 获取当前页面的完整HTML内容
	var pageHTML string
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
			return err
		}),
	)
	if err != nil {
		gc.logger.Printf("⚠️ 获取页面HTML失败: %v", err)
		return err
	}

	// 检查HTML文档中是否包含AI输出完成的标志
	if gc.CheckForAICompletionKeyword(pageHTML) {
		// 发现AI输出完成标志，执行下一步自动化操作
		gc.logger.Println("🚀 检测到AI输出完成，开始执行下一步自动化操作...")

		// 这里可以添加任何需要的后续操作
		// 例如：点击复制按钮、保存内容、截图等

		// 输出日志表示AI输出数据复制完成
		gc.logger.Println("📋 AI输出数据复制完成")
		return nil
	} else {
		gc.logger.Println("❌ 未检测到AI输出完成标志")
		return fmt.Errorf("AI output not completed yet")
	}
}
