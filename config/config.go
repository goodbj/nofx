package config

import (
	"nofx/experience"
	"nofx/mcp"
	"os"
	"strconv"
	"strings"
)

// Global configuration instance
var global *Config

// Config is the global configuration (loaded from .env)
// Only contains truly global config, trading related config is at trader/strategy level
// ModelPricingConfig holds pricing information for a model
type ModelPricingConfig struct {
	InputPrice  float64 // Price per million tokens for input
	OutputPrice float64 // Price per million tokens for output
}

// CostDisplayUnit represents the currency unit for cost display
const (
	USD  = "USD"
	CNY  = "CNY"
	BOTH = "BOTH"
)

type Config struct {
	// Service configuration
	APIServerPort       int
	JWTSecret           string
	JWTExpirationDays   int // JWT token expiration in days
	RegistrationEnabled bool
	MaxUsers            int // Maximum number of users allowed (0 = unlimited, default = 10)

	// Database configuration
	DBType     string // sqlite or postgres
	DBPath     string // SQLite database file path
	DBHost     string // PostgreSQL host
	DBPort     int    // PostgreSQL port
	DBUser     string // PostgreSQL user
	DBPassword string // PostgreSQL password
	DBName     string // PostgreSQL database name
	DBSSLMode  string // PostgreSQL SSL mode

	// Security configuration
	// TransportEncryption enables browser-side encryption for API keys
	// Requires HTTPS or localhost. Set to false for HTTP access via IP.
	TransportEncryption bool

	// Experience improvement (anonymous usage statistics)
	// Helps us understand product usage and improve the experience
	// Set EXPERIENCE_IMPROVEMENT=false to disable
	ExperienceImprovement bool

	// Model token limits configuration
	ModelMaxTokens map[string]int

	// Model pricing configuration (per million tokens)
	ModelPricing map[string]ModelPricingConfig

	// Cost display configuration
	CostDisplayUnit string  // USD, CNY, or BOTH
	USDCNYRate      float64 // Exchange rate from USD to CNY

	// Market data provider API keys
	AlpacaAPIKey    string // Alpaca API key for US stocks
	AlpacaSecretKey string // Alpaca secret key
	TwelveDataKey   string // TwelveData API key for forex & metals
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	cfg := &Config{
		APIServerPort:         8080,
		JWTExpirationDays:     7, // Default: 7 days for JWT token expiration
		RegistrationEnabled:   true,
		MaxUsers:              10,   // Default: 10 users allowed
		ExperienceImprovement: true, // Default: enabled to help improve the product
		// Database defaults
		DBType:    "sqlite",
		DBPath:    "data/data.db",
		DBHost:    "localhost",
		DBPort:    5432,
		DBUser:    "postgres",
		DBName:    "nofx",
		DBSSLMode: "disable",
	}

	// Load from environment variables
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = strings.TrimSpace(v)
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "default-jwt-secret-change-in-production"
	}

	if v := os.Getenv("JWT_EXPIRATION_DAYS"); v != "" {
		if days, err := strconv.Atoi(v); err == nil && days > 0 {
			cfg.JWTExpirationDays = days
		}
	}

	if v := os.Getenv("REGISTRATION_ENABLED"); v != "" {
		cfg.RegistrationEnabled = strings.ToLower(v) == "true"
	}

	if v := os.Getenv("MAX_USERS"); v != "" {
		if maxUsers, err := strconv.Atoi(v); err == nil && maxUsers >= 0 {
			cfg.MaxUsers = maxUsers
		}
	}

	if v := os.Getenv("NOFX_BACKEND_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 {
			cfg.APIServerPort = port
		}
	}

	// Transport encryption: default false for easier deployment
	// Set TRANSPORT_ENCRYPTION=true to enable (requires HTTPS or localhost)
	if v := os.Getenv("TRANSPORT_ENCRYPTION"); v != "" {
		cfg.TransportEncryption = strings.ToLower(v) == "true"
	}

	// Experience improvement: anonymous usage statistics
	// Default enabled, set EXPERIENCE_IMPROVEMENT=false to disable
	if v := os.Getenv("EXPERIENCE_IMPROVEMENT"); v != "" {
		cfg.ExperienceImprovement = strings.ToLower(v) != "false"
	}

	// Load model token limits from environment variables
	loadModelTokenLimits(cfg)

	// Load model pricing from environment variables
	loadModelPricing(cfg)

	// Load cost display configuration
	loadCostDisplayConfig(cfg)

	// Market data provider API keys
	cfg.AlpacaAPIKey = os.Getenv("ALPACA_API_KEY")
	cfg.AlpacaSecretKey = os.Getenv("ALPACA_SECRET_KEY")
	cfg.TwelveDataKey = os.Getenv("TWELVEDATA_API_KEY")

	// Database configuration
	if v := os.Getenv("DB_TYPE"); v != "" {
		cfg.DBType = strings.ToLower(v)
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 {
			cfg.DBPort = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.DBSSLMode = v
	}

	global = cfg

	// Initialize experience improvement (installation ID will be set after database init)
	experience.Init(cfg.ExperienceImprovement, "")

	// Set up AI token usage tracking callback
	mcp.TokenUsageCallback = func(usage mcp.TokenUsage) {
		experience.TrackAIUsage(experience.AIUsageEvent{
			ModelProvider: usage.Provider,
			ModelName:     usage.Model,
			InputTokens:   usage.PromptTokens,
			OutputTokens:  usage.CompletionTokens,
		})
	}
	return cfg
}

// Load model token limits from environment variables
func loadModelTokenLimits(cfg *Config) {
	// Ensure ModelMaxTokens map is initialized
	if cfg.ModelMaxTokens == nil {
		cfg.ModelMaxTokens = make(map[string]int)
		// Set default values
		cfg.ModelMaxTokens["deepseek"] = 32768
		cfg.ModelMaxTokens["gpt-4"] = 128000
		cfg.ModelMaxTokens["gpt-3.5"] = 16384
		cfg.ModelMaxTokens["claude"] = 200000
		cfg.ModelMaxTokens["qwen"] = 32768
		cfg.ModelMaxTokens["ollama"] = 4096
		cfg.ModelMaxTokens["default"] = 4096
	}

	// Load individual model limits from environment variables
	if v := os.Getenv("MODEL_MAX_TOKENS_DEEPSEEK"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["deepseek"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_GPT4"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["gpt-4"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_GPT35"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["gpt-3.5"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_CLAUDE"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["claude"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_QWEN"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["qwen"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_DEFAULT"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["default"] = limit
		}
	}
	if v := os.Getenv("MODEL_MAX_TOKENS_OLLAMA"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			cfg.ModelMaxTokens["ollama"] = limit
		}
	}
}

// Load cost display configuration
func loadCostDisplayConfig(cfg *Config) {
	// Load cost display unit
	if v := os.Getenv("COST_DISPLAY_UNIT"); v != "" {
		unit := strings.ToUpper(strings.TrimSpace(v))
		switch unit {
		case USD, CNY, BOTH:
			cfg.CostDisplayUnit = unit
		default:
			cfg.CostDisplayUnit = BOTH // Default to BOTH
		}
	} else {
		cfg.CostDisplayUnit = BOTH // Default to BOTH
	}

	// Load exchange rate
	if v := os.Getenv("USD_TO_CNY_RATE"); v != "" {
		if rate, err := strconv.ParseFloat(v, 64); err == nil && rate > 0 {
			cfg.USDCNYRate = rate
		} else {
			cfg.USDCNYRate = 7.2 // Default rate
		}
	} else {
		cfg.USDCNYRate = 7.2 // Default rate
	}
}

// Load model pricing from environment variables
func loadModelPricing(cfg *Config) {
	// Ensure ModelPricing map is initialized
	if cfg.ModelPricing == nil {
		cfg.ModelPricing = make(map[string]ModelPricingConfig)
		// Set default values
		cfg.ModelPricing["qwen-max"] = ModelPricingConfig{InputPrice: 2.40, OutputPrice: 9.60}
		cfg.ModelPricing["qwen-plus"] = ModelPricingConfig{InputPrice: 1.00, OutputPrice: 4.00}
		cfg.ModelPricing["gpt-4o"] = ModelPricingConfig{InputPrice: 2.50, OutputPrice: 10.00}
		cfg.ModelPricing["gpt-4omini"] = ModelPricingConfig{InputPrice: 0.15, OutputPrice: 0.60}
		cfg.ModelPricing["claude-sonnet"] = ModelPricingConfig{InputPrice: 3.00, OutputPrice: 15.00}
		cfg.ModelPricing["deepseek"] = ModelPricingConfig{InputPrice: 0.20, OutputPrice: 0.80}
		cfg.ModelPricing["ollama"] = ModelPricingConfig{InputPrice: 0.00, OutputPrice: 0.00}
		cfg.ModelPricing["ernie"] = ModelPricingConfig{InputPrice: 0.08, OutputPrice: 0.20}
		cfg.ModelPricing["default"] = ModelPricingConfig{InputPrice: 0.50, OutputPrice: 2.00}
	}

	// Load individual model pricing from environment variables
	if v := os.Getenv("MODEL_PRICE_QWEN_MAX_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["qwen-max"]
			pricing.InputPrice = price
			cfg.ModelPricing["qwen-max"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_QWEN_MAX_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["qwen-max"]
			pricing.OutputPrice = price
			cfg.ModelPricing["qwen-max"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_QWEN_PLUS_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["qwen-plus"]
			pricing.InputPrice = price
			cfg.ModelPricing["qwen-plus"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_QWEN_PLUS_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["qwen-plus"]
			pricing.OutputPrice = price
			cfg.ModelPricing["qwen-plus"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_GPT4O_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["gpt-4o"]
			pricing.InputPrice = price
			cfg.ModelPricing["gpt-4o"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_GPT4O_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["gpt-4o"]
			pricing.OutputPrice = price
			cfg.ModelPricing["gpt-4o"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_GPT4OMINI_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["gpt-4omini"]
			pricing.InputPrice = price
			cfg.ModelPricing["gpt-4omini"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_GPT4OMINI_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["gpt-4omini"]
			pricing.OutputPrice = price
			cfg.ModelPricing["gpt-4omini"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_CLAUDE_SONNET_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["claude-sonnet"]
			pricing.InputPrice = price
			cfg.ModelPricing["claude-sonnet"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_CLAUDE_SONNET_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["claude-sonnet"]
			pricing.OutputPrice = price
			cfg.ModelPricing["claude-sonnet"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_DEEPSEEK_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["deepseek"]
			pricing.InputPrice = price
			cfg.ModelPricing["deepseek"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_DEEPSEEK_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["deepseek"]
			pricing.OutputPrice = price
			cfg.ModelPricing["deepseek"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_WENXIN_ERINIE35_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["ernie"]
			pricing.InputPrice = price
			cfg.ModelPricing["ernie"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_WENXIN_ERINIE35_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["ernie"]
			pricing.OutputPrice = price
			cfg.ModelPricing["ernie"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_DEFAULT_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["default"]
			pricing.InputPrice = price
			cfg.ModelPricing["default"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_DEFAULT_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["default"]
			pricing.OutputPrice = price
			cfg.ModelPricing["default"] = pricing
		}
	}

	// Load Ollama model pricing from environment variables
	if v := os.Getenv("MODEL_PRICE_OLLAMA_INPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["ollama"]
			pricing.InputPrice = price
			cfg.ModelPricing["ollama"] = pricing
		}
	}
	if v := os.Getenv("MODEL_PRICE_OLLAMA_OUTPUT"); v != "" {
		if price, err := strconv.ParseFloat(v, 64); err == nil && price >= 0 {
			pricing := cfg.ModelPricing["ollama"]
			pricing.OutputPrice = price
			cfg.ModelPricing["ollama"] = pricing
		}
	}
}

// Init 初始化全局配置
func Init() {
	LoadConfig()
}

// Get returns the global configuration
func Get() *Config {
	if global == nil {
		Init()
	}
	return global
}
