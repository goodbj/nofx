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

	fmt.Println("=== NOFX 数据库探索 ===")
	fmt.Printf("正在查看数据库: %s\n\n", dbPath)

	// 获取所有表名
	var tables []string
	err = db.Raw("SELECT name FROM sqlite_master WHERE type='table';").Scan(&tables).Error
	if err != nil {
		log.Fatalf("无法获取表列表: %v", err)
	}

	fmt.Println("数据库中存在的表:")
	for i, table := range tables {
		fmt.Printf("%d. %s\n", i+1, table)
	}

	fmt.Println("\n=== 表结构详情 ===")

	// 查看各个表的结构和数据量
	for _, table := range tables {
		if table == "sqlite_sequence" || table == "sqlite_stat1" || table == "schema_migrations" {
			continue // 跳过系统表
		}

		fmt.Printf("\n--- 表: %s ---\n", table)

		// 定义列结构体
		type ColumnInfo struct {
			Cid          int64   `gorm:"column:cid"`
			Name         string  `gorm:"column:name"`
			Type         string  `gorm:"column:type"`
			NotNull      int64   `gorm:"column:notnull"`
			DefaultValue *string `gorm:"column:dflt_value"`
			PrimaryKey   int64   `gorm:"column:pk"`
		}

		// 获取表结构
		var columns []ColumnInfo
		err = db.Raw(fmt.Sprintf("PRAGMA table_info(%s);", table)).Scan(&columns).Error
		if err != nil {
			fmt.Printf("无法获取表 %s 的结构: %v\n", table, err)
			continue
		}

		fmt.Println("列结构:")
		for _, col := range columns {
			fmt.Printf("  - %s (%s)", col.Name, col.Type)
			if col.NotNull == 1 {
				fmt.Print(" NOT NULL")
			}
			if col.PrimaryKey == 1 {
				fmt.Print(" PRIMARY KEY")
			}
			if col.DefaultValue != nil {
				fmt.Printf(" DEFAULT '%s'", *col.DefaultValue)
			}
			fmt.Println()
		}

		// 计算表中的记录数
		var count int64
		err = db.Table(table).Count(&count).Error
		if err != nil {
			fmt.Printf("无法计算表 %s 的记录数: %v\n", table, err)
		} else {
			fmt.Printf("总记录数: %d\n", count)
		}

		// 显示前几条记录（如果有数据的话）
		if count > 0 {
			var sampleData []map[string]interface{}
			err = db.Table(table).Limit(2).Find(&sampleData).Error
			if err != nil {
				fmt.Printf("无法获取表 %s 的示例数据: %v\n", table, err)
			} else if len(sampleData) > 0 {
				fmt.Println("示例数据:")
				for i, record := range sampleData {
					fmt.Printf("  记录 %d: ", i+1)
					jsonBytes, _ := json.Marshal(record)
					fmt.Printf("%s\n", jsonBytes)
				}
			}
		}
	}

	fmt.Println("\n=== 探索完成 ===")
}
