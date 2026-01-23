package main

import (
	"fmt"
	"nofx/mcp"
	"strings"
)

// 模拟真实的交易上下文数据
func generateRealTradingContext() string {
	var sb strings.Builder
	
	// 1. 系统提示词部分（约2000字符）
	sb.WriteString(`# 📊 交易决策AI助手

你是一个专业的加密货币交易决策AI，基于技术分析为用户提供交易建议。

## 决策输出格式（JSON）
你必须严格按照以下JSON格式输出决策：
[
  {
    "action": "open_position",
    "symbol": "BTCUSDT",
    "side": "LONG",
    "leverage": 10,
    "position_size_usd": 500,
    "stop_loss": 43000.50,
    "take_profit": 44500.25,
    "confidence": 85,
    "risk_usd": 25.50,
    "reasoning": "技术分析理由..."
  }
]

---

`)

	// 2. 账户状态（约500字符）
	sb.WriteString(`## 账户状态

总权益: 10000.00 USDT | 可用余额: 7500.00 USDT (75.0%) | 总盈亏: +5.23% | 保证金使用率: 25.0% | 持仓数: 2

`)

	// 3. 历史交易统计（约800字符）
	sb.WriteString(`## 历史交易统计

**指标说明**:
- 盈利因子: 总盈利 ÷ 总亏损（>1表示盈利，>1.5为良好，>2为优秀）
- 夏普比率: (平均收益 - 无风险收益) ÷ 收益标准差（>1良好，>2优秀）
- 盈亏比: 平均盈利 ÷ 平均亏损（>1.5为良好，>2为优秀）
- 最大回撤: 资金曲线从峰值到谷底的最大跌幅（<20%为低风险）

**当前数据**:
- 总交易: 45 笔
- 盈利因子: 1.85
- 夏普比率: 1.42
- 盈亏比: 2.15
- 总盈亏: +523.50 USDT
- 平均盈利: +85.50 USDT
- 平均亏损: -39.80 USDT
- 最大回撤: 12.5%

`)

	// 4. 当前持仓（约600字符）
	sb.WriteString(`## 当前持仓

1. BTCUSDT LONG | 进场 43200.0000 当前 43560.5000 | 数量 0.2315 | 仓位价值 10083.91 USDT | 盈亏 +0.83% | 盈亏金额 +83.91 USDT | 峰值盈亏 1.25% | 杠杆 10x | 保证金 1008 USDT | 强平价 38880.0000

2. ETHUSDT LONG | 进场 2600.0000 当前 2635.8000 | 数量 1.9230 | 仓位价值 5068.25 USDT | 盈亏 +1.38% | 盈亏金额 +68.25 USDT | 峰值盈亏 1.80% | 杠杆 10x | 保证金 507 USDT | 强平价 2340.0000

`)

	// 5. 候选币种（重点：包含大量K线数据，这是压缩的主要目标）
	sb.WriteString(`## 候选币种

### 1. BTCUSDT

当前价格: 43560.4200

#### 1h 时间框架 (从旧到新)

` + "```" + `
时间(UTC)      开盘      最高      最低      收盘      成交量
`)

	// 生成50根1h K线（每根约80字符，共4000字符）
	for i := 1; i <= 50; i++ {
		base := 43000.0 + float64(i)*10.0
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f\n",
			i, base, base+50.0, base-30.0, base+20.0, 1234567.89+float64(i)*1000))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	// 6. 4h K线（再增加50根，约4000字符）
	sb.WriteString(`#### 4h 时间框架 (从旧到新)

` + "```" + `
时间(UTC)      开盘      最高      最低      收盘      成交量
`)

	for i := 1; i <= 50; i++ {
		base := 42500.0 + float64(i)*20.0
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f\n",
			i, base, base+100.0, base-60.0, base+40.0, 9876543.21+float64(i)*5000))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	// 7. 1d K线（再增加30根，约2400字符）
	sb.WriteString(`#### 1d 时间框架 (从旧到新)

` + "```" + `
时间(UTC)      开盘      最高      最低      收盘      成交量
`)

	for i := 1; i <= 30; i++ {
		base := 40000.0 + float64(i)*100.0
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f\n",
			i, base, base+500.0, base-300.0, base+200.0, 50000000.0+float64(i)*100000))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	// 8. 添加第二个币种（ETH）
	sb.WriteString(`### 2. ETHUSDT

当前价格: 2635.8000

#### 1h 时间框架 (从旧到新)

` + "```" + `
时间(UTC)      开盘      最高      最低      收盘      成交量
`)

	for i := 1; i <= 50; i++ {
		base := 2600.0 + float64(i)*0.5
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f\n",
			i, base, base+5.0, base-3.0, base+2.0, 234567.89+float64(i)*100))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	// 9. ETH 4h K线
	sb.WriteString(`#### 4h 时间框架 (从旧到新)

` + "```" + `
时间(UTC)      开盘      最高      最低      收盘      成交量
`)

	for i := 1; i <= 50; i++ {
		base := 2550.0 + float64(i)*1.5
		sb.WriteString(fmt.Sprintf("01-%02d 00:00  %-9.4f %-9.4f %-9.4f %-9.4f %-12.2f\n",
			i, base, base+10.0, base-6.0, base+4.0, 987654.32+float64(i)*500))
	}
	sb.WriteString("    <- 当前\n")
	sb.WriteString("```\n\n")

	return sb.String()
}

func main() {
	fmt.Println("🧪 测试 MCP 上下文压缩功能")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// 生成真实的交易上下文数据
	realContext := generateRealTradingContext()
	fmt.Printf("✅ 生成真实交易上下文: %d 字符\n", len(realContext))
	fmt.Println()

	// 测试场景1: 你的实际配置（64K模型，自动检测）
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("📌 场景1: 你的实际模型配置")
	fmt.Println("   模型: Ollama zoyer2/Qwen2.5-Coder-7B-Instruct-Q4_K_M-64K-CLINE")
	fmt.Println(strings.Repeat("-", 70))

	client1 := &mcp.Client{}
	client1 = mcp.NewClient().(*mcp.Client)
	client1.Provider = "ollama"
	client1.Model = "zoyer2/Qwen2.5-Coder-7B-Instruct-Q4_K_M-64K-CLINE"
	
	contextSize1 := client1.getModelContextSize()
	usagePercent1 := float64(len(realContext)) / float64(contextSize1) * 100
	
	fmt.Printf("   检测到上下文大小: %d (%.0fK)\n", contextSize1, float64(contextSize1)/1024)
	fmt.Printf("   Prompt 长度: %d 字符\n", len(realContext))
	fmt.Printf("   使用率: %.1f%%\n", usagePercent1)
	
	compressed1 := client1.compressContextIfNeeded(realContext)
	
	if len(compressed1) < len(realContext) {
		fmt.Printf("   ✅ 触发压缩!\n")
		fmt.Printf("   压缩后: %d 字符\n", len(compressed1))
		fmt.Printf("   压缩率: %.1f%%\n", float64(len(compressed1))/float64(len(realContext))*100)
		fmt.Printf("   节省: %d 字符 (%.1f%%)\n", 
			len(realContext)-len(compressed1),
			float64(len(realContext)-len(compressed1))/float64(len(realContext))*100)
	} else {
		fmt.Printf("   ℹ️  未触发压缩（使用率 %.1f%% < 60%%）\n", usagePercent1)
	}
	fmt.Println()

	// 测试场景2: 手动设置为16K（模拟内存紧张场景）
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("📌 场景2: 手动设置 16K 上下文（适合内存紧张时）")
	fmt.Println(strings.Repeat("-", 70))

	client2 := mcp.NewClient().(*mcp.Client)
	client2.Provider = "ollama"
	client2.Model = "qwen2.5-coder-7b"
	client2.config.ModelContextSize = 16384 // 手动设置
	
	contextSize2 := client2.getModelContextSize()
	usagePercent2 := float64(len(realContext)) / float64(contextSize2) * 100
	
	fmt.Printf("   配置的上下文大小: %d (%.0fK)\n", contextSize2, float64(contextSize2)/1024)
	fmt.Printf("   Prompt 长度: %d 字符\n", len(realContext))
	fmt.Printf("   使用率: %.1f%%\n", usagePercent2)
	
	compressed2 := client2.compressContextIfNeeded(realContext)
	
	if len(compressed2) < len(realContext) {
		fmt.Printf("   ✅ 触发压缩!\n")
		fmt.Printf("   压缩后: %d 字符\n", len(compressed2))
		fmt.Printf("   压缩率: %.1f%%\n", float64(len(compressed2))/float64(len(realContext))*100)
		fmt.Printf("   节省: %d 字符 (%.1f%%)\n", 
			len(realContext)-len(compressed2),
			float64(len(realContext)-len(compressed2))/float64(len(realContext))*100)
		
		// 统计K线数据变化
		originalKlines := strings.Count(realContext, "01-")
		compressedKlines := strings.Count(compressed2, "01-")
		fmt.Printf("   K线数据: %d 根 → %d 根\n", originalKlines, compressedKlines)
	} else {
		fmt.Printf("   ℹ️  未触发压缩\n")
	}
	fmt.Println()

	// 测试场景3: 手动设置为8K（激进压缩）
	fmt.Println(strings.Repeat("-", 70))
	fmt.Println("📌 场景3: 手动设置 8K 上下文（激进压缩模式）")
	fmt.Println(strings.Repeat("-", 70))

	client3 := mcp.NewClient().(*mcp.Client)
	client3.Provider = "ollama"
	client3.Model = "qwen2.5-coder-7b"
	client3.config.ModelContextSize = 8192 // 手动设置
	
	contextSize3 := client3.getModelContextSize()
	usagePercent3 := float64(len(realContext)) / float64(contextSize3) * 100
	
	fmt.Printf("   配置的上下文大小: %d (%.0fK)\n", contextSize3, float64(contextSize3)/1024)
	fmt.Printf("   Prompt 长度: %d 字符\n", len(realContext))
	fmt.Printf("   使用率: %.1f%%\n", usagePercent3)
	
	compressed3 := client3.compressContextIfNeeded(realContext)
	
	if len(compressed3) < len(realContext) {
		fmt.Printf("   ✅ 触发激进压缩!\n")
		fmt.Printf("   压缩后: %d 字符\n", len(compressed3))
		fmt.Printf("   压缩率: %.1f%%\n", float64(len(compressed3))/float64(len(realContext))*100)
		fmt.Printf("   节省: %d 字符 (%.1f%%)\n", 
			len(realContext)-len(compressed3),
			float64(len(realContext)-len(compressed3))/float64(len(realContext))*100)
		
		// 统计K线数据变化
		originalKlines := strings.Count(realContext, "01-")
		compressedKlines := strings.Count(compressed3, "01-")
		fmt.Printf("   K线数据: %d 根 → %d 根 (保留 %.1f%%)\n", 
			originalKlines, compressedKlines,
			float64(compressedKlines)/float64(originalKlines)*100)
		
		// 检查是否移除了成交量
		hasVolumeOriginal := strings.Contains(realContext, "成交量")
		hasVolumeCompressed := strings.Count(compressed3, "成交量") > 0
		if hasVolumeOriginal && !hasVolumeCompressed {
			fmt.Printf("   ✅ 已移除成交量列（进一步压缩）\n")
		}
	} else {
		fmt.Printf("   ℹ️  未触发压缩\n")
	}
	fmt.Println()

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ 测试完成！")
	fmt.Println()
	fmt.Println("💡 建议:")
	fmt.Println("   - 你的64K模型在大多数情况下不需要压缩")
	fmt.Println("   - 如果遇到内存不足(OOM)，可设置 MCP_MODEL_CONTEXT_SIZE=16384")
	fmt.Println("   - 如果推理速度慢，可设置 MCP_MODEL_CONTEXT_SIZE=8192 激进压缩")
	fmt.Println()
}
