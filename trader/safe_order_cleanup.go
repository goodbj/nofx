package trader

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// SafeOrderCleanupManager 安全的订单清理管理器
// 采用多重验证机制，确保不会误删持仓相关的有效委托
type SafeOrderCleanupManager struct {
	trader   Trader
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewSafeOrderCleanupManager 创建安全的订单清理管理器
func NewSafeOrderCleanupManager(trader Trader, interval time.Duration) *SafeOrderCleanupManager {
	return &SafeOrderCleanupManager{
		trader:   trader,
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start 启动安全清理任务
func (m *SafeOrderCleanupManager) Start() {
	go m.run()
	slog.Info("Started safe order cleanup manager", "interval", m.interval)
}

// Stop 停止清理任务
func (m *SafeOrderCleanupManager) Stop() {
	close(m.stopCh)
	<-m.doneCh
	slog.Info("Stopped safe order cleanup manager")
}

// run 运行清理循环
func (m *SafeOrderCleanupManager) run() {
	defer close(m.doneCh)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// 立即执行一次
	m.performSafeCleanup()

	for {
		select {
		case <-ticker.C:
			m.performSafeCleanup()
		case <-m.stopCh:
			slog.Info("Safe order cleanup manager received stop signal")
			return
		}
	}
}

// performSafeCleanup 执行安全的清理操作
// 采用多重验证机制确保安全
func (m *SafeOrderCleanupManager) performSafeCleanup() {
	slog.Info("Starting safe order cleanup process")

	// 第一层验证：检查持仓状态
	isEmpty, err := m.isPositionTrulyEmpty()
	if err != nil {
		slog.Error("Failed to check position status", "error", err)
		return
	}

	if !isEmpty {
		slog.Info("Position is not empty, skipping cleanup")
		return
	}

	slog.Info("Position is empty, proceeding with safety checks")

	// 第二层验证：获取当前所有订单
	openOrders, err := m.trader.GetOpenOrders("")
	if err != nil {
		slog.Error("Failed to get open orders", "error", err)
		return
	}

	if len(openOrders) == 0 {
		slog.Info("No open orders found")
		return
	}

	slog.Info("Found open orders, performing detailed analysis", "count", len(openOrders))

	// 第三层验证：详细分析每个订单
	ordersToCancel := make([]OpenOrder, 0)
	safeOrders := make([]OpenOrder, 0)

	for _, order := range openOrders {
		if m.isOrderTrulyOrphaned(order) {
			ordersToCancel = append(ordersToCancel, order)
			slog.Info("Identified orphaned order",
				"symbol", order.Symbol,
				"orderID", order.OrderID,
				"type", order.Type,
				"side", order.Side)
		} else {
			safeOrders = append(safeOrders, order)
			slog.Info("Order is safe to keep",
				"symbol", order.Symbol,
				"orderID", order.OrderID,
				"type", order.Type,
				"side", order.Side)
		}
	}

	// 第四层验证：最终确认
	if len(ordersToCancel) == 0 {
		slog.Info("No orphaned orders to cancel")
		return
	}

	slog.Info("Pre-cleanup summary",
		"orphaned_orders", len(ordersToCancel),
		"safe_orders", len(safeOrders))

	// 执行清理前的最后确认
	if !m.finalSafetyCheck() {
		slog.Warn("Final safety check failed, aborting cleanup")
		return
	}

	// 执行清理操作
	canceledCount := 0
	for _, order := range ordersToCancel {
		err := m.cancelOrderWithRetry(order)
		if err != nil {
			slog.Error("Failed to cancel order",
				"symbol", order.Symbol,
				"orderID", order.OrderID,
				"error", err)
		} else {
			canceledCount++
			slog.Info("Successfully canceled order",
				"symbol", order.Symbol,
				"orderID", order.OrderID)
		}
	}

	slog.Info("Safe order cleanup completed",
		"attempted_cancellations", len(ordersToCancel),
		"successful_cancellations", canceledCount)
}

// isPositionTrulyEmpty 严格检查持仓是否为空
func (m *SafeOrderCleanupManager) isPositionTrulyEmpty() (bool, error) {
	positions, err := m.trader.GetPositions()
	if err != nil {
		return false, fmt.Errorf("failed to get positions: %w", err)
	}

	// 检查所有持仓
	for _, pos := range positions {
		posSymbol, ok := pos["symbol"].(string)
		if !ok {
			continue
		}

		posAmt, ok := pos["positionAmt"].(float64)
		if !ok {
			continue
		}

		// 使用更严格的阈值判断
		if posAmt > 0.000000001 || posAmt < -0.000000001 {
			slog.Info("Found non-empty position",
				"symbol", posSymbol,
				"amount", posAmt)
			return false, nil
		}
	}

	slog.Info("All positions are empty")
	return true, nil
}

// isOrderTrulyOrphaned 判断订单是否真正为孤儿订单
func (m *SafeOrderCleanupManager) isOrderTrulyOrphaned(order OpenOrder) bool {
	// 检查是否为测试订单或特殊订单
	if m.isSpecialOrder(order) {
		return false
	}

	// 检查是否为系统重要订单
	if m.isSystemCriticalOrder(order) {
		return false
	}

	// 检查订单类型和时间
	if m.isRecentOrder(order) {
		return false
	}

	// 检查是否为正常交易订单
	if m.isNormalTradingOrder(order) {
		return false
	}

	// 经过所有检查后，认为是孤儿订单
	return true
}

// isSpecialOrder 检查是否为特殊订单
func (m *SafeOrderCleanupManager) isSpecialOrder(order OpenOrder) bool {
	// 检查订单ID是否包含特殊标识
	specialPrefixes := []string{"TEST", "SYSTEM", "ADMIN", "MAINTENANCE"}
	for _, prefix := range specialPrefixes {
		if strings.HasPrefix(order.OrderID, prefix) {
			return true
		}
	}
	return false
}

// isSystemCriticalOrder 检查是否为系统关键订单
func (m *SafeOrderCleanupManager) isSystemCriticalOrder(order OpenOrder) bool {
	// 检查是否为流动性提供订单等
	criticalTypes := []string{"LIQUIDITY", "MAKER"}
	for _, ct := range criticalTypes {
		if strings.Contains(strings.ToUpper(order.Type), ct) {
			return true
		}
	}
	return false
}

// isRecentOrder 检查是否为近期订单
func (m *SafeOrderCleanupManager) isRecentOrder(order OpenOrder) bool {
	// 检查订单创建时间（如果可用）
	// 这里可以根据实际的订单结构添加时间检查逻辑
	// 暂时返回false，表示不进行时间检查
	return false
}

// isNormalTradingOrder 检查是否为正常的交易订单
func (m *SafeOrderCleanupManager) isNormalTradingOrder(order OpenOrder) bool {
	// 检查订单是否符合正常的交易模式
	normalTypes := []string{"LIMIT", "MARKET", "STOP", "TAKE_PROFIT"}
	for _, nt := range normalTypes {
		if strings.Contains(strings.ToUpper(order.Type), nt) {
			return true
		}
	}
	return false
}

// finalSafetyCheck 最终安全检查
func (m *SafeOrderCleanupManager) finalSafetyCheck() bool {
	// 再次确认持仓状态
	isEmpty, err := m.isPositionTrulyEmpty()
	if err != nil {
		slog.Error("Final safety check failed: cannot verify position status", "error", err)
		return false
	}

	if !isEmpty {
		slog.Warn("Final safety check failed: position is not empty")
		return false
	}

	// 检查系统状态
	if !m.isSystemStable() {
		slog.Warn("Final safety check failed: system is not stable")
		return false
	}

	slog.Info("Final safety check passed")
	return true
}

// isSystemStable 检查系统是否稳定
func (m *SafeOrderCleanupManager) isSystemStable() bool {
	// 这里可以添加系统稳定性检查逻辑
	// 例如：检查是否有其他正在进行的交易操作
	// 暂时返回true
	return true
}

// cancelOrderWithRetry 带重试机制的订单取消
func (m *SafeOrderCleanupManager) cancelOrderWithRetry(order OpenOrder) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := m.cancelSingleOrder(order)
		if err == nil {
			return nil
		}

		lastErr = err
		slog.Warn("Failed to cancel order, retrying",
			"attempt", i+1,
			"symbol", order.Symbol,
			"orderID", order.OrderID,
			"error", err)

		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1)) // 递增延迟
		}
	}

	return fmt.Errorf("failed to cancel order after %d attempts: %w", maxRetries, lastErr)
}

// cancelSingleOrder 取消单个订单
func (m *SafeOrderCleanupManager) cancelSingleOrder(order OpenOrder) error {
	// 使用类型断言来访问特定的交易商实现
	switch v := m.trader.(type) {
	case *FuturesTrader:
		return m.cancelBinanceOrder(v, order.Symbol, order.OrderID)
	case *AsterTrader:
		return m.cancelAsterOrder(v, order.Symbol, order.OrderID)
	default:
		return fmt.Errorf("unsupported trader type for order cancellation")
	}
}

// cancelBinanceOrder 取消Binance订单
func (m *SafeOrderCleanupManager) cancelBinanceOrder(trader *FuturesTrader, symbol, orderID string) error {
	slog.Info("Canceling Binance order", "symbol", symbol, "orderID", orderID)

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
		return fmt.Errorf("failed to cancel Binance order: %w", err)
	}

	slog.Info("Successfully canceled Binance order", "symbol", symbol, "orderID", orderID)
	return nil
}

// cancelAsterOrder 取消Aster订单
func (m *SafeOrderCleanupManager) cancelAsterOrder(trader *AsterTrader, symbol, orderID string) error {
	slog.Info("Canceling Aster order", "symbol", symbol, "orderID", orderID)

	params := map[string]interface{}{
		"symbol":  symbol,
		"orderId": orderID,
	}

	_, err := trader.request("DELETE", "/fapi/v3/order", params)
	if err != nil {
		return fmt.Errorf("failed to cancel Aster order: %w", err)
	}

	slog.Info("Successfully canceled Aster order", "symbol", symbol, "orderID", orderID)
	return nil
}

// ManualCleanup 立即执行手动清理
func (m *SafeOrderCleanupManager) ManualCleanup() error {
	slog.Info("Manual cleanup triggered")
	m.performSafeCleanup()
	return nil
}
