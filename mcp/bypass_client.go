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
	// Create base client
	base, ok := baseClient.(*Client)
	if !ok {
		// If not a base client, create a new one
		base = NewClient().(*Client)
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

// Override other hook methods to delegate to base client
func (bc *BypassClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	return bc.Client.hooks.buildMCPRequestBody(systemPrompt, userPrompt)
}

func (bc *BypassClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	return bc.Client.hooks.buildRequestBodyFromRequest(req)
}

func (bc *BypassClient) buildUrl() string {
	return bc.Client.hooks.buildUrl()
}

func (bc *BypassClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	return bc.Client.hooks.buildRequest(url, jsonData)
}

func (bc *BypassClient) setAuthHeader(reqHeaders http.Header) {
	bc.Client.hooks.setAuthHeader(reqHeaders)
}

func (bc *BypassClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	return bc.Client.hooks.marshalRequestBody(requestBody)
}

func (bc *BypassClient) parseMCPResponse(body []byte) (string, error) {
	return bc.Client.hooks.parseMCPResponse(body)
}

func (bc *BypassClient) isRetryableError(err error) bool {
	return bc.Client.hooks.isRetryableError(err)
}
