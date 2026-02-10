package main

import (
	"fmt"
	"nofx/config"
	"nofx/store"
)

func main() {
	fmt.Println("🔍 检查交易员配置")
	fmt.Println("==================")

	// 初始化配置
	config.Init()

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		return
	}
	defer s.Close()

	// 获取所有用户
	userIDs, err := s.User().GetAllIDs()
	if err != nil {
		fmt.Printf("❌ 获取用户列表失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 个用户\n", len(userIDs))

	// 查找目标交易员
	targetTraderID := "75103af7_guardian-ai_1770569967"
	found := false

	for _, userID := range userIDs {
		user, err := s.User().GetByID(userID)
		if err != nil {
			continue
		}

		// 获取该用户的所有交易员
		traders, err := s.Trader().List(userID)
		if err != nil {
			continue
		}

		fmt.Printf("\n📋 用户: %s\n", user.Email)
		fmt.Printf("   交易员数量: %d\n", len(traders))

		for _, trader := range traders {
			fmt.Printf("   - ID: %s, 名称: %s\n", trader.ID, trader.Name)

			if trader.ID == targetTraderID {
				found = true
				fmt.Printf("\n🎯 找到目标交易员:\n")
				fmt.Printf("   ID: %s\n", trader.ID)
				fmt.Printf("   名称: %s\n", trader.Name)
				fmt.Printf("   AI模型ID: %s\n", trader.AIModelID)
				fmt.Printf("   交易所ID: %s\n", trader.ExchangeID)
				fmt.Printf("   策略ID: %s\n", trader.StrategyID)
				fmt.Printf("   初始余额: %.2f\n", trader.InitialBalance)
				fmt.Printf("   扫描间隔: %d分钟\n", trader.ScanIntervalMinutes)
				fmt.Printf("   运行状态: %t\n", trader.IsRunning)

				// 获取完整配置
				fullConfig, err := s.Trader().GetFullConfig(userID, trader.ID)
				if err != nil {
					fmt.Printf("   ❌ 获取完整配置失败: %v\n", err)
					continue
				}

				if fullConfig.Exchange != nil {
					fmt.Printf("\n🔗 关联交易所:\n")
					fmt.Printf("   名称: %s\n", fullConfig.Exchange.Name)
					fmt.Printf("   类型: %s\n", fullConfig.Exchange.ExchangeType)
					fmt.Printf("   测试网: %t\n", fullConfig.Exchange.Testnet)
					fmt.Printf("   自定义API URL: %s\n", fullConfig.Exchange.CustomAPIURL)
					fmt.Printf("   API密钥状态: %s\n", func() string {
						if fullConfig.Exchange.APIKey == "" {
							return "❌ 未设置"
						}
						return "✅ 已设置"
					}())
				}

				if fullConfig.Strategy != nil {
					fmt.Printf("\n📝 关联策略:\n")
					fmt.Printf("   名称: %s\n", fullConfig.Strategy.Name)
					fmt.Printf("   描述: %s\n", fullConfig.Strategy.Description)
				}

				if fullConfig.AIModel != nil {
					fmt.Printf("\n🤖 关联AI模型:\n")
					fmt.Printf("   名称: %s\n", fullConfig.AIModel.Name)
					fmt.Printf("   提供商: %s\n", fullConfig.AIModel.Provider)
				}
			}
		}
	}

	if !found {
		fmt.Printf("\n❌ 未找到交易员 %s\n", targetTraderID)
		fmt.Println("可用的交易员:")
		for _, userID := range userIDs {
			traders, err := s.Trader().List(userID)
			if err != nil {
				continue
			}
			for _, trader := range traders {
				fmt.Printf("   - %s (%s)\n", trader.Name, trader.ID)
			}
		}
	}
}
