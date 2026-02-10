package main

import (
	"fmt"
	"log"
	"nofx/config"
	"nofx/manager"
	"nofx/store"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔧 交易员内存加载修复工具")
	fmt.Println("========================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 初始化交易员管理器
	traderManager := manager.NewTraderManager()

	// 获取所有用户
	userIDs, err := s.User().GetAllIDs()
	if err != nil {
		log.Fatalf("❌ 获取用户列表失败: %v", err)
	}

	// 需要修复的交易员
	traderIDs := []string{
		"83faf7b3_deepseek_1770656396",    // 实盘交易员
		"75103af7_guardian-ai_1770569967", // 虚拟盘交易员
	}

	fmt.Println("📋 开始修复交易员内存加载问题:")
	fmt.Println("===========================")

	for _, userID := range userIDs {
		// 从数据库加载所有用户交易员到内存
		fmt.Printf("🔄 为用户 %s 加载所有交易员到内存...\n", userID)
		err = traderManager.LoadUserTradersFromStore(s, userID)
		if err != nil {
			fmt.Printf("❌ 加载用户 %s 的交易员失败: %v\n", userID, err)
			continue
		}

		// 检查特定交易员是否已加载
		for _, targetTraderID := range traderIDs {
			_, err := traderManager.GetTrader(targetTraderID)
			if err != nil {
				fmt.Printf("❌ 交易员 %s 仍未加载到内存\n", targetTraderID)

				// 尝试单独加载该交易员
				fullConfig, configErr := s.Trader().GetFullConfig(userID, targetTraderID)
				if configErr != nil {
					fmt.Printf("   ❌ 获取交易员配置失败: %v\n", configErr)
					continue
				}

				fmt.Printf("   ✅ 交易员配置存在，尝试手动加载...\n")
				fmt.Printf("   - 名称: %s\n", fullConfig.Trader.Name)
				fmt.Printf("   - 交易所: %s\n", fullConfig.Exchange.AccountName)
				fmt.Printf("   - 测试网: %v\n", fullConfig.Exchange.Testnet)

				// 尝试重新加载用户的所有交易员
				loadErr := traderManager.LoadUserTradersFromStore(s, userID)
				if loadErr != nil {
					fmt.Printf("   ❌ 重新加载用户交易员失败: %v\n", loadErr)
					fmt.Printf("   💡 建议: 检查API密钥权限和网络连接\n")
				} else {
					// 再次检查交易员是否已加载
					_, err := traderManager.GetTrader(targetTraderID)
					if err != nil {
						fmt.Printf("   ❌ 交易员仍未能加载到内存，可能存在API密钥或网络问题\n")
						fmt.Printf("   💡 建议: 检查API密钥权限和网络连接\n")
					} else {
						fmt.Printf("   ✅ 交易员已成功加载到内存\n")
					}
				}
			} else {
				fmt.Printf("✅ 交易员 %s 已成功加载到内存\n", targetTraderID)
			}
		}
	}

	// 输出当前内存中的所有交易员
	fmt.Println("\n📊 内存中当前交易员列表:")
	fmt.Println("======================")
	allTraders := traderManager.GetAllTraders()
	if len(allTraders) == 0 {
		fmt.Println("   暂无交易员加载到内存")
	} else {
		for traderID, _ := range allTraders {
			fmt.Printf("   - %s\n", traderID)
		}
	}

	fmt.Println("\n✅ 修复流程完成!")
	fmt.Println("💡 提示: 重启后端服务将重新加载所有交易员配置")
	fmt.Println("🔧 如果问题持续存在，请检查API密钥权限和网络连接")
}
