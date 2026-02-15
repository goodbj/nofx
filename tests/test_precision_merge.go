package main

import (
	"fmt"
	"nofx/trader"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	fmt.Println("🔬 精度系统合并验证测试")
	fmt.Println("=========================")

	// 创建测试用的客户端
	client := futures.NewClient("", "") // 空字符串表示使用默认配置

	// 创建精度管理器
	precisionManager := trader.NewPrecisionManager(client)

	// 测试交易对
	testSymbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}

	fmt.Println("1. 测试精度信息获取...")
	for _, symbol := range testSymbols {
		info, err := precisionManager.GetPrecisionInfo(symbol)
		if err != nil {
			fmt.Printf("   ❌ %s 获取精度信息失败: %v\n", symbol, err)
		} else {
			fmt.Printf("   ✅ %s - StepSize: %f, TickSize: %f, Precision: %d\n",
				symbol, info.StepSize, info.TickSize, info.Precision)
		}
	}

	fmt.Println("\n2. 测试数量格式化...")
	testQuantities := []struct {
		symbol   string
		quantity float64
	}{
		{"BTCUSDT", 1.23456789},
		{"ETHUSDT", 10.12345678},
		{"SOLUSDT", 100.987654321},
	}

	for _, tq := range testQuantities {
		formatted, err := precisionManager.FormatQuantityWithValidation(tq.symbol, tq.quantity)
		if err != nil {
			fmt.Printf("   ❌ %s 格式化失败: %v\n", tq.symbol, err)
		} else {
			fmt.Printf("   ✅ %s: %.8f -> %s\n", tq.symbol, tq.quantity, formatted)
		}
	}

	fmt.Println("\n3. 测试价格格式化...")
	testPrices := []struct {
		symbol string
		price  float64
	}{
		{"BTCUSDT", 45678.12345},
		{"ETHUSDT", 2345.6789},
		{"SOLUSDT", 123.45678},
	}

	for _, tp := range testPrices {
		formatted, err := precisionManager.FormatPriceWithValidation(tp.symbol, tp.price)
		if err != nil {
			fmt.Printf("   ❌ %s 价格格式化失败: %v\n", tp.symbol, err)
		} else {
			fmt.Printf("   ✅ %s: %.8f -> %s\n", tp.symbol, tp.price, formatted)
		}
	}

	fmt.Println("\n✅ 精度系统合并验证完成")
	fmt.Println("   系统现在使用统一的PrecisionManager进行所有精度处理")
}
