package main

import (
	"fmt"
	"nofx/config"
	"nofx/store"
)

func main() {
	fmt.Println("🔍 检查交易所API配置")
	fmt.Println("====================")

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

	// 检查虚拟盘交易员的交易所
	virtualExchangeID := "bac26dfa-a4fa-43e6-8a63-ce492f6d2585"
	virtualExchange, err := s.Exchange().GetByID(virtualExchangeID)
	if err != nil {
		fmt.Printf("❌ 获取虚拟盘交易所失败: %v\n", err)
		return
	}

	fmt.Printf("虚拟盘交易所配置:\n")
	fmt.Printf("  ID: %s\n", virtualExchange.ID)
	fmt.Printf("  Name: %s\n", virtualExchange.Name)
	fmt.Printf("  Exchange Type: %s\n", virtualExchange.Exchange)
	fmt.Printf("  CustomAPIURL: '%s'\n", virtualExchange.CustomAPIURL)
	fmt.Printf("  Enabled: %t\n", virtualExchange.Enabled)
	fmt.Println()

	// 检查实盘交易员的交易所
	realExchangeID := "83faf7b3-526a-401d-be60-240d8293f5d5"
	realExchange, err := s.Exchange().GetByID(realExchangeID)
	if err != nil {
		fmt.Printf("❌ 获取实盘交易所失败: %v\n", err)
		return
	}

	fmt.Printf("实盘交易所配置:\n")
	fmt.Printf("  ID: %s\n", realExchange.ID)
	fmt.Printf("  Name: %s\n", realExchange.Name)
	fmt.Printf("  Exchange Type: %s\n", realExchange.Exchange)
	fmt.Printf("  CustomAPIURL: '%s'\n", realExchange.CustomAPIURL)
	fmt.Printf("  Enabled: %t\n", realExchange.Enabled)
	fmt.Println()

	// 分析问题
	fmt.Println("问题分析:")
	if virtualExchange.CustomAPIURL == "https://fapi.binance.com" {
		fmt.Println("❌ 虚拟盘交易员配置错误：使用了实盘API地址")
		fmt.Println("✅ 应该使用: https://testnet.binancefuture.com")
	} else if virtualExchange.CustomAPIURL == "https://testnet.binancefuture.com" {
		fmt.Println("✅ 虚拟盘交易员配置正确")
	} else {
		fmt.Printf("⚠️  虚拟盘交易员使用了未知的API地址: %s\n", virtualExchange.CustomAPIURL)
	}

	if realExchange.CustomAPIURL == "https://fapi.binance.com" {
		fmt.Println("✅ 实盘交易员配置正确")
	} else if realExchange.CustomAPIURL == "https://testnet.binancefuture.com" {
		fmt.Println("❌ 实盘交易员配置错误：使用了测试网API地址")
	} else {
		fmt.Printf("⚠️  实盘交易员使用了未知的API地址: %s\n", realExchange.CustomAPIURL)
	}
}
