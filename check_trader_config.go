package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run check_trader_config.go <trader_id>")
		os.Exit(1)
	}

	traderID := os.Args[1]
	fmt.Printf("🔍 检查交易员配置: %s\n", traderID)
	fmt.Println("========================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化加密服务
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatalf("加密服务初始化失败: %v", err)
	}

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 获取交易员完整配置
	fullConfig, err := s.Trader().GetFullConfig("e8e9da5d-c213-4111-b12d-96a0c5875385", traderID)
	if err != nil {
		log.Fatalf("未找到交易员配置: %v", err)
	}

	traderConfig := fullConfig.Trader
	exchangeCfg := fullConfig.Exchange

	fmt.Printf("📋 交易员配置详情:\n")
	fmt.Printf("   ID: %s\n", traderConfig.ID)
	fmt.Printf("   名称: %s\n", traderConfig.Name)
	fmt.Printf("   AI模型ID: %s\n", traderConfig.AIModelID)
	fmt.Printf("   交易所ID: %s\n", traderConfig.ExchangeID)
	fmt.Printf("   策略ID: %s\n", traderConfig.StrategyID)
	fmt.Printf("   初始余额: %.2f\n", traderConfig.InitialBalance)
	fmt.Printf("   扫描间隔: %d分钟\n", traderConfig.ScanIntervalMinutes)
	fmt.Printf("   运行状态: %t\n", traderConfig.IsRunning)
	fmt.Printf("   交叉保证金: %t\n", traderConfig.IsCrossMargin)
	fmt.Printf("   显示在竞赛中: %t\n", traderConfig.ShowInCompetition)

	if exchangeCfg != nil {
		fmt.Printf("\n📋 交易所配置详情:\n")
		fmt.Printf("   账户名: %s\n", exchangeCfg.AccountName)
		fmt.Printf("   类型: %s\n", exchangeCfg.ExchangeType)
		fmt.Printf("   测试网: %t\n", exchangeCfg.Testnet)
		fmt.Printf("   自定义URL: %s\n", exchangeCfg.CustomAPIURL)
		fmt.Printf("   启用状态: %t\n", exchangeCfg.Enabled)

		// 解密API密钥
		if exchangeCfg.APIKey != "" {
			apiKey, err := cryptoService.DecryptFromStorage(string(exchangeCfg.APIKey))
			if err != nil {
				fmt.Printf("   API密钥解密失败: %v\n", err)
			} else {
				fmt.Printf("   API密钥: %s\n", apiKey)
			}
		}

		if exchangeCfg.SecretKey != "" {
			secretKey, err := cryptoService.DecryptFromStorage(string(exchangeCfg.SecretKey))
			if err != nil {
				fmt.Printf("   密钥解密失败: %v\n", err)
			} else {
				fmt.Printf("   密钥: %s\n", secretKey)
			}
		}
	} else {
		fmt.Printf("❌ 未找到交易所配置\n")
	}
}
