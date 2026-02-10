package main

import (
	"fmt"
	"strings"
)

// 模拟新规则下的逻辑：完全忽略testnet开关，只看customAPIURL
func getRealExchangeEndpoint(customAPIURL string, testnet bool) string {
	realExchangeEndpoint := ""

	// 根据新规则：完全忽略Testnet开关，只看CustomAPIURL
	// 但本地代理地址仍需特殊处理，因为它们通常不应该作为最终目标
	if customAPIURL != "" && strings.TrimSpace(customAPIURL) != "" {
		// 检查是否为本地代理地址
		isLocalhost := strings.Contains(customAPIURL, "://localhost:") ||
			strings.Contains(customAPIURL, "://127.0.0.1:")

		if isLocalhost {
			// 如果是本地代理地址，使用默认主网API
			// 这是因为本地代理通常是中间节点，不是最终目标
			realExchangeEndpoint = "https://fapi.binance.com" // 默认主网API URL
		} else {
			// 如果是有效的非本地地址，则直接使用它
			realExchangeEndpoint = customAPIURL
		}
	} else {
		// 如果CustomAPIURL为空，则使用交易所默认API地址
		realExchangeEndpoint = "https://fapi.binance.com" // 默认主网API URL
	}

	return realExchangeEndpoint
}

func main() {
	fmt.Println("🧪 完全忽略Testnet开关的新逻辑验证")
	fmt.Println("==================================")

	// 测试用例 - 根据新规则：完全忽略testnet开关，只看customAPIURL
	testCases := []struct {
		name         string
		customAPIURL string
		testnet      bool // 这个字段现在被忽略
		expected     string
	}{
		{
			name:         "自定义URL为空 -> 使用默认主网API（忽略testnet=true）",
			customAPIURL: "",
			testnet:      true, // 即使testnet为true也应被忽略
			expected:     "https://fapi.binance.com",
		},
		{
			name:         "自定义URL为空格 -> 使用默认主网API（忽略testnet=false）",
			customAPIURL: "   ",
			testnet:      false, // 即使testnet为false也应被忽略
			expected:     "https://fapi.binance.com",
		},
		{
			name:         "自定义URL为有效地址 -> 使用自定义URL（忽略testnet=true）",
			customAPIURL: "https://custom.binance.com",
			testnet:      true, // testnet被忽略
			expected:     "https://custom.binance.com",
		},
		{
			name:         "自定义URL为测试网地址 -> 使用自定义URL（忽略testnet=false）",
			customAPIURL: "https://testnet.binancefuture.com",
			testnet:      false, // testnet被忽略
			expected:     "https://testnet.binancefuture.com",
		},
		{
			name:         "自定义URL为本地代理且为有效地址 -> 使用默认主网API（忽略testnet）",
			customAPIURL: "http://localhost:8081",
			testnet:      true, // testnet被忽略
			expected:     "https://fapi.binance.com",
		},
		{
			name:         "自定义URL为127.0.0.1代理 -> 使用默认主网API（忽略testnet）",
			customAPIURL: "http://127.0.0.1:8081",
			testnet:      false, // testnet被忽略
			expected:     "https://fapi.binance.com",
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
