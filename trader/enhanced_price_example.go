package trader

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// EnhancedFuturesTrader 增强的期货交易器，包含优雅的错误处理
type EnhancedFuturesTrader struct {
	*futures.Client
	validator *EnhancedPriceValidator
}

// NewEnhancedFuturesTrader 创建增强的期货交易器
func NewEnhancedFuturesTrader(apiKey, secretKey string, strictMode bool) *EnhancedFuturesTrader {
	client := futures.NewClient(apiKey, secretKey)
	validator := NewEnhancedPriceValidator(strictMode)

	return &EnhancedFuturesTrader{
		Client:    client,
		validator: validator,
	}
}

// GetMarketPriceWithEnhancedErrorHandling 带增强错误处理的价格获取
func (t *EnhancedFuturesTrader) GetMarketPriceWithEnhancedErrorHandling(symbol string) (float64, error) {
	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 调用API
	prices, err := t.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		// 处理网络错误
		priceErr := &PriceError{
			Type:    PriceErrorNetwork,
			Symbol:  symbol,
			Message: "API call failed",
			Details: err.Error(),
		}

		// 记录详细错误信息
		t.validator.LogErrorWithDetails(priceErr)

		// 返回用户友好的错误信息
		return 0, fmt.Errorf(t.validator.CreateUserFriendlyErrorMessage(priceErr))
	}

	// 检查响应数据
	if len(prices) == 0 {
		priceErr := &PriceError{
			Type:    PriceErrorNoData,
			Symbol:  symbol,
			Message: "no price data returned",
			Details: "API response contains no price information",
		}

		t.validator.LogErrorWithDetails(priceErr)
		return 0, fmt.Errorf(t.validator.CreateUserFriendlyErrorMessage(priceErr))
	}

	// 验证价格字符串
	priceStr := prices[0].Price
	price, priceErr := t.validator.ValidatePriceString(priceStr, symbol)
	if priceErr != nil {
		t.validator.LogErrorWithDetails(priceErr)

		// 如果是无效交易对错误，提供替代建议
		if priceErr.IsInvalidSymbol() {
			suggestions := t.validator.GetCommonSymbolVariants(symbol)
			if len(suggestions) > 0 {
				priceErr.Details += fmt.Sprintf(". 可能的正确交易对: %v", suggestions)
			}
		}

		return 0, fmt.Errorf(t.validator.CreateUserFriendlyErrorMessage(priceErr))
	}

	// 成功获取价格
	fmt.Printf("✅ 成功获取 %s 价格: %.8f\n", symbol, price)
	return price, nil
}

// GetMarketPriceWithFallback 带降级机制的价格获取
func (t *EnhancedFuturesTrader) GetMarketPriceWithFallback(symbol string) (float64, error) {
	// 首先尝试正常获取
	price, err := t.GetMarketPriceWithEnhancedErrorHandling(symbol)
	if err == nil {
		return price, nil
	}

	fmt.Printf("⚠️  %s 价格获取失败，尝试降级方案: %v\n", symbol, err)

	// 降级方案1：获取所有价格然后筛选
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	allPrices, err := t.NewListPricesService().Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("降级方案也失败: %w", err)
	}

	// 在所有价格中查找
	for _, priceData := range allPrices {
		if priceData.Symbol == symbol {
			price, priceErr := t.validator.ValidatePriceString(priceData.Price, symbol)
			if priceErr == nil {
				fmt.Printf("✅ 通过降级方案成功获取 %s 价格: %.8f\n", symbol, price)
				return price, nil
			}
		}
	}

	// 如果还是找不到，返回最终错误
	finalErr := &PriceError{
		Type:    PriceErrorInvalidSymbol,
		Symbol:  symbol,
		Message: "symbol not found in any data source",
		Details: "tried both direct and fallback methods",
	}

	t.validator.LogErrorWithDetails(finalErr)
	return 0, fmt.Errorf(t.validator.CreateUserFriendlyErrorMessage(finalErr))
}

// BatchGetPrices 批量获取多个交易对价格（带错误处理）
func (t *EnhancedFuturesTrader) BatchGetPrices(symbols []string) map[string]float64 {
	results := make(map[string]float64)

	for _, symbol := range symbols {
		price, err := t.GetMarketPriceWithFallback(symbol)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", symbol, err)
			// 可以选择跳过或记录错误
			continue
		}
		results[symbol] = price
	}

	return results
}

// ExampleUsage 使用示例
func ExampleUsage() {
	// 创建增强交易器（无需真实API密钥用于演示）
	trader := NewEnhancedFuturesTrader("", "", true)

	// 测试各种情况
	testSymbols := []string{"ENSOUSDT", "BTCUSDT", "INVALID_SYMBOL", "ENSUSDT"}

	fmt.Println("=== 价格获取测试 ===")
	for _, symbol := range testSymbols {
		fmt.Printf("\n🔍 测试 %s:\n", symbol)

		price, err := trader.GetMarketPriceWithFallback(symbol)
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
		} else {
			fmt.Printf("✅ 价格: %.8f\n", price)
		}
	}

	// 批量获取示例
	fmt.Println("\n=== 批量价格获取 ===")
	batchResults := trader.BatchGetPrices([]string{"BTCUSDT", "ETHUSDT", "SOLUSDT"})
	for symbol, price := range batchResults {
		fmt.Printf("%s: %.2f\n", symbol, price)
	}
}
