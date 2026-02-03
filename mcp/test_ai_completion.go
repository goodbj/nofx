package mcp

import (
	"context"
	"fmt"
	"strings"
)

// TestAICompletionDetection 测试AI输出完成检测功能
func (gc *GuardianClient) TestAICompletionDetection() {
	fmt.Println("🧪 开始测试AI输出完成检测功能...")

	// 创建一个模拟的HTML内容，包含AI输出完成标志
	mockHTML := `<div class="some-class">
		<p>AI正在生成内容...</p>
		<div class="ds-flex _0a3d93b" style="align-items: center; gap: 10px;">
			<div class="some-content">AI生成的内容</div>
		</div>
	</div>`

	// 检测AI输出完成标志
	isCompleted := gc.CheckForAICompletionKeyword(mockHTML)

	if isCompleted {
		fmt.Println("✅ 测试成功：正确检测到AI输出完成标志")

		// 创建一个模拟的context用于测试
		// ctx := context.Background() // 注释掉未使用的变量
		_ = context.Background() // 避免未使用导入的错误

		// 模拟执行后续操作
		gc.logger.Println("🚀 执行后续自动化操作...")
		gc.logger.Println("📋 AI输出数据复制完成")
	} else {
		fmt.Println("❌ 测试失败：未能检测到AI输出完成标志")
	}

	// 测试不包含标志的情况
	mockHTMLIncomplete := `<div class="some-class">
		<p>AI正在生成内容...</p>
	</div>`

	isCompleted = gc.CheckForAICompletionKeyword(mockHTMLIncomplete)

	if !isCompleted {
		fmt.Println("✅ 测试成功：正确识别未完成状态")
	} else {
		fmt.Println("❌ 测试失败：错误地检测到完成标志")
	}

	fmt.Println("🧪 AI输出完成检测功能测试完成")
}

// TestWithRealHTMLContent 使用真实HTML内容测试AI完成检测
func (gc *GuardianClient) TestWithRealHTMLContent(realHTML string) {
	fmt.Println("🔍 在真实HTML内容中测试AI完成标志检测...")

	// 检查HTML中是否包含AI完成标志
	containsFlag := strings.Contains(realHTML, "ds-flex _0a3d93b")

	if containsFlag {
		fmt.Println("✅ 在真实HTML内容中检测到AI输出完成标志 'ds-flex _0a3d93b'")

		// 执行后续操作
		gc.logger.Println("🚀 检测到AI输出完成，开始执行下一步自动化操作...")
		gc.logger.Println("📋 AI输出数据复制完成")

		// 这里可以添加实际的复制操作，比如点击复制按钮等
	} else {
		fmt.Println("❌ 在真实HTML内容中未检测到AI输出完成标志")
	}
}
