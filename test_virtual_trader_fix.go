package main

import (
	"fmt"
	"strings"
)

// 模拟修复后的逻辑
func getRealExchangeEndpoint(customAPIURL string, testnet bool) string {
	realExchangeEndpoint := ""
	
	// 修复：确保空字符串也被正确处理
	if customAPIURL != "" && strings.TrimSpace(customAPIURL) != "" && 
	   !strings.Contains(customAPIURL, "://localhost:") && 
	   !strings.Contains(customAPIURL, "://127.0.0.1:") {
		// 如果CustomAPIURL不为空且不包含本地地址，则使用它作为真实的交易所URL
		realExchangeEndpoint = customAPIURL
	} else if testnet {
		realExchangeEndpoint = "https://testnet.binancefuture.com"
	} else {
		realExchangeEndpoint = "https://fapi.binance.com" // 默认主网API URL
	}
	
	return realExchangeEndpoint
}

func main() {
	fmt.Println("🧪 虚拟盘测试网逻辑修复验证")
	fmt.Println("==============================")
	
	// 测试用例
	testCases := []struct {
		name         string
		customAPIURL string
		testnet      bool
		expected     string
	}{
		{
			name:         "虚拟盘 + 使用测试网 (空字符串)",
			customAPIURL: "",
			testnet:      true,
			expected:     "https://testnet.binancefuture.com",
		},
		{
			name:         "虚拟盘 + 使用测试网 (只有空格)",
			customAPIURL: "   ",
			testnet:      true,
			expected:     "https://testnet.binancefuture.com",
		},
		{
			name:         "虚拟盘 + 不使用测试网 (空字符串)",
			customAPIURL: "",
			testnet:      false,
			expected:     "https://fapi.binance.com",
		},
		{
			name:         "实盘 + 自定义URL",
			customAPIURL: "https://custom.binance.com",
			testnet:      false,
			expected:     "https://custom.binance.com",
		},
		{
			name:         "虚拟盘 + 本地代理URL",
			customAPIURL: "http://localhost:8081",
			testnet:      true,
			expected:     "https://testnet.binancefuture.com",
		},
		{
			name:         "虚拟盘 + 127.0.0.1代理URL",
			customAPIURL: "http://127.0.0.1:8081",
			testnet:      true,
			expected:     "https://testnet.binancefuture.com",
		},
	}
	
	fmt.Println("📋 测试结果:")
	fmt.Println()
	
	allPassed := true
	for i, tc := range testCases {
		result := getRealExchangeEndpoint(tc.customAPIURL, tc.testnet)
		passed := result == tc.expected
		
		if !passed {
			allPassed = false
		}
		
		status := "✅"
		if !passed {
			status = "❌"
		}
		
		fmt.Printf("%d. %s %s\n", i+1, status, tc.name)
		fmt.Printf("   输入: CustomAPIURL='%s', Testnet=%t\n", tc.customAPIURL, tc.testnet)
		fmt.Printf("   期望: %s\n", tc.expected)
		fmt.Printf("   实际: %s\n", result)
		if !passed {
			fmt.Printf("   ❌ 不匹配!\n")
		}
		fmt.Println()
	}
	
	fmt.Println("==============================")
	if allPassed {
		fmt.Println("🎉 所有测试通过！虚拟盘测试网逻辑修复成功")
	} else {
		fmt.Println("⚠️  部分测试失败，请检查逻辑")
	}
}