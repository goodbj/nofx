package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 网络连通性测试")
	fmt.Println("========================")

	// 测试基本网络连接
	testURLs := []string{
		"https://www.baidu.com",
		"https://www.google.com",
		"https://fapi.binance.com",
		"https://testnet.binancefuture.com",
	}

	for _, url := range testURLs {
		fmt.Printf("测试连接: %s\n", url)
		start := time.Now()
		
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		
		resp, err := client.Get(url)
		duration := time.Since(start)
		
		if err != nil {
			fmt.Printf("  ❌ 连接失败: %v (耗时: %v)\n", err, duration)
		} else {
			fmt.Printf("  ✅ 连接成功: 状态码 %d (耗时: %v)\n", resp.StatusCode, duration)
			resp.Body.Close()
		}
		fmt.Println()
	}

	// 测试DNS解析
	fmt.Println("=== DNS解析测试 ===")
	domains := []string{
		"fapi.binance.com",
		"testnet.binancefuture.com",
		"api.binance.com",
	}

	for _, domain := range domains {
		fmt.Printf("解析域名: %s\n", domain)
		ips, err := net.LookupIP(domain)
		if err != nil {
			fmt.Printf("  ❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 解析成功: %v\n", ips)
		}
		fmt.Println()
	}
}