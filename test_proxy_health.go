package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 测试代理服务健康状态...")

	// 测试代理服务健康检查
	url := "http://localhost:8081/health"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 无法连接到代理服务: %v\n", err)
		fmt.Println("💡 代理服务可能未启动或端口被占用")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	if resp.StatusCode == 200 {
		fmt.Printf("✅ 代理服务健康状态正常! 状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
	} else {
		fmt.Printf("⚠️ 代理服务响应异常! 状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
	}

	fmt.Println("\n🔍 测试代理转发功能...")
	// 简单测试代理转发功能
	proxyURL := "http://localhost:8081/fapi/v1/ping"
	resp2, err := client.Get(proxyURL)
	if err != nil {
		fmt.Printf("🔄 代理转发测试失败 (这可能是正常的，如果没有目标服务): %v\n", err)
	} else {
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		fmt.Printf("✅ 代理转发测试完成! 状态码: %d, 响应长度: %d 字节\n", resp2.StatusCode, len(body2))
	}
}
