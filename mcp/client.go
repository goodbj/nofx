package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	ProviderCustom = "custom"

	MCPClientTemperature = 0.5
)

var (
	DefaultTimeout = 120 * time.Second

	MaxRetryTimes = 3

	retryableErrors = []string{
		"EOF",
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"no such host",
		"stream error",   // HTTP/2 stream error
		"INTERNAL_ERROR", // Server internal error
	}

	// TokenUsageCallback is called after each AI request with token usage info
	TokenUsageCallback func(usage TokenUsage)
)

// TokenUsage represents token usage from AI API response
type TokenUsage struct {
	Provider         string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Client AI API configuration
type Client struct {
	Provider   string
	APIKey     string
	BaseURL    string
	Model      string
	UseFullURL bool // Whether to use full URL (without appending /chat/completions)
	MaxTokens  int  // Maximum tokens for AI response

	httpClient *http.Client
	logger     Logger  // Logger (replaceable)
	config     *Config // Config object (stores all configurations)

	// hooks are used to implement dynamic dispatch (polymorphism)
	// When DeepSeekClient embeds Client, hooks point to DeepSeekClient
	// This way methods called in call() are automatically dispatched to the overridden version in subclass
	hooks clientHooks
}

// New creates default client (backward compatible)
//
// Deprecated: Recommend using NewClient(...opts) for better flexibility
func New() AIClient {
	return NewClient()
}

// NewClient creates client (supports options pattern)
//
// Usage examples:
//
//	// Basic usage (backward compatible)
//	client := mcp.NewClient()
//
//	// Custom logger
//	client := mcp.NewClient(mcp.WithLogger(customLogger))
//
//	// Custom timeout
//	client := mcp.NewClient(mcp.WithTimeout(60*time.Second))
//
//	// Combine multiple options
//	client := mcp.NewClient(
//	    mcp.WithDeepSeekConfig("sk-xxx"),
//	    mcp.WithLogger(customLogger),
//	    mcp.WithTimeout(60*time.Second),
//	)
func NewClient(opts ...ClientOption) AIClient {
	// 1. Create default config
	cfg := DefaultConfig()

	// 2. Apply user options
	for _, opt := range opts {
		opt(cfg)
	}

	// 3. Create client instance
	client := &Client{
		Provider:   cfg.Provider,
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		UseFullURL: cfg.UseFullURL,
		httpClient: cfg.HTTPClient,
		logger:     cfg.Logger,
		config:     cfg,
	}

	// 4. Set default Provider (if not set)
	if client.Provider == "" {
		client.Provider = ProviderDeepSeek
		client.BaseURL = DefaultDeepSeekBaseURL
		client.Model = DefaultDeepSeekModel
	}

	// 5. Set hooks to point to self
	client.hooks = client

	return client
}

// SetCustomAPI sets custom OpenAI-compatible API
func (client *Client) SetAPIKey(apiKey, apiURL, customModel string) {
	client.Provider = ProviderCustom
	client.APIKey = apiKey

	if apiURL != "" {
		client.BaseURL = apiURL
		// Check if URL ends with #, if so use full URL (without appending /chat/completions)
		if strings.HasSuffix(apiURL, "#") {
			client.BaseURL = strings.TrimSuffix(apiURL, "#")
			client.UseFullURL = true
		} else {
			client.UseFullURL = false
		}
		// SECURITY: Validate that the URL is a valid AI endpoint, not another service
		if !IsValidAIEndpointURL(client.BaseURL) {
			client.logger.Errorf("❌ Invalid AI endpoint URL detected: %s, reverting to default", client.BaseURL)
			// Revert to a safe default if invalid URL is detected
			client.BaseURL = "https://api.openai.com/v1"
			client.UseFullURL = false
		}
	}

	client.Model = customModel
}

func (client *Client) SetTimeout(timeout time.Duration) {
	client.httpClient.Timeout = timeout
}

// CallWithMessages template method - fixed retry flow (cannot be overridden)
func (client *Client) CallWithMessages(systemPrompt, userPrompt string) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set, please call SetAPIKey first")
	}

	// Fixed retry flow
	var lastErr error
	maxRetries := client.config.MaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.logger.Warnf("⚠️  AI API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}

		// Call the fixed single-call flow
		result, err := client.hooks.call(systemPrompt, userPrompt)
		if err == nil {
			if attempt > 1 {
				client.logger.Infof("✓ AI API retry succeeded")
			}
			return result, nil
		}

		lastErr = err
		// Check if error is retryable via hooks (supports custom retry strategy in subclass)
		if !client.hooks.isRetryableError(err) {
			return "", err
		}

		// Wait before retry
		if attempt < maxRetries {
			waitTime := client.config.RetryWaitBase * time.Duration(attempt)
			client.logger.Infof("⏳ Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

func (client *Client) setAuthHeader(reqHeader http.Header) {
	reqHeader.Set("Authorization", fmt.Sprintf("Bearer %s", client.APIKey))
}

func (client *Client) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// 应用上下文压缩（如果启用）
	if client.config.EnableContextCompression {
		userPrompt = client.compressContextIfNeeded(userPrompt)
	}

	// Build messages array
	messages := []map[string]string{}

	// If system prompt exists, add system message
	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}
	// Add user message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": userPrompt,
	})

	// Build request body
	requestBody := map[string]interface{}{
		"model":       client.Model,
		"messages":    messages,
		"temperature": client.config.Temperature, // Use configured temperature
	}
	// OpenAI newer models use max_completion_tokens instead of max_tokens
	if client.Provider == ProviderOpenAI {
		requestBody["max_completion_tokens"] = client.MaxTokens
	} else {
		requestBody["max_tokens"] = client.MaxTokens
	}
	return requestBody
}

// ============================================================
// Context Compression Functions (上下文压缩功能)
// ============================================================

// compressContextIfNeeded 根据当前prompt长度和模型上下文大小，自动压缩上下文
func (client *Client) compressContextIfNeeded(userPrompt string) string {
	// 检测模型上下文大小
	modelContextSize := client.getModelContextSize()
	if modelContextSize == 0 {
		client.logger.Debugf("[MCP Compression] Model context size unknown, skipping compression")
		return userPrompt
	}

	// 计算当前prompt长度（粗略估算：1个字符≈1个token）
	currentLength := len(userPrompt)
	usagePercent := float64(currentLength) / float64(modelContextSize) * 100

	client.logger.Debugf("[MCP Compression] Current prompt length: %d, Model context: %d, Usage: %.1f%%",
		currentLength, modelContextSize, usagePercent)

	// 如果使用率<60%，不压缩
	if usagePercent < 60 {
		client.logger.Debugf("[MCP Compression] Usage < 60%%, no compression needed")
		return userPrompt
	}

	// 应用压缩
	client.logger.Infof("🗜️ [MCP Compression] Applying context compression (usage: %.1f%%)", usagePercent)
	compressed := client.compressKlineData(userPrompt, modelContextSize, currentLength)

	compressionRatio := float64(len(compressed)) / float64(currentLength) * 100
	client.logger.Infof("✅ [MCP Compression] Compressed: %d → %d chars (%.1f%%)",
		currentLength, len(compressed), compressionRatio)

	return compressed
}

// getModelContextSize 获取当前模型的上下文大小
func (client *Client) getModelContextSize() int {
	// 如果配置中手动指定了大小，直接使用
	if client.config.ModelContextSize > 0 {
		return client.config.ModelContextSize
	}

	// 根据模型名称自动检测
	modelLower := strings.ToLower(client.Model)

	// Ollama 模型检测
	if strings.Contains(modelLower, "64k") {
		return 65536 // 64K
	}
	if strings.Contains(modelLower, "32k") {
		return 32768 // 32K
	}
	if strings.Contains(modelLower, "128k") {
		return 131072 // 128K
	}

	// 根据 Provider 和 Model 推测
	switch client.Provider {
	case "ollama":
		// Ollama 默认小模型
		if strings.Contains(modelLower, "qwen") || strings.Contains(modelLower, "coder") {
			return 8192 // 默认8K
		}
		return 4096 // 默认4K

	case ProviderDeepSeek:
		return 32768 // DeepSeek 通常32K

	case ProviderOpenAI:
		if strings.Contains(modelLower, "gpt-4") {
			return 128000 // GPT-4 128K
		}
		return 16384 // GPT-3.5 16K

	case ProviderClaude:
		return 200000 // Claude 200K

	case ProviderQwen:
		return 32768 // 通义千问 32K

	case ProviderGemini:
		return 1000000 // Gemini Pro 1M (1000K)

	default:
		client.logger.Warnf("[MCP Compression] Unknown provider: %s, model: %s", client.Provider, client.Model)
		return 0 // 未知模型，不压缩
	}
}

// compressKlineData 压缩K线数据部分（主要压缩目标）
func (client *Client) compressKlineData(prompt string, modelContextSize, currentLength int) string {
	// 计算压缩级别
	usagePercent := float64(currentLength) / float64(modelContextSize) * 100
	var compressionLevel int

	if modelContextSize >= 131072 { // >128K
		compressionLevel = 0 // 不压缩
	} else if modelContextSize >= 32768 { // 32K-128K
		if usagePercent > 90 {
			compressionLevel = 1 // 轻度压缩
		} else {
			compressionLevel = 0
		}
	} else if modelContextSize >= 8192 { // 8K-32K
		if usagePercent > 80 {
			compressionLevel = 2 // 中度压缩
		} else {
			compressionLevel = 1
		}
	} else { // <8K
		compressionLevel = 3 // 激进压缩
	}

	client.logger.Debugf("[MCP Compression] Compression level: %d", compressionLevel)

	switch compressionLevel {
	case 0:
		return prompt // 无压缩

	case 1: // 轻度压缩：只压缩K线数量
		return client.compressKlineCount(prompt, 20)

	case 2: // 中度压缩：压缩K线数量+移除成交量
		compressed := client.compressKlineCount(prompt, 12)
		return client.removeVolumeColumn(compressed)

	case 3: // 激进压缩：大幅压缩K线+移除成交量+简化数值
		compressed := client.compressKlineCount(prompt, 8)
		compressed = client.removeVolumeColumn(compressed)
		return client.simplifyNumbers(compressed)

	default:
		return prompt
	}
}

// compressKlineCount 压缩K线数据的行数
func (client *Client) compressKlineCount(prompt string, maxLines int) string {
	// 查找所有K线数据块（``` 代码块）
	klineBlockRegex := regexp.MustCompile("(?s)```\n(时间.*?\n)((?:.*?\n)+?)```")

	return klineBlockRegex.ReplaceAllStringFunc(prompt, func(match string) string {
		// 解析K线块
		parts := klineBlockRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		header := parts[1]    // 表头
		dataLines := parts[2] // 数据行

		// 分割数据行
		lines := strings.Split(strings.TrimSpace(dataLines), "\n")

		// 如果行数已经少于maxLines，不处理
		if len(lines) <= maxLines {
			return match
		}

		// 只保留最后maxLines行
		startIdx := len(lines) - maxLines
		compressedLines := lines[startIdx:]

		// 重新构建
		return fmt.Sprintf("```\n%s%s\n```", header, strings.Join(compressedLines, "\n"))
	})
}

// removeVolumeColumn 移除K线数据中的成交量列
func (client *Client) removeVolumeColumn(prompt string) string {
	// 替换表头
	prompt = strings.ReplaceAll(prompt, "时间(UTC)      开盘      最高      最低      收盘      成交量",
		"时间(UTC)      开盘      最高      最低      收盘")
	prompt = strings.ReplaceAll(prompt, "Time(UTC)      Open      High      Low       Close     Volume",
		"Time(UTC)      Open      High      Low       Close")

	// 移除数据行的最后一列（成交量）
	volumeRegex := regexp.MustCompile(`(\d{2}-\d{2} \d{2}:\d{2})\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+\S+`)
	return volumeRegex.ReplaceAllString(prompt, "$1 $2 $3 $4 $5")
}

// simplifyNumbers 简化数值精度（保留4位小数 → 2位小数）
func (client *Client) simplifyNumbers(prompt string) string {
	// 将价格从4位小数简化为2位小数
	// 例如：43560.4200 → 43560.42
	numberRegex := regexp.MustCompile(`\b(\d+)\.(\d{2})\d{2}\b`)
	return numberRegex.ReplaceAllString(prompt, "$1.$2")
}

// can be used to marshal the request body and can be overridden
func (client *Client) marshalRequestBody(requestBody map[string]any) ([]byte, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}
	return jsonData, nil
}

func (client *Client) parseMCPResponse(body []byte) (string, error) {
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API returned empty response")
	}

	// Report token usage if callback is set
	if TokenUsageCallback != nil && result.Usage.TotalTokens > 0 {
		TokenUsageCallback(TokenUsage{
			Provider:         client.Provider,
			Model:            client.Model,
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		})
	}

	return result.Choices[0].Message.Content, nil
}

func (client *Client) buildUrl() string {
	if client.UseFullURL {
		return client.BaseURL
	}
	return fmt.Sprintf("%s/chat/completions", client.BaseURL)
}

func (client *Client) buildRequest(url string, jsonData []byte) (*http.Request, error) {
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("fail to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set auth header via hooks (supports overriding in subclass)
	client.hooks.setAuthHeader(req.Header)

	return req, nil
}

// call single AI API call (fixed flow, cannot be overridden)
func (client *Client) call(systemPrompt, userPrompt string) (string, error) {
	// Print current AI configuration
	client.logger.Infof("📡 [%s] Request AI Server: BaseURL: %s", client.String(), client.BaseURL)
	client.logger.Debugf("[%s] UseFullURL: %v", client.String(), client.UseFullURL)
	if len(client.APIKey) > 8 {
		client.logger.Debugf("[%s]   API Key: %s...%s", client.String(), client.APIKey[:4], client.APIKey[len(client.APIKey)-4:])
	}

	// Validate BaseURL is not empty
	if client.BaseURL == "" {
		return "", fmt.Errorf("AI API BaseURL is empty, please check AI model configuration")
	}

	// Step 1: Build request body (via hooks for dynamic dispatch)
	requestBody := client.hooks.buildMCPRequestBody(systemPrompt, userPrompt)

	// Step 2: Serialize request body (via hooks for dynamic dispatch)
	jsonData, err := client.hooks.marshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	// DEBUG: 在发送给 AI 之前打印最终请求体（仅在 debug 日志级别下生效）
	client.logger.Debugf("[MCP %s] Final request body before send:\n%s", client.String(), string(jsonData))

	// Step 3: Build URL (via hooks for dynamic dispatch)
	requestUrl := client.hooks.buildUrl()
	client.logger.Infof("📡 [MCP %s] Request URL: %s", client.String(), requestUrl)

	// Validate URL is properly formatted
	if _, err := url.ParseRequestURI(requestUrl); err != nil {
		return "", fmt.Errorf("invalid AI API URL format: %w", err)
	}

	// Step 4: Create HTTP request (fixed logic)
	req, err := client.hooks.buildRequest(requestUrl, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Step 5: Send HTTP request (fixed logic)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Step 6: Read response body (fixed logic)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Step 7: Check HTTP status code (fixed logic)
	if resp.StatusCode != http.StatusOK {
		errorMessage := string(body)
		// Log the error with more context for debugging
		client.logger.Errorf("❌ AI API Error (Status %d): %s", resp.StatusCode, errorMessage)
		client.logger.Errorf("❌ Request details - URL: %s, Model: %s, Provider: %s", requestUrl, client.Model, client.Provider)
		return "", fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, errorMessage)
	}

	// Step 8: Parse response (via hooks for dynamic dispatch)
	result, err := client.hooks.parseMCPResponse(body)
	if err != nil {
		return "", fmt.Errorf("fail to parse AI server response: %w", err)
	}

	return result, nil
}

func (client *Client) String() string {
	return fmt.Sprintf("[Provider: %s, Model: %s]",
		client.Provider, client.Model)
}

// isRetryableError determines if error is retryable (network errors, timeouts, etc.)
func (client *Client) isRetryableError(err error) bool {
	errStr := err.Error()
	// Network errors, timeouts, EOF, etc. can be retried
	for _, retryable := range client.config.RetryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	return false
}

// ============================================================
// Builder Pattern API (Advanced Features)
// ============================================================

// CallWithRequest calls AI API using Request object (supports advanced features)
//
// This method supports:
// - Multi-turn conversation history
// - Fine-grained parameter control (temperature, top_p, penalties, etc.)
// - Function Calling / Tools
// - Streaming response (future support)
//
// Usage example:
//
//	request := NewRequestBuilder().
//	    WithSystemPrompt("You are helpful").
//	    WithUserPrompt("Hello").
//	    WithTemperature(0.8).
//	    Build()
//	result, err := client.CallWithRequest(request)
func (client *Client) CallWithRequest(req *Request) (string, error) {
	if client.APIKey == "" {
		return "", fmt.Errorf("AI API key not set, please call SetAPIKey first")
	}

	// If Model is not set in Request, use Client's Model
	if req.Model == "" {
		req.Model = client.Model
	}

	// Fixed retry flow
	var lastErr error
	maxRetries := client.config.MaxRetries

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			client.logger.Warnf("⚠️  AI API call failed, retrying (%d/%d)...", attempt, maxRetries)
		}

		// Call single request
		result, err := client.callWithRequest(req)
		if err == nil {
			if attempt > 1 {
				client.logger.Infof("✓ AI API retry succeeded")
			}
			return result, nil
		}

		lastErr = err
		// Check if error is retryable
		if !client.hooks.isRetryableError(err) {
			return "", err
		}

		// Wait before retry
		if attempt < maxRetries {
			waitTime := client.config.RetryWaitBase * time.Duration(attempt)
			client.logger.Infof("⏳ Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("still failed after %d retries: %w", maxRetries, lastErr)
}

// callWithRequest single AI API call (using Request object)
func (client *Client) callWithRequest(req *Request) (string, error) {
	// Print current AI configuration
	client.logger.Infof("📡 [%s] Request AI Server with Builder: BaseURL: %s", client.String(), client.BaseURL)
	client.logger.Debugf("[%s] Messages count: %d", client.String(), len(req.Messages))

	// Build request body (from Request object) - use hooks for polymorphism
	requestBody := client.hooks.buildRequestBodyFromRequest(req)

	// Serialize request body
	jsonData, err := client.hooks.marshalRequestBody(requestBody)
	if err != nil {
		return "", err
	}

	// Build URL
	url := client.hooks.buildUrl()
	client.logger.Infof("📡 [MCP %s] Request URL: %s", client.String(), url)

	// Create HTTP request
	httpReq, err := client.hooks.buildRequest(url, jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Send HTTP request
	resp, err := client.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	result, err := client.hooks.parseMCPResponse(body)
	if err != nil {
		return "", fmt.Errorf("fail to parse AI server response: %w", err)
	}

	return result, nil
}

// buildRequestBodyFromRequest builds request body from Request object
func (client *Client) buildRequestBodyFromRequest(req *Request) map[string]any {
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

	// Add optional parameters (only add non-nil parameters)
	if req.Temperature != nil {
		requestBody["temperature"] = *req.Temperature
	} else {
		// If not set in Request, use Client's configuration
		requestBody["temperature"] = client.config.Temperature
	}

	// OpenAI newer models use max_completion_tokens instead of max_tokens
	tokenKey := "max_tokens"
	if client.Provider == ProviderOpenAI {
		tokenKey = "max_completion_tokens"
	}
	if req.MaxTokens != nil {
		requestBody[tokenKey] = *req.MaxTokens
	} else {
		// If not set in Request, use Client's MaxTokens
		requestBody[tokenKey] = client.MaxTokens
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

	if len(req.Tools) > 0 {
		requestBody["tools"] = req.Tools
	}

	if req.ToolChoice != "" {
		requestBody["tool_choice"] = req.ToolChoice
	}

	if req.Stream {
		requestBody["stream"] = true
	}

	return requestBody
}

// OpenLongLivedBrowser opens a long-lived browser window (default implementation returns error)
func (client *Client) OpenLongLivedBrowser(targetURL string) error {
	return fmt.Errorf("OpenLongLivedBrowser is not supported by this client type: %s", client.Provider)
}
