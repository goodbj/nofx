package main

import (
	"fmt"
	"nofx/config"
	"nofx/manager"
	"nofx/store"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 交易员状态验证工具")
	fmt.Println("====================")

	// 加载环境变量
	_ = godotenv.Load()
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

	// 初始化交易员管理器
	traderManager := manager.NewTraderManager()

	// 获取所有用户并加载交易员
	userIDs, err := s.User().GetAllIDs()
	if err != nil {
		fmt.Printf("❌ 获取用户列表失败: %v\n", err)
		return
	}

	for _, userID := range userIDs {
		err = traderManager.LoadUserTradersFromStore(s, userID)
		if err != nil {
			fmt.Printf("❌ 加载用户 %s 的交易员失败: %v\n", userID, err)
			continue
		}
	}

	// 检查目标交易员
	targetTraders := []struct {
		id   string
		name string
	}{
		{"83faf7b3_deepseek_1770656396", "实盘交易员"},
		{"75103af7_guardian-ai_1770569967", "虚拟盘交易员"},
	}

	fmt.Println("\n📋 交易员内存状态检查:")
	fmt.Println("=====================")

	for _, tt := range targetTraders {
		_, err := traderManager.GetTrader(tt.id)
		if err != nil {
			fmt.Printf("❌ %s (%s) - 未加载到内存\n", tt.name, tt.id)
		} else {
			fmt.Printf("✅ %s (%s) - 已加载到内存\n", tt.name, tt.id)
		}
	}

	// 显示所有已加载的交易员
	fmt.Println("\n📊 所有已加载的交易员:")
	fmt.Println("===================")
	allTraders := traderManager.GetAllTraders()
	count := 0
	for id, _ := range allTraders {
		fmt.Printf("   • %s\n", id)
		count++
	}

	fmt.Printf("\n总计: %d 个交易员已加载到内存\n", count)

	fmt.Println("\n💡 如果实盘交易员仍未加载，请检查API密钥权限")
	fmt.Println("🔧 然后重启后端服务使配置生效")
}
