package mcp

import (
	"fmt"
)

// RunDemo 运行AI完成检测演示
func RunDemo() {
	fmt.Println("🤖 Guardian Client AI完成检测功能演示")
	fmt.Println("==================================================")

	DemoAICompletionDetection()

	fmt.Println("\n==================================================")
	fmt.Println("✅ 演示完成!")
}

// 这个函数可以被外部调用以运行演示
func ExecuteDemo() {
	RunDemo()
}
