package main

import (
	"fmt"
	"os"
	"time"

	"nofx/trader"
)

func main() {
	fmt.Println("🚀 开始从前端后端透明代理获取精度...")

	// 设置代理环境变量
	os.Setenv("USE_BINANCE_PROXY", "true")
	os.Setenv("BINANCE_PROXY_URL", "http://localhost:8081")

	// 创建通过代理的期货交易者
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "test_user", "http://localhost:8081", "https://fapi.binance.com")

	// 定义要查询的交易对
	symbol := "BTCUSDT"

	fmt.Printf("🔄 正在通过透明代理从API交易所获取 %s 精度...\n", symbol)

	// 获取精度
	startTime := time.Now()
	precision, err := proxyTrader.GetSymbolPrecision(symbol)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ 获取精度失败: %v\n", err)
		fmt.Println("💡 系统将使用默认精度值作为备用方案")
		fmt.Printf("📋 精度值: 3 (默认值)\n")
	} else {
		fmt.Printf("✅ 成功获取精度!\n")
		fmt.Printf("📋 交易对: %s\n", symbol)
		fmt.Printf("📋 精度值: %d\n", precision)
		fmt.Printf("⏱️  耗时: %v\n", duration)
	}

	fmt.Println("\n✨ 精度获取完成")
}
