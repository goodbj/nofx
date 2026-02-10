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
	fmt.Println("🔍 简化API调用测试")
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
	fmt.Printf("   API密钥: %s\n", apiKey[:10]+"...")
	fmt.Printf("   密钥: %s\n", secretKey[:10]+"...")

	// 测试直接使用Binance SDK
	fmt.Println("\n=== 测试Binance SDK ===")
	testBinanceSDK(apiKey, secretKey, exchange.Testnet, exchange.CustomAPIURL)
}

func testBinanceSDK(apiKey, secretKey string, testnet bool, customURL string) {
	fmt.Println("创建Binance客户端...")
	
	// 创建客户端
	client := futures.NewClient(apiKey, secretKey)
	
	// 设置端点
	endpoint := "https://fapi.binance.com"
	if testnet {
		endpoint = "https://testnet.binancefuture.com"
	}
	if customURL != "" {
		endpoint = customURL
	}
	client.BaseURL = endpoint
	
	fmt.Printf("使用端点: %s\n", endpoint)

	// 设置超时
	client.HTTPClient.Timeout = 30 * time.Second

	fmt.Println("同步服务器时间...")
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

	fmt.Println("获取账户信息...")
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
	fmt.Printf("   资产数量: %d\n", len(account.Assets))
	
	// 显示前几个资产
	for i, asset := range account.Assets {
		if i >= 3 {
			break
		}
		fmt.Printf("   资产%d: %s - 余额: %s, 未实现盈亏: %s\n", 
			i+1, asset.Asset, asset.WalletBalance, asset.UnrealizedProfit)
	}
}