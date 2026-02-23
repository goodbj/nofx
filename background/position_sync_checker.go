package background

import (
	"fmt"
	"time"

	"nofx/manager"
	"nofx/store"

	"github.com/sirupsen/logrus"
)

// PositionSyncChecker 持仓同步检查器 - 使用原生平仓功能处理不一致记录
type PositionSyncChecker struct {
	store         *store.Store
	traderManager *manager.TraderManager
	checkInterval time.Duration
	logger        *logrus.Logger
	isRunning     bool
	stopChan      chan struct{}
}

// NewPositionSyncChecker 创建新的持仓同步检查器实例
func NewPositionSyncChecker(st *store.Store, tm *manager.TraderManager, interval time.Duration, logger *logrus.Logger) *PositionSyncChecker {
	return &PositionSyncChecker{
		store:         st,
		traderManager: tm,
		checkInterval: interval,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动后台检查服务
func (psc *PositionSyncChecker) Start() {
	if psc.isRunning {
		psc.logger.Println("持仓同步检查服务已在运行中")
		return
	}

	psc.isRunning = true
	psc.logger.Infof("启动持仓同步检查服务，检查间隔: %v", psc.checkInterval)

	// 立即执行一次检查
	go psc.performCheck()

	// 启动定时检查
	ticker := time.NewTicker(psc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go psc.performCheck()
		case <-psc.stopChan:
			psc.logger.Println("停止持仓同步检查服务")
			psc.isRunning = false
			return
		}
	}
}

// Stop 停止后台检查服务
func (psc *PositionSyncChecker) Stop() {
	if !psc.isRunning {
		return
	}

	close(psc.stopChan)
}

// performCheck 执行单次检查
func (psc *PositionSyncChecker) performCheck() {
	psc.logger.Infoln("开始执行持仓同步检查...")

	startTime := time.Now()

	// 检查并处理不一致的持仓记录
	processedCount, err := psc.syncPositions()
	if err != nil {
		psc.logger.Errorf("检查过程中出现错误: %v", err)
		return
	}

	duration := time.Since(startTime)
	psc.logger.Infof("检查完成 - 处理记录数: %d, 耗时: %v", processedCount, duration)
}

// syncPositions 同步持仓状态
func (psc *PositionSyncChecker) syncPositions() (int, error) {
	// 获取所有数据库中的OPEN持仓记录
	dbPositions, err := psc.store.Position().GetOpenPositions("")
	if err != nil {
		return 0, fmt.Errorf("查询数据库持仓记录失败: %w", err)
	}

	processedCount := 0

	// 对每个交易员的持仓进行检查
	traderMap := make(map[string][]*store.TraderPosition)
	for _, pos := range dbPositions {
		traderMap[pos.TraderID] = append(traderMap[pos.TraderID], pos)
	}

	// 检查每个交易员的实际持仓与数据库记录是否一致
	for traderID, positions := range traderMap {
		err := psc.checkTraderPositions(traderID, positions)
		if err != nil {
			psc.logger.Errorf("检查交易员 %s 持仓时出现错误: %v", traderID, err)
			continue
		}
		processedCount += len(positions)
	}

	return processedCount, nil
}

// checkTraderPositions 检查单个交易员的持仓一致性
func (psc *PositionSyncChecker) checkTraderPositions(traderID string, dbPositions []*store.TraderPosition) error {
	// 获取实际的交易所持仓
	actualPositions, err := psc.getActualExchangePositions(traderID)
	if err != nil {
		psc.logger.Errorf("获取交易员 %s 实际持仓失败: %v", traderID, err)
		return err
	}

	// 创建实际持仓的映射以便快速查找
	actualPosMap := make(map[string]string) // symbol -> side
	for _, pos := range actualPositions {
		key := fmt.Sprintf("%s_%s", pos.Symbol, pos.Side) // 使用 symbol_side 作为唯一标识
		actualPosMap[key] = pos.Side
	}

	// 检查数据库中的每个持仓是否在实际持仓中存在
	for _, dbPos := range dbPositions {
		dbKey := fmt.Sprintf("%s_%s", dbPos.Symbol, dbPos.Side)

		// 如果数据库中的持仓在实际持仓中不存在，则需要清理
		if _, exists := actualPosMap[dbKey]; !exists {
			psc.logger.Infof("发现不一致持仓: 交易员=%s, 币种=%s, 方向=%s, 数据库状态=OPEN",
				traderID, dbPos.Symbol, dbPos.Side)

			// 调用原生平仓功能来清理这个不一致的持仓
			err := psc.cleanupOrphanedPosition(traderID, dbPos)
			if err != nil {
				psc.logger.Errorf("清理孤儿持仓失败: %s, 错误: %v", dbKey, err)
				continue
			}
		}
	}

	return nil
}

// getActualExchangePositions 获取实际的交易所持仓
func (psc *PositionSyncChecker) getActualExchangePositions(traderID string) ([]*store.TraderPosition, error) {
	// 获取交易员实例
	t, err := psc.traderManager.GetTrader(traderID)
	if err != nil {
		return nil, fmt.Errorf("获取交易员实例失败: %w", err)
	}

	// 通过交易员接口获取实际持仓
	actualPositions := []*store.TraderPosition{}

	// 通过交易员接口获取实际持仓
	actualPosData, err := t.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取实际持仓数据失败: %w", err)
	}

	// 将实际持仓数据转换为TraderPosition格式
	for _, posData := range actualPosData {
		if symbol, ok := posData["symbol"].(string); ok {
			side := "LONG"
			if posSide, ok := posData["side"].(string); ok {
				side = posSide
			}

			// 检查持仓数量是否大于0
			if qty, ok := posData["positionAmt"].(float64); ok {
				if qty == 0 || (qty < 0 && side == "long") || (qty > 0 && side == "short") {
					continue // 忽略数量为0的持仓
				}
			}

			actualPositions = append(actualPositions, &store.TraderPosition{
				Symbol: symbol,
				Side:   side,
			})
		}
	}

	return actualPositions, nil
}

// cleanupOrphanedPosition 清理孤儿持仓（已被交易所自动平仓但在数据库中仍为OPEN的记录）
func (psc *PositionSyncChecker) cleanupOrphanedPosition(traderID string, dbPos *store.TraderPosition) error {
	psc.logger.Infof("开始清理孤儿持仓: 交易员=%s, 币种=%s, 方向=%s", traderID, dbPos.Symbol, dbPos.Side)

	// 获取交易员实例
	autoTrader, err := psc.traderManager.GetTrader(traderID)
	if err != nil {
		return fmt.Errorf("获取交易员实例失败: %w", err)
	}

	// 获取AutoTrader内部的Trader接口
	traderInterface := autoTrader.GetTrader()

	// 使用交易员的平仓功能尝试平仓（即使实际已无持仓，也不会造成影响）
	// 这样可以确保数据库记录得到正确更新
	switch dbPos.Side {
	case "LONG", "long":
		// 尝试平多仓 - 如果实际已无持仓，此操作将无效果
		_, err := traderInterface.CloseLong(dbPos.Symbol, 0) // 0 表示平全部
		if err != nil {
			psc.logger.Debugf("平多仓操作结果 (可能正常): %v", err)
			// 即使错误也继续，因为实际持仓可能已不存在
		}
	case "SHORT", "short":
		// 尝试平空仓 - 如果实际已无持仓，此操作将无效果
		_, err := traderInterface.CloseShort(dbPos.Symbol, 0) // 0 表示平全部
		if err != nil {
			psc.logger.Debugf("平空仓操作结果 (可能正常): %v", err)
			// 即使错误也继续，因为实际持仓可能已不存在
		}
	default:
		return fmt.Errorf("未知的持仓方向: %s", dbPos.Side)
	}

	// 更新数据库中的持仓状态为CLOSED
	err = psc.updatePositionToClosed(dbPos, "系统检测到持仓已不存在，自动关闭")
	if err != nil {
		return fmt.Errorf("更新数据库持仓状态失败: %w", err)
	}

	psc.logger.Infof("成功清理孤儿持仓: 交易员=%s, 币种=%s, 方向=%s", traderID, dbPos.Symbol, dbPos.Side)
	return nil
}

// updatePositionToClosed 更新持仓记录为已关闭状态
func (psc *PositionSyncChecker) updatePositionToClosed(pos *store.TraderPosition, reason string) error {
	// 使用PositionStore的ClosePosition方法来正确更新持仓状态
	err := psc.store.Position().ClosePosition(pos.ID, 0, "", 0, 0, reason)
	if err != nil {
		return fmt.Errorf("更新持仓状态失败: %w", err)
	}

	return nil
}

// IsRunning 检查服务是否正在运行
func (psc *PositionSyncChecker) IsRunning() bool {
	return psc.isRunning
}
