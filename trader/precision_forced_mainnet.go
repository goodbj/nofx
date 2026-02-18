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

// GetSymbolPrecisionForcedFromManager 通过PrecisionManager强制获取符号精度
func GetSymbolPrecisionForcedFromManager(manager PrecisionManagerInterface, symbol string) (int, error) {
	// 类型断言获取具体实现
	if pm, ok := manager.(*FileBasedPrecisionManager); ok {
		info, err := pm.GetPrecisionInfo(symbol)
		if err != nil {
			return 0, err
		}
		return info.Precision, nil
	}

	// 如果是其他实现，使用普通方法
	info, err := manager.GetPrecisionInfo(symbol)
	if err != nil {
		return 0, err
	}
	return info.Precision, nil
}

// GetPricePrecisionForcedFromManager 通过PrecisionManager强制获取价格精度
func GetPricePrecisionForcedFromManager(manager PrecisionManagerInterface, symbol string) (int, error) {
	// 类型断言获取具体实现
	if pm, ok := manager.(*FileBasedPrecisionManager); ok {
		info, err := pm.GetPrecisionInfo(symbol)
		if err != nil {
			return 0, err
		}
		return info.Precision, nil
	}

	// 如果是其他实现，使用普通方法
	info, err := manager.GetPrecisionInfo(symbol)
	if err != nil {
		return 0, err
	}
	return info.Precision, nil // 使用相同的Precision字段
}

// Deprecated: 旧的函数，保留是为了向后兼容
// 请使用FileBasedPrecisionManager中的方法
func GetSymbolPrecisionForced(client *futures.Client, symbol string) (int, error) {
	// 创建FileBasedPrecisionManager实例
	manager := NewFileBasedPrecisionManager(client, "trader/jingdu.json")
	return GetSymbolPrecisionForcedFromManager(manager, symbol)
}

// Deprecated: 旧的函数，保留是为了向后兼容
// 请使用FileBasedPrecisionManager中的方法
func GetPricePrecisionForced(client *futures.Client, symbol string) (int, error) {
	// 创建FileBasedPrecisionManager实例
	manager := NewFileBasedPrecisionManager(client, "trader/jingdu.json")
	return GetPricePrecisionForcedFromManager(manager, symbol)
}
