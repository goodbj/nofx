package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Strategy 结构体定义，匹配数据库中的策略表
type Strategy struct {
	ID          string      `gorm:"primaryKey;column:id" json:"id"`
	Name        string      `gorm:"column:name;not null" json:"name"`
	Description string      `gorm:"column:description" json:"description"`
	Config      interface{} `gorm:"column:config;type:text" json:"config"`
	CreatedAt   interface{} `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   interface{} `gorm:"column:updated_at" json:"updated_at"`
}

// 从 store/strategy.go 获取的 KlineConfig 结构
type KlineConfig struct {
	// primary timeframe: "1m", "3m", "5m", "15m", "1h", "4h"
	PrimaryTimeframe string `json:"primary_timeframe"`
	// primary timeframe K-line count
	PrimaryCount int `json:"primary_count"`
	// longer timeframe
	LongerTimeframe string `json:"longer_timeframe,omitempty"`
	// longer timeframe K-line count
	LongerCount int `json:"longer_count,omitempty"`
	// whether to enable multi-timeframe analysis
	EnableMultiTimeframe bool `json:"enable_multi_timeframe"`
	// selected timeframe list (new: supports multi-timeframe selection)
	SelectedTimeframes []string `json:"selected_timeframes,omitempty"`
	// custom K-line counts for each timeframe
	TimeframeCounts map[string]int `json:"timeframe_counts,omitempty"`
	// trading style preset ("short", "medium", "long")
	TradingStylePreset string `json:"trading_style_preset,omitempty"`
}

func main() {
	// 设置数据库路径
	dbPath := filepath.Join("data", "data.db")

	// 检查数据库文件是否存在
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("数据库文件不存在: %s\n", dbPath)
		return
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	fmt.Println("=== NOFX 策略表详细查看 ===")
	fmt.Printf("正在查看数据库: %s\n\n", dbPath)

	// 获取策略表的所有记录
	var strategies []Strategy
	result := db.Table("strategies").Find(&strategies)
	if result.Error != nil {
		log.Fatalf("无法查询策略表: %v", result.Error)
	}

	fmt.Printf("策略表中共有 %d 条记录\n\n", len(strategies))

	for i, strategy := range strategies {
		fmt.Printf("--- 策略 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", strategy.ID)
		fmt.Printf("名称: %s\n", strategy.Name)
		fmt.Printf("描述: %s\n", strategy.Description)

		// 解析并格式化配置字段
		fmt.Println("配置:")
		if strategy.Config != nil {
			// 尝试将其转换为 JSON 字符串并格式化
			configBytes, err := json.MarshalIndent(strategy.Config, "  ", "  ")
			if err != nil {
				fmt.Printf("  %v\n", strategy.Config)
			} else {
				fmt.Printf("  %s\n", string(configBytes))

				// 尝试解析为 KlineConfig 结构
				var klineConfig KlineConfig
				if configMap, ok := strategy.Config.(map[string]interface{}); ok {
					configBytes, _ := json.Marshal(configMap)
					if err := json.Unmarshal(configBytes, &klineConfig); err == nil {
						fmt.Println("  K线配置解析:")
						fmt.Printf("    主要时间周期: %s\n", klineConfig.PrimaryTimeframe)
						fmt.Printf("    主要K线数量: %d\n", klineConfig.PrimaryCount)
						fmt.Printf("    较长时间周期: %s\n", klineConfig.LongerTimeframe)
						fmt.Printf("    较长K线数量: %d\n", klineConfig.LongerCount)
						fmt.Printf("    启用多时间周期分析: %t\n", klineConfig.EnableMultiTimeframe)
						fmt.Printf("    选择的时间周期: %v\n", klineConfig.SelectedTimeframes)
						fmt.Printf("    时间周期计数映射: %v\n", klineConfig.TimeframeCounts)
						fmt.Printf("    交易风格预设: %s\n", klineConfig.TradingStylePreset)
					}
				}
			}
		} else {
			fmt.Println("  无配置")
		}

		fmt.Printf("创建时间: %v\n", strategy.CreatedAt)
		fmt.Printf("更新时间: %v\n", strategy.UpdatedAt)
		fmt.Println()
	}

	if len(strategies) == 0 {
		fmt.Println("没有找到任何策略记录")
	}

	fmt.Println("=== 查看完成 ===")
}
