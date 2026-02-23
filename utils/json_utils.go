package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// FilterReasoningQuotes 专门过滤JSON字符串中reasoning字段的引号
func FilterReasoningQuotes(jsonStr string) (string, error) {
	// 首先解析原始JSON到map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 递归处理所有字段，找到reasoning字段并过滤其值
	filterReasoningInMap(data)

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

	// 递归处理所有字段，找到reasoning字段并替换引号
	replaceReasoningQuotesInMap(data)

	// 重新序列化为JSON字符串
	filteredBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(filteredBytes), nil
}

// replaceReasoningQuotesInMap 递归处理map中的reasoning字段，将英文引号替换成中文引号
func replaceReasoningQuotesInMap(data map[string]interface{}) {
	for key, value := range data {
		if strings.ToLower(key) == "reasoning" {
			// 如果值是字符串，替换其中的引号
			if str, ok := value.(string); ok {
				data[key] = replaceQuotesWithChinese(str)
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套的map
			replaceReasoningQuotesInMap(nestedMap)
		} else if slice, ok := value.([]interface{}); ok {
			// 处理数组中的元素
			for _, item := range slice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					replaceReasoningQuotesInMap(itemMap)
				}
			}
		}
	}
}

// replaceQuotesWithChinese 将字符串中的英文双引号替换成中文双引号
func replaceQuotesWithChinese(input string) string {
	// 将英文双引号 " 替换为中文双引号 "
	result := strings.ReplaceAll(input, "\"", "＂") // 全角引号

	// 如果还需要处理单引号，也可以替换
	result = strings.ReplaceAll(result, "'", "＇") // 全角单引号

	return result
}

// filterReasoningInMap 递归处理map中的reasoning字段
func filterReasoningInMap(data map[string]interface{}) {
	for key, value := range data {
		if strings.ToLower(key) == "reasoning" {
			// 如果值是字符串，过滤其中的引号
			if str, ok := value.(string); ok {
				data[key] = filterQuotesInString(str)
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套的map
			filterReasoningInMap(nestedMap)
		} else if slice, ok := value.([]interface{}); ok {
			// 处理数组中的元素
			for _, item := range slice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					filterReasoningInMap(itemMap)
				}
			}
		}
	}
}

// filterQuotesInString 过滤字符串中的引号
func filterQuotesInString(input string) string {
	// 替换双引号为圆括号，保留语义但避免JSON解析问题
	re := regexp.MustCompile(`"([^"]*)"`)

	// 先处理引号对
	result := re.ReplaceAllStringFunc(input, func(matched string) string {
		inner := matched[1 : len(matched)-1] // 移除外层引号
		return "(" + inner + ")"             // 用圆括号包围
	})

	// 还可以用其他方式处理，如直接移除或使用其他标点符号
	// 例如，将双引号替换成全角引号或其他符号
	result = strings.ReplaceAll(result, "\"", "''")
	result = strings.ReplaceAll(result, "'", "`")

	return result
}
