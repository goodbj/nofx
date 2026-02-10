package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 网络连接测试")
	fmt.Println("================")

	// 测试基本网络连接
	url := "https://fapi.binance.com/fapi/v1/time"
	fmt.Printf("测试URL: %s\n", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ 连接成功\n")
	fmt.Printf("状态码: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Printf("网络连接正常\n")
	} else {
		fmt.Printf("连接异常，状态码: %d\n", resp.StatusCode)
	}
}
