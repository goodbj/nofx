package guardian

import (
	"context"
	"log"
	"nofx/config"
)

// Guardian 主守护程序结构
type Guardian struct {
	config       *config.GuardianConfig
	listener     *PromptListener
	browserAgent *BrowserAutomation
	dataTransfer *DataTransfer
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewGuardian 创建新的守护程序实例
func NewGuardian(config *config.GuardianConfig) *Guardian {
	ctx, cancel := context.WithCancel(context.Background())

	return &Guardian{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动守护程序
func (g *Guardian) Start() error {
	log.Println("🚀 Starting AI Automation Guardian...")

	// 初始化各模块
	if err := g.InitializeModules(); err != nil {
		return err
	}

	// 设置监听器回调函数
	g.listener.SetOnPrompt(func(systemPrompt, userPrompt string) {
		if err := g.ProcessPrompt(systemPrompt, userPrompt); err != nil {
			log.Printf("❌ Error processing prompt: %v", err)
		}
	})

	// 启动监听模块
	if err := g.listener.Start(); err != nil {
		return err
	}

	log.Println("✅ AI Automation Guardian started successfully")
	return nil
}

// Stop 停止守护程序
func (g *Guardian) Stop() {
	log.Println("🛑 Stopping AI Automation Guardian...")
	g.cancel()
	if g.browserAgent != nil {
		g.browserAgent.Close()
	}
	log.Println("✅ AI Automation Guardian stopped")
}

// InitializeModules 初始化各功能模块
func (g *Guardian) InitializeModules() error {
	// 初始化监听模块
	g.listener = NewPromptListener(g.config)

	// 初始化浏览器自动化模块
	var err error
	g.browserAgent, err = NewBrowserAutomation(g.config.BrowserConfig)
	if err != nil {
		return err
	}

	// 初始化数据传输模块
	g.dataTransfer = NewDataTransfer(g.config.DataTransferConfig)

	return nil
}

// ProcessPrompt 处理从NoFx获取的提示词
func (g *Guardian) ProcessPrompt(systemPrompt, userPrompt string) error {
	log.Printf("📨 Received prompt from NoFx. System prompt length: %d, User prompt length: %d",
		len(systemPrompt), len(userPrompt))

	// 将提示词发送到浏览器AI服务
	aiResponse, err := g.browserAgent.ProcessPrompt(systemPrompt, userPrompt)
	if err != nil {
		log.Printf("❌ Error processing prompt in browser: %v", err)
		return err
	}

	log.Printf("✅ AI response received, length: %d", len(aiResponse))

	// 将AI响应传回NoFx执行
	err = g.dataTransfer.SendDecisionToNoFx(aiResponse)
	if err != nil {
		log.Printf("❌ Error sending decision to NoFx: %v", err)
		return err
	}

	log.Println("✅ Decision sent to NoFx for execution")
	return nil
}
