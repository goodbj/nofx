package main

import (
	"context"
	"fmt"
	"log"
	"nofx/config"
	"nofx/crypto"
	"nofx/store"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔍 深度API调用链路调试")
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

	// 获取实盘01配置（应该工作正常的配置）
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

	fmt.Printf("📋 交易所配置详情:\n")
	fmt.Printf("   账户名: %s\n", exchange.AccountName)
	fmt.Printf("   测试网: %t\n", exchange.Testnet)
	fmt.Printf("   自定义URL: %s\n", exchange.CustomAPIURL)
	fmt.Printf("   API密钥: %s\n", apiKey)
	fmt.Printf("   密钥: %s\n", secretKey)

	// 创建Binance客户端进行深度调试
	fmt.Println("\n=== 深度API调用调试 ===")

	client := futures.NewClient(apiKey, secretKey)
	if exchange.CustomAPIURL != "" {
		client.BaseURL = exchange.CustomAPIURL
	} else if exchange.Testnet {
		client.BaseURL = "https://testnet.binancefuture.com"
	} else {
		client.BaseURL = "https://fapi.binance.com"
	}

	// 设置详细的HTTP客户端配置用于调试
	client.HTTPClient.Timeout = 30 * time.Second

	fmt.Printf("🌐 连接端点: %s\n", client.BaseURL)

	// 测试1: 服务器时间（不需要签名）
	fmt.Println("\n--- 测试1: 服务器时间 ---")
	testServerTime(client)

	// 测试2: 账户信息（需要签名）
	fmt.Println("\n--- 测试2: 账户信息 ---")
	testAccountInfo(client)

	// 测试3: 持仓信息（需要签名）
	fmt.Println("\n--- 测试3: 持仓信息 ---")
	testPositions(client)
}

func testServerTime(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	service := client.NewServerTimeService()

	startTime := time.Now()
	serverTime, err := service.Do(ctx)
	endTime := time.Now()

	fmt.Printf("请求耗时: %v\n", endTime.Sub(startTime))

	if err != nil {
		fmt.Printf("❌ 服务器时间获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 服务器时间获取成功: %d\n", serverTime)
}

func testAccountInfo(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	service := client.NewGetAccountService()

	startTime := time.Now()
	account, err := service.Do(ctx)
	endTime := time.Now()

	fmt.Printf("请求耗时: %v\n", endTime.Sub(startTime))

	if err != nil {
		fmt.Printf("❌ 账户信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 账户信息获取成功\n")
	fmt.Printf("   总资产: %.2f USDT\n", account.TotalWalletBalance)
	fmt.Printf("   可用余额: %.2f USDT\n", account.AvailableBalance)
	fmt.Printf("   持仓数量: %d\n", len(account.Positions))
}

func testPositions(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	service := client.NewGetPositionRiskService()

	startTime := time.Now()
	positions, err := service.Do(ctx)
	endTime := time.Now()

	fmt.Printf("请求耗时: %v\n", endTime.Sub(startTime))

	if err != nil {
		fmt.Printf("❌ 持仓信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 持仓信息获取成功\n")
	fmt.Printf("   持仓数量: %d\n", len(positions))

	nonZeroPositions := 0
	for _, pos := range positions {
		if pos.PositionAmt != "0" {
			nonZeroPositions++
			fmt.Printf("   %s: %s (%s)\n", pos.Symbol, pos.PositionAmt, pos.EntryPrice)
		}
	}

	if nonZeroPositions == 0 {
		fmt.Printf("   当前无持仓\n")
	}
}
