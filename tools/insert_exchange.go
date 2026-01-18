package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

func main() {
	// Connect to the SQLite database
	db, err := sql.Open("sqlite", "../data/data.db")
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	// Check if exchanges already exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM exchanges;").Scan(&count)
	if err != nil {
		log.Fatal("❌ Error checking exchanges:", err)
	}

	if count > 0 {
		fmt.Printf("ℹ️  Database already has %d exchange(s), skipping initialization\n", count)
		return
	}

	// Insert a default exchange
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`
		INSERT INTO exchanges 
		(id, exchange_type, account_name, user_id, name, type, enabled, api_key, secret_key, passphrase, testnet, custom_api_url, hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key, lighter_wallet_addr, lighter_private_key, lighter_api_key_private_key, lighter_api_key_index, created_at, updated_at) 
		VALUES 
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"00000000-0000-0000-0000-000000000001", // id
		"binance",                              // exchange_type
		"Default Binance Account",              // account_name
		"default",                              // user_id
		"Binance",                              // name
		"cex",                                  // type
		1,                                      // enabled (true)
		"",                                     // api_key
		"",                                     // secret_key
		"",                                     // passphrase
		0,                                      // testnet (false)
		"",                                     // custom_api_url
		"",                                     // hyperliquid_wallet_addr
		"",                                     // aster_user
		"",                                     // aster_signer
		"",                                     // aster_private_key
		"",                                     // lighter_wallet_addr
		"",                                     // lighter_private_key
		"",                                     // lighter_api_key_private_key
		0,                                      // lighter_api_key_index
		now,                                    // created_at
		now,                                    // updated_at
	)

	if err != nil {
		log.Fatal("❌ Error inserting default exchange:", err)
	}

	fmt.Println("✅ Successfully inserted default exchange into database")

	// Verify the insertion
	var newCount int
	err = db.QueryRow("SELECT COUNT(*) FROM exchanges;").Scan(&newCount)
	if err != nil {
		log.Fatal("❌ Error checking exchanges after insert:", err)
	}

	fmt.Printf("📊 Total exchanges in database: %d\n", newCount)
}
