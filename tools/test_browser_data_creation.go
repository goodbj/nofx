package main

import (
	"fmt"
	"nofx/mcp"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🔍 浏览器数据目录创建测试")
	fmt.Println("========================")

	// 检查环境变量
	envDir := os.Getenv("BROWSER_DATA_DIR")
	fmt.Printf("环境变量 BROWSER_DATA_DIR: '%s'\n", envDir)

	// 检查常量值
	fmt.Printf("GuardianBrowserDataDir 常量: '%s'\n", mcp.GuardianBrowserDataDir)

	// 模拟创建交易员浏览器数据目录的逻辑
	testTraderID := "test_trader_123"
	dirName := filepath.Join(mcp.GuardianBrowserDataDir, testTraderID)
	fmt.Printf("将要创建的目录名: '%s'\n", dirName)

	// 检查完整路径
	absPath, err := filepath.Abs(dirName)
	if err != nil {
		fmt.Printf("路径解析错误: %v\n", err)
		return
	}
	fmt.Printf("绝对路径: '%s'\n", absPath)

	// 检查父目录是否存在
	parentDir := filepath.Dir(absPath)
	fmt.Printf("父目录: '%s'\n", parentDir)

	if _, err := os.Stat(parentDir); err == nil {
		fmt.Printf("✅ 父目录存在: %s\n", parentDir)
	} else {
		fmt.Printf("❌ 父目录不存在: %s\n", parentDir)
	}

	// 尝试创建测试目录
	fmt.Println("\n尝试创建测试目录...")
	if err := os.MkdirAll(dirName, 0755); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功创建目录: %s\n", dirName)

		// 验证创建位置
		if _, err := os.Stat(dirName); err == nil {
			fmt.Printf("✅ 目录确实存在: %s\n", dirName)

			// 检查目录位置是否正确
			absCreated, _ := filepath.Abs(dirName)
			expectedParent := filepath.Join("E:\\AI\\nofx_Dev\\data\\Chrome_browser_data")

			if filepath.HasPrefix(absCreated, expectedParent) {
				fmt.Printf("✅ 目录位置正确，在 %s 下\n", expectedParent)
			} else {
				fmt.Printf("❌ 目录位置错误！期望在 %s 下，实际在 %s\n", expectedParent, absCreated)
			}
		}

		// 清理测试目录
		os.RemoveAll(dirName)
		fmt.Printf("🧹 已清理测试目录: %s\n", dirName)
	}
}
