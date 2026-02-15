package trader

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"nofx/logger"

	"github.com/adshao/go-binance/v2/futures"
)

// DebugMarketPrice 获取市场价格并详细记录调试信息
func DebugMarketPrice(client *futures.Client, symbol string) (float64, error) {
	logger.Infof("🔍 开始调试 %s 市场价格获取", symbol)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 记录请求详情
	logger.Debugf("🌐 请求参数 - Symbol: %s, Timeout: %v", symbol, 20*time.Second)

	prices, err := client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		logger.Errorf("❌ API调用失败: %v", err)
		logger.Errorf("   错误类型: %T", err)
		if ctx.Err() == context.DeadlineExceeded {
			logger.Errorf("   超时详情: 请求超过20秒未响应")
		}
		return 0, fmt.Errorf("API调用失败: %w", err)
	}

	// 详细记录响应数据
	logger.Debugf("✅ API调用成功")
	logger.Debugf("   返回价格数量: %d", len(prices))

	if len(prices) == 0 {
		logger.Warnf("⚠️  API返回空的价格列表")
		return 0, fmt.Errorf("price not found for symbol %s", symbol)
	}

	priceData := prices[0]
	logger.Debugf("   价格数据详情:")
	logger.Debugf("     Symbol: %s", priceData.Symbol)
	logger.Debugf("     Price字段类型: %T", priceData.Price)
	logger.Debugf("     Price字段值: '%s'", priceData.Price)
	logger.Debugf("     Price字段长度: %d", len(priceData.Price))

	// 检查空字符串情况
	if priceData.Price == "" {
		logger.Errorf("❌ 检测到空的价格字符串!")
		logger.Errorf("   Symbol: %s", priceData.Symbol)
		logger.Errorf("   Price字段: '%s' (长度: %d)", priceData.Price, len(priceData.Price))
		return 0, fmt.Errorf("empty price string received for symbol %s", symbol)
	}

	// 尝试解析
	price, err := strconv.ParseFloat(priceData.Price, 64)
	if err != nil {
		logger.Errorf("❌ 价格解析失败:")
		logger.Errorf("   原始数据: '%s'", priceData.Price)
		logger.Errorf("   错误信息: %v", err)
		logger.Errorf("   错误类型: %T", err)
		return 0, fmt.Errorf("failed to parse price '%s': %w", priceData.Price, err)
	}

	logger.Infof("✅ %s 价格获取成功: %.8f", symbol, price)
	return price, nil
}

// ValidatePriceResponse 验证价格响应数据的完整性
func ValidatePriceResponse(prices []*futures.SymbolPrice, symbol string) error {
	if len(prices) == 0 {
		return fmt.Errorf("empty price response for symbol %s", symbol)
	}

	priceData := prices[0]

	// 检查基本字段
	if priceData.Symbol == "" {
		return fmt.Errorf("missing symbol in price response")
	}

	if priceData.Symbol != symbol {
		return fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, priceData.Symbol)
	}

	// 检查价格字段
	if priceData.Price == "" {
		return fmt.Errorf("empty price field for symbol %s", symbol)
	}

	// 检查价格格式
	if _, err := strconv.ParseFloat(priceData.Price, 64); err != nil {
		return fmt.Errorf("invalid price format '%s' for symbol %s: %w", priceData.Price, symbol, err)
	}

	return nil
}

// GetPriceWithFallback 带有降级机制的价格获取
func GetPriceWithFallback(client *futures.Client, symbol string) (float64, error) {
	// 首先尝试正常获取
	price, err := DebugMarketPrice(client, symbol)
	if err == nil {
		return price, nil
	}

	logger.Warnf("⚠️  %s 价格获取失败，尝试降级方案: %v", symbol, err)

	// 降级方案1：获取所有价格然后筛选
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	allPrices, err := client.NewListPricesService().Do(ctx)
	if err != nil {
		logger.Errorf("❌ 降级方案也失败: %v", err)
		return 0, fmt.Errorf("all price fetching methods failed for %s", symbol)
	}

	// 在所有价格中查找目标符号
	for _, priceInfo := range allPrices {
		if priceInfo.Symbol == symbol && priceInfo.Price != "" {
			if price, err := strconv.ParseFloat(priceInfo.Price, 64); err == nil {
				logger.Infof("✅ 通过降级方案获取到 %s 价格: %.8f", symbol, price)
				return price, nil
			}
		}
	}

	logger.Errorf("❌ 无法通过任何方式获取 %s 的价格", symbol)
	return 0, fmt.Errorf("price not available for symbol %s", symbol)
}
