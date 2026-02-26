package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FilterReasoningQuotes 专门过滤JSON字符串中reasoning字段的引号
func FilterReasoningQuotes(jsonStr string) (string, error) {
	// 首先解析原始JSON到map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 条件性处理reasoning字段 - 只有在必要时才启用
	// 递归处理所有字段，找到reasoning字段并过滤其值
	// filterReasoningInMap(data) // 可选启用

	// 重新序列化为JSON字符串
	filteredBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(filteredBytes), nil
}

// ReplaceReasoningQuotesWithChinese 专门将JSON中reasoning字段的英文双引号替换成中文双引号
func ReplaceReasoningQuotesWithChinese(jsonStr string) (string, error) {
	// 首先解析原始JSON到map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 条件性处理reasoning字段 - 只有在必要时才启用
	// 递归处理所有字段，找到reasoning字段并替换引号
	// replaceReasoningQuotesInMap(data) // 可选启用

	// 重新序列化为JSON字符串
	filteredBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(filteredBytes), nil
}

// replaceReasoningQuotesInMap 递归处理map中的reasoning字段，将英文引号替换成中文引号
// func replaceReasoningQuotesInMap(data map[string]interface{}) {
// 	for key, value := range data {
// 		if strings.ToLower(key) == "reasoning" {
// 			// 如果值是字符串，替换其中的引号
// 			if str, ok := value.(string); ok {
// 				data[key] = replaceQuotesWithChinese(str)
// 			}
// 		} else if nestedMap, ok := value.(map[string]interface{}); ok {
// 			// 递归处理嵌套的map
// 			replaceReasoningQuotesInMap(nestedMap)
// 	} else if slice, ok := value.([]interface{}); ok {
// 			// 处理数组中的元素
// 			for _, item := range slice {
// 				if itemMap, ok := item.(map[string]interface{}); ok {
// 					replaceReasoningQuotesInMap(itemMap)
// 				}
// 			}
// 		}
// 	}
// }

// replaceQuotesWithChinese 将字符串中的英文双引号替换成中文双引号
func replaceQuotesWithChinese(input string) string {
	// 将英文双引号 " 替换为中文双引号 "
	result := strings.ReplaceAll(input, "\"", "＂") // 全角引号

	// 如果还需要处理单引号，也可以替换
	result = strings.ReplaceAll(result, "'", "＇") // 全角单引号

	return result
}

// filterReasoningInMap 递归处理map中的reasoning字段
// func filterReasoningInMap(data map[string]interface{}) {
// 	for key, value := range data {
// 		if strings.ToLower(key) == "reasoning" {
// 			// 修复reasoning字段值中的JSON格式问题
// 			if str, ok := value.(string); ok {
// 				data[key] = filterReasoningString(str)
// 			}
// 		} else if nestedMap, ok := value.(map[string]interface{}); ok {
// 			// 递归处理嵌套的map
// 			filterReasoningInMap(nestedMap)
// 	} else if slice, ok := value.([]interface{}); ok {
// 			// 处理数组中的元素
// 			for _, item := range slice {
// 				if itemMap, ok := item.(map[string]interface{}); ok {
// 					filterReasoningInMap(itemMap)
// 				}
// 			}
// 		}
// 	}
// }

// filterQuotesInString 过滤字符串中的引号
func filterQuotesInString(input string) string {
	// 修复字符串中的引号问题，避免JSON解析错误
	result := strings.ReplaceAll(input, "\"", "\\\"")
	result = strings.ReplaceAll(result, "'", "\\'")
	return result
}

// filterReasoningString 专门处理reasoning字段字符串的过滤
func filterReasoningString(input string) string {
	// 修复reasoning字段中的特殊字符，确保JSON兼容性
	result := strings.ReplaceAll(input, "\"", "\\\"")
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.ReplaceAll(result, "\r", " ")
	result = strings.ReplaceAll(result, "\t", " ")

	// 处理可能导致JSON解析错误的字符
	result = strings.ReplaceAll(result, "\\\"", "\"")

	return result
}

// ValidateDecisionJSON 验证决策JSON格式（周期专用）
func ValidateDecisionJSON(jsonStr string) error {
	// 检查JSON基本格式
	if !strings.HasPrefix(strings.TrimSpace(jsonStr), "[") &&
		!strings.HasPrefix(strings.TrimSpace(jsonStr), "{") {
		return fmt.Errorf("JSON必须以[或{开头")
	}

	// 检查是否包含必要的决策字段
	requiredFields := []string{"symbol", "action", "reasoning"}
	for _, field := range requiredFields {
		if !strings.Contains(strings.ToLower(jsonStr), field) {
			return fmt.Errorf("缺少必要字段: %s", field)
		}
	}

	// 注释掉reasoning字段检查，保持AI输出的原始状态
	// 检查reasoning字段格式
	// if strings.Contains(jsonStr, "reasoning'") && !strings.Contains(jsonStr, "reasoning\"") {
	// 	return fmt.Errorf("reasoning字段引号格式错误")
	// }

	return nil
}

// FixPeriodDecisionJSON 修复周期决策JSON格式
func FixPeriodDecisionJSON(jsonStr string) string {
	// 注释掉reasoning字段修复，保持AI输出的原始状态
	// 1. 修复reasoning字段的引号问题
	// jsonStr = strings.ReplaceAll(jsonStr, "reasoning': '", "reasoning\":\"")
	// jsonStr = strings.ReplaceAll(jsonStr, "',", \"\",")
	// jsonStr = strings.ReplaceAll(jsonStr, "'}", \"\"}")

	// 2. 修复Unicode转义字符
	jsonStr = strings.ReplaceAll(jsonStr, "\u003c", "<")
	jsonStr = strings.ReplaceAll(jsonStr, "\u003e", ">")
	jsonStr = strings.ReplaceAll(jsonStr, "\u0026", "&")
	jsonStr = strings.ReplaceAll(jsonStr, "\u0027", "'")
	jsonStr = strings.ReplaceAll(jsonStr, "\u0022", "\"")

	// 3. 处理空symbol字段
	if strings.Contains(jsonStr, `"symbol": ""`) {
		jsonStr = strings.ReplaceAll(jsonStr, `"symbol": ""`, `"symbol": "ALL"`)
	}

	// 4. 修复常见的JSON格式问题
	jsonStr = strings.ReplaceAll(jsonStr, "\"", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\"", "\"")

	// 5. 确保数组格式正确
	if strings.Contains(jsonStr, "[{") && !strings.Contains(jsonStr, "[") {
		jsonStr = "[" + jsonStr
	}

	return jsonStr
}

// ConvertStringNumbersToJSONNumbers将JSON中字符串格式的数字字段转换为真正的数字
func ConvertStringNumbersToJSONNumbers(jsonStr string) (string, error) {
	//首先解析原始JSON到interface{}切片
	var decisions []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &decisions); err != nil {
		// 如果解析为数组失败，尝试解析为单个对象
		var singleDecision map[string]interface{}
		if singleErr := json.Unmarshal([]byte(jsonStr), &singleDecision); singleErr == nil {
			decisions = []map[string]interface{}{singleDecision}
		} else {
			return "", fmt.Errorf("failed to parse JSON as array or object: %w", err)
		}
	}

	//定义需要转换的数字字段
	numberFields := []string{
		"leverage",
		"position_size_usd",
		"stop_loss",
		"take_profit",
		"confidence",
		"risk_usd",
		"close_percentage",
		"new_stop_loss",
		"new_take_profit",
		"trail_percentage",
		"activation_price",
		"target_roi",
		"max_roi",
		"time_limit_hours",
		"callback_rate",
		"additional_position_size_usd",
	}

	//处理每个决策对象
	for _, decision := range decisions {
		for _, field := range numberFields {
			if value, exists := decision[field]; exists {
				// 如果字段值是字符串，尝试转换为数字
				if strValue, ok := value.(string); ok {
					//尝解析为整数
					if intValue, err := strconv.Atoi(strings.TrimSpace(strValue)); err == nil {
						decision[field] = intValue
					} else {
						//尝解析为浮点数
						if floatValue, err := strconv.ParseFloat(strings.TrimSpace(strValue), 64); err == nil {
							decision[field] = floatValue
						}
						// 如果都失败了，保持原字符串值（让后续的JSON解析处理错误）
					}
				}
				// 如果已经是数字类型，保持不变
			}
		}
	}

	// 重新序列化为JSON字符串
	resultBytes, err := json.Marshal(decisions)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}
