package main

import (
	"fmt"
	"log"
	"nofx/crypto"
	"nofx/store"

	"github.com/google/uuid"
)

func main() {
	// Initialize database
	st, err := store.NewWithConfig(store.DBConfig{
		Type: store.DBTypeSQLite,
		Path: "../data/data.db",
	})
	if err != nil {
		log.Fatal("❌ Failed to initialize database:", err)
	}
	defer st.Close()

	// Check if any exchanges already exist
	var exchangeCount int64
	err = st.GormDB().Table("exchanges").Count(&exchangeCount).Error
	if err != nil {
		log.Fatal("❌ Error counting exchanges:", err)
	}

	if exchangeCount > 0 {
		fmt.Printf("ℹ️  Database already has %d exchange(s), skipping initialization\n", exchangeCount)
		return
	}

	// Initialize encryption service for encrypted fields
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		log.Fatal("❌ Failed to initialize encryption service:", err)
	}
	crypto.SetGlobalCryptoService(cryptoService)

	// Create a default exchange
	exchangeID := uuid.New().String()
	name, typ := getExchangeNameAndType("binance")

	exchange := &store.Exchange{
		ID:           exchangeID,
		ExchangeType: "binance",
		AccountName:  "Default Binance Account",
		UserID:       "default",
		Name:         name,
		Type:         typ,
		Enabled:      true,
		APIKey:       crypto.EncryptedString(""),
		SecretKey:    crypto.EncryptedString(""),
	}

	if err := st.GormDB().Create(exchange).Error; err != nil {
		log.Fatal("❌ Failed to create default exchange:", err)
	}

	fmt.Printf("✅ Created default exchange: %s (%s)\n", exchange.Name, exchange.ExchangeType)
	fmt.Printf("📊 Total exchanges in database: 1\n")
}

func getExchangeNameAndType(exchangeType string) (string, string) {
	switch exchangeType {
	case "binance":
		return "Binance", "cex"
	case "bybit":
		return "Bybit", "cex"
	case "okx":
		return "OKX", "cex"
	case "bitget":
		return "Bitget", "cex"
	case "hyperliquid":
		return "Hyperliquid", "dex"
	case "aster":
		return "Aster", "dex"
	case "lighter":
		return "Lighter", "dex"
	default:
		return exchangeType, "cex"
	}
}
