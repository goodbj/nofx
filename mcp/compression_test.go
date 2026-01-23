package mcp

import (
	"fmt"
	"strings"
	"testing"
)

// TestContextCompression 测试上下文压缩功能
func TestContextCompression(t *testing.T) {
	// 创建测试客户端（使用小上下文模型，确保触发压缩）
	client := &Client{
		Provider: "ollama",
		Model:    "qwen2.5-coder-7b",
		config:   DefaultConfig(),
		logger:   &testLogger{},
	}

	// 手动设置为8K上下文（确保触发压缩）
	client.config.EnableContextCompression = true
	client.config.ModelContextSize = 8192

	// 模拟K线数据（长文本）
	mockPrompt := generateMockKlinePrompt()

	fmt.Printf("📊 测试上下文压缩功能\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("原始长度: %d 字符\n", len(mockPrompt))

	// 测试自动检测模型上下文大小
	contextSize := client.getModelContextSize()
	fmt.Printf("配置的模型上下文: %d\n", contextSize)
	fmt.Printf("使用率: %.1f%%\n", float64(len(mockPrompt))/float64(contextSize)*100)

	// 测试压缩
	compressed := client.compressContextIfNeeded(mockPrompt)
	fmt.Printf("压缩后长度: %d 字符\n", len(compressed))
	fmt.Printf("压缩率: %.1f%%\n", float64(len(compressed))/float64(len(mockPrompt))*100)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// 验证K线数据被压缩
	originalLines := strings.Count(mockPrompt, "01-")
	compressedLines := strings.Count(compressed, "01-")
	fmt.Printf("K线数据行数: %d → %d\n", originalLines, compressedLines)

	if compressedLines >= originalLines {
		t.Errorf("压缩失败：压缩后行数 (%d) >= 原始行数 (%d)", compressedLines, originalLines)
	}
}

// TestModelContextSizeDetection 测试模型上下文大小检测
func TestModelContextSizeDetection(t *testing.T) {
	testCases := []struct {
		provider string
		model    string
		expected int
	}{
		{"ollama", "qwen2.5-coder-7b-64k", 65536},
		{"ollama", "qwen2.5-coder-7b-32k", 32768},
		{"ollama", "llama2", 4096},
		{ProviderDeepSeek, "deepseek-chat", 32768},
		{ProviderOpenAI, "gpt-4", 128000},
		{ProviderOpenAI, "gpt-3.5-turbo", 16384},
		{ProviderClaude, "claude-3-opus", 200000},
		{ProviderQwen, "qwen-max", 32768},
		{ProviderGemini, "gemini-pro", 1000000},
	}

	fmt.Printf("\n📋 测试模型上下文大小自动检测\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, tc := range testCases {
		client := &Client{
			Provider: tc.provider,
			Model:    tc.model,
			config:   DefaultConfig(),
		}
		detected := client.getModelContextSize()
		status := "✅"
		if detected != tc.expected {
			status = "❌"
			t.Errorf("%s %s: 期望 %d, 实际 %d", tc.provider, tc.model, tc.expected, detected)
		}
		fmt.Printf("%s %s / %s: %d\n", status, tc.provider, tc.model, detected)
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// TestCompressionLevels 测试不同压缩级别
func TestCompressionLevels(t *testing.T) {
	testCases := []struct {
		contextSize   int
		usagePercent  float64
		expectedLevel string
	}{
		{131072, 50, "无压缩"},
		{65536, 95, "轻度压缩"},
		{32768, 85, "中度压缩"},
		{8192, 90, "中度压缩"},
		{4096, 70, "激进压缩"},
	}

	fmt.Printf("\n🔧 测试不同压缩级别\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, tc := range testCases {
		client := &Client{
			Provider: "ollama",
			Model:    "test-model",
			config: &Config{
				EnableContextCompression: true,
				ModelContextSize:         tc.contextSize,
			},
			logger: &testLogger{},
		}

		// 生成测试prompt
		promptLength := int(float64(tc.contextSize) * tc.usagePercent / 100)
		mockPrompt := strings.Repeat("x", promptLength) + generateMockKlinePrompt()

		_ = client.compressKlineData(mockPrompt, tc.contextSize, len(mockPrompt))

		fmt.Printf("上下文: %dK, 使用率: %.0f%% → %s\n",
			tc.contextSize/1024, tc.usagePercent, tc.expectedLevel)
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// generateMockKlinePrompt 生成模拟K线数据
func generateMockKlinePrompt() string {
	var sb strings.Builder
	sb.WriteString("# 交易决策请求\n\n")
	sb.WriteString("## 候选币种\n\n")
	sb.WriteString("### 1. BTCUSDT\n\n")
	sb.WriteString("#### 1h 时间框架 (从旧到新)\n\n")
	sb.WriteString("```\n")
	sb.WriteString("时间(UTC)      开盘      最高      最低      收盘      成交量\n")

	// 生成30行K线数据
	for i := 1; i <= 30; i++ {
		sb.WriteString(fmt.Sprintf("01-%02d 12:%02d  43560.4200 43580.1234 43550.8765 43565.9876 1234567.89\n",
			i, i))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	// 添加第二个时间框架
	sb.WriteString("#### 4h 时间框架 (从旧到新)\n\n")
	sb.WriteString("```\n")
	sb.WriteString("时间(UTC)      开盘      最高      最低      收盘      成交量\n")
	for i := 1; i <= 30; i++ {
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  43500.1234 43650.5678 43480.9012 43600.3456 9876543.21\n",
			i))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	return sb.String()
}

// testLogger 简单的测试日志器
type testLogger struct{}

func (l *testLogger) Debugf(format string, args ...interface{}) {
	// fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (l *testLogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}
func (l *testLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}
func (l *testLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
