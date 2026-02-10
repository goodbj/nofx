package main

import (
	"fmt"
	"nofx/store"
)

func main() {
	// 创建数据库连接
	db, err := store.New("data/data.db")
	if err != nil {
		fmt.Printf("无法连接数据库: %v\n", err)
		return
	}
	defer db.Close()

	// 尝试查询交易所配置
	var exchanges []*store.Exchange
	err = db.GormDB().Find(&exchanges).Error
	if err != nil {
		fmt.Printf("查询交易所配置失败: %v\n", err)
		return
	}

	fmt.Printf("找到 %d 个交易所配置\n", len(exchanges))
	for i, exchange := range exchanges {
		fmt.Printf("[%d] ID: %s, Type: %s, Testnet: %t, CustomAPIURL: %s\n",
			i+1, exchange.ID, exchange.ExchangeType, exchange.Testnet, exchange.CustomAPIURL)
	}
}
