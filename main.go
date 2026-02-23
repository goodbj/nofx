package main

import (
	"errors"
	"fmt"
	"nofx/api"
	"nofx/auth"
	"nofx/background"
	"nofx/backtest"
	"nofx/config"
	"nofx/crypto"
	"nofx/experience"
	"nofx/logger"
	"nofx/manager"
	"nofx/mcp"
	"nofx/store"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// normalizeUserIDMain copies the logic from api/backtest.go
func normalizeUserIDMain(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return "default"
	}
	return id
}

func main() {
	// Load .env environment variables
	_ = godotenv.Load()

	// Initialize logger with level from environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // default log level
	}
	logger.InitWithSimpleConfig(logLevel)

	logger.Infof("🔧 Configured log level: %s", logLevel)

	logger.Info("╔════════════════════════════════════════════════════════════╗")
	logger.Info("║           🚀 NOFX - AI-Powered Trading System              ║")
	logger.Info("╚════════════════════════════════════════════════════════════╝")

	// Initialize global configuration (loaded from .env)
	config.Init()
	cfg := config.Get()
	logger.Info("✅ Configuration loaded")

	// Initialize encryption service BEFORE database (so EncryptedString can decrypt on read)
	logger.Info("🔐 Initializing encryption service...")
	cryptoService, err := crypto.NewCryptoService()
	if err != nil {
		logger.Fatalf("❌ Failed to initialize encryption service: %v", err)
	}
	crypto.SetGlobalCryptoService(cryptoService)
	logger.Info("✅ Encryption service initialized successfully")

	// Initialize database from configuration
	// For backward compatibility: command line arg overrides config (SQLite only)
	if len(os.Args) > 1 {
		cfg.DBPath = os.Args[1]
	}
	// Ensure data directory exists (for SQLite)
	if cfg.DBType == "sqlite" {
		if dir := filepath.Dir(cfg.DBPath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				logger.Errorf("Failed to create data directory: %v", err)
			}
		}
	}

	logger.Infof("📋 Initializing database (%s)...", cfg.DBType)
	dbType := store.DBTypeSQLite
	if cfg.DBType == "postgres" {
		dbType = store.DBTypePostgres
	}
	st, err := store.NewWithConfig(store.DBConfig{
		Type:     dbType,
		Path:     cfg.DBPath,
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		logger.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer st.Close()
	backtest.UseDatabaseWithType(st.DB(), st.DBType() == store.DBTypePostgres)

	// Initialize installation ID for experience improvement (anonymous statistics)
	initInstallationID(st)

	// Set JWT secret
	auth.SetJWTSecret(cfg.JWTSecret)
	logger.Info("🔑 JWT secret configured")

	// WebSocket market monitor is NO LONGER USED
	// All K-line data now comes from CoinAnk API instead of Binance WebSocket cache
	// Commented out to reduce unnecessary connections:
	// go market.NewWSMonitor(150).Start(nil)
	// logger.Info("📊 WebSocket market monitor started")
	// time.Sleep(500 * time.Millisecond)
	logger.Info("📊 Using CoinAnk API for all market data (WebSocket cache disabled)")

	// Create TraderManager and BacktestManager
	traderManager := manager.NewTraderManager()
	mcpClient := newSharedMCPClient()
	backtestManager := backtest.NewManager(mcpClient)

	// Set the AI resolver for backtest manager to handle AI configuration
	backtestManager.SetAIResolver(func(cfg *backtest.BacktestConfig) error {
		if cfg == nil {
			return errors.New("config is nil")
		}
		if st == nil {
			return errors.New("System database not ready, cannot load AI model configuration")
		}

		cfg.UserID = normalizeUserIDMain(cfg.UserID)
		modelID := strings.TrimSpace(cfg.AIModelID)

		var (
			model *store.AIModel
			err   error
		)

		if modelID != "" {
			model, err = st.AIModel().Get(cfg.UserID, modelID)
			if err != nil {
				return fmt.Errorf("Failed to load AI model: %w", err)
			}
		} else {
			model, err = st.AIModel().GetDefault(cfg.UserID)
			if err != nil {
				return fmt.Errorf("No available AI model found: %w", err)
			}
			cfg.AIModelID = model.ID
		}

		if !model.Enabled {
			return fmt.Errorf("AI model %s is not enabled yet", model.Name)
		}

		apiKey := strings.TrimSpace(string(model.APIKey))
		if apiKey == "" {
			return fmt.Errorf("AI model %s is missing API Key, please configure it in the system first", model.Name)
		}

		provider := strings.ToLower(strings.TrimSpace(model.Provider))
		// Ensure provider is never empty or "inherit" - infer from model name if needed
		if provider == "" || provider == "inherit" {
			modelNameLower := strings.ToLower(model.Name)
			if strings.Contains(modelNameLower, "claude") || strings.Contains(modelNameLower, "anthropic") {
				provider = "anthropic"
			} else if strings.Contains(modelNameLower, "gpt") || strings.Contains(modelNameLower, "openai") {
				provider = "openai"
			} else if strings.Contains(modelNameLower, "gemini") || strings.Contains(modelNameLower, "google") {
				provider = "google"
			} else if strings.Contains(modelNameLower, "deepseek") {
				provider = "deepseek"
			} else if model.CustomAPIURL != "" {
				provider = "custom"
			} else {
				provider = "openai" // default fallback
			}
			logger.Infof("📊 Inferred AI provider '%s' from model name '%s'", provider, model.Name)
		}
		cfg.AICfg.Provider = provider
		cfg.AICfg.APIKey = apiKey
		cfg.AICfg.BaseURL = strings.TrimSpace(model.CustomAPIURL)
		modelName := strings.TrimSpace(model.CustomModelName)
		if cfg.AICfg.Model == "" {
			cfg.AICfg.Model = modelName
		}
		cfg.AICfg.Model = strings.TrimSpace(cfg.AICfg.Model)

		if cfg.AICfg.Provider == "custom" {
			if cfg.AICfg.BaseURL == "" {
				return errors.New("Custom AI model requires API URL configuration")
			}
			if cfg.AICfg.Model == "" {
				return errors.New("Custom AI model requires model name configuration")
			}
		}

		return nil
	})

	if err := backtestManager.RestoreRuns(); err != nil {
		logger.Warnf("⚠️ Failed to restore backtest history: %v", err)
	}

	// Verify and ensure all environment variables are properly set before loading traders
	logger.Info("🔧 Initializing environment variables...")

	// Proxy settings verification
	useProxy := os.Getenv("USE_BINANCE_PROXY")
	proxyURL := os.Getenv("BINANCE_PROXY_URL")

	if useProxy == "true" {
		if proxyURL == "" {
			// Set default proxy URL if not specified
			defaultProxyURL := "http://localhost:8081"
			os.Setenv("BINANCE_PROXY_URL", defaultProxyURL)
			proxyURL = defaultProxyURL
			logger.Infof("🔧 Proxy enabled, default proxy URL set to %s", defaultProxyURL)
		}
		logger.Infof("🔧 Proxy configuration: USE_BINANCE_PROXY=%s, BINANCE_PROXY_URL=%s", useProxy, proxyURL)
	} else {
		logger.Info("🔧 Proxy disabled or not configured")
	}

	// Ensure other critical environment variables are set
	criticalEnvVars := []string{
		"JWT_SECRET",
		"DATA_ENCRYPTION_KEY",
		"RSA_PRIVATE_KEY",
	}

	for _, envVar := range criticalEnvVars {
		if value := os.Getenv(envVar); value == "" {
			logger.Warnf("⚠️ Environment variable %s is not set", envVar)
		}
	}

	// Small delay to ensure environment variables are fully propagated
	time.Sleep(100 * time.Millisecond)
	logger.Info("🔧 Environment variables initialization completed")

	// Load all traders from database to memory (may auto-start traders with IsRunning=true)
	logger.Info("🔄 Loading traders from database...")
	if err := traderManager.LoadTradersFromStore(st); err != nil {
		logger.Fatalf("❌ Failed to load traders: %v", err)
	}
	logger.Info("✅ Traders loaded successfully")

	// Display loaded trader information
	traders, err := st.Trader().List("default")
	if err != nil {
		logger.Fatalf("❌ Failed to get trader list: %v", err)
	}

	logger.Info("🤖 AI Trader Configurations in Database:")
	if len(traders) == 0 {
		logger.Info("  (No trader configurations, please create via Web interface)")
	} else {
		for _, t := range traders {
			status := "❌ Stopped"
			if t.IsRunning {
				status = "✅ Running"
			}
			logger.Infof("  • %s [%s] %s - AI Model: %s, Exchange: %s",
				t.Name, t.ID[:8], status, t.AIModelID, t.ExchangeID)
		}
	}

	// Start position sync checker service
	logger.Info("🔄 启动持仓状态一致性检查服务...")
	go func() {
		positionSyncChecker := background.NewPositionSyncChecker(st, traderManager, time.Hour, logger.Log)
		go positionSyncChecker.Start()
		logger.Info("✅ 持仓状态检查服务已启动，将每小时自动检查一次")
	}()

	// Start API server
	server := api.NewServer(traderManager, st, cryptoService, backtestManager, cfg.APIServerPort)
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatalf("❌ Failed to start API server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("✅ System started successfully, waiting for trading commands...")
	logger.Info("📌 Tip: Use Ctrl+C to stop the system")

	<-quit
	logger.Info("📴 Shutdown signal received, closing system...")

	// Stop all traders
	traderManager.StopAll()
	logger.Info("✅ System shut down safely")
}

// newSharedMCPClient creates a shared MCP AI client (for backtesting)
func newSharedMCPClient() mcp.AIClient {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		logger.Warn("⚠️ DEEPSEEK_API_KEY not set, AI features will be unavailable")
		return nil
	}
	return mcp.NewDeepSeekClient()
}

// initInstallationID initializes the anonymous installation ID for experience improvement
// This ID is persisted in database and used for anonymous usage statistics
func initInstallationID(st *store.Store) {
	const key = "installation_id"

	// Try to load from database
	installationID, err := st.GetSystemConfig(key)
	if err != nil {
		logger.Warnf("⚠️ Failed to load installation ID: %v", err)
	}

	// Generate new ID if not exists
	if installationID == "" {
		installationID = uuid.New().String()
		if err := st.SetSystemConfig(key, installationID); err != nil {
			logger.Warnf("⚠️ Failed to save installation ID: %v", err)
		}
		logger.Infof("📊 Generated new installation ID: %s", installationID[:8]+"...")
	}

	// Set installation ID in experience module
	experience.SetInstallationID(installationID)
}
