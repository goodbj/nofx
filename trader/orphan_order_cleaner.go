package trader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// OrphanOrderCleaner 负责清理与当前持仓无关的孤儿委托
type OrphanOrderCleaner struct {
	trader Trader
}

// NewOrphanOrderCleaner 创建一个新的孤儿委托清理器
func NewOrphanOrderCleaner(trader Trader) *OrphanOrderCleaner {
	return &OrphanOrderCleaner{
		trader: trader,
	}
}

// SafeCleanupOrphanedOrders 安全地清理所有孤儿委托（传统订单+Algo订单）
func (c *OrphanOrderCleaner) SafeCleanupOrphanedOrders() error {
	positions, err := c.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// 1. 清理传统孤儿订单
	openOrders, err := c.trader.GetOpenOrders("")
	if err != nil {
		return fmt.Errorf("failed to get open orders: %w", err)
	}

	// 按交易对分组订单，便于处理
	ordersBySymbol := make(map[string][]OpenOrder)
	for _, order := range openOrders {
		ordersBySymbol[order.Symbol] = append(ordersBySymbol[order.Symbol], order)
	}

	canceledCount := 0

	// 对每个交易对的订单进行检查和清理
	for _, orders := range ordersBySymbol {
		for _, order := range orders {
			if c.shouldCancelOrder(order, positions) {
				// 在取消前再次确认
				if c.shouldReallyCancelOrder(order, positions) {
					err := c.cancelOrderBySymbolAndID(order.Symbol, order.OrderID)
					if err != nil {
						slog.Info("Failed to cancel orphaned order", "symbol", order.Symbol, "orderID", order.OrderID, "error", err)
						continue
					}
					slog.Info("Canceled orphaned order", "symbol", order.Symbol, "orderID", order.OrderID, "type", order.Type)
					canceledCount++
				}
			}
		}
	}

	// 2. 清理Algo孤儿订单（Binance已迁移，止损/止盈主要是Algo单）
	if ft, ok := c.trader.(*FuturesTrader); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		algoOrders, algoErr := ft.client.NewListOpenAlgoOrdersService().
			Do(ctx)
		if algoErr != nil {
			slog.Warn("Failed to list all Algo orders for orphan cleanup", "error", algoErr)
		} else {
			for _, algoOrder := range algoOrders {
				symbol := algoOrder.Symbol
				// 检查Algo单是否是孤儿
				isOrphan := true
				for _, pos := range positions {
					posSymbol, ok := pos["symbol"].(string)
					if !ok {
						continue
					}
					if !strings.EqualFold(posSymbol, symbol) {
						continue
					}
					posAmt, ok := pos["positionAmt"].(float64)
					if !ok {
						continue
					}
					if posAmt > 0.00000001 || posAmt < -0.00000001 {
						isOrphan = false
						break
					}
				}

				if isOrphan {
					ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel2()
					_, cancelErr := ft.client.NewCancelAlgoOrderService().
						AlgoID(algoOrder.AlgoId).
						Do(ctx2)
					if cancelErr != nil {
						slog.Warn("Failed to cancel orphaned Algo order", "symbol", symbol, "algoId", algoOrder.AlgoId, "error", cancelErr)
					} else {
						slog.Info("Canceled orphaned Algo order", "symbol", symbol, "algoId", algoOrder.AlgoId, "type", algoOrder.OrderType)
						canceledCount++
					}
				}
			}
		}
	}

	if canceledCount > 0 {
		slog.Info("Completed orphaned order cleanup", "canceled_count", canceledCount)
	} else {
		slog.Info("No orphaned orders to clean up")
	}

	return nil
}

// SafeCleanupOrphanedOrdersForSymbol 清理指定交易对的孤儿委托（传统订单+Algo订单）
func (c *OrphanOrderCleaner) SafeCleanupOrphanedOrdersForSymbol(symbol string) error {
	positions, err := c.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// 1. 清理传统孤儿订单
	openOrders, err := c.trader.GetOpenOrders(symbol)
	if err != nil {
		return fmt.Errorf("failed to get open orders for %s: %w", symbol, err)
	}

	canceledCount := 0

	for _, order := range openOrders {
		if c.shouldCancelOrder(order, positions) {
			// 在取消前再次确认
			if c.shouldReallyCancelOrder(order, positions) {
				err := c.cancelOrderBySymbolAndID(order.Symbol, order.OrderID)
				if err != nil {
					slog.Info("Failed to cancel orphaned order", "symbol", order.Symbol, "orderID", order.OrderID, "error", err)
					continue
				}
				slog.Info("Canceled orphaned order", "symbol", order.Symbol, "orderID", order.OrderID, "type", order.Type)
				canceledCount++
			}
		}
	}

	// 2. 清理Algo孤儿订单（Binance已迁移，止损/止盈主要是Algo单）
	// 对于Binance FuturesTrader，直接通过类型断言访问Algo API
	if ft, ok := c.trader.(*FuturesTrader); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		algoOrders, algoErr := ft.client.NewListOpenAlgoOrdersService().
			Symbol(symbol).
			Do(ctx)
		if algoErr != nil {
			slog.Warn("Failed to list Algo orders for orphan cleanup", "symbol", symbol, "error", algoErr)
		} else {
			for _, algoOrder := range algoOrders {
				// 检查Algo单是否是孤儿（没有对应持仓）
				isOrphan := true
				for _, pos := range positions {
					posSymbol, ok := pos["symbol"].(string)
					if !ok {
						continue
					}
					if !strings.EqualFold(posSymbol, symbol) {
						continue
					}
					posAmt, ok := pos["positionAmt"].(float64)
					if !ok {
						continue
					}
					if posAmt > 0.00000001 || posAmt < -0.00000001 {
						isOrphan = false
						break
					}
				}

				if isOrphan {
					ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel2()
					_, cancelErr := ft.client.NewCancelAlgoOrderService().
						AlgoID(algoOrder.AlgoId).
						Do(ctx2)
					if cancelErr != nil {
						slog.Warn("Failed to cancel orphaned Algo order", "symbol", symbol, "algoId", algoOrder.AlgoId, "error", cancelErr)
					} else {
						slog.Info("Canceled orphaned Algo order", "symbol", symbol, "algoId", algoOrder.AlgoId, "type", algoOrder.OrderType)
						canceledCount++
					}
				}
			}
		}
	}

	if canceledCount > 0 {
		slog.Info("Completed orphaned order cleanup for symbol", "symbol", symbol, "canceled_count", canceledCount)
	} else {
		slog.Info("No orphaned orders to clean up for symbol", "symbol", symbol)
	}

	return nil
}

// shouldCancelOrder 判断是否应该取消委托
func (c *OrphanOrderCleaner) shouldCancelOrder(order OpenOrder, positions []map[string]interface{}) bool {
	// 检查订单是否与任何持仓相关
	for _, pos := range positions {
		if c.isOrderRelatedToPosition(order, pos) {
			// 订单与持仓相关，不应取消
			return false
		}
	}

	// 没有相关持仓，可以考虑取消
	return true
}

// shouldReallyCancelOrder 在取消前进行最终确认
func (c *OrphanOrderCleaner) shouldReallyCancelOrder(order OpenOrder, positions []map[string]interface{}) bool {
	// 再次获取最新的持仓信息以确保准确性
	latestPositions, err := c.trader.GetPositions()
	if err != nil {
		slog.Warn("Failed to get latest positions for final check, proceeding with caution", "error", err)
		latestPositions = positions // 使用传入的持仓信息
	}

	// 重新检查订单是否与任何持仓相关
	for _, pos := range latestPositions {
		if c.isOrderRelatedToPosition(order, pos) {
			// 即使在最后时刻发现订单与持仓相关，也不应取消
			return false
		}
	}

	return true
}

// isOrderRelatedToPosition 检查委托是否与持仓相关
func (c *OrphanOrderCleaner) isOrderRelatedToPosition(order OpenOrder, pos map[string]interface{}) bool {
	// 检查交易对是否匹配
	posSymbol, ok := pos["symbol"].(string)
	if !ok {
		return false
	}

	if strings.ToUpper(order.Symbol) != strings.ToUpper(posSymbol) {
		return false
	}

	// 获取持仓方向
	posSide, ok := pos["side"].(string)
	if !ok {
		return false
	}

	// 获取持仓数量
	posAmt, ok := pos["positionAmt"].(float64)
	if !ok {
		return false
	}

	// 如果持仓数量接近0，认为没有有效持仓
	if posAmt < 0.00000001 && posAmt > -0.00000001 {
		return false
	}

	// 分析订单类型和方向
	switch strings.ToUpper(order.Type) {
	case "STOP_MARKET", "STOP", "TAKE_PROFIT_MARKET", "TAKE_PROFIT":
		// 止损止盈订单通常与持仓相关
		// 检查订单方向是否与持仓方向相反（通常是平仓订单）
		isCloseOrder := (strings.ToUpper(order.Side) == "SELL" && strings.ToUpper(posSide) == "LONG") ||
			(strings.ToUpper(order.Side) == "BUY" && strings.ToUpper(posSide) == "SHORT")

		if isCloseOrder {
			// 这是平仓订单，与持仓相关
			return true
		}

		// 检查是否为加仓订单（较少见，但可能存在）
		isOpenOrder := (strings.ToUpper(order.Side) == "BUY" && strings.ToUpper(posSide) == "LONG") ||
			(strings.ToUpper(order.Side) == "SELL" && strings.ToUpper(posSide) == "SHORT")

		if isOpenOrder {
			// 加仓订单也与持仓相关
			return true
		}

	case "LIMIT", "MARKET":
		// 限价单和市价单的处理
		// 如果是平仓方向的订单，与持仓相关
		isCloseOrder := (strings.ToUpper(order.Side) == "SELL" && strings.ToUpper(posSide) == "LONG") ||
			(strings.ToUpper(order.Side) == "BUY" && strings.ToUpper(posSide) == "SHORT")

		if isCloseOrder {
			return true
		}

		// 如果是开仓方向的订单，也要考虑
		isOpenOrder := (strings.ToUpper(order.Side) == "BUY" && strings.ToUpper(posSide) == "LONG") ||
			(strings.ToUpper(order.Side) == "SELL" && strings.ToUpper(posSide) == "SHORT")

		if isOpenOrder {
			return true
		}
	}

	return false
}

// cancelOrderBySymbolAndID 根据交易对和订单ID取消订单
func (c *OrphanOrderCleaner) cancelOrderBySymbolAndID(symbol, orderID string) error {
	// 由于Trader接口没有直接按订单ID取消的方法，我们使用一个通用的取消方法
	// 实际上，我们会尝试使用扩展接口或特定交易所的API来按ID取消订单

	// 首先尝试获取订单详情以确认是否需要取消
	orders, err := c.trader.GetOpenOrders(symbol)
	if err != nil {
		return fmt.Errorf("failed to get open orders for %s: %w", symbol, err)
	}

	// 查找特定订单
	for _, order := range orders {
		if order.OrderID == orderID {
			// 这里需要一个能够按订单ID取消的方法
			// 由于标准接口不支持，我们可能需要类型断言来访问特定实现
			// 或者实现一个通用的取消方法
			return c.cancelOrderByIDUsingSpecificTrader(symbol, orderID)
		}
	}

	return fmt.Errorf("order %s not found for symbol %s", orderID, symbol)
}

// cancelOrderByIDUsingSpecificTrader 尝试使用特定交易商接口取消订单
func (c *OrphanOrderCleaner) cancelOrderByIDUsingSpecificTrader(symbol, orderID string) error {
	// 使用类型断言来尝试访问具体的交易商类型
	switch v := c.trader.(type) {
	case *FuturesTrader:
		// 对于Binance期货交易商，使用其API直接取消订单
		return c.cancelBinanceOrder(v, symbol, orderID)
	case *AsterTrader:
		// 对于Aster交易商，使用其API取消订单
		return c.cancelAsterOrder(v, symbol, orderID)
	default:
		// 对于其他交易商，尝试使用通用方法
		return c.cancelOrderBySymbolAndFallback(symbol, orderID)
	}
}

// cancelBinanceOrder 使用Binance API取消订单
func (c *OrphanOrderCleaner) cancelBinanceOrder(trader *FuturesTrader, symbol, orderID string) error {
	slog.Info("Canceling orphaned order on Binance", "symbol", symbol, "orderID", orderID)

	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID format: %s", orderID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = trader.client.NewCancelOrderService().
		Symbol(symbol).
		OrderID(orderIDInt).
		Do(ctx)

	if err != nil {
		slog.Error("Failed to cancel Binance order", "symbol", symbol, "orderID", orderID, "error", err)
		return fmt.Errorf("failed to cancel Binance order: %w", err)
	}

	slog.Info("Successfully canceled Binance order", "symbol", symbol, "orderID", orderID)
	return nil
}

// cancelAsterOrder 使用Aster API取消订单
func (c *OrphanOrderCleaner) cancelAsterOrder(trader *AsterTrader, symbol, orderID string) error {
	slog.Info("Canceling orphaned order on Aster", "symbol", symbol, "orderID", orderID)

	params := map[string]interface{}{
		"symbol":  symbol,
		"orderId": orderID,
	}

	_, err := trader.request("DELETE", "/fapi/v3/order", params)
	if err != nil {
		slog.Error("Failed to cancel Aster order", "symbol", symbol, "orderID", orderID, "error", err)
		return fmt.Errorf("failed to cancel Aster order: %w", err)
	}

	slog.Info("Successfully canceled Aster order", "symbol", symbol, "orderID", orderID)
	return nil
}

// cancelOrderBySymbolAndFallback 尝试通过获取订单列表后按ID取消
func (c *OrphanOrderCleaner) cancelOrderBySymbolAndFallback(symbol, orderID string) error {
	// 获取所有开放订单
	orders, err := c.trader.GetOpenOrders(symbol)
	if err != nil {
		return fmt.Errorf("failed to get open orders for fallback cancellation: %w", err)
	}

	// 检查目标订单是否存在
	found := false
	for _, order := range orders {
		if order.OrderID == orderID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("order %s not found for symbol %s", orderID, symbol)
	}

	// 由于接口限制，我们使用CancelAllOrders作为后备方案
	// 但在实际使用中，这可能会取消其他有效订单，所以需要谨慎
	// 在孤儿订单清理场景中，这通常是可接受的，因为我们只在持仓为空时运行

	slog.Info("Using fallback cancellation method - canceling all orders for symbol", "symbol", symbol)
	return c.trader.CancelAllOrders(symbol)
}

// IsPositionEmpty 检查持仓是否为空
func (c *OrphanOrderCleaner) IsPositionEmpty() (bool, error) {
	positions, err := c.trader.GetPositions()
	if err != nil {
		return false, fmt.Errorf("failed to get positions: %w", err)
	}

	// 检查是否所有持仓的数量都接近0
	for _, pos := range positions {
		posAmt, ok := pos["positionAmt"].(float64)
		if !ok {
			continue
		}

		// 如果有任何持仓的数量不接近0，则持仓不为空
		if posAmt > 0.00000001 || posAmt < -0.00000001 {
			return false, nil
		}
	}

	return true, nil
}

// CleanupIfPositionEmpty 如果持仓为空则清理孤儿订单
func (c *OrphanOrderCleaner) CleanupIfPositionEmpty() error {
	isEmpty, err := c.IsPositionEmpty()
	if err != nil {
		return fmt.Errorf("failed to check position status: %w", err)
	}

	if isEmpty {
		slog.Info("Position is empty, starting orphaned order cleanup")
		err = c.SafeCleanupOrphanedOrders()
		if err != nil {
			slog.Error("Failed to clean up orphaned orders", "error", err)
			return err
		}
		slog.Info("Successfully cleaned up orphaned orders")
	} else {
		slog.Info("Position is not empty, skipping orphaned order cleanup")
	}

	return nil
}

// CleanupOrphanedOrdersForSymbolIfNoPosition 检查特定交易对是否有持仓，如果没有则清理该交易对的孤儿订单
func (c *OrphanOrderCleaner) CleanupOrphanedOrdersForSymbolIfNoPosition(symbol string) error {
	positions, err := c.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// 检查指定交易对是否有持仓
	hasPosition := false
	for _, pos := range positions {
		posSymbol, ok := pos["symbol"].(string)
		if !ok {
			continue
		}

		if strings.EqualFold(posSymbol, symbol) {
			posAmt, ok := pos["positionAmt"].(float64)
			if !ok {
				continue
			}

			// 检查持仓数量是否接近0
			if posAmt > 0.00000001 || posAmt < -0.00000001 {
				hasPosition = true
				break
			}
		}
	}

	if !hasPosition {
		slog.Info("No position for symbol, cleaning up orphaned orders", "symbol", symbol)
		return c.SafeCleanupOrphanedOrdersForSymbol(symbol)
	} else {
		slog.Info("Position exists for symbol, skipping orphaned order cleanup", "symbol", symbol)
		return nil
	}
}
