package main

import (
	"fmt"
	"os"
	"time"

	"nofx/trader"
)

// ProxyPrecisionTest 代理精度获取测试
type ProxyPrecisionTest struct {
	testSymbol string
}

// NewProxyPrecisionTest 创建新的代理精度测试实例
func NewProxyPrecisionTest(symbol string) *ProxyPrecisionTest {
	return &ProxyPrecisionTest{
		testSymbol: symbol,
	}
}

// Run 执行代理精度获取测试
func (ppt *ProxyPrecisionTest) Run() {
	fmt.Println("🔍 开始代理精度获取测试...")
	fmt.Printf("📊 测试交易对: %s\n", ppt.testSymbol)

	// 设置代理环境变量
	os.Setenv("USE_BINANCE_PROXY", "true")
	os.Setenv("BINANCE_PROXY_URL", "http://localhost:8081")

	fmt.Println("\n=== 通过代理获取精度 ===")

	// 创建通过代理的期货交易者
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "test_user", "http://localhost:8081", "https://fapi.binance.com")

	fmt.Printf("🔄 通过代理获取 %s 的精度信息...\n", ppt.testSymbol)

	// 通过代理获取精度
	startTime := time.Now()
	precision, err := proxyTrader.GetSymbolPrecision(ppt.testSymbol)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ 通过代理获取精度失败: %v\n", err)
	} else {
		fmt.Printf("✅ 通过代理获取精度成功: %d (耗时: %v)\n", precision, duration)
	}

	// 跳过PrecisionManager直接访问测试（字段未导出）
	fmt.Println("⚠️ 跳过PrecisionManager直接访问测试 (precisionManager字段未导出)")
	fmt.Println("💡 实际应用中，精度管理器会在后台自动工作")

	// 测试实盘和模拟盘的不同目标URL
	fmt.Println("\n=== 测试不同目标URL (实盘vs模拟盘) ===")

	// 实盘测试
	fmt.Println("🔄 测试实盘API (https://fapi.binance.com)...")
	realTrader := trader.NewFuturesTraderViaProxy("", "", "test_user", "http://localhost:8081", "https://fapi.binance.com")
	realPrecision, err := realTrader.GetSymbolPrecision(ppt.testSymbol)
	if err != nil {
		fmt.Printf("   ❌ 实盘API获取精度失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 实盘API获取精度成功: %d\n", realPrecision)
	}

	// 模拟盘测试
	fmt.Println("🔄 测试模拟盘API (https://testnet.binancefuture.com)...")
	testnetTrader := trader.NewFuturesTraderViaProxy("", "", "test_user", "http://localhost:8081", "https://testnet.binancefuture.com")
	testnetPrecision, err := testnetTrader.GetSymbolPrecision(ppt.testSymbol)
	if err != nil {
		fmt.Printf("   ❌ 模拟盘API获取精度失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 模拟盘API获取精度成功: %d\n", testnetPrecision)
	}
}

// RunProxyPrecisionTest 运行代理精度测试的便捷函数
func RunProxyPrecisionTest() {
	test := NewProxyPrecisionTest("BTCUSDT")
	test.Run()
}

func main() {
	fmt.Println("🚀 开始代理精度获取测试")
	RunProxyPrecisionTest()
	fmt.Println("\n✅ 代理精度获取测试完成")
}
