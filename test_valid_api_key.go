package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	fmt.Println("🔍 Binance API密钥验证测试")
	fmt.Println("========================")

	// 使用已验证正确的API密钥
	apiKey := "61L4IvYjkiKoFGIiyhEgKChBcwuvDco1hP4cSMb8W8ZFHz9gO12oBvvcAVcY3jyG"
	secretKey := "E2AAFZdB7akVgEavRImX0i6RIsGG9TQ07fhgREGs2wVJWbFNUfP5BoNI0n0sYCqo"

	fmt.Printf("🔑 测试API密钥: %s\n", apiKey[:10]+"...")
	fmt.Printf("🔑 测试密钥: %s\n", secretKey[:10]+"...")

	// 创建Binance客户端
	client := futures.NewClient(apiKey, secretKey)
	client.BaseURL = "https://fapi.binance.com" // 主网

	// 设置超时
	client.HTTPClient.Timeout = 30 * time.Second

	fmt.Println("\n=== 测试1: 获取服务器时间 ===")
	testServerTime(client)

	fmt.Println("\n=== 测试2: 获取账户信息 ===")
	testAccountInfo(client)

	fmt.Println("\n=== 测试3: 获取持仓信息 ===")
	testPositions(client)
}

func testServerTime(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timeService := client.NewServerTimeService()
	serverTime, err := timeService.Do(ctx)
	if err != nil {
		fmt.Printf("❌ 服务器时间获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 服务器时间获取成功: %d\n", serverTime)
}

func testAccountInfo(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	accountService := client.NewGetAccountService()
	account, err := accountService.Do(ctx)
	if err != nil {
		fmt.Printf("❌ 账户信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 账户信息获取成功\n")
	fmt.Printf("   总资产: %.2f USDT\n", account.TotalWalletBalance)
	fmt.Printf("   可用余额: %.2f USDT\n", account.AvailableBalance)
	fmt.Printf("   未实现盈亏: %.2f USDT\n", account.TotalUnrealizedProfit)
	fmt.Printf("   持仓数量: %d\n", len(account.Positions))
}

func testPositions(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	positionService := client.NewGetPositionRiskService()
	positions, err := positionService.Do(ctx)
	if err != nil {
		fmt.Printf("❌ 持仓信息获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 持仓信息获取成功\n")
	fmt.Printf("   持仓数量: %d\n", len(positions))

	// 显示非零持仓
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
