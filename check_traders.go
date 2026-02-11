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

	// 获取所有用户ID
	userIDs, err := s.User().GetAllIDs()
	if err != nil {
		fmt.Printf("❌ 获取用户ID列表失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 个用户\n", len(userIDs))

	for _, userID := range userIDs {
		fmt.Printf("\n📋 用户ID: %s\n", userID)

		// 加载该用户的所有交易员
		traders, err := s.Trader().List(userID)
		if err != nil {
			fmt.Printf("   ❌ 获取交易员失败: %v\n", err)
			continue
		}

		fmt.Printf("   🤖 该用户有 %d 个交易员:\n", len(traders))
		for _, trader := range traders {
			fmt.Printf("     - ID: %s\n", trader.ID)
			fmt.Printf("       Name: %s\n", trader.Name)
			fmt.Printf("       AI Model: %s\n", trader.AIModelID)
			fmt.Printf("       Exchange ID: %s\n", trader.ExchangeID)
			fmt.Printf("       Strategy ID: %s\n", trader.StrategyID)
			fmt.Printf("       Enabled: %t\n", trader.IsRunning) // 使用IsRunning字段
			fmt.Println()
		}
	}
}
