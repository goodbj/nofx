package mcp

import (
	"fmt"
	"net/http"
)

// BypassClient wraps an existing client and optionally bypasses API calls using browser automation
type BypassClient struct {
	*Client                           // Embed base client
	browserProvider BrowserAIProvider // Browser automation provider
	enableBypass    bool              // Flag to enable/disable bypass
}

// NewBypassClient creates a new client with optional API bypass capability
func NewBypassClient(baseClient AIClient, enableBypass bool, browserProviderType string) AIClient {
	// Create base client - preserve original client configuration if possible
	var base *Client
	if client, ok := baseClient.(*Client); ok {
		// If it's already a base client, use it directly
		base = client
	} else {
		// If not a base client, create a new one but try to preserve key configuration
		base = NewClient().(*Client)
		// Copy essential configuration from the original client if it has those methods
		if origClient, ok := baseClient.(interface{ GetProvider() string }); ok {
			base.Provider = origClient.GetProvider()
		}
		if origClient, ok := baseClient.(interface{ GetModel() string }); ok {
			base.Model = origClient.GetModel()
		}
		if origClient, ok := baseClient.(interface{ GetBaseURL() string }); ok {
			base.BaseURL = origClient.GetBaseURL()
		}
		if origClient, ok := baseClient.(interface{ GetAPIKey() string }); ok {
			base.APIKey = origClient.GetAPIKey()
		}
	}

	client := &BypassClient{
		Client:       base,
		enableBypass: enableBypass,
	}

	// If bypass is enabled, create the browser provider
	if enableBypass {
		provider, err := CreateBrowserAIProvider(browserProviderType)
		if err != nil {
			client.logger.Errorf("Failed to create browser provider for %s: %v", browserProviderType, err)
			client.enableBypass = false // Disable bypass on error
		} else {
			client.browserProvider = provider
		}
	}

	// Set hooks to point to self for method interception
	client.Client.hooks = client

	return client
}

// call intercepts the API call and optionally uses browser automation
func (bc *BypassClient) call(systemPrompt, userPrompt string) (string, error) {
	if bc.enableBypass && bc.browserProvider != nil {
		// Use browser automation bypass
		bc.logger.Infof("🔄 Using browser automation bypass for API call")
		result, err := bc.browserProvider.CallWithMessages(systemPrompt, userPrompt)
		if err != nil {
			bc.logger.Errorf("❌ Browser automation failed: %v", err)
			// IMPORTANT: Do not fall back to API call, return error directly
			return "", fmt.Errorf("browser automation failed and fallback to API is disabled: %w", err)
		}
		return result, nil
	}

	// If bypass is not enabled, use standard API call
	bc.logger.Infof("📡 Using standard API call")
	return bc.Client.call(systemPrompt, userPrompt)
}

// CallWithMessages implements the AIClient interface
func (bc *BypassClient) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	return bc.call(systemPrompt, userPrompt)
}

// CallWithRequest implements the AIClient interface
func (bc *BypassClient) CallWithRequest(req *Request) (string, error) {
	if bc.enableBypass && bc.browserProvider != nil {
		// Convert Request to messages format for browser automation
		var systemMsg, userMsg string
		for _, msg := range req.Messages {
			if msg.Role == "system" {
				systemMsg = msg.Content
			} else if msg.Role == "user" || msg.Role == "assistant" {
				userMsg = msg.Content
			}
		}
		return bc.browserProvider.CallWithMessages(systemMsg, userMsg)
	}

	// Fall back to base client's implementation
	return bc.Client.CallWithRequest(req)
}

// OpenLongLivedBrowser implements the AIClient interface
func (bc *BypassClient) OpenLongLivedBrowser(targetURL string) error {
	if bc.browserProvider != nil {
		// If the browser provider has this method, use it
		if opener, ok := bc.browserProvider.(interface{ OpenLongLivedBrowser(string) error }); ok {
			return opener.OpenLongLivedBrowser(targetURL)
		}
	}

	// Otherwise, fall back to base client's original hooks if not ourselves
	if bc.Client.hooks != bc {
		if opener, ok := bc.Client.hooks.(interface{ OpenLongLivedBrowser(string) error }); ok {
			return opener.OpenLongLivedBrowser(targetURL)
		}
	}

	return fmt.Errorf("OpenLongLivedBrowser not supported")
}

// Override hook methods to delegate to base client's original hooks
// (to avoid circular reference when base client's hooks are not BypassClient)
func (bc *BypassClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// If the original hooks are not ourselves, use them; otherwise, use the Client's default implementation
	if bc.Client.hooks != bc {
		return bc.Client.hooks.buildMCPRequestBody(systemPrompt, userPrompt)
	}
	// Otherwise, call the embedded Client's methods directly to avoid circular reference
	return bc.Client.buildMCPRequestBody(systemPrompt, userPrompt)
}

func (bc *BypassClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.buildRequestBodyFromRequest(req)
	}
	return bc.Client.buildRequestBodyFromRequest(req)
}

func (bc *BypassClient) buildUrl() string {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.buildUrl()
	}
	return bc.Client.buildUrl()
}

func (bc *BypassClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.buildRequest(url, jsonData)
	}
	return bc.Client.buildRequest(url, jsonData)
}

func (bc *BypassClient) setAuthHeader(reqHeaders http.Header) {
	if bc.Client.hooks != bc {
		bc.Client.hooks.setAuthHeader(reqHeaders)
	} else {
		bc.Client.setAuthHeader(reqHeaders)
	}
}

func (bc *BypassClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.marshalRequestBody(requestBody)
	}
	return bc.Client.marshalRequestBody(requestBody)
}

func (bc *BypassClient) parseMCPResponse(body []byte) (string, error) {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.parseMCPResponse(body)
	}
	return bc.Client.parseMCPResponse(body)
}

func (bc *BypassClient) isRetryableError(err error) bool {
	if bc.Client.hooks != bc {
		return bc.Client.hooks.isRetryableError(err)
	}
	return bc.Client.isRetryableError(err)
}
