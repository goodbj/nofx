package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// ClaudeGuardianProvider implements a specialized provider for Claude with browser automation
type ClaudeGuardianProvider struct {
	*GuardianClient
}

// NewClaudeGuardianProvider creates a new Claude Guardian provider
func NewClaudeGuardianProvider() BrowserAIProvider {
	config := DefaultConfig()
	config.Provider = "claude-browser"
	config.BaseURL = "https://claude.ai/chat"
	config.Model = "claude-browser-automation"

	client := NewGuardianClientFromConfig(*config)
	provider := &ClaudeGuardianProvider{
		GuardianClient: client,
	}

	return provider
}

// GetServiceName returns the service name
func (p *ClaudeGuardianProvider) GetServiceName() string {
	return "claude-browser"
}

// GetDefaultURL returns the default URL
func (p *ClaudeGuardianProvider) GetDefaultURL() string {
	return "https://claude.ai/chat"
}

// GetInputSelectors returns Claude-specific input selectors
func (p *ClaudeGuardianProvider) GetInputSelectors() []string {
	return []string{
		"div[data-testid='composer']",
		"textarea[placeholder*='Message']",
		"textarea[aria-label='Message']",
		"div.ProseMirror[contenteditable='true']",
		"div[role='textbox']",
		"textarea[placeholder='Message']",
		"div[contenteditable='true'][data-placeholder]",
	}
}

// GetSubmitSelectors returns Claude-specific submit selectors
func (p *ClaudeGuardianProvider) GetSubmitSelectors() []string {
	return []string{
		"button[data-testid='send-button']",
		"button[aria-label='Send message']",
		"button[type='submit']",
		"div[data-testid='send-button']",
		"button.send-button",
		"button[data-action-button='true']",
		"button[aria-label='Send']",
	}
}

// GetResponseSelectors returns Claude-specific response selectors
func (p *ClaudeGuardianProvider) GetResponseSelectors() []string {
	return []string{
		"div[data-testid='message']",
		"div.claude-message",
		"div[data-is-streaming='false']",
		"div.markdown.prose",
		"pre",
		"code",
		"div.group.flex",
		"div.ml-composer-response",
	}
}

// Custom extraction method for Claude responses
func (p *ClaudeGuardianProvider) extractClaudeContent(ctx context.Context) (string, error) {
	var aiContent string

	// Try multiple selectors specific to Claude
	selectors := []string{
		"div[data-is-streaming='false'] div[data-testid='message-content']",
		"div[data-testid='message-content']",
		"div.markdown.prose",
		"div[data-testid='message']",
		"div.claude-message",
		"div.ml-composer-response",
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
					p.logger.Printf("✅ Extracted Claude content from selector '%s', length: %d", selector, len(aiContent))
					return nil
				}

				// If InnerHTML fails, try OuterHTML
				var outerHTML string
				err = chromedp.OuterHTML(selector, &outerHTML).Do(ctx)
				if err == nil && strings.TrimSpace(outerHTML) != "" {
					aiContent = strings.TrimSpace(outerHTML)
					p.logger.Printf("✅ Extracted Claude content from selector '%s' (outerHTML), length: %d", selector, len(aiContent))
					return nil
				}

				// If HTML methods fail, try getting text content
				var content string
				err = chromedp.Text(selector, &content).Do(ctx)
				if err == nil && strings.TrimSpace(content) != "" {
					aiContent = strings.TrimSpace(content)
					p.logger.Printf("✅ Extracted Claude text content from selector '%s', length: %d", selector, len(aiContent))
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
			// Look for Claude-specific response containers
			const responseElements = document.querySelectorAll('div[data-is-streaming="false"]');
			if (responseElements.length > 0) {
				// Get the last (most recent) response
				const lastResponse = responseElements[responseElements.length - 1];
				return lastResponse.innerText || lastResponse.textContent || lastResponse.innerHTML || '';
			}
			
			// Look for message content divs
			const contentElements = document.querySelectorAll('div[data-testid="message-content"]');
			if (contentElements.length > 0) {
				return Array.from(contentElements).map(el => el.innerText || el.textContent || el.innerHTML).pop() || '';
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
		p.logger.Printf("✅ Extracted Claude content using JavaScript evaluation, Length: %d", len(aiContent))
		return p.GuardianClient.cleanAndDecodeContent(aiContent), nil
	}

	return "", fmt.Errorf("no Claude response content found")
}

// Override the call method to use Claude-specific logic
func (p *ClaudeGuardianProvider) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	p.logger.Println("🤖 Claude Guardian: Starting browser automation for AI processing")

	// Combine system and user prompts
	var combinedPrompt string
	if systemPrompt != "" && userPrompt != "" {
		combinedPrompt = fmt.Sprintf("System Prompt:\n%s\n\nUser Prompt:\n%s", systemPrompt, userPrompt)
	} else if userPrompt != "" {
		combinedPrompt = userPrompt
	} else {
		combinedPrompt = systemPrompt
	}

	// Perform browser automation with Claude-specific logic
	result, err := p.performClaudeAutomation(combinedPrompt)
	if err != nil {
		p.logger.Printf("❌ Claude Guardian Browser Automation failed: %v", err)
		return "", fmt.Errorf("claude guardian browser automation failed: %w", err)
	}

	p.logger.Println("✅ Claude Guardian Browser Automation completed successfully")
	return result, nil
}

// performClaudeAutomation handles the Claude-specific browser automation process
func (p *ClaudeGuardianProvider) performClaudeAutomation(prompt string) (string, error) {
	// Set up Chrome options for Claude
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

	p.logger.Printf("🌐 Navigating to Claude URL: %s", targetURL)

	// Navigate to Claude
	if err := chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.Sleep(3*time.Second),
	); err != nil {
		return "", fmt.Errorf("failed to navigate to Claude: %w", err)
	}

	p.logger.Println("✅ Successfully navigated to Claude")

	// Find and fill the input field with Claude-specific selectors
	inputSelectors := p.GetInputSelectors()

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			p.logger.Printf("🔍 Trying Claude input selectors: %v", inputSelectors)

			for _, selector := range inputSelectors {
				p.logger.Printf("🔍 Attempting to find input with selector: %s", selector)

				// Wait for element to be visible
				err := chromedp.WaitVisible(selector).Do(ctx)
				if err == nil {
					p.logger.Printf("✅ Found Claude input element with selector: %s", selector)

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

					p.logger.Printf("✅ Set prompt to Claude input field using selector: %s", selector)
					return nil
				} else {
					p.logger.Printf("❌ Claude input selector %s not found: %v", selector, err)
				}
			}
			return fmt.Errorf("no Claude input element found with any of the attempted selectors")
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to fill Claude input field: %w", err)
	}

	p.logger.Println("✅ Successfully filled Claude input field")

	// Click submit button
	submitSelectors := p.GetSubmitSelectors()

	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, selector := range submitSelectors {
				p.logger.Printf("👆 Attempting to click Claude submit button: %s", selector)

				err := chromedp.Click(selector).Do(ctx)
				if err == nil {
					p.logger.Printf("✅ Successfully clicked Claude submit button: %s", selector)
					return nil
				} else {
					p.logger.Printf("⚠️ Failed to click Claude submit button %s: %v", selector, err)
				}
			}
			return fmt.Errorf("failed to click any Claude submit button")
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to click Claude submit button: %w", err)
	}

	p.logger.Println("✅ Submit button clicked, waiting for response...")

	// Wait for and extract Claude response
	response, err := p.waitForAndExtractClaudeResponse(ctx)
	if err != nil {
		p.logger.Printf("⚠️ Error waiting for Claude response: %v", err)
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

	p.logger.Printf("✅ Successfully retrieved Claude response, length: %d", len(response))

	// Clean the response
	cleanedResponse := p.GuardianClient.cleanAndDecodeContent(response)
	p.logger.Printf("🧹 Claude response cleaned, final length: %d", len(cleanedResponse))

	return cleanedResponse, nil
}

// waitForAndExtractClaudeResponse waits for Claude to finish responding and extracts the content
func (p *ClaudeGuardianProvider) waitForAndExtractClaudeResponse(ctx context.Context) (string, error) {
	// Wait for a period to allow response to load
	initialWaitSeconds := getEnvInt("GUARDIAN_INITIAL_WAIT_SECONDS", 5)
	p.logger.Printf("⏳ Waiting %d seconds for Claude to start response...", initialWaitSeconds)
	time.Sleep(time.Duration(initialWaitSeconds) * time.Second)

	// Now try to extract the response
	response, err := p.extractClaudeContent(ctx)
	if err != nil {
		return "", err
	}

	// Additional wait to ensure full response is loaded
	time.Sleep(2 * time.Second)

	// Double-check and extract again if possible
	secondResponse, err := p.extractClaudeContent(ctx)
	if err == nil && len(secondResponse) > len(response) {
		p.logger.Println("💬 Claude response grew, using updated content")
		response = secondResponse
	}

	return response, nil
}

// init registers the Claude Guardian provider
func init() {
	RegisterBrowserAIProvider("claude-browser", func() BrowserAIProvider {
		return NewClaudeGuardianProvider()
	})
}
