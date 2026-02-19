package trader

import (
	"context"
	"fmt"
	"time"

	"nofx/logger"

	"github.com/adshao/go-binance/v2/futures"
)

// IntegratedErrorHandlingTrader 集成错误处理的交易器
type IntegratedErrorHandlingTrader struct {
	*futures.Client
	classifier *ErrorClassifier
	validator  *EnhancedPriceValidator
}

// NewIntegratedErrorHandlingTrader 创建集成错误处理的交易器
func NewIntegratedErrorHandlingTrader(apiKey, secretKey string) *IntegratedErrorHandlingTrader {
	client := futures.NewClient(apiKey, secretKey)
	classifier := NewErrorClassifier(true)
	validator := NewEnhancedPriceValidator(true)

	return &IntegratedErrorHandlingTrader{
		Client:     client,
		classifier: classifier,
		validator:  validator,
	}
}

// GetMarketPrice 集成错误处理的价格获取方法
func (t *IntegratedErrorHandlingTrader) GetMarketPrice(symbol string) (float64, error) {
	// 设置上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 记录开始时间用于性能监控
	startTime := time.Now()

	// 调用币安API
	prices, err := t.NewListPricesService().Symbol(symbol).Do(ctx)

	// 记录API调用耗时
	duration := time.Since(startTime)
	logger.Debugf("API调用耗时 %s: %v", symbol, duration)

	if err != nil {
		// 网络或API错误处理
		contextInfo := map[string]interface{}{
			"symbol":   symbol,
			"duration": duration.String(),
			"timeout":  "15s",
		}

		categorizedErr := t.classifier.ClassifyError(err, contextInfo)
		logger.Errorf("❌ %s", t.classifier.FormatErrorForDisplay(categorizedErr))

		// 记录技术详情用于调试
		logger.Debugf("技术详情: %s", t.classifier.GetTechnicalDetails(categorizedErr))

		return 0, fmt.Errorf(t.classifier.FormatErrorForDisplay(categorizedErr))
	}

	// 检查响应数据
	if len(prices) == 0 {
		emptyErr := fmt.Errorf("no price data returned for symbol %s", symbol)
		contextInfo := map[string]interface{}{
			"symbol":   symbol,
			"response": "empty",
		}

		categorizedErr := t.classifier.ClassifyError(emptyErr, contextInfo)
		logger.Warnf("⚠️  %s", t.classifier.FormatErrorForDisplay(categorizedErr))

		return 0, fmt.Errorf(t.classifier.FormatErrorForDisplay(categorizedErr))
	}

	// 验证价格数据
	priceStr := prices[0].Price
	price, validationErr := t.validator.ValidatePriceString(priceStr, symbol)

	if validationErr != nil {
		// 价格验证失败
		contextInfo := map[string]interface{}{
			"symbol":     symbol,
			"price_data": priceStr,
			"raw_error":  validationErr.Error(),
		}

		categorizedErr := t.classifier.ClassifyError(validationErr, contextInfo)

		// 如果是无效交易对，提供替代建议
		if validationErr.IsInvalidSymbol() {
			suggestions := t.validator.GetCommonSymbolVariants(symbol)
			if len(suggestions) > 0 {
				logger.Infof("💡 可能的正确交易对: %v", suggestions)
			}
		}

		logger.Errorf("❌ %s", t.classifier.FormatErrorForDisplay(categorizedErr))
		return 0, fmt.Errorf(t.classifier.FormatErrorForDisplay(categorizedErr))
	}

	// 成功获取价格
	logger.Infof("✅ 成功获取 %s 价格: %.8f (耗时: %v)", symbol, price, duration)
	return price, nil
}

// GetMarketPriceWithFallback 带降级机制的价格获取
func (t *IntegratedErrorHandlingTrader) GetMarketPriceWithFallback(symbol string) (float64, error) {
	// 首先尝试正常获取
	price, err := t.GetMarketPrice(symbol)
	if err == nil {
		return price, nil
	}

	logger.Warnf("⚠️  %s 价格获取失败，尝试降级方案: %v", symbol, err)

	// 降级方案：获取所有价格然后筛选
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	allPrices, err := t.NewListPricesService().Do(ctx)
	if err != nil {
		fallbackErr := fmt.Errorf("降级方案也失败: %w", err)
		contextInfo := map[string]interface{}{
			"symbol":        symbol,
			"primary_error": err.Error(),
			"fallback":      "failed",
		}

		categorizedErr := t.classifier.ClassifyError(fallbackErr, contextInfo)
		logger.Errorf("❌ %s", t.classifier.FormatErrorForDisplay(categorizedErr))
		return 0, fmt.Errorf(t.classifier.FormatErrorForDisplay(categorizedErr))
	}

	// 在所有价格中查找
	for _, priceData := range allPrices {
		if priceData.Symbol == symbol {
			price, validationErr := t.validator.ValidatePriceString(priceData.Price, symbol)
			if validationErr == nil {
				logger.Infof("✅ 通过降级方案成功获取 %s 价格: %.8f", symbol, price)
				return price, nil
			}
		}
	}

	// 最终失败
	finalErr := fmt.Errorf("symbol %s not found in any data source after all attempts", symbol)
	contextInfo := map[string]interface{}{
		"symbol":       symbol,
		"total_prices": len(allPrices),
		"attempts":     "primary + fallback",
	}

	categorizedErr := t.classifier.ClassifyError(finalErr, contextInfo)
	logger.Errorf("❌ %s", t.classifier.FormatErrorForDisplay(categorizedErr))

	return 0, fmt.Errorf(t.classifier.FormatErrorForDisplay(categorizedErr))
}

// BatchGetPrices 批量获取价格（带错误处理）
func (t *IntegratedErrorHandlingTrader) BatchGetPrices(symbols []string) map[string]float64 {
	results := make(map[string]float64)
	failedSymbols := make([]string, 0)

	logger.Infof("🔍 开始批量获取 %d 个交易对价格", len(symbols))

	for _, symbol := range symbols {
		price, err := t.GetMarketPriceWithFallback(symbol)
		if err != nil {
			failedSymbols = append(failedSymbols, symbol)
			continue
		}
		results[symbol] = price
	}

	// 报告结果
	successCount := len(results)
	failureCount := len(failedSymbols)

	logger.Infof("📊 批量获取完成: 成功 %d, 失败 %d", successCount, failureCount)

	if failureCount > 0 {
		logger.Warnf("❌ 获取失败的交易对: %v", failedSymbols)
	}

	return results
}

// ReplaceExistingTraderMethods 替换现有交易器的方法
func (t *IntegratedErrorHandlingTrader) ReplaceExistingTraderMethods() {
	// 这个方法展示了如何替换现有的交易器方法
	// 在实际应用中，您需要将现有交易器的GetMarketPrice方法
	// 指向这个集成版本

	logger.Infof("🔄 已准备替换现有交易器的错误处理逻辑")
	logger.Infof("新特性包括:")
	logger.Infof("  • 自动错误分类和差异化提示")
	logger.Infof("  • 风控机制、系统bug、运行时错误的明确区分")
	logger.Infof("  • 用户友好的错误信息和建议")
	logger.Infof("  • 详细的技术调试信息")
	logger.Infof("  • 智能降级机制")
}

// IntegrationExample 集成使用示例
func IntegrationExample() {
	// 创建集成交易器（使用空密钥仅作演示）
	trader := NewIntegratedErrorHandlingTrader("", "")

	// 替换现有方法
	trader.ReplaceExistingTraderMethods()

	// 测试各种场景
	testSymbols := []string{"ENSOUSDT", "BTCUSDT", "INVALID_SYMBOL"}

	logger.Infof("=== 集成测试开始 ===")

	// 单个获取测试
	for _, symbol := range testSymbols {
		logger.Infof("\n🔍 测试获取 %s 价格:", symbol)
		price, err := trader.GetMarketPriceWithFallback(symbol)
		if err != nil {
			logger.Errorf("❌ %s: %v", symbol, err)
		} else {
			logger.Infof("✅ %s: %.8f", symbol, price)
		}
	}

	// 批量获取测试
	logger.Infof("\n=== 批量获取测试 ===")
	batchResults := trader.BatchGetPrices([]string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "INVALID"})

	logger.Infof("批量结果:")
	for symbol, price := range batchResults {
		logger.Infof("  %s: %.2f", symbol, price)
	}

	logger.Infof("✅ 集成测试完成！")
}
