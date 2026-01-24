package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nofx/config"
	guardian "nofx/guardian/internal"
)

func main() {
	log.Println("🚀 Starting AI Automation Guardian...")

	// 创建默认配置
	guardianConfig := config.NewDefaultGuardianConfig()

	// 从环境变量覆盖配置（如果有的话）
	if apiEndpoint := os.Getenv("NOFX_API_ENDPOINT"); apiEndpoint != "" {
		guardianConfig.DataTransferConfig.NofxAPIEndpoint = apiEndpoint
	}

	if apiKey := os.Getenv("GUARDIAN_API_KEY"); apiKey != "" {
		guardianConfig.DataTransferConfig.APIKey = apiKey
	}

	if authToken := os.Getenv("GUARDIAN_AUTH_TOKEN"); authToken != "" {
		guardianConfig.DataTransferConfig.AuthToken = authToken
	}

	if aiProvider := os.Getenv("GUARDIAN_AI_PROVIDER"); aiProvider != "" {
		guardianConfig.AIProvider = aiProvider
	}

	if aiEndpoint := os.Getenv("GUARDIAN_AI_ENDPOINT"); aiEndpoint != "" {
		guardianConfig.AIEndpoint = aiEndpoint
		guardianConfig.PageURL = aiEndpoint
	}

	// 创建守护程序实例
	guard := guardian.NewGuardian(guardianConfig)

	// 启动守护程序
	if err := guard.Start(); err != nil {
		log.Fatalf("❌ Failed to start guardian: %v", err)
	}

	log.Println("✅ AI Automation Guardian is running...")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-sigChan
	log.Printf("⚠️ Received signal: %s, shutting down gracefully...", sig)

	// 停止守护程序
	guard.Stop()

	// 等待一点时间确保清理完成
	time.Sleep(2 * time.Second)

	log.Println("🛑 AI Automation Guardian shutdown complete")
}
