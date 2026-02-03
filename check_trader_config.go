package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 连接到数据库
	db, err := sql.Open("sqlite3", "./nofx.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询所有交易员的data_access_method字段
	rows, err := db.Query(`
		SELECT id, name, data_access_method 
		FROM traders 
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("交易员配置检查:")
	fmt.Println("================")

	for rows.Next() {
		var id, name, dataAccessMethod string
		err := rows.Scan(&id, &name, &dataAccessMethod)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("ID: %s\n", id)
		fmt.Printf("名称: %s\n", name)
		fmt.Printf("数据访问方式: %s\n", dataAccessMethod)
		fmt.Println("----------------")
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
}
