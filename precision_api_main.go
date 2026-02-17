package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"trader"
)

// PrecisionAPITest 精度API调取测试
type PrecisionAPITest struct {
	realAPI    string
	testnetAPI string
}

// NewPrecisionAPITest 创建新的精度API测试
func NewPrecisionAPITest() *PrecisionAPITest {
	return &PrecisionAPITest{
		realAPI:    "https://fapi.binance.com/fapi/v1/exchangeInfo",
		testnetAPI: "https://testnet.binancefuture.com/fapi/v1/exchangeInfo",
	}
}

// Run 运行精度API测试
func (pat *PrecisionAPITest) Run() {
	fmt.Println("🔍 精度信息API调取测试")
	fmt.Println("🎯 目标：通过API从服务器获取正确的精度信息")
	fmt.Println("📋 测试地址：")
	fmt.Printf("   实盘API: %s\n", pat.realAPI)
	fmt.Printf("   模拟盘API: %s\n", pat.testnetAPI)
	fmt.Println()

	// 测试1：透明代理访问实盘API
	pat.testProxyAccess("实盘API", pat.realAPI, "https://fapi.binance.com")

	// 测试2：透明代理访问模拟盘API
	pat.testProxyAccess("模拟盘API", pat.testnetAPI, "https://testnet.binancefuture.com")

	// 测试3：直连访问实盘API
	pat.testDirectAccess("实盘API直连", pat.realAPI)

	// 测试4：直连访问模拟盘API
	pat.testDirectAccess("模拟盘API直连", pat.testnetAPI)
}

// testProxyAccess 测试透明代理访问
func (pat *PrecisionAPITest) testProxyAccess(testName, apiURL, targetEndpoint string) {
	fmt.Printf("=== %s 透明代理测试 ===\n", testName)
	fmt.Printf("🌐 API地址: %s\n", apiURL)
	fmt.Printf("🎯 目标端点: %s\n", targetEndpoint)

	// 创建代理交易者
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "precision_test_user", "http://localhost:8081", targetEndpoint)
	if proxyTrader == nil {
		fmt.Printf("❌ 无法创建代理交易者\n")
		return
	}

	fmt.Println("🔄 通过代理获取exchangeInfo...")

	// 记录开始时间
	startTime := time.Now()

	// 直接调用exchangeInfo API（不使用默认值回退）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exchangeInfo, err := proxyTrader.GetClient().NewExchangeInfoService().Do(ctx)
	duration := time.Since(startTime)

	fmt.Printf("⏱️  请求耗时: %v\n", duration)

	if err != nil {
		fmt.Printf("❌ 代理访问失败: %v\n", err)
		fmt.Println("📋 错误详情:")
		fmt.Printf("   错误类型: %T\n", err)
		fmt.Printf("   错误信息: %v\n", err)
		if ctx.Err() != nil {
			fmt.Printf("   上下文错误: %v\n", ctx.Err())
		}
	} else {
		fmt.Printf("✅ 代理访问成功\n")
		fmt.Printf("📊 返回数据统计:\n")
		fmt.Printf("   交易对数量: %d\n", len(exchangeInfo.Symbols))
		fmt.Printf("   服务器时间: %d\n", exchangeInfo.ServerTime)
		fmt.Printf("   时区: %s\n", exchangeInfo.Timezone)

		// 显示前几个交易对的精度信息示例
		if len(exchangeInfo.Symbols) > 0 {
			fmt.Println("📋 精度信息示例 (前5个交易对):")
			for i, symbol := range exchangeInfo.Symbols {
				if i >= 5 {
					break
				}
				fmt.Printf("   %s: 价格精度=%d, 数量精度=%d\n",
					symbol.Symbol, symbol.QuotePrecision, symbol.BaseAssetPrecision)

				// 显示LOT_SIZE过滤器信息
				for _, filter := range symbol.Filters {
					if filterType, ok := filter["filterType"].(string); ok && filterType == "LOT_SIZE" {
						if stepSize, ok := filter["stepSize"].(string); ok {
							fmt.Printf("       Step Size: %s\n", stepSize)
						}
						if minQty, ok := filter["minQty"].(string); ok {
							fmt.Printf("       Min Qty: %s\n", minQty)
						}
					}
				}
			}
		}
	}

	fmt.Println()
}

// testDirectAccess 测试直连访问
func (pat *PrecisionAPITest) testDirectAccess(testName, apiURL string) {
	fmt.Printf("=== %s 直连测试 ===\n", testName)
	fmt.Printf("🌐 API地址: %s\n", apiURL)

	// 创建直连客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	fmt.Println("🔄 直连获取exchangeInfo...")

	// 记录开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	fmt.Printf("⏱️  请求耗时: %v\n", duration)

	if err != nil {
		fmt.Printf("❌ 直连访问失败: %v\n", err)
		fmt.Println("📋 错误详情:")
		fmt.Printf("   错误类型: %T\n", err)
		fmt.Printf("   错误信息: %v\n", err)
		if resp != nil {
			fmt.Printf("   响应状态: %s\n", resp.Status)
		}
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ 直连访问成功\n")
	fmt.Printf("📊 响应信息:\n")
	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// 读取响应体大小
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("❌ 读取响应体失败: %v\n", err)
		return
	}

	fmt.Printf("   响应体大小: %d bytes (前%d字节)\n", n, n)
	fmt.Printf("   响应内容预览: %s\n", string(buf[:n]))

	fmt.Println()
}

// RunPrecisionAPITest 运行精度API测试的便捷函数
func RunPrecisionAPITest() {
	test := NewPrecisionAPITest()
	test.Run()
}

func main() {
	fmt.Println("🚀 精度信息API调取测试程序")
	fmt.Println("📋 本程序将测试通过API从服务器获取精度信息")
	fmt.Println("⚠️  不使用默认值回退机制")
	fmt.Println()

	RunPrecisionAPITest()
}
