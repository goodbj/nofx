package trader

import (
	"context"
	"encoding/json"
	"fmt"
	"nofx/kernel"
	"nofx/logger"
	"nofx/mcp"
	"os"
	"strings"
)

// GuardianIntegration 将Guardian功能集成到NoFx系统中
type GuardianIntegration struct {
	client mcp.AIClient
}

// NewGuardianIntegration 创建新的Guardian集成实例
func NewGuardianIntegration(client mcp.AIClient) *GuardianIntegration {
	return &GuardianIntegration{
		client: client,
	}
}

// ProcessWithExternalAI 使用外部AI服务（如DeepSeek）处理决策
// 这将在系统生成决策后，但在执行前调用
func (gi *GuardianIntegration) ProcessWithExternalAI(systemPrompt, userPrompt string) ([]kernel.Decision, error) {
	logger.Warn("🤖 Guardian: Processing with external AI service...")
	logger.Warnf("📋 System prompt length: %d", len(systemPrompt))
	logger.Warnf("📋 User prompt length: %d", len(userPrompt))

	// 使用外部AI客户端处理提示词
	aiResponse, err := gi.client.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		logger.Errorf("❌ Guardian: Error calling external AI: %v", err)
		return nil, fmt.Errorf("failed to call external AI: %w", err)
	}

	logger.Warnf("✅ Guardian: Received response from external AI, length: %d", len(aiResponse))
	logger.Debugf("📋 Response preview: %.200s...", aiResponse) // 只记录前200个字符以避免日志过长

	// 尝試解析AI響應為決策數組
	var decisions []kernel.Decision
	err = json.Unmarshal([]byte(aiResponse), &decisions)
	if err != nil {
		// 如果不是數組，嘗試解析為單個決策
		var singleDecision kernel.Decision
		err2 := json.Unmarshal([]byte(aiResponse), &singleDecision)
		if err2 != nil {
			// 都解析失敗，嘗試從文本中提取決策
			logger.Warnf("⚠️ Guardian: AI response is not valid JSON, attempting to extract decisions from text...")
			decisions, err = gi.extractDecisionsFromText(aiResponse)
			if err != nil {
				logger.Errorf("❌ Guardian: Failed to extract decisions from text: %v", err)
				logger.Warnf("📋 AI Response Content: %s", aiResponse) // 输出完整响应内容用于调试
				return nil, fmt.Errorf("failed to parse AI response as JSON or extract from text: %w", err)
			}
		} else {
			decisions = []kernel.Decision{singleDecision}
		}
	}

	logger.Warnf("✅ Guardian: Successfully parsed %d decisions from external AI", len(decisions))

	// 验证决策
	for i, decision := range decisions {
		logger.Warnf("🤖 Guardian: Decision #%d - %s %s (confidence: %d)",
			i+1, decision.Action, decision.Symbol, decision.Confidence)
	}

	return decisions, nil
}

// extractDecisionsFromText 从AI响应文本中提取决策（如果AI返回的是纯文本而非JSON）
func (gi *GuardianIntegration) extractDecisionsFromText(response string) ([]kernel.Decision, error) {
	// 这里可以实现从文本中提取决策的逻辑
	// 由于AI可能不会严格按照JSON格式返回，我们需要从文本中提取决策
	// 这是一个增强实现，支持多种格式的响应

	// 首先尝试查找JSON数组模式
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonPart := response[startIdx : endIdx+1]
		var decisions []kernel.Decision
		err := json.Unmarshal([]byte(jsonPart), &decisions)
		if err == nil {
			return decisions, nil
		}
		// 如果JSON解析失败，记录错误但继续尝试其他方法
		logger.Warnf("⚠️ Failed to parse JSON from response: %v", err)
	}

	// 尝试查找单个JSON对象模式（花括号）
	objStartIdx := strings.Index(response, "{")
	objEndIdx := strings.LastIndex(response, "}")
	if objStartIdx != -1 && objEndIdx != -1 && objEndIdx > objStartIdx {
		jsonObj := response[objStartIdx : objEndIdx+1]
		var decision kernel.Decision
		err := json.Unmarshal([]byte(jsonObj), &decision)
		if err == nil {
			return []kernel.Decision{decision}, nil
		}
		logger.Warnf("⚠️ Failed to parse single JSON object from response: %v", err)
	}

	// 如果AI服务返回了结构化的文本决策，尝试从文本中解析
	// 检查是否包含决策关键词
	if strings.Contains(strings.ToLower(response), "action") ||
		strings.Contains(strings.ToLower(response), "symbol") ||
		strings.Contains(strings.ToLower(response), "decision") {

		// 创建一个简单的决策对象，包含AI响应的主要内容
		decision := kernel.Decision{
			Symbol:     "UNKNOWN", // 无法从文本中确定具体符号
			Action:     "unknown",
			Reasoning:  response[:min(len(response), 500)], // 只取前500个字符作为理由
			Confidence: 50,                                 // 默认置信度
		}
		logger.Infof("📝 Created decision from text response, length: %d", len(response))
		return []kernel.Decision{decision}, nil
	}

	// 如果以上方法都失败，创建一个包含AI响应的通用决策
	logger.Warnf("📝 No structured data found in response, creating generic decision")
	decision := kernel.Decision{
		Symbol:     "GENERIC",
		Action:     "info",
		Reasoning:  "AI Response: " + response[:min(len(response), 1000)], // 只取前1000字符
		Confidence: 0,
	}
	return []kernel.Decision{decision}, nil
}

// ProcessDecisionWithGuardian 在决策执行前通过Guardian进行处理
func (gi *GuardianIntegration) ProcessDecisionWithGuardian(ctx context.Context, originalDecisions []kernel.Decision, systemPrompt, userPrompt string) ([]kernel.Decision, error) {
	logger.Warn("🛡️ Guardian Integration: Processing decisions with external AI analysis")

	// 缓存 Prompt 到全局变量（用于前端查看）
	kernel.SetLastPrompt(systemPrompt, userPrompt)

	// 如果原始决策为空，直接返回
	if len(originalDecisions) == 0 {
		logger.Warn("⚠️ Guardian: No original decisions to process, attempting to get decisions from external AI")
		// 如果没有原始决策，从外部AI获取
		return gi.ProcessWithExternalAI(systemPrompt, userPrompt)
	}

	// 记录原始决策
	logger.Warnf("📊 Guardian: Original decision count: %d", len(originalDecisions))
	for i, decision := range originalDecisions {
		logger.Warnf("  [%d] %s %s (confidence: %d)", i+1, decision.Action, decision.Symbol, decision.Confidence)
	}

	// 将原始决策提供给外部AI进行分析和可能的优化
	enhancedPrompt := gi.createEnhancedPrompt(originalDecisions, systemPrompt, userPrompt)

	// 使用增强的提示词调用外部AI
	enhancedDecisions, err := gi.ProcessWithExternalAI(systemPrompt, enhancedPrompt)
	if err != nil {
		logger.Warnf("⚠️ Guardian: External AI processing failed: %v, using original decisions", err)
		// 如果外部AI处理失败，返回原始决策
		return originalDecisions, nil
	}

	// 如果外部AI提供了新决策，使用新决策
	if len(enhancedDecisions) > 0 {
		logger.Warnf("✅ Guardian: Enhanced decision count: %d", len(enhancedDecisions))
		return enhancedDecisions, nil
	}

	// 否则返回原始决策
	logger.Warn("ℹ️ Guardian: No enhanced decisions from external AI, using original decisions")
	return originalDecisions, nil
}

// createEnhancedPrompt 创建增强提示词，将原始决策提供给外部AI进行分析
func (gi *GuardianIntegration) createEnhancedPrompt(originalDecisions []kernel.Decision, systemPrompt, userPrompt string) string {
	originalDecisionsJSON, _ := json.Marshal(originalDecisions)

	enhancedPrompt := fmt.Sprintf(`Previous AI system generated the following decisions:
%s

Please review these decisions and either:
1. Confirm and optimize them if they look good
2. Modify them if improvements can be made
3. Provide completely new decisions if the original ones are inadequate

Original user context:
%s

Please respond with your final decisions in JSON format:`,
		string(originalDecisionsJSON), userPrompt)

	return enhancedPrompt
}

// GetGuardianClient 获取配置好的Guardian客户端
func GetGuardianClient(provider string) mcp.AIClient {
	// 当提供者为guardian时，使用浏览器自动化客户端
	if provider == "" || strings.ToLower(provider) == "guardian" {
		// 检查是否通过环境变量指定了特定AI服务
		aiServiceType := os.Getenv("GUARDIAN_AI_SERVICE")
		if aiServiceType != "" {
			return GetGuardianClientWithTarget(aiServiceType)
		}

		// 默认创建基本的GuardianClient，允许动态配置
		client := mcp.NewGuardianClient()
		return client
	}

	// 对于其他提供者，仍然返回对应的客户端
	switch strings.ToLower(provider) {
	case "deepseek":
		return mcp.NewDeepSeekClient()
	case "openai":
		return mcp.NewOpenAIClient()
	case "claude":
		return mcp.NewClaudeClient()
	case "qwen":
		return mcp.NewQwenClient()
	case "gemini":
		return mcp.NewGeminiClient()
	case "ollama":
		return mcp.NewOllamaClient()
	default:
		// 默认返回通用客户端
		return mcp.NewClient()
	}
}

// GetGuardianClientWithBaseURL 根据Base URL创建Guardian客户端
func GetGuardianClientWithBaseURL(baseURL string) mcp.AIClient {
	client := mcp.NewGuardianClient()

	// 如果提供了Base URL，则动态设置配置
	if baseURL != "" && baseURL != "http://localhost:8888" && baseURL != "http://127.0.0.1:8888" {
		guardianClient, ok := client.(*mcp.GuardianClient)
		if ok {
			guardianClient.SetDynamicConfig(baseURL)

			// 特别处理Docker环境显示设置
			if strings.Contains(strings.ToLower(baseURL), "display=true") {
				guardianClient.DisplayEnabled = true
			}
		}
	}

	return client
}

// GetGuardianClientWithTarget 获取具有特定目标AI服务的Guardian客户端
func GetGuardianClientWithTarget(aiServiceType string) mcp.AIClient {
	switch strings.ToLower(aiServiceType) {
	case "deepseek":
		return mcp.NewGuardianClientWithTarget(
			"https://chat.deepseek.com/",
			"#prompt-textarea",      // 输入框选择器
			"button[type='submit']", // 提交按钮选择器
			".font-light",           // 响应内容选择器
		)
	case "chatgpt":
		return mcp.NewGuardianClientWithTarget(
			"https://chat.openai.com/",
			"textarea[placeholder*='Send a message']", // 输入框选择器
			"button[data-testid='send-button']",       // 提交按钮选择器
			"[data-message-author-role='assistant']",  // 响应内容选择器
		)
	case "claude":
		return mcp.NewGuardianClientWithTarget(
			"https://claude.ai/",
			"div[data-is-empty='true'] div",          // 输入框选择器
			"button[data-testid='send-button']",      // 提交按钮选择器
			"div[data-testid='assistant-reply'] div", // 响应内容选择器
		)
	case "gemini":
		return mcp.NewGuardianClientWithTarget(
			"https://gemini.google.com/",
			"textarea[aria-label*='Describe what you need']", // 输入框选择器
			"button[aria-label*='Send']",                     // 提交按钮选择器
			"div[data-read-aloud]",                           // 响应内容选择器
		)
	default:
		// 默认使用DeepSeek配置
		return mcp.NewGuardianClient()
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
