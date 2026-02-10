package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 测试API连接状态")
	fmt.Println("==================")

	// 测试后端API
	testAPI("http://localhost:8888/api/health", "后端健康检查")
	testAPI("http://localhost:8888/api/supported-exchanges", "交易所支持列表")
	testAPI("http://localhost:3300", "前端服务")
}

func testAPI(url, name string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ %s: 连接失败 - %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ %s: 连接正常 (状态码: %d)\n", name, resp.StatusCode)
	} else {
		fmt.Printf("⚠️  %s: 状态码异常 - %d\n", name, resp.StatusCode)
	}
}
