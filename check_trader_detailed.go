package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 交易员状态检查工具")
	fmt.Println("====================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化加密服务
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatalf("❌ 加密服务初始化失败: %v", err)
	}

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 获取所有用户
	userIDs, err := s.User().GetAllIDs()
	if err != nil {
		log.Fatalf("❌ 获取用户列表失败: %v", err)
	}

	// 检查交易员配置
	traderIDs := []string{
		"83faf7b3_deepseek_1770656396",    // 实盘交易员
		"75103af7_guardian-ai_1770569967", // 虚拟盘交易员
	}

	fmt.Println("📋 交易员配置检查:")
	fmt.Println("=================")

	foundTraders := 0

	for _, userID := range userIDs {
		// 获取用户的所有交易员
		traders, err := s.Trader().List(userID)
		if err != nil {
			fmt.Printf("❌ 获取用户 %s 的交易员列表失败: %v\n", userID, err)
			continue
		}

		for _, trader := range traders {
			// 检查是否是我们要找的交易员
			for _, targetID := range traderIDs {
				if trader.ID == targetID {
					foundTraders++
					fmt.Printf("\n🔍 找到交易员: %s\n", trader.ID)
					fmt.Println("------------------------")
					fmt.Printf("   名称: %s\n", trader.Name)
					fmt.Printf("   用户ID: %s\n", trader.UserID)
					fmt.Printf("   AI模型ID: %s\n", trader.AIModelID)
					fmt.Printf("   交易所ID: %s\n", trader.ExchangeID)
					fmt.Printf("   策略ID: %s\n", trader.StrategyID)
					fmt.Printf("   初始余额: %.2f\n", trader.InitialBalance)
					fmt.Printf("   扫描间隔: %d 分钟\n", trader.ScanIntervalMinutes)
					fmt.Printf("   交叉保证金: %v\n", trader.IsCrossMargin)
					fmt.Printf("   运行状态: %v\n", trader.IsRunning)

					// 获取完整配置信息
					fullConfig, err := s.Trader().GetFullConfig(userID, trader.ID)
					if err != nil {
						fmt.Printf("   ❌ 获取完整配置失败: %v\n", err)
						continue
					}

					// 显示AI模型信息
					if fullConfig.AIModel != nil {
						fmt.Printf("   AI模型: %s (%s)\n", fullConfig.AIModel.Name, fullConfig.AIModel.Provider)
						fmt.Printf("   AI模型启用: %v\n", fullConfig.AIModel.Enabled)
					}

					// 显示交易所信息
					if fullConfig.Exchange != nil {
						fmt.Printf("   交易所: %s (%s)\n", fullConfig.Exchange.AccountName, fullConfig.Exchange.ExchangeType)
						fmt.Printf("   交易所启用: %v\n", fullConfig.Exchange.Enabled)
						fmt.Printf("   测试网: %v\n", fullConfig.Exchange.Testnet)
						if fullConfig.Exchange.CustomAPIURL != "" {
							fmt.Printf("   自定义API URL: %s\n", fullConfig.Exchange.CustomAPIURL)
						}

						// 解密并显示API密钥信息
						if string(fullConfig.Exchange.APIKey) != "" {
							decryptedAPIKey, err := cryptoService.DecryptFromStorage(string(fullConfig.Exchange.APIKey))
							if err != nil {
								fmt.Printf("   API密钥: 解密失败\n")
							} else {
								if len(decryptedAPIKey) >= 12 {
									fmt.Printf("   API密钥: %s...%s\n", decryptedAPIKey[:8], decryptedAPIKey[len(decryptedAPIKey)-4:])
								} else {
									fmt.Printf("   API密钥: %s\n", decryptedAPIKey)
								}
							}
						}
					}

					// 显示策略信息
					if fullConfig.Strategy != nil {
						fmt.Printf("   策略: %s\n", fullConfig.Strategy.Name)
					}

					// 分析问题原因
					fmt.Printf("\n   📊 问题分析:\n")
					if trader.ID == "83faf7b3_deepseek_1770656396" {
						// 实盘交易员 - API权限问题
						fmt.Printf("   ⚠️  实盘交易员可能因API密钥权限不足导致初始化失败\n")
						fmt.Printf("   🔧 建议检查:\n")
						fmt.Printf("      - API密钥是否具有交易权限\n")
						fmt.Printf("      - API密钥是否被禁用\n")
						fmt.Printf("      - IP白名单设置\n")
						fmt.Printf("      - API密钥长度是否符合要求\n")
					} else if trader.ID == "75103af7_guardian-ai_1770569967" {
						// 虚拟盘交易员 - 网络连接问题
						fmt.Printf("   ⚠️  虚拟盘交易员可能因网络连接问题导致500错误\n")
						fmt.Printf("   🔧 建议检查:\n")
						fmt.Printf("      - 网络连接状态\n")
						fmt.Printf("      - 代理服务配置\n")
						fmt.Printf("      - Binance测试网API可用性\n")
						fmt.Printf("      - 防火墙设置\n")
					}
				}
			}
		}
	}

	if foundTraders == 0 {
		fmt.Println("❌ 未找到指定的交易员配置")
		fmt.Println("可用的交易员:")
		for _, userID := range userIDs {
			traders, err := s.Trader().List(userID)
			if err != nil {
				continue
			}
			for _, trader := range traders {
				fmt.Printf("  - ID: %s, 名称: %s\n", trader.ID, trader.Name)
			}
		}
		return
	}

	fmt.Println("\n🔧 修复建议:")
	fmt.Println("============")
	fmt.Println("1. 实盘交易员修复:")
	fmt.Println("   - 登录Binance账户检查API密钥权限")
	fmt.Println("   - 确保API密钥具有读取账户信息权限")
	fmt.Println("   - 检查IP白名单设置")
	fmt.Println("   - 验证API密钥长度要求")
	fmt.Println("   - 重启后端服务重新加载配置")
	fmt.Println()
	fmt.Println("2. 虚拟盘交易员修复:")
	fmt.Println("   - 检查网络连接")
	fmt.Println("   - 验证代理服务状态")
	fmt.Println("   - 测试Binance测试网API连通性")
	fmt.Println("   - 检查防火墙设置")
	fmt.Println()
	fmt.Println("3. 通用解决方案:")
	fmt.Println("   - 重启后端服务")
	fmt.Println("   - 检查.env配置文件")
	fmt.Println("   - 验证数据库连接")
	fmt.Println("   - 查看详细日志信息")

	fmt.Println("\n📝 下一步操作:")
	fmt.Println("=============")
	fmt.Println("1. 根据上述分析检查对应的API密钥和网络配置")
	fmt.Println("2. 修正配置后重启后端服务")
	fmt.Println("3. 重新测试余额获取功能")
	fmt.Println("4. 如问题持续存在，查看详细日志进行深入诊断")
}
