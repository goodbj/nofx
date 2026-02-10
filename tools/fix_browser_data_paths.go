package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔧 修复浏览器数据目录路径配置")
	fmt.Println("================================")

	// 项目根目录
	rootDir := ".."

	// 目标正确路径
	correctPath := filepath.Join(rootDir, "data", "browser_data")
	incorrectPath := filepath.Join(rootDir, "storage")

	fmt.Printf("✅ 正确路径: %s\n", correctPath)
	fmt.Printf("❌ 当前错误路径示例: %s\n", incorrectPath)
	fmt.Println()

	// 检查并创建正确的目录
	if err := os.MkdirAll(correctPath, 0755); err != nil {
		fmt.Printf("❌ 创建目录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 确保目录存在: %s\n", correctPath)

	// 查找所有需要修复的文件
	filesToCheck := []string{
		"start_nofx_dev_docker.bat",
		"stop_nofx_dev_display.bat",
		"start_nofx_dev.bat",
	}

	fixedCount := 0

	for _, filename := range filesToCheck {
		filePath := filepath.Join(rootDir, filename)
		if content, err := os.ReadFile(filePath); err == nil {
			originalContent := string(content)

			// 替换错误的路径引用
			fixedContent := strings.ReplaceAll(originalContent, "E:\\AI\\nofx_Dev\\storage", "E:\\AI\\nofx_Dev\\data\\browser_data")
			fixedContent = strings.ReplaceAll(fixedContent, "storage/", "data/browser_data/")

			if fixedContent != originalContent {
				// 写回文件
				if err := os.WriteFile(filePath, []byte(fixedContent), 0644); err != nil {
					fmt.Printf("❌ 修复文件失败 %s: %v\n", filename, err)
				} else {
					fmt.Printf("✅ 已修复文件: %s\n", filename)
					fixedCount++
				}
			} else {
				fmt.Printf("✅ 文件无需修复: %s\n", filename)
			}
		} else {
			fmt.Printf("⚠️  文件不存在或无法读取: %s\n", filename)
		}
	}

	fmt.Println()
	fmt.Printf("📋 修复完成: %d 个文件已更新\n", fixedCount)
	fmt.Println("✅ 浏览器数据目录路径已统一配置为: data/browser_data")
}
