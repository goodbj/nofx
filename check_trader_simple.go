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

		// 获取该用户的所有交易员配置
		traderConfigs, err := s.Trader().GetFullConfigs(userID)
		if err != nil {
			continue
		}

		fmt.Printf("\n📋 用户: %s\n", user.Email)
		fmt.Printf("   交易员数量: %d\n", len(traderConfigs))

		for _, config := range traderConfigs {
			fmt.Printf("   - ID: %s, 名称: %s\n", config.Trader.ID, config.Trader.Name)

			if config.Trader.ID == targetTraderID {
				found = true
				fmt.Printf("\n🎯 找到目标交易员:\n")
				fmt.Printf("   ID: %s\n", config.Trader.ID)
				fmt.Printf("   名称: %s\n", config.Trader.Name)
				fmt.Printf("   AI模型: %s\n", config.Trader.AIModelID)
				fmt.Printf("   策略ID: %s\n", config.Trader.StrategyID)

				if config.Exchange != nil {
					fmt.Printf("   交易所: %s (%s)\n", config.Exchange.Name, config.Exchange.ExchangeType)
					fmt.Printf("   测试网: %t\n", config.Exchange.Testnet)
					fmt.Printf("   API密钥: %s\n", func() string {
						if config.Exchange.APIKey == "" {
							return "未设置"
						}
						return "已设置"
					}())
				}

				if config.Strategy != nil {
					fmt.Printf("   策略名称: %s\n", config.Strategy.Name)
				}
			}
		}
	}

	if !found {
		fmt.Printf("\n❌ 未找到交易员 %s\n", targetTraderID)
		fmt.Println("可用的交易员:")
		for _, userID := range userIDs {
			user, err := s.User().GetByID(userID)
			if err != nil {
				continue
			}
			traderConfigs, err := s.Trader().GetFullConfigs(userID)
			if err != nil {
				continue
			}
			for _, config := range traderConfigs {
				fmt.Printf("   - %s (%s)\n", config.Trader.Name, config.Trader.ID)
			}
		}
	}
}
