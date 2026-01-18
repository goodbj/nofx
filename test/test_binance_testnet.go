package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	// Get API credentials from environment variables
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	apiURL := os.Getenv("BINANCE_API_URL")

	if apiKey == "" || secretKey == "" || apiURL == "" {
		log.Fatal("Missing required environment variables: BINANCE_API_KEY, BINANCE_SECRET_KEY, BINANCE_API_URL")
	}

	// Create futures client
	client := futures.NewClient(apiKey, secretKey)
	client.BaseURL = apiURL

	// Test connection by fetching server time
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Connected successfully! Server time: %d\n", serverTime)

	// Test getting account balance
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Printf("Failed to get account: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Account balances: %+v\n", account)
	fmt.Println("Test completed successfully!")
}
