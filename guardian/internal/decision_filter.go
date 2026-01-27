package guardian

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// DecisionFilter 决策过滤器
type DecisionFilter struct {
	MinConfidence  int      // 最低置信度
	MaxRiskUSD     float64  // 最大风险金额
	AllowedActions []string // 允许的操作
	AllowedSymbols []string // 允许的交易对
}

// NewDecisionFilter 创建新的决策过滤器
func NewDecisionFilter() *DecisionFilter {
	return &DecisionFilter{
		MinConfidence:  10,    // 默认最低置信度为10
		MaxRiskUSD:     100.0, // 默认最大风险为100 USDT
		AllowedActions: []string{"open_long", "open_short", "close_long", "close_short", "hold", "wait", "update_stop_loss", "update_take_profit", "partial_close", "trailing_stop", "dynamic_take_profit", "add_to_position"},
		AllowedSymbols: []string{}, // 空表示允许所有交易对
	}
}

// ValidateDecision 验证决策是否符合基本要求
func (df *DecisionFilter) ValidateDecision(decision Decision) error {
	// 检查交易动作
	action := strings.ToLower(decision.Action)
	validAction := false
	for _, allowedAction := range df.AllowedActions {
		if action == strings.ToLower(allowedAction) {
			validAction = true
			break
		}
	}
	if !validAction {
		return fmt.Errorf("invalid action: %s, allowed actions: %v", decision.Action, df.AllowedActions)
	}

	// 检查交易对符号
	if err := df.validateSymbol(decision.Symbol); err != nil {
		return fmt.Errorf("invalid symbol: %v", err)
	}

	// 检查置信度
	if decision.Confidence < df.MinConfidence {
		return fmt.Errorf("confidence %d is below minimum threshold %d", decision.Confidence, df.MinConfidence)
	}
	if decision.Confidence > 100 {
		return fmt.Errorf("confidence %d exceeds maximum value 100", decision.Confidence)
	}

	// 检查风险金额
	if decision.RiskUSD > df.MaxRiskUSD {
		return fmt.Errorf("risk amount %f exceeds maximum allowed risk %f", decision.RiskUSD, df.MaxRiskUSD)
	}

	return nil
}

// validateSymbol 验证交易对符号
func (df *DecisionFilter) validateSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	// 检查是否在允许的交易对列表中（如果列表不为空）
	if len(df.AllowedSymbols) > 0 {
		found := false
		for _, allowedSymbol := range df.AllowedSymbols {
			if strings.ToUpper(symbol) == strings.ToUpper(allowedSymbol) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("symbol %s is not in allowed list: %v", symbol, df.AllowedSymbols)
		}
	}

	// 使用正则表达式检查交易对格式 (例如: BTCUSDT, ETH-BTC 等)
	symbolRegex := regexp.MustCompile(`^[A-Z0-9]+[/_\-]?[A-Z0-9]+$`)
	if !symbolRegex.MatchString(symbol) {
		return fmt.Errorf("symbol %s has invalid format", symbol)
	}

	return nil
}

// ParseAndValidateDecision 解析并验证决策
func (df *DecisionFilter) ParseAndValidateDecision(decisionStr string) ([]Decision, error) {
	log.Printf("🔍 开始解析和验证决策: %d 字符", len(decisionStr))

	// 尝试解析为决策数组
	var decisions []Decision
	err := json.Unmarshal([]byte(decisionStr), &decisions)
	if err != nil {
		// 如果不是数组，尝试解析为单个决策
		var singleDecision Decision
		err2 := json.Unmarshal([]byte(decisionStr), &singleDecision)
		if err2 != nil {
			// 都解析失败，尝试解析为包装过的响应
			return df.parseWrappedResponse(decisionStr)
		}

		decisions = []Decision{singleDecision}
	}

	// 验证每个决策
	for i, decision := range decisions {
		// 如果是默认HOLD决策（来自非JSON响应的包装），跳过验证或给出警告
		if decision.Action == "hold" && decision.Reasoning == decisionStr && len(decisions) == 1 {
			log.Printf("⚠️ 接收到非JSON格式的AI响应，已包装为HOLD决策，需要进一步处理: %s", TruncateString(decisionStr, 100))
			continue // 对于包装的响应，我们允许通过，但记录警告
		}

		if err := df.ValidateDecision(decision); err != nil {
			return nil, fmt.Errorf("validation failed for decision #%d: %v", i+1, err)
		}

		log.Printf("✅ 决策 #%d 验证通过: %s %s (置信度: %d)",
			i+1, decision.Action, decision.Symbol, decision.Confidence)
	}

	log.Printf("✅ 所有 %d 个决策验证通过", len(decisions))
	return decisions, nil
}

// parseWrappedResponse 解析包装过的响应
func (df *DecisionFilter) parseWrappedResponse(responseStr string) ([]Decision, error) {
	// 尝试查找JSON对象嵌入在文本中的情况
	jsonRegex := regexp.MustCompile(`\[{.*}]|\{.*}`)
	matches := jsonRegex.FindAllString(responseStr, -1)

	for _, match := range matches {
		var decisions []Decision
		if err := json.Unmarshal([]byte(match), &decisions); err == nil {
			log.Printf("🔍 在文本中找到JSON格式的决策: %d 个", len(decisions))

			// 验证找到的决策
			for i, decision := range decisions {
				if err := df.ValidateDecision(decision); err != nil {
					log.Printf("⚠️ 决策 #%d 验证警告: %v", i+1, err)
					continue
				}

				log.Printf("✅ 决策 #%d 验证通过: %s %s (置信度: %d)",
					i+1, decision.Action, decision.Symbol, decision.Confidence)
			}

			return decisions, nil
		}

		// 如果作为数组解析失败，尝试作为单个对象解析
		var singleDecision Decision
		if err := json.Unmarshal([]byte(match), &singleDecision); err == nil {
			if err := df.ValidateDecision(singleDecision); err == nil {
				log.Printf("✅ 解析为单个决策并验证通过: %s %s (置信度: %d)",
					singleDecision.Action, singleDecision.Symbol, singleDecision.Confidence)
				return []Decision{singleDecision}, nil
			}
		}
	}

	// 如果找不到有效的JSON，返回包装的HOLD决策
	log.Printf("⚠️ 无法从响应中解析出有效决策，将响应包装为HOLD决策: %s", TruncateString(responseStr, 100))

	defaultDecision := Decision{
		Action:     "hold",
		Symbol:     "BTCUSDT",   // 默认交易对
		Confidence: 50,          // 默认置信度
		Reasoning:  responseStr, // 将原始响应作为理由
	}

	return []Decision{defaultDecision}, nil
}

// FilterDecisions 过滤决策，只返回通过验证的决策
func (df *DecisionFilter) FilterDecisions(decisions []Decision) []Decision {
	var filtered []Decision

	for _, decision := range decisions {
		if err := df.ValidateDecision(decision); err == nil {
			filtered = append(filtered, decision)
		} else {
			log.Printf("❌ 决策被过滤掉: %v", err)
		}
	}

	return filtered
}

// SetMaxRisk 设置最大风险金额
func (df *DecisionFilter) SetMaxRisk(risk float64) {
	df.MaxRiskUSD = risk
	log.Printf("🔧 设置最大风险金额为: %f", risk)
}

// SetMinConfidence 设置最低置信度
func (df *DecisionFilter) SetMinConfidence(confidence int) {
	df.MinConfidence = confidence
	log.Printf("🔧 设置最低置信度为: %d", confidence)
}

// AddAllowedSymbol 添加允许的交易对
func (df *DecisionFilter) AddAllowedSymbol(symbol string) {
	df.AllowedSymbols = append(df.AllowedSymbols, symbol)
	log.Printf("🔧 添加允许的交易对: %s", symbol)
}

// RemoveAllowedSymbol 移除允许的交易对
func (df *DecisionFilter) RemoveAllowedSymbol(symbol string) {
	for i, s := range df.AllowedSymbols {
		if strings.ToUpper(s) == strings.ToUpper(symbol) {
			df.AllowedSymbols = append(df.AllowedSymbols[:i], df.AllowedSymbols[i+1:]...)
			log.Printf("🔧 移除允许的交易对: %s", symbol)
			return
		}
	}
}

// IsSymbolAllowed 检查交易对是否被允许
func (df *DecisionFilter) IsSymbolAllowed(symbol string) bool {
	if len(df.AllowedSymbols) == 0 {
		return true // 如果没有限制，则允许所有交易对
	}

	for _, allowedSymbol := range df.AllowedSymbols {
		if strings.ToUpper(symbol) == strings.ToUpper(allowedSymbol) {
			return true
		}
	}

	return false
}
