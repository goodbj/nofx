package main

import (
	"fmt"
	"nofx/mcp"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🔍 验证浏览器数据目录路径配置")
	fmt.Println("================================")

	// 检查环境变量
	envDir := os.Getenv("BROWSER_DATA_DIR")
	if envDir != "" {
		fmt.Printf("✅ 环境变量 BROWSER_DATA_DIR: %s\n", envDir)
	} else {
		fmt.Printf("⚠️  环境变量 BROWSER_DATA_DIR 未设置，使用默认值\n")
	}

	// 检查常量值
	fmt.Printf("✅ GuardianBrowserDataDir 常量值: %s\n", mcp.GuardianBrowserDataDir)

	// 检查目录是否存在
	dataDir := filepath.Join("..", "data", "browser_data")
	if _, err := os.Stat(dataDir); err == nil {
		fmt.Printf("✅ 目标目录存在: %s\n", dataDir)

		// 检查目录内容
		entries, err := os.ReadDir(dataDir)
		if err == nil {
			fmt.Printf("📁 目录包含 %d 个项目\n", len(entries))
		}
	} else {
		fmt.Printf("❌ 目标目录不存在: %s\n", dataDir)
	}

	// 检查旧目录是否存在
	oldDir := filepath.Join("..", "data", "guardian_browser_data")
	if _, err := os.Stat(oldDir); err == nil {
		fmt.Printf("⚠️  旧目录仍然存在: %s\n", oldDir)
	} else {
		fmt.Printf("✅ 旧目录已清理: %s\n", oldDir)
	}

	fmt.Println("\n📋 路径配置验证完成")
}
