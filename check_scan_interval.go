package main

import (
	"fmt"
	"log"
	"nofx/store"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 连接到数据库
	db, err := gorm.Open(sqlite.Open("traders.db"), &gorm.Config{})
	if err != nil {
		log.Printf("尝试连接SQLite数据库失败: %v", err)

		// 尝试其他可能的数据库文件位置
		dbPaths := []string{"./traders.db", "./data/traders.db", "../traders.db", "db.sqlite", "./db.sqlite"}
		for _, path := range dbPaths {
			db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
			if err == nil {
				log.Printf("成功连接到数据库: %s", path)
				break
			}
		}

		if err != nil {
			log.Fatal("无法连接到任何数据库文件")
		}
	}

	// 查询所有交易员的扫描间隔配置
	var traders []store.Trader
	result := db.Find(&traders)
	if result.Error != nil {
		log.Fatalf("查询交易员数据失败: %v", result.Error)
	}

	fmt.Println("交易员扫描间隔配置检查:")
	fmt.Println("=========================")
	for _, trader := range traders {
		fmt.Printf("交易员: %s (ID: %s)\n", trader.Name, trader.ID[:8]+"...")
		fmt.Printf("  扫描间隔(分钟): %d\n", trader.ScanIntervalMinutes)
		fmt.Printf("  是否运行: %t\n", trader.IsRunning)
		fmt.Printf("  AI模型: %s\n", trader.AIModelID)
		fmt.Printf("  交易所: %s\n", trader.ExchangeID)
		fmt.Println()
	}
}
