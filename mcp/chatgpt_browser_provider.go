package mcp

// ChatGPTBrowserProvider implements the BrowserAIProvider interface for ChatGPT
type ChatGPTBrowserProvider struct {
	*BaseBrowserAIProvider
}

// NewChatGPTBrowserProvider creates a new ChatGPT browser provider
func NewChatGPTBrowserProvider() BrowserAIProvider {
	provider := &ChatGPTBrowserProvider{
		BaseBrowserAIProvider: NewBaseBrowserAIProvider("chatgpt-browser", "https://chat.openai.com"),
	}

	// Set ChatGPT-specific selectors
	provider.InputSelectors = []string{
		"textarea[placeholder='Send a message']",
		"textarea[id='prompt-textarea']",
		"textarea[aria-label='Chat text input']",
		"#chat-input",
		"[data-testid='chat-input']",
		"div[contenteditable='true'][tabindex='0']",
	}

	provider.SubmitSelectors = []string{
		"button[data-testid='send-button']",
		"button.send-button",
		"button[aria-label='Send']",
		".send-btn",
		"button[data-testid='send-button'] div",
	}

	provider.ResponseSelectors = []string{
		"div.text-message span",
		"div[data-testid='response-container']",
		"div.message-response",
		".chat-message-content",
		"div.markdown prose",
		"pre",
		"code",
	}

	return provider
}

// init registers the ChatGPT browser provider
func init() {
	RegisterBrowserAIProvider("chatgpt-browser", NewChatGPTBrowserProvider)
}
