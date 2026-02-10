package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/store"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 查找所有交易员")
	fmt.Println("==================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 查询所有交易员
	var traders []store.Trader
	result := s.GormDB().Find(&traders)
	if result.Error != nil {
		log.Fatalf("查询交易员失败: %v", result.Error)
	}

	fmt.Printf("找到 %d 个交易员:\n", len(traders))
	fmt.Println("==================")

	for i, trader := range traders {
		fmt.Printf("%d. ID: %s\n", i+1, trader.ID)
		fmt.Printf("   名称: %s\n", trader.Name)
		fmt.Printf("   用户ID: %s\n", trader.UserID)
		fmt.Printf("   AI模型ID: %s\n", trader.AIModelID)
		fmt.Printf("   交易所ID: %s\n", trader.ExchangeID)
		fmt.Printf("   策略ID: %s\n", trader.StrategyID)
		fmt.Printf("   运行状态: %t\n", trader.IsRunning)
		fmt.Println("------------------")
	}
}
