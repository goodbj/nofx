package main

import (
	"fmt"
	"os"
	"time"

	"nofx/trader"
)

// PrecisionSimulationDemo 精度获取模拟演示
type PrecisionSimulationDemo struct {
	testSymbol string
}

// NewPrecisionSimulationDemo 创建新的精度模拟演示实例
func NewPrecisionSimulationDemo(symbol string) *PrecisionSimulationDemo {
	return &PrecisionSimulationDemo{
		testSymbol: symbol,
	}
}

// Run 执行精度获取模拟演示
func (psd *PrecisionSimulationDemo) Run() {
	fmt.Println("🚀 前端后端已运行，开始模拟获取精度...")
	fmt.Printf("📊 测试交易对: %s\n", psd.testSymbol)

	// 验证代理服务状态
	psd.verifyProxyService()

	// 演示不同的精度获取方式
	psd.demoDirectAPICall()
	psd.demoProxyModeCall()
	psd.demoPrecisionManagerCall()

	// 演示实盘和模拟盘的精度获取
	psd.demoRealAndTestnetPrecision()

	fmt.Println("\n✅ 精度获取模拟演示完成!")
	fmt.Println("💡 以上演示展示了在前端后端运行情况下，如何通过不同方式获取交易对精度")
}

// verifyProxyService 验证代理服务状态
func (psd *PrecisionSimulationDemo) verifyProxyService() {
	fmt.Println("\n🔍 验证代理服务状态...")

	// 检查代理环境变量
	useProxy := os.Getenv("USE_BINANCE_PROXY")
	proxyURL := os.Getenv("BINANCE_PROXY_URL")

	if useProxy == "true" || useProxy == "1" {
		fmt.Printf("✅ 代理模式已启用: %s\n", proxyURL)
	} else {
		fmt.Printf("⚠️ 代理模式未启用，当前设置: USE_BINANCE_PROXY=%s\n", useProxy)
		fmt.Printf("💡 如需启用代理模式，请设置环境变量: USE_BINANCE_PROXY=true BINANCE_PROXY_URL=http://localhost:8081\n")
	}
}

// demoDirectAPICall 演示直接API调用
func (psd *PrecisionSimulationDemo) demoDirectAPICall() {
	fmt.Println("\n=== 演示1: 直接API调用获取精度 ===")
	fmt.Printf("🔄 尝试直接从Binance API获取 %s 精度...\n", psd.testSymbol)

	// 创建期货交易者（直接连接）
	trader := trader.NewFuturesTrader("", "", "demo_user", "")

	startTime := time.Now()
	precision, err := trader.GetSymbolPrecision(psd.testSymbol)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("⚠️ 直接API调用失败: %v (耗时: %v)\n", err, duration)
		fmt.Println("💡 这可能是由于网络限制或API访问问题，代理模式可以解决此类问题")
	} else {
		fmt.Printf("✅ 直接API调用成功: 精度=%d (耗时: %v)\n", precision, duration)
	}
}

// demoProxyModeCall 演示代理模式调用
func (psd *PrecisionSimulationDemo) demoProxyModeCall() {
	fmt.Println("\n=== 演示2: 代理模式调用获取精度 ===")
	fmt.Printf("🔄 尝试通过代理获取 %s 精度...\n", psd.testSymbol)

	// 创建通过代理的期货交易者
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "demo_user", "http://localhost:8081", "https://fapi.binance.com")

	startTime := time.Now()
	precision, err := proxyTrader.GetSymbolPrecision(psd.testSymbol)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("⚠️ 代理模式调用失败: %v (耗时: %v)\n", err, duration)
		fmt.Println("💡 代理服务可能无法访问目标API，但仍会返回默认精度值")
	} else {
		fmt.Printf("✅ 代理模式调用成功: 精度=%d (耗时: %v)\n", precision, duration)
	}
}

// demoPrecisionManagerCall 演示精度管理器调用
func (psd *PrecisionSimulationDemo) demoPrecisionManagerCall() {
	fmt.Println("\n=== 演示3: 精度管理器调用 ===")
	fmt.Printf("🔄 演示PrecisionManager在后台的工作原理...\n")

	// 演示精度管理器的工作原理
	fmt.Println("💡 PrecisionManager在交易过程中自动工作:")
	fmt.Println("   • 通过GetSymbolPrecision获取精度信息")
	fmt.Println("   • 自动缓存精度信息以提高性能")
	fmt.Println("   • 在数量和价格格式化时自动应用精度规则")
	fmt.Println("   • 提供默认精度值作为后备方案")

	// 实际使用中，精度管理器会在调用GetSymbolPrecision等方法时发挥作用
	trader := trader.NewFuturesTrader("", "", "demo_user", "")
	startTime := time.Now()
	precision, err := trader.GetSymbolPrecision(psd.testSymbol)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("⚠️ 精度获取过程中使用了默认值: %v (耗时: %v)\n", err, duration)
	} else {
		fmt.Printf("✅ 精度获取成功: 精度=%d (耗时: %v)\n", precision, duration)
	}

	fmt.Println("💡 无论直接调用GetSymbolPrecision还是内部使用PrecisionManager，工作流程相同")
}

// demoRealAndTestnetPrecision 演示实盘和模拟盘精度获取
func (psd *PrecisionSimulationDemo) demoRealAndTestnetPrecision() {
	fmt.Println("\n=== 演示4: 实盘与模拟盘精度获取对比 ===")

	// 实盘精度获取
	fmt.Printf("🔄 获取实盘API精度 (%s)...\n", psd.testSymbol)
	realTrader := trader.NewFuturesTraderViaProxy("", "", "demo_user", "http://localhost:8081", "https://fapi.binance.com")
	realPrecision, err := realTrader.GetSymbolPrecision(psd.testSymbol)
	if err != nil {
		fmt.Printf("   ⚠️ 实盘API精度获取失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 实盘API精度: %d\n", realPrecision)
	}

	// 模拟盘精度获取
	fmt.Printf("🔄 获取模拟盘API精度 (%s)...\n", psd.testSymbol)
	testnetTrader := trader.NewFuturesTraderViaProxy("", "", "demo_user", "http://localhost:8081", "https://testnet.binancefuture.com")
	testnetPrecision, err := testnetTrader.GetSymbolPrecision(psd.testSymbol)
	if err != nil {
		fmt.Printf("   ⚠️ 模拟盘API精度获取失败: %v\n", err)
	} else {
		fmt.Printf("   ✅ 模拟盘API精度: %d\n", testnetPrecision)
	}

	fmt.Println("\n💡 总结:")
	fmt.Println("   • 透明代理架构允许无缝切换实盘和模拟盘")
	fmt.Println("   • 精度获取具有多重容错机制")
	fmt.Println("   • 当网络访问受限时，系统会使用默认精度值保证功能可用")
	fmt.Println("   • PrecisionManager提供统一的精度管理接口")
}

// RunPrecisionSimulationDemo 运行精度模拟演示的便捷函数
func RunPrecisionSimulationDemo() {
	demo := NewPrecisionSimulationDemo("BTCUSDT")
	demo.Run()
}

func main() {
	fmt.Println("🌟 精度获取模拟演示")
	fmt.Println("📋 模拟在前端后端运行状态下获取交易对精度的完整流程")
	RunPrecisionSimulationDemo()
}
