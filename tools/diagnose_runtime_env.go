package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("🔍 运行时环境变量诊断")
	fmt.Println("========================")

	// 检查所有相关环境变量
	relevantVars := []string{
		"BROWSER_DATA_DIR",
		"GUARDIAN_USER_DATA_DIR",
		"PATH",
		"HOME",
		"USERPROFILE",
	}

	fmt.Println("相关环境变量:")
	for _, varName := range relevantVars {
		value := os.Getenv(varName)
		if value != "" {
			fmt.Printf("  %s = %s\n", varName, value)
		} else {
			fmt.Printf("  %s = (未设置)\n", varName)
		}
	}

	fmt.Println("\n所有环境变量中包含 'browser' 或 'data' 的:")
	for _, env := range os.Environ() {
		if strings.Contains(strings.ToLower(env), "browser") ||
			strings.Contains(strings.ToLower(env), "data") {
			fmt.Printf("  %s\n", env)
		}
	}

	fmt.Printf("\n当前工作目录: %s\n", os.Getenv("PWD"))
	fmt.Printf("可执行文件路径: %s\n", os.Args[0])
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
}
