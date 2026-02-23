package background

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// PositionConsistencyChecker 持仓状态一致性检查器
type PositionConsistencyChecker struct {
	db            *sql.DB
	checkInterval time.Duration
	logger        *logrus.Logger
	isRunning     bool
	stopChan      chan struct{}
}

// NewPositionChecker 创建新的检查器实例
func NewPositionChecker(db *sql.DB, interval time.Duration, logger *logrus.Logger) *PositionConsistencyChecker {
	return &PositionConsistencyChecker{
		db:            db,
		checkInterval: interval,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动后台检查服务
func (pcc *PositionConsistencyChecker) Start() {
	if pcc.isRunning {
		pcc.logger.Println("检查服务已在运行中")
		return
	}

	pcc.isRunning = true
	pcc.logger.Infof("启动持仓状态一致性检查服务，检查间隔: %v", pcc.checkInterval)

	// 立即执行一次检查
	go pcc.performCheck()

	// 启动定时检查
	ticker := time.NewTicker(pcc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go pcc.performCheck()
		case <-pcc.stopChan:
			pcc.logger.Println("停止持仓状态检查服务")
			pcc.isRunning = false
			return
		}
	}
}

// Stop 停止后台检查服务
func (pcc *PositionConsistencyChecker) Stop() {
	if !pcc.isRunning {
		return
	}

	close(pcc.stopChan)
}

// performCheck 执行单次检查
func (pcc *PositionConsistencyChecker) performCheck() {
	pcc.logger.Println("开始执行持仓状态一致性检查...")

	startTime := time.Now()

	// 检查并修复不一致记录
	fixedCount, err := pcc.checkAndFixPositions()
	if err != nil {
		pcc.logger.Errorf("检查过程中出现错误: %v", err)
		return
	}

	duration := time.Since(startTime)
	pcc.logger.Infof("检查完成 - 修复记录数: %d, 耗时: %v", fixedCount, duration)
}

// checkAndFixPositions 检查并修复持仓状态
func (pcc *PositionConsistencyChecker) checkAndFixPositions() (int, error) {
	// 查询所有OPEN状态的持仓记录
	query := `
		SELECT id, trader_id, symbol, status, open_time, close_time
		FROM trader_positions 
		WHERE status = 'OPEN'
		ORDER BY trader_id, open_time
	`

	rows, err := pcc.db.Query(query)
	if err != nil {
		return 0, fmt.Errorf("查询持仓记录失败: %w", err)
	}
	defer rows.Close()

	var recordsToFix []int64
	now := time.Now()

	// 识别需要修复的记录
	for rows.Next() {
		var id int64
		var traderID, symbol, status string
		var openTime time.Time
		var closeTime *time.Time

		err := rows.Scan(&id, &traderID, &symbol, &status, &openTime, &closeTime)
		if err != nil {
			pcc.logger.Errorf("扫描记录失败 (ID: %d): %v", id, err)
			continue
		}

		// 检查是否应该自动关闭（开仓超过24小时）
		if now.Sub(openTime) > 24*time.Hour {
			recordsToFix = append(recordsToFix, id)
		}
	}

	if len(recordsToFix) == 0 {
		pcc.logger.Println("未发现需要修复的记录")
		return 0, nil
	}

	// 执行批量修复
	fixedCount, err := pcc.batchFixPositions(recordsToFix)
	if err != nil {
		return 0, fmt.Errorf("批量修复失败: %w", err)
	}

	pcc.logger.Infof("成功修复 %d 条记录", fixedCount)
	return fixedCount, nil
}

// batchFixPositions 批量修复持仓状态
func (pcc *PositionConsistencyChecker) batchFixPositions(recordIDs []int64) (int, error) {
	tx, err := pcc.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	fixedCount := 0
	now := time.Now()
	reason := "系统自动检测到持仓状态不一致，且开仓时间超过24小时"

	// 批量更新
	for _, id := range recordIDs {
		updateQuery := `
			UPDATE trader_positions 
			SET status = 'CLOSED',
				close_time = $1,
				close_reason = $2,
				updated_at = $1
			WHERE id = $3
		`

		result, err := tx.Exec(updateQuery, now, reason, id)
		if err != nil {
			pcc.logger.Errorf("更新记录失败 (ID: %d): %v", id, err)
			continue
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			pcc.logger.Errorf("获取影响行数失败 (ID: %d): %v", id, err)
			continue
		}

		if rowsAffected > 0 {
			fixedCount++
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("提交事务失败: %w", err)
	}

	return fixedCount, nil
}

// IsRunning 检查服务是否正在运行
func (pcc *PositionConsistencyChecker) IsRunning() bool {
	return pcc.isRunning
}
