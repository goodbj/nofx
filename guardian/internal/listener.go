package guardian

import (
	"context"
	"log"
	"nofx/config"
	"nofx/kernel"
	"time"
)

// PromptListener 提示词监听器
type PromptListener struct {
	config   *config.GuardianConfig
	ctx      context.Context
	cancel   context.CancelFunc
	onPrompt func(string, string) // 回调函数，接收系统提示词和用户提示词
}

// NewPromptListener 创建新的提示词监听器
func NewPromptListener(config *config.GuardianConfig) *PromptListener {
	ctx, cancel := context.WithCancel(context.Background())

	return &PromptListener{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动监听器
func (pl *PromptListener) Start() error {
	log.Println("👂 Starting prompt listener...")

	// 启动定期检查任务
	go pl.pollForPrompts()

	return nil
}

// Stop 停止监听器
func (pl *PromptListener) Stop() {
	pl.cancel()
	log.Println("🛑 Prompt listener stopped")
}

// SetOnPrompt 设置提示词回调函数
func (pl *PromptListener) SetOnPrompt(callback func(string, string)) {
	pl.onPrompt = callback
}

// pollForPrompts 定期轮询最新的提示词
func (pl *PromptListener) pollForPrompts() {
	ticker := time.NewTicker(time.Duration(pl.config.ListenInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pl.ctx.Done():
			return
		case <-ticker.C:
			// 获取最新的提示词
			systemPrompt, userPrompt, timestamp := kernel.GetLastPrompt()

			// 如果有新的提示词（时间戳比上次检查的更新），则触发回调
			if !timestamp.IsZero() && timestamp.After(time.Now().Add(-time.Duration(pl.config.ListenInterval+1000)*time.Millisecond)) {
				log.Printf("🎯 New prompt detected, system prompt length: %d, user prompt length: %d",
					len(systemPrompt), len(userPrompt))

				if pl.onPrompt != nil {
					pl.onPrompt(systemPrompt, userPrompt)
				}
			}
		}
	}
}

// GetLastPrompt 获取最后的提示词（直接调用内核函数）
func (pl *PromptListener) GetLastPrompt() (string, string, time.Time) {
	return kernel.GetLastPrompt()
}
