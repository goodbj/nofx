package trader

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// PrecisionHandler 通用精度处理器
type PrecisionHandler struct {
	client        *futures.Client
	symbolCache   map[string]*SymbolPrecisionInfo
	cacheMutex    sync.RWMutex
	cacheDuration time.Duration
}

// SymbolPrecisionInfo 交易对精度信息
type SymbolPrecisionInfo struct {
	Symbol            string
	PricePrecision    int
	QuantityPrecision int
	MinPrice          float64
	MaxPrice          float64
	TickSize          float64
	MinQuantity       float64
	MaxQuantity       float64
	StepSize          float64
	MinNotional       float64
	LastUpdated       time.Time
}

// NewPrecisionHandler 创建精度处理器
func NewPrecisionHandler(client *futures.Client) *PrecisionHandler {
	return &PrecisionHandler{
		client:        client,
		symbolCache:   make(map[string]*SymbolPrecisionInfo),
		cacheDuration: 5 * time.Minute, // 5分钟缓存
	}
}

// GetSymbolPrecision 获取交易对精度信息
func (ph *PrecisionHandler) GetSymbolPrecision(symbol string) (*SymbolPrecisionInfo, error) {
	// 检查缓存
	ph.cacheMutex.RLock()
	if info, exists := ph.symbolCache[symbol]; exists && time.Since(info.LastUpdated) < ph.cacheDuration {
		ph.cacheMutex.RUnlock()
		return info, nil
	}
	ph.cacheMutex.RUnlock()

	// 获取交易所信息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeInfo, err := ph.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange info: %w", err)
	}

	// 查找指定交易对
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			info := ph.parseSymbolInfo(&s)
			// 更新缓存
			ph.cacheMutex.Lock()
			ph.symbolCache[symbol] = info
			ph.cacheMutex.Unlock()
			return info, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found", symbol)
}

// parseSymbolInfo 解析交易对信息
func (ph *PrecisionHandler) parseSymbolInfo(symbol *futures.Symbol) *SymbolPrecisionInfo {
	info := &SymbolPrecisionInfo{
		Symbol:            symbol.Symbol,
		PricePrecision:    symbol.PricePrecision,
		QuantityPrecision: symbol.QuantityPrecision,
		LastUpdated:       time.Now(),
	}

	// 解析过滤器
	for _, filter := range symbol.Filters {
		switch filterType := filter["filterType"].(string); filterType {
		case "PRICE_FILTER":
			info.MinPrice, _ = strconv.ParseFloat(filter["minPrice"].(string), 64)
			info.MaxPrice, _ = strconv.ParseFloat(filter["maxPrice"].(string), 64)
			info.TickSize, _ = strconv.ParseFloat(filter["tickSize"].(string), 64)
		case "LOT_SIZE":
			info.MinQuantity, _ = strconv.ParseFloat(filter["minQty"].(string), 64)
			info.MaxQuantity, _ = strconv.ParseFloat(filter["maxQty"].(string), 64)
			info.StepSize, _ = strconv.ParseFloat(filter["stepSize"].(string), 64)
		case "MIN_NOTIONAL":
			info.MinNotional, _ = strconv.ParseFloat(filter["notional"].(string), 64)
		}
	}

	return info
}

// AdjustQuantity 调整交易数量到符合精度要求
func (ph *PrecisionHandler) AdjustQuantity(symbol string, quantity float64) (float64, error) {
	info, err := ph.GetSymbolPrecision(symbol)
	if err != nil {
		// 如果获取精度信息失败，使用默认处理
		return ph.defaultAdjustQuantity(quantity), nil
	}

	// 首先检查最小数量限制
	if quantity < info.MinQuantity {
		return 0, fmt.Errorf("quantity %.8f is below minimum %.8f", quantity, info.MinQuantity)
	}

	// 检查最大数量限制
	if quantity > info.MaxQuantity {
		quantity = info.MaxQuantity
	}

	// 按步长调整数量
	if info.StepSize > 0 {
		quantity = ph.roundToStepSize(quantity, info.StepSize)
	}

	// 按精度位数调整
	multiplier := math.Pow10(info.QuantityPrecision)
	quantity = math.Round(quantity*multiplier) / multiplier

	// 再次检查是否满足最小数量要求（调整后可能不满足）
	if quantity < info.MinQuantity {
		return 0, fmt.Errorf("adjusted quantity %.8f is below minimum %.8f", quantity, info.MinQuantity)
	}

	return quantity, nil
}

// AdjustPrice 调整价格到符合精度要求
func (ph *PrecisionHandler) AdjustPrice(symbol string, price float64) (float64, error) {
	info, err := ph.GetSymbolPrecision(symbol)
	if err != nil {
		// 如果获取精度信息失败，使用默认处理
		return ph.defaultAdjustPrice(price), nil
	}

	// 检查价格范围
	if price < info.MinPrice {
		price = info.MinPrice
	}
	if price > info.MaxPrice {
		price = info.MaxPrice
	}

	// 按tick size调整
	if info.TickSize > 0 {
		price = ph.roundToStepSize(price, info.TickSize)
	}

	// 按精度位数调整
	multiplier := math.Pow10(info.PricePrecision)
	price = math.Round(price*multiplier) / multiplier

	return price, nil
}

// ValidateOrder 验证订单是否符合精度要求
func (ph *PrecisionHandler) ValidateOrder(symbol string, quantity, price float64) error {
	info, err := ph.GetSymbolPrecision(symbol)
	if err != nil {
		return err
	}

	// 验证数量
	if quantity < info.MinQuantity {
		return fmt.Errorf("quantity %.8f is below minimum %.8f", quantity, info.MinQuantity)
	}
	if quantity > info.MaxQuantity {
		return fmt.Errorf("quantity %.8f exceeds maximum %.8f", quantity, info.MaxQuantity)
	}

	// 验证价格
	if price < info.MinPrice {
		return fmt.Errorf("price %.8f is below minimum %.8f", price, info.MinPrice)
	}
	if price > info.MaxPrice {
		return fmt.Errorf("price %.8f exceeds maximum %.8f", price, info.MaxPrice)
	}

	// 验证最小名义金额
	notional := quantity * price
	if notional < info.MinNotional {
		return fmt.Errorf("order notional %.8f is below minimum %.8f", notional, info.MinNotional)
	}

	return nil
}

// roundToStepSize 将数值调整到步长的倍数
func (ph *PrecisionHandler) roundToStepSize(value, stepSize float64) float64 {
	if stepSize <= 0 {
		return value
	}
	steps := value / stepSize
	roundedSteps := math.Round(steps)
	return roundedSteps * stepSize
}

// defaultAdjustQuantity 默认数量调整（当无法获取精度信息时使用）
func (ph *PrecisionHandler) defaultAdjustQuantity(quantity float64) float64 {
	// 默认保留6位小数
	multiplier := math.Pow10(6)
	return math.Round(quantity*multiplier) / multiplier
}

// defaultAdjustPrice 默认价格调整（当无法获取精度信息时使用）
func (ph *PrecisionHandler) defaultAdjustPrice(price float64) float64 {
	// 默认保留2位小数
	multiplier := math.Pow10(2)
	return math.Round(price*multiplier) / multiplier
}

// FormatQuantity 格式化数量为字符串
func (ph *PrecisionHandler) FormatQuantity(symbol string, quantity float64) (string, error) {
	adjustedQty, err := ph.AdjustQuantity(symbol, quantity)
	if err != nil {
		return "", err
	}

	info, err := ph.GetSymbolPrecision(symbol)
	if err != nil {
		// 使用默认格式
		return fmt.Sprintf("%.6f", adjustedQty), nil
	}

	format := fmt.Sprintf("%%.%df", info.QuantityPrecision)
	return fmt.Sprintf(format, adjustedQty), nil
}

// FormatPrice 格式化价格为字符串
func (ph *PrecisionHandler) FormatPrice(symbol string, price float64) (string, error) {
	adjustedPrice, err := ph.AdjustPrice(symbol, price)
	if err != nil {
		return "", err
	}

	info, err := ph.GetSymbolPrecision(symbol)
	if err != nil {
		// 使用默认格式
		return fmt.Sprintf("%.2f", adjustedPrice), nil
	}

	format := fmt.Sprintf("%%.%df", info.PricePrecision)
	return fmt.Sprintf(format, adjustedPrice), nil
}

// ClearCache 清除缓存
func (ph *PrecisionHandler) ClearCache() {
	ph.cacheMutex.Lock()
	ph.symbolCache = make(map[string]*SymbolPrecisionInfo)
	ph.cacheMutex.Unlock()
}

// GetCacheSize 获取缓存大小
func (ph *PrecisionHandler) GetCacheSize() int {
	ph.cacheMutex.RLock()
	defer ph.cacheMutex.RUnlock()
	return len(ph.symbolCache)
}
