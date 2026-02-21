package trader

import (
	"log/slog"
	"time"
)

// OrphanOrderCleanupTask 定期清理孤儿订单的任务
type OrphanOrderCleanupTask struct {
	cleaner  *OrphanOrderCleaner
	interval time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewOrphanOrderCleanupTask 创建一个新的孤儿订单清理任务
func NewOrphanOrderCleanupTask(trader Trader, interval time.Duration) *OrphanOrderCleanupTask {
	return &OrphanOrderCleanupTask{
		cleaner:  NewOrphanOrderCleaner(trader),
		interval: interval,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start 启动定期清理任务
func (t *OrphanOrderCleanupTask) Start() {
	go t.run()
	slog.Info("Started orphan order cleanup task", "interval", t.interval)
}

// Stop 停止定期清理任务
func (t *OrphanOrderCleanupTask) Stop() {
	close(t.stopCh)
	<-t.doneCh
	slog.Info("Stopped orphan order cleanup task")
}

// run 运行清理任务
func (t *OrphanOrderCleanupTask) run() {
	defer close(t.doneCh)

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	// 立即执行一次
	t.performCleanup()

	for {
		select {
		case <-ticker.C:
			t.performCleanup()
		case <-t.stopCh:
			slog.Info("Orphan order cleanup task received stop signal")
			return
		}
	}
}

// performCleanup 执行清理操作
func (t *OrphanOrderCleanupTask) performCleanup() {
	isEmpty, err := t.cleaner.IsPositionEmpty()
	if err != nil {
		slog.Error("Failed to check if position is empty", "error", err)
		return
	}

	if isEmpty {
		slog.Info("Position is empty, cleaning up orphaned orders")
		err := t.cleaner.SafeCleanupOrphanedOrders()
		if err != nil {
			slog.Error("Failed to clean up orphaned orders", "error", err)
		}
	} else {
		slog.Info("Position is not empty, skipping orphaned order cleanup")
	}
}

// CleanupNow 立即执行一次清理
func (t *OrphanOrderCleanupTask) CleanupNow() error {
	slog.Info("Manually triggering orphan order cleanup")

	isEmpty, err := t.cleaner.IsPositionEmpty()
	if err != nil {
		return err
	}

	if isEmpty {
		slog.Info("Position is empty, cleaning up orphaned orders")
		return t.cleaner.SafeCleanupOrphanedOrders()
	} else {
		slog.Info("Position is not empty, skipping orphaned order cleanup")
		return nil
	}
}
