package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"nofx/trader"
	
	"github.com/adshao/go-binance/v2/futures"
)

// SimulatePrecisionTest 模拟精度获取测试
type SimulatePrecisionTest struct {
	testSymbol string
}

// NewSimulatePrecisionTest 创建新的精度测试实例
func NewSimulatePrecisionTest(symbol string) *SimulatePrecisionTest {
	return &SimulatePrecisionTest{
		testSymbol: symbol,
	}
}

// Run 执行精度获取模拟测试
func (spt *SimulatePrecisionTest) Run() {
	fmt.Println("🔍 开始模拟精度获取测试...")
	fmt.Printf("📊 测试交易对: %s\n", spt.testSymbol)

	// 测试1: 直接通过Binance API获取精度
	fmt.Println("\n=== 1. 直接API精度获取 ===")
	spt.testDirectAPICall()

	// 测试2: 通过PrecisionManager获取精度
	fmt.Println("\n=== 2. PrecisionManager精度获取 ===")
	spt.testPrecisionManager()

	// 测试3: 通过代理模式获取精度
	fmt.Println("\n=== 3. 代理模式精度获取 ===")
	spt.testProxyMode()
}

// testDirectAPICall 测试直接API调用
func (spt *SimulatePrecisionTest) testDirectAPICall() {
	fmt.Println("🌐 正在通过直接API调用获取精度...")

	// 创建期货客户端
	client := futures.NewClient("", "") // 空的API密钥，只测试公共接口

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 调用exchangeInfo API
	fmt.Printf("🔄 正在获取 %s 的exchangeInfo...\n", spt.testSymbol)
	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		fmt.Printf("❌ 直接API调用失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 直接API调用成功，获取到 %d 个交易对信息\n", len(exchangeInfo.Symbols))

	// 查找指定交易对的精度信息
	found := false
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == spt.testSymbol {
			fmt.Printf("📋 %s 精度信息:\n", spt.testSymbol)
			fmt.Printf("   价格精度: %d\n", s.PricePrecision)
			fmt.Printf("   数量精度: %d\n", s.QuantityPrecision)

			// 查找step size和tick size
			for _, filter := range s.Filters {
				if filterType, ok := filter["filterType"].(string); ok {
					switch filterType {
					case "LOT_SIZE":
						if stepSize, ok := filter["stepSize"].(string); ok {
							fmt.Printf("   Step Size: %s\n", stepSize)
						}
						if minQty, ok := filter["minQty"].(string); ok {
							fmt.Printf("   Min Qty: %s\n", minQty)
						}
						if maxQty, ok := filter["maxQty"].(string); ok {
							fmt.Printf("   Max Qty: %s\n", maxQty)
						}
					case "PRICE_FILTER":
						if tickSize, ok := filter["tickSize"].(string); ok {
							fmt.Printf("   Tick Size: %s\n", tickSize)
						}
						if minPrice, ok := filter["minPrice"].(string); ok {
							fmt.Printf("   Min Price: %s\n", minPrice)
						}
						if maxPrice, ok := filter["maxPrice"].(string); ok {
							fmt.Printf("   Max Price: %s\n", maxPrice)
						}
					}
				}
			}
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("⚠️ 未找到 %s 的精度信息\n", spt.testSymbol)
	}
}

// testPrecisionManager 测试PrecisionManager
func (spt *SimulatePrecisionTest) testPrecisionManager() {
	fmt.Println("🔧 正在通过PrecisionManager获取精度...")

	// 创建一个期货客户端
	client := futures.NewClient("", "")
	
	// 创建精度管理器
	precisionMgr := trader.NewPrecisionManager(client)

	// 获取精度信息
	fmt.Printf("🔄 正在获取 %s 精度信息...\n", spt.testSymbol)
	precisionInfo, err := precisionMgr.GetPrecisionInfo(spt.testSymbol)
	if err != nil {
		fmt.Printf("❌ PrecisionManager获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ PrecisionManager获取成功:\n")
	fmt.Printf("   交易对: %s\n", precisionInfo.Symbol)
	fmt.Printf("   Step Size: %f\n", precisionInfo.StepSize)
	fmt.Printf("   Tick Size: %f\n", precisionInfo.TickSize)
	fmt.Printf("   最小数量: %f\n", precisionInfo.MinQty)
	fmt.Printf("   最大数量: %f\n", precisionInfo.MaxQty)
	fmt.Printf("   精度: %d\n", precisionInfo.Precision)
	fmt.Printf("   更新时间: %v\n", precisionInfo.LastUpdate)
}

// testProxyMode 测试代理模式
func (spt *SimulatePrecisionTest) testProxyMode() {
	fmt.Println("🔗 正在通过代理模式获取精度...")

	// 检查是否启用了代理
	useProxy := os.Getenv("USE_BINANCE_PROXY")
	if useProxy != "true" && useProxy != "1" {
		fmt.Println("⚠️ 代理模式未启用 (USE_BINANCE_PROXY != true)")
		
		// 即使未启用代理，我们也可以测试代理相关的功能
		fmt.Println("💡 演示代理模式的工作原理:")
		fmt.Println("   1. 通过NewFuturesTraderViaProxy创建交易者")
		fmt.Println("   2. CustomTransport拦截HTTP请求")
		fmt.Println("   3. 添加X-Target-URL头部指向真实API")
		fmt.Println("   4. 请求发送到代理服务 (localhost:8081)")
		fmt.Println("   5. 代理服务转发请求到真实交易所API")
		fmt.Println("   6. 响应返回给客户端")
		return
	}

	// 如果启用了代理，创建代理交易者
	proxyURL := os.Getenv("BINANCE_PROXY_URL")
	if proxyURL == "" {
		proxyURL = "http://localhost:8081"
	}

	fmt.Printf("🌐 代理URL: %s\n", proxyURL)
	
	// 创建通过代理的期货交易者
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "test_user", proxyURL, "https://fapi.binance.com")
	
	// 通过代理获取精度
	precisionInfo, err := proxyTrader.GetSymbolPrecision(spt.testSymbol)
	if err != nil {
		fmt.Printf("❌ 代理模式获取精度失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 代理模式获取精度成功: %d\n", precisionInfo)
	
	// 由于无法直接访问precisionManager，跳过此测试
	fmt.Println("⚠️ 无法直接访问代理交易者的PrecisionManager (字段未导出)")
	fmt.Println("💡 在实际应用中，会通过交易者的方法间接使用PrecisionManager")
}

// RunSimulatePrecisionTest 运行精度模拟测试的便捷函数
func RunSimulatePrecisionTest() {
	test := NewSimulatePrecisionTest("BTCUSDT")
	test.Run()
}

func main() {
	fmt.Println("🚀 开始模拟精度获取测试")
	RunSimulatePrecisionTest()
	fmt.Println("\n✅ 精度获取模拟测试完成")
}