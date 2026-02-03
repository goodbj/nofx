package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func checkDatabase() {
	// 连接到数据库
	db, err := sql.Open("sqlite3", "./nofx.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("=== 检查交易员表结构 ===")

	// 查看表结构
	rows, err := db.Query("PRAGMA table_info(traders)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("traders表字段:")
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt_value sql.NullString
		err := rows.Scan(&cid, &name, &typ, &notnull, &dflt_value, &pk)
		if err != nil {
			log.Fatal(err)
		}
		defaultVal := ""
		if dflt_value.Valid {
			defaultVal = dflt_value.String
		}
		fmt.Printf("  %d. %s (%s) NOT NULL:%d PK:%d DEFAULT:%s\n", cid, name, typ, notnull, pk, defaultVal)
	}

	fmt.Println("\n=== 检查交易员数据 ===")

	// 查询所有交易员的数据
	traderRows, err := db.Query(`
		SELECT id, name, data_access_method, exchange_id, ai_model_id, is_running
		FROM traders 
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer traderRows.Close()

	count := 0
	for traderRows.Next() {
		var id, name, dataAccessMethod, exchangeID, aiModelID string
		var isRunning bool
		err := traderRows.Scan(&id, &name, &dataAccessMethod, &exchangeID, &aiModelID, &isRunning)
		if err != nil {
			log.Fatal(err)
		}

		count++
		fmt.Printf("\n%d. 交易员信息:\n", count)
		fmt.Printf("   ID: %s\n", id)
		fmt.Printf("   名称: %s\n", name)
		fmt.Printf("   数据访问方式: %s\n", dataAccessMethod)
		fmt.Printf("   交易所ID: %s\n", exchangeID)
		fmt.Printf("   AI模型ID: %s\n", aiModelID)
		fmt.Printf("   是否运行中: %t\n", isRunning)
	}

	if count == 0 {
		fmt.Println("没有找到任何交易员数据")
	}

	fmt.Printf("\n总共找到 %d 个交易员记录\n", count)
}

func main() {
	checkDatabase()
}
