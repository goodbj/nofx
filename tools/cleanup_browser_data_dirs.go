package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🧹 清理错误的浏览器数据目录")
	fmt.Println("========================")

	// 项目根目录
	dataDir := filepath.Join(".", "data")

	// 读取 data 目录下的所有子目录
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("❌ 无法读取数据目录: %v\n", err)
		return
	}

	// 找出所有以 Chrome_browser_data_ 开头但在 data 目录下的目录（错误位置）
	var wrongDirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, "Chrome_browser_data_") &&
				!strings.Contains(name, string(filepath.Separator)) {
				// 这是一个错误位置的目录
				wrongDirs = append(wrongDirs, name)
			}
		}
	}

	if len(wrongDirs) == 0 {
		fmt.Println("✅ 没有发现错误的目录结构")
		return
	}

	fmt.Printf("❌ 发现 %d 个错误位置的目录:\n", len(wrongDirs))
	for i, dir := range wrongDirs {
		fmt.Printf("  %d. %s\n", i+1, dir)
	}

	fmt.Println("\n🗑️  开始清理错误的目录...")
	for _, dir := range wrongDirs {
		fullPath := filepath.Join(dataDir, dir)
		if err := os.RemoveAll(fullPath); err != nil {
			fmt.Printf("❌ 删除目录 %s 失败: %v\n", dir, err)
		} else {
			fmt.Printf("✅ 已删除错误目录: %s\n", dir)
		}
	}

	fmt.Println("\n✅ 清理完成!")

	// 验证清理结果
	fmt.Println("\n🔍 验证清理结果:")
	correctDir := filepath.Join(dataDir, "Chrome_browser_data")
	if _, err := os.Stat(correctDir); err == nil {
		fmt.Printf("✅ 正确的主目录存在: %s\n", correctDir)

		// 检查主目录下的内容
		subEntries, err := os.ReadDir(correctDir)
		if err == nil {
			fmt.Printf("📁 主目录包含 %d 个子项\n", len(subEntries))
		}
	} else {
		fmt.Printf("⚠️  正确的主目录不存在: %s\n", correctDir)
	}
}
