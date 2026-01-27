package guardian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nofx/kernel"
	"nofx/mcp"
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
	log.Println("🤖 Guardian: Processing with external AI service...")

	// 使用外部AI客户端处理提示词
	aiResponse, err := gi.client.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		log.Printf("❌ Guardian: Error calling external AI: %v", err)
		return nil, fmt.Errorf("failed to call external AI: %w", err)
	}

	log.Printf("✅ Guardian: Received response from external AI, length: %d", len(aiResponse))

	// 尝试解析AI响应为决策数组
	var decisions []kernel.Decision
	err = json.Unmarshal([]byte(aiResponse), &decisions)
	if err != nil {
		// 如果不是数组，尝试解析为单个决策
		var singleDecision kernel.Decision
		err2 := json.Unmarshal([]byte(aiResponse), &singleDecision)
		if err2 != nil {
			// 都解析失败，尝试从文本中提取决策
			log.Printf("⚠️ Guardian: AI response is not valid JSON, attempting to extract decisions from text...")
			decisions, err = gi.extractDecisionsFromText(aiResponse)
			if err != nil {
				log.Printf("❌ Guardian: Failed to extract decisions from text: %v", err)
				return nil, fmt.Errorf("failed to parse AI response as JSON or extract from text: %w", err)
			}
		} else {
			decisions = []kernel.Decision{singleDecision}
		}
	}

	log.Printf("✅ Guardian: Successfully parsed %d decisions from external AI", len(decisions))

	// 验证决策
	for i, decision := range decisions {
		log.Printf("🤖 Guardian: Decision #%d - %s %s (confidence: %d)",
			i+1, decision.Action, decision.Symbol, decision.Confidence)
	}

	return decisions, nil
}

// extractDecisionsFromText 从AI响应文本中提取决策（如果AI返回的是纯文本而非JSON）
func (gi *GuardianIntegration) extractDecisionsFromText(response string) ([]kernel.Decision, error) {
	// 这里可以实现从文本中提取决策的逻辑
	// 由于AI可能不会严格按照JSON格式返回，我们需要从文本中提取决策
	// 这是一个简化实现，实际应用中可能需要更复杂的解析逻辑

	// 查找JSON数组模式
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")

	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonPart := response[startIdx : endIdx+1]
		var decisions []kernel.Decision
		err := json.Unmarshal([]byte(jsonPart), &decisions)
		if err == nil {
			return decisions, nil
		}
	}

	// 如果无法提取，返回错误
	return nil, fmt.Errorf("could not extract decisions from text response")
}

// ProcessDecisionWithGuardian 在决策执行前通过Guardian进行处理
func (gi *GuardianIntegration) ProcessDecisionWithGuardian(ctx context.Context, originalDecisions []kernel.Decision, systemPrompt, userPrompt string) ([]kernel.Decision, error) {
	log.Println("🛡️ Guardian Integration: Processing decisions with external AI analysis")

	// 如果原始决策为空，直接返回
	if len(originalDecisions) == 0 {
		log.Println("⚠️ Guardian: No original decisions to process, attempting to get decisions from external AI")
		// 如果没有原始决策，从外部AI获取
		return gi.ProcessWithExternalAI(systemPrompt, userPrompt)
	}

	// 记录原始决策
	log.Printf("📊 Guardian: Original decision count: %d", len(originalDecisions))
	for i, decision := range originalDecisions {
		log.Printf("  [%d] %s %s (confidence: %d)", i+1, decision.Action, decision.Symbol, decision.Confidence)
	}

	// 将原始决策提供给外部AI进行分析和可能的优化
	enhancedPrompt := gi.createEnhancedPrompt(originalDecisions, systemPrompt, userPrompt)

	enhancedDecisions, err := gi.ProcessWithExternalAI(systemPrompt, enhancedPrompt)
	if err != nil {
		log.Printf("⚠️ Guardian: External AI processing failed: %v, using original decisions", err)
		// 如果外部AI处理失败，返回原始决策
		return originalDecisions, nil
	}

	// 如果外部AI提供了新决策，使用新决策
	if len(enhancedDecisions) > 0 {
		log.Printf("✅ Guardian: Enhanced decision count: %d", len(enhancedDecisions))
		return enhancedDecisions, nil
	}

	// 否则返回原始决策
	log.Println("ℹ️ Guardian: No enhanced decisions from external AI, using original decisions")
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
	// 根据配置返回相应的AI客户端，这里使用DeepSeek作为默认
	if provider == "" || strings.ToLower(provider) == "deepseek" {
		// 返回DeepSeek客户端
		client := mcp.NewDeepSeekClient()
		return client
	}

	// 可以扩展支持其他AI提供商
	return mcp.NewDeepSeekClient() // 默认返回DeepSeek
}
