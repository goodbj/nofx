package main

import (
	"fmt"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"
)

func main() {
	fmt.Println("🔧 创建测试交易所配置")
	fmt.Println("========================")

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

	// 检查是否已存在测试交易所
	exchanges, err := s.Exchange().List("default")
	if err != nil {
		fmt.Printf("❌ 获取交易所列表失败: %v\n", err)
		return
	}

	fmt.Printf("📊 当前交易所数量: %d\n", len(exchanges))

	// 如果没有交易所，创建一个测试交易所
	if len(exchanges) == 0 {
		testExchange := &store.Exchange{
			ID:           "test-exchange-001",
			UserID:       "default",
			Name:         "Binance Test Account",
			AccountName:  "Test Account",
			ExchangeType: "binance",
			Type:         "cex",
			Enabled:      true,
			Testnet:      false,
			APIKey:       crypto.EncryptedString("test-api-key"),
			SecretKey:    crypto.EncryptedString("test-secret-key"),
			CustomAPIURL: "",
		}

		if err := s.GormDB().Create(testExchange).Error; err != nil {
			fmt.Printf("❌ 创建测试交易所失败: %v\n", err)
			return
		}
		fmt.Println("✅ 测试交易所创建成功")
	} else {
		fmt.Println("✅ 已存在交易所配置")
		for i, ex := range exchanges {
			fmt.Printf("   %d. %s (%s) - %s\n", i+1, ex.Name, ex.ExchangeType, map[bool]string{true: "启用", false: "禁用"}[ex.Enabled])
		}
	}
}
