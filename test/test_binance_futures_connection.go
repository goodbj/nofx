package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	// 从环境变量获取API凭据
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")

	// 如果环境变量未设置，使用用户提供的测试密钥
	if apiKey == "" {
		apiKey = "53zcowwoNVUiRWyg46RQb2nRwSQhUNwYeBEHetyJsjB0MLms1rqxABxWjXClkYoA"
	}
	if secretKey == "" {
		secretKey = "ZkvgnmrJiyOjuix0QxSa1eeQBwvlBENcwY5xBVp9uS085oDUpQHQcxHNfk0dppuj"
	}

	if apiKey == "" || secretKey == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_SECRET_KEY environment variables must be set")
	}

	// 显示API密钥的部分内容（出于安全考虑）
	keyDisplayLen := 8
	if len(apiKey) < keyDisplayLen {
		keyDisplayLen = len(apiKey)
	}
	fmt.Println("🚀 Starting Binance Futures Testnet Connection Test...")
	fmt.Printf("Using API Key: %s...\n", apiKey[:keyDisplayLen]) // 只显示前几位

	// 创建币安期货客户端，使用测试网端点
	client := futures.NewClient(apiKey, secretKey)
	client.BaseURL = "https://testnet.binancefuture.com" // 使用币安测试网

	fmt.Println("🔗 Connecting to Binance Testnet at:", client.BaseURL)

	// 测试获取服务器时间
	fmt.Println("\n🕐 Testing server time...")
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ Error getting server time: %v\n", err)
		return
	}
	fmt.Printf("✅ Server time: %d\n", serverTime)

	// 测试获取账户信息
	fmt.Println("\n💳 Testing account info...")
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ Error getting account info: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully connected to Binance Testnet!\n")
	fmt.Printf("💰 Total wallet balance: %s USDT\n", account.TotalWalletBalance)
	fmt.Printf("💵 Available balance: %s USDT\n", account.AvailableBalance)
	fmt.Printf("📊 Unrealized profit: %s USDT\n", account.TotalUnrealizedProfit)
	fmt.Printf("📈 Number of assets: %d\n", len(account.Assets))
	fmt.Printf("📦 Number of positions: %d\n", len(account.Positions))

	// 显示资产信息
	fmt.Println("\n💼 Assets:")
	for i, asset := range account.Assets {
		if i >= 10 { // 只显示前10个
			fmt.Printf("  ... and %d more assets\n", len(account.Assets)-10)
			break
		}
		fmt.Printf("  - %s: Wallet=%s, Available=%s, Unrealized PnL=%s\n",
			asset.Asset,
			asset.WalletBalance,
			asset.AvailableBalance,
			asset.UnrealizedProfit)
	}

	// 显示持仓信息（使用GetPositionRiskService而不是直接访问account.Positions）
	fmt.Println("\n📊 Positions:")
	positions, err := client.NewGetPositionRiskService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ Error getting positions: %v\n", err)
	} else {
		activePositions := 0
		for _, position := range positions {
			posAmtFloat, _ := strconv.ParseFloat(position.PositionAmt, 64)
			if posAmtFloat != 0 { // 只显示有持仓的
				activePositions++
				side := "📊"
				if posAmtFloat > 0 {
					side = "📈 LONG"
				} else {
					side = "📉 SHORT"
				}
				fmt.Printf("  %s %s: Amount=%s, Entry Price=%s, Mark Price=%s, Unrealized PnL=%s\n",
					side,
					position.Symbol,
					position.PositionAmt,
					position.EntryPrice,
					position.MarkPrice,
					position.UnRealizedProfit)
			}
		}
		if activePositions == 0 {
			fmt.Println("  📭 No active positions")
		}
	}

	// 测试获取订单信息
	fmt.Println("\n📋 Testing open orders...")
	openOrders, err := client.NewListOpenOrdersService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ Error getting open orders: %v\n", err)
	} else {
		fmt.Printf("✅ Open orders: %d\n", len(openOrders))
		for i, order := range openOrders {
			if i >= 5 { // 只显示前5个
				fmt.Printf("  ... and %d more orders\n", len(openOrders)-5)
				break
			}
			side := "📊"
			if order.Side == "BUY" {
				side = "📈 BUY"
			} else {
				side = "📉 SELL"
			}
			fmt.Printf("  - %s %s: ID=%d, Type=%s, Price=%s, Qty=%s, Side=%s\n",
				side,
				order.Symbol,
				order.OrderID,
				order.Type,
				order.Price,
				order.OrigQuantity,
				order.Side)
		}
	}

	// 测试获取价格
	fmt.Println("\n💹 Testing market prices...")
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ Error getting prices: %v\n", err)
	} else {
		fmt.Printf("✅ Got %d prices\n", len(prices))
		// 显示前几个主要交易对的价格
		pairsToShow := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
		for _, pair := range pairsToShow {
			for _, price := range prices {
				if price.Symbol == pair {
					fmt.Printf("  🪙 %s: %s\n", price.Symbol, price.Price)
					break
				}
			}
		}
	}

	fmt.Println("\n🎉 Binance Futures Testnet connection test completed successfully!")
}