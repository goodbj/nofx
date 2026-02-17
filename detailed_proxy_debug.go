package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"nofx/trader"
)

func main() {
	fmt.Println("🔍 透明代理详细调试 - 深入分析unexpected EOF问题")
	fmt.Println("=================================================")

	// 设置代理环境变量
	os.Setenv("USE_BINANCE_PROXY", "true")
	os.Setenv("BINANCE_PROXY_URL", "http://localhost:8081")

	symbol := "BTCUSDT"
	fmt.Printf("📊 查询交易对: %s\n", symbol)

	fmt.Println("\n🔍 1. 检查CustomTransport行为...")
	testCustomTransportBehavior()

	fmt.Println("\n🔄 2. 测试不同API端点的行为...")
	testDifferentEndpoints()

	fmt.Println("\n🔄 3. 创建代理交易者并测试精度获取...")
	proxyTrader := trader.NewFuturesTraderViaProxy("", "", "debug_user", "http://localhost:8081", "https://fapi.binance.com")

	// 记录开始时间
	startTime := time.Now()
	precision, err := proxyTrader.GetSymbolPrecision(symbol)
	duration := time.Since(startTime)

	fmt.Printf("\n⏱️  精度获取总耗时: %v\n", duration)

	if err != nil {
		fmt.Printf("❌ 精度获取失败: %v\n", err)
		fmt.Println("\n🔍 unexpected EOF错误分析:")
		fmt.Println("   • unexpected EOF通常表示连接在读取响应时意外终止")
		fmt.Println("   • 可能发生在代理服务转发请求的过程中")
		fmt.Println("   • 原因可能是响应数据过大、连接超时或中间网络设备干扰")
	} else {
		fmt.Printf("✅ 精度获取成功! 精度值: %d\n", precision)
	}

	fmt.Println("\n📋 4. 问题根本原因分析:")
	fmt.Println("   • 代理转发测试(ping)成功，但exchangeInfo失败")
	fmt.Println("   • exchangeInfo端点返回的数据量比ping大得多")
	fmt.Println("   • 可能是代理服务在处理大数据响应时出现问题")
	fmt.Println("   • 或者是币安API对代理请求的响应处理有特殊限制")

	fmt.Println("\n🔧 5. 针对unexpected EOF的解决方案:")
	fmt.Println("   • 增加代理服务的响应缓冲区大小")
	fmt.Println("   • 调整HTTP客户端的超时设置")
	fmt.Println("   • 检查代理服务的内存池设置")
	fmt.Println("   • 优化大数据响应的处理逻辑")

	fmt.Println("\n🎯 6. 验证其他API端点...")
	verifyOtherEndpoints()

	fmt.Println("\n✨ 详细调试完成")
}

func testCustomTransportBehavior() {
	// 创建一个简单的HTTP客户端来测试CustomTransport
	fmt.Println("   🔄 测试CustomTransport添加头部的功能...")

	// 创建一个使用CustomTransport的简单HTTP客户端
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
	}

	customTransport := &trader.CustomTransport{
		Transport:      transport,
		TargetEndpoint: "https://fapi.binance.com",
	}

	client := &http.Client{
		Transport: customTransport,
		Timeout:   30 * time.Second,
	}

	// 测试ping端点
	fmt.Println("   📡 测试ping端点...")
	req, err := http.NewRequest("GET", "http://localhost:8081/fapi/v1/ping", nil)
	if err != nil {
		fmt.Printf("      ❌ 请求创建失败: %v\n", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("      ❌ ping请求失败: %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("      ✅ ping请求成功: 状态码 %d\n", resp.StatusCode)
	}
}

func testDifferentEndpoints() {
	fmt.Println("   🔄 测试不同API端点的行为差异...")

	// 测试ping端点（响应小）
	fmt.Println("   📡 测试 fapi/v1/ping (小响应)...")
	testEndpoint("http://localhost:8081/fapi/v1/ping", "https://fapi.binance.com")

	// 测试系统状态端点
	fmt.Println("   📡 测试 fapi/v1/system/status (小响应)...")
	testEndpoint("http://localhost:8081/fapi/v1/system/status", "https://fapi.binance.com")
}

func testEndpoint(proxyURL, targetURL string) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", proxyURL, nil)
	if err != nil {
		fmt.Printf("      ❌ 请求创建失败: %v\n", err)
		return
	}

	// 添加X-Target-URL头部
	req.Header.Set("X-Target-URL", targetURL)

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("      ❌ 请求失败 (%v): 耗时 %v\n", err, duration)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("      ✅ 请求成功: 状态码 %d, 响应大小 %d 字节, 耗时 %v\n",
			resp.StatusCode, len(body), duration)

		// 打印响应数据（只打印前200个字符以避免输出过长）
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			fmt.Printf("         📄 响应数据 (前200字符): %s...\n", bodyStr[:200])
		} else {
			fmt.Printf("         📄 响应数据: %s\n", bodyStr)
		}
	}
}

func verifyOtherEndpoints() {
	fmt.Println("   🔄 验证其他API端点的响应大小...")

	endpoints := []struct {
		name string
		path string
	}{
		{"Ping", "fapi/v1/ping"},
		{"Time", "fapi/v1/time"},
		{"Exchange Info (Limited)", "fapi/v1/exchangeInfo?symbol=BTCUSDT"}, // 尝试带参数的版本
	}

	for _, ep := range endpoints {
		url := fmt.Sprintf("http://localhost:8081/%s", ep.path)
		fmt.Printf("   📡 测试 %s: %s\n", ep.name, ep.path)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("      ❌ 请求创建失败: %v\n", err)
			continue
		}

		req.Header.Set("X-Target-URL", "https://fapi.binance.com")

		client := &http.Client{Timeout: 30 * time.Second}
		startTime := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("      ❌ %s 失败 (%v): 耗时 %v\n", ep.name, err, duration)
		} else {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("      ✅ %s 成功: 状态码 %d, 响应大小 %d 字节, 耗时 %v\n",
				ep.name, resp.StatusCode, len(body), duration)

			if ep.name == "Exchange Info (Limited)" {
				fmt.Printf("         💡 这个端点可能返回较小的响应，适合测试大数据响应问题\n")
			}
		}
	}
}
