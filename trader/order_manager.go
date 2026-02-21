package trader

import (
	"encoding/json"
	"fmt"
	"nofx/logger"
	"sync"
	"time"
)

// OrderManager 管理订单操作的安全锁机制
type OrderManager struct {
	mutex         sync.RWMutex
	isCleaning    bool
	pendingOrders []interface{} // 使用 interface{} 以避免循环依赖
	cleanupTimer  *time.Timer
}

// 全局订单管理器实例
var orderManager = &OrderManager{}

// StartCleanup 开始清理并锁定新订单
func (om *OrderManager) StartCleanup() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	om.isCleaning = true
	logger.Info("Order submission locked, starting cleanup...")

	// 设置超时，防止长时间锁定
	om.cleanupTimer = time.AfterFunc(30*time.Second, func() {
		om.EndCleanup()
		logger.Warn("Cleanup timeout reached, releasing order lock")
	})
}

// EndCleanup 结束清理并解锁
func (om *OrderManager) EndCleanup() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	om.isCleaning = false
	if om.cleanupTimer != nil {
		om.cleanupTimer.Stop()
		om.cleanupTimer = nil
	}
	logger.Info("Order submission unlocked")

	// 处理清理期间积压的订单请求
	om.processPendingOrders()
}

// IsCleaning 检查是否正在进行清理
func (om *OrderManager) IsCleaning() bool {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	return om.isCleaning
}

// CanSubmitOrder 检查是否可以提交订单
func (om *OrderManager) CanSubmitOrder() bool {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	return !om.isCleaning
}

// SubmitOrderWithLock 尝试提交订单（带安全锁）
func (om *OrderManager) SubmitOrderWithLock(order interface{}) error {
	om.mutex.RLock()

	if om.isCleaning {
		om.mutex.RUnlock()

		// 返回错误，不允许在清理期间提交订单
		return fmt.Errorf("cannot submit order during cleanup")
	}

	om.mutex.RUnlock()
	return nil // 让调用方执行实际的下单逻辑
}

// ProcessPendingOrders 处理积压的订单
func (om *OrderManager) processPendingOrders() {
	if len(om.pendingOrders) == 0 {
		return
	}

	logger.Infof("Processing %d pending orders after cleanup", len(om.pendingOrders))

	for _, order := range om.pendingOrders {
		// 注意：这里只是占位符，实际实现需要具体的订单处理逻辑
		_ = order
		// 实际的订单处理逻辑应该在这里
	}

	om.pendingOrders = nil
}

// CleanupOrdersWhenNoPositions 清理无持仓时的挂单
func (om *OrderManager) CleanupOrdersWhenNoPositions(getPositionsFunc func() ([]map[string]interface{}, error), getOpenOrdersFunc func(string) ([]OpenOrder, error), cancelOrderFunc func(string) error, symbol string) error {
	om.StartCleanup()
	defer om.EndCleanup()

	logger.Info("Starting order cleanup with safety lock...")

	// 检查持仓状态
	positions, err := getPositionsFunc()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	hasActivePositions := false
	for _, pos := range positions {
		if qty, ok := pos["quantity"].(*json.Number); ok {
			if qtyFloat, err := qty.Float64(); err == nil && qtyFloat != 0 {
				hasActivePositions = true
				break
			}
		} else if qty, ok := pos["quantity"].(float64); ok {
			if qty != 0 {
				hasActivePositions = true
				break
			}
		} else if qty, ok := pos["quantity"].(int); ok {
			if qty != 0 {
				hasActivePositions = true
				break
			}
		}
	}

	// 如果确实没有持仓，清理所有挂单
	if !hasActivePositions {
		openOrders, err := getOpenOrdersFunc(symbol)
		if err != nil {
			return fmt.Errorf("failed to get open orders: %w", err)
		}

		if len(openOrders) > 0 {
			logger.Infof("Cleaning up %d orphaned orders...", len(openOrders))

			for _, order := range openOrders {
				logger.Infof("Cancelling orphaned order: %s for %s",
					order.OrderID, order.Symbol)

				if err := cancelOrderFunc(order.OrderID); err != nil {
					logger.Errorf("Failed to cancel order %s: %v", order.OrderID, err)
				}
			}

			logger.Info("Order cleanup completed")
		} else {
			logger.Info("No open orders to clean up")
		}
	}

	return nil
}

// GetOrderManager 获取全局订单管理器实例
func GetOrderManager() *OrderManager {
	return orderManager
}
