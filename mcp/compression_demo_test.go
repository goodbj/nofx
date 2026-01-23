package mcp

import (
	"fmt"
	"strings"
	"testing"
)

// TestCompressionDemo 演示压缩效果
func TestCompressionDemo(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("🗜️  MCP 上下文压缩功能演示\n")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// 测试场景1: 64K 模型（用户的模型）
	testCompressionScenario(t, "Ollama Qwen2.5-Coder-7B-64K", "ollama", "qwen2.5-coder-7b-64k", 65536)

	// 测试场景2: 8K 小模型
	testCompressionScenario(t, "Ollama 小模型 8K", "ollama", "qwen-8k", 8192)

	// 测试场景3: 手动设置8K（模拟用户场景）
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("📌 场景3: 手动设置为 8K 上下文（适合低显存显卡）\n")
	fmt.Println(strings.Repeat("-", 60))

	client := &Client{
		Provider: "ollama",
		Model:    "custom-model",
		config: &Config{
			EnableContextCompression: true,
			ModelContextSize:         8192, // 手动设置
		},
		logger: &testLogger{},
	}

	mockPrompt := generateMockKlinePrompt()
	compressed := client.compressContextIfNeeded(mockPrompt)

	fmt.Printf("原始长度: %d 字符\n", len(mockPrompt))
	fmt.Printf("压缩后长度: %d 字符\n", len(compressed))
	fmt.Printf("压缩率: %.1f%%\n", float64(len(compressed))/float64(len(mockPrompt))*100)
	fmt.Printf("节省: %d 字符\n\n", len(mockPrompt)-len(compressed))
}

func testCompressionScenario(t *testing.T, name, provider, model string, contextSize int) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Printf("📌 场景: %s\n", name)
	fmt.Println(strings.Repeat("-", 60))

	client := &Client{
		Provider: provider,
		Model:    model,
		config:   DefaultConfig(),
		logger:   &testLogger{},
	}

	mockPrompt := generateMockKlinePrompt()
	usagePercent := float64(len(mockPrompt)) / float64(contextSize) * 100

	fmt.Printf("模型上下文: %d (%.0fK)\n", contextSize, float64(contextSize)/1024)
	fmt.Printf("Prompt 长度: %d 字符\n", len(mockPrompt))
	fmt.Printf("使用率: %.1f%%\n", usagePercent)

	compressed := client.compressContextIfNeeded(mockPrompt)

	if len(compressed) < len(mockPrompt) {
		fmt.Printf("✅ 触发压缩\n")
		fmt.Printf("   压缩后: %d 字符 (%.1f%%)\n", len(compressed), float64(len(compressed))/float64(len(mockPrompt))*100)
		fmt.Printf("   节省: %d 字符\n", len(mockPrompt)-len(compressed))
	} else {
		fmt.Printf("ℹ️  未触发压缩（使用率<60%%）\n")
	}
}
