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

	// 查找特定交易员
	traderID := "75103af7_guardian-ai_1770569967"
	trader, err := s.Trader().GetByID(traderID)
	if err != nil {
		fmt.Printf("❌ 交易员查找失败: %v\n", err)
		// 列出所有交易员
		users, err := s.User().GetAllIDs()
		if err != nil {
			fmt.Printf("❌ 获取用户列表失败: %v\n", err)
			return
		}

		for _, userID := range users {
			user, err := s.User().GetByID(userID)
			if err != nil {
				continue
			}
			traders, err := s.Trader().FindByUserID(user.ID)
			if err != nil {
				continue
			}
			fmt.Printf("用户 %s 的交易员:\n", user.Email)
			for _, t := range traders {
				fmt.Printf("  - ID: %s, 名称: %s\n", t.ID, t.Name)
			}
		}
		return
	}

	fmt.Printf("✅ 找到交易员:\n")
	fmt.Printf("   ID: %s\n", trader.ID)
	fmt.Printf("   名称: %s\n", trader.Name)
	fmt.Printf("   交易所ID: %s\n", trader.ExchangeID)
	fmt.Printf("   AI模型ID: %s\n", trader.AIModelID)
	fmt.Printf("   策略ID: %s\n", trader.StrategyID)
	fmt.Printf("   启用状态: %t\n", trader.Enabled)

	// 检查关联的交易所
	if trader.ExchangeID != "" {
		exchange, err := s.Exchange().GetByID(trader.UserID, trader.ExchangeID)
		if err != nil {
			fmt.Printf("❌ 交易所查找失败: %v\n", err)
		} else {
			fmt.Printf("\n🔗 关联交易所:\n")
			fmt.Printf("   ID: %s\n", exchange.ID)
			fmt.Printf("   名称: %s\n", exchange.Name)
			fmt.Printf("   类型: %s\n", exchange.ExchangeType)
			fmt.Printf("   测试网: %t\n", exchange.Testnet)
			fmt.Printf("   自定义API URL: %s\n", exchange.CustomAPIURL)
			fmt.Printf("   API密钥状态: %s\n", func() string {
				if exchange.APIKey == "" {
					return "❌ 未设置"
				}
				return "✅ 已设置"
			}())
		}
	}
}
