package kernel

import (
	"fmt"
	"nofx/decision"
	"nofx/logger"
	"nofx/mcp"
	"nofx/store"
	"strings"
	"time"
)

// EnhancedDecisionEngine 集成增强错误处理的决策引擎
type EnhancedDecisionEngine struct {
	*StrategyEngine
	errorHandler *decision.DecisionErrorHandler
	strictMode   bool
}

// NewEnhancedDecisionEngine 创建增强决策引擎
func NewEnhancedDecisionEngine(strategyEngine *StrategyEngine, strictMode bool) *EnhancedDecisionEngine {
	return &EnhancedDecisionEngine{
		StrategyEngine: strategyEngine,
		errorHandler:   decision.NewDecisionErrorHandler(strictMode),
		strictMode:     strictMode,
	}
}

// GetFullDecisionWithEnhancedErrorHandling 带增强错误处理的完整决策获取
func (e *EnhancedDecisionEngine) GetFullDecisionWithEnhancedErrorHandling(
	ctx *Context,
	mcpClient mcp.AIClient,
	variant string,
) (*FullDecision, error) {

	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if e.StrategyEngine == nil {
		defaultConfig := store.GetDefaultStrategyConfig("en")
		e.StrategyEngine = NewStrategyEngine(&defaultConfig)
	}

	// 1. 获取市场数据
	if len(ctx.MarketDataMap) == 0 {
		if err := fetchMarketDataWithStrategy(ctx, e.StrategyEngine); err != nil {
			return nil, fmt.Errorf("failed to fetch market data: %w", err)
		}
	}

	// 2. 构建提示词
	riskConfig := e.StrategyEngine.GetRiskControlConfig()
	systemPrompt := e.StrategyEngine.BuildSystemPrompt(ctx.Account.TotalEquity, variant)
	userPrompt := e.StrategyEngine.BuildUserPrompt(ctx)

	// 缓存提示词
	lastPromptMutex.Lock()
	lastSystemPrompt = systemPrompt
	lastUserPrompt = userPrompt
	lastPromptTime = time.Now()
	lastPromptMutex.Unlock()

	// 3. 调用AI API
	aiCallStart := time.Now()
	aiResponse, err := mcpClient.CallWithMessages(systemPrompt, userPrompt)
	aiCallDuration := time.Since(aiCallStart)

	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 4. 使用增强错误处理解析AI响应
	enhancedDecision, allErrors, parseErr := e.parseAIResponseWithEnhancedErrorHandling(
		aiResponse,
		ctx.Account.TotalEquity,
		riskConfig.BTCETHMaxLeverage,
		riskConfig.AltcoinMaxLeverage,
		riskConfig.BTCETHMaxPositionValueRatio,
		riskConfig.AltcoinMaxPositionValueRatio,
	)

	if enhancedDecision != nil {
		enhancedDecision.Timestamp = time.Now()
		enhancedDecision.SystemPrompt = systemPrompt
		enhancedDecision.UserPrompt = userPrompt
		enhancedDecision.AIRequestDurationMs = aiCallDuration.Milliseconds()
		enhancedDecision.RawResponse = aiResponse
	}

	// 5. 处理错误和生成报告
	if len(allErrors) > 0 {
		userReport := e.errorHandler.GenerateUserFriendlyReport(allErrors)
		logger.Warnf("⚠️ 决策处理警告:\n%s", userReport)

		// 技术详情记录到调试日志
		if !e.strictMode {
			techDetails := e.errorHandler.GetTechnicalDetails(allErrors)
			logger.Debugf("技术详情:\n%s", techDetails)
		}
	}

	if parseErr != nil {
		return enhancedDecision, fmt.Errorf("enhanced parsing failed: %w", parseErr)
	}

	return enhancedDecision, nil
}

// parseAIResponseWithEnhancedErrorHandling 使用增强错误处理解析AI响应
func (e *EnhancedDecisionEngine) parseAIResponseWithEnhancedErrorHandling(
	aiResponse string,
	accountEquity float64,
	btcEthLeverage, altcoinLeverage int,
	btcEthPosRatio, altcoinPosRatio float64,
) (*FullDecision, []*decision.DecisionError, error) {

	// 1. 提取思维链
	cotTrace := extractCoTTrace(aiResponse)

	// 2. 使用增强错误处理提取决策
	enhancedDecisions, allErrors, err := e.extractDecisionsWithEnhancedErrorHandling(aiResponse)
	if err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: []Decision{},
		}, allErrors, fmt.Errorf("failed to extract decisions: %w", err)
	}

	// 3. 转换为原始决策格式（保持向后兼容）
	originalDecisions := e.convertEnhancedToOriginal(enhancedDecisions)

	// 4. 验证决策
	if err := validateDecisions(originalDecisions, accountEquity, btcEthLeverage, altcoinLeverage, btcEthPosRatio, altcoinPosRatio); err != nil {
		return &FullDecision{
			CoTTrace:  cotTrace,
			Decisions: originalDecisions,
		}, allErrors, fmt.Errorf("decision validation failed: %w", err)
	}

	return &FullDecision{
		CoTTrace:  cotTrace,
		Decisions: originalDecisions,
	}, allErrors, nil
}

// extractDecisionsWithEnhancedErrorHandling 使用增强错误处理提取决策
func (e *EnhancedDecisionEngine) extractDecisionsWithEnhancedErrorHandling(response string) ([]*decision.DecisionEnhanced, []*decision.DecisionError, error) {
	s := removeInvisibleRunes(response)
	s = strings.TrimSpace(s)
	s = fixMissingQuotes(s)

	var jsonPart string
	if match := reDecisionTag.FindStringSubmatch(s); match != nil && len(match) > 1 {
		jsonPart = strings.TrimSpace(match[1])
		logger.Infof("✓ 使用 <decision> 标签提取JSON")
	} else {
		jsonPart = s
		logger.Infof("⚠️ 未找到 <decision> 标签，搜索完整文本中的JSON")
	}

	jsonPart = fixMissingQuotes(jsonPart)

	if m := reJSONFence.FindStringSubmatch(jsonPart); m != nil && len(m) > 1 {
		jsonContent := strings.TrimSpace(m[1])
		jsonContent = compactArrayOpen(jsonContent)
		jsonContent = fixMissingQuotes(jsonContent)

		// 使用增强错误处理解析
		return e.errorHandler.ParseDecisionsWithAutoFix(jsonContent)
	}

	// 处理裸JSON数组
	jsonContent := strings.TrimSpace(reJSONArray.FindString(jsonPart))
	if jsonContent == "" {
		logger.Infof("⚠️ [安全降级] AI未输出JSON决策，进入安全等待模式")

		// 创建安全降级决策
		fallbackDecision := &decision.DecisionEnhanced{
			Symbol:    "ALL",
			Action:    "wait",
			Reasoning: fmt.Sprintf("模型未输出结构化JSON决策，进入安全等待模式；摘要: %s", truncateString(jsonPart, 240)),
		}

		return []*decision.DecisionEnhanced{fallbackDecision}, nil, nil
	}

	jsonContent = compactArrayOpen(jsonContent)
	jsonContent = fixMissingQuotes(jsonContent)

	// 使用增强错误处理解析
	return e.errorHandler.ParseDecisionsWithAutoFix(jsonContent)
}

// convertEnhancedToOriginal 将增强决策转换为原始格式（保持兼容性）
func (e *EnhancedDecisionEngine) convertEnhancedToOriginal(enhanced []*decision.DecisionEnhanced) []Decision {
	original := make([]Decision, len(enhanced))
	for i, enh := range enhanced {
		original[i] = Decision{
			Symbol:          enh.Symbol,
			Action:          enh.Action,
			Leverage:        int(enh.Leverage), // float64转int以保持兼容性
			PositionSizeUSD: enh.PositionSizeUSD,
			StopLoss:        enh.StopLoss,
			TakeProfit:      enh.TakeProfit,
			NewStopLoss:     enh.NewStopLoss,
			NewTakeProfit:   enh.NewTakeProfit,
			ClosePercentage: enh.ClosePercentage,
			Confidence:      enh.Confidence,
			RiskUSD:         enh.RiskUSD,
			Reasoning:       enh.Reasoning,
		}
	}
	return original
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// IntegrationWithExistingEngine 将增强处理集成到现有引擎
func IntegrationWithExistingEngine(strategyEngine *StrategyEngine) *EnhancedDecisionEngine {
	// 创建增强引擎包装现有引擎
	enhancedEngine := NewEnhancedDecisionEngine(strategyEngine, false)

	// 可以选择替换某些方法或保持原有接口
	logger.Infof("✅ 增强决策引擎已集成到现有系统")
	logger.Infof("新特性:")
	logger.Infof("  • 自动修正AI输出的JSON格式错误")
	logger.Infof("  • 处理浮点数杠杆值（如2.5倍）")
	logger.Infof("  • 智能验证和修正无效决策参数")
	logger.Infof("  • 友好的错误提示和建议")

	return enhancedEngine
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 假设您有一个现有的策略引擎
	// existingEngine := NewStrategyEngine(config)

	// 集成增强错误处理
	// enhancedEngine := IntegrationWithExistingEngine(existingEngine)

	// 使用增强引擎处理决策
	// decision, err := enhancedEngine.GetFullDecisionWithEnhancedErrorHandling(ctx, mcpClient, "default")

	logger.Infof("使用增强引擎的完整示例代码已在集成文件中提供")
}
