package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔧 修复环境变量加载问题")
	fmt.Println("========================")
	
	// 读取 .env 文件并设置环境变量
	envFile := filepath.Join("..", ".env")
	file, err := os.Open(envFile)
	if err != nil {
		fmt.Printf("❌ 无法读取 .env 文件: %v\n", err)
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	found := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "BROWSER_DATA_DIR=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				if value != "" {
					os.Setenv("BROWSER_DATA_DIR", value)
					fmt.Printf("✅ 设置环境变量 BROWSER_DATA_DIR = %s\n", value)
					found = true
					break
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ 读取文件错误: %v\n", err)
		return
	}
	
	if !found {
		fmt.Println("❌ 未在 .env 文件中找到 BROWSER_DATA_DIR 设置")
		// 设置默认值
		defaultPath := "./data/browser_data"
		os.Setenv("BROWSER_DATA_DIR", defaultPath)
		fmt.Printf("🔧 设置默认环境变量 BROWSER_DATA_DIR = %s\n", defaultPath)
	}
	
	// 验证设置结果
	envValue := os.Getenv("BROWSER_DATA_DIR")
	fmt.Printf("✅ 最终环境变量值: '%s'\n", envValue)
	
	// 确保目录存在
	if envValue != "" {
		absPath, err := filepath.Abs(filepath.Join("..", envValue))
		if err != nil {
			fmt.Printf("❌ 无法解析路径: %v\n", err)
			return
		}
		
		fmt.Printf("📁 绝对路径: %s\n", absPath)
		
		if err := os.MkdirAll(absPath, 0755); err != nil {
			fmt.Printf("❌ 创建目录失败: %v\n", err)
			return
		}
		fmt.Printf("✅ 确保目录存在: %s\n", absPath)
	}
	
	fmt.Println("\n✅ 环境变量修复完成")
}