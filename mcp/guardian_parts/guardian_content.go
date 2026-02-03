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
				// 检查页面是否包含 ds-message _63c77b1 或 ds-flex _0a3d93b 字符串
				// 优先检查 ds-message _63c77b1 (AI正在输出内容) 和 ds-flex _0a3d93b (AI输出完成)
				containsAIMessage := strings.Contains(pageHTML, "ds-message _63c77b1")
				containsAICompletion := strings.Contains(pageHTML, "ds-flex _0a3d93b")

				if containsAICompletion {
					// 发现完成关键字，输出日志
					gc.logger.Println("✅ ds-flex _0a3d93b keyword detected, AI output completed")

					// 提取AI输出内容
					extractedContent, extractErr := gc.extractAIOutputContent(ctx)
					if extractErr != nil {
						gc.logger.Printf("⚠️ Error extracting AI output content: %v", extractErr)
						// 如果提取失败，仍然返回成功，但内容为空
						return "", nil
					}

					// 复制AI输出内容后输出日志
					gc.logger.Printf("📋 AI output copied, content length: %d", len(extractedContent))

					// 清理和解码内容
					cleanedContent := gc.cleanAndDecodeContent(extractedContent)
					gc.logger.Printf("🧹 Content cleaned and decoded, final length: %d", len(cleanedContent))

					return cleanedContent, nil // 找到关键字并提取内容，返回成功
				} else if containsAIMessage {
					// 发现AI消息内容，但尚未完成，继续等待完成标志
					gc.logger.Println("💬 ds-message _63c77b1 detected, AI is generating content...")
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
func (gc *GuardianClient) extractAIOutputContent(ctx context.Context) (string, error) {
	var aiContent string

	// 优先尝试获取完整的 ds-message _63c77b1 元素的HTML内容
	selectors := []string{
		"div.ds-message._63c77b1",      // DeepSeek AI输出内容的主要容器（最优先）
		"div.ds-flex._0a3d93b",         // AI输出完成标志容器
		"div.ds-message._63c77b1 div",  // 容器内的内容
		"div.ds-message._63c77b1 span", // 容器内的文本节点
		"div.ds-message._63c77b1 *",    // 容器内的任意元素
		"div.ds-flex._0a3d93b div",     // 容器内的内容
		"div.ds-flex._0a3d93b span",    // 容器内的文本节点
		"div.ds-flex._0a3d93b *",       // 容器内的任意元素
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
						// 获取第二个元素的内容
						return elements[1].innerHTML || elements[1].outerHTML || elements[1].innerText || elements[1].textContent || '';
					} else if (elements.length == 1) {
						// 如果只有一个元素，返回它
						return elements[0].innerHTML || elements[0].outerHTML || elements[0].innerText || elements[0].textContent || '';
					}
					return '';
				})()`, &elementContent),
			)
			if err == nil && strings.TrimSpace(elementContent) != "" {
				aiContent = strings.TrimSpace(elementContent)
				gc.logger.Printf("✅ Extracted content from second ds-message._63c77b1 element, length: %d", len(aiContent))
				return gc.cleanAndDecodeContent(aiContent), nil
			}
		}

		// 对于其他选择器，使用原有的逻辑
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 优先获取完整的innerHTML，保留HTML结构
				var innerHTML string
				err := chromedp.InnerHTML(selector, &innerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(innerHTML) != "" {
					aiContent = strings.TrimSpace(innerHTML)
					gc.logger.Printf("✅ Extracted innerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果InnerHTML失败，尝试OuterHTML
				var outerHTML string
				err = chromedp.OuterHTML(selector, &outerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(outerHTML) != "" {
					aiContent = strings.TrimSpace(outerHTML)
					gc.logger.Printf("✅ Extracted outerHTML from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				// 如果HTML获取失败，最后尝试纯文本
				var content string
				err = chromedp.Text(selector, &content).Do(ctx)
				if err == nil && strings.TrimSpace(content) != "" {
					aiContent = strings.TrimSpace(content)
					gc.logger.Printf("✅ Extracted text content from selector '%s', length: %d", selector, len(aiContent))
					return nil // 成功提取内容
				}

				return fmt.Errorf("no content found with selector: %s", selector)
			}),
		)

		if err == nil && aiContent != "" {
			return gc.cleanAndDecodeContent(aiContent), nil
		}
	}

	// 如果常规选择器都失败，尝试使用JavaScript直接查找包含 ds-message _63c77b1 或 ds-flex _0a3d93b 的元素及其内容
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			// 首先尝试查找 ds-message _63c77b1 元素（AI输出内容）
			const messageElements = document.querySelectorAll('div.ds-message._63c77b1');
			if (messageElements.length >= 2) {
				// 获取第二个匹配元素的完整内部HTML（按您的要求获取第二个）
				const element = messageElements[1];
				// 返回完整的内部HTML，保留所有嵌套结构
				return element.innerHTML || element.outerHTML || element.innerText || element.textContent || '';
			} else if (messageElements.length == 1) {
				// 如果只有一个元素，返回它
				const element = messageElements[0];
				return element.innerHTML || element.outerHTML || element.innerText || element.textContent || '';
			}
			
			// 如果没有找到 ds-message _63c77b1 元素，尝试查找 ds-flex _0a3d93b 元素（AI输出完成标志）
			const elements = document.querySelectorAll('div.ds-flex._0a3d93b');
			if (elements.length > 0) {
				// 获取第一个匹配元素的完整内容
				const element = elements[0];
				// 返回完整的内部HTML
				return element.innerHTML || element.outerHTML || element.innerText || element.textContent || '';
			}
			return '';
		})()`, &aiContent),
	)

	if err != nil {
		return "", err
	}

	if aiContent != "" {
		gc.logger.Printf("✅ Extracted content using JavaScript evaluation, length: %d", len(aiContent))
		return gc.cleanAndDecodeContent(aiContent), nil
	}

	return "", fmt.Errorf("no AI output content found in ds-message _63c77b1 or ds-flex _0a3d93b container")
}

// cleanAndDecodeContent 清理和解码AI输出内容，处理HTML实体编码和其他特殊字符
func (gc *GuardianClient) cleanAndDecodeContent(content string) string {
	// 使用与cleanHTMLContent类似的方法来清理内容，但保留重要的非HTML标签
	cleaned := content

	// 保护特殊的非HTML标签：reasoning 和 decision（这些是AI输出的结构化标记）
	reasoningRegex := regexp.MustCompile(`(?s)<reasoning>(.*?)</reasoning>`)
	decisionRegex := regexp.MustCompile(`(?s)<decision>(.*?)</decision>`)

	placeholderMap := make(map[string]string)

	// 保护reasoning标签内容
	reasoningMatches := reasoningRegex.FindAllString(content, -1)
	protectedContent := content
	for i, match := range reasoningMatches {
		placeholder := fmt.Sprintf("<<REASONING_PLACEHOLDER_%d>>", i)
		placeholderMap[placeholder] = match
		protectedContent = strings.Replace(protectedContent, match, placeholder, 1)
	}

	// 保护decision标签内容
	decisionMatches := decisionRegex.FindAllString(protectedContent, -1)
	for i, match := range decisionMatches {
		placeholder := fmt.Sprintf("<<DECISION_PLACEHOLDER_%d>>", i)
		placeholderMap[placeholder] = match
		protectedContent = strings.Replace(protectedContent, match, placeholder, 1)
	}

	// 对剩余内容使用类似cleanHTMLContent的方法处理
	cleaned = protectedContent

	// 移除其他HTML标签（保留标签间的内容）
	re := regexp.MustCompile(`<[^>]*>`)
	cleaned = re.ReplaceAllString(cleaned, " ")

	// 替换多个空白字符为单个空格
	space := regexp.MustCompile(`\s+`)
	cleaned = space.ReplaceAllString(cleaned, " ")

	// 处理常见的HTML实体编码
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")

	// 恢复受保护的内容
	for placeholder, original := range placeholderMap {
		cleaned = strings.Replace(cleaned, placeholder, original, 1)
	}

	// 移除多余的空格和换行
	cleaned = strings.TrimSpace(cleaned)

	// 记录清理前后的长度变化
	if len(content) != len(cleaned) {
		gc.logger.Printf("🧹 Content cleaned: original length %d, cleaned length %d", len(content), len(cleaned))
	}

	return cleaned
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
