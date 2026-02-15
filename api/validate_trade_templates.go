package api

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TradeTemplateValidator 交易模板验证器
type TradeTemplateValidator struct {
	logger *logrus.Logger
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Info     []string `json:"info,omitempty"`
}

// NewTradeTemplateValidator 创建验证器
func NewTradeTemplateValidator(logger *logrus.Logger) *TradeTemplateValidator {
	return &TradeTemplateValidator{logger: logger}
}

// ValidateTradeDecision 验证单个交易决策
func (v *TradeTemplateValidator) ValidateTradeDecision(decision map[string]interface{}, currentPrice float64) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Warnings: []string{},
		Errors:   []string{},
		Info:     []string{},
	}

	// 0. 基础数据清洗和类型检查
	if err := v.sanitizeAndValidateBasicTypes(decision, &result); err != nil {
		result.Valid = false
		return result
	}

	// 1. 检查必需字段
	symbol, ok := decision["symbol"].(string)
	if !ok || symbol == "" {
		result.Errors = append(result.Errors, "❌ 缺少 symbol 字段或格式错误")
		result.Valid = false
		return result
	}
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	decision["symbol"] = symbol // 清洗后写回

	action, ok := decision["action"].(string)
	if !ok || action == "" {
		result.Errors = append(result.Errors, "❌ 缺少 action 字段或格式错误")
		result.Valid = false
		return result
	}
	action = strings.TrimSpace(strings.ToLower(action))
	decision["action"] = action // 清洗后写回

	// 验证 symbol 格式（必须是 XXXUSDT 格式）
	if !strings.HasSuffix(symbol, "USDT") && !strings.HasSuffix(symbol, "BUSD") {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ Symbol 格式错误: %s (必须以 USDT 或 BUSD 结尾)", symbol))
		result.Valid = false
		return result
	}

	// 检查价格是否有效
	if currentPrice <= 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 当前价格无效: %.2f (必须 > 0)", currentPrice))
		result.Valid = false
		return result
	}

	result.Info = append(result.Info, fmt.Sprintf("📊 Symbol: %s | Action: %s | 当前价格: %.2f", symbol, action, currentPrice))

	// 2. 根据不同动作类型验证
	switch action {
	case "open_long", "open_short":
		v.validateOpenPosition(decision, currentPrice, &result)
	case "close_long", "close_short":
		v.validateClosePosition(decision, &result)
	case "update_stop_loss":
		v.validateUpdateStopLoss(decision, currentPrice, &result)
	case "update_take_profit":
		v.validateUpdateTakeProfit(decision, currentPrice, &result)
	case "partial_close":
		v.validatePartialClose(decision, &result)
	case "trailing_stop":
		v.validateTrailingStop(decision, currentPrice, &result)
	case "dynamic_take_profit":
		v.validateDynamicTakeProfit(decision, &result)
	case "dynamic_stop_loss":
		v.validateDynamicStopLoss(decision, &result)
	case "oco_order":
		v.validateOCOOrder(decision, currentPrice, &result)
	case "bracket_order":
		v.validateBracketOrder(decision, currentPrice, &result)
	case "add_to_position":
		v.validateAddToPosition(decision, &result)
	case "hold", "wait":
		result.Info = append(result.Info, "✅ 观望动作，无需特殊验证")
	default:
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 未知动作类型: %s", action))
		result.Valid = false
		return result
	}

	// 3. 检查 confidence 字段
	if confidence, ok := decision["confidence"].(float64); ok {
		if math.IsNaN(confidence) || math.IsInf(confidence, 0) {
			result.Errors = append(result.Errors, "❌ Confidence 值无效 (NaN 或 Inf)")
			result.Valid = false
		} else if confidence < 0 || confidence > 100 {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ Confidence 必须在 0-100 之间，当前: %.0f", confidence))
			result.Valid = false
		} else if confidence < 70 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ Confidence 较低 (%.0f)，建议 ≥70", confidence))
		} else {
			result.Info = append(result.Info, fmt.Sprintf("✅ Confidence: %.0f (合理)", confidence))
		}
	} else {
		result.Warnings = append(result.Warnings, "⚠️ 缺少 confidence 字段")
	}

	return result
}

// sanitizeAndValidateBasicTypes 清洗和验证基础数据类型
func (v *TradeTemplateValidator) sanitizeAndValidateBasicTypes(decision map[string]interface{}, result *ValidationResult) error {
	// 检查所有数值字段，确保不是 NaN 或 Inf
	numericFields := []string{
		"leverage", "position_size_usd", "stop_loss", "take_profit", "confidence", "risk_usd",
		"new_stop_loss", "new_take_profit", "close_percentage", "trail_percentage", "callback_rate",
		"activation_price", "target_roi", "max_roi", "time_limit_hours", "initial_risk_percent",
		"profit_threshold_percent", "additional_position_size_usd",
	}

	for _, field := range numericFields {
		if val, ok := decision[field]; ok {
			// 尝试转换为 float64
			var floatVal float64
			switch v := val.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			case int64:
				floatVal = float64(v)
			case string:
				// 尝试解析字符串数字
				parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 的值 '%v' 不是有效数字", field, v))
					return fmt.Errorf("invalid numeric field: %s", field)
				}
				floatVal = parsed
				decision[field] = floatVal // 转换后写回
			default:
				result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 的类型错误: %T", field, v))
				return fmt.Errorf("invalid field type: %s", field)
			}

			// 检查是否为 NaN 或 Inf
			if math.IsNaN(floatVal) {
				result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 的值为 NaN (非数字)", field))
				return fmt.Errorf("NaN value in field: %s", field)
			}
			if math.IsInf(floatVal, 0) {
				result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 的值为 Infinity (无穷大)", field))
				return fmt.Errorf("Inf value in field: %s", field)
			}

			// 检查是否为负数（某些字段不允许负数）
			noNegativeFields := []string{"leverage", "position_size_usd", "confidence", "close_percentage", "callback_rate", "additional_position_size_usd"}
			for _, noNegField := range noNegativeFields {
				if field == noNegField && floatVal < 0 {
					result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 不能为负数，当前: %.2f", field, floatVal))
					return fmt.Errorf("negative value in field: %s", field)
				}
			}

			// 检查是否为 0（某些字段不允许为 0）
			noZeroFields := []string{"leverage", "position_size_usd", "stop_loss", "take_profit", "new_stop_loss", "new_take_profit"}
			for _, noZeroField := range noZeroFields {
				if field == noZeroField && floatVal == 0 {
					result.Errors = append(result.Errors, fmt.Sprintf("【风控】 ❌ 字段 %s 不能为 0", field))
					return fmt.Errorf("zero value in field: %s", field)
				}
			}
		}
	}

	return nil
}

// validateOpenPosition 验证开仓动作
func (v *TradeTemplateValidator) validateOpenPosition(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	isLong := decision["action"].(string) == "open_long"

	// 检查杠杆
	leverage, ok := decision["leverage"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 leverage 字段")
		result.Valid = false
		return
	}

	symbol := decision["symbol"].(string)
	isBTCETH := strings.HasPrefix(symbol, "BTC") || strings.HasPrefix(symbol, "ETH")

	maxLeverage := 5.0
	if isBTCETH {
		maxLeverage = 20.0
	}

	if leverage < 1 || leverage > maxLeverage {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 杠杆超出范围 (1-%.0f)，当前: %.0f", maxLeverage, leverage))
		result.Valid = false
	} else if leverage > 10 && !isBTCETH {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 山寨币杠杆较高 (%.0fx)，风险极大！建议 ≤5x", leverage))
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 杠杆: %.0fx (合理)", leverage))
	}

	// 检查仓位大小
	positionSizeUSD, ok := decision["position_size_usd"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 position_size_usd 字段")
		result.Valid = false
		return
	}

	minSize := 12.0
	if isBTCETH {
		minSize = 60.0
	}

	// 检查过大的仓位（可能是输入错误）
	maxReasonableSize := 100000.0 // 10万 USDT
	if positionSizeUSD > maxReasonableSize {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 仓位过大 (%.2f USDT)，超出合理范围 (最大 %.0f)，请检查是否输入错误", positionSizeUSD, maxReasonableSize))
		result.Valid = false
		return
	}

	if positionSizeUSD < minSize {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 仓位过小 (最小 %.0f USDT)，当前: %.2f", minSize, positionSizeUSD))
		result.Valid = false
	} else if positionSizeUSD > 50000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 仓位较大 (%.2f USDT)，请确认风险承受能力", positionSizeUSD))
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 仓位大小: %.2f USDT", positionSizeUSD))
	}

	// 检查止损止盈
	stopLoss, hasStopLoss := decision["stop_loss"].(float64)
	takeProfit, hasTakeProfit := decision["take_profit"].(float64)

	if !hasStopLoss || !hasTakeProfit {
		result.Errors = append(result.Errors, "❌ 开仓必须设置 stop_loss 和 take_profit")
		result.Valid = false
		return
	}

	// 检查异常价格（可能是输入错误）- 价格不能相差超过200%
	priceRatio := math.Max(stopLoss, takeProfit) / math.Min(stopLoss, takeProfit)
	if priceRatio > 3.0 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 止损(%.2f)与止盈(%.2f)价格差异过大(%.1f倍)，请检查输入", stopLoss, takeProfit, priceRatio))
		result.Valid = false
		return
	}

	// 检查价格是否偏离当前价过远（可能输入错误）
	if stopLoss > currentPrice*2 || takeProfit > currentPrice*2 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 止损或止盈价格异常 (SL:%.2f, TP:%.2f, 当前:%.2f)，超出200%%范围，请检查输入", stopLoss, takeProfit, currentPrice))
		result.Valid = false
		return
	}

	if stopLoss < currentPrice*0.1 || takeProfit < currentPrice*0.1 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 止损或止盈价格过低 (SL:%.2f, TP:%.2f, 当前:%.2f)，小于当前价10%%，请检查输入", stopLoss, takeProfit, currentPrice))
		result.Valid = false
		return
	}

	// 验证止损止盈方向
	if isLong {
		// 做多：止损 < 当前价 < 止盈
		if stopLoss >= currentPrice {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ 多单止损价 (%.2f) 必须低于当前价 (%.2f)", stopLoss, currentPrice))
			result.Valid = false
		}
		if takeProfit <= currentPrice {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ 多单止盈价 (%.2f) 必须高于当前价 (%.2f)", takeProfit, currentPrice))
			result.Valid = false
		}
	} else {
		// 做空：止盈 < 当前价 < 止损
		if stopLoss <= currentPrice {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ 空单止损价 (%.2f) 必须高于当前价 (%.2f)", stopLoss, currentPrice))
			result.Valid = false
		}
		if takeProfit >= currentPrice {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ 空单止盈价 (%.2f) 必须低于当前价 (%.2f)", takeProfit, currentPrice))
			result.Valid = false
		}
	}

	// 计算风险回报比
	if result.Valid {
		slDistance := math.Abs(currentPrice - stopLoss)
		tpDistance := math.Abs(takeProfit - currentPrice)

		// 防止除以零
		if slDistance == 0 {
			result.Errors = append(result.Errors, "❌ 止损价格与当前价相同，无风险保护")
			result.Valid = false
			return
		}

		rrRatio := tpDistance / slDistance

		if rrRatio < 1.0 {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ 风险回报比过低 (1:%.2f)，亏损风险大于盈利预期，不建议交易", rrRatio))
			result.Valid = false
		} else if rrRatio < 1.5 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 风险回报比偏低 (1:%.2f)，建议 ≥1.5", rrRatio))
		} else {
			result.Info = append(result.Info, fmt.Sprintf("✅ 风险回报比: 1:%.2f (优秀)", rrRatio))
		}

		slPercent := (slDistance / currentPrice) * 100
		tpPercent := (tpDistance / currentPrice) * 100

		// 检查过小的止损止盈距离
		if slPercent < 0.3 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止损距离过小 (%.2f%%)，可能因波动而频繁触发", slPercent))
		}
		if tpPercent < 0.5 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止盈距离过小 (%.2f%%)，盈利空间有限", tpPercent))
		}

		// 检查过大的止损距离
		if slPercent > 15 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止损距离过大 (%.2f%%)，单笔风险过高", slPercent))
		}

		result.Info = append(result.Info, fmt.Sprintf("📊 止损: %.2f (%.2f%%) | 止盈: %.2f (%.2f%%)", stopLoss, slPercent, takeProfit, tpPercent))
	}

	// 检查 risk_usd
	if riskUSD, ok := decision["risk_usd"].(float64); ok {
		expectedRisk := positionSizeUSD * leverage * (math.Abs(currentPrice-stopLoss) / currentPrice)
		deviation := math.Abs(riskUSD - expectedRisk)
		deviationPercent := (deviation / expectedRisk) * 100

		if deviationPercent > 50 {
			result.Errors = append(result.Errors, fmt.Sprintf("❌ risk_usd (%.2f) 与计算值 (%.2f) 相差过大 (%.1f%%)，请检查计算", riskUSD, expectedRisk, deviationPercent))
			result.Valid = false
		} else if deviationPercent > 20 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ risk_usd (%.2f) 与计算值 (%.2f) 相差 %.1f%%", riskUSD, expectedRisk, deviationPercent))
		}
	}
}

// validateClosePosition 验证平仓动作
func (v *TradeTemplateValidator) validateClosePosition(decision map[string]interface{}, result *ValidationResult) {
	result.Info = append(result.Info, "✅ 平仓动作：将关闭所有相关仓位")
	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateUpdateStopLoss 验证更新止损
func (v *TradeTemplateValidator) validateUpdateStopLoss(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	newStopLoss, ok := decision["new_stop_loss"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 new_stop_loss 字段")
		result.Valid = false
		return
	}

	// 检查异常价格
	if newStopLoss > currentPrice*2 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 新止损价格 (%.2f) 异常高，超出当前价 (%.2f) 200%%，请检查输入", newStopLoss, currentPrice))
		result.Valid = false
		return
	}

	if newStopLoss < currentPrice*0.1 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 新止损价格 (%.2f) 过低，小于当前价 (%.2f) 10%%，请检查输入", newStopLoss, currentPrice))
		result.Valid = false
		return
	}

	// 假设是多单（实际应该从持仓查询）
	// 这里只做基本的合理性检查
	slDistance := math.Abs(currentPrice - newStopLoss)
	slPercent := (slDistance / currentPrice) * 100

	if slPercent < 0.3 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止损距离过近 (%.2f%%)，可能因波动而频繁触发", slPercent))
	} else if slPercent > 15 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止损距离过远 (%.2f%%)，风险较大", slPercent))
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 新止损: %.2f (距当前价 %.2f%%)", newStopLoss, slPercent))
	}

	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateUpdateTakeProfit 验证更新止盈
func (v *TradeTemplateValidator) validateUpdateTakeProfit(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	newTakeProfit, ok := decision["new_take_profit"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 new_take_profit 字段")
		result.Valid = false
		return
	}

	// 检查异常价格
	if newTakeProfit > currentPrice*3 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 新止盈价格 (%.2f) 异常高，超出当前价 (%.2f) 300%%，请检查输入", newTakeProfit, currentPrice))
		result.Valid = false
		return
	}

	if newTakeProfit < currentPrice*0.1 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 新止盈价格 (%.2f) 过低，小于当前价 (%.2f) 10%%，请检查输入", newTakeProfit, currentPrice))
		result.Valid = false
		return
	}

	tpDistance := math.Abs(newTakeProfit - currentPrice)
	tpPercent := (tpDistance / currentPrice) * 100

	if tpPercent < 0.5 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止盈目标过近 (%.2f%%)，收益有限", tpPercent))
	} else if tpPercent > 50 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 止盈目标过远 (%.2f%%)，可能难以达成", tpPercent))
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 新止盈: %.2f (距当前价 %.2f%%)", newTakeProfit, tpPercent))
	}

	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validatePartialClose 验证部分平仓
func (v *TradeTemplateValidator) validatePartialClose(decision map[string]interface{}, result *ValidationResult) {
	closePercentage, ok := decision["close_percentage"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 close_percentage 字段")
		result.Valid = false
		return
	}

	if closePercentage <= 0 || closePercentage > 100 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ close_percentage 必须在 1-100 之间，当前: %.0f", closePercentage))
		result.Valid = false
	} else if closePercentage < 25 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 平仓比例较小 (%.0f%%)，可能影响不大", closePercentage))
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 部分平仓: %.0f%% 的仓位", closePercentage))
	}

	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateTrailingStop 验证追踪止损
func (v *TradeTemplateValidator) validateTrailingStop(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	callbackRate, ok := decision["callback_rate"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 callback_rate 字段")
		result.Valid = false
		return
	}

	// 币安API要求：callback_rate 范围 [0.1, 10]，其中 1.0 = 1%
	if callbackRate < 0.1 || callbackRate > 10.0 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ callback_rate 必须在 0.1-10 之间 (1.0=1%%)，当前: %.1f", callbackRate))
		result.Valid = false
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 追踪回调率: %.1f%% (格式正确)", callbackRate))
	}

	// 检查激活价格（可选）
	if activationPrice, ok := decision["activation_price"].(float64); ok {
		distance := math.Abs(activationPrice - currentPrice)
		distancePercent := (distance / currentPrice) * 100
		result.Info = append(result.Info, fmt.Sprintf("📊 激活价格: %.2f (距当前价 %.2f%%)", activationPrice, distancePercent))

		if distancePercent > 20 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 激活价格与当前价相差较大 (%.2f%%)，请确认", distancePercent))
		}
	} else {
		result.Info = append(result.Info, "📊 未设置激活价格，将使用当前市场价")
	}

	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateDynamicTakeProfit 验证动态止盈
func (v *TradeTemplateValidator) validateDynamicTakeProfit(decision map[string]interface{}, result *ValidationResult) {
	targetROI, hasTarget := decision["target_roi"].(float64)
	maxROI, hasMax := decision["max_roi"].(float64)
	timeLimitHours, hasTime := decision["time_limit_hours"].(float64)

	if !hasTarget {
		result.Errors = append(result.Errors, "❌ 缺少 target_roi 字段")
		result.Valid = false
	}
	if !hasMax {
		result.Errors = append(result.Errors, "❌ 缺少 max_roi 字段")
		result.Valid = false
	}
	if !hasTime {
		result.Errors = append(result.Errors, "❌ 缺少 time_limit_hours 字段")
		result.Valid = false
	}

	if !result.Valid {
		return
	}

	if maxROI <= targetROI {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ max_roi (%.1f) 必须大于 target_roi (%.1f)", maxROI, targetROI))
		result.Valid = false
	}

	if targetROI < 2.0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ target_roi 较低 (%.1f%%)，建议 ≥2%%", targetROI))
	}

	if timeLimitHours < 1 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 时间限制过短 (%.1f小时)", timeLimitHours))
	}

	result.Info = append(result.Info, fmt.Sprintf("✅ 动态止盈: 目标%.1f%% | 最大%.1f%% | 时限%.1f小时", targetROI, maxROI, timeLimitHours))
	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateDynamicStopLoss 验证动态止损
func (v *TradeTemplateValidator) validateDynamicStopLoss(decision map[string]interface{}, result *ValidationResult) {
	initialRiskPercent, hasInitial := decision["initial_risk_percent"].(float64)
	profitThreshold, hasThreshold := decision["profit_threshold_percent"].(float64)

	if !hasInitial {
		result.Errors = append(result.Errors, "❌ 缺少 initial_risk_percent 字段")
		result.Valid = false
	}
	if !hasThreshold {
		result.Errors = append(result.Errors, "❌ 缺少 profit_threshold_percent 字段")
		result.Valid = false
	}

	if !result.Valid {
		return
	}

	if initialRiskPercent < 0.5 || initialRiskPercent > 10 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 初始风险 (%.1f%%) 超出常规范围 (0.5-10%%)", initialRiskPercent))
	}

	if profitThreshold < 1.0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 盈利阈值过低 (%.1f%%)，建议 ≥1%%", profitThreshold))
	}

	result.Info = append(result.Info, fmt.Sprintf("✅ 动态止损: 初始风险%.1f%% | 盈利阈值%.1f%%", initialRiskPercent, profitThreshold))
	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// validateOCOOrder 验证OCO订单
func (v *TradeTemplateValidator) validateOCOOrder(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	stopLoss, hasStopLoss := decision["stop_loss"].(float64)
	takeProfit, hasTakeProfit := decision["take_profit"].(float64)

	if !hasStopLoss || !hasTakeProfit {
		result.Errors = append(result.Errors, "❌ OCO订单必须同时设置 stop_loss 和 take_profit")
		result.Valid = false
		return
	}

	// 检查价格合理性
	if stopLoss == takeProfit {
		result.Errors = append(result.Errors, "❌ 止损价和止盈价不能相同")
		result.Valid = false
	}

	result.Info = append(result.Info, fmt.Sprintf("✅ OCO订单: 止损%.2f | 止盈%.2f", stopLoss, takeProfit))
	result.Warnings = append(result.Warnings, "⚠️ 提示：OCO仅适用于已有持仓，新开仓不支持")
}

// validateBracketOrder 验证括号订单
func (v *TradeTemplateValidator) validateBracketOrder(decision map[string]interface{}, currentPrice float64, result *ValidationResult) {
	// 括号订单本质上是 开仓 + OCO
	v.validateOpenPosition(decision, currentPrice, result)

	if result.Valid {
		result.Info = append(result.Info, "✅ 括号订单: 将同时创建开仓和止损止盈订单")
	}
}

// validateAddToPosition 验证加仓
func (v *TradeTemplateValidator) validateAddToPosition(decision map[string]interface{}, result *ValidationResult) {
	additionalSize, ok := decision["additional_position_size_usd"].(float64)
	if !ok {
		result.Errors = append(result.Errors, "❌ 缺少 additional_position_size_usd 字段")
		result.Valid = false
		return
	}

	// 检查过大的加仓金额（可能是输入错误）
	maxAdditionalSize := 50000.0 // 5万 USDT
	if additionalSize > maxAdditionalSize {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 加仓金额过大 (%.2f USDT)，超出合理范围 (最大 %.0f)，请检查输入", additionalSize, maxAdditionalSize))
		result.Valid = false
		return
	}

	if additionalSize < 12 {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ 加仓金额过小 (最小12 USDT)，当前: %.2f", additionalSize))
		result.Valid = false
	} else if additionalSize > 20000 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("⚠️ 加仓金额较大 (%.2f USDT)，请确认风险", additionalSize))
	}

	addType, ok := decision["add_position_type"].(string)
	if !ok || (strings.ToUpper(addType) != "LONG" && strings.ToUpper(addType) != "SHORT") {
		result.Errors = append(result.Errors, fmt.Sprintf("❌ add_position_type 必须是 'long' 或 'short'，当前: %v", addType))
		result.Valid = false
	} else {
		result.Info = append(result.Info, fmt.Sprintf("✅ 加仓: %.2f USDT 到 %s 仓位", additionalSize, strings.ToUpper(addType)))
	}

	result.Warnings = append(result.Warnings, "⚠️ 警告：仅在盈利仓位时加仓，永远不要追亏损！")
	result.Warnings = append(result.Warnings, "⚠️ 提示：确保有对应的持仓存在")
}

// HandleValidateTemplate HTTP处理函数
func (s *Server) HandleValidateTemplate(c *gin.Context) {
	var decisions []map[string]interface{}
	if err := c.ShouldBindJSON(&decisions); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format: " + err.Error()})
		return
	}

	// 模拟获取当前价格（实际应该从市场API获取）
	mockPrices := map[string]float64{
		"BTCUSDT": 43560.42,
		"ETHUSDT": 2635.80,
		"SOLUSDT": 98.45,
		"BNBUSDT": 300.50,
		"ADAUSDT": 0.52,
		"XRPUSDT": 0.55,
	}

	validator := NewTradeTemplateValidator(s.logger)
	results := make([]map[string]interface{}, 0, len(decisions))

	s.logger.Info("========================================")
	s.logger.Info("🔍 开始验证交易模板")
	s.logger.Info("========================================")

	for i, decision := range decisions {
		s.logger.Infof("\n--- 验证第 %d 个决策 ---", i+1)

		symbol, _ := decision["symbol"].(string)
		currentPrice := mockPrices[symbol]
		if currentPrice == 0 {
			currentPrice = 100.0 // 默认价格
		}

		validationResult := validator.ValidateTradeDecision(decision, currentPrice)

		// 输出详细日志
		s.logger.Infof("📊 决策内容: %s", formatJSON(decision))

		for _, info := range validationResult.Info {
			s.logger.Info(info)
		}
		for _, warning := range validationResult.Warnings {
			s.logger.Warn(warning)
		}
		for _, err := range validationResult.Errors {
			s.logger.Error(err)
		}

		if validationResult.Valid {
			s.logger.Info("✅ 验证通过")
		} else {
			s.logger.Error("❌ 验证失败")
		}

		results = append(results, map[string]interface{}{
			"index":      i,
			"decision":   decision,
			"validation": validationResult,
		})
	}

	s.logger.Info("========================================")
	s.logger.Info("✅ 验证完成")
	s.logger.Info("========================================")

	c.JSON(200, gin.H{
		"total_decisions": len(decisions),
		"results":         results,
	})
}

// formatJSON 格式化JSON输出
func formatJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
