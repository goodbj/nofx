package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("=== AI Models Configuration ===")
	fmt.Println()

	// Query AI models
	rows, err := db.Query(`
		SELECT id, name, provider, enabled, custom_api_url, custom_model_name 
		FROM ai_models 
		WHERE user_id='default' 
		ORDER BY id
	`)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var id, name, provider, customAPIURL, customModelName string
		var enabled bool

		err := rows.Scan(&id, &name, &provider, &enabled, &customAPIURL, &customModelName)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}

		found = true
		fmt.Printf("Model ID: %s\n", id)
		fmt.Printf("  Name: %s\n", name)
		fmt.Printf("  Provider: %s\n", provider)
		fmt.Printf("  Enabled: %v\n", enabled)
		fmt.Printf("  Custom API URL: %q\n", customAPIURL)
		fmt.Printf("  Custom Model Name: %q\n", customModelName)
		fmt.Println()

		if enabled && customAPIURL == "" {
			fmt.Printf("⚠️  WARNING: Enabled model '%s' has empty Custom API URL!\n", name)
			fmt.Println()
		}
	}

	if !found {
		fmt.Println("❌ No AI models found for user 'default'")
		fmt.Println("   Please configure AI model in the web interface first.")
	}
}
