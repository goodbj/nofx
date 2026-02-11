package main

import (
	"context"
	"fmt"
	"nofx/store"
	"nofx/trader"
	"time"
)

func main() {
	fmt.Println("🔍 比较实盘vs虚拟盘API调用")
	fmt.Println("========================")

	// Initialize store
	st, err := store.NewStore()
	if err != nil {
		fmt.Printf("❌ Failed to initialize store: %v\n", err)
		return
	}
	defer st.Close()

	// Get exchanges
	exchanges, err := st.Exchange().List("")
	if err != nil {
		fmt.Printf("❌ Failed to get exchanges: %v\n", err)
		return
	}

	// Find real and testnet exchanges
	var realExchange, testnetExchange *store.Exchange
	for _, exchange := range exchanges {
		if exchange.Exchange == "binance" && exchange.Enabled {
			if exchange.CustomAPIURL == "https://fapi.binance.com" {
				realExchange = exchange
				fmt.Printf("✅ 找到实盘交易所: %s (ID: %s)\n", exchange.Name, exchange.ID[:8])
			} else if exchange.CustomAPIURL == "https://testnet.binancefuture.com" {
				testnetExchange = exchange
				fmt.Printf("✅ 找到虚拟盘交易所: %s (ID: %s)\n", exchange.Name, exchange.ID[:8])
			}
		}
	}

	if realExchange == nil || testnetExchange == nil {
		fmt.Println("❌ 未找到完整的实盘/虚拟盘配置")
		return
	}

	// Test real exchange
	fmt.Println("\n🧪 测试实盘API调用...")
	testExchangeAPI(realExchange, "实盘")

	// Test testnet exchange
	fmt.Println("\n🧪 测试虚拟盘API调用...")
	testExchangeAPI(testnetExchange, "虚拟盘")

	fmt.Println("\n✅ 测试完成")
}

func testExchangeAPI(exchange *store.Exchange, name string) {
	// Decrypt API keys
	apiKey := string(exchange.APIKey)
	secretKey := string(exchange.SecretKey)

	fmt.Printf("  交易所: %s\n", exchange.Name)
	fmt.Printf("  API URL: %s\n", exchange.CustomAPIURL)
	fmt.Printf("  API密钥长度: %d\n", len(apiKey))
	fmt.Printf("  密钥长度: %d\n", len(secretKey))

	// Create trader instance
	var traderInstance *trader.FuturesTrader
	if exchange.CustomAPIURL != "" {
		traderInstance = trader.NewFuturesTrader(apiKey, secretKey, "test_user", exchange.CustomAPIURL)
	} else {
		traderInstance = trader.NewFuturesTrader(apiKey, secretKey, "test_user", "")
	}

	// Test server time
	fmt.Printf("  🕐 测试服务器时间同步...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverTime, err := traderInstance.Client().NewServerTimeService().Do(ctx)
	if err != nil {
		fmt.Printf(" ❌ 失败: %v\n", err)
		return
	}
	fmt.Printf(" ✅ 成功 (时间: %d)\n", serverTime)

	// Test account info
	fmt.Printf("  💰 测试账户信息获取...")
	account, err := traderInstance.Client().NewGetAccountService().Do(ctx)
	if err != nil {
		fmt.Printf(" ❌ 失败: %v\n", err)
		return
	}
	fmt.Printf(" ✅ 成功\n")
	fmt.Printf("    总余额: %s\n", account.TotalWalletBalance)
	fmt.Printf("    可用余额: %s\n", account.AvailableBalance)
	fmt.Printf("    未实现盈亏: %s\n", account.TotalUnrealizedProfit)
}
