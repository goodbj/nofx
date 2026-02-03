package mcp

// ClaudeBrowserProvider implements the BrowserAIProvider interface for Claude
type ClaudeBrowserProvider struct {
	*BaseBrowserAIProvider
}

// NewClaudeBrowserProvider creates a new Claude browser provider
func NewClaudeBrowserProvider() BrowserAIProvider {
	provider := &ClaudeBrowserProvider{
		BaseBrowserAIProvider: NewBaseBrowserAIProvider("claude-browser", "https://claude.ai/chat"),
	}

	// Set Claude-specific selectors
	provider.InputSelectors = []string{
		"div[data-testid='composer']",
		"textarea[placeholder*='Message']",
		"textarea[aria-label='Message']",
		"div.ProseMirror[contenteditable='true']",
		"div[role='textbox']",
	}

	provider.SubmitSelectors = []string{
		"button[data-testid='send-button']",
		"button[aria-label='Send message']",
		"button[type='submit']",
		"div[data-testid='send-button']",
		"button.send-button",
	}

	provider.ResponseSelectors = []string{
		"div[data-testid='message']",
		"div.claude-message",
		"div[data-is-streaming='false']",
		"div.markdown prose",
		"pre",
		"code",
		"div.group.flex",
	}

	return provider
}

// init registers the Claude browser provider
func init() {
	RegisterBrowserAIProvider("claude-browser", NewClaudeBrowserProvider)
}
