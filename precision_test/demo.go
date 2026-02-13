package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	fmt.Println("🔬 精度管理器核心逻辑验证")
	fmt.Println(strings.Repeat("=", 45))

	// 演示核心精度处理逻辑
	demoPrecisionLogic()

	fmt.Println("\n✅ 验证完成")
}

func demoPrecisionLogic() {
	fmt.Println("\n🎯 核心精度处理演示:")

	// 测试用例
	testCases := []struct {
		quantity float64
		stepSize float64
		expected string
		desc     string
	}{
		{1.23456789, 0.001, "1.234", "正常对齐"},
		{0.0015, 0.001, "0.001", "向下取整"},
		{0.001000000001, 0.001, "0.001", "浮点数精度处理"},
		{1000.999, 1.0, "1000", "整数stepSize"},
		{0.1234500000001, 0.0001, "0.1234", "尾随零处理"},
	}

	for _, tc := range testCases {
		result := alignToStepSize(tc.quantity, tc.stepSize)
		precision := calculatePrecision(tc.stepSize)
		formatted := formatWithPrecision(result, precision)

		status := "✅"
		if formatted != tc.expected {
			status = "❌"
		}

		fmt.Printf("  %s %s:\n", status, tc.desc)
		fmt.Printf("     原始: %.12f\n", tc.quantity)
		fmt.Printf("     stepSize: %f\n", tc.stepSize)
		fmt.Printf("     对齐后: %.12f\n", result)
		fmt.Printf("     格式化: %s (期望: %s)\n", formatted, tc.expected)
		fmt.Println()
	}
}

// 核心精度处理函数
func alignToStepSize(quantity, stepSize float64) float64 {
	if stepSize <= 0 {
		return quantity
	}

	// 向下取整到最近的stepSize倍数
	aligned := math.Floor(quantity/stepSize) * stepSize

	// 处理浮点数精度问题
	aligned = math.Round(aligned*1e10) / 1e10

	return aligned
}

func formatWithPrecision(quantity float64, precision int) string {
	if precision < 0 {
		precision = 3
	}

	format := fmt.Sprintf("%%.%df", precision)
	formatted := fmt.Sprintf(format, quantity)

	// 移除尾随零
	formatted = strings.TrimRight(formatted, "0")
	if strings.HasSuffix(formatted, ".") {
		formatted = formatted[:len(formatted)-1]
	}

	return formatted
}

func calculatePrecision(stepSize float64) int {
	stepSizeStr := fmt.Sprintf("%f", stepSize)
	stepSizeStr = strings.TrimRight(stepSizeStr, "0")
	if strings.HasSuffix(stepSizeStr, ".") {
		stepSizeStr = stepSizeStr[:len(stepSizeStr)-1]
	}

	dotIndex := strings.Index(stepSizeStr, ".")
	if dotIndex == -1 {
		return 0
	}

	return len(stepSizeStr) - dotIndex - 1
}
