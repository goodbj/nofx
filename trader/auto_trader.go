package trader

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"nofx/experience"
	"nofx/kernel"
	"nofx/logger"
	"nofx/market"
	"nofx/mcp"
	"nofx/store"
	"os"
	"strings"
	"sync"
	"time"
)

// AutoTraderConfig auto trading configuration (simplified version - AI makes all decisions)
type AutoTraderConfig struct {
	// Trader identification
	ID      string // Trader unique identifier (for log directory, etc.)
	Name    string // Trader display name
	AIModel string // AI model: "qwen" or "deepseek"

	// Trading platform selection
	Exchange   string // Exchange type: "binance", "bybit", "okx", "bitget", "hyperliquid", "aster" or "lighter"
	ExchangeID string // Exchange account UUID (for multi-account support)

	// Binance API configuration
	BinanceAPIKey    string
	BinanceSecretKey string

	// Binance custom endpoint (for testnet or custom API endpoints)
	BinanceCustomAPIURL string // Custom API endpoint for Binance (e.g., testnet)

	// Bybit API configuration
	BybitAPIKey    string
	BybitSecretKey string

	// OKX API configuration
	OKXAPIKey     string
	OKXSecretKey  string
	OKXPassphrase string

	// Bitget API configuration
	BitgetAPIKey     string
	BitgetSecretKey  string
	BitgetPassphrase string

	// Hyperliquid configuration
	HyperliquidPrivateKey string
	HyperliquidWalletAddr string
	HyperliquidTestnet    bool

	// Aster configuration
	AsterUser       string // Aster main wallet address
	AsterSigner     string // Aster API wallet address
	AsterPrivateKey string // Aster API wallet private key

	// LIGHTER configuration
	LighterWalletAddr       string // LIGHTER wallet address (L1 wallet)
	LighterPrivateKey       string // LIGHTER L1 private key (for account identification)
	LighterAPIKeyPrivateKey string // LIGHTER API Key private key (40 bytes, for transaction signing)
	LighterAPIKeyIndex      int    // LIGHTER API Key index (0-255)
	LighterTestnet          bool   // Whether to use testnet

	// AI configuration
	UseQwen     bool
	DeepSeekKey string
	QwenKey     string

	// Custom AI API configuration
	CustomAPIURL    string
	CustomAPIKey    string
	CustomModelName string

	// Scan configuration
	ScanInterval time.Duration // Scan interval (recommended 3 minutes)

	// Account configuration
	InitialBalance float64 // Initial balance (for P&L calculation, must be set manually)

	// Risk control (only as hints, AI can make autonomous decisions)
	MaxDailyLoss    float64       // Maximum daily loss percentage (hint)
	MaxDrawdown     float64       // Maximum drawdown percentage (hint)
	StopTradingTime time.Duration // Pause duration after risk control triggers

	// Position mode
	IsCrossMargin bool // true=cross margin mode, false=isolated margin mode

	// Competition visibility
	ShowInCompetition bool // Whether to show in competition page

	// Trader delegates AI model implementation to the AI model itself
	// Trader should not care about whether AI uses API calls or web automation

	// Strategy configuration (use complete strategy config)
	StrategyConfig *store.StrategyConfig // Strategy configuration (includes coin sources, indicators, risk control, prompts, etc.)
}

// TraderTradeRecord represents a single trade record for tracking purposes
type TraderTradeRecord struct {
	Timestamp time.Time
	Symbol    string
	Action    string
	PnL       float64
}

// TradeFrequencyTracker tracks trade counts for enforcing frequency limits
type TradeFrequencyTracker struct {
	dailyTrades           int
	hourlyTrades          int
	symbolHourlyTrades    map[string]int // Count per symbol per hour
	dailyResetTime        time.Time
	hourlyResetTime       time.Time
	symbolHourlyResetTime time.Time
	mutex                 sync.RWMutex
}

// AutoTrader automatic trader
type AutoTrader struct {
	id                    string // Trader unique identifier
	name                  string // Trader display name
	aiModel               string // AI model name
	exchange              string // Trading platform type (binance/bybit/etc)
	exchangeID            string // Exchange account UUID
	showInCompetition     bool   // Whether to show in competition page
	config                AutoTraderConfig
	trader                Trader // Use Trader interface (supports multiple platforms)
	mcpClient             mcp.AIClient
	store                 *store.Store           // Data storage (decision records, etc.)
	strategyEngine        *kernel.StrategyEngine // Strategy engine (uses strategy configuration)
	cycleNumber           int                    // Current cycle number
	initialBalance        float64
	dailyPnL              float64
	customPrompt          string // Custom trading strategy prompt
	overrideBasePrompt    bool   // Whether to override base prompt
	lastResetTime         time.Time
	stopUntil             time.Time
	isRunning             bool
	isRunningMutex        sync.RWMutex           // Mutex to protect isRunning flag
	executionMutex        sync.Mutex             // Mutex to prevent concurrent executions (for manual scans)
	isExecuting           bool                   // Flag to indicate if a decision cycle is currently executing
	startTime             time.Time              // System start time
	callCount             int                    // AI call count
	positionFirstSeenTime map[string]int64       // Position first seen time (symbol_side -> timestamp in milliseconds)
	stopMonitorCh         chan struct{}          // Used to stop monitoring goroutine
	monitorWg             sync.WaitGroup         // Used to wait for monitoring goroutine to finish
	peakPnLCache          map[string]float64     // Peak profit cache (symbol -> peak P&L percentage)
	peakPnLCacheMutex     sync.RWMutex           // Cache read-write lock
	lastBalanceSyncTime   time.Time              // Last balance sync time
	userID                string                 // User ID
	tradeFrequencyTracker *TradeFrequencyTracker // Tracks trade frequencies for limits
	lastManualScanTime    time.Time              // 🔥 新增：上次手动扫描时间（用于延迟系统扫描）
	nextSystemScanTime    time.Time              // 🔥 新增：下次系统扫描时间
	scanDelayDuration     time.Duration          // 🔥 新增：手动扫描后延迟系统扫描的时长

}

// calculateScanDelay 智能计算手动扫描后的延迟时长
// 规则：
//   - 扫描周期 >= 8分钟：延迟5分钟（保留手动扫描的价值）
//   - 扫描周期 < 8分钟：延迟 = 周期 × 0.8（避免手动扫描意义不大）
func calculateScanDelay(scanInterval time.Duration) time.Duration {
	const minDelayMinutes = 5.0  // 最小延迟5分钟
	const thresholdMinutes = 8.0 // 阈值：8分钟

	intervalMinutes := scanInterval.Minutes()

	// 如果扫描周期 >= 8分钟，使用固定5分钟延迟（手动扫描有意义）
	if intervalMinutes >= thresholdMinutes {
		return time.Duration(minDelayMinutes) * time.Minute
	}

	// 如果扫描周期 < 8分钟，延迟 = 周期 × 0.8（避免过于频繁）
	delayMinutes := intervalMinutes * 0.8

	// 确保至少延迟2分钟（避免过短）
	if delayMinutes < 2.0 {
		delayMinutes = 2.0
	}

	return time.Duration(delayMinutes * float64(time.Minute))
}

// NewAutoTrader creates an automatic trader
// st parameter is used to store decision records to database
func NewAutoTrader(config AutoTraderConfig, st *store.Store, userID string) (*AutoTrader, error) {
	// Set default values
	if config.ID == "" {
		config.ID = "default_trader"
	}
	if config.Name == "" {
		config.Name = "Default Trader"
	}
	if config.AIModel == "" {
		if config.UseQwen {
			config.AIModel = "qwen"
		} else {
			config.AIModel = "deepseek"
		}
	}

	// Initialize AI client based on provider
	var mcpClient mcp.AIClient
	aiModel := config.AIModel
	if config.UseQwen && aiModel == "" {
		aiModel = "qwen"
	}

	switch aiModel {
	case "claude":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewClaudeClient()
		mcpClient = mcp.NewClaudeClient()
		mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using Claude AI (delegating implementation to AI model)", config.Name)

	case "kimi":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewKimiClient()
		mcpClient = mcp.NewKimiClient()
		mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using Kimi AI (delegating implementation to AI model)", config.Name)

	case "gemini":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewGeminiClient()
		mcpClient = mcp.NewGeminiClient()
		mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using Google Gemini AI (delegating implementation to AI model)", config.Name)

	case "grok":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewGrokClient()
		mcpClient = mcp.NewGrokClient()
		mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using xAI Grok AI (delegating implementation to AI model)", config.Name)

	case "openai":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewOpenAIClient()
		mcpClient = mcp.NewOpenAIClient()
		mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using OpenAI (delegating implementation to AI model)", config.Name)

	case "qwen":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewQwenClient()
		// Check if bypass is enabled via custom configuration
		bypassEnabled := config.CustomModelName == "bypass" || strings.Contains(strings.ToLower(config.CustomAPIURL), "bypass")
		if bypassEnabled {
			// Even when using browser automation, we need to set API key for base client validation
			mcpClient.SetAPIKey("dummy-key-for-browser-automation", config.CustomAPIURL, config.CustomModelName)
			// For bypass clients, we need to ensure trader ID is passed through
			if guardianClient, ok := mcpClient.(*mcp.GuardianClient); ok {
				mcpClient = mcp.NewBypassClientWithTraderID(mcpClient, true, "guardian-ai", guardianClient.GetTraderID())
			} else {
				mcpClient = mcp.NewBypassClient(mcpClient, true, "guardian-ai")
			}
			logger.Infof("🤖 [%%s] Using QWEN with Browser Automation Bypass", config.Name)
		} else {
			apiKey := config.QwenKey
			if apiKey == "" {
				apiKey = config.CustomAPIKey
			}
			mcpClient.SetAPIKey(apiKey, config.CustomAPIURL, config.CustomModelName)
			logger.Infof("🤖 [%%s] Using Alibaba Cloud Qwen AI", config.Name)
		}

	case "custom":
		// Check if bypass is enabled via custom configuration
		bypassEnabled := config.CustomModelName == "bypass" || strings.Contains(strings.ToLower(config.CustomAPIURL), "bypass")
		mcpClient = mcp.New()
		if bypassEnabled {
			// Even when using browser automation, we need to set API key for base client validation
			mcpClient.SetAPIKey("dummy-key-for-browser-automation", config.CustomAPIURL, config.CustomModelName)
			// For custom models, we'll use guardian-ai as the default browser provider
			// For bypass clients, we need to ensure trader ID is passed through
			if guardianClient, ok := mcpClient.(*mcp.GuardianClient); ok {
				mcpClient = mcp.NewBypassClientWithTraderID(mcpClient, true, "guardian-ai", guardianClient.GetTraderID())
			} else {
				mcpClient = mcp.NewBypassClient(mcpClient, true, "guardian-ai")
			}
			logger.Infof("🤖 [%s] Using CUSTOM with Browser Automation Bypass", config.Name)
		} else {
			mcpClient.SetAPIKey(config.CustomAPIKey, config.CustomAPIURL, config.CustomModelName)
			logger.Infof("🤖 [%s] Using custom AI API: %s (model: %s)", config.Name, config.CustomAPIURL, config.CustomModelName)
		}

	case "ollama":
		// Check if bypass is enabled via custom configuration
		bypassEnabled := config.CustomModelName == "bypass" || strings.Contains(strings.ToLower(config.CustomAPIURL), "bypass")
		mcpClient = mcp.NewOllamaClient()
		if bypassEnabled {
			// Even when using browser automation, we need to set API key for base client validation
			mcpClient.SetAPIKey("dummy-key-for-browser-automation", config.CustomAPIURL, config.CustomModelName)
			// For bypass clients, we need to ensure trader ID is passed through
			if guardianClient, ok := mcpClient.(*mcp.GuardianClient); ok {
				mcpClient = mcp.NewBypassClientWithTraderID(mcpClient, true, "guardian-ai", guardianClient.GetTraderID())
			} else {
				mcpClient = mcp.NewBypassClient(mcpClient, true, "guardian-ai")
			}
			logger.Infof("🤖 [%s] Using OLLAMA with Browser Automation Bypass", config.Name)
		} else {
			// Ollama typically doesn't need an API key, but we'll use it if provided
			apiKey := config.CustomAPIKey
			if apiKey == "" {
				apiKey = "ollama" // Default API key for Ollama if not provided
			}
			mcpClient.SetAPIKey(apiKey, config.CustomAPIURL, config.CustomModelName)
			logger.Infof("🤖 [%s] Using Ollama AI: %s (model: %s)", config.Name, config.CustomAPIURL, config.CustomModelName)
		}

	case "guardian":
		// AI model handles its own implementation internally
		logger.Warnf("🚨 [GUARDIAN DEBUG] Creating GuardianClient with TraderID: %s", config.ID)
		mcpClient = mcp.NewGuardianClientWithOptions(
			mcp.WithTraderID(config.ID), // Pass trader ID for data isolation
		)
		if guardianOriginal, ok := mcpClient.(*mcp.GuardianClient); ok {
			logger.Warnf("🚨 [GUARDIAN DEBUG] Original GuardianClient TraderID: '%s'", guardianOriginal.GetTraderID())
		}
		if config.CustomAPIURL != "" {
			if guardianClient, ok := mcpClient.(*mcp.GuardianClient); ok {
				guardianClient.SetDynamicConfig(config.CustomAPIURL)
			}
		}
		logger.Infof("🤖 [%s] Using Guardian AI (delegating implementation to AI model)", config.Name)

	case "guardian-ai":
		// Trader delegates to AI model - let the AI model handle its own implementation
		baseClient := mcp.NewGuardianClientWithOptions(
			mcp.WithTraderID(config.ID), // Pass trader ID for data isolation
		)
		// AI model internally decides whether to use browser automation or API calls
		// Trader just uses the unified interface
		mcpClient = baseClient
		logger.Infof("🤖 [%s] Using Guardian AI (delegating to AI model for implementation)", config.Name)

	case "deepseek":
		// AI model handles its own implementation internally
		mcpClient = mcp.NewDeepSeekClient()
		apiKey := config.DeepSeekKey
		if apiKey == "" {
			apiKey = config.CustomAPIKey
		}
		mcpClient.SetAPIKey(apiKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using DeepSeek AI (delegating implementation to AI model)", config.Name)

	default: // deepseek or empty
		mcpClient = mcp.NewDeepSeekClient()
		apiKey := config.DeepSeekKey
		if apiKey == "" {
			apiKey = config.CustomAPIKey
		}
		mcpClient.SetAPIKey(apiKey, config.CustomAPIURL, config.CustomModelName)
		logger.Infof("🤖 [%s] Using DeepSeek AI", config.Name)
	}

	if config.CustomAPIURL != "" || config.CustomModelName != "" {
		logger.Infof("🔧 [%s] Custom config - URL: %s, Model: %s", config.Name, config.CustomAPIURL, config.CustomModelName)
	}

	// Set default trading platform
	if config.Exchange == "" {
		config.Exchange = "binance"
	}

	// Create corresponding trader based on configuration
	var trader Trader
	var err error

	// Record position mode (general)
	marginModeStr := "Cross Margin"
	if !config.IsCrossMargin {
		marginModeStr = "Isolated Margin"
	}
	logger.Infof("📊 [%s] Position mode: %s", config.Name, marginModeStr)

	switch config.Exchange {
	case "binance":
		logger.Infof("🏦 [%s] Using Binance Futures trading", config.Name)
		// 检查全局代理开关
		useProxyGlobal := os.Getenv("USE_BINANCE_PROXY") == "true"

		// 根据全局设置决定使用哪种端点
		var endpoint string
		var targetEndpoint string
		if useProxyGlobal {
			// 对于代理模式，我们需要连接到代理服务，但告诉代理真正的目标URL
			proxyURL := os.Getenv("BINANCE_PROXY_URL")
			if proxyURL == "" {
				proxyURL = "http://localhost:8081" // 默认代理URL
			}
			endpoint = proxyURL
			// 真正的目标端点应该是自定义API URL或默认的交易所URL
			targetEndpoint = getBinanceCustomEndpointForAutoTrader(&config)
			logger.Infof("🔧 [DEBUG] BinanceCustomAPIURL from config: '%s', targetEndpoint after getBinanceCustomEndpointForAutoTrader: '%s'",
				config.BinanceCustomAPIURL, targetEndpoint)
			if targetEndpoint == "" {
				targetEndpoint = "https://fapi.binance.com" // Always default to mainnet
			}
			logger.Infof("🔧 [DEBUG] Final targetEndpoint: '%s'", targetEndpoint)

			// 在代理模式下，创建一个连接到代理服务的交易者实例
			originalTrader := NewFuturesTraderViaProxy(config.BinanceAPIKey, config.BinanceSecretKey, userID, endpoint, targetEndpoint)
			trader = NewProxyTraderWrapperWithAuth(originalTrader, "proxy", os.Getenv("BINANCE_PROXY_URL"), config.BinanceAPIKey, config.BinanceSecretKey, targetEndpoint)
		} else {
			endpoint = getBinanceCustomEndpointForAutoTrader(&config)
			targetEndpoint = endpoint

			// 在非代理模式下，直接连接到真实交易所
			originalTrader := NewFuturesTrader(config.BinanceAPIKey, config.BinanceSecretKey, userID, endpoint)
			trader = NewProxyTraderWrapperWithAuth(originalTrader, "native", "", config.BinanceAPIKey, config.BinanceSecretKey, targetEndpoint)
		}
	case "bybit":
		logger.Infof("🏦 [%s] Using Bybit Futures trading", config.Name)
		trader = NewBybitTrader(config.BybitAPIKey, config.BybitSecretKey)
	case "okx":
		logger.Infof("🏦 [%s] Using OKX Futures trading", config.Name)
		trader = NewOKXTrader(config.OKXAPIKey, config.OKXSecretKey, config.OKXPassphrase)
	case "bitget":
		logger.Infof("🏦 [%s] Using Bitget Futures trading", config.Name)
		trader = NewBitgetTrader(config.BitgetAPIKey, config.BitgetSecretKey, config.BitgetPassphrase)
	case "hyperliquid":
		logger.Infof("🏦 [%s] Using Hyperliquid trading", config.Name)
		trader, err = NewHyperliquidTrader(config.HyperliquidPrivateKey, config.HyperliquidWalletAddr, config.HyperliquidTestnet)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Hyperliquid trader: %w", err)
		}
	case "aster":
		logger.Infof("🏦 [%s] Using Aster trading", config.Name)
		trader, err = NewAsterTrader(config.AsterUser, config.AsterSigner, config.AsterPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Aster trader: %w", err)
		}
	case "lighter":
		logger.Infof("🏦 [%s] Using LIGHTER trading", config.Name)

		if config.LighterWalletAddr == "" || config.LighterAPIKeyPrivateKey == "" {
			return nil, fmt.Errorf("Lighter requires wallet address and API Key private key")
		}

		// Lighter only supports mainnet (testnet disabled)
		trader, err = NewLighterTraderV2(
			config.LighterWalletAddr,
			config.LighterAPIKeyPrivateKey,
			config.LighterAPIKeyIndex,
			false, // Always use mainnet for Lighter
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize LIGHTER trader: %w", err)
		}
		logger.Infof("✓ LIGHTER trader initialized successfully")
	default:
		return nil, fmt.Errorf("unsupported trading platform: %s", config.Exchange)
	}

	// Validate initial balance configuration, auto-fetch from exchange if 0
	if config.InitialBalance <= 0 {
		logger.Infof("📊 [%s] Initial balance not set, attempting to fetch current balance from exchange...", config.Name)
		account, err := trader.GetBalance()
		if err != nil {
			return nil, fmt.Errorf("initial balance not set and unable to fetch balance from exchange: %w", err)
		}
		// Try multiple balance field names (different exchanges return different formats)
		balanceKeys := []string{"total_equity", "totalWalletBalance", "wallet_balance", "totalEq", "balance"}
		var foundBalance float64
		for _, key := range balanceKeys {
			if balance, ok := account[key].(float64); ok && balance > 0 {
				foundBalance = balance
				break
			}
		}
		if foundBalance > 0 {
			config.InitialBalance = foundBalance
			logger.Infof("✓ [%s] Auto-fetched initial balance: %.2f USDT", config.Name, foundBalance)
			// Save to database so it persists across restarts
			if st != nil {
				if err := st.Trader().UpdateInitialBalance(userID, config.ID, foundBalance); err != nil {
					logger.Infof("⚠️  [%s] Failed to save initial balance to database: %v", config.Name, err)
				} else {
					logger.Infof("✓ [%s] Initial balance saved to database", config.Name)
				}
			}
		} else {
			return nil, fmt.Errorf("initial balance must be greater than 0, please set InitialBalance in config or ensure exchange account has balance")
		}
	}

	// Get last cycle number (for recovery)
	var cycleNumber int
	if st != nil {
		cycleNumber, _ = st.Decision().GetLastCycleNumber(config.ID)
		logger.Infof("📊 [%s] Decision records will be stored to database", config.Name)
	}

	// Create strategy engine (must have strategy config)
	if config.StrategyConfig == nil {
		return nil, fmt.Errorf("[%s] strategy not configured", config.Name)
	}
	strategyEngine := kernel.NewStrategyEngine(config.StrategyConfig)
	logger.Infof("✓ [%s] Using strategy engine (strategy configuration loaded)", config.Name)

	// 🔥 智能计算手动扫描延迟时长（基于扫描周期）
	scanDelayDuration := calculateScanDelay(config.ScanInterval)
	logger.Infof("⏰ [%s] 手动扫描延迟配置: %.0f分钟 (扫描周期: %.0f分钟)",
		config.Name, scanDelayDuration.Minutes(), config.ScanInterval.Minutes())

	return &AutoTrader{
		id:                    config.ID,
		name:                  config.Name,
		aiModel:               config.AIModel,
		exchange:              config.Exchange,
		exchangeID:            config.ExchangeID,
		showInCompetition:     config.ShowInCompetition,
		config:                config,
		trader:                trader,
		mcpClient:             mcpClient,
		store:                 st,
		strategyEngine:        strategyEngine,
		cycleNumber:           cycleNumber,
		initialBalance:        config.InitialBalance,
		lastResetTime:         time.Now(),
		startTime:             time.Now(),
		callCount:             0,
		isRunning:             false,
		positionFirstSeenTime: make(map[string]int64),
		stopMonitorCh:         make(chan struct{}),
		monitorWg:             sync.WaitGroup{},
		peakPnLCache:          make(map[string]float64),
		peakPnLCacheMutex:     sync.RWMutex{},
		lastBalanceSyncTime:   time.Now(),
		userID:                userID,
		tradeFrequencyTracker: &TradeFrequencyTracker{
			dailyTrades:           0,
			hourlyTrades:          0,
			symbolHourlyTrades:    make(map[string]int),
			dailyResetTime:        time.Now(),
			hourlyResetTime:       time.Now(),
			symbolHourlyResetTime: time.Now(),
		},
		scanDelayDuration:  5 * time.Minute,                     // 🔥 新增：手动扫描后默认延迟5分钟（基于实测：本地大模型扫描2-4分钟）
		nextSystemScanTime: time.Now().Add(config.ScanInterval), // 🔥 新增：初始化下次扫描时间
	}, nil
}

// Run runs the automatic trading main loop
func (at *AutoTrader) Run() error {
	// 初始化随机种子以确保随机数真正随机
	rand.Seed(time.Now().UnixNano())

	at.isRunningMutex.Lock()
	at.isRunning = true
	at.isRunningMutex.Unlock()

	at.stopMonitorCh = make(chan struct{})
	at.startTime = time.Now()

	logger.Info("🚀 AI-driven automatic trading system started")
	logger.Infof("💰 Initial balance: %.2f USDT", at.initialBalance)
	logger.Infof("⚙️  Scan interval: %v", at.config.ScanInterval)
	logger.Info("🤖 AI will make full decisions on leverage, position size, stop loss/take profit, etc.")
	at.monitorWg.Add(1)
	defer at.monitorWg.Done()

	// Start drawdown monitoring
	at.startDrawdownMonitor()

	// Start Lighter order sync if using Lighter exchange
	if at.exchange == "lighter" {
		if lighterTrader, ok := at.trader.(*LighterTraderV2); ok && at.store != nil {
			lighterTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Lighter order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start Hyperliquid order sync if using Hyperliquid exchange
	if at.exchange == "hyperliquid" {
		if hyperliquidTrader, ok := at.trader.(*HyperliquidTrader); ok && at.store != nil {
			hyperliquidTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Hyperliquid order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start Bybit order sync if using Bybit exchange
	if at.exchange == "bybit" {
		if bybitTrader, ok := at.trader.(*BybitTrader); ok && at.store != nil {
			bybitTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Bybit order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start OKX order sync if using OKX exchange
	if at.exchange == "okx" {
		if okxTrader, ok := at.trader.(*OKXTrader); ok && at.store != nil {
			okxTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] OKX order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start Bitget order sync if using Bitget exchange
	if at.exchange == "bitget" {
		if bitgetTrader, ok := at.trader.(*BitgetTrader); ok && at.store != nil {
			bitgetTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Bitget order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start Aster order sync if using Aster exchange
	if at.exchange == "aster" {
		if asterTrader, ok := at.trader.(*AsterTrader); ok && at.store != nil {
			asterTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Aster order+position sync enabled (every 30s)", at.name)
		}
	}

	// Start Binance order sync if using Binance exchange
	// Handle both direct FuturesTrader and ProxyTraderWrapper cases
	if at.exchange == "binance" {
		// Check if it's a direct FuturesTrader
		if binanceTrader, ok := at.trader.(*FuturesTrader); ok && at.store != nil {
			binanceTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
			logger.Infof("🔄 [%s] Binance order+position sync enabled (every 30s) - Direct Trader", at.name)
		} else if proxyWrapper, ok := at.trader.(*ProxyTraderWrapper); ok {
			// For proxy wrapper, we need to check if the underlying trader is FuturesTrader
			if proxyWrapper.trader != nil {
				if binanceTrader, ok := proxyWrapper.trader.(*FuturesTrader); ok && at.store != nil {
					binanceTrader.StartOrderSync(at.id, at.exchangeID, at.exchange, at.store, 30*time.Second)
					logger.Infof("🔄 [%s] Binance order+position sync enabled (every 30s) - Proxy Wrapped Trader", at.name)
				} else {
					logger.Infof("⚠️ [%s] Binance trader type mismatch - expected *FuturesTrader, got %T", at.name, proxyWrapper.trader)
				}
			} else {
				logger.Infof("⚠️ [%s] Proxy wrapper has nil underlying trader", at.name)
			}
		} else {
			logger.Infof("⚠️ [%s] Binance trader type mismatch - expected *FuturesTrader or *ProxyTraderWrapper, got %T", at.name, at.trader)
		}
	}

	ticker := time.NewTicker(at.config.ScanInterval)
	defer ticker.Stop()

	// Execute immediately on first run
	if err := at.runCycleWithExecutionLock(); err != nil {
		logger.Infof("❌ Execution failed: %v", err)
	}
	// 🔥 初始化下次扫描时间
	at.nextSystemScanTime = time.Now().Add(at.config.ScanInterval)

	for {
		at.isRunningMutex.RLock()
		running := at.isRunning
		at.isRunningMutex.RUnlock()

		if !running {
			break
		}

		select {
		case <-ticker.C:
			// 🔥 新增：检查是否需要因手动扫描而延迟
			now := time.Now()
			if !at.lastManualScanTime.IsZero() && now.Before(at.lastManualScanTime.Add(at.scanDelayDuration)) {
				remainingDelay := at.lastManualScanTime.Add(at.scanDelayDuration).Sub(now)
				logger.Infof("⏳ [%s] 系统扫描延迟中（手动扫描后 %.0f 秒内不触发），剩余: %.0f秒",
					at.name, at.scanDelayDuration.Seconds(), remainingDelay.Seconds())
				// 更新下次扫描时间
				at.nextSystemScanTime = at.lastManualScanTime.Add(at.scanDelayDuration).Add(at.config.ScanInterval)
				continue
			}

			if err := at.runCycleWithExecutionLock(); err != nil {
				// Only log error if it's not because manual scan is in progress
				if err.Error() != "automatic scan skipped: manual scan in progress" {
					logger.Infof("❌ Execution failed: %v", err)
				}
			}

			// 添加随机延迟以避免人机检测
			randomDelay := time.Duration(rand.Intn(30)) * time.Second // 随机0-30秒延迟
			logger.Infof("🎲 [%s] Adding random delay: %.0f seconds to avoid bot detection", at.name, randomDelay.Seconds())
			time.Sleep(randomDelay)

			// 🔥 更新下次扫描时间（加上随机延迟）
			at.nextSystemScanTime = time.Now().Add(at.config.ScanInterval)
		case <-at.stopMonitorCh:
			logger.Infof("[%s] ⏹ Stop signal received, exiting automatic trading main loop", at.name)
			return nil
		}
	}

	return nil
}

// Stop stops the automatic trading
func (at *AutoTrader) Stop() {
	at.isRunningMutex.Lock()
	if !at.isRunning {
		at.isRunningMutex.Unlock()
		return
	}
	at.isRunning = false
	at.isRunningMutex.Unlock()

	close(at.stopMonitorCh) // Notify monitoring goroutine to stop
	at.monitorWg.Wait()     // Wait for monitoring goroutine to finish
	logger.Info("⏹ Automatic trading system stopped")
}

// runCycle runs one trading cycle (using AI full decision-making)
func (at *AutoTrader) runCycle() error {
	at.callCount++

	logger.Info("\n" + strings.Repeat("=", 70) + "\n")
	logger.Infof("⏰ %s - AI decision cycle #%d", time.Now().Format("2006-01-02 15:04:05"), at.callCount)
	logger.Info(strings.Repeat("=", 70))

	// 0. Check if trader is stopped (early exit to prevent trades after Stop() is called)
	at.isRunningMutex.RLock()
	running := at.isRunning
	at.isRunningMutex.RUnlock()
	if !running {
		// Enhanced stop reason detection to avoid false positives
		stopReason := at.determineStopReason()

		logger.Infof("⏹ [%s] Trader was stopped before starting cycle #%d", stopReason, at.callCount)

		// Enhanced error message with specific reason
		var errorMessage string
		switch stopReason {
		case "USER_MANUAL_STOP":
			errorMessage = fmt.Sprintf("[USER_MANUAL_STOP] Trader was manually stopped by user before starting decision cycle #%d - this is normal behavior when user stops trader", at.callCount)
		case "RISK_CONTROL_AUTO_PAUSE":
			remaining := at.stopUntil.Sub(time.Now())
			errorMessage = fmt.Sprintf("[RISK_CONTROL_AUTO_PAUSE] Trading automatically paused by system risk control for %.0f more minutes - cycle #%d blocked for safety", remaining.Minutes(), at.callCount)
		case "SYSTEM_ERROR_STOP":
			errorMessage = fmt.Sprintf("[SYSTEM_ERROR_STOP] Trader stopped due to system error or abnormal condition in cycle #%d - please check system logs", at.callCount)
		default:
			errorMessage = fmt.Sprintf("[UNKNOWN_STOP] Trader stopped for unknown reason before starting cycle #%d", at.callCount)
		}

		// Create minimal record for this case
		minimalRecord := &store.DecisionRecord{
			Success:      false,
			ErrorMessage: errorMessage,
		}
		at.saveDecision(minimalRecord)
		return nil
	}

	// Create decision record
	record := &store.DecisionRecord{
		ExecutionLog: []string{},
		Success:      true,
	}

	// 1. Check if trading needs to be stopped (risk control auto-pause)
	if time.Now().Before(at.stopUntil) {
		remaining := at.stopUntil.Sub(time.Now())
		logger.Infof("⏸ [RISK_CONTROL_AUTO_PAUSE] Trading automatically paused by system risk control, remaining %.0f minutes", remaining.Minutes())
		logger.Infof("   - Safety mechanism activated to protect capital")
		logger.Infof("   - Trading will resume automatically after pause period")

		record.Success = false
		record.ErrorMessage = fmt.Sprintf("[RISK_CONTROL_AUTO_PAUSE] Trading automatically paused by system risk control for %.0f more minutes - safety mechanism activated to protect capital", remaining.Minutes())
		at.saveDecision(record)
		return nil
	}

	// 2. Reset daily P&L (reset every day)
	if time.Since(at.lastResetTime) > 24*time.Hour {
		at.dailyPnL = 0
		at.lastResetTime = time.Now()
		logger.Info("📅 Daily P&L reset")
	}

	// 4. Collect trading context
	ctx, err := at.buildTradingContext()
	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("Failed to build trading context: %v", err)
		at.saveDecision(record)
		return fmt.Errorf("failed to build trading context: %w", err)
	}

	// Save equity snapshot independently (decoupled from AI decision, used for drawing profit curve)
	at.saveEquitySnapshot(ctx)

	logger.Info(strings.Repeat("=", 70))
	for _, coin := range ctx.CandidateCoins {
		record.CandidateCoins = append(record.CandidateCoins, coin.Symbol)
	}

	logger.Infof("📊 Account equity: %.2f USDT | Available: %.2f USDT | Positions: %d",
		ctx.Account.TotalEquity, ctx.Account.AvailableBalance, ctx.Account.PositionCount)

	// 5. Use strategy engine to call AI for decision
	logger.Infof("🤖 Requesting AI analysis and decision... [Strategy Engine]")
	aiDecision, err := kernel.GetFullDecisionWithStrategy(ctx, at.mcpClient, at.strategyEngine, "balanced")

	if aiDecision != nil && aiDecision.AIRequestDurationMs > 0 {
		record.AIRequestDurationMs = aiDecision.AIRequestDurationMs
		logger.Infof("⏱️ AI call duration: %.2f seconds", float64(record.AIRequestDurationMs)/1000)
		record.ExecutionLog = append(record.ExecutionLog,
			fmt.Sprintf("AI call duration: %d ms", record.AIRequestDurationMs))
	}

	// Save chain of thought, decisions, and input prompt even if there's an error (for debugging)
	if aiDecision != nil {
		record.SystemPrompt = aiDecision.SystemPrompt // Save system prompt
		record.InputPrompt = aiDecision.UserPrompt
		record.CoTTrace = aiDecision.CoTTrace
		record.RawResponse = aiDecision.RawResponse // Save raw AI response for debugging
		if len(aiDecision.Decisions) > 0 {
			decisionJSON, _ := json.MarshalIndent(aiDecision.Decisions, "", "  ")
			record.DecisionJSON = string(decisionJSON)
		}
	}

	if err != nil {
		record.Success = false
		record.ErrorMessage = fmt.Sprintf("Failed to get AI decision: %v", err)

		// Print system prompt and AI chain of thought (output even with errors for debugging)
		// DISABLED: Removed to reduce log size and improve restart performance
		// if aiDecision != nil {
		// 	logger.Info("\n" + strings.Repeat("=", 70) + "\n")
		// 	logger.Infof("📋 System prompt (error case)")
		// 	logger.Info(strings.Repeat("=", 70))
		// 	logger.Info(aiDecision.SystemPrompt)
		// 	logger.Info(strings.Repeat("=", 70))

		// 	if aiDecision.CoTTrace != "" {
		// 		logger.Info("\n" + strings.Repeat("-", 70) + "\n")
		// 		logger.Info("💭 AI chain of thought analysis (error case):")
		// 		logger.Info(strings.Repeat("-", 70))
		// 		logger.Info(aiDecision.CoTTrace)
		// 		logger.Info(strings.Repeat("-", 70))
		// 	}
		// }

		at.saveDecision(record)
		return fmt.Errorf("failed to get AI decision: %w", err)
	}

	// // 5. Print system prompt
	// logger.Infof("\n" + strings.Repeat("=", 70))
	// logger.Infof("📋 System prompt [template: %s]", at.systemPromptTemplate)
	// logger.Info(strings.Repeat("=", 70))
	// logger.Info(decision.SystemPrompt)
	// logger.Infof(strings.Repeat("=", 70) + "\n")

	// 6. Print AI chain of thought
	// logger.Infof("\n" + strings.Repeat("-", 70))
	// logger.Info("💭 AI chain of thought analysis:")
	// logger.Info(strings.Repeat("-", 70))
	// logger.Info(decision.CoTTrace)
	// logger.Infof(strings.Repeat("-", 70) + "\n")

	// 7. Print AI decisions
	// logger.Infof("📋 AI decision list (%d items):\n", len(kernel.Decisions))
	// for i, d := range kernel.Decisions {
	//     logger.Infof("  [%d] %s: %s - %s", i+1, d.Symbol, d.Action, d.Reasoning)
	//     if d.Action == "open_long" || d.Action == "open_short" {
	//        logger.Infof("      Leverage: %dx | Position: %.2f USDT | Stop loss: %.4f | Take profit: %.4f",
	//           d.Leverage, d.PositionSizeUSD, d.StopLoss, d.TakeProfit)
	//     }
	// }
	logger.Info()
	logger.Info(strings.Repeat("-", 70))
	// 8. Sort decisions: ensure close positions first, then open positions (prevent position stacking overflow)
	logger.Info(strings.Repeat("-", 70))

	// 8. Sort decisions: ensure close positions first, then open positions (prevent position stacking overflow)
	sortedDecisions := sortDecisionsByPriority(aiDecision.Decisions)

	logger.Info("🔄 Execution order (optimized): Close positions first → Open positions later")
	for i, d := range sortedDecisions {
		logger.Infof("  [%d] %s %s", i+1, d.Symbol, d.Action)
	}
	logger.Info()

	// Check if trader is stopped before executing any decisions (prevent trades after Stop())
	at.isRunningMutex.RLock()
	running = at.isRunning
	at.isRunningMutex.RUnlock()
	if !running {
		stopReason := at.determineStopReason()

		logger.Infof("⏹ [%s] Trader stopped before decision execution, cycle #%d", stopReason, at.callCount)
		logger.Infof("   - AI decisions generated: %d", len(sortedDecisions))
		logger.Infof("   - Execution aborted to respect stop command")

		// Enhanced error message with specific reason
		var errorMessage string
		switch stopReason {
		case "USER_MANUAL_STOP":
			errorMessage = fmt.Sprintf("[USER_MANUAL_STOP] Trader was manually stopped by user before executing %d decisions - this is normal behavior when user stops trader during AI processing", len(sortedDecisions))
		case "RISK_CONTROL_AUTO_PAUSE":
			remaining := at.stopUntil.Sub(time.Now())
			errorMessage = fmt.Sprintf("[RISK_CONTROL_AUTO_PAUSE] Trading automatically paused by system risk control for %.0f more minutes - execution blocked for safety", remaining.Minutes())
		case "SYSTEM_ERROR_STOP":
			errorMessage = fmt.Sprintf("[SYSTEM_ERROR_STOP] Trader stopped due to system error before executing %d decisions - please check system logs", len(sortedDecisions))
		default:
			errorMessage = fmt.Sprintf("[UNKNOWN_STOP] Trader stopped for unknown reason before executing %d decisions", len(sortedDecisions))
		}

		// Even though we're not executing the decisions, save the AI-generated decision record for tracking
		record.Success = false
		record.ErrorMessage = errorMessage
		if err := at.saveDecision(record); err != nil {
			logger.Infof("⚠ Failed to save decision record: %v", err)
		}
		return nil
	}

	// Execute decisions and record results
	for _, d := range sortedDecisions {
		// Check if trader is stopped before each decision (allow immediate stop during execution)
		at.isRunningMutex.RLock()
		running = at.isRunning
		at.isRunningMutex.RUnlock()
		if !running {
			stopReason := at.determineStopReason()

			logger.Infof("⏹ [%s] Trader stopped during decision execution, aborting remaining decisions", stopReason)
			break
		}

		actionRecord := store.DecisionAction{
			Action:     d.Action,
			Symbol:     d.Symbol,
			Quantity:   0,
			Leverage:   d.Leverage,
			Price:      0,
			StopLoss:   d.StopLoss,
			TakeProfit: d.TakeProfit,
			Confidence: d.Confidence,
			Reasoning:  d.Reasoning,
			Timestamp:  time.Now().UTC(),
			Success:    false,

			// Additional parameters for advanced action types
			NewStopLoss:               d.NewStopLoss,
			NewTakeProfit:             d.NewTakeProfit,
			ClosePercentage:           d.ClosePercentage,
			TrailPercentage:           d.TrailPercentage,
			CallbackRate:              d.CallbackRate,
			TargetROI:                 d.TargetROI,
			MaxROI:                    d.MaxROI,
			TimeLimitHours:            d.TimeLimitHours,
			AdditionalPositionSizeUSD: d.AdditionalPositionSizeUSD,
		}

		if err := at.executeDecisionWithRecord(&d, &actionRecord); err != nil {
			logger.Infof("❌ Failed to execute decision (%s %s): %v", d.Symbol, d.Action, err)
			actionRecord.Error = err.Error()
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("❌ %s %s failed: %v", d.Symbol, d.Action, err))
		} else {
			actionRecord.Success = true
			record.ExecutionLog = append(record.ExecutionLog, fmt.Sprintf("✓ %s %s succeeded", d.Symbol, d.Action))
			// Brief delay after successful execution
			time.Sleep(1 * time.Second)
		}

		record.Decisions = append(record.Decisions, actionRecord)
	}

	// 9. Save decision record
	if err := at.saveDecision(record); err != nil {
		logger.Infof("⚠ Failed to save decision record: %v", err)
	}

	return nil
}

// buildTradingContext builds trading context
func (at *AutoTrader) buildTradingContext() (*kernel.Context, error) {
	// 1. Get account information
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", err)
	}

	// Get account fields
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0
	totalEquity := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Use totalEquity directly if provided by trader (more accurate)
	if eq, ok := balance["totalEquity"].(float64); ok && eq > 0 {
		totalEquity = eq
	} else {
		// Fallback: Total Equity = Wallet balance + Unrealized profit
		totalEquity = totalWalletBalance + totalUnrealizedProfit
	}

	// 2. Get position information
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var positionInfos []kernel.PositionInfo
	totalMarginUsed := 0.0

	// Current position key set (for cleaning up closed position records)
	currentPositionKeys := make(map[string]bool)

	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity, ok := pos["positionAmt"].(float64)
		if !ok {
			// Try to get as json.Number if it fails as float64
			quantityNum, ok2 := pos["positionAmt"].(*json.Number)
			if !ok2 {
				logger.Warnf("Failed to get position amount for %s, using 0", pos["symbol"])
				quantity = 0
			} else {
				var err error
				quantity, err = quantityNum.Float64()
				if err != nil {
					logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
					quantity = 0
				}
			}
		}
		if quantity < 0 {
			quantity = -quantity // Short position quantity is negative, convert to positive
		}

		// Skip closed positions (quantity = 0), prevent "ghost positions" from being passed to AI
		if quantity == 0 {
			continue
		}

		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		// Calculate margin used (estimated)
		leverage := 10 // Default value, should actually be fetched from position info
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed

		// Calculate P&L percentage (based on margin, considering leverage)
		pnlPct := calculatePnLPercentage(unrealizedPnl, marginUsed)

		// Get position open time from exchange (preferred) or fallback to local tracking
		posKey := symbol + "_" + side
		currentPositionKeys[posKey] = true

		var updateTime int64
		// Priority 1: Get from database (trader_positions table) - most accurate
		if at.store != nil {
			if dbPos, err := at.store.Position().GetOpenPositionBySymbol(at.id, symbol, side); err == nil && dbPos != nil {
				if dbPos.EntryTime > 0 {
					updateTime = dbPos.EntryTime
				}
			}
		}
		// Priority 2: Get from exchange API (Bybit: createdTime, OKX: createdTime)
		if updateTime == 0 {
			if createdTime, ok := pos["createdTime"].(int64); ok && createdTime > 0 {
				updateTime = createdTime
			}
		}
		// Priority 3: Fallback to local tracking
		if updateTime == 0 {
			if _, exists := at.positionFirstSeenTime[posKey]; !exists {
				at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()
			}
			updateTime = at.positionFirstSeenTime[posKey]
		}

		// Get peak profit rate for this position
		at.peakPnLCacheMutex.RLock()
		peakPnlPct := at.peakPnLCache[posKey]
		at.peakPnLCacheMutex.RUnlock()

		positionInfos = append(positionInfos, kernel.PositionInfo{
			Symbol:           symbol,
			Side:             side,
			EntryPrice:       entryPrice,
			MarkPrice:        markPrice,
			Quantity:         quantity,
			Leverage:         leverage,
			UnrealizedPnL:    unrealizedPnl,
			UnrealizedPnLPct: pnlPct,
			PeakPnLPct:       peakPnlPct,
			LiquidationPrice: liquidationPrice,
			MarginUsed:       marginUsed,
			UpdateTime:       updateTime,
		})
	}

	// Clean up closed position records
	for key := range at.positionFirstSeenTime {
		if !currentPositionKeys[key] {
			delete(at.positionFirstSeenTime, key)
		}
	}

	// 3. Use strategy engine to get candidate coins (must have strategy engine)
	if at.strategyEngine == nil {
		return nil, fmt.Errorf("trader has no strategy engine configured")
	}
	candidateCoins, err := at.strategyEngine.GetCandidateCoins()
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate coins: %w", err)
	}
	// [DEBUG] Strategy candidate coins count - Enable when debugging strategy selection
	// logger.Infof("📋 [%s] Strategy engine fetched candidate coins: %d", at.name, len(candidateCoins))

	// 4. Calculate total P&L
	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	// 5. Get leverage from strategy config
	strategyConfig := at.strategyEngine.GetConfig()
	btcEthLeverage := strategyConfig.RiskControl.BTCETHMaxLeverage
	altcoinLeverage := strategyConfig.RiskControl.AltcoinMaxLeverage
	// [DEBUG] Leverage configuration - Enable when verifying leverage settings
	// logger.Infof("📋 [%s] Strategy leverage config: BTC/ETH=%dx, Altcoin=%dx", at.name, btcEthLeverage, altcoinLeverage)

	// 6. Build context
	ctx := &kernel.Context{
		CurrentTime:     time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		RuntimeMinutes:  int(time.Since(at.startTime).Minutes()),
		CallCount:       at.callCount,
		BTCETHLeverage:  btcEthLeverage,
		AltcoinLeverage: altcoinLeverage,
		Account: kernel.AccountInfo{
			TotalEquity:      totalEquity,
			AvailableBalance: availableBalance,
			UnrealizedPnL:    totalUnrealizedProfit,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			MarginUsed:       totalMarginUsed,
			MarginUsedPct:    marginUsedPct,
			PositionCount:    len(positionInfos),
		},
		Positions:      positionInfos,
		CandidateCoins: candidateCoins,
	}

	// 7. Add recent closed trades (if store is available)
	if at.store != nil {
		// Get recent 10 closed trades for AI context
		recentTrades, err := at.store.Position().GetRecentTrades(at.id, 10)
		if err != nil {
			logger.Infof("⚠️ [%s] Failed to get recent trades: %v", at.name, err)
		} else {
			// [DEBUG] Recent trades count - Enable when debugging trade history
			// logger.Infof("📊 [%s] Found %d recent closed trades for AI context", at.name, len(recentTrades))
			for _, trade := range recentTrades {
				// Convert Unix timestamps to formatted strings for AI readability
				entryTimeStr := ""
				if trade.EntryTime > 0 {
					entryTimeStr = time.Unix(trade.EntryTime, 0).UTC().Format("01-02 15:04 UTC")
				}
				exitTimeStr := ""
				if trade.ExitTime > 0 {
					exitTimeStr = time.Unix(trade.ExitTime, 0).UTC().Format("01-02 15:04 UTC")
				}

				ctx.RecentOrders = append(ctx.RecentOrders, kernel.RecentOrder{
					Symbol:       trade.Symbol,
					Side:         trade.Side,
					EntryPrice:   trade.EntryPrice,
					ExitPrice:    trade.ExitPrice,
					RealizedPnL:  trade.RealizedPnL,
					PnLPct:       trade.PnLPct,
					EntryTime:    entryTimeStr,
					ExitTime:     exitTimeStr,
					HoldDuration: trade.HoldDuration,
				})
			}
		}
		// Get trading statistics for AI context
		stats, err := at.store.Position().GetFullStats(at.id)
		if err != nil {
			logger.Infof("⚠️ [%s] Failed to get trading stats: %v", at.name, err)
		} else if stats == nil {
			logger.Infof("⚠️ [%s] GetFullStats returned nil", at.name)
		} else if stats.TotalTrades == 0 {
			logger.Infof("⚠️ [%s] GetFullStats returned 0 trades (traderID=%s)", at.name, at.id)
		} else {
			ctx.TradingStats = &kernel.TradingStats{
				TotalTrades:    stats.TotalTrades,
				WinRate:        stats.WinRate,
				ProfitFactor:   stats.ProfitFactor,
				SharpeRatio:    stats.SharpeRatio,
				TotalPnL:       stats.TotalPnL,
				AvgWin:         stats.AvgWin,
				AvgLoss:        stats.AvgLoss,
				MaxDrawdownPct: stats.MaxDrawdownPct,
			}
			// [DEBUG] Trading statistics - Enable when analyzing performance
			// logger.Infof("📈 [%s] Trading stats: %d trades, %.1f%% win rate, PF=%.2f, Sharpe=%.2f, DD=%.1f%%",
			// 	at.name, stats.TotalTrades, stats.WinRate, stats.ProfitFactor, stats.SharpeRatio, stats.MaxDrawdownPct)
		}
	} else {
		logger.Infof("⚠️ [%s] Store is nil, cannot get recent trades", at.name)
	}

	// 8. Get quantitative data (if enabled in strategy config)
	if strategyConfig.Indicators.EnableQuantData {
		// Collect symbols to query (candidate coins + position coins)
		symbolsToQuery := make(map[string]bool)
		for _, coin := range candidateCoins {
			symbolsToQuery[coin.Symbol] = true
		}
		for _, pos := range positionInfos {
			symbolsToQuery[pos.Symbol] = true
		}

		symbols := make([]string, 0, len(symbolsToQuery))
		for sym := range symbolsToQuery {
			symbols = append(symbols, sym)
		}

		// [DEBUG] Quantitative data fetching - Enable when debugging quant data interface
		// logger.Infof("📊 [%s] Fetching quantitative data for %d symbols...", at.name, len(symbols))
		ctx.QuantDataMap = at.strategyEngine.FetchQuantDataBatch(symbols)
		// [DEBUG] Quantitative data result - Enable when debugging quant data interface
		// logger.Infof("📊 [%s] Successfully fetched quantitative data for %d symbols", at.name, len(ctx.QuantDataMap))
	}

	// 9. Get OI ranking data (market-wide position changes)
	if strategyConfig.Indicators.EnableOIRanking {
		// [DEBUG] OI ranking fetch - Enable when debugging OI ranking feature
		// logger.Infof("📊 [%s] Fetching OI ranking data...", at.name)
		ctx.OIRankingData = at.strategyEngine.FetchOIRankingData()
		// [DEBUG] OI ranking result - Enable when debugging OI ranking feature
		// if ctx.OIRankingData != nil {
		// 	logger.Infof("📊 [%s] OI ranking data ready: %d top, %d low positions",
		// 		at.name, len(ctx.OIRankingData.TopPositions), len(ctx.OIRankingData.LowPositions))
		// }
	}

	// 10. Get NetFlow ranking data (market-wide fund flow)
	if strategyConfig.Indicators.EnableNetFlowRanking {
		// [DEBUG] NetFlow ranking fetch - Enable when debugging fund flow feature
		// logger.Infof("💰 [%s] Fetching NetFlow ranking data...", at.name)
		ctx.NetFlowRankingData = at.strategyEngine.FetchNetFlowRankingData()
		// [DEBUG] NetFlow ranking result - Enable when debugging fund flow feature
		// if ctx.NetFlowRankingData != nil {
		// 	logger.Infof("💰 [%s] NetFlow ranking data ready: inst_in=%d, inst_out=%d",
		// 		at.name, len(ctx.NetFlowRankingData.InstitutionFutureTop), len(ctx.NetFlowRankingData.InstitutionFutureLow))
		// }
	}

	// 11. Get Price ranking data (market-wide gainers/losers)
	if strategyConfig.Indicators.EnablePriceRanking {
		// [DEBUG] Price ranking fetch - Enable when debugging price ranking feature
		// logger.Infof("📈 [%s] Fetching Price ranking data...", at.name)
		ctx.PriceRankingData = at.strategyEngine.FetchPriceRankingData()
		// [DEBUG] Price ranking result - Enable when debugging price ranking feature
		// if ctx.PriceRankingData != nil {
		// 	logger.Infof("📈 [%s] Price ranking data ready for %d durations",
		// 		at.name, len(ctx.PriceRankingData.Durations))
		// }
	}

	return ctx, nil
}

// executeDecisionWithRecord executes AI decision and records detailed information
func (at *AutoTrader) executeDecisionWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	// TODO: Execute BEFORE_DECISION_EXECUTE hook (currently disabled due to missing type definition)
	// beforeHookResult := hook.HookExec[hook.DecisionExecutionResult](hook.BEFORE_DECISION_EXECUTE, decision, at.id)
	// if beforeHookResult != nil {
	// 	if err := beforeHookResult.Error(); err != nil {
	// 		logger.Errorf("❌ Before decision execute hook error: %v", err)
	// 	}
	// }

	// Log the incoming decision for debugging
	logger.Infof("  🤖 Processing AI decision: Symbol=%s, Action=%s, Confidence=%d", decision.Symbol, decision.Action, decision.Confidence)
	// [DEBUG] Decision parameters - Enable when debugging specific decision execution
	// Log additional parameters based on action type
	// switch decision.Action {
	// case "update_stop_loss":
	// 	logger.Infof("     New Stop Loss: %.4f", decision.NewStopLoss)
	// case "update_take_profit":
	// 	logger.Infof("     New Take Profit: %.4f", decision.NewTakeProfit)
	// case "partial_close":
	// 	logger.Infof("     Close Percentage: %.2f%%", decision.ClosePercentage)
	// case "open_long", "open_short":
	// 	logger.Infof("     Leverage: %d, Position Size: %.2f, Stop Loss: %.4f, Take Profit: %.4f", decision.Leverage, decision.PositionSizeUSD, decision.StopLoss, decision.TakeProfit)
	// case "trailing_stop":
	// 	logger.Infof("     Trail Percentage: %.2f%%, Activation Price: %.4f", decision.TrailPercentage, decision.ActivationPrice)
	// case "dynamic_take_profit":
	// 	logger.Infof("     Target ROI: %.2f%%, Max ROI: %.2f%%, Time Limit: %.2f hours", decision.TargetROI, decision.MaxROI, decision.TimeLimitHours)
	// }
	switch decision.Action {
	case "open_long":
		return at.executeOpenLongWithRecord(decision, actionRecord)
	case "open_short":
		return at.executeOpenShortWithRecord(decision, actionRecord)
	case "close_long":
		return at.executeCloseLongWithRecord(decision, actionRecord)
	case "close_short":
		return at.executeCloseShortWithRecord(decision, actionRecord)
	case "hold", "wait":
		// No execution needed, just record
		return nil
	case "update_stop_loss":
		return at.executeUpdateStopLossWithRecord(decision, actionRecord)
	case "update_take_profit":
		return at.executeUpdateTakeProfitWithRecord(decision, actionRecord)
	case "partial_close":
		return at.executePartialCloseWithRecord(decision, actionRecord)
	case "trailing_stop":
		return at.executeTrailingStopWithRecord(decision, actionRecord)
	case "dynamic_take_profit":
		return at.executeDynamicTakeProfitWithRecord(decision, actionRecord)
	case "oco_order":
		return at.executeOCOOrderWithRecord(decision, actionRecord)
	case "bracket_order":
		return at.executeBracketOrderWithRecord(decision, actionRecord)
	case "add_to_position":
		return at.executeAddToPositionWithRecord(decision, actionRecord)
	default:
		return fmt.Errorf("unknown action: %s", decision.Action)
	}
}

// ExecuteDecision executes a trading decision from external sources (e.g., debate consensus)
// This is a public method that can be called by other modules
func (at *AutoTrader) ExecuteDecision(d *kernel.Decision) error {
	logger.Infof("[%s] Executing external decision: %s %s", at.name, d.Action, d.Symbol)

	// Create a minimal action record for tracking
	actionRecord := &store.DecisionAction{
		Symbol:     d.Symbol,
		Action:     d.Action,
		Leverage:   d.Leverage,
		StopLoss:   d.StopLoss,
		TakeProfit: d.TakeProfit,
		Confidence: d.Confidence,
		Reasoning:  d.Reasoning,

		// Additional parameters for advanced action types
		NewStopLoss:               d.NewStopLoss,
		NewTakeProfit:             d.NewTakeProfit,
		ClosePercentage:           d.ClosePercentage,
		TrailPercentage:           d.TrailPercentage,
		CallbackRate:              d.CallbackRate,
		TargetROI:                 d.TargetROI,
		MaxROI:                    d.MaxROI,
		TimeLimitHours:            d.TimeLimitHours,
		AdditionalPositionSizeUSD: d.AdditionalPositionSizeUSD,
	}

	// Execute the decision
	err := at.executeDecisionWithRecord(d, actionRecord)
	if err != nil {
		logger.Errorf("[%s] External decision execution failed: %v", at.name, err)
		return err
	}

	logger.Infof("[%s] External decision executed successfully: %s %s", at.name, d.Action, d.Symbol)
	return nil
}

// TriggerDecision triggers a new decision cycle immediately
// GenerateFullPrompt generates the complete prompt (System + User) with real-time data
// but does NOT call the AI. Used for manual copy-paste workflow.
func (at *AutoTrader) GenerateFullPrompt(variant string) (string, string, error) {
	// 1. Collect trading context (K-lines, indicators, account, positions)
	ctx, err := at.buildTradingContext()
	if err != nil {
		return "", "", fmt.Errorf("failed to build trading context: %w", err)
	}

	if at.strategyEngine == nil {
		return "", "", fmt.Errorf("strategy engine not initialized")
	}

	// 1.5. Enhance context with additional market data for positions and candidate coins
	// This ensures AI has complete information about both held positions and potential trades
	if err := at.enhanceMarketData(ctx); err != nil {
		logger.Warnf("[%s] Warning: failed to enhance market data: %v", at.name, err)
		// Continue with existing data even if enhancement fails
	}

	// 2. Build System Prompt
	systemPrompt := at.strategyEngine.BuildSystemPrompt(ctx.Account.TotalEquity, variant)

	// 3. Build User Prompt (contains real-time market data)
	userPrompt := at.strategyEngine.BuildUserPrompt(ctx)

	return systemPrompt, userPrompt, nil
}

// enhanceMarketData enhances the context with additional market data for positions and candidate coins
// This ensures AI has complete information about both held positions and potential trades
func (at *AutoTrader) enhanceMarketData(ctx *kernel.Context) error {
	// Ensure MarketDataMap is initialized
	if ctx.MarketDataMap == nil {
		ctx.MarketDataMap = make(map[string]*market.Data)
	}

	// Get strategy configuration for timeframe settings
	strategyConfig := at.strategyEngine.GetConfig()
	timeframes := strategyConfig.Indicators.Klines.SelectedTimeframes
	primaryTimeframe := strategyConfig.Indicators.Klines.PrimaryTimeframe
	klineCount := strategyConfig.Indicators.Klines.PrimaryCount
	if klineCount <= 0 {
		klineCount = 50
	}

	// Collect all symbols that need market data
	symbolsToQuery := make(map[string]bool)

	// Add symbols from positions
	for _, pos := range ctx.Positions {
		symbolsToQuery[pos.Symbol] = true
	}

	// Add symbols from candidate coins
	for _, coin := range ctx.CandidateCoins {
		symbolsToQuery[coin.Symbol] = true
	}

	// Fetch market data for all required symbols
	for symbol := range symbolsToQuery {
		data, err := market.GetWithTimeframes(symbol, timeframes, primaryTimeframe, klineCount)
		if err != nil {
			logger.Warnf("[%s] Failed to get market data for %s: %v", at.name, symbol, err)
			// Continue with other symbols even if one fails
			continue
		}
		ctx.MarketDataMap[symbol] = data
	}

	logger.Infof("[%s] Enhanced market data for %d symbols (positions + candidates)", at.name, len(ctx.MarketDataMap))

	return nil
}

func (at *AutoTrader) TriggerDecision() (map[string]interface{}, error) {
	startTime := time.Now()
	logger.Infof("🔄 Manual trigger: Starting new decision cycle for %s", at.name)

	// Check if trader is running
	at.isRunningMutex.RLock()
	running := at.isRunning
	at.isRunningMutex.RUnlock()
	if !running {
		logger.Errorf("❌ Trader %s is not running, cannot execute manual trigger", at.name)
		return nil, fmt.Errorf("trader is not running")
	}

	logger.Infof("✅ Trader %s is running, acquiring execution mutex", at.name)

	// Acquire execution mutex to prevent concurrent executions
	at.executionMutex.Lock()
	// Check if already executing to prevent concurrent runs
	if at.isExecuting {
		at.executionMutex.Unlock()
		logger.Warnf("⚠️ Decision cycle for %s is already executing, please wait for completion", at.name)
		return nil, fmt.Errorf("decision cycle is already executing, please wait for completion")
	}
	// Mark as executing
	at.isExecuting = true
	at.executionMutex.Unlock()
	logger.Infof("🔒 Execution mutex acquired for %s, starting runCycle", at.name)

	// Ensure we reset the executing flag when done
	defer func() {
		logger.Infof("🔓 Releasing execution mutex for %s", at.name)
		at.executionMutex.Lock()
		at.isExecuting = false
		at.executionMutex.Unlock()
	}()

	logger.Infof("🚀 Calling runCycle for manual trigger on %s", at.name)
	// 🔥 记录手动扫描时间（用于延迟系统扫描）
	at.lastManualScanTime = time.Now()

	// Call the main decision cycle - let it run to completion
	err := at.runCycle()
	executionTime := time.Since(startTime)
	if err != nil {
		logger.Errorf("❌ Manual trigger decision cycle failed for %s: %v (took %v)", at.name, err, executionTime)
		return nil, err
	}

	logger.Infof("✅ Preparing result for %s after successful runCycle (took %v)", at.name, executionTime)

	// 🔥 更新下次系统扫描时间（手动扫描后延迟）
	at.nextSystemScanTime = at.lastManualScanTime.Add(at.scanDelayDuration).Add(at.config.ScanInterval)
	logger.Infof("⏰ [%s] 下次系统扫描时间已更新为: %s (手动扫描后延迟%.0f秒)",
		at.name, at.nextSystemScanTime.Format("15:04:05"), at.scanDelayDuration.Seconds())

	// Return success status and some info
	result := map[string]interface{}{
		"success":                  true,
		"timestamp":                time.Now().Unix(),
		"cycle_number":             at.callCount,
		"execution_time_ms":        executionTime.Milliseconds(),
		"execution_time_formatted": executionTime.String(),
		"message":                  "Decision cycle completed successfully",
	}

	logger.Infof("✅ Manual trigger decision cycle completed for %s (took %v)", at.name, executionTime)
	return result, nil
}

// runCycleWithExecutionLock runs a decision cycle with execution lock to prevent conflicts with manual scans
func (at *AutoTrader) runCycleWithExecutionLock() error {
	startTime := time.Now()
	// Acquire execution mutex to prevent concurrent executions
	at.executionMutex.Lock()
	// Check if already executing to prevent concurrent runs
	if at.isExecuting {
		at.executionMutex.Unlock()
		logger.Infof("⚠️ Automatic scan skipped: manual scan in progress for %s", at.name)
		return fmt.Errorf("automatic scan skipped: manual scan in progress")
	}
	// Mark as executing
	at.isExecuting = true
	at.executionMutex.Unlock()

	// Ensure we reset the executing flag when done
	defer func() {
		at.executionMutex.Lock()
		at.isExecuting = false
		at.executionMutex.Unlock()
	}()

	// Call the main decision cycle
	err := at.runCycle()
	executionTime := time.Since(startTime)
	if err != nil {
		logger.Errorf("Automatic scan cycle failed for %s: %v (took %v)", at.name, err, executionTime)
	} else {
		logger.Infof("Automatic scan cycle completed for %s (took %v)", at.name, executionTime)
	}
	return err
}

// executeOpenLongWithRecord executes open long position and records detailed information
func (at *AutoTrader) executeOpenLongWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  📈 Open long: %s", decision.Symbol)

	// [CODE ENFORCED] Check trade frequency limits
	if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
		return err
	}

	// ⚠️ Get current positions for multiple checks
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// [CODE ENFORCED] Check max positions limit
	if err := at.enforceMaxPositions(len(positions)); err != nil {
		return err
	}

	// Check if there's already a position in the same symbol and direction
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol && pos["side"] == "long" {
			return fmt.Errorf("❌ [POSITION_EXISTS_LIMIT] %s already has long position - Close existing position first", decision.Symbol)
		}
	}

	// Get current price
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// Get balance (needed for multiple checks)
	balance, err := at.trader.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}
	availableBalance := 0.0
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// [CODE ENFORCED] Check daily loss limit before opening new positions
	if err := at.enforceDailyLossLimit(); err != nil {
		return err
	}

	// [CODE ENFORCED] Check current margin usage
	marginUsedPct := 0.0
	if marginUsed, ok := balance["marginUsed"].(float64); ok && marginUsed > 0 {
		if totalEquity, ok := balance["totalEquity"].(float64); ok && totalEquity > 0 {
			marginUsedPct = (marginUsed / totalEquity) * 100
		}
	}
	if err := at.enforceMaxMarginUsage(marginUsedPct); err != nil {
		return err
	}

	// Get equity for position value ratio check
	equity := 0.0
	if eq, ok := balance["totalEquity"].(float64); ok && eq > 0 {
		equity = eq
	} else if eq, ok := balance["totalWalletBalance"].(float64); ok && eq > 0 {
		equity = eq
	} else {
		equity = availableBalance // Fallback to available balance
	}

	// [CODE ENFORCED] Position Value Ratio Check: position_value <= equity × ratio
	adjustedPositionSize, wasCapped := at.enforcePositionValueRatio(decision.PositionSizeUSD, equity, decision.Symbol)
	if wasCapped {
		decision.PositionSizeUSD = adjustedPositionSize
	}

	// ⚠️ Auto-adjust position size if insufficient margin
	// Formula: totalRequired = positionSize/leverage + positionSize*0.001 + positionSize/leverage*0.01
	//        = positionSize * (1.01/leverage + 0.001)
	marginFactor := 1.01/float64(decision.Leverage) + 0.001
	maxAffordablePositionSize := availableBalance / marginFactor

	actualPositionSize := decision.PositionSizeUSD
	if actualPositionSize > maxAffordablePositionSize {
		// Use 98% of max to leave buffer for price fluctuation
		adjustedSize := maxAffordablePositionSize * 0.98
		logger.Infof("  ⚠️ Position size %.2f exceeds max affordable %.2f, auto-reducing to %.2f",
			actualPositionSize, maxAffordablePositionSize, adjustedSize)
		actualPositionSize = adjustedSize
		decision.PositionSizeUSD = actualPositionSize
	}

	// [CODE ENFORCED] Minimum position size check
	if err := at.enforceMinPositionSize(decision.PositionSizeUSD); err != nil {
		return err
	}

	// Calculate quantity with adjusted position size
	quantity := actualPositionSize / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// Set margin mode
	if err := at.trader.SetMarginMode(decision.Symbol, at.config.IsCrossMargin); err != nil {
		logger.Infof("  ⚠️ Failed to set margin mode: %v", err)
		// Continue execution, doesn't affect trading
	}

	// Open position
	order, err := at.trader.OpenLong(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// Record order ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	logger.Infof("  ✓ Position opened successfully, order ID: %v, quantity: %.4f", order["orderId"], quantity)

	// Record order to database and poll for confirmation
	at.recordAndConfirmOrder(order, decision.Symbol, "open_long", quantity, marketData.CurrentPrice, decision.Leverage, 0)

	// Record position opening time
	posKey := decision.Symbol + "_long"
	at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()

	// Set stop loss and take profit
	if decision.StopLoss > 0 {
		if err := at.trader.SetStopLoss(decision.Symbol, "LONG", quantity, decision.StopLoss); err != nil {
			logger.Errorf("  ❌ Failed to set stop loss for %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to set stop loss: %w", err)
		}
		logger.Infof("  ✓ Stop loss set for %s: %.4f", decision.Symbol, decision.StopLoss)
	}
	if decision.TakeProfit > 0 {
		if err := at.trader.SetTakeProfit(decision.Symbol, "LONG", quantity, decision.TakeProfit); err != nil {
			logger.Errorf("  ❌ Failed to set take profit for %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to set take profit: %w", err)
		}
		logger.Infof("  ✓ Take profit set for %s: %.4f", decision.Symbol, decision.TakeProfit)
	}

	return nil
}

// executeOpenShortWithRecord executes open short position and records detailed information
func (at *AutoTrader) executeOpenShortWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  📉 Open short: %s", decision.Symbol)

	// [CODE ENFORCED] Check trade frequency limits
	if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
		return err
	}

	// ⚠️ Get current positions for multiple checks
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// [CODE ENFORCED] Check max positions limit
	if err := at.enforceMaxPositions(len(positions)); err != nil {
		return err
	}

	// Check if there's already a position in the same symbol and direction
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol && pos["side"] == "short" {
			return fmt.Errorf("❌ [POSITION_EXISTS_LIMIT] %s already has short position - Close existing position first", decision.Symbol)
		}
	}

	// Get current price
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}

	// Get balance (needed for multiple checks)
	balance, err := at.trader.GetBalance()
	if err != nil {
		return fmt.Errorf("failed to get account balance: %w", err)
	}
	availableBalance := 0.0
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// [CODE ENFORCED] Check daily loss limit before opening new positions
	if err := at.enforceDailyLossLimit(); err != nil {
		return err
	}

	// [CODE ENFORCED] Check current margin usage
	marginUsedPct := 0.0
	if marginUsed, ok := balance["marginUsed"].(float64); ok && marginUsed > 0 {
		if totalEquity, ok := balance["totalEquity"].(float64); ok && totalEquity > 0 {
			marginUsedPct = (marginUsed / totalEquity) * 100
		}
	}
	if err := at.enforceMaxMarginUsage(marginUsedPct); err != nil {
		return err
	}

	// Get equity for position value ratio check
	equity := 0.0
	if eq, ok := balance["totalEquity"].(float64); ok && eq > 0 {
		equity = eq
	} else if eq, ok := balance["totalWalletBalance"].(float64); ok && eq > 0 {
		equity = eq
	} else {
		equity = availableBalance // Fallback to available balance
	}

	// [CODE ENFORCED] Position Value Ratio Check: position_value <= equity × ratio
	adjustedPositionSize, wasCapped := at.enforcePositionValueRatio(decision.PositionSizeUSD, equity, decision.Symbol)
	if wasCapped {
		decision.PositionSizeUSD = adjustedPositionSize
	}

	// ⚠️ Auto-adjust position size if insufficient margin
	// Formula: totalRequired = positionSize/leverage + positionSize*0.001 + positionSize/leverage*0.01
	//        = positionSize * (1.01/leverage + 0.001)
	marginFactor := 1.01/float64(decision.Leverage) + 0.001
	maxAffordablePositionSize := availableBalance / marginFactor

	actualPositionSize := decision.PositionSizeUSD
	if actualPositionSize > maxAffordablePositionSize {
		// Use 98% of max to leave buffer for price fluctuation
		adjustedSize := maxAffordablePositionSize * 0.98
		logger.Infof("  ⚠️ Position size %.2f exceeds max affordable %.2f, auto-reducing to %.2f",
			actualPositionSize, maxAffordablePositionSize, adjustedSize)
		actualPositionSize = adjustedSize
		decision.PositionSizeUSD = actualPositionSize
	}

	// [CODE ENFORCED] Minimum position size check
	if err := at.enforceMinPositionSize(decision.PositionSizeUSD); err != nil {
		return err
	}

	// Calculate quantity with adjusted position size
	quantity := actualPositionSize / marketData.CurrentPrice
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice

	// Set margin mode
	if err := at.trader.SetMarginMode(decision.Symbol, at.config.IsCrossMargin); err != nil {
		logger.Infof("  ⚠️ Failed to set margin mode: %v", err)
		// Continue execution, doesn't affect trading
	}

	// Open position
	order, err := at.trader.OpenShort(decision.Symbol, quantity, decision.Leverage)
	if err != nil {
		return err
	}

	// Record order ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	logger.Infof("  ✓ Position opened successfully, order ID: %v, quantity: %.4f", order["orderId"], quantity)

	// Record order to database and poll for confirmation
	at.recordAndConfirmOrder(order, decision.Symbol, "open_short", quantity, marketData.CurrentPrice, decision.Leverage, 0)

	// Record position opening time
	posKey := decision.Symbol + "_short"
	at.positionFirstSeenTime[posKey] = time.Now().UnixMilli()

	// Set stop loss and take profit
	if decision.StopLoss > 0 {
		if err := at.trader.SetStopLoss(decision.Symbol, "SHORT", quantity, decision.StopLoss); err != nil {
			logger.Errorf("  ❌ Failed to set stop loss for %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to set stop loss: %w", err)
		}
		logger.Infof("  ✓ Stop loss set for %s: %.4f", decision.Symbol, decision.StopLoss)
	}
	if decision.TakeProfit > 0 {
		if err := at.trader.SetTakeProfit(decision.Symbol, "SHORT", quantity, decision.TakeProfit); err != nil {
			logger.Errorf("  ❌ Failed to set take profit for %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to set take profit: %w", err)
		}
		logger.Infof("  ✓ Take profit set for %s: %.4f", decision.Symbol, decision.TakeProfit)
	}

	return nil
}

// executeCloseLongWithRecord executes close long position and records detailed information
func (at *AutoTrader) executeCloseLongWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  ✖️ Close long: %s", decision.Symbol)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current price
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}
	actionRecord.Price = marketData.CurrentPrice

	// [CODE ENFORCED] Check minimum hold time before closing position
	if err := at.enforceMinHoldTime(decision.Symbol, "long"); err != nil {
		return err
	}

	// Normalize symbol for database lookup
	normalizedSymbol := market.Normalize(decision.Symbol)

	// Get entry price and quantity - prioritize local database for accurate quantity
	var entryPrice float64
	var quantity float64

	// First try to get from local database (more accurate for quantity)
	if at.store != nil {
		if openPos, err := at.store.Position().GetOpenPositionBySymbol(at.id, normalizedSymbol, "LONG"); err == nil && openPos != nil {
			quantity = openPos.Quantity
			entryPrice = openPos.EntryPrice
			logger.Infof("  📊 Using local position data: qty=%.8f, entry=%.2f", quantity, entryPrice)
		}
	}

	// Fallback to exchange API if local data not found
	if quantity == 0 {
		positions, err := at.trader.GetPositions()
		if err == nil {
			for _, pos := range positions {
				if pos["symbol"] == decision.Symbol && pos["side"] == "long" {
					if ep, ok := pos["entryPrice"].(float64); ok {
						entryPrice = ep
					}
					quantityTemp, ok := pos["positionAmt"].(float64)
					if !ok {
						// Try to get as json.Number if it fails as float64
						quantityNum, ok2 := pos["positionAmt"].(*json.Number)
						if ok2 {
							var err error
							quantityTemp, err = quantityNum.Float64()
							if err != nil {
								logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
								quantityTemp = 0
							}
						}
					}
					if quantityTemp > 0 {
						quantity = quantityTemp
					}
					break
				}
			}
		}
		logger.Infof("  📊 Using exchange position data: qty=%.8f, entry=%.2f", quantity, entryPrice)
	}

	// Close position
	order, err := at.trader.CloseLong(decision.Symbol, 0) // 0 = close all
	if err != nil {
		return err
	}

	// After closing position, ensure all related pending orders are cancelled
	if err := at.trader.CancelAllOrders(decision.Symbol); err != nil {
		logger.Infof("  ⚠️ Failed to cancel pending orders after closing long position: %v", err)
	} else {
		logger.Infof("  ✓ Cancelled all pending orders after closing long position for %s", decision.Symbol)
	}

	// Record order ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	// Record order to database and poll for confirmation
	at.recordAndConfirmOrder(order, decision.Symbol, "close_long", quantity, marketData.CurrentPrice, 0, entryPrice)

	logger.Infof("  ✓ Position closed successfully")
	return nil
}

// executeCloseShortWithRecord executes close short position and records detailed information
func (at *AutoTrader) executeCloseShortWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  ✖️ Close short: %s", decision.Symbol)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current price
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		return err
	}
	actionRecord.Price = marketData.CurrentPrice

	// [CODE ENFORCED] Check minimum hold time before closing position
	if err := at.enforceMinHoldTime(decision.Symbol, "short"); err != nil {
		return err
	}

	// Normalize symbol for database lookup
	normalizedSymbol := market.Normalize(decision.Symbol)

	// Get entry price and quantity - prioritize local database for accurate quantity
	var entryPrice float64
	var quantity float64

	// First try to get from local database (more accurate for quantity)
	if at.store != nil {
		if openPos, err := at.store.Position().GetOpenPositionBySymbol(at.id, normalizedSymbol, "SHORT"); err == nil && openPos != nil {
			quantity = openPos.Quantity
			entryPrice = openPos.EntryPrice
			logger.Infof("  📊 Using local position data: qty=%.8f, entry=%.2f", quantity, entryPrice)
		}
	}

	// Fallback to exchange API if local data not found
	if quantity == 0 {
		positions, err := at.trader.GetPositions()
		if err == nil {
			for _, pos := range positions {
				if pos["symbol"] == decision.Symbol && pos["side"] == "short" {
					if ep, ok := pos["entryPrice"].(float64); ok {
						entryPrice = ep
					}
					quantityTemp, ok := pos["positionAmt"].(float64)
					if !ok {
						// Try to get as json.Number if it fails as float64
						quantityNum, ok2 := pos["positionAmt"].(*json.Number)
						if ok2 {
							var err error
							quantityTemp, err = quantityNum.Float64()
							if err != nil {
								logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
								quantityTemp = 0
							}
						}
					}
					if quantityTemp != 0 {
						quantity = -quantityTemp // positionAmt is negative for short
					}
					break
				}
			}
		}
		logger.Infof("  📊 Using exchange position data: qty=%.8f, entry=%.2f", quantity, entryPrice)
	}

	// Close position
	order, err := at.trader.CloseShort(decision.Symbol, 0) // 0 = close all
	if err != nil {
		return err
	}

	// After closing position, ensure all related pending orders are cancelled
	if err := at.trader.CancelAllOrders(decision.Symbol); err != nil {
		logger.Infof("  ⚠️ Failed to cancel pending orders after closing short position: %v", err)
	} else {
		logger.Infof("  ✓ Cancelled all pending orders after closing short position for %s", decision.Symbol)
	}

	// Record order ID
	if orderID, ok := order["orderId"].(int64); ok {
		actionRecord.OrderID = orderID
	}

	// Record order to database and poll for confirmation
	at.recordAndConfirmOrder(order, decision.Symbol, "close_short", quantity, marketData.CurrentPrice, 0, entryPrice)

	logger.Infof("  ✓ Position closed successfully")
	return nil
}

// executeTrailingStopWithRecord executes trailing stop and records detailed information
func (at *AutoTrader) executeTrailingStopWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Setting trailing stop: %s, Trail %%: %.2f%%, Activation Price: %.4f",
		decision.Symbol, decision.TrailPercentage, decision.ActivationPrice)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	if foundPos == nil {
		return fmt.Errorf("no position found for symbol %s", decision.Symbol)
	}

	// Get the current position quantity
	qtyFloat, ok := foundPos["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		quantity, ok2 := foundPos["positionAmt"].(*json.Number)
		if !ok2 {
			return fmt.Errorf("failed to get position amount")
		}
		var err error
		qtyFloat, err = quantity.Float64()
		if err != nil {
			return fmt.Errorf("failed to convert quantity to float: %w", err)
		}
	}

	// Convert negative quantity to positive if needed
	if qtyFloat < 0 {
		qtyFloat = math.Abs(qtyFloat)
	}

	// Get position side
	positionSide, ok := foundPos["positionSide"].(string)
	if !ok {
		// Some exchanges use 'side' instead of 'positionSide'
		positionSide, _ = foundPos["side"].(string)
	}

	// Determine side for trailing stop
	side := "LONG"
	if positionSide == "SHORT" || (positionSide == "" && strings.Contains(strings.ToUpper(foundPos["symbol"].(string)), "USDT") && foundPos["side"].(string) == "SHORT") {
		side = "SHORT"
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
	}

	// Use current market price as activation price if not provided in decision
	activationPrice := decision.ActivationPrice
	if activationPrice == 0 && marketData != nil {
		activationPrice = marketData.CurrentPrice
		logger.Infof("  💡 Using current market price as activation price: %.4f", activationPrice)
	}

	// Debug log for exchange API call
	logger.Infof("  📡 Attempting to submit trailing stop to exchange: Symbol=%s, Side=%s, TrailPercentage=%.4f, ActivationPrice=%.4f",
		decision.Symbol, side, decision.TrailPercentage, activationPrice)

	// Submit trailing stop order to exchange
	logger.Infof("  📡 Submitting trailing stop to exchange: Symbol=%s, Side=%s, CallbackRate=%.4f, ActivationPrice=%.4f",
		decision.Symbol, side, decision.CallbackRate, activationPrice)

	err = at.trader.SetTrailingStop(decision.Symbol, side, qtyFloat, decision.CallbackRate, activationPrice)
	if err != nil {
		logger.Errorf("  ❌ Failed to set trailing stop for %s: %v", decision.Symbol, err)
		return fmt.Errorf("failed to set trailing stop: %w", err)
	}

	logger.Infof("  ✓ Trailing stop set successfully for %s, Callback Rate: %.2f%%, Activation Price: %.4f",
		decision.Symbol, decision.CallbackRate*100, decision.ActivationPrice)

	// Record the trailing stop action to database
	if at.store != nil {
		orderID := fmt.Sprintf("TS_%s_%d", decision.Symbol, time.Now().Unix())
		// Record the trailing stop as an action
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    side,
			OrderAction:     "trailing_stop",
			Type:            "TRAILING_STOP_MARKET", // Or TRAILING_STOP_LIMIT depending on implementation
			Side:            "TRAILING_STOP",
			Quantity:        qtyFloat,
			Price:           decision.ActivationPrice, // Activation price for trailing stop
			Status:          "ACTIVE",                 // Status indicating the trailing stop is active
			FilledQuantity:  0,                        // Not filled yet, just activated
			AvgFillPrice:    0,                        // Will be filled when triggered
			Commission:      0,                        // No commission for trailing stop setup
			FilledAt:        0,                        // Will be set when triggered
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		// Add market price at time of activation for reference
		if marketData != nil {
			orderRecord.AvgFillPrice = marketData.CurrentPrice // Current market price at activation time
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record trailing stop: %v", err)
		} else {
			logger.Infof("  📊 Trailing stop recorded: %s trail %%: %.2f%%, activation price: %.4f, current price: %.4f",
				decision.Symbol, decision.TrailPercentage, decision.ActivationPrice, orderRecord.AvgFillPrice)
		}
	}

	return nil
}

// executeDynamicTakeProfitWithRecord executes dynamic take profit and records detailed information
func (at *AutoTrader) executeDynamicTakeProfitWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Setting dynamic take profit: %s, Target ROI: %.2f%%, Max ROI: %.2f%%, Time Limit: %.2f hours",
		decision.Symbol, decision.TargetROI, decision.MaxROI, decision.TimeLimitHours)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	if foundPos == nil {
		return fmt.Errorf("no position found for symbol %s", decision.Symbol)
	}

	// Get the current position quantity
	qtyFloat, ok := foundPos["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		quantity, ok2 := foundPos["positionAmt"].(*json.Number)
		if !ok2 {
			return fmt.Errorf("failed to get position amount")
		}
		var err error
		qtyFloat, err = quantity.Float64()
		if err != nil {
			return fmt.Errorf("failed to convert quantity to float: %w", err)
		}
	}

	// Convert negative quantity to positive if needed
	if qtyFloat < 0 {
		qtyFloat = math.Abs(qtyFloat)
	}

	// Get position side
	positionSide, ok := foundPos["positionSide"].(string)
	if !ok {
		// Some exchanges use 'side' instead of 'positionSide'
		positionSide, _ = foundPos["side"].(string)
	}

	// Determine side for dynamic take profit
	side := "LONG"
	if positionSide == "SHORT" || (positionSide == "" && strings.Contains(strings.ToUpper(foundPos["symbol"].(string)), "USDT") && foundPos["side"].(string) == "SHORT") {
		side = "SHORT"
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
	}

	// Debug log for exchange API call
	logger.Infof("  📡 Submitting dynamic take profit to exchange: Symbol=%s, Side=%s, TargetROI=%.4f, MaxROI=%.4f, TimeLimitHours=%.4f",
		decision.Symbol, side, decision.TargetROI, decision.MaxROI, decision.TimeLimitHours)
	// Set dynamic take profit - this would be a custom implementation
	// Since exchanges don't directly support dynamic take profit with ROI targets,
	// we would need to implement this logic in our system
	err = at.setDynamicTakeProfit(decision.Symbol, side, qtyFloat, decision.TargetROI, decision.MaxROI, decision.TimeLimitHours)
	if err != nil {
		logger.Errorf("  ❌ Failed to set dynamic take profit for %s: %v", decision.Symbol, err)
		return fmt.Errorf("failed to set dynamic take profit: %w", err)
	}

	logger.Infof("  ✓ Dynamic take profit set successfully for %s, Target ROI: %.2f%%, Max ROI: %.2f%%, Time Limit: %.2f hours",
		decision.Symbol, decision.TargetROI, decision.MaxROI, decision.TimeLimitHours)

	// Record the dynamic take profit action to database
	if at.store != nil {
		orderID := fmt.Sprintf("DTP_%s_%d", decision.Symbol, time.Now().Unix())
		// Record the dynamic take profit as an action
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    side,
			OrderAction:     "dynamic_take_profit",
			Type:            "DYNAMIC_TAKE_PROFIT", // Custom order type
			Side:            "TAKE_PROFIT_DYNAMIC",
			Quantity:        qtyFloat,
			Price:           0,        // Not applicable for dynamic take profit based on ROI
			Status:          "ACTIVE", // Status indicating the dynamic take profit is active
			FilledQuantity:  0,        // Not filled yet, just activated
			AvgFillPrice:    0,        // Will be filled when triggered
			Commission:      0,        // No commission for dynamic take profit setup
			FilledAt:        0,        // Will be set when triggered
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		// Add market price at time of activation for reference
		if marketData != nil {
			orderRecord.AvgFillPrice = marketData.CurrentPrice // Current market price at activation time
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record dynamic take profit: %v", err)
		} else {
			logger.Infof("  📊 Dynamic take profit recorded: %s target ROI: %.2f%%, current price: %.4f",
				decision.Symbol, decision.TargetROI, orderRecord.AvgFillPrice)
		}
	}

	return nil
}

// setDynamicTakeProfit implements custom dynamic take profit logic based on ROI targets
func (at *AutoTrader) setDynamicTakeProfit(symbol, side string, quantity, targetROI, maxROI, timeLimitHours float64) error {
	// This is a placeholder implementation - in a real system, you'd need to implement
	// custom logic that monitors the position and closes it when ROI targets are met
	logger.Infof("  📈 Implementing dynamic take profit logic for %s: targetROI=%.2f%%, maxROI=%.2f%%, timeLimit=%.2f hours",
		symbol, targetROI, maxROI, timeLimitHours)

	// In a real implementation, you'd start a goroutine that monitors the position
	// and executes the take profit when conditions are met
	// For now, we'll simulate this by setting regular take profit based on target ROI
	// First get current market price to calculate target price
	marketData, err := market.Get(symbol)
	if err != nil {
		logger.Errorf("  ❌ Failed to get market data for dynamic take profit for %s: %v", symbol, err)
		return fmt.Errorf("failed to get market data: %w", err)
	}

	// Calculate target price based on ROI
	var targetPrice float64
	if side == "LONG" {
		// For long positions: targetPrice = currentPrice * (1 + targetROI/100)
		targetPrice = marketData.CurrentPrice * (1 + targetROI/100)
	} else {
		// For short positions: targetPrice = currentPrice * (1 - targetROI/100)
		targetPrice = marketData.CurrentPrice * (1 - targetROI/100)
	}

	// Set the take profit order at the calculated target price
	if err := at.trader.SetTakeProfit(symbol, side, quantity, targetPrice); err != nil {
		logger.Errorf("  ❌ Failed to set take profit for dynamic take profit for %s: %v", symbol, err)
		return fmt.Errorf("failed to set take profit: %w", err)
	}

	logger.Infof("  ✓ Dynamic take profit set for %s at price %.4f (based on %.2f%% ROI)", symbol, targetPrice, targetROI)
	return nil
}

// GetID gets trader ID
func (at *AutoTrader) GetID() string {
	return at.id
}

// GetName gets trader name
func (at *AutoTrader) GetName() string {
	return at.name
}

// GetAIModel gets AI model
func (at *AutoTrader) GetAIModel() string {
	return at.aiModel
}

// GetExchange gets exchange
func (at *AutoTrader) GetExchange() string {
	return at.exchange
}

// GetShowInCompetition returns whether trader should be shown in competition
func (at *AutoTrader) GetShowInCompetition() bool {
	return at.showInCompetition
}

// SetShowInCompetition sets whether trader should be shown in competition
func (at *AutoTrader) SetShowInCompetition(show bool) {
	at.showInCompetition = show
}

// SetCustomPrompt sets custom trading strategy prompt
func (at *AutoTrader) SetCustomPrompt(prompt string) {
	at.customPrompt = prompt
}

// SetOverrideBasePrompt sets whether to override base prompt
func (at *AutoTrader) SetOverrideBasePrompt(override bool) {
	at.overrideBasePrompt = override
}

// GetSystemPromptTemplate gets current system prompt template name (from strategy config)
func (at *AutoTrader) GetSystemPromptTemplate() string {
	if at.strategyEngine != nil {
		config := at.strategyEngine.GetConfig()
		if config.CustomPrompt != "" {
			return "custom"
		}
	}
	return "strategy"
}

// saveEquitySnapshot saves equity snapshot independently (for drawing profit curve, decoupled from AI decision)
func (at *AutoTrader) saveEquitySnapshot(ctx *kernel.Context) {
	if at.store == nil || ctx == nil {
		return
	}

	snapshot := &store.EquitySnapshot{
		TraderID:      at.id,
		Timestamp:     time.Now().UTC(),
		TotalEquity:   ctx.Account.TotalEquity,
		Balance:       ctx.Account.TotalEquity - ctx.Account.UnrealizedPnL,
		UnrealizedPnL: ctx.Account.UnrealizedPnL,
		PositionCount: ctx.Account.PositionCount,
		MarginUsedPct: ctx.Account.MarginUsedPct,
	}

	if err := at.store.Equity().Save(snapshot); err != nil {
		logger.Infof("⚠️ Failed to save equity snapshot: %v", err)
	}
}

// saveDecision saves AI decision log to database (only records AI input/output, for debugging)
func (at *AutoTrader) saveDecision(record *store.DecisionRecord) error {
	if at.store == nil {
		return nil
	}

	at.cycleNumber++
	record.CycleNumber = at.cycleNumber
	record.TraderID = at.id

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}

	if err := at.store.Decision().LogDecision(record); err != nil {
		logger.Infof("⚠️ Failed to save decision record: %v", err)
		return err
	}

	logger.Infof("📝 Decision record saved: trader=%s, cycle=%d", at.id, at.cycleNumber)
	return nil
}

// determineStopReason analyzes the current state to determine the specific reason for trader stop
func (at *AutoTrader) determineStopReason() string {
	// Check risk control pause first (highest priority)
	if time.Now().Before(at.stopUntil) {
		return "RISK_CONTROL_AUTO_PAUSE"
	}

	// For manual scans, we keep it simple - if it's not risk control pause,
	// we assume it's user manual stop since this is the final task
	return "USER_MANUAL_STOP"
}

// GetStore gets data store (for external access to decision records, etc.)
func (at *AutoTrader) GetStore() *store.Store {
	return at.store
}

// GetStatus gets system status (for API)
func (at *AutoTrader) GetStatus() map[string]interface{} {
	aiProvider := "DeepSeek"
	if at.config.UseQwen {
		aiProvider = "Qwen"
	}

	at.isRunningMutex.RLock()
	isRunning := at.isRunning
	at.isRunningMutex.RUnlock()

	// 🔥 计算是否被手动扫描延迟
	isDelayedByManual := false
	secondsUntilNextScan := 0
	if isRunning {
		now := time.Now()
		if !at.lastManualScanTime.IsZero() && now.Before(at.lastManualScanTime.Add(at.scanDelayDuration)) {
			isDelayedByManual = true
		}
		// 计算距离下次扫描的秒数
		if at.nextSystemScanTime.After(now) {
			secondsUntilNextScan = int(at.nextSystemScanTime.Sub(now).Seconds())
		}
	}

	return map[string]interface{}{
		"trader_id":               at.id,
		"trader_name":             at.name,
		"ai_model":                at.aiModel,
		"exchange":                at.exchange,
		"is_running":              isRunning,
		"is_executing":            at.isExecuting, // 🔥 系统扫描状态
		"start_time":              at.startTime.Format(time.RFC3339),
		"runtime_minutes":         int(time.Since(at.startTime).Minutes()),
		"call_count":              at.callCount,
		"initial_balance":         at.initialBalance,
		"scan_interval":           at.config.ScanInterval.String(),
		"stop_until":              at.stopUntil.Format(time.RFC3339),
		"last_reset_time":         at.lastResetTime.Format(time.RFC3339),
		"ai_provider":             aiProvider,
		"is_delayed_by_manual":    isDelayedByManual,            // 🔥 新增：是否因手动扫描而延迟
		"next_scan_time":          at.nextSystemScanTime.Unix(), // 🔥 新增：下次扫描时间戳
		"seconds_until_next_scan": secondsUntilNextScan,         // 🔥 新增：距离下次扫描的秒数
	}
}

// GetAccountInfo gets account information (for API)
func (at *AutoTrader) GetAccountInfo() (map[string]interface{}, error) {
	balance, err := at.trader.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Get account fields
	totalWalletBalance := 0.0
	totalUnrealizedProfit := 0.0
	availableBalance := 0.0
	totalEquity := 0.0

	if wallet, ok := balance["totalWalletBalance"].(float64); ok {
		totalWalletBalance = wallet
	}
	if unrealized, ok := balance["totalUnrealizedProfit"].(float64); ok {
		totalUnrealizedProfit = unrealized
	}
	if avail, ok := balance["availableBalance"].(float64); ok {
		availableBalance = avail
	}

	// Use totalEquity directly if provided by trader (more accurate)
	if eq, ok := balance["totalEquity"].(float64); ok && eq > 0 {
		totalEquity = eq
	} else {
		// Fallback: Total Equity = Wallet balance + Unrealized profit
		totalEquity = totalWalletBalance + totalUnrealizedProfit
	}

	// Get positions to calculate total margin
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	totalMarginUsed := 0.0
	totalUnrealizedPnLCalculated := 0.0
	for _, pos := range positions {
		markPrice := pos["markPrice"].(float64)
		quantity, ok := pos["positionAmt"].(float64)
		if !ok {
			// Try to get as json.Number if it fails as float64
			quantityNum, ok2 := pos["positionAmt"].(*json.Number)
			if !ok2 {
				logger.Warnf("Failed to get position amount for %s, using 0", pos["symbol"])
				quantity = 0
			} else {
				var err error
				quantity, err = quantityNum.Float64()
				if err != nil {
					logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
					quantity = 0
				}
			}
		}
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		totalUnrealizedPnLCalculated += unrealizedPnl

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}
		marginUsed := (quantity * markPrice) / float64(leverage)
		totalMarginUsed += marginUsed
	}

	// Verify unrealized P&L consistency (API value vs calculated from positions)
	// Note: Lighter API may return 0 for unrealized PnL, this is a known limitation
	diff := math.Abs(totalUnrealizedProfit - totalUnrealizedPnLCalculated)
	if diff > 5.0 { // Only warn if difference is significant (> 5 USDT)
		logger.Infof("⚠️ Unrealized P&L inconsistency (Lighter API limitation): API=%.4f, Calculated=%.4f, Diff=%.4f",
			totalUnrealizedProfit, totalUnrealizedPnLCalculated, diff)
	}

	totalPnL := totalEquity - at.initialBalance
	totalPnLPct := 0.0
	if at.initialBalance > 0 {
		totalPnLPct = (totalPnL / at.initialBalance) * 100
	} else {
		logger.Infof("⚠️ Initial Balance abnormal: %.2f, cannot calculate P&L percentage", at.initialBalance)
	}

	marginUsedPct := 0.0
	if totalEquity > 0 {
		marginUsedPct = (totalMarginUsed / totalEquity) * 100
	}

	return map[string]interface{}{
		// Core fields
		"total_equity":      totalEquity,           // Account equity = wallet + unrealized
		"wallet_balance":    totalWalletBalance,    // Wallet balance (excluding unrealized P&L)
		"unrealized_profit": totalUnrealizedProfit, // Unrealized P&L (official value from exchange API)
		"available_balance": availableBalance,      // Available balance

		// P&L statistics
		"total_pnl":       totalPnL,          // Total P&L = equity - initial
		"total_pnl_pct":   totalPnLPct,       // Total P&L percentage
		"initial_balance": at.initialBalance, // Initial balance
		"daily_pnl":       at.dailyPnL,       // Daily P&L

		// Position information
		"position_count":  len(positions),  // Position count
		"margin_used":     totalMarginUsed, // Margin used
		"margin_used_pct": marginUsedPct,   // Margin usage rate
	}, nil
}

// GetPositions gets position list (for API)
func (at *AutoTrader) GetPositions() ([]map[string]interface{}, error) {
	positions, err := at.trader.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity, ok := pos["positionAmt"].(float64)
		if !ok {
			// Try to get as json.Number if it fails as float64
			quantityNum, ok2 := pos["positionAmt"].(*json.Number)
			if !ok2 {
				logger.Warnf("Failed to get position amount for %s, using 0", pos["symbol"])
				quantity = 0
			} else {
				var err error
				quantity, err = quantityNum.Float64()
				if err != nil {
					logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
					quantity = 0
				}
			}
		}
		if quantity < 0 {
			quantity = -quantity
		}
		unrealizedPnl := pos["unRealizedProfit"].(float64)
		liquidationPrice := pos["liquidationPrice"].(float64)

		leverage := 10
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}

		// Calculate margin used
		marginUsed := (quantity * markPrice) / float64(leverage)

		// Calculate P&L percentage (based on margin)
		pnlPct := calculatePnLPercentage(unrealizedPnl, marginUsed)

		result = append(result, map[string]interface{}{
			"symbol":             symbol,
			"side":               side,
			"entry_price":        entryPrice,
			"mark_price":         markPrice,
			"quantity":           quantity,
			"leverage":           leverage,
			"unrealized_pnl":     unrealizedPnl,
			"unrealized_pnl_pct": pnlPct,
			"liquidation_price":  liquidationPrice,
			"margin_used":        marginUsed,
		})
	}

	return result, nil
}

// calculatePnLPercentage calculates P&L percentage (based on margin, automatically considers leverage)
// Return rate = Unrealized P&L / Margin × 100%
func calculatePnLPercentage(unrealizedPnl, marginUsed float64) float64 {
	if marginUsed > 0 {
		return (unrealizedPnl / marginUsed) * 100
	}
	return 0.0
}

// sortDecisionsByPriority sorts decisions: close positions first, then open positions, finally hold/wait
// This avoids position stacking overflow when changing positions
func sortDecisionsByPriority(decisions []kernel.Decision) []kernel.Decision {
	if len(decisions) <= 1 {
		return decisions
	}

	// Define priority
	getActionPriority := func(action string) int {
		switch action {
		case "close_long", "close_short":
			return 1 // Highest priority: close positions first
		case "open_long", "open_short":
			return 2 // Second priority: open positions later
		case "hold", "wait":
			return 3 // Lowest priority: wait
		default:
			return 999 // Unknown actions at the end
		}
	}

	// Copy decision list
	sorted := make([]kernel.Decision, len(decisions))
	copy(sorted, decisions)

	// Sort by priority
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if getActionPriority(sorted[i].Action) > getActionPriority(sorted[j].Action) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// startDrawdownMonitor starts drawdown monitoring
func (at *AutoTrader) startDrawdownMonitor() {
	at.monitorWg.Add(1)
	go func() {
		defer at.monitorWg.Done()

		ticker := time.NewTicker(1 * time.Minute) // Check every minute
		defer ticker.Stop()

		logger.Info("📊 Started position drawdown monitoring (check every minute)")

		for {
			select {
			case <-ticker.C:
				at.checkPositionDrawdown()
			case <-at.stopMonitorCh:
				logger.Info("⏹ Stopped position drawdown monitoring")
				return
			}
		}
	}()
}

// checkPositionDrawdown checks position drawdown situation
func (at *AutoTrader) checkPositionDrawdown() {
	// Get current positions
	positions, err := at.trader.GetPositions()
	if err != nil {
		logger.Infof("❌ Drawdown monitoring: failed to get positions: %v", err)
		return
	}

	for _, pos := range positions {
		symbol := pos["symbol"].(string)
		side := pos["side"].(string)
		entryPrice := pos["entryPrice"].(float64)
		markPrice := pos["markPrice"].(float64)
		quantity, ok := pos["positionAmt"].(float64)
		if !ok {
			// Try to get as json.Number if it fails as float64
			quantityNum, ok2 := pos["positionAmt"].(*json.Number)
			if !ok2 {
				logger.Warnf("Failed to get position amount for %s, using 0", pos["symbol"])
				quantity = 0
			} else {
				var err error
				quantity, err = quantityNum.Float64()
				if err != nil {
					logger.Warnf("Failed to convert position amount to float for %s, using 0: %v", pos["symbol"], err)
					quantity = 0
				}
			}
		}
		if quantity < 0 {
			quantity = -quantity // Short position quantity is negative, convert to positive
		}

		// Calculate current P&L percentage
		leverage := 10 // Default value
		if lev, ok := pos["leverage"].(float64); ok {
			leverage = int(lev)
		}

		var currentPnLPct float64
		if side == "long" {
			currentPnLPct = ((markPrice - entryPrice) / entryPrice) * float64(leverage) * 100
		} else {
			currentPnLPct = ((entryPrice - markPrice) / entryPrice) * float64(leverage) * 100
		}

		// Calculate position value for loss checking
		positionValue := math.Abs(markPrice * quantity)

		// Check maximum loss per trade (applies to existing positions)
		if err := at.enforceMaxLossPerTrade(symbol, currentPnLPct*positionValue/100, positionValue); err != nil {
			logger.Infof("🚨 Max loss per trade exceeded: %v", err)

			// Execute emergency close position due to excessive loss
			if closeErr := at.emergencyClosePosition(symbol, side); closeErr != nil {
				logger.Infof("❌ Max loss close position failed (%s %s): %v", symbol, side, closeErr)
			} else {
				logger.Infof("✅ Max loss close position succeeded: %s %s", symbol, side)
				// Clear cache for this position after closing
				at.ClearPeakPnLCache(symbol, side)
				continue // Skip further processing for this position since it's closed
			}
		}

		// Construct unique position identifier (distinguish long/short)
		posKey := symbol + "_" + side

		// Get historical peak profit for this position
		at.peakPnLCacheMutex.RLock()
		peakPnLPct, exists := at.peakPnLCache[posKey]
		at.peakPnLCacheMutex.RUnlock()

		if !exists {
			// If no historical peak record, use current P&L as initial value
			peakPnLPct = currentPnLPct
			at.UpdatePeakPnL(symbol, side, currentPnLPct)
		} else {
			// Update peak cache
			at.UpdatePeakPnL(symbol, side, currentPnLPct)
		}

		// Calculate drawdown (magnitude of decline from peak)
		var drawdownPct float64
		if peakPnLPct > 0 && currentPnLPct < peakPnLPct {
			drawdownPct = ((peakPnLPct - currentPnLPct) / peakPnLPct) * 100
		}

		// Check close position condition: profit > 5% and drawdown >= 40%
		if currentPnLPct > 5.0 && drawdownPct >= 40.0 {
			logger.Infof("🚨 Drawdown close position condition triggered: %s %s | Current profit: %.2f%% | Peak profit: %.2f%% | Drawdown: %.2f%%",
				symbol, side, currentPnLPct, peakPnLPct, drawdownPct)

			// Execute close position
			if err := at.emergencyClosePosition(symbol, side); err != nil {
				logger.Infof("❌ Drawdown close position failed (%s %s): %v", symbol, side, err)
			} else {
				logger.Infof("✅ Drawdown close position succeeded: %s %s", symbol, side)
				// Clear cache for this position after closing
				at.ClearPeakPnLCache(symbol, side)
			}
		} else if currentPnLPct > 5.0 {
			// Record situations close to close position condition (for debugging)
			logger.Infof("📊 Drawdown monitoring: %s %s | Profit: %.2f%% | Peak: %.2f%% | Drawdown: %.2f%%",
				symbol, side, currentPnLPct, peakPnLPct, drawdownPct)
		}
	}
}

// emergencyClosePosition emergency close position function
func (at *AutoTrader) emergencyClosePosition(symbol, side string) error {
	switch side {
	case "long":
		order, err := at.trader.CloseLong(symbol, 0) // 0 = close all
		if err != nil {
			return err
		}
		logger.Infof("✅ Emergency close long position succeeded, order ID: %v", order["orderId"])
	case "short":
		order, err := at.trader.CloseShort(symbol, 0) // 0 = close all
		if err != nil {
			return err
		}
		logger.Infof("✅ Emergency close short position succeeded, order ID: %v", order["orderId"])
	default:
		return fmt.Errorf("unknown position direction: %s", side)
	}

	return nil
}

// GetPeakPnLCache gets peak profit cache
func (at *AutoTrader) GetPeakPnLCache() map[string]float64 {
	at.peakPnLCacheMutex.RLock()
	defer at.peakPnLCacheMutex.RUnlock()

	// Return a copy of the cache
	cache := make(map[string]float64)
	for k, v := range at.peakPnLCache {
		cache[k] = v
	}
	return cache
}

// UpdatePeakPnL updates peak profit cache
func (at *AutoTrader) UpdatePeakPnL(symbol, side string, currentPnLPct float64) {
	at.peakPnLCacheMutex.Lock()
	defer at.peakPnLCacheMutex.Unlock()

	posKey := symbol + "_" + side
	if peak, exists := at.peakPnLCache[posKey]; exists {
		// Update peak (if long, take larger value; if short, currentPnLPct is negative, also compare)
		if currentPnLPct > peak {
			at.peakPnLCache[posKey] = currentPnLPct
		}
	} else {
		// First time recording
		at.peakPnLCache[posKey] = currentPnLPct
	}
}

// ClearPeakPnLCache clears peak cache for specified position
func (at *AutoTrader) ClearPeakPnLCache(symbol, side string) {
	at.peakPnLCacheMutex.Lock()
	defer at.peakPnLCacheMutex.Unlock()

	posKey := symbol + "_" + side
	delete(at.peakPnLCache, posKey)
}

// recordAndConfirmOrder polls order status for actual fill data and records position
// action: open_long, open_short, close_long, close_short
// entryPrice: entry price when closing (0 when opening)
func (at *AutoTrader) recordAndConfirmOrder(orderResult map[string]interface{}, symbol, action string, quantity float64, price float64, leverage int, entryPrice float64) {
	if at.store == nil {
		return
	}

	// Get order ID (supports multiple types)
	var orderID string
	switch v := orderResult["orderId"].(type) {
	case int64:
		orderID = fmt.Sprintf("%d", v)
	case float64:
		orderID = fmt.Sprintf("%.0f", v)
	case string:
		orderID = v
	default:
		orderID = fmt.Sprintf("%v", v)
	}

	if orderID == "" || orderID == "0" {
		logger.Infof("  ⚠️ Order ID is empty, skipping record")
		return
	}

	// Determine positionSide
	var positionSide string
	switch action {
	case "open_long", "close_long":
		positionSide = "LONG"
	case "open_short", "close_short":
		positionSide = "SHORT"
	}

	var actualPrice = price
	var actualQty = quantity
	var fee float64

	// Exchanges with OrderSync: Skip immediate order recording, let OrderSync handle it
	// This ensures accurate data from GetTrades API and avoids duplicate records
	switch at.exchange {
	case "binance", "lighter", "hyperliquid", "bybit", "okx", "bitget", "aster":
		logger.Infof("  📝 Order submitted (id: %s), will be synced by OrderSync", orderID)
		return
	}

	// For exchanges without OrderSync (e.g., Binance): record immediately and poll for fill data
	orderRecord := at.createOrderRecord(orderID, symbol, action, positionSide, quantity, price, leverage)
	if err := at.store.Order().CreateOrder(orderRecord); err != nil {
		logger.Infof("  ⚠️ Failed to record order: %v", err)
	} else {
		logger.Infof("  📝 Order recorded: %s [%s] %s", orderID, action, symbol)
	}

	// Wait for order to be filled and get actual fill data
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 5; i++ {
		status, err := at.trader.GetOrderStatus(symbol, orderID)
		if err == nil {
			statusStr, _ := status["status"].(string)
			if statusStr == "FILLED" {
				// Get actual fill price
				if avgPrice, ok := status["avgPrice"].(float64); ok && avgPrice > 0 {
					actualPrice = avgPrice
				}
				// Get actual executed quantity
				if execQty, ok := status["executedQty"].(float64); ok && execQty > 0 {
					actualQty = execQty
				}
				// Get commission/fee
				if commission, ok := status["commission"].(float64); ok {
					fee = commission
				}
				logger.Infof("  ✅ Order filled: avgPrice=%.6f, qty=%.6f, fee=%.6f", actualPrice, actualQty, fee)

				// Update order status to FILLED
				if err := at.store.Order().UpdateOrderStatus(orderRecord.ID, "FILLED", actualQty, actualPrice, fee); err != nil {
					logger.Infof("  ⚠️ Failed to update order status: %v", err)
				}

				// Record fill details
				at.recordOrderFill(orderRecord.ID, orderID, symbol, action, actualPrice, actualQty, fee)
				break
			} else if statusStr == "CANCELED" || statusStr == "EXPIRED" || statusStr == "REJECTED" {
				logger.Infof("  ⚠️ Order %s, skipping position record", statusStr)

				// Update order status
				if err := at.store.Order().UpdateOrderStatus(orderRecord.ID, statusStr, 0, 0, 0); err != nil {
					logger.Infof("  ⚠️ Failed to update order status: %v", err)
				}
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Normalize symbol for position record consistency
	normalizedSymbolForPosition := market.Normalize(symbol)

	logger.Infof("  📝 Recording position (ID: %s, action: %s, price: %.6f, qty: %.6f, fee: %.4f)",
		orderID, action, actualPrice, actualQty, fee)

	// Record position change with actual fill data (use normalized symbol)
	at.recordPositionChange(orderID, normalizedSymbolForPosition, positionSide, action, actualQty, actualPrice, leverage, entryPrice, fee)

	// Send anonymous trade statistics for experience improvement (async, non-blocking)
	// This helps us understand overall product usage across all deployments
	experience.TrackTrade(experience.TradeEvent{
		Exchange:  at.exchange,
		TradeType: action,
		Symbol:    symbol,
		AmountUSD: actualPrice * actualQty,
		Leverage:  leverage,
		UserID:    at.userID,
		TraderID:  at.id,
	})
}

// recordPositionChange records position change (create record on open, update record on close)
func (at *AutoTrader) recordPositionChange(orderID, symbol, side, action string, quantity, price float64, leverage int, entryPrice float64, fee float64) {
	if at.store == nil || at.exchangeID == "" {
		logger.Debugf("⚠️ Skipping position record: store or exchangeID not available")
		return
	}

	// Use PositionBuilder for all position changes (consistent with OrderSync)
	posBuilder := store.NewPositionBuilder(at.store.Position())
	tradeTimeMs := time.Now().UTC().UnixMilli()

	if err := posBuilder.ProcessTrade(
		at.id, at.exchangeID, at.exchange,
		symbol, side, action,
		quantity, price, fee, 0, // realizedPnL will be calculated by PositionBuilder
		tradeTimeMs, orderID,
	); err != nil {
		logger.Infof("  ⚠️ Failed to process position change via PositionBuilder: %v", err)
	} else {
		logger.Infof("  📍 Position updated via PositionBuilder: %s (action: %s, qty: %.6f)", orderID, action, quantity)
	}
}

// createOrderRecord creates an order record struct from order details
func (at *AutoTrader) createOrderRecord(orderID, symbol, action, positionSide string, quantity, price float64, leverage int) *store.TraderOrder {
	// Determine order type (market for auto trader)
	orderType := "MARKET"

	// Determine side (BUY/SELL)
	var side string
	switch action {
	case "open_long", "close_short":
		side = "BUY"
	case "open_short", "close_long":
		side = "SELL"
	}

	// Use action as orderAction directly (keep lowercase format)
	orderAction := action

	// Determine if it's a reduce only order
	reduceOnly := (action == "close_long" || action == "close_short")

	// Normalize symbol for consistency
	normalizedSymbol := market.Normalize(symbol)

	return &store.TraderOrder{
		TraderID:        at.id,
		ExchangeID:      at.exchangeID,
		ExchangeType:    at.exchange,
		ExchangeOrderID: orderID,
		Symbol:          normalizedSymbol,
		Side:            side,
		PositionSide:    positionSide,
		Type:            orderType,
		TimeInForce:     "GTC",
		Quantity:        quantity,
		Price:           price,
		Status:          "NEW",
		FilledQuantity:  0,
		AvgFillPrice:    0,
		Commission:      0,
		CommissionAsset: "USDT",
		Leverage:        leverage,
		ReduceOnly:      reduceOnly,
		ClosePosition:   reduceOnly,
		OrderAction:     orderAction,
		CreatedAt:       time.Now().UTC().UnixMilli(),
		UpdatedAt:       time.Now().UTC().UnixMilli(),
	}
}

// recordOrderFill records order fill/trade details
func (at *AutoTrader) recordOrderFill(orderRecordID int64, exchangeOrderID, symbol, action string, price, quantity, fee float64) {
	if at.store == nil {
		return
	}

	// Determine side (BUY/SELL)
	var side string
	switch action {
	case "open_long", "close_short":
		side = "BUY"
	case "open_short", "close_long":
		side = "SELL"
	}

	// Generate a simple trade ID (exchange doesn't always provide one)
	tradeID := fmt.Sprintf("%s-%d", exchangeOrderID, time.Now().UnixNano())

	// Normalize symbol for consistency
	normalizedSymbol := market.Normalize(symbol)

	fill := &store.TraderFill{
		TraderID:        at.id,
		ExchangeID:      at.exchangeID,
		ExchangeType:    at.exchange,
		OrderID:         orderRecordID,
		ExchangeOrderID: exchangeOrderID,
		ExchangeTradeID: tradeID,
		Symbol:          normalizedSymbol,
		Side:            side,
		Price:           price,
		Quantity:        quantity,
		QuoteQuantity:   price * quantity,
		Commission:      fee,
		CommissionAsset: "USDT",
		RealizedPnL:     0,     // Will be calculated for close orders
		IsMaker:         false, // Market orders are usually taker
		CreatedAt:       time.Now().UTC().UnixMilli(),
	}

	// Calculate realized PnL for close orders
	if action == "close_long" || action == "close_short" {
		// Try to get the entry price from the open position
		var positionSide string
		if action == "close_long" {
			positionSide = "LONG"
		} else {
			positionSide = "SHORT"
		}

		if openPos, err := at.store.Position().GetOpenPositionBySymbol(at.id, symbol, positionSide); err == nil && openPos != nil {
			if positionSide == "LONG" {
				fill.RealizedPnL = (price - openPos.EntryPrice) * quantity
			} else {
				fill.RealizedPnL = (openPos.EntryPrice - price) * quantity
			}
		}
	}

	if err := at.store.Order().CreateFill(fill); err != nil {
		logger.Infof("  ⚠️ Failed to record fill: %v", err)
	} else {
		logger.Infof("  📋 Fill recorded: %.4f @ %.6f, fee: %.4f", quantity, price, fee)
	}
}

// ============================================================================
// Risk Control Helpers
// ============================================================================

// isBTCETH checks if a symbol is BTC or ETH
func isBTCETH(symbol string) bool {
	symbol = strings.ToUpper(symbol)
	return strings.HasPrefix(symbol, "BTC") || strings.HasPrefix(symbol, "ETH")
}

// enforcePositionValueRatio checks and enforces position value ratio limits (CODE ENFORCED)
// Returns the adjusted position size (capped if necessary) and whether the position was capped
// positionSizeUSD: the original position size in USD
// equity: the account equity
// symbol: the trading symbol
func (at *AutoTrader) enforcePositionValueRatio(positionSizeUSD float64, equity float64, symbol string) (float64, bool) {
	if at.config.StrategyConfig == nil {
		return positionSizeUSD, false
	}

	riskControl := at.config.StrategyConfig.RiskControl

	// Get the appropriate position value ratio limit
	var maxPositionValueRatio float64
	if isBTCETH(symbol) {
		maxPositionValueRatio = riskControl.BTCETHMaxPositionValueRatio
		if maxPositionValueRatio <= 0 {
			maxPositionValueRatio = 5.0 // Default: 5x for BTC/ETH
		}
	} else {
		maxPositionValueRatio = riskControl.AltcoinMaxPositionValueRatio
		if maxPositionValueRatio <= 0 {
			maxPositionValueRatio = 1.0 // Default: 1x for altcoins
		}
	}

	// Calculate max allowed position value = equity × ratio
	maxPositionValue := equity * maxPositionValueRatio

	// Check if position size exceeds limit
	if positionSizeUSD > maxPositionValue {
		logger.Infof("  ⚠️ [POSITION_VALUE_RATIO_LIMIT] Position %.2f USDT exceeds limit (equity %.2f × %.1fx = %.2f USDT max for %s), capping",
			positionSizeUSD, equity, maxPositionValueRatio, maxPositionValue, symbol)
		return maxPositionValue, true
	}

	return positionSizeUSD, false
}

// enforceMinPositionSize checks minimum position size (CODE ENFORCED)
func (at *AutoTrader) enforceMinPositionSize(positionSizeUSD float64) error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	minSize := at.config.StrategyConfig.RiskControl.MinPositionSize
	if minSize <= 0 {
		minSize = 12 // Default: 12 USDT
	}

	if positionSizeUSD < minSize {
		return fmt.Errorf("❌ [MIN_POSITION_SIZE_LIMIT] Position size: %.2f USDT below minimum: %.2f USDT - Trade blocked", positionSizeUSD, minSize)
	}
	return nil
}

// enforceMaxPositions checks maximum positions count (CODE ENFORCED)
func (at *AutoTrader) enforceMaxPositions(currentPositionCount int) error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	maxPositions := at.config.StrategyConfig.RiskControl.MaxPositions
	if maxPositions <= 0 {
		maxPositions = 3 // Default: 3 positions
	}

	if currentPositionCount >= maxPositions {
		return fmt.Errorf("❌ [MAX_POSITIONS_LIMIT] Current positions: %d/%d - Position opening blocked", currentPositionCount, maxPositions)
	}
	return nil
}

// enforceMaxMarginUsage checks maximum margin usage (CODE ENFORCED)
func (at *AutoTrader) enforceMaxMarginUsage(currentMarginUsagePct float64) error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	maxMarginUsage := at.config.StrategyConfig.RiskControl.MaxMarginUsage
	if maxMarginUsage <= 0 {
		maxMarginUsage = 0.9 // Default: 90%
	}

	if currentMarginUsagePct > maxMarginUsage*100 {
		return fmt.Errorf("❌ [MAX_MARGIN_LIMIT] Current margin usage: %.2f%% exceeds limit: %.0f%% - Trade blocked", currentMarginUsagePct, maxMarginUsage*100)
	}
	return nil
}

// enforceMinHoldTime checks minimum hold time for positions (CODE ENFORCED)
func (at *AutoTrader) enforceMinHoldTime(symbol string, side string) error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	minHoldTimeMinutes := at.config.StrategyConfig.RiskControl.MinHoldTimeMinutes
	if minHoldTimeMinutes <= 0 {
		return nil // If not set, don't enforce
	}

	positionKey := fmt.Sprintf("%s_%s", symbol, side)
	firstSeenTime, exists := at.positionFirstSeenTime[positionKey]
	if !exists {
		return nil // If no first seen time, let it pass
	}

	positionHeldDuration := time.Since(time.UnixMilli(firstSeenTime))
	minHoldDuration := time.Duration(minHoldTimeMinutes) * time.Minute

	if positionHeldDuration < minHoldDuration {
		remainingTime := minHoldDuration - positionHeldDuration
		return fmt.Errorf("❌ [MIN_HOLD_TIME_LIMIT] Position %s held: %v, minimum required: %v - Closing blocked (wait %v)",
			positionKey, positionHeldDuration.Round(time.Second), minHoldDuration, remainingTime.Round(time.Second))
	}

	return nil
}

// enforceMaxLossPerTrade checks maximum loss per single trade (CODE ENFORCED)
func (at *AutoTrader) enforceMaxLossPerTrade(symbol string, unrealizedPnL float64, positionValue float64) error {
	if at.config.StrategyConfig == nil || positionValue <= 0 {
		return nil
	}

	maxLossPercent := at.config.StrategyConfig.RiskControl.MaxLossPerTradePercent
	if maxLossPercent <= 0 {
		return nil // If not set, don't enforce
	}

	currentLossPercent := (math.Abs(unrealizedPnL) / positionValue) * 100

	if unrealizedPnL < 0 && currentLossPercent > maxLossPercent {
		return fmt.Errorf("❌ [MAX_LOSS_PER_TRADE_LIMIT] Position %s loss: %.2f%% exceeds limit: %.2f%% - Current loss: %.2f USDT",
			symbol, currentLossPercent, maxLossPercent, unrealizedPnL)
	}

	return nil
}

// enforceDailyLossLimit checks daily loss limit (CODE ENFORCED)
func (at *AutoTrader) enforceDailyLossLimit() error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	dailyLossLimitPercent := at.config.StrategyConfig.RiskControl.DailyLossLimitPercent
	if dailyLossLimitPercent <= 0 {
		return nil // If not set, don't enforce
	}

	// Reset daily P&L if day has changed
	currentTime := time.Now()
	if currentTime.Day() != at.lastResetTime.Day() ||
		currentTime.Month() != at.lastResetTime.Month() ||
		currentTime.Year() != at.lastResetTime.Year() {
		at.dailyPnL = 0
		at.lastResetTime = currentTime
	}

	currentDailyLossPercent := (math.Abs(at.dailyPnL) / at.initialBalance) * 100

	if at.dailyPnL < 0 && currentDailyLossPercent > dailyLossLimitPercent {
		return fmt.Errorf("❌ [DAILY_LOSS_LIMIT] Daily loss: %.2f%% exceeds limit: %.2f%% - Current daily loss: %.2f USDT",
			currentDailyLossPercent, dailyLossLimitPercent, at.dailyPnL)
	}

	return nil
}

// enforceTradeFrequencyLimits checks trade frequency limits (CODE ENFORCED)
func (at *AutoTrader) enforceTradeFrequencyLimits(symbol string) error {
	if at.config.StrategyConfig == nil {
		return nil
	}

	// Get the trade frequency limits from the strategy config
	riskControl := at.config.StrategyConfig.RiskControl
	maxDailyTrades := riskControl.MaxDailyTrades
	maxHourlyTrades := riskControl.MaxHourlyTrades
	maxTradesPerSymbolPerHour := riskControl.MaxTradesPerSymbolPerHour

	// Skip enforcement if all limits are disabled (set to 0 or less)
	if maxDailyTrades <= 0 && maxHourlyTrades <= 0 && maxTradesPerSymbolPerHour <= 0 {
		return nil
	}

	tracker := at.tradeFrequencyTracker
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	// Reset counters if necessary
	currentTime := time.Now()

	// Reset daily counter if day has changed
	if currentTime.Day() != tracker.dailyResetTime.Day() ||
		currentTime.Month() != tracker.dailyResetTime.Month() ||
		currentTime.Year() != tracker.dailyResetTime.Year() {
		tracker.dailyTrades = 0
		tracker.dailyResetTime = currentTime
	}

	// Reset hourly counter if hour has changed
	if currentTime.Hour() != tracker.hourlyResetTime.Hour() ||
		currentTime.Day() != tracker.hourlyResetTime.Day() ||
		currentTime.Month() != tracker.hourlyResetTime.Month() ||
		currentTime.Year() != tracker.hourlyResetTime.Year() {
		tracker.hourlyTrades = 0
		tracker.symbolHourlyTrades = make(map[string]int) // Reset symbol-specific counters
		tracker.hourlyResetTime = currentTime
		tracker.symbolHourlyResetTime = currentTime
	}

	// Check daily trade limit
	if maxDailyTrades > 0 && tracker.dailyTrades >= maxDailyTrades {
		return fmt.Errorf("❌ [DAILY_TRADE_LIMIT] Daily trades: %d/%d reached - Trading blocked for today", tracker.dailyTrades, maxDailyTrades)
	}

	// Check hourly trade limit
	if maxHourlyTrades > 0 && tracker.hourlyTrades >= maxHourlyTrades {
		return fmt.Errorf("❌ [HOURLY_TRADE_LIMIT] Hourly trades: %d/%d reached - Trading blocked for this hour", tracker.hourlyTrades, maxHourlyTrades)
	}

	// Check symbol-specific hourly trade limit
	if maxTradesPerSymbolPerHour > 0 {
		symbolTrades := tracker.symbolHourlyTrades[symbol]
		if symbolTrades >= maxTradesPerSymbolPerHour {
			return fmt.Errorf("❌ [SYMBOL_HOURLY_LIMIT] Symbol '%s' hourly trades: %d/%d reached - Trading blocked for this symbol", symbol, symbolTrades, maxTradesPerSymbolPerHour)
		}
	}

	// If we pass all checks, increment the counters
	tracker.dailyTrades++
	tracker.hourlyTrades++
	tracker.symbolHourlyTrades[symbol]++

	return nil
}

// getSideFromAction converts order action to side (BUY/SELL)
func getSideFromAction(action string) string {
	switch action {
	case "open_long", "close_short":
		return "BUY"
	case "open_short", "close_long":
		return "SELL"
	default:
		return "BUY"
	}
}

// GetOpenOrders returns open orders (pending SL/TP) from exchange
func (at *AutoTrader) GetOpenOrders(symbol string) ([]OpenOrder, error) {
	return at.trader.GetOpenOrders(symbol)
}

// getBinanceCustomEndpointForAutoTrader extracts the custom API endpoint for Binance from auto trader config
func getBinanceCustomEndpointForAutoTrader(config *AutoTraderConfig) string {
	// 判断条件a：如果CustomAPIURL为空，使用默认主网API
	if config.BinanceCustomAPIURL == "" {
		return "https://fapi.binance.com" // 默认主网API URL
	}

	// 判断条件b：如果CustomAPIURL不为空且不是空白字符，直接使用CustomAPIURL
	trimmedURL := strings.TrimSpace(config.BinanceCustomAPIURL)
	if trimmedURL == "" {
		return "https://fapi.binance.com" // 默认主网API URL
	}

	// 判断条件c：检查是否为本地代理地址（localhost或127.0.0.1）
	isLocalhost := strings.Contains(trimmedURL, "://localhost:") ||
		strings.Contains(trimmedURL, "://127.0.0.1:")

	// 判断条件d：如果是本地代理地址，使用默认主网API
	if isLocalhost {
		return "https://fapi.binance.com" // 默认主网API URL
	} else {
		// 如果不是本地代理地址，直接使用CustomAPIURL
		return trimmedURL
	}
}

// executeUpdateStopLossWithRecord executes update stop loss and records detailed information
func (at *AutoTrader) executeUpdateStopLossWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Update stop loss: %s to %.4f", decision.Symbol, decision.NewStopLoss)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	if foundPos == nil {
		return fmt.Errorf("no position found for symbol %s", decision.Symbol)
	}

	// Get the current position quantity
	qtyFloat, ok := foundPos["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		quantity, ok2 := foundPos["positionAmt"].(*json.Number)
		if !ok2 {
			return fmt.Errorf("failed to get position amount")
		}
		var err error
		qtyFloat, err = quantity.Float64()
		if err != nil {
			return fmt.Errorf("failed to convert quantity to float: %w", err)
		}
	}

	// Convert negative quantity to positive if needed
	if qtyFloat < 0 {
		qtyFloat = math.Abs(qtyFloat)
	}

	// Get position side
	positionSide, ok := foundPos["positionSide"].(string)
	if !ok {
		// Some exchanges use 'side' instead of 'positionSide'
		positionSide, _ = foundPos["side"].(string)
	}

	// Determine side for stop loss
	side := "LONG"
	if positionSide == "SHORT" || (positionSide == "" && strings.Contains(strings.ToUpper(foundPos["symbol"].(string)), "USDT") && foundPos["side"].(string) == "SHORT") {
		side = "SHORT"
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
	}

	// Use current market price if new stop loss price is 0
	newStopLossPrice := decision.NewStopLoss
	if newStopLossPrice == 0 && marketData != nil {
		newStopLossPrice = marketData.CurrentPrice
		logger.Infof("  💡 Using current market price as stop loss price: %.4f", newStopLossPrice)
	}

	// Debug log for exchange API call
	logger.Infof("  📡 Submitting stop loss update to exchange: Symbol=%s, Side=%s, NewStopLoss=%.4f", decision.Symbol, side, newStopLossPrice)
	// Update stop loss
	err = at.trader.UpdateStopLoss(decision.Symbol, side, newStopLossPrice)
	if err != nil {
		logger.Errorf("  ❌ Failed to update stop loss for %s: %v", decision.Symbol, err)
		return fmt.Errorf("failed to update stop loss: %w", err)
	}

	logger.Infof("  ✓ Stop loss updated successfully for %s to %.4f", decision.Symbol, decision.NewStopLoss)

	// Record the stop loss update action to database
	if at.store != nil {
		orderID := fmt.Sprintf("SL_%s_%d", decision.Symbol, time.Now().Unix())
		// Record the stop loss update as an action
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    side,
			OrderAction:     "update_stop_loss",
			Type:            "STOP_MARKET", // Or STOP_LIMIT depending on implementation
			Side:            "STOP_LOSS",
			Quantity:        qtyFloat,
			Price:           decision.NewStopLoss, // Target stop loss price
			Status:          "UPDATED",            // Status indicating the stop loss was updated
			FilledQuantity:  0,                    // Not filled yet, just updated
			AvgFillPrice:    0,                    // Will be filled when triggered
			Commission:      0,                    // No commission for stop loss updates
			FilledAt:        0,                    // Will be set when triggered
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		// Add market price at time of update for reference
		if marketData != nil {
			orderRecord.AvgFillPrice = marketData.CurrentPrice // Current market price at update time
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record stop loss update: %v", err)
		} else {
			logger.Infof("  📊 Stop loss update recorded: %s new SL: %.4f, current price: %.4f",
				decision.Symbol, decision.NewStopLoss, orderRecord.AvgFillPrice)
		}
	}

	return nil
}

// executeUpdateTakeProfitWithRecord executes update take profit and records detailed information
func (at *AutoTrader) executeUpdateTakeProfitWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Update take profit: %s to %.4f", decision.Symbol, decision.NewTakeProfit)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	if foundPos == nil {
		return fmt.Errorf("no position found for symbol %s", decision.Symbol)
	}

	// Get the current position quantity
	qtyFloat, ok := foundPos["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		quantity, ok2 := foundPos["positionAmt"].(*json.Number)
		if !ok2 {
			return fmt.Errorf("failed to get position amount")
		}
		var err error
		qtyFloat, err = quantity.Float64()
		if err != nil {
			return fmt.Errorf("failed to convert quantity to float: %w", err)
		}
	}

	// Convert negative quantity to positive if needed
	if qtyFloat < 0 {
		qtyFloat = math.Abs(qtyFloat)
	}

	// Get position side
	positionSide, ok := foundPos["positionSide"].(string)
	if !ok {
		// Some exchanges use 'side' instead of 'positionSide'
		positionSide, _ = foundPos["side"].(string)
	}

	// Determine side for take profit
	side := "LONG"
	if positionSide == "SHORT" || (positionSide == "" && strings.Contains(strings.ToUpper(foundPos["symbol"].(string)), "USDT") && foundPos["side"].(string) == "SHORT") {
		side = "SHORT"
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
	}

	// Use current market price if new take profit price is 0
	newTakeProfitPrice := decision.NewTakeProfit
	if newTakeProfitPrice == 0 && marketData != nil {
		newTakeProfitPrice = marketData.CurrentPrice
		logger.Infof("  💡 Using current market price as take profit price: %.4f", newTakeProfitPrice)
	}

	// Debug log for exchange API call
	logger.Infof("  📡 Submitting take profit update to exchange: Symbol=%s, Side=%s, NewTakeProfit=%.4f", decision.Symbol, side, newTakeProfitPrice)
	// Update take profit
	err = at.trader.UpdateTakeProfit(decision.Symbol, side, newTakeProfitPrice)
	if err != nil {
		logger.Errorf("  ❌ Failed to update take profit for %s: %v", decision.Symbol, err)
		return fmt.Errorf("failed to update take profit: %w", err)
	}

	logger.Infof("  ✓ Take profit updated successfully for %s to %.4f", decision.Symbol, decision.NewTakeProfit)

	// Record the take profit update action to database
	if at.store != nil {
		orderID := fmt.Sprintf("TP_%s_%d", decision.Symbol, time.Now().Unix())
		// Record the take profit update as an action
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    side,
			OrderAction:     "update_take_profit",
			Type:            "TAKE_PROFIT_MARKET", // Or TAKE_PROFIT_LIMIT depending on implementation
			Side:            "TAKE_PROFIT",
			Quantity:        qtyFloat,
			Price:           decision.NewTakeProfit, // Target take profit price
			Status:          "UPDATED",              // Status indicating the take profit was updated
			FilledQuantity:  0,                      // Not filled yet, just updated
			AvgFillPrice:    0,                      // Will be filled when triggered
			Commission:      0,                      // No commission for take profit updates
			FilledAt:        0,                      // Will be set when triggered
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		// Add market price at time of update for reference
		if marketData != nil {
			orderRecord.AvgFillPrice = marketData.CurrentPrice // Current market price at update time
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record take profit update: %v", err)
		} else {
			logger.Infof("  📊 Take profit update recorded: %s new TP: %.4f, current price: %.4f",
				decision.Symbol, decision.NewTakeProfit, orderRecord.AvgFillPrice)
		}
	}

	return nil
}

// executePartialCloseWithRecord executes partial close and records detailed information
func (at *AutoTrader) executePartialCloseWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Partial close: %s, %.2f%%", decision.Symbol, decision.ClosePercentage)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	if foundPos == nil {
		return fmt.Errorf("no position found for symbol %s", decision.Symbol)
	}

	// Get the current position quantity
	qtyFloat, ok := foundPos["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		quantity, ok2 := foundPos["positionAmt"].(*json.Number)
		if !ok2 {
			return fmt.Errorf("failed to get position amount")
		}
		var err error
		qtyFloat, err = quantity.Float64()
		if err != nil {
			return fmt.Errorf("failed to convert quantity to float: %w", err)
		}
	}

	// Determine if long or short position
	isLong := qtyFloat > 0
	absQty := math.Abs(qtyFloat)

	// Calculate partial close quantity
	partialQty := absQty * decision.ClosePercentage / 100

	// Determine position type for closing (used by PartialClose function)
	positionType := "short"
	if isLong {
		positionType = "long" // Long position to close
	} else {
		positionType = "short" // Short position to close
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
	}

	// Debug log for exchange API call
	logger.Infof("  📡 Submitting partial close to exchange: Symbol=%s, Position Type=%s, ClosePercentage=%.2f%%", decision.Symbol, positionType, decision.ClosePercentage)
	// Close partial position
	order, err := at.trader.PartialClose(decision.Symbol, positionType, decision.ClosePercentage)
	if err != nil {
		logger.Errorf("  ❌ Failed to partially close position for %s: %v", decision.Symbol, err)
		return fmt.Errorf("failed to partially close position: %w", err)
	}

	logger.Infof("  ✓ Partial close executed successfully for %s, %.4f quantity", decision.Symbol, partialQty)

	// Record order to database with market price reference
	at.recordAndConfirmOrder(order, decision.Symbol, "partial_close", partialQty, marketData.CurrentPrice, 0, decision.ClosePercentage)
	return nil
}

// executeOCOOrderWithRecord executes OCO (One-Cancels-Other) order action
func (at *AutoTrader) executeOCOOrderWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Executing OCO order: %s", decision.Symbol)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine quantity and side
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	// Determine if we're working with an existing position or opening a new one
	isExistingPosition := foundPos != nil

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
		return err
	}

	// Calculate quantity based on position size if opening a new position
	var quantity float64
	if !isExistingPosition {
		if decision.PositionSizeUSD > 0 {
			quantity = decision.PositionSizeUSD / marketData.CurrentPrice
			actionRecord.Quantity = quantity
		} else {
			// Default to 0.001 BTC equivalent if no position size specified
			quantity = 0.001
			actionRecord.Quantity = quantity
		}
	} else {
		// For existing position, use current position quantity
		qtyFloat, ok := foundPos["positionAmt"].(float64)
		if !ok {
			// Try to get as json.Number if it fails as float64
			quantityNum, ok2 := foundPos["positionAmt"].(*json.Number)
			if !ok2 {
				return fmt.Errorf("failed to get position amount")
			}
			var err error
			qtyFloat, err = quantityNum.Float64()
			if err != nil {
				return fmt.Errorf("failed to convert quantity to float: %w", err)
			}
		}
		quantity = math.Abs(qtyFloat) // Use absolute value
		actionRecord.Quantity = quantity
	}

	actionRecord.Price = marketData.CurrentPrice

	// Debug log for exchange API call
	logger.Infof("  📡 Submitting OCO order to exchange: Symbol=%s, Quantity=%.8f, StopLoss=%.4f, TakeProfit=%.4f",
		decision.Symbol, quantity, decision.StopLoss, decision.TakeProfit)

	// Submit OCO order to exchange
	// Note: OCO orders are not universally supported by all exchanges
	// This implementation submits both stop-loss and take-profit orders separately
	// which achieves similar functionality to an OCO order
	var submitErrors []string

	// Submit OCO order based on existing position
	// If we have an existing position, set protective stop-loss/take-profit
	// If no position exists, create entry orders with protective stops
	if decision.StopLoss > 0 || decision.TakeProfit > 0 {
		if isExistingPosition {
			// For existing positions, set protective stop-loss and take-profit
			positionSide := "LONG"
			if foundPos != nil {
				positionSideVal, ok := foundPos["positionSide"].(string)
				if !ok {
					positionSideVal, _ = foundPos["side"].(string)
				}
				if positionSideVal == "SHORT" {
					positionSide = "SHORT"
				}
			}

			// Submit stop loss order
			if decision.StopLoss > 0 {
				if err := at.trader.SetStopLoss(decision.Symbol, positionSide, quantity, decision.StopLoss); err != nil {
					logger.Infof("  ⚠️ Failed to set stop loss: %v", err)
					submitErrors = append(submitErrors, fmt.Sprintf("stop-loss: %v", err))
				} else {
					logger.Infof("  ✓ Stop loss set successfully: %.4f", decision.StopLoss)
				}
			}

			// Submit take profit order
			if decision.TakeProfit > 0 {
				if err := at.trader.SetTakeProfit(decision.Symbol, positionSide, quantity, decision.TakeProfit); err != nil {
					logger.Infof("  ⚠️ Failed to set take profit: %v", err)
					submitErrors = append(submitErrors, fmt.Sprintf("take-profit: %v", err))
				} else {
					logger.Infof("  ✓ Take profit set successfully: %.4f", decision.TakeProfit)
				}
			}
		} else {
			// For new positions, we should ideally create entry orders with stop/limit orders
			// But since we don't have a direct method for this, we'll handle based on the intent
			// If stop loss is below current price and take profit is above, likely trying to go LONG
			// If stop loss is above current price and take profit is below, likely trying to go SHORT

			marketPrice := marketData.CurrentPrice
			var intendedSide string

			// Determine intended trade direction based on SL/TP prices vs current market price
			if decision.StopLoss < marketPrice && decision.TakeProfit > marketPrice {
				// Likely trying to go LONG: StopLoss below market, TakeProfit above market
				intendedSide = "LONG"
			} else if decision.StopLoss > marketPrice && decision.TakeProfit < marketPrice {
				// Likely trying to go SHORT: StopLoss above market, TakeProfit below market
				intendedSide = "SHORT"
			} else {
				// Ambiguous or invalid configuration
				logger.Warnf("  ⚠️ Ambiguous OCO order configuration for %s: Market=%.2f, SL=%.2f, TP=%.2f",
					decision.Symbol, marketPrice, decision.StopLoss, decision.TakeProfit)
				submitErrors = append(submitErrors, fmt.Sprintf("invalid OCO configuration for entry: market=%.2f, sl=%.2f, tp=%.2f",
					marketPrice, decision.StopLoss, decision.TakeProfit))
			}

			if intendedSide != "" {
				// We could implement a proper entry order here, but for now we'll just log the intent
				logger.Infof("  📊 OCO order intent: Enter %s position for %s, SL: %.4f, TP: %.4f",
					intendedSide, decision.Symbol, decision.StopLoss, decision.TakeProfit)
				logger.Infof("  ⚠️ Note: Entry orders with stop/limit not yet implemented in OCO. Please create entry position first.")
				submitErrors = append(submitErrors, fmt.Sprintf("entry orders with OCO not supported, create position first"))
			}
		}
	}

	if len(submitErrors) == 2 {
		// Both orders failed
		return fmt.Errorf("OCO order submission failed: %s", strings.Join(submitErrors, ", "))
	}

	// Log successful execution
	logger.Infof("  ✓ OCO order executed successfully for %s, StopLoss: %.4f, TakeProfit: %.4f",
		decision.Symbol, decision.StopLoss, decision.TakeProfit)

	// Record the OCO order action to database
	if at.store != nil {
		orderID := fmt.Sprintf("OCO_%s_%d", decision.Symbol, time.Now().Unix())
		// Record the OCO order as an action
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    "BOTH", // Indicates both stop-loss and take-profit
			OrderAction:     "oco_order",
			Type:            "OCO", // One-Cancels-Other order type
			Side:            "OCO",
			Quantity:        quantity,
			Price:           marketData.CurrentPrice, // Reference price at time of order
			StopPrice:       decision.StopLoss,       // Associated stop loss price (using correct field name)
			Status:          "ACTIVE",                // Status indicating the OCO is active
			FilledQuantity:  0,                       // Not filled yet, just activated
			AvgFillPrice:    0,                       // Will be filled when triggered
			Commission:      0,                       // No commission for OCO setup
			FilledAt:        0,                       // Will be set when triggered
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record OCO order: %v", err)
		} else {
			logger.Infof("  📊 OCO order recorded: %s SL: %.4f TP: %.4f",
				decision.Symbol, decision.StopLoss, decision.TakeProfit)
		}
	}

	return nil
}

// executeBracketOrderWithRecord executes bracket order action (Step 1: Open position, Step 2: Set SL/TP)
func (at *AutoTrader) executeBracketOrderWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  🔄 Executing bracket order (2-step): %s", decision.Symbol)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to check if position already exists
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	// Get current market price for reference
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
		return err
	}

	var quantity float64
	var positionSide string

	// STEP 1: Open position if not exists
	if foundPos == nil {
		logger.Infof("  📊 Step 1/2: Opening new position for %s", decision.Symbol)

		// Calculate quantity based on position size
		if decision.PositionSizeUSD > 0 {
			quantity = decision.PositionSizeUSD / marketData.CurrentPrice
		} else {
			return fmt.Errorf("bracket order requires position_size_usd")
		}

		// Determine trade direction based on stop loss and take profit prices
		// LONG: StopLoss < CurrentPrice < TakeProfit
		// SHORT: TakeProfit < CurrentPrice < StopLoss
		if decision.StopLoss < marketData.CurrentPrice && decision.TakeProfit > marketData.CurrentPrice {
			positionSide = "LONG"
			logger.Infof("  📈 Detected LONG position intent (SL: %.2f < Price: %.2f < TP: %.2f)",
				decision.StopLoss, marketData.CurrentPrice, decision.TakeProfit)
		} else if decision.StopLoss > marketData.CurrentPrice && decision.TakeProfit < marketData.CurrentPrice {
			positionSide = "SHORT"
			logger.Infof("  📉 Detected SHORT position intent (TP: %.2f < Price: %.2f < SL: %.2f)",
				decision.TakeProfit, marketData.CurrentPrice, decision.StopLoss)
		} else {
			return fmt.Errorf("invalid bracket order configuration: cannot determine direction (Market=%.2f, SL=%.2f, TP=%.2f)",
				marketData.CurrentPrice, decision.StopLoss, decision.TakeProfit)
		}

		// Open the position
		var openResult map[string]interface{}
		if positionSide == "LONG" {
			openResult, err = at.trader.OpenLong(decision.Symbol, quantity, decision.Leverage)
		} else {
			openResult, err = at.trader.OpenShort(decision.Symbol, quantity, decision.Leverage)
		}

		if err != nil {
			return fmt.Errorf("failed to open %s position: %w", positionSide, err)
		}

		logger.Infof("  ✓ Step 1/2 Complete: %s position opened successfully (OrderID: %v)",
			positionSide, openResult["orderId"])

		// Wait a moment for position to be established
		time.Sleep(500 * time.Millisecond)

		// Refresh positions after opening
		positions, err = at.trader.GetPositions()
		if err != nil {
			logger.Warnf("  ⚠️ Failed to refresh positions: %v", err)
		} else {
			for _, pos := range positions {
				if pos["symbol"] == decision.Symbol {
					foundPos = pos
					break
				}
			}
		}

		actionRecord.Quantity = quantity
		actionRecord.Price = marketData.CurrentPrice
	} else {
		logger.Infof("  ℹ️ Position already exists for %s, will set SL/TP only", decision.Symbol)

		// Get existing position details
		qtyFloat, ok := foundPos["positionAmt"].(float64)
		if !ok {
			quantityNum, ok2 := foundPos["positionAmt"].(*json.Number)
			if !ok2 {
				return fmt.Errorf("failed to get position amount")
			}
			qtyFloat, err = quantityNum.Float64()
			if err != nil {
				return fmt.Errorf("failed to convert quantity to float: %w", err)
			}
		}
		quantity = math.Abs(qtyFloat)

		// Determine position side
		positionSideVal, ok := foundPos["positionSide"].(string)
		if !ok {
			positionSideVal, _ = foundPos["side"].(string)
		}
		if positionSideVal == "SHORT" {
			positionSide = "SHORT"
		} else {
			positionSide = "LONG"
		}

		actionRecord.Quantity = quantity
		actionRecord.Price = marketData.CurrentPrice
	}

	// STEP 2: Set stop loss and take profit
	logger.Infof("  📊 Step 2/2: Setting SL/TP for %s %s position", decision.Symbol, positionSide)

	var submitErrors []string

	// Set stop-loss order
	if decision.StopLoss > 0 {
		if err := at.trader.SetStopLoss(decision.Symbol, positionSide, quantity, decision.StopLoss); err != nil {
			logger.Infof("  ⚠️ Failed to set stop loss: %v", err)
			submitErrors = append(submitErrors, fmt.Sprintf("stop-loss: %v", err))
		} else {
			logger.Infof("  ✓ Stop loss set successfully: %.4f", decision.StopLoss)
		}
	}

	// Set take-profit order
	if decision.TakeProfit > 0 {
		if err := at.trader.SetTakeProfit(decision.Symbol, positionSide, quantity, decision.TakeProfit); err != nil {
			logger.Infof("  ⚠️ Failed to set take profit: %v", err)
			submitErrors = append(submitErrors, fmt.Sprintf("take-profit: %v", err))
		} else {
			logger.Infof("  ✓ Take profit set successfully: %.4f", decision.TakeProfit)
		}
	}

	// Check if both SL/TP failed
	if len(submitErrors) == 2 {
		return fmt.Errorf("bracket order SL/TP failed: %s (position was opened, but protective orders failed)", strings.Join(submitErrors, ", "))
	}

	// Log successful execution
	if len(submitErrors) == 0 {
		logger.Infof("  ✅ Step 2/2 Complete: Bracket order fully executed for %s (SL: %.4f, TP: %.4f)",
			decision.Symbol, decision.StopLoss, decision.TakeProfit)
	} else {
		logger.Infof("  ⚠️ Bracket order partially executed: %s", strings.Join(submitErrors, ", "))
	}

	// Record the bracket order action to database
	if at.store != nil {
		orderID := fmt.Sprintf("BRACKET_%s_%d", decision.Symbol, time.Now().Unix())
		orderRecord := &store.TraderOrder{
			TraderID:        at.id,
			ExchangeID:      at.exchangeID,
			ExchangeType:    at.exchange,
			ExchangeOrderID: orderID,
			Symbol:          decision.Symbol,
			PositionSide:    positionSide,
			OrderAction:     "bracket_order",
			Type:            "BRACKET",
			Side:            positionSide,
			Quantity:        quantity,
			Price:           marketData.CurrentPrice,
			StopPrice:       decision.StopLoss,
			Status:          "ACTIVE",
			FilledQuantity:  quantity,
			AvgFillPrice:    marketData.CurrentPrice,
			Commission:      0,
			FilledAt:        time.Now().UTC().UnixMilli(),
			CreatedAt:       time.Now().UTC().UnixMilli(),
			UpdatedAt:       time.Now().UTC().UnixMilli(),
		}

		if err := at.store.Order().CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to record bracket order: %v", err)
		} else {
			logger.Infof("  📊 Bracket order recorded: %s %s SL: %.4f TP: %.4f",
				decision.Symbol, positionSide, decision.StopLoss, decision.TakeProfit)
		}
	}

	return nil
}

// executeAddToPositionWithRecord executes add to position action and records detailed information
func (at *AutoTrader) executeAddToPositionWithRecord(decision *kernel.Decision, actionRecord *store.DecisionAction) error {
	logger.Infof("  ➕ Adding to position: %s", decision.Symbol)

	// Only enforce trade frequency limits for position opening actions (open_long, open_short)
	// Updates to existing positions (stop loss, take profit, etc.) should not be limited
	if decision.Action == "open_long" || decision.Action == "open_short" {
		if err := at.enforceTradeFrequencyLimits(decision.Symbol); err != nil {
			return err
		}
	}

	// Get current positions to determine existing position type
	positions, err := at.trader.GetPositions()
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Find the existing position for this symbol
	var foundPos map[string]interface{}
	for _, pos := range positions {
		if pos["symbol"] == decision.Symbol {
			foundPos = pos
			break
		}
	}

	// Get current market price for calculations
	marketData, err := market.Get(decision.Symbol)
	if err != nil {
		logger.Warnf("  ⚠️ Failed to get market data for %s: %v", decision.Symbol, err)
		return err
	}

	// Determine if we're adding to a long or short position
	addType := decision.AddPositionType
	if addType == "" {
		// If not specified in decision, try to determine from existing position
		if foundPos != nil {
			posSide, ok := foundPos["side"].(string)
			if !ok {
				posSide, _ = foundPos["positionSide"].(string)
			}
			addType = strings.ToUpper(posSide)
		} else {
			// If no existing position, default to adding to long
			addType = "LONG"
		}
	} else {
		// Convert user input to uppercase for validation
		addType = strings.ToUpper(addType)
	}

	// Validate addType
	if addType != "LONG" && addType != "SHORT" {
		return fmt.Errorf("invalid add position type: %s, must be 'LONG' or 'SHORT'", addType)
	}

	// Calculate quantity based on additional position size
	additionalSize := decision.AdditionalPositionSizeUSD
	if additionalSize <= 0 {
		// If not specified in decision, use PositionSizeUSD as fallback
		additionalSize = decision.PositionSizeUSD
	}
	if additionalSize <= 0 {
		return fmt.Errorf("additional position size must be greater than 0")
	}

	quantity := additionalSize / marketData.CurrentPrice

	logger.Infof("  📊 Adding to %s position: %s, additional size: $%.2f, quantity: %.6f",
		addType, decision.Symbol, additionalSize, quantity)

	// Execute the additional position based on type
	var order map[string]interface{}
	switch addType {
	case "LONG":
		order, err = at.trader.OpenLong(decision.Symbol, quantity, decision.Leverage)
		if err != nil {
			logger.Errorf("  ❌ Failed to open long position %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to add to long position: %w", err)
		}
	case "SHORT":
		order, err = at.trader.OpenShort(decision.Symbol, quantity, decision.Leverage)
		if err != nil {
			logger.Errorf("  ❌ Failed to open short position %s: %v", decision.Symbol, err)
			return fmt.Errorf("failed to add to short position: %w", err)
		}
	default:
		return fmt.Errorf("unsupported position type: %s", addType)
	}

	logger.Infof("  ✓ Successfully added to %s position: %s, quantity: %.6f", addType, decision.Symbol, quantity)

	// Update action record
	actionRecord.Quantity = quantity
	actionRecord.Price = marketData.CurrentPrice
	actionRecord.Leverage = decision.Leverage

	// Record order to database and poll for confirmation
	at.recordAndConfirmOrder(order, decision.Symbol, "add_to_position", quantity, marketData.CurrentPrice, decision.Leverage, 0)

	// Set stop loss and take profit if provided
	if decision.StopLoss > 0 {
		if err := at.trader.SetStopLoss(decision.Symbol, addType, quantity, decision.StopLoss); err != nil {
			logger.Errorf("  ❌ Failed to set stop loss for %s: %v", decision.Symbol, err)
			// Don't return error here as the main position addition was successful
		} else {
			logger.Infof("  ✓ Stop loss set for %s: %.4f", decision.Symbol, decision.StopLoss)
		}
	}
	if decision.TakeProfit > 0 {
		if err := at.trader.SetTakeProfit(decision.Symbol, addType, quantity, decision.TakeProfit); err != nil {
			logger.Errorf("  ❌ Failed to set take profit for %s: %v", decision.Symbol, err)
			// Don't return error here as the main position addition was successful
		} else {
			logger.Infof("  ✓ Take profit set for %s: %.4f", decision.Symbol, decision.TakeProfit)
		}
	}

	return nil
}

// GetTrader returns the underlying trader instance
func (at *AutoTrader) GetTrader() Trader {
	return at.trader
}

// GetExchangeID returns the exchange account UUID
func (at *AutoTrader) GetExchangeID() string {
	return at.exchangeID
}

// GetExchangeType returns the exchange type
func (at *AutoTrader) GetExchangeType() string {
	return at.exchange
}
