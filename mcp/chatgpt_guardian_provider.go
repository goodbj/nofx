package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// ChatGPTGuardianProvider implements a specialized provider for ChatGPT with browser automation
type ChatGPTGuardianProvider struct {
	*GuardianClient
}

// NewChatGPTGuardianProvider creates a new ChatGPT Guardian provider
func NewChatGPTGuardianProvider() BrowserAIProvider {
	config := DefaultConfig()
	config.Provider = "chatgpt"
	config.BaseURL = "https://chat.openai.com"
	config.Model = "chatgpt-automation"

	client := NewGuardianClientFromConfig(*config)
	provider := &ChatGPTGuardianProvider{
		GuardianClient: client,
	}

	return provider
}

// GetServiceName returns the service name
func (p *ChatGPTGuardianProvider) GetServiceName() string {
	return "chatgpt"
}

// GetDefaultURL returns the default URL
func (p *ChatGPTGuardianProvider) GetDefaultURL() string {
	return "https://chat.openai.com"
}

// GetInputSelectors returns ChatGPT-specific input selectors
func (p *ChatGPTGuardianProvider) GetInputSelectors() []string {
	return []string{
		"textarea[placeholder='Send a message']",
		"textarea[id='prompt-textarea']",
		"textarea[aria-label='Chat text input']",
		"#chat-input",
		"[data-testid='chat-input']",
		"div[contenteditable='true'][tabindex='0']",
	}
}

// GetSubmitSelectors returns ChatGPT-specific submit selectors
func (p *ChatGPTGuardianProvider) GetSubmitSelectors() []string {
	return []string{
		"button[data-testid='send-button']",
		"button.send-button",
		"button[aria-label='Send']",
		".send-btn",
		"button[data-testid='send-button'] div",
	}
}

// GetResponseSelectors returns ChatGPT-specific response selectors
func (p *ChatGPTGuardianProvider) GetResponseSelectors() []string {
	return []string{
		"div.text-message span",
		"div[data-testid='response-container']",
		"div.message-response",
		".chat-message-content",
		"div.markdown.prose",
		"pre",
		"code",
	}
}

// Custom extraction method for ChatGPT responses
func (p *ChatGPTGuardianProvider) extractChatGPTContent(ctx context.Context) (string, error) {
	var aiContent string

	// Try multiple selectors specific to ChatGPT
	selectors := []string{
		"div.text-message span",
		"div.markdown.prose",
		"div[data-testid='conversation-turn-LAST'] div[data-message-author-role='assistant']",
		"div[data-testid='conversation-turn-LAST'] div.markdown",
		"div[data-message-author-role='assistant']",
		".markdown",
		"pre",
		"code",
	}

	for _, selector := range selectors {
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				// Try getting innerHTML first
				var innerHTML string
				err := chromedp.InnerHTML(selector, &innerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(innerHTML) != "" {
					aiContent = strings.TrimSpace(innerHTML)
					p.logger.Printf("✅ Extracted ChatGPT content from selector '%s', length: %d", selector, len(aiContent))
					return nil
				}

				// If InnerHTML fails, try OuterHTML
				var outerHTML string
				err = chromedp.OuterHTML(selector, &outerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(outerHTML) != "" {
					aiContent = strings.TrimSpace(outerHTML)
					p.logger.Printf("✅ Extracted ChatGPT content from selector '%s' (outerHTML), length: %d", selector, len(aiContent))
					return nil
				}

				// If HTML methods fail, try getting text content
				var content string
				err = chromedp.Text(selector, &content).Do(ctx)
				if err == nil && strings.TrimSpace(content) != "" {
					aiContent = strings.TrimSpace(content)
					p.logger.Printf("✅ Extracted ChatGPT text content from selector '%s', length: %d", selector, len(aiContent))
					return nil
				}

				return fmt.Errorf("no content found with selector: %s", selector)
			}),
		)

		if err == nil && aiContent != "" {
			return p.GuardianClient.cleanAndDecodeContent(aiContent), nil
		}
	}

	// If standard selectors fail, try JavaScript evaluation
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			// Look for ChatGPT-specific response containers
			const responseElements = document.querySelectorAll('div[data-message-author-role="assistant"]');
			if (responseElements.length > 0) {
				// Get the last (most recent) response
				const lastResponse = responseElements[responseElements.length - 1];
				return lastResponse.innerText || lastResponse.textContent || lastResponse.innerHTML || '';
			}
			
			// Fallback: get any element with markdown class
			const markdownElements = document.querySelectorAll('.markdown');
			if (markdownElements.length > 0) {
				return Array.from(markdownElements).map(el => el.innerText || el.textContent || el.innerHTML).join('\\n\\n');
			}
			
			return '';
		})()`, &aiContent),
	)

	if err != nil {
		return "", err
	}

	if aiContent != "" {
		p.logger.Printf("✅ Extracted ChatGPT content using JavaScript evaluation, length: %d", len(aiContent))
		return p.GuardianClient.cleanAndDecodeContent(aiContent), nil
	}

	return "", fmt.Errorf("no ChatGPT response content found")
}

// Override the call method to use ChatGPT-specific logic
func (p *ChatGPTGuardianProvider) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	p.logger.Println("🤖 ChatGPT Guardian: Starting browser automation for AI processing")

	// Combine system and user prompts
	var combinedPrompt string
	if systemPrompt != "" && userPrompt != "" {
		combinedPrompt = fmt.Sprintf("System Prompt:\n%s\n\nUser Prompt:\n%s", systemPrompt, userPrompt)
	} else if userPrompt != "" {
		combinedPrompt = userPrompt
	} else {
		combinedPrompt = systemPrompt
	}

	// Perform browser automation with ChatGPT-specific logic
	result, err := p.performChatGPTAutomation(combinedPrompt)
	if err != nil {
		p.logger.Printf("❌ ChatGPT Guardian Browser Automation failed: %v", err)
		return "", fmt.Errorf("chatgpt guardian browser automation failed: %w", err)
	}

	p.logger.Println("✅ ChatGPT Guardian Browser Automation completed successfully")
	return result, nil
}

// performChatGPTAutomation handles the ChatGPT-specific browser automation process
func (p *ChatGPTGuardianProvider) performChatGPTAutomation(prompt string) (string, error) {
	// Set up Chrome options for ChatGPT
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-web-security", false),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-plugins-discovery", false),
		chromedp.Flag("incognito", false),
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("renderer-process-limit", "1"),
		chromedp.Flag("max_old_space_size", "4096"),
		chromedp.Flag("no-first-run", "true"),
		chromedp.Flag("no-default-browser-check", "true"),
		chromedp.Flag("disable-backgrounding-occluded-windows", "false"), // 确保窗口可见
		chromedp.Flag("disable-renderer-backgrounding", "true"),          // 防止后台渲染
		chromedp.Flag("disable-background-timer-throttling", "true"),     // 防止定时器节流
		chromedp.Flag("disable-background-networking", "false"),          // 允许后台网络活动
		chromedp.Flag("window-size", "1000,850"),
		chromedp.Flag("user-data-dir", fmt.Sprintf("%s_%d_%p", GuardianBrowserDataDir, time.Now().UnixNano(), p)), // 每个实例使用独立的用户数据目录
		chromedp.Flag("profile-directory", "Default"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Set timeout
	forceCloseTimeoutSeconds := getEnvInt("GUARDIAN_FORCE_CLOSE_TIMEOUT_SECONDS", 290)
	timeout := time.Duration(forceCloseTimeoutSeconds) * time.Second
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel()

	// 验证URL格式
	targetURL := p.ProviderConfig.BaseURL
	if targetURL == "" {
		targetURL = DefaultGuardianBaseURL
		p.logger.Printf("⚠️ BaseURL is empty, using default: %s", targetURL)
	}
	if _, err := url.Parse(targetURL); err != nil {
		return "", fmt.Errorf("invalid target URL '%s': %w", targetURL, err)
	}

	p.logger.Printf("🌐 Navigating to ChatGPT URL: %s", targetURL)

	// Navigate to ChatGPT
	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(3*time.Second),
	); err != nil {
		return "", fmt.Errorf("failed to navigate to ChatGPT: %w", err)
	}

	p.logger.Println("✅ Successfully navigated to ChatGPT")

	// Find and fill the input field with ChatGPT-specific selectors
	inputSelectors := p.GetInputSelectors()

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			p.logger.Printf("🔍 Trying ChatGPT input selectors: %v", inputSelectors)

			for _, selector := range inputSelectors {
				p.logger.Printf("🔍 Attempting to find input with selector: %s", selector)

				// Wait for element to be visible
				err := chromedp.WaitVisible(selector).Do(ctx)
				if err == nil {
					p.logger.Printf("✅ Found ChatGPT input element with selector: %s", selector)

					// Clear and fill the input
					err = chromedp.Clear(selector).Do(ctx)
					if err != nil {
						p.logger.Printf("⚠️ Could not clear input field %s: %v", selector, err)
					}

					// Focus the element
					err = chromedp.Focus(selector).Do(ctx)
					if err != nil {
						p.logger.Printf("⚠️ Focus failed for %s: %v", selector, err)
					}

					// Set the prompt
					err = chromedp.SetValue(selector, prompt).Do(ctx)
					if err != nil {
						p.logger.Printf("❌ SetValue failed for %s: %v", selector, err)
						continue // Try next selector
					}

					p.logger.Printf("✅ Set prompt to ChatGPT input field using selector: %s", selector)
					return nil
				} else {
					p.logger.Printf("❌ ChatGPT input selector %s not found: %v", selector, err)
				}
			}
			return fmt.Errorf("no ChatGPT input element found with any of the attempted selectors")
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fill ChatGPT input field: %w", err)
	}

	p.logger.Println("✅ Successfully filled ChatGPT input field")

	// Click submit button
	submitSelectors := p.GetSubmitSelectors()

	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, selector := range submitSelectors {
				p.logger.Printf("👆 Attempting to click ChatGPT submit button: %s", selector)

				err := chromedp.Click(selector).Do(ctx)
				if err == nil {
					p.logger.Printf("✅ Successfully clicked ChatGPT submit button: %s", selector)
					return nil
				} else {
					p.logger.Printf("⚠️ Failed to click ChatGPT submit button %s: %v", selector, err)
				}
			}
			return fmt.Errorf("failed to click any ChatGPT submit button")
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to click ChatGPT submit button: %w", err)
	}

	p.logger.Println("✅ Submit button clicked, waiting for response...")

	// Wait for and extract ChatGPT response
	response, err := p.waitForAndExtractChatGPTResponse(ctx)
	if err != nil {
		p.logger.Printf("⚠️ Error waiting for ChatGPT response: %v", err)
		// Try to get whatever content is available
		var fallbackContent string
		chromedp.Run(ctx,
			chromedp.Evaluate("document.body.innerText", &fallbackContent),
		)
		if fallbackContent != "" {
			return p.GuardianClient.cleanAndDecodeContent(fallbackContent), nil
		}
		return "", err
	}

	p.logger.Printf("✅ Successfully retrieved ChatGPT response, length: %d", len(response))

	// Clean the response
	cleanedResponse := p.GuardianClient.cleanAndDecodeContent(response)
	p.logger.Printf("🧹 ChatGPT response cleaned, final length: %d", len(cleanedResponse))

	return cleanedResponse, nil
}

// waitForAndExtractChatGPTResponse waits for ChatGPT to finish responding and extracts the content
func (p *ChatGPTGuardianProvider) waitForAndExtractChatGPTResponse(ctx context.Context) (string, error) {
	// Wait for a period to allow response to load
	initialWaitSeconds := getEnvInt("GUARDIAN_INITIAL_WAIT_SECONDS", 5)
	p.logger.Printf("⏳ Waiting %d seconds for ChatGPT to start response...", initialWaitSeconds)
	time.Sleep(time.Duration(initialWaitSeconds) * time.Second)

	// Now try to extract the response
	response, err := p.extractChatGPTContent(ctx)
	if err != nil {
		return "", err
	}

	// Additional wait to ensure full response is loaded
	time.Sleep(2 * time.Second)

	// Double-check and extract again if possible
	secondResponse, err := p.extractChatGPTContent(ctx)
	if err == nil && len(secondResponse) > len(response) {
		p.logger.Println("💬 ChatGPT response grew, using updated content")
		response = secondResponse
	}

	return response, nil
}

// init registers the ChatGPT Guardian provider
func init() {
	RegisterBrowserAIProvider("chatgpt", func() BrowserAIProvider {
		return NewChatGPTGuardianProvider()
	})
}
