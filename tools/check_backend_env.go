package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("🔍 检查运行中后端服务的环境变量")
	fmt.Println("================================")

	// 尝试调用后端API获取环境变量信息
	resp, err := http.Get("http://localhost:8888/api/debug/env")
	if err != nil {
		fmt.Printf("❌ 无法连接到后端服务: %v\n", err)
		fmt.Println("请确保后端服务正在运行在 http://localhost:8888")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ 成功连接到后端服务")
		fmt.Printf("响应状态: %s\n", resp.Status)
		// 这里可以读取响应体来获取环境变量信息
	} else {
		fmt.Printf("⚠️  后端服务返回状态: %s\n", resp.Status)
	}

	// 也检查本地环境变量作为对比
	fmt.Println("\n本地环境变量对比:")
	localBrowserDataDir := os.Getenv("BROWSER_DATA_DIR")
	if localBrowserDataDir != "" {
		fmt.Printf("本地 BROWSER_DATA_DIR: %s\n", localBrowserDataDir)
	} else {
		fmt.Println("本地 BROWSER_DATA_DIR: (未设置)")
	}
}
