package trader

import (
	"fmt"
	"strconv"
	"strings"
)

// SafeParseFloat 安全解析浮点数，处理空字符串情况
func SafeParseFloat(value interface{}, fieldName string) (float64, error) {
	switch v := value.(type) {
	case string:
		// 检查空字符串
		if strings.TrimSpace(v) == "" {
			return 0, fmt.Errorf("field '%s' contains empty string", fieldName)
		}
		// 尝试解析
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse field '%s' value '%s': %w", fieldName, v, err)
		}
		return parsed, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case nil:
		return 0, fmt.Errorf("field '%s' is nil", fieldName)
	default:
		return 0, fmt.Errorf("field '%s' has unsupported type: %T", fieldName, v)
	}
}

// ValidateAndParsePrice 验证并解析价格数据
func ValidateAndParsePrice(price interface{}, symbol string) (float64, error) {
	priceFloat, err := SafeParseFloat(price, "price")
	if err != nil {
		return 0, fmt.Errorf("failed to get market price for %s: %w", symbol, err)
	}

	// 验证价格合理性
	if priceFloat <= 0 {
		return 0, fmt.Errorf("invalid price for %s: %.8f (must be positive)", symbol, priceFloat)
	}

	if priceFloat > 1000000 { // 100万美元上限检查
		return 0, fmt.Errorf("suspiciously high price for %s: %.8f", symbol, priceFloat)
	}

	return priceFloat, nil
}

// ValidateAndParseQuantity 验证并解析数量数据
func ValidateAndParseQuantity(quantity interface{}, symbol string) (float64, error) {
	qtyFloat, err := SafeParseFloat(quantity, "quantity")
	if err != nil {
		return 0, fmt.Errorf("failed to parse quantity for %s: %w", symbol, err)
	}

	// 验证数量合理性
	if qtyFloat < 0 {
		return 0, fmt.Errorf("negative quantity for %s: %.8f", symbol, qtyFloat)
	}

	if qtyFloat > 10000000 { // 1000万上限检查
		return 0, fmt.Errorf("suspiciously large quantity for %s: %.8f", symbol, qtyFloat)
	}

	return qtyFloat, nil
}
