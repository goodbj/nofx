package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔧 修复浏览器数据目录路径问题")
	fmt.Println("================================")

	// 项目根目录
	projectRoot := ".."
	dataDir := filepath.Join(projectRoot, "data")

	// 正确的浏览器数据目录
	correctBaseDir := filepath.Join(dataDir, "browser_data")

	fmt.Printf("✅ 项目根目录: %s\n", filepath.Clean(projectRoot))
	fmt.Printf("✅ 数据目录: %s\n", dataDir)
	fmt.Printf("✅ 正确的基础目录: %s\n", correctBaseDir)
	fmt.Println()

	// 检查当前目录下的错误目录
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("❌ 无法读取数据目录: %v\n", err)
		return
	}

	wrongDirs := []string{}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "browser_data_") &&
			!strings.HasPrefix(entry.Name(), "browser_data"+string(filepath.Separator)) {
			wrongDirs = append(wrongDirs, entry.Name())
		}
	}

	if len(wrongDirs) == 0 {
		fmt.Println("✅ 没有发现错误的目录结构")
		return
	}

	fmt.Printf("❌ 发现 %d 个可能错误的目录:\n", len(wrongDirs))
	for _, dir := range wrongDirs {
		fmt.Printf("  - %s\n", dir)
	}
	fmt.Println()

	// 询问是否要清理
	fmt.Print("是否要清理这些错误目录? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		fmt.Println("取消操作")
		return
	}

	// 清理错误目录
	cleanedCount := 0
	for _, dirName := range wrongDirs {
		dirPath := filepath.Join(dataDir, dirName)
		if err := os.RemoveAll(dirPath); err != nil {
			fmt.Printf("❌ 删除目录失败 %s: %v\n", dirName, err)
		} else {
			fmt.Printf("✅ 已删除目录: %s\n", dirName)
			cleanedCount++
		}
	}

	fmt.Printf("\n📋 清理完成: %d 个目录已删除\n", cleanedCount)
	fmt.Println("💡 建议: 重启后端服务，让系统重新创建正确的目录结构")
}
