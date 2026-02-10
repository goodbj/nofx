package main

import (
	"context"
	"fmt"
	"log"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"
	"nofx/trader"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 API调用对比测试")
	fmt.Println("========================")

	// 加载环境变量
	_ = godotenv.Load()
	config.Init()

	// 初始化加密服务
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatalf("加密服务初始化失败: %v", err)
	}

	// 初始化数据库
	s, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "data/data.db",
	})
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer s.Close()

	// 获取实盘01配置
	var exchange store.Exchange
	err = s.GormDB().Where("account_name = ? AND enabled = ?", "实盘01", true).First(&exchange).Error
	if err != nil {
		log.Fatalf("未找到实盘01配置: %v", err)
	}

	// 解密API密钥
	apiKey, err := cryptoService.DecryptFromStorage(string(exchange.APIKey))
	if err != nil {
		log.Fatalf("API密钥解密失败: %v", err)
	}

	secretKey, err := cryptoService.DecryptFromStorage(string(exchange.SecretKey))
	if err != nil {
		log.Fatalf("密钥解密失败: %v", err)
	}

	fmt.Printf("📋 测试配置:\n")
	fmt.Printf("   账户名称: %s\n", exchange.AccountName)
	fmt.Printf("   测试网: %t\n", exchange.Testnet)
	fmt.Printf("   自定义URL: '%s'\n", exchange.CustomAPIURL)
	fmt.Printf("   API密钥长度: %d\n", len(apiKey))
	fmt.Printf("   密钥长度: %d\n", len(secretKey))

	// 测试1: 直接使用Binance SDK（模拟稳定版方式）
	fmt.Println("\n=== 测试1: 直接使用Binance SDK ===")
	testDirectBinanceSDK(apiKey, secretKey, exchange.Testnet)

	// 测试2: 使用FuturesTrader（开发版方式）
	fmt.Println("\n=== 测试2: 使用FuturesTrader ===")
	testFuturesTrader(apiKey, secretKey, exchange.Testnet, exchange.CustomAPIURL)

	// 测试3: 使用AutoTrader配置方式
	fmt.Println("\n=== 测试3: 使用AutoTrader配置 ===")
	testAutoTrader(apiKey, secretKey, exchange.Testnet, exchange.CustomAPIURL)
}

func testDirectBinanceSDK(apiKey, secretKey string, testnet bool) {
	// 创建客户端
	client := futures.NewClient(apiKey, secretKey)

	// 设置端点
	if testnet {
		client.BaseURL = "https://testnet.binancefuture.com"
	} else {
		client.BaseURL = "https://fapi.binance.com"
	}

	// 同步服务器时间
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	serverTime, err := client.NewServerTimeService().Do(ctx)
	if err != nil {
		fmt.Printf("❌ 服务器时间同步失败: %v\n", err)
		return
	}

	now := time.Now().UnixMilli()
	offset := now - serverTime
	client.TimeOffset = offset
	fmt.Printf("✓ 服务器时间同步成功，偏移: %dms\n", offset)

	// 获取账户信息
	account, err := client.NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Printf("❌ 账户信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✓ 账户信息获取成功:\n")
	fmt.Printf("   总余额: %s\n", account.TotalWalletBalance)
	fmt.Printf("   可用余额: %s\n", account.AvailableBalance)
	fmt.Printf("   未实现盈亏: %s\n", account.TotalUnrealizedProfit)
}

func testFuturesTrader(apiKey, secretKey string, testnet bool, customURL string) {
	// 创建FuturesTrader
	var traderInstance *trader.FuturesTrader
	if customURL != "" {
		traderInstance = trader.NewFuturesTrader(apiKey, secretKey, "test_user", customURL)
	} else {
		endpoint := "https://fapi.binance.com"
		if testnet {
			endpoint = "https://testnet.binancefuture.com"
		}
		traderInstance = trader.NewFuturesTrader(apiKey, secretKey, "test_user", endpoint)
	}

	// 获取余额
	balance, err := traderInstance.GetBalance()
	if err != nil {
		fmt.Printf("❌ 余额获取失败: %v\n", err)
		return
	}

	fmt.Printf("✓ 余额获取成功:\n")
	fmt.Printf("   总余额: %.2f\n", balance["totalWalletBalance"])
	fmt.Printf("   可用余额: %.2f\n", balance["availableBalance"])
	fmt.Printf("   未实现盈亏: %.2f\n", balance["totalUnrealizedProfit"])
}

func testAutoTrader(apiKey, secretKey string, testnet bool, customURL string) {
	// 创建AutoTrader配置
	config := trader.AutoTraderConfig{
		ID:                  "test_trader",
		Name:                "测试交易员",
		AIModel:             "test",
		Exchange:            "binance",
		BinanceAPIKey:       apiKey,
		BinanceSecretKey:    secretKey,
		BinanceCustomAPIURL: customURL,
		ExchangeTestnet:     testnet,
		InitialBalance:      10000,
		IsCrossMargin:       true,
		ScanInterval:        5 * time.Minute,
	}

	// 创建AutoTrader
	autoTrader, err := trader.NewAutoTrader(config, nil, "test_user")
	if err != nil {
		fmt.Printf("❌ AutoTrader创建失败: %v\n", err)
		return
	}

	// 获取账户信息
	accountInfo, err := autoTrader.GetAccountInfo()
	if err != nil {
		fmt.Printf("❌ 账户信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✓ 账户信息获取成功:\n")
	fmt.Printf("   总净值: %.2f\n", accountInfo["total_equity"])
	fmt.Printf("   可用余额: %.2f\n", accountInfo["available_balance"])
	fmt.Printf("   总盈亏: %.2f\n", accountInfo["total_pnl"])
}
