package main

import (
	"log"
	"nofx/store"
)

func main() {
	log.Println("🔍 Checking strategy configurations for K-line settings...")

	// Initialize the store
	s, err := store.New("data/data.db")
	if err != nil {
		log.Fatalf("❌ Failed to initialize store: %v", err)
	}
	defer s.Close()

	// Get all strategies
	allStrategies, err := s.Strategy().List("default")
	if err != nil {
		log.Printf("⚠️ Could not load default strategies: %v", err)
		allStrategies = []*store.Strategy{}
	}

	// Get all user strategies (try to load strategies for various user IDs)
	userIDs := []string{"default", "admin", ""}
	for _, userID := range userIDs {
		userStrategies, err := s.Strategy().List(userID)
		if err == nil && len(userStrategies) > 0 {
			allStrategies = append(allStrategies, userStrategies...)
		}
	}

	log.Printf("📋 Found %d strategies to check", len(allStrategies))

	updatedCount := 0
	for _, strategy := range allStrategies {
		needsUpdate := false

		// Parse the config
		config, err := strategy.ParseConfig()
		if err != nil {
			log.Printf("⚠️ Could not parse config for strategy %s: %v", strategy.ID, err)
			continue
		}

		// Check if K-line config has the expected fields
		klineCfg := config.Indicators.Klines

		// Check for missing K-line configuration fields
		if klineCfg.PrimaryTimeframe == "" {
			klineCfg.PrimaryTimeframe = "5m"
			needsUpdate = true
			log.Printf("🔧 Setting default PrimaryTimeframe for strategy %s", strategy.ID)
		}
		if klineCfg.PrimaryCount == 0 {
			klineCfg.PrimaryCount = 30
			needsUpdate = true
			log.Printf("🔧 Setting default PrimaryCount for strategy %s", strategy.ID)
		}
		if klineCfg.LongerTimeframe == "" {
			klineCfg.LongerTimeframe = "4h"
			needsUpdate = true
			log.Printf("🔧 Setting default LongerTimeframe for strategy %s", strategy.ID)
		}
		if klineCfg.LongerCount == 0 {
			klineCfg.LongerCount = 10
			needsUpdate = true
			log.Printf("🔧 Setting default LongerCount for strategy %s", strategy.ID)
		}
		if klineCfg.SelectedTimeframes == nil || len(klineCfg.SelectedTimeframes) == 0 {
			klineCfg.SelectedTimeframes = []string{"5m", "15m", "1h", "4h"}
			needsUpdate = true
			log.Printf("🔧 Setting default SelectedTimeframes for strategy %s", strategy.ID)
		}
		if klineCfg.TimeframeCounts == nil {
			klineCfg.TimeframeCounts = map[string]int{
				"1m":  120,
				"3m":  120,
				"5m":  120,
				"15m": 80,
				"30m": 60,
				"1h":  50,
				"2h":  40,
				"4h":  30,
				"1d":  20,
				"1w":  10,
			}
			needsUpdate = true
			log.Printf("🔧 Setting default TimeframeCounts for strategy %s", strategy.ID)
		}
		if klineCfg.TradingStylePreset == "" {
			klineCfg.TradingStylePreset = "short"
			needsUpdate = true
			log.Printf("🔧 Setting default TradingStylePreset for strategy %s", strategy.ID)
		}

		// Update the config if needed
		if needsUpdate {
			config.Indicators.Klines = klineCfg
			if err := strategy.SetConfig(config); err != nil {
				log.Printf("❌ Could not update config for strategy %s: %v", strategy.ID, err)
				continue
			}

			// Save the updated strategy
			if err := s.Strategy().Update(strategy); err != nil {
				log.Printf("❌ Could not save updated strategy %s: %v", strategy.ID, err)
				continue
			}

			log.Printf("✅ Updated strategy configuration for %s", strategy.ID)
			updatedCount++
		}
	}

	log.Printf("✅ Completed strategy configuration check. Updated %d strategies.", updatedCount)

	// Also ensure the default strategy exists with proper configuration
	defaultStrategy, err := s.Strategy().GetDefault()
	if err != nil {
		log.Println("⚠️ No default strategy found, creating one...")
		defaultConfig := store.GetDefaultStrategyConfig("en")

		defaultStrategy = &store.Strategy{
			ID:            "default",
			UserID:        "default",
			Name:          "Default Strategy",
			Description:   "Default strategy configuration",
			IsDefault:     true,
			IsActive:      false,
			IsPublic:      false,
			ConfigVisible: true,
		}

		if err := defaultStrategy.SetConfig(&defaultConfig); err != nil {
			log.Printf("❌ Could not set config for default strategy: %v", err)
		} else {
			if err := s.Strategy().Create(defaultStrategy); err != nil {
				log.Printf("❌ Could not create default strategy: %v", err)
			} else {
				log.Println("✅ Created default strategy with proper configuration")
			}
		}
	} else {
		log.Println("📋 Default strategy already exists")
	}
}
