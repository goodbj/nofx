package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func main() {
	fmt.Println("🔍 NOFX 测试代码索引")
	fmt.Println("==================")
	fmt.Println("以下是可以运行的各种测试脚本：")
	fmt.Println()

	fmt.Println("1. 🔗 Binance Testnet 连接测试")
	fmt.Println("   用途：测试与币安期货测试网的连接")
	fmt.Println("   命令：go run test/test_binance_testnet.go")
	fmt.Println()

	fmt.Println("2. 📊 Binance 详细连接测试")
	fmt.Println("   用途：详细测试币安期货测试网各项功能")
	fmt.Println("   命令：go run test/test_binance_futures_connection.go")
	fmt.Println()

	fmt.Println("3. 🤖 Binance 系统交易者测试")
	fmt.Println("   用途：使用系统内置交易者结构测试币安连接")
	fmt.Println("   命令：go run test/test_binance_futures_trader.go")
	fmt.Println()

	fmt.Println("💡 提示：")
	fmt.Println("- 所有测试脚本都位于 ./test/ 目录中")
	fmt.Println("- 运行前请确保已安装必要的依赖包")
	fmt.Println("- 部分测试可能需要设置环境变量或使用内置的测试密钥")
	fmt.Println()

	// 尝试列出test目录中的所有测试文件
	fmt.Println("📁 当前测试文件列表：")
	cmd := exec.Command("dir", "test", "/b")
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		cmd = exec.Command("ls", "-la", "test")
	}
	output, err := cmd.Output()
	if err == nil {
		fmt.Printf("%s", output)
	} else {
		fmt.Println("   无法列出文件")
	}
}
