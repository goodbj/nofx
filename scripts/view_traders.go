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

// Trader 结构体定义，匹配数据库中的交易者表
type Trader struct {
	ID                   string      `gorm:"primaryKey;column:id" json:"id"`
	Name                 string      `gorm:"column:name;not null" json:"name"`
	ExchangeID           string      `gorm:"column:exchange_id;not null" json:"exchange_id"`
	ExchangeType         string      `gorm:"column:exchange_type;not null" json:"exchange_type"`
	APIKey               string      `gorm:"column:api_key" json:"api_key"`
	APISecret            string      `gorm:"column:api_secret" json:"api_secret"`
	Passphrase           *string     `gorm:"column:passphrase" json:"passphrase"`
	Config               interface{} `gorm:"column:config;type:text" json:"config"`
	Enabled              bool        `gorm:"column:enabled" json:"enabled"`
	AutoSync             bool        `gorm:"column:auto_sync" json:"auto_sync"`
	MaxPositions         int         `gorm:"column:max_positions" json:"max_positions"`
	Leverage             float64     `gorm:"column:leverage" json:"leverage"`
	MinOrderSize         float64     `gorm:"column:min_order_size" json:"min_order_size"`
	MaxOrderSize         float64     `gorm:"column:max_order_size" json:"max_order_size"`
	RiskPerTrade         float64     `gorm:"column:risk_per_trade" json:"risk_per_trade"`
	StopLossPercentage   float64     `gorm:"column:stop_loss_percentage" json:"stop_loss_percentage"`
	TakeProfitPercentage float64     `gorm:"column:take_profit_percentage" json:"take_profit_percentage"`
	CreatedAt            interface{} `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            interface{} `gorm:"column:updated_at" json:"updated_at"`
	Deprecated           *bool       `gorm:"column:deprecated" json:"deprecated"`
	DeprecatedReason     *string     `gorm:"column:deprecated_reason" json:"deprecated_reason"`
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

	fmt.Println("=== NOFX 交易者表详细查看 ===")
	fmt.Printf("正在查看数据库: %s\n\n", dbPath)

	// 获取交易者表的所有记录
	var traders []Trader
	result := db.Table("traders").Find(&traders)
	if result.Error != nil {
		log.Fatalf("无法查询交易者表: %v", result.Error)
	}

	fmt.Printf("交易者表中共有 %d 条记录\n\n", len(traders))

	for i, trader := range traders {
		fmt.Printf("--- 交易者 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", trader.ID)
		fmt.Printf("名称: %s\n", trader.Name)
		fmt.Printf("交易所ID: %s\n", trader.ExchangeID)
		fmt.Printf("交易所类型: %s\n", trader.ExchangeType)
		fmt.Printf("API密钥: %s\n", trader.APIKey)
		fmt.Printf("启用: %t\n", trader.Enabled)
		fmt.Printf("自动同步: %t\n", trader.AutoSync)
		fmt.Printf("最大持仓数: %d\n", trader.MaxPositions)
		fmt.Printf("杠杆: %.2f\n", trader.Leverage)
		fmt.Printf("最小订单大小: %.2f\n", trader.MinOrderSize)
		fmt.Printf("最大订单大小: %.2f\n", trader.MaxOrderSize)
		fmt.Printf("每笔交易风险: %.2f\n", trader.RiskPerTrade)
		fmt.Printf("止损百分比: %.2f\n", trader.StopLossPercentage)
		fmt.Printf("止盈百分比: %.2f\n", trader.TakeProfitPercentage)

		if trader.Deprecated != nil {
			fmt.Printf("已弃用: %t\n", *trader.Deprecated)
			if trader.DeprecatedReason != nil {
				fmt.Printf("弃用原因: %s\n", *trader.DeprecatedReason)
			}
		} else {
			fmt.Println("已弃用: false")
		}

		// 解析并格式化配置字段
		fmt.Println("配置:")
		if trader.Config != nil {
			// 尝试将其转换为 JSON 字符串并格式化
			configBytes, err := json.MarshalIndent(trader.Config, "  ", "  ")
			if err != nil {
				fmt.Printf("  %v\n", trader.Config)
			} else {
				fmt.Printf("  %s\n", string(configBytes))
			}
		} else {
			fmt.Println("  无配置")
		}

		fmt.Printf("创建时间: %v\n", trader.CreatedAt)
		fmt.Printf("更新时间: %v\n", trader.UpdatedAt)
		fmt.Println()
	}

	if len(traders) == 0 {
		fmt.Println("没有找到任何交易者记录")
	}

	fmt.Println("=== 查看完成 ===")
}
