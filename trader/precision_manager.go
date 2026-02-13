package trader

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"nofx/logger"

	"github.com/adshao/go-binance/v2/futures"
)

// PrecisionManager 精度管理器 - 通用精度处理解决方案
type PrecisionManager struct {
	client        *futures.Client
	symbolCache   map[string]*SymbolPrecisionInfo
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
}

// SymbolPrecisionInfo 交易对精度信息
type SymbolPrecisionInfo struct {
	Symbol     string
	StepSize   float64
	TickSize   float64
	MinQty     float64
	MaxQty     float64
	Precision  int
	LastUpdate time.Time
}

// NewPrecisionManager 创建精度管理器
func NewPrecisionManager(client *futures.Client) *PrecisionManager {
	return &PrecisionManager{
		client:        client,
		symbolCache:   make(map[string]*SymbolPrecisionInfo),
		cacheDuration: 5 * time.Minute, // 5分钟缓存
	}
}

// GetPrecisionInfo 获取交易对精度信息（带缓存和自动刷新）
func (pm *PrecisionManager) GetPrecisionInfo(symbol string) (*SymbolPrecisionInfo, error) {
	// 1. 先检查缓存
	pm.cacheMutex.RLock()
	if info, exists := pm.symbolCache[symbol]; exists {
		// 检查缓存是否过期
		if time.Since(info.LastUpdate) < pm.cacheDuration {
			pm.cacheMutex.RUnlock()
			logger.Debugf("🎯 使用缓存的精度信息: %s (stepSize: %f)", symbol, info.StepSize)
			return info, nil
		}
	}
	pm.cacheMutex.RUnlock()

	// 2. 缓存过期或不存在，从API获取
	return pm.fetchPrecisionInfo(symbol)
}

// fetchPrecisionInfo 从API获取精度信息
func (pm *PrecisionManager) fetchPrecisionInfo(symbol string) (*SymbolPrecisionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	exchangeInfo, err := pm.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			info := &SymbolPrecisionInfo{
				Symbol:     symbol,
				LastUpdate: time.Now(),
			}

			// 解析过滤器
			for _, filter := range s.Filters {
				filterType, _ := filter["filterType"].(string)
				switch filterType {
				case "LOT_SIZE":
					if stepSizeStr, ok := filter["stepSize"].(string); ok {
						info.StepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
						info.Precision = calculatePrecisionFromStepSize(stepSizeStr)
					}
					if minQtyStr, ok := filter["minQty"].(string); ok {
						info.MinQty, _ = strconv.ParseFloat(minQtyStr, 64)
					}
					if maxQtyStr, ok := filter["maxQty"].(string); ok {
						info.MaxQty, _ = strconv.ParseFloat(maxQtyStr, 64)
					}
				case "PRICE_FILTER":
					if tickSizeStr, ok := filter["tickSize"].(string); ok {
						info.TickSize, _ = strconv.ParseFloat(tickSizeStr, 64)
					}
				}
			}

			// 更新缓存
			pm.cacheMutex.Lock()
			pm.symbolCache[symbol] = info
			pm.cacheMutex.Unlock()

			logger.Infof("✅ 获取到 %s 精度信息: stepSize=%f, precision=%d",
				symbol, info.StepSize, info.Precision)
			return info, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found in exchange info", symbol)
}

// FormatQuantityWithValidation 通用的数量格式化方法（带完整验证）
func (pm *PrecisionManager) FormatQuantityWithValidation(symbol string, quantity float64) (string, error) {
	// 1. 获取精度信息
	info, err := pm.GetPrecisionInfo(symbol)
	if err != nil {
		// 降级到默认处理
		logger.Warnf("⚠️ 无法获取 %s 精度信息，使用默认处理: %v", symbol, err)
		return fmt.Sprintf("%.3f", quantity), nil
	}

	// 2. 边界检查
	if quantity < info.MinQty {
		return "", fmt.Errorf("quantity %.6f is less than minimum allowed %.6f", quantity, info.MinQty)
	}
	if quantity > info.MaxQty {
		return "", fmt.Errorf("quantity %.6f exceeds maximum allowed %.6f", quantity, info.MaxQty)
	}

	// 3. StepSize对齐（核心精度处理）
	alignedQty := pm.alignToStepSize(quantity, info.StepSize)

	// 4. 精度格式化
	formatted := pm.formatWithPrecision(alignedQty, info.Precision)

	// 5. 最终验证
	if err := pm.validateFormattedQuantity(formatted, info); err != nil {
		return "", fmt.Errorf("validation failed after formatting: %w", err)
	}

	logger.Debugf("🎯 %s 数量处理: 原始=%.8f -> 对齐=%.8f -> 格式化=%s",
		symbol, quantity, alignedQty, formatted)

	return formatted, nil
}

// alignToStepSize 按StepSize对齐数量
func (pm *PrecisionManager) alignToStepSize(quantity, stepSize float64) float64 {
	if stepSize <= 0 {
		return quantity
	}

	// 向下取整到最近的stepSize倍数
	// 这样可以确保不会超出精度限制
	aligned := math.Floor(quantity/stepSize) * stepSize

	// 处理浮点数精度问题
	// 例如：0.0030000000000000004 -> 0.003
	aligned = math.Round(aligned*1e10) / 1e10

	return aligned
}

// formatWithPrecision 按精度格式化
func (pm *PrecisionManager) formatWithPrecision(quantity float64, precision int) string {
	if precision < 0 {
		precision = 3 // 默认精度
	}

	format := fmt.Sprintf("%%.%df", precision)
	formatted := fmt.Sprintf(format, quantity)

	// 移除尾随零
	formatted = strings.TrimRight(formatted, "0")
	if strings.HasSuffix(formatted, ".") {
		formatted = formatted[:len(formatted)-1]
	}

	return formatted
}

// validateFormattedQuantity 验证格式化后的数量
func (pm *PrecisionManager) validateFormattedQuantity(formatted string, info *SymbolPrecisionInfo) error {
	// 重新解析验证
	parsedQty, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		return fmt.Errorf("failed to parse formatted quantity: %w", err)
	}

	// 检查是否符合stepSize要求
	if info.StepSize > 0 {
		remainder := math.Mod(parsedQty, info.StepSize)
		// 允许很小的浮点数误差
		if math.Abs(remainder) > 1e-10 && math.Abs(remainder-info.StepSize) > 1e-10 {
			return fmt.Errorf("quantity %s does not align with stepSize %f (remainder: %f)",
				formatted, info.StepSize, remainder)
		}
	}

	return nil
}

// FormatPriceWithValidation 通用的价格格式化方法
func (pm *PrecisionManager) FormatPriceWithValidation(symbol string, price float64) (string, error) {
	info, err := pm.GetPrecisionInfo(symbol)
	if err != nil {
		return fmt.Sprintf("%.2f", price), nil
	}

	// Price对齐到tickSize
	alignedPrice := pm.alignToTickSize(price, info.TickSize)

	// 价格精度通常比数量精度低
	pricePrecision := info.Precision
	if pricePrecision > 8 {
		pricePrecision = 8 // 价格精度上限
	}

	formatted := pm.formatWithPrecision(alignedPrice, pricePrecision)

	logger.Debugf("💰 %s 价格处理: 原始=%.8f -> 对齐=%.8f -> 格式化=%s",
		symbol, price, alignedPrice, formatted)

	return formatted, nil
}

// alignToTickSize 按TickSize对齐价格
func (pm *PrecisionManager) alignToTickSize(price, tickSize float64) float64 {
	if tickSize <= 0 {
		return price
	}

	// 四舍五入到最近的tickSize倍数
	aligned := math.Round(price/tickSize) * tickSize

	// 处理浮点数精度问题
	aligned = math.Round(aligned*1e10) / 1e10

	return aligned
}

// calculatePrecisionFromStepSize 从stepSize计算精度
func calculatePrecisionFromStepSize(stepSize string) int {
	// 移除尾随零
	stepSize = strings.TrimRight(stepSize, "0")
	if strings.HasSuffix(stepSize, ".") {
		stepSize = stepSize[:len(stepSize)-1]
	}

	// 找到小数点
	dotIndex := strings.Index(stepSize, ".")
	if dotIndex == -1 {
		return 0
	}

	return len(stepSize) - dotIndex - 1
}

// ClearCache 清除缓存（用于测试或强制刷新）
func (pm *PrecisionManager) ClearCache() {
	pm.cacheMutex.Lock()
	pm.symbolCache = make(map[string]*SymbolPrecisionInfo)
	pm.cacheMutex.Unlock()
	logger.Infof("🧹 精度缓存已清除")
}

// GetCacheStats 获取缓存统计信息
func (pm *PrecisionManager) GetCacheStats() map[string]interface{} {
	pm.cacheMutex.RLock()
	defer pm.cacheMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["cache_size"] = len(pm.symbolCache)
	stats["cache_duration"] = pm.cacheDuration.String()

	// 统计即将过期的条目
	expiringSoon := 0
	for _, info := range pm.symbolCache {
		if time.Since(info.LastUpdate) > pm.cacheDuration-time.Minute {
			expiringSoon++
		}
	}
	stats["expiring_soon"] = expiringSoon

	return stats
}
