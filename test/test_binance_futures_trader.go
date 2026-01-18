package main

import (
	"fmt"
	"os"

	"nofx/trader"
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
		fmt.Println("BINANCE_API_KEY and BINANCE_SECRET_KEY environment variables must be set")
		return
	}

	// 显示API密钥的部分内容（出于安全考虑）
	keyDisplayLen := 8
	if len(apiKey) < keyDisplayLen {
		keyDisplayLen = len(apiKey)
	}
	fmt.Println("🚀 Starting Binance Futures Trader Testnet Connection Test...")
	fmt.Printf("Using API Key: %s...\n", apiKey[:keyDisplayLen]) // 只显示前几位

	// 创建币安期货交易者，使用测试网端点
	// 注意：使用自定义端点参数传入测试网URL
	traderInstance := trader.NewFuturesTrader(apiKey, secretKey, "test-user", "https://testnet.binancefuture.com")

	fmt.Println("🔗 Connected using system's FuturesTrader with testnet endpoint")

	// 测试获取账户余额
	fmt.Println("\n💳 Testing account balance...")
	balance, err := traderInstance.GetBalance()
	if err != nil {
		fmt.Printf("❌ Error getting balance: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully got account balance!\n")
	fmt.Printf("💰 Total wallet balance: %.2f USDT\n", balance["totalWalletBalance"])
	fmt.Printf("💵 Available balance: %.2f USDT\n", balance["availableBalance"])
	fmt.Printf("📊 Unrealized profit: %.2f USDT\n", balance["totalUnrealizedProfit"])

	// 测试获取持仓信息
	fmt.Println("\n📊 Testing positions...")
	positions, err := traderInstance.GetPositions()
	if err != nil {
		fmt.Printf("❌ Error getting positions: %v\n", err)
	} else {
		fmt.Printf("✅ Got %d positions\n", len(positions))
		if len(positions) > 0 {
			for i, pos := range positions {
				if i >= 5 { // 只显示前5个
					fmt.Printf("  ... and %d more positions\n", len(positions)-5)
					break
				}
				side := "📊"
				if pos["side"] == "long" {
					side = "📈 LONG"
				} else {
					side = "📉 SHORT"
				}
				fmt.Printf("  %s %s: Amount=%.4f, Entry Price=%.4f, PnL=%.4f\n",
					side,
					pos["symbol"],
					pos["positionAmt"],
					pos["entryPrice"],
					pos["unRealizedProfit"])
			}
		} else {
			fmt.Println("  📭 No active positions")
		}
	}

	// 测试获取市场价格
	fmt.Println("\n💹 Testing market prices (BTCUSDT)...")
	price, err := traderInstance.GetMarketPrice("BTCUSDT")
	if err != nil {
		fmt.Printf("❌ Error getting BTCUSDT price: %v\n", err)
	} else {
		fmt.Printf("✅ BTCUSDT price: %.2f\n", price)
	}

	fmt.Println("\n🎉 Binance Futures Trader Testnet connection test completed successfully!")
}