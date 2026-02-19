package trader

import (
	"fmt"
	"strconv"
	"strings"
)

// PriceErrorType 定义价格获取错误的类型
type PriceErrorType int

const (
	PriceErrorUnknown PriceErrorType = iota
	PriceErrorInvalidSymbol
	PriceErrorNoData
	PriceErrorEmptyResponse
	PriceErrorParseFailed
	PriceErrorNetwork
	PriceErrorRateLimit
)

// PriceError 价格获取错误的结构化表示
type PriceError struct {
	Type    PriceErrorType
	Symbol  string
	Message string
	Details string
}

func (e *PriceError) Error() string {
	return fmt.Sprintf("price error for %s: %s (%s)", e.Symbol, e.Message, e.Details)
}

// IsInvalidSymbol 检查是否为无效交易对错误
func (e *PriceError) IsInvalidSymbol() bool {
	return e.Type == PriceErrorInvalidSymbol
}

// IsNoData 检查是否为无数据错误
func (e *PriceError) IsNoData() bool {
	return e.Type == PriceErrorNoData || e.Type == PriceErrorEmptyResponse
}

// IsNetworkError 检查是否为网络错误
func (e *PriceError) IsNetworkError() bool {
	return e.Type == PriceErrorNetwork || e.Type == PriceErrorRateLimit
}

// EnhancedPriceValidator 增强的价格验证器
type EnhancedPriceValidator struct {
	// 可以添加配置选项
	strictMode bool
}

// NewEnhancedPriceValidator 创建新的增强价格验证器
func NewEnhancedPriceValidator(strictMode bool) *EnhancedPriceValidator {
	return &EnhancedPriceValidator{
		strictMode: strictMode,
	}
}

// ValidatePriceString 验证价格字符串并提供详细的错误信息
func (v *EnhancedPriceValidator) ValidatePriceString(priceStr, symbol string) (float64, *PriceError) {
	// 检查空字符串
	if strings.TrimSpace(priceStr) == "" {
		return 0, &PriceError{
			Type:    PriceErrorEmptyResponse,
			Symbol:  symbol,
			Message: "empty price response",
			Details: "API returned empty price string",
		}
	}

	// 检查是否为"null"字符串（某些API可能返回这个）
	if strings.ToLower(strings.TrimSpace(priceStr)) == "null" {
		return 0, &PriceError{
			Type:    PriceErrorNoData,
			Symbol:  symbol,
			Message: "null price response",
			Details: "API returned null value instead of price",
		}
	}

	// 尝试解析数值
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, &PriceError{
			Type:    PriceErrorParseFailed,
			Symbol:  symbol,
			Message: "failed to parse price",
			Details: fmt.Sprintf("cannot convert '%s' to float64: %v", priceStr, err),
		}
	}

	// 验证价格合理性（严格模式下）
	if v.strictMode {
		if price <= 0 {
			return 0, &PriceError{
				Type:    PriceErrorInvalidSymbol,
				Symbol:  symbol,
				Message: "invalid price value",
				Details: fmt.Sprintf("price %.8f is not positive", price),
			}
		}

		// 检查是否为异常大的值
		if price > 1000000 { // 100万美元，可以根据需要调整
			return 0, &PriceError{
				Type:    PriceErrorInvalidSymbol,
				Symbol:  symbol,
				Message: "suspiciously high price",
				Details: fmt.Sprintf("price %.2f seems unusually high", price),
			}
		}
	}

	return price, nil
}

// CreateUserFriendlyErrorMessage 为用户创建友好的错误信息
func (v *EnhancedPriceValidator) CreateUserFriendlyErrorMessage(err *PriceError) string {
	switch err.Type {
	case PriceErrorInvalidSymbol:
		return fmt.Sprintf("❌ 交易对 '%s' 无效或不支持\n💡 建议：检查交易对名称是否正确，或该交易对可能未在交易所上线", err.Symbol)

	case PriceErrorNoData:
		return fmt.Sprintf("⚠️  暂无 '%s' 的价格数据\n💡 建议：该交易对可能暂停交易或刚上线尚未有数据", err.Symbol)

	case PriceErrorEmptyResponse:
		return fmt.Sprintf("⚠️  价格服务返回空数据\n💡 建议：稍后重试或检查网络连接")

	case PriceErrorParseFailed:
		return fmt.Sprintf("❌ 价格数据格式错误\n💡 建议：API响应格式异常，请联系技术支持")

	case PriceErrorNetwork:
		return fmt.Sprintf("🌐 网络连接问题\n💡 建议：检查网络连接后重试")

	case PriceErrorRateLimit:
		return fmt.Sprintf("⏳ API调用频率限制\n💡 建议：等待片刻后重试")

	default:
		return fmt.Sprintf("❌ 获取 '%s' 价格时发生未知错误\n💡 建议：请稍后重试或联系技术支持", err.Symbol)
	}
}

// LogErrorWithDetails 记录详细的错误信息（用于调试）
func (v *EnhancedPriceValidator) LogErrorWithDetails(err *PriceError) {
	// 这里可以集成到您的日志系统
	fmt.Printf("[PRICE_ERROR] Symbol: %s, Type: %d, Message: %s, Details: %s\n",
		err.Symbol, err.Type, err.Message, err.Details)
}

// GetCommonSymbolVariants 获取常见交易对的变体（用于建议）
func (v *EnhancedPriceValidator) GetCommonSymbolVariants(symbol string) []string {
	// 移除USDT后缀进行匹配
	baseSymbol := strings.TrimSuffix(strings.ToUpper(symbol), "USDT")

	variants := []string{
		baseSymbol + "USDT",
		baseSymbol + "-USDT",
		baseSymbol + "_USDT",
	}

	// 特殊处理一些常见情况
	switch baseSymbol {
	case "ENSO":
		variants = append(variants, "ENSUSDT", "ENSO-USDT")
	case "ENS":
		variants = append(variants, "ENSOUSDT", "ENS-USDT")
	}

	return variants
}
