package main

import (
	"fmt"
	"nofx/config"
	"nofx/store"
)

func main() {
	fmt.Println("🔍 检查交易所配置")
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

	// 查找目标交易所
	exchangeID := "bac26dfa-a4fa-43e6-8a63-ce492f6d2585"

	exchanges, err := s.Exchange().List("e8e9da5d-c213-4111-b12d-96a0c5875385")
	if err != nil {
		fmt.Printf("❌ 获取交易所列表失败: %v\n", err)
		return
	}

	fmt.Printf("📊 找到 %d 个交易所配置\n", len(exchanges))

	for _, exchange := range exchanges {
		fmt.Printf("\n📋 交易所: %s\n", exchange.Name)
		fmt.Printf("   ID: %s\n", exchange.ID)
		fmt.Printf("   类型: %s\n", exchange.ExchangeType)
		fmt.Printf("   测试网: %t\n", exchange.Testnet)
		fmt.Printf("   自定义API URL: %s\n", exchange.CustomAPIURL)
		fmt.Printf("   启用状态: %t\n", exchange.Enabled)
		fmt.Printf("   API密钥长度: %d\n", len(exchange.APIKey))
		fmt.Printf("   密钥长度: %d\n", len(exchange.SecretKey))

		if exchange.ID == exchangeID {
			fmt.Printf("\n🎯 目标交易所配置详情:\n")
			fmt.Printf("   账户名称: %s\n", exchange.AccountName)
			fmt.Printf("   API密钥(前10位): %s\n", string(exchange.APIKey)[:10])
			fmt.Printf("   密钥(前10位): %s\n", string(exchange.SecretKey)[:10])
		}
	}
}
