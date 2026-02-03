package mcp

import (
	"fmt"
	"strings"
)

// GuardianAIProvider implements a general browser automation AI provider
// that can work with various AI services through browser automation
type GuardianAIProvider struct {
	*BaseBrowserAIProvider
}

// NewGuardianAIProvider creates a new Guardian AI provider
func NewGuardianAIProvider() BrowserAIProvider {
	provider := &GuardianAIProvider{
		BaseBrowserAIProvider: NewBaseBrowserAIProvider("guardian-ai", "https://chat.deepseek.com"),
	}

	// Set generic selectors that can work with multiple AI services
	provider.InputSelectors = []string{
		"textarea[placeholder*='message'], textarea[placeholder*='Message']",
		"textarea[placeholder*='input'], textarea[placeholder*='Input']",
		"textarea[aria-label*='input'], textarea[aria-label*='text']",
		"textarea[role='textbox']",
		"input[type='text']",
		"div[contenteditable='true']",
	}

	provider.SubmitSelectors = []string{
		"button[type='submit']",
		"button[aria-label*='send'], button[title*='send']",
		"button[data-testid*='send']",
		".send-button, #send-button",
		"button.send-button",
		"button[aria-label='Send']",
		".send-btn",
	}

	provider.ResponseSelectors = []string{
		"[data-testid*='response'], [data-testid*='answer']",
		".response, .answer, .result",
		"div[class*='message']",
		"div.markdown-body",
		"pre",
		"code",
	}

	return provider
}

// SetDynamicService allows changing the target service dynamically
func (g *GuardianAIProvider) SetDynamicService(serviceType string, targetURL string) {
	g.ServiceName = fmt.Sprintf("guardian-%s", strings.ToLower(serviceType))
	g.DefaultURL = targetURL

	// Update selectors based on service type
	switch strings.ToLower(serviceType) {
	case "deepseek":
		g.InputSelectors = []string{
			"textarea[placeholder='给 DeepSeek 发送消息 ']",
			"textarea._27c9245.ds-scroll-area.d96f2d2a",
			"textarea._27c9245.ds-scroll-area.d96f2d2a[placeholder='给 DeepSeek 发送消息 ']",
			"textarea[placeholder='Send a message']",
			"div.public-DraftEditor-content",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
		}

		g.SubmitSelectors = []string{
			"div._7436101.ds-icon-button.ds-icon-button--l.ds-icon-button--sizing-container[role='button'][aria-disabled='false']",
			"div._7436101.ds-icon-button[role='button']:not([aria-disabled='true'])",
			"div.ds-icon-button[role='button']:not([aria-disabled='true'])",
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

		g.ResponseSelectors = []string{
			"div.ds-flex._0a3d93b", // DeepSeek's AI output completion marker
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown-body",
			"pre",
			"code",
		}
	case "chatgpt":
		g.InputSelectors = []string{
			"textarea[placeholder='Send a message']",
			"textarea[id='prompt-textarea']",
			"textarea[aria-label='Chat text input']",
			"#chat-input",
			"[data-testid='chat-input']",
			"div[contenteditable='true'][tabindex='0']",
		}

		g.SubmitSelectors = []string{
			"button[data-testid='send-button']",
			"button.send-button",
			"button[aria-label='Send']",
			".send-btn",
			"button[data-testid='send-button'] div",
		}

		g.ResponseSelectors = []string{
			"div.text-message span",
			"div[data-testid='response-container']",
			"div.message-response",
			".chat-message-content",
			"div.markdown.prose",
			"pre",
			"code",
		}
	case "claude":
		g.InputSelectors = []string{
			"div[data-testid='composer']",
			"textarea[placeholder*='Message']",
			"textarea[aria-label='Message']",
			"div.ProseMirror[contenteditable='true']",
			"div[role='textbox']",
			"textarea[placeholder='Message']",
			"div[contenteditable='true'][data-placeholder]",
		}

		g.SubmitSelectors = []string{
			"button[data-testid='send-button']",
			"button[aria-label='Send message']",
			"button[type='submit']",
			"div[data-testid='send-button']",
			"button.send-button",
			"button[data-action-button='true']",
			"button[aria-label='Send']",
		}

		g.ResponseSelectors = []string{
			"div[data-testid='message-content']",
			"div[data-is-streaming='false']",
			"div.markdown.prose",
			"pre",
			"code",
			"div.group.flex",
			"div.ml-composer-response",
		}
		// Add more cases for other services as needed
	}
}

// CallWithMessages overrides the base method to customize behavior for Guardian
func (g *GuardianAIProvider) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	// Set the provider-specific configurations before calling
	if g.GuardianClient.BaseURL == "" || g.GuardianClient.BaseURL == "https://chat.deepseek.com" {
		// If using default URL or empty, assume DeepSeek
		g.SetDynamicService("deepseek", g.DefaultURL)
	}

	return g.GuardianClient.CallWithMessages(systemPrompt, userPrompt)
}

// init registers the Guardian AI provider
func init() {
	RegisterBrowserAIProvider("guardian-ai", NewGuardianAIProvider)
}
