package trader

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nofx/logger"

	"github.com/adshao/go-binance/v2/futures"
)

// fetchExchangeInfoWithForcedMainnet 强制使用主网获取exchangeInfo
// 专门用于精度信息获取，不影响其他交易功能
func fetchExchangeInfoWithForcedMainnet(client *futures.Client, timeout time.Duration) (*futures.ExchangeInfo, error) {
	// 保存原始BaseURL
	originalBaseURL := client.BaseURL

	// 强制使用主网地址获取精度信息
	forcedMainnetURL := "https://fapi.binance.com"
	client.BaseURL = forcedMainnetURL

	logger.Infof("🔧 [精度专用强制连接] 切换到主网地址获取精度信息: %s", forcedMainnetURL)

	// 创建新的HTTP客户端（如果需要）
	originalClient := client.HTTPClient
	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{
			Timeout: timeout,
		}
	} else {
		client.HTTPClient.Timeout = timeout
	}

	// 执行API调用
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	startTime := time.Now()
	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	elapsed := time.Since(startTime)

	// 恢复原始配置
	client.BaseURL = originalBaseURL
	client.HTTPClient = originalClient

	if err != nil {
		logger.Warnf("❌ [精度专用强制连接] 主网获取失败 (耗时: %v): %v", elapsed, err)
		return nil, fmt.Errorf("forced mainnet exchange info fetch failed: %w", err)
	}

	logger.Infof("✅ [精度专用强制连接] 主网获取成功 (耗时: %v), 符号数量: %d", elapsed, len(exchangeInfo.Symbols))
	return exchangeInfo, nil
}

// GetSymbolPrecisionForced 强制使用主网获取符号精度（仅用于精度获取）
func GetSymbolPrecisionForced(client *futures.Client, symbol string) (int, error) {
	// 使用强制主网连接获取exchangeInfo
	exchangeInfo, err := fetchExchangeInfoWithForcedMainnet(client, 30*time.Second)
	if err != nil {
		logger.Warnf("⚠️ 强制主网获取精度失败 %s, 使用默认精度 3: %v", symbol, err)
		return 3, nil // 降级到默认精度
	}

	// 从返回的数据中提取精度信息
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			// 获取LOT_SIZE过滤器的精度
			for _, filter := range s.Filters {
				if filterType, ok := filter["filterType"].(string); ok && filterType == "LOT_SIZE" {
					if stepSize, ok := filter["stepSize"].(string); ok {
						precision := calculatePrecisionForced(stepSize)
						logger.Infof("✅ [强制主网] %s 数量精度: %d (步长: %s)", symbol, precision, stepSize)
						return precision, nil
					}
				}
			}
		}
	}

	logger.Warnf("⚠️ [强制主网] %s 精度信息未找到, 使用默认精度 3", symbol)
	return 3, nil
}

// GetPricePrecisionForced 强制使用主网获取价格精度（仅用于精度获取）
func GetPricePrecisionForced(client *futures.Client, symbol string) (int, error) {
	// 使用强制主网连接获取exchangeInfo
	exchangeInfo, err := fetchExchangeInfoWithForcedMainnet(client, 30*time.Second)
	if err != nil {
		logger.Warnf("⚠️ 强制主网获取价格精度失败 %s, 使用默认精度 2: %v", symbol, err)
		return 2, nil // 降级到默认精度
	}

	// 从返回的数据中提取价格精度信息
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			// 获取PRICE_FILTER过滤器的精度
			for _, filter := range s.Filters {
				if filterType, ok := filter["filterType"].(string); ok && filterType == "PRICE_FILTER" {
					if tickSize, ok := filter["tickSize"].(string); ok {
						precision := calculatePrecisionForced(tickSize)
						logger.Infof("✅ [强制主网] %s 价格精度: %d (价格步长: %s)", symbol, precision, tickSize)
						return precision, nil
					}
				}
			}
		}
	}

	logger.Warnf("⚠️ [强制主网] %s 价格精度信息未找到, 使用默认精度 2", symbol)
	return 2, nil
}

// calculatePrecisionForced 计算精度位数（避免重复定义）
func calculatePrecisionForced(stepSize string) int {
	if stepSize == "" {
		return 0
	}

	// 移除尾随零
	stepSize = strings.TrimRight(stepSize, "0")

	// 如果是整数（如"1"）
	if !strings.Contains(stepSize, ".") {
		return 0
	}

	// 计算小数位数
	parts := strings.Split(stepSize, ".")
	if len(parts) == 2 {
		return len(parts[1])
	}

	return 0
}
