package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// ConnectionTest 连接测试结构体
type ConnectionTest struct {
	testSymbol string
}

// NewConnectionTest 创建新的连接测试实例
func NewConnectionTest(symbol string) *ConnectionTest {
	return &ConnectionTest{
		testSymbol: symbol,
	}
}

// Run 执行完整的连接测试
func (ct *ConnectionTest) Run() {
	fmt.Println("🔍 开始连接测试 - 实盘和虚拟盘精度获取验证")

	// 1. 测试实盘连接
	fmt.Println("\n=== 实盘连接测试 ===")
	ct.testRealMarket()

	// 2. 测试虚拟盘连接
	fmt.Println("\n=== 虚拟盘连接测试 ===")
	ct.testTestnetMarket()

	// 3. 测试代理连接
	fmt.Println("\n=== 代理连接测试 ===")
	ct.testProxyConnection()

	fmt.Println("\n✅ 连接测试完成")
}

func (ct *ConnectionTest) testRealMarket() {
	fmt.Println("🌐 测试实盘市场连接...")

	// 使用空的API密钥进行测试（只测试连接）
	client := futures.NewClient("", "")

	// 设置较短的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("🔄 正在获取 %s 的exchangeInfo...\n", ct.testSymbol)
	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		fmt.Printf("❌ 实盘连接失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 实盘连接成功，获取到 %d 个交易对信息\n", len(exchangeInfo.Symbols))

	// 查找特定交易对的精度信息
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == ct.testSymbol {
			fmt.Printf("📋 %s 精度信息:\n", ct.testSymbol)
			fmt.Printf("   价格精度: %d\n", s.PricePrecision)
			fmt.Printf("   数量精度: %d\n", s.QuantityPrecision)

			// 查找step size
			for _, filter := range s.Filters {
				if filterType, ok := filter["filterType"].(string); ok {
					switch filterType {
					case "LOT_SIZE":
						if stepSize, ok := filter["stepSize"].(string); ok {
							fmt.Printf("   Step Size: %s\n", stepSize)
						}
					case "PRICE_FILTER":
						if tickSize, ok := filter["tickSize"].(string); ok {
							fmt.Printf("   Tick Size: %s\n", tickSize)
						}
					}
				}
			}
			return
		}
	}

	fmt.Printf("⚠️ 未找到 %s 的精度信息\n", ct.testSymbol)
}

func (ct *ConnectionTest) testTestnetMarket() {
	fmt.Println("🌐 测试虚拟盘市场连接...")

	// 虚拟盘客户端
	client := futures.NewClient("", "")
	client.BaseURL = "https://testnet.binancefuture.com"

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("🔄 正在获取 %s 的exchangeInfo...\n", ct.testSymbol)
	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		fmt.Printf("❌ 虚拟盘连接失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 虚拟盘连接成功，获取到 %d 个交易对信息\n", len(exchangeInfo.Symbols))

	// 查找特定交易对的精度信息
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == ct.testSymbol {
			fmt.Printf("📋 %s 精度信息:\n", ct.testSymbol)
			fmt.Printf("   价格精度: %d\n", s.PricePrecision)
			fmt.Printf("   数量精度: %d\n", s.QuantityPrecision)

			// 查找step size
			for _, filter := range s.Filters {
				if filterType, ok := filter["filterType"].(string); ok {
					switch filterType {
					case "LOT_SIZE":
						if stepSize, ok := filter["stepSize"].(string); ok {
							fmt.Printf("   Step Size: %s\n", stepSize)
						}
					case "PRICE_FILTER":
						if tickSize, ok := filter["tickSize"].(string); ok {
							fmt.Printf("   Tick Size: %s\n", tickSize)
						}
					}
				}
			}
			return
		}
	}

	fmt.Printf("⚠️ 未找到 %s 的精度信息\n", ct.testSymbol)
}

func (ct *ConnectionTest) testProxyConnection() {
	fmt.Println("🌐 测试代理连接...")

	// 检查环境变量中的代理设置
	proxyURL := os.Getenv("HTTP_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}

	if proxyURL == "" {
		fmt.Println("⚠️ 未检测到代理设置，跳过代理测试")
		return
	}

	fmt.Printf("🔍 检测到代理设置: %s\n", proxyURL)

	// 解析代理URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		fmt.Printf("❌ 代理URL解析失败: %v\n", err)
		return
	}

	// 创建带代理的HTTP客户端
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}

	// 测试代理连接到Binance
	testURLs := []string{
		"https://fapi.binance.com/fapi/v1/exchangeInfo",
		"https://testnet.binancefuture.com/fapi/v1/exchangeInfo",
	}

	for _, testURL := range testURLs {
		fmt.Printf("🔄 测试代理连接到: %s\n", testURL)

		req, err := http.NewRequest("GET", testURL, nil)
		if err != nil {
			fmt.Printf("❌ 请求创建失败: %v\n", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 代理连接失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("✅ 代理连接成功，状态码: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			fmt.Printf("✅ %s 通过代理可正常访问\n", testURL)
		}
	}
}

// RunConnectionTest 运行连接测试的便捷函数
func RunConnectionTest() {
	test := NewConnectionTest("BTCUSDT")
	test.Run()
}
