package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	// Connect to the SQLite database
	db, err := sql.Open("sqlite", "../data/data.db")
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Check if database is accessible
	fmt.Println("✅ Successfully connected to database")

	// Check all tables
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		log.Fatal("❌ Error querying tables:", err)
	}
	defer rows.Close()

	fmt.Println("\n📋 Tables in database:")
	tables := make([]string, 0)
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatal("❌ Error scanning table name:", err)
		}
		fmt.Printf("  - %s\n", tableName)
		tables = append(tables, tableName)
	}

	// Check specific tables
	checkTable(db, "users")
	checkTable(db, "exchanges")
	checkTable(db, "traders")

	// Check exchange data specifically
	fmt.Println("\n💱 Exchange data:")
	exchangeRows, err := db.Query("SELECT * FROM exchanges LIMIT 10;")
	if err != nil {
		fmt.Printf("  No exchanges table or error querying: %v\n", err)
	} else {
		defer exchangeRows.Close()

		// Get column names
		columns, err := exchangeRows.Columns()
		if err != nil {
			fmt.Printf("  Error getting columns: %v\n", err)
		} else {
			for exchangeRows.Next() {
				// Create a slice of interface{}'s to represent each column
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				// Scan the result into the column pointers
				if err := exchangeRows.Scan(valuePtrs...); err != nil {
					fmt.Printf("  Error scanning row: %v\n", err)
					continue
				}

				// Create our map and retrieve the value for each column from the pointers slice
				entry := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					b, ok := val.([]byte)
					if ok {
						// Attempt to unmarshal as JSON if it looks like JSON
						var jsonVal interface{}
						if err := json.Unmarshal(b, &jsonVal); err == nil {
							entry[col] = jsonVal
						} else {
							entry[col] = string(b)
						}
					} else {
						entry[col] = val
					}
				}
				fmt.Printf("  %+v\n", entry)
			}
		}
	}
}

func checkTable(db *sql.DB, tableName string) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName)
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		fmt.Printf("❌ Error querying %s table: %v\n", tableName, err)
	} else {
		fmt.Printf("%s Count: %d\n", getEmojiForTable(tableName), count)
	}
}

func getEmojiForTable(tableName string) string {
	switch tableName {
	case "users":
		return "👥 Users"
	case "exchanges":
		return "💱 Exchanges"
	case "traders":
		return "🤖 Traders"
	default:
		return fmt.Sprintf("📁 %s", tableName)
	}
}
