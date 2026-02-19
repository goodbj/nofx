package decision

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DecisionEnhanced 增强版决策结构体，支持浮点数杠杆
type DecisionEnhanced struct {
	Symbol          string  `json:"symbol"`
	Action          string  `json:"action"`
	Leverage        float64 `json:"leverage,omitempty"` // 改为float64支持小数杠杆
	PositionSizeUSD float64 `json:"position_size_usd,omitempty"`
	StopLoss        float64 `json:"stop_loss,omitempty"`
	TakeProfit      float64 `json:"take_profit,omitempty"`
	NewStopLoss     float64 `json:"new_stop_loss,omitempty"`
	NewTakeProfit   float64 `json:"new_take_profit,omitempty"`
	ClosePercentage float64 `json:"close_percentage,omitempty"`
	MaxDrawdown     float64 `json:"max_drawdown,omitempty"`
	MinTargetProfit float64 `json:"min_target_profit,omitempty"`
	MaxPositionUSD  float64 `json:"max_position_usd,omitempty"`
	MaxDailyLoss    float64 `json:"max_daily_loss,omitempty"`
	TimeInForce     string  `json:"time_in_force,omitempty"`
	Confidence      int     `json:"confidence,omitempty"`
	RiskUSD         float64 `json:"risk_usd,omitempty"`
	Reasoning       string  `json:"reasoning"`
}

// ErrorType 错误类型定义
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeLeverageTypeMismatch
	ErrorTypeInvalidJSON
	ErrorTypeMissingField
	ErrorTypeValueOutOfRange
)

// DecisionError 决策处理错误结构
type DecisionError struct {
	Type         ErrorType
	Field        string
	OriginalVal  interface{}
	CorrectedVal interface{}
	Message      string
	Suggestion   string
}

func (e *DecisionError) Error() string {
	return fmt.Sprintf("决策处理错误 [%s]: %s", e.Field, e.Message)
}

// DecisionErrorHandler 决策错误处理器
type DecisionErrorHandler struct {
	strictMode bool
}

// NewDecisionErrorHandler 创建新的决策错误处理器
func NewDecisionErrorHandler(strictMode bool) *DecisionErrorHandler {
	return &DecisionErrorHandler{
		strictMode: strictMode,
	}
}

// PreprocessJSON 预处理JSON，修正常见错误
func (h *DecisionErrorHandler) PreprocessJSON(jsonContent string) (string, []*DecisionError) {
	errors := make([]*DecisionError, 0)

	// 1. 修复杠杆类型问题：将整数形式的浮点数转换为标准格式
	leveragePattern := regexp.MustCompile(`"leverage"\s*:\s*(\d+(?:\.\d+)?)`)
	matches := leveragePattern.FindAllStringSubmatch(jsonContent, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			originalValue := match[1]
			// 检查是否需要修正格式
			if strings.Contains(originalValue, ".") && !strings.Contains(jsonContent, fmt.Sprintf(`"leverage": %s`, originalValue)) {
				// 确保浮点数格式正确
				if parsed, err := strconv.ParseFloat(originalValue, 64); err == nil {
					correctedValue := fmt.Sprintf("%.1f", parsed)
					oldPattern := fmt.Sprintf(`"leverage"\s*:\s*%s`, originalValue)
					newPattern := fmt.Sprintf(`"leverage": %s`, correctedValue)
					jsonContent = regexp.MustCompile(oldPattern).ReplaceAllString(jsonContent, newPattern)

					errors = append(errors, &DecisionError{
						Type:         ErrorTypeLeverageTypeMismatch,
						Field:        "leverage",
						OriginalVal:  originalValue,
						CorrectedVal: correctedValue,
						Message:      fmt.Sprintf("杠杆值格式已自动修正: %s → %s", originalValue, correctedValue),
						Suggestion:   "AI应输出标准JSON数字格式",
					})
				}
			}
		}
	}

	// 2. 修复缺失的引号问题
	jsonContent = h.fixMissingQuotes(jsonContent)

	// 3. 验证JSON结构完整性
	if !h.isValidJSON(jsonContent) {
		errors = append(errors, &DecisionError{
			Type:       ErrorTypeInvalidJSON,
			Field:      "json",
			Message:    "JSON格式无效",
			Suggestion: "请检查AI输出的JSON格式是否正确",
		})
	}

	return jsonContent, errors
}

// fixMissingQuotes 修复缺失的引号
func (h *DecisionErrorHandler) fixMissingQuotes(jsonStr string) string {
	// 替换中文引号
	jsonStr = strings.ReplaceAll(jsonStr, "“", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "”", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "‘", "'")
	jsonStr = strings.ReplaceAll(jsonStr, "’", "'")

	// 修复常见的JSON格式问题
	// 确保键名有引号
	keyPattern := regexp.MustCompile(`([{,]\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
	jsonStr = keyPattern.ReplaceAllString(jsonStr, `$1"$2":`)

	return jsonStr
}

// isValidJSON 验证JSON有效性
func (h *DecisionErrorHandler) isValidJSON(jsonStr string) bool {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp) == nil
}

// ParseDecisionsWithAutoFix 带自动修正的决策解析
func (h *DecisionErrorHandler) ParseDecisionsWithAutoFix(jsonContent string) ([]*DecisionEnhanced, []*DecisionError, error) {
	// 第一步：预处理JSON
	processedJSON, preprocessErrors := h.PreprocessJSON(jsonContent)

	// 第二步：尝试解析增强版决策
	var decisions []*DecisionEnhanced
	err := json.Unmarshal([]byte(processedJSON), &decisions)

	if err != nil {
		// 如果解析失败，尝试解析原始Decision结构
		var originalDecisions []Decision
		if originalErr := json.Unmarshal([]byte(processedJSON), &originalDecisions); originalErr == nil {
			// 转换为增强版结构
			decisions = h.convertToEnhanced(originalDecisions)
			// 添加转换警告
			preprocessErrors = append(preprocessErrors, &DecisionError{
				Type:       ErrorTypeLeverageTypeMismatch,
				Field:      "structure",
				Message:    "使用兼容模式解析决策",
				Suggestion: "建议更新AI输出格式以支持浮点数杠杆",
			})
		} else {
			// 完全失败
			return nil, preprocessErrors, fmt.Errorf("JSON解析失败: %w", err)
		}
	}

	// 第三步：验证和修正决策数据
	validationErrors := h.validateAndFixDecisions(decisions)
	allErrors := append(preprocessErrors, validationErrors...)

	return decisions, allErrors, nil
}

// convertToEnhanced 将原始决策转换为增强版
func (h *DecisionErrorHandler) convertToEnhanced(original []Decision) []*DecisionEnhanced {
	enhanced := make([]*DecisionEnhanced, len(original))
	for i, orig := range original {
		enhanced[i] = &DecisionEnhanced{
			Symbol:          orig.Symbol,
			Action:          orig.Action,
			Leverage:        float64(orig.Leverage), // int转float64
			PositionSizeUSD: orig.PositionSizeUSD,
			StopLoss:        orig.StopLoss,
			TakeProfit:      orig.TakeProfit,
			NewStopLoss:     orig.NewStopLoss,
			NewTakeProfit:   orig.NewTakeProfit,
			ClosePercentage: orig.ClosePercentage,
			MaxDrawdown:     orig.MaxDrawdown,
			MinTargetProfit: orig.MinTargetProfit,
			MaxPositionUSD:  orig.MaxPositionUSD,
			MaxDailyLoss:    orig.MaxDailyLoss,
			TimeInForce:     orig.TimeInForce,
			Confidence:      orig.Confidence,
			RiskUSD:         orig.RiskUSD,
			Reasoning:       orig.Reasoning,
		}
	}
	return enhanced
}

// validateAndFixDecisions 验证并修正决策数据
func (h *DecisionErrorHandler) validateAndFixDecisions(decisions []*DecisionEnhanced) []*DecisionError {
	errors := make([]*DecisionError, 0)

	for i, decision := range decisions {
		// 验证杠杆值
		if decision.Leverage <= 0 {
			correctedLeverage := 1.0
			errors = append(errors, &DecisionError{
				Type:         ErrorTypeValueOutOfRange,
				Field:        fmt.Sprintf("decision[%d].leverage", i),
				OriginalVal:  decision.Leverage,
				CorrectedVal: correctedLeverage,
				Message:      fmt.Sprintf("杠杆值无效 (%.2f)，已修正为 %.1f", decision.Leverage, correctedLeverage),
				Suggestion:   "杠杆值应为正数",
			})
			decision.Leverage = correctedLeverage
		} else if decision.Leverage > 125 {
			correctedLeverage := 125.0
			errors = append(errors, &DecisionError{
				Type:         ErrorTypeValueOutOfRange,
				Field:        fmt.Sprintf("decision[%d].leverage", i),
				OriginalVal:  decision.Leverage,
				CorrectedVal: correctedLeverage,
				Message:      fmt.Sprintf("杠杆值过高 (%.2f)，已限制为 %.1f", decision.Leverage, correctedLeverage),
				Suggestion:   "最大杠杆限制为125倍",
			})
			decision.Leverage = correctedLeverage
		}

		// 验证其他字段...
		if decision.Confidence < 0 || decision.Confidence > 100 {
			correctedConfidence := 80
			if decision.Confidence < 0 {
				correctedConfidence = 0
			} else if decision.Confidence > 100 {
				correctedConfidence = 100
			}
			errors = append(errors, &DecisionError{
				Type:         ErrorTypeValueOutOfRange,
				Field:        fmt.Sprintf("decision[%d].confidence", i),
				OriginalVal:  decision.Confidence,
				CorrectedVal: correctedConfidence,
				Message:      fmt.Sprintf("置信度超出范围 (%d)，已修正为 %d", decision.Confidence, correctedConfidence),
				Suggestion:   "置信度应在0-100范围内",
			})
			decision.Confidence = correctedConfidence
		}
	}

	return errors
}

// GenerateUserFriendlyReport 生成用户友好的错误报告
func (h *DecisionErrorHandler) GenerateUserFriendlyReport(errors []*DecisionError) string {
	if len(errors) == 0 {
		return "✅ 决策解析成功，无错误"
	}

	var report strings.Builder
	report.WriteString("⚠️  决策处理报告:\n")

	criticalErrors := 0
	warningErrors := 0

	for _, err := range errors {
		switch err.Type {
		case ErrorTypeInvalidJSON:
			criticalErrors++
			report.WriteString(fmt.Sprintf("🔴 严重错误: %s\n", err.Message))
			if err.Suggestion != "" {
				report.WriteString(fmt.Sprintf("   💡 建议: %s\n", err.Suggestion))
			}
		case ErrorTypeLeverageTypeMismatch, ErrorTypeValueOutOfRange:
			warningErrors++
			report.WriteString(fmt.Sprintf("🟡 警告: %s\n", err.Message))
			if err.Suggestion != "" {
				report.WriteString(fmt.Sprintf("   💡 建议: %s\n", err.Suggestion))
			}
		default:
			report.WriteString(fmt.Sprintf("🔵 信息: %s\n", err.Message))
		}
	}

	report.WriteString(fmt.Sprintf("\n📊 统计: %d个严重错误, %d个警告, %d个信息提示",
		criticalErrors, warningErrors, len(errors)-criticalErrors-warningErrors))

	return report.String()
}

// GetTechnicalDetails 获取技术详情
func (h *DecisionErrorHandler) GetTechnicalDetails(errors []*DecisionError) string {
	if len(errors) == 0 {
		return "无技术错误详情"
	}

	var details strings.Builder
	details.WriteString("=== 技术错误详情 ===\n")

	for i, err := range errors {
		details.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, err.Field, err.Message))
		details.WriteString(fmt.Sprintf("   类型: %v\n", err.Type))
		details.WriteString(fmt.Sprintf("   原始值: %v\n", err.OriginalVal))
		details.WriteString(fmt.Sprintf("   修正值: %v\n", err.CorrectedVal))
		if err.Suggestion != "" {
			details.WriteString(fmt.Sprintf("   建议: %s\n", err.Suggestion))
		}
		details.WriteString("\n")
	}

	return details.String()
}
