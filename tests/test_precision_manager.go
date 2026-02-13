package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"nofx/trader"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	fmt.Println("🔬 精度管理器测试")
	fmt.Println(strings.Repeat("=", 50))

	// 创建测试用的客户端（实际使用时替换为真实API密钥）
	client := futures.NewClient("test", "test")

	// 创建精度管理器
	precisionManager := trader.NewPrecisionManager(client)

	// 测试目标交易对
	testSymbols := []string{"BTCUSDT", "ETHUSDT", "AKEUSDT"}

	fmt.Println("📋 测试场景:")
	fmt.Println("1. 正常数量格式化")
	fmt.Println("2. 边界情况测试")
	fmt.Println("3. 缓存机制验证")
	fmt.Println("4. 并发安全性测试")
	fmt.Println()

	// 测试1: 正常数量格式化
	fmt.Println("🧪 测试1: 正常数量格式化")
	testNormalFormatting(precisionManager, testSymbols)

	// 测试2: 边界情况
	fmt.Println("\n🧪 测试2: 边界情况测试")
	testEdgeCases(precisionManager)

	// 测试3: 缓存机制
	fmt.Println("\n🧪 测试3: 缓存机制验证")
	testCacheMechanism(precisionManager, testSymbols[0])

	// 测试4: 并发安全性
	fmt.Println("\n🧪 测试4: 并发安全性测试")
	testConcurrentSafety(precisionManager, testSymbols[0])

	fmt.Println("\n✅ 所有测试完成")
}

func testNormalFormatting(pm *trader.PrecisionManager, symbols []string) {
	testCases := []float64{0.123456789, 1.23456789, 12.3456789, 123.456789}

	for _, symbol := range symbols {
		fmt.Printf("\n📊 %s 测试:\n", symbol)
		for _, qty := range testCases {
			formatted, err := pm.FormatQuantityWithValidation(symbol, qty)
			if err != nil {
				fmt.Printf("  ❌ %.8f -> 错误: %v\n", qty, err)
			} else {
				fmt.Printf("  ✅ %.8f -> %s\n", qty, formatted)
			}
		}
	}
}

func testEdgeCases(pm *trader.PrecisionManager) {
	// 测试各种边界情况
	edgeCases := []struct {
		symbol   string
		quantity float64
		desc     string
	}{
		{"AKEUSDT", 0.0009, "小于最小数量"},
		{"AKEUSDT", 0.001, "等于最小数量"},
		{"AKEUSDT", 0.001000000001, "微小超出"},
		{"AKEUSDT", 1000.001, "接近最大数量"},
		{"BTCUSDT", 0.000999999999, "浮点数精度问题"},
		{"BTCUSDT", 0.1234500000001, "尾随零处理"},
	}

	for _, tc := range edgeCases {
		formatted, err := pm.FormatQuantityWithValidation(tc.symbol, tc.quantity)
		status := "✅"
		if err != nil {
			status = "❌"
		}
		fmt.Printf("  %s %s (%s): %.12f -> %s\n",
			status, tc.symbol, tc.desc, tc.quantity, formatted)
		if err != nil {
			fmt.Printf("     错误: %v\n", err)
		}
	}
}

func testCacheMechanism(pm *trader.PrecisionManager, symbol string) {
	fmt.Printf("测试 %s 的缓存机制:\n", symbol)

	// 第一次调用（应该从API获取）
	start := time.Now()
	_, err1 := pm.FormatQuantityWithValidation(symbol, 1.234567)
	duration1 := time.Since(start)

	// 第二次调用（应该使用缓存）
	start = time.Now()
	_, err2 := pm.FormatQuantityWithValidation(symbol, 2.345678)
	duration2 := time.Since(start)

	fmt.Printf("  第一次调用: %v (耗时: %v)\n", err1 == nil, duration1)
	fmt.Printf("  第二次调用: %v (耗时: %v)\n", err2 == nil, duration2)
	fmt.Printf("  性能提升: %.2fx\n", float64(duration1)/float64(duration2))

	// 显示缓存统计
	stats := pm.GetCacheStats()
	fmt.Printf("  缓存统计: %+v\n", stats)
}

func testConcurrentSafety(pm *trader.PrecisionManager, symbol string) {
	fmt.Printf("测试 %s 的并发安全性:\n", symbol)

	// 启动多个goroutine同时处理
	const numGoroutines = 10
	const numOperations = 100

	results := make(chan string, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numOperations; j++ {
				// 生成随机数量
				quantity := 0.001 + rand.Float64()*1000

				formatted, err := pm.FormatQuantityWithValidation(symbol, quantity)
				if err != nil {
					results <- fmt.Sprintf("G%d-O%d: 错误 %v", goroutineID, j, err)
				} else {
					results <- fmt.Sprintf("G%d-O%d: %.6f -> %s", goroutineID, j, quantity, formatted)
				}
			}
		}(i)
	}

	// 收集结果
	successCount := 0
	errorCount := 0

	for i := 0; i < numGoroutines*numOperations; i++ {
		result := <-results
		if strings.Contains(result, "错误") {
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("  成功处理: %d\n", successCount)
	fmt.Printf("  错误数量: %d\n", errorCount)
	fmt.Printf("  成功率: %.2f%%\n", float64(successCount)/float64(successCount+errorCount)*100)
}

// 模拟测试用的exchange info服务
func init() {
	// 这里可以设置mock数据来测试实际功能
	// 在实际使用中，会调用真实的币安API
}
