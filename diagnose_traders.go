package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔧 交易员余额获取问题诊断工具")
	fmt.Println("==============================")

	// 测试两个交易员
	traderIDs := []string{
		"83faf7b3_deepseek_1770656396",    // 实盘交易员
		"75103af7_guardian-ai_1770569967", // 虚拟盘交易员
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, traderID := range traderIDs {
		fmt.Printf("\n🔍 测试交易员: %s\n", traderID)
		fmt.Println("------------------------")

		// 测试账户信息API
		url := fmt.Sprintf("http://localhost:8888/api/account?trader_id=%s", traderID)
		resp, err := client.Get(url)

		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// 读取响应内容
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("状态码: %d\n", resp.StatusCode)
		fmt.Printf("响应内容: %s\n", string(body))

		if resp.StatusCode == 200 {
			fmt.Printf("✅ 交易员 %s 余额获取成功\n", traderID)
		} else if resp.StatusCode == 404 {
			fmt.Printf("❌ 交易员 %s 未找到 (404) - 可能未正确加载到内存\n", traderID)
		} else if resp.StatusCode == 500 {
			fmt.Printf("❌ 交易员 %s 服务器内部错误 (500) - 网络或API连接问题\n", traderID)
		} else {
			fmt.Printf("⚠️  交易员 %s 返回状态: %d\n", traderID, resp.StatusCode)
		}
	}

	fmt.Println("\n📋 问题分析和修复建议:")
	fmt.Println("=====================")
	fmt.Println("1. 实盘交易员 (83faf7b3_deepseek_1770656396):")
	fmt.Println("   - 404错误表明交易员未加载到内存")
	fmt.Println("   - 原因：API密钥权限不足导致初始化失败")
	fmt.Println("   - 解决方案：检查并更新API密钥权限")
	fmt.Println()
	fmt.Println("2. 虚拟盘交易员 (75103af7_guardian-ai_1770569967):")
	fmt.Println("   - 500错误表明服务器内部错误")
	fmt.Println("   - 原因：网络连接问题或代理配置错误")
	fmt.Println("   - 解决方案：检查网络连接和代理服务")
	fmt.Println()
	fmt.Println("🔧 推荐修复步骤:")
	fmt.Println("1. 验证API密钥权限（特别是实盘交易员）")
	fmt.Println("2. 检查网络连接和代理服务状态")
	fmt.Println("3. 重启后端服务重新加载交易员配置")
	fmt.Println("4. 测试API连接连通性")
}
