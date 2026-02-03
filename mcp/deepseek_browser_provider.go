package mcp

// DeepSeekBrowserProvider implements the BrowserAIProvider interface for DeepSeek
type DeepSeekBrowserProvider struct {
	*BaseBrowserAIProvider
}

// NewDeepSeekBrowserProvider creates a new DeepSeek browser provider
func NewDeepSeekBrowserProvider() BrowserAIProvider {
	provider := &DeepSeekBrowserProvider{
		BaseBrowserAIProvider: NewBaseBrowserAIProvider("deepseek-browser", "https://chat.deepseek.com"),
	}

	// Set DeepSeek-specific selectors
	provider.InputSelectors = []string{
		"textarea[placeholder='给 DeepSeek 发送消息 ']",
		"textarea._27c9245.ds-scroll-area.d96f2d2a",
		"textarea._27c9245.ds-scroll-area.d96f2d2a[placeholder='给 DeepSeek 发送消息 ']",
		"textarea[placeholder='Send a message']",
		"div.public-DraftEditor-content",
		"textarea[aria-label='Chat text input']",
		"#chat-input",
		"[data-testid='chat-input']",
	}

	provider.SubmitSelectors = []string{
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

	provider.ResponseSelectors = []string{
		"div.ds-flex._0a3d93b", // 主要的AI输出完成标识容器
		"div.text-message span",
		"div[data-testid='response-container']",
		"div.message-response",
		".chat-message-content",
		"div.markdown-body",
		"pre",
		"code",
	}

	return provider
}

// init registers the DeepSeek browser provider
func init() {
	RegisterBrowserAIProvider("deepseek-browser", NewDeepSeekBrowserProvider)
}
