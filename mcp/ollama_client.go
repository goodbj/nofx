package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nofx/logger"
)

// OllamaClient Ollama client implementation
type OllamaClient struct {
	*Client // Embed base client
}

// NewOllamaClient creates new Ollama client
func NewOllamaClient() AIClient {
	client := &OllamaClient{
		Client: &Client{
			Provider:   ProviderOllama,
			BaseURL:    DefaultOllamaBaseURL,
			Model:      DefaultOllamaModel,
			MaxTokens:  2000, // Use same default as in DefaultConfig
			httpClient: &http.Client{Timeout: 300 * time.Second}, // Increase timeout for Ollama (5 minutes for large models)
			logger:     logger.NewMCPLogger(), // Use logger from imported package
			config:     DefaultConfig(),
		},
	}
	client.hooks = client // Set hooks to self for method override
	return client
}

// Constants for Ollama
const (
	ProviderOllama       = "ollama"
	DefaultOllamaBaseURL = "http://127.0.0.1:11434"
	DefaultOllamaModel   = "llama3.1"
)

// buildUrl Override URL building for Ollama (uses /api/generate instead of /chat/completions)
func (c *OllamaClient) buildUrl() string {
	// For Ollama, we use /api/generate endpoint
	baseURL := strings.TrimSuffix(c.BaseURL, "/")
	return fmt.Sprintf("%s/api/generate", baseURL)
}

// buildRequest Override request building for Ollama
func (c *OllamaClient) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	// Create HTTP request with the JSON data as body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers for Ollama
	req.Header.Set("Content-Type", "application/json")

	// For Ollama, we don't always use Authorization header
	// Only set it if API key is provided (some Ollama setups may require it)
	if c.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	}

	return req, nil
}

// marshalRequestBody Override request body marshaling for Ollama
func (c *OllamaClient) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	// For Ollama, we need to convert the OpenAI-style request to Ollama format
	ollamaBody := make(map[string]interface{})
	
	// Copy model
	if model, ok := requestBody["model"].(string); ok && model != "" {
		ollamaBody["model"] = model
	} else {
		ollamaBody["model"] = c.Model
	}
	
	// Convert messages to prompt
	if messages, ok := requestBody["messages"].([]interface{}); ok {
		var promptBuilder strings.Builder
		for _, msgInterface := range messages {
			if msg, ok := msgInterface.(map[string]interface{}); ok {
				if role, ok := msg["role"].(string); ok {
					if content, ok := msg["content"].(string); ok {
						promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(role), content))
					}
				}
			}
		}
		ollamaBody["prompt"] = promptBuilder.String()
	}
	
	// Copy options from OpenAI request to Ollama format
	if options, ok := requestBody["options"].(map[string]interface{}); ok {
		ollamaBody["options"] = options
	} else {
		// Set default options
		opts := make(map[string]interface{})
		
		if temp, ok := requestBody["temperature"]; ok {
			opts["temperature"] = temp
		} else if c.config.Temperature > 0 {
			opts["temperature"] = c.config.Temperature
		}
		
		if maxTokens, ok := requestBody["max_tokens"]; ok {
			opts["num_predict"] = maxTokens
		} else if c.MaxTokens > 0 {
			opts["num_predict"] = c.MaxTokens
		}
		
		if len(opts) > 0 {
			ollamaBody["options"] = opts
		}
	}
	
	// Set streaming to false for compatibility
	ollamaBody["stream"] = false
	
	return json.Marshal(ollamaBody)
}

// parseMCPResponse Override response parsing for Ollama
func (c *OllamaClient) parseMCPResponse(body []byte) (string, error) {
	var ollamaResp struct {
		Model     string `json:"model"`
		CreatedAt string `json:"created_at"`
		Response  string `json:"response"`
		Done      bool   `json:"done"`
		Context   []int  `json:"context"`
	}
	
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse Ollama response: %w", err)
	}
	
	// Return the response from Ollama
	return ollamaResp.Response, nil
}

// isRetryableError Override retry logic for Ollama
func (c *OllamaClient) isRetryableError(err error) bool {
	errStr := err.Error()
	
	// Common Ollama-specific retry conditions
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "connection reset") {
		return true
	}
	
	// Default retry logic
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF")
}

// CallWithRequest Implement CallWithRequest for Ollama
func (c *OllamaClient) CallWithRequest(req *Request) (string, error) {
	if c.APIKey == "" {
		// Ollama doesn't always require an API key, so we allow empty keys
		c.logger.Warnf("⚠️  Ollama API key is not set, some configurations may require it")
	}
	
	// Fixed retry flow
	var lastErr error
	maxRetries := c.config.MaxRetries
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			c.logger.Warnf("⚠️  Ollama API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}
		
		// Call single request
		result, err := c.callWithRequest(req)
		if err == nil {
			if attempt > 1 {
				c.logger.Infof("✓ Ollama API retry succeeded")
			}
			return result, nil
		}
		
		lastErr = err
		// Check if error is retryable
		if !c.hooks.isRetryableError(err) {
			return "", err
		}
		
		// Wait before retry
		if attempt < maxRetries {
			waitTime := c.config.RetryWaitBase * time.Duration(attempt)
			c.logger.Infof("⏳ Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}
	
	return "", fmt.Errorf("Ollama API call still failed after %d retries: %w", maxRetries, lastErr)
}

// buildRequestBodyFromRequest Override request body building for Ollama
func (c *OllamaClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	// Convert Message to API format
	messages := make([]map[string]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Build basic request body
	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}

	// Add optional parameters (Ollama specific)
	if req.Temperature != nil {
		requestBody["temperature"] = *req.Temperature
	} else {
		// If not set in Request, use Client's configuration
		requestBody["temperature"] = c.config.Temperature
	}

	// Ollama uses "num_predict" instead of "max_tokens"
	if req.MaxTokens != nil {
		requestBody["num_predict"] = *req.MaxTokens
	} else {
		// If not set in Request, use Client's MaxTokens
		requestBody["num_predict"] = c.MaxTokens
	}

	if req.TopP != nil {
		requestBody["top_p"] = *req.TopP
	}

	if req.FrequencyPenalty != nil {
		requestBody["frequency_penalty"] = *req.FrequencyPenalty
	}

	if req.PresencePenalty != nil {
		requestBody["presence_penalty"] = *req.PresencePenalty
	}

	if len(req.Stop) > 0 {
		requestBody["stop"] = req.Stop
	}

	// Ollama doesn't support tools in the same way as OpenAI, so we skip tools

	// Ollama doesn't support streaming in the same way as OpenAI, so we set to false
	requestBody["stream"] = false

	return requestBody
}

// callWithRequest Internal implementation for Ollama
func (c *OllamaClient) callWithRequest(req *Request) (string, error) {
	// Print current AI configuration
	c.logger.Infof("📡 [Ollama] Request Ollama Server: BaseURL: %s", c.BaseURL)
	c.logger.Debugf("[Ollama] Messages count: %d", len(req.Messages))

	// Build request body (from Request object) - use Ollama specific implementation
	requestBody := c.buildRequestBodyFromRequest(req)

	// Serialize request body
	jsonData, err := c.hooks.marshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	// Build URL
	url := c.hooks.buildUrl()
	c.logger.Infof("📡 [MCP Ollama] Request URL: %s", url)

	// Create HTTP request
	httpReq, err := c.hooks.buildRequest(url, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama request: %w", err)
	}

	// Send HTTP request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send Ollama request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Ollama response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	result, err := c.hooks.parseMCPResponse(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse Ollama server response: %w", err)
	}

	return result, nil
}