// 集成精度处理器到现有交易系统

package trader

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// IntegratedTrader 集成精度处理的交易器
type IntegratedTrader struct {
	*FuturesTrader
	precisionHandler *PrecisionHandler
}

// NewIntegratedTrader 创建集成交易器
func NewIntegratedTrader(client *futures.Client) *IntegratedTrader {
	return &IntegratedTrader{
		FuturesTrader:    &FuturesTrader{client: client},
		precisionHandler: NewPrecisionHandler(client),
	}
}

// OpenLongWithPrecision 带精度处理的开多单
func (t *IntegratedTrader) OpenLongWithPrecision(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 调整数量精度
	adjustedQty, err := t.precisionHandler.AdjustQuantity(symbol, quantity)
	if err != nil {
		return nil, fmt.Errorf("quantity adjustment failed: %w", err)
	}

	// 验证订单
	price, err := t.GetMarketPrice(symbol)
	if err != nil {
		return nil, err
	}

	limitPrice := price * 1.01 // 略高于市价确保成交

	if err := t.precisionHandler.ValidateOrder(symbol, adjustedQty, limitPrice); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// 格式化参数
	qtyStr, err := t.precisionHandler.FormatQuantity(symbol, adjustedQty)
	if err != nil {
		return nil, err
	}

	priceStr, err := t.precisionHandler.FormatPrice(symbol, limitPrice)
	if err != nil {
		return nil, err
	}

	// 执行交易
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeBuy).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(qtyStr).
		Price(priceStr).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return map[string]interface{}{
		"orderId":     order.OrderID,
		"symbol":      order.Symbol,
		"origQty":     order.OrigQuantity,
		"price":       order.Price,
		"adjustedQty": adjustedQty,
		"originalQty": quantity,
	}, nil
}

// OpenShortWithPrecision 带精度处理的开空单
func (t *IntegratedTrader) OpenShortWithPrecision(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 调整数量精度
	adjustedQty, err := t.precisionHandler.AdjustQuantity(symbol, quantity)
	if err != nil {
		return nil, fmt.Errorf("quantity adjustment failed: %w", err)
	}

	// 验证订单
	price, err := t.GetMarketPrice(symbol)
	if err != nil {
		return nil, err
	}

	limitPrice := price * 0.99 // 略低于市价确保成交

	if err := t.precisionHandler.ValidateOrder(symbol, adjustedQty, limitPrice); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// 格式化参数
	qtyStr, err := t.precisionHandler.FormatQuantity(symbol, adjustedQty)
	if err != nil {
		return nil, err
	}

	priceStr, err := t.precisionHandler.FormatPrice(symbol, limitPrice)
	if err != nil {
		return nil, err
	}

	// 执行交易
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeSell).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(qtyStr).
		Price(priceStr).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return map[string]interface{}{
		"orderId":     order.OrderID,
		"symbol":      order.Symbol,
		"origQty":     order.OrigQuantity,
		"price":       order.Price,
		"adjustedQty": adjustedQty,
		"originalQty": quantity,
	}, nil
}

// PlaceOrderWithPrecision 通用下单函数
func (t *IntegratedTrader) PlaceOrderWithPrecision(symbol string, side futures.SideType,
	orderType futures.OrderType, quantity float64, price float64) (map[string]interface{}, error) {

	// 调整数量和价格精度
	adjustedQty, err := t.precisionHandler.AdjustQuantity(symbol, quantity)
	if err != nil {
		return nil, fmt.Errorf("quantity adjustment failed: %w", err)
	}

	adjustedPrice, err := t.precisionHandler.AdjustPrice(symbol, price)
	if err != nil {
		return nil, fmt.Errorf("price adjustment failed: %w", err)
	}

	// 验证订单
	if err := t.precisionHandler.ValidateOrder(symbol, adjustedQty, adjustedPrice); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// 格式化参数
	qtyStr, err := t.precisionHandler.FormatQuantity(symbol, adjustedQty)
	if err != nil {
		return nil, err
	}

	priceStr, err := t.precisionHandler.FormatPrice(symbol, adjustedPrice)
	if err != nil {
		return nil, err
	}

	// 执行交易
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(orderType).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(qtyStr).
		Price(priceStr).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	return map[string]interface{}{
		"orderId":       order.OrderID,
		"symbol":        order.Symbol,
		"origQty":       order.OrigQuantity,
		"price":         order.Price,
		"adjustedQty":   adjustedQty,
		"adjustedPrice": adjustedPrice,
		"originalQty":   quantity,
		"originalPrice": price,
	}, nil
}

// GetSymbolPrecisionInfo 获取交易对精度信息
func (t *IntegratedTrader) GetSymbolPrecisionInfo(symbol string) (*SymbolPrecisionInfo, error) {
	return t.precisionHandler.GetSymbolPrecision(symbol)
}

// ValidateQuantity 验证数量是否符合精度要求
func (t *IntegratedTrader) ValidateQuantity(symbol string, quantity float64) error {
	_, err := t.precisionHandler.AdjustQuantity(symbol, quantity)
	return err
}

// ValidatePrice 验证价格是否符合精度要求
func (t *IntegratedTrader) ValidatePrice(symbol string, price float64) error {
	_, err := t.precisionHandler.AdjustPrice(symbol, price)
	return err
}

// FormatQuantityForAPI 格式化数量用于API调用
func (t *IntegratedTrader) FormatQuantityForAPI(symbol string, quantity float64) (string, error) {
	return t.precisionHandler.FormatQuantity(symbol, quantity)
}

// FormatPriceForAPI 格式化价格用于API调用
func (t *IntegratedTrader) FormatPriceForAPI(symbol string, price float64) (string, error) {
	return t.precisionHandler.FormatPrice(symbol, price)
}

// ClearPrecisionCache 清除精度缓存
func (t *IntegratedTrader) ClearPrecisionCache() {
	t.precisionHandler.ClearCache()
}

// GetPrecisionCacheSize 获取精度缓存大小
func (t *IntegratedTrader) GetPrecisionCacheSize() int {
	return t.precisionHandler.GetCacheSize()
}
