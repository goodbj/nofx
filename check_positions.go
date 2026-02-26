package main

import (
	"fmt"
	"nofx/store"
)

func main() {
	st, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "./data/data.db",
	})
	if err != nil {
		fmt.Printf("Error creating store: %v\n", err)
		return
	}
	defer st.Close()

	// 获取所有开放持仓
	positions, err := st.Position().GetOpenPositions("")
	if err != nil {
		fmt.Printf("Error getting open positions: %v\n", err)
		return
	}

	fmt.Printf("Found %d OPEN positions in database\n", len(positions))

	for _, pos := range positions {
		traderIDPrefix := pos.TraderID
		if len(traderIDPrefix) > 8 {
			traderIDPrefix = traderIDPrefix[:8]
		}
		fmt.Printf("- ID: %d, Trader: %s, Symbol: %s, Side: %s, EntryTime: %d\n",
			pos.ID, traderIDPrefix, pos.Symbol, pos.Side, pos.EntryTime)
	}
}
