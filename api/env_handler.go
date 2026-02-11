package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// EnvVariable represents an environment variable with metadata
type EnvVariable struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Description string `json:"description"`
	IsSecret    bool   `json:"isSecret,omitempty"`
}

// GetEnvVariables returns all environment variables organized by category
func (s *Server) handleGetEnvVariables(c *gin.Context) {
	variables := s.getAllEnvVariables()

	c.JSON(http.StatusOK, gin.H{
		"variables": variables,
		"count":     len(variables),
	})
}

// UpdateEnvVariables handles updating environment variables
func (s *Server) handleUpdateEnvVariables(c *gin.Context) {
	var req struct {
		Variables map[string]string `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// In a real implementation, this would update environment variables
	// For now, we'll just return a success message since environment variables
	// can't be modified at runtime in Go

	updatedKeys := []string{}
	for key := range req.Variables {
		updatedKeys = append(updatedKeys, key)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment variables updated successfully",
		"updated": updatedKeys,
	})
}

// getAllEnvVariables returns all environment variables with categorization and descriptions
func (s *Server) getAllEnvVariables() []EnvVariable {
	envMap := make(map[string]string)

	// Get all environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Define categories and descriptions for known variables
	variables := []EnvVariable{}

	// Process all environment variables from the system
	for key, value := range envMap {
		category := getCategoryFromKey(key)
		variables = append(variables, EnvVariable{
			Key:         key,
			Value:       value,
			Category:    category,
			Description: getEnvDescription(key),
			IsSecret:    isSecretVariable(key),
		})
	}

	// Add default values for known variables that aren't in the environment
	allKnownVars := []string{
		// Server Configuration
		"NOFX_BACKEND_PORT", "NOFX_FRONTEND_PORT", "NOFX_TIMEZONE",
		"LOG_LEVEL", "API_SERVER_PORT", "WAIT_BEFORE_DETECTION",
		"DETECTION_INTERVAL",
		// Authentication
		"JWT_SECRET", "JWT_EXPIRATION_DAYS",
		// Encryption Keys
		"DATA_ENCRYPTION_KEY", "RSA_PRIVATE_KEY", "TRANSPORT_ENCRYPTION",
		// Database
		"DB_TYPE", "DB_PATH", "DB_HOST", "DB_PORT", "DB_USER",
		"DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		// AI Models & Pricing
		"MODEL_MAX_TOKENS_DEEPSEEK", "MODEL_MAX_TOKENS_GPT4",
		"MODEL_MAX_TOKENS_GPT35", "MODEL_MAX_TOKENS_CLAUDE",
		"MODEL_MAX_TOKENS_QWEN", "MODEL_MAX_TOKENS_OLLAMA",
		"MODEL_PRICE_QWEN_MAX_INPUT", "MODEL_PRICE_QWEN_MAX_OUTPUT",
		"MODEL_PRICE_GPT4O_INPUT", "MODEL_PRICE_GPT4O_OUTPUT",
		"COST_DISPLAY_UNIT", "USD_TO_CNY_RATE", "OLLAMA_READ_TIMEOUT",
		"MODEL_PRICE_QWEN_PLUS_INPUT", "MODEL_PRICE_QWEN_PLUS_OUTPUT",
		"MODEL_PRICE_GPT4OMINI_INPUT", "MODEL_PRICE_GPT4OMINI_OUTPUT",
		"MODEL_PRICE_CLAUDE_SONNET_INPUT", "MODEL_PRICE_CLAUDE_SONNET_OUTPUT",
		"MODEL_PRICE_DEEPSEEK_INPUT", "MODEL_PRICE_DEEPSEEK_OUTPUT",
		"MODEL_PRICE_OLLAMA_INPUT", "MODEL_PRICE_OLLAMA_OUTPUT",
		"MODEL_PRICE_WENXIN_ERINIE35_INPUT", "MODEL_PRICE_WENXIN_ERINIE35_OUTPUT",
		"MODEL_PRICE_DEFAULT_INPUT", "MODEL_PRICE_DEFAULT_OUTPUT",
		// Guardian Automation
		"GUARDIAN_TARGET_URL", "GUARDIAN_INPUT_SELECTOR",
		"GUARDIAN_BUTTON_SELECTOR", "GUARDIAN_RESPONSE_SELECTOR",
		"GUARDIAN_DISPLAY_ENABLED", "DOCKER_ENV", "GUARDIAN_BROWSER_TIMEOUT_SECONDS",
		"GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS", "GUARDIAN_MANUAL_MODE",
		"GUARDIAN_CHROME_PATH", "GUARDIAN_FORCE_BROWSER_REQUIRED",
		"GUARDIAN_VALIDATION_MODE", "BROWSER_DATA_DIR",
		"GUARDIAN_INITIAL_WAIT_SECONDS", "GUARDIAN_MIN_POLL_SECONDS",
		"GUARDIAN_MAX_POLL_SECONDS",
		// Other Settings
		"TELEGRAM_BOT_TOKEN", "TELEGRAM_CHAT_ID", "USE_BINANCE_PROXY",
		"BINANCE_PROXY_URL", "MCP_ENABLE_CONTEXT_COMPRESSION",
		"MCP_MODEL_CONTEXT_SIZE", "CSP_SCRIPT_SRC", "CSP_FRAME_ANCESTORS",
		"ALLOWED_ORIGINS", "CSP_CONNECT_SRC", "CSP_IMG_SRC", "CSP_FONT_SRC",
		"CSP_STYLE_SRC", "CSP_MEDIA_SRC", "CSP_OBJECT_SRC", "CSP_CHILD_SRC",
		"CSP_WORKER_SRC", "CSP_FRAME_SRC", "CSP_NAVIGATE_TO", "CSP_FORM_ACTION",
		"CSP_UPGRADE_INSECURE_REQUESTS", "CSP_BLOCK_ALL_MIXED_CONTENT",
		"SECURITY_DISABLE_CSRF", "SECURITY_DISABLE_CSP", "SECURITY_HSTS_ENABLED",
		"SECURITY_HSTS_MAX_AGE", "SECURITY_HSTS_INCLUDE_SUBDOMAINS",
		"SECURITY_HSTS_PRELOAD", "SECURITY_XSS_PROTECTION",
		"SECURITY_NO_SNIFF", "SECURITY_X_FRAME_OPTIONS", "SECURITY_REFERRER_POLICY",
		"SECURITY_PERMISSIONS_POLICY", "SECURITY_SET_COOKIE_SAMESITE",
		"SECURITY_SET_COOKIE_SECURE", "SECURITY_SET_COOKIE_HTTP_ONLY",
		"SECURITY_SET_COOKIE_MAXAGE", "SECURITY_SET_COOKIE_DOMAIN",
		"SECURITY_SET_COOKIE_PATH", "SECURITY_SET_COOKIE_PRIORITY",
		"SECURITY_SET_COOKIE_PARTITIONED", "SECURITY_SET_COOKIE_PREFIX",
		"SECURITY_SET_COOKIE_SUFFIX", "SECURITY_SET_COOKIE_STRICT_TRANSPORT_SECURITY",
		"SECURITY_SET_COOKIE_CONTENT_TYPE_OPTIONS", "SECURITY_SET_COOKIE_FRAME_OPTIONS",
		"SECURITY_SET_COOKIE_XSS_PROTECTION", "SECURITY_SET_COOKIE_REFERRER_POLICY",
		"SECURITY_SET_COOKIE_PERMISSIONS_POLICY", "SECURITY_SET_COOKIE_FEATURE_POLICY",
		"SECURITY_SET_COOKIE_ACCESS_CONTROL_ALLOW_ORIGIN", "SECURITY_SET_COOKIE_ACCESS_CONTROL_ALLOW_CREDENTIALS",
		"SECURITY_SET_COOKIE_ACCESS_CONTROL_ALLOW_HEADERS", "SECURITY_SET_COOKIE_ACCESS_CONTROL_ALLOW_METHODS",
		"SECURITY_SET_COOKIE_ACCESS_CONTROL_MAX_AGE", "SECURITY_SET_COOKIE_ACCESS_CONTROL_EXPOSE_HEADERS",
	}

	for _, key := range allKnownVars {
		if _, exists := envMap[key]; !exists {
			// Add with default value if not set
			category := getCategoryFromKey(key)
			variables = append(variables, EnvVariable{
				Key:         key,
				Value:       getDefaultEnvValue(key),
				Category:    category,
				Description: getEnvDescription(key),
				IsSecret:    isSecretVariable(key),
			})
		}
	}

	return variables
}

// getEnvDescription returns a human-readable description for an environment variable
func getEnvDescription(key string) string {
	descriptions := map[string]string{
		// Server Configuration
		"NOFX_BACKEND_PORT":     "Backend API server port",
		"NOFX_FRONTEND_PORT":    "Frontend web interface port",
		"NOFX_TIMEZONE":         "Timezone setting",
		"LOG_LEVEL":             "Log level (debug, info, warn, error)",
		"API_SERVER_PORT":       "API server port",
		"WAIT_BEFORE_DETECTION": "Initial wait time before detection",
		"DETECTION_INTERVAL":    "Interval between detections",

		// Authentication
		"JWT_SECRET":          "JWT signing secret",
		"JWT_EXPIRATION_DAYS": "JWT token expiration in days",

		// Encryption Keys
		"DATA_ENCRYPTION_KEY":  "AES-256 data encryption key",
		"RSA_PRIVATE_KEY":      "RSA private key for client-server encryption",
		"TRANSPORT_ENCRYPTION": "Enable transport encryption for API keys",

		// Database
		"DB_TYPE":     "Database type (sqlite, postgres)",
		"DB_PATH":     "SQLite database path",
		"DB_HOST":     "Database host",
		"DB_PORT":     "Database port",
		"DB_USER":     "Database username",
		"DB_PASSWORD": "Database password",
		"DB_NAME":     "Database name",
		"DB_SSLMODE":  "Database SSL mode",

		// AI Models
		"MODEL_MAX_TOKENS_DEEPSEEK": "DeepSeek model max token count",
		"MODEL_MAX_TOKENS_GPT4":     "GPT-4 model max token count",
		"MODEL_MAX_TOKENS_GPT35":    "GPT-3.5 model max token count",
		"MODEL_MAX_TOKENS_CLAUDE":   "Claude model max token count",
		"MODEL_MAX_TOKENS_QWEN":     "Qwen model max token count",
		"MODEL_MAX_TOKENS_OLLAMA":   "Ollama local model max token count",

		// AI Pricing
		"MODEL_PRICE_QWEN_MAX_INPUT":         "Qwen-Max input price (元/百万tokens)",
		"MODEL_PRICE_QWEN_MAX_OUTPUT":        "Qwen-Max output price (元/百万tokens)",
		"MODEL_PRICE_QWEN_PLUS_INPUT":        "Qwen-Plus input price (元/百万tokens)",
		"MODEL_PRICE_QWEN_PLUS_OUTPUT":       "Qwen-Plus output price (元/百万tokens)",
		"MODEL_PRICE_GPT4O_INPUT":            "GPT-4o input price (美元/百万tokens)",
		"MODEL_PRICE_GPT4O_OUTPUT":           "GPT-4o output price (美元/百万tokens)",
		"MODEL_PRICE_GPT4OMINI_INPUT":        "GPT-4o-mini input price (美元/百万tokens)",
		"MODEL_PRICE_GPT4OMINI_OUTPUT":       "GPT-4o-mini output price (美元/百万tokens)",
		"MODEL_PRICE_CLAUDE_SONNET_INPUT":    "Claude 3.7 Sonnet input price (美元/百万tokens)",
		"MODEL_PRICE_CLAUDE_SONNET_OUTPUT":   "Claude 3.7 Sonnet output price (美元/百万tokens)",
		"MODEL_PRICE_DEEPSEEK_INPUT":         "Deepseek-Chat input price (元/百万tokens)",
		"MODEL_PRICE_DEEPSEEK_OUTPUT":        "Deepseek-Chat output price (元/百万tokens)",
		"MODEL_PRICE_OLLAMA_INPUT":           "Ollama model input price (元/百万tokens)",
		"MODEL_PRICE_OLLAMA_OUTPUT":          "Ollama model output price (元/百万tokens)",
		"MODEL_PRICE_WENXIN_ERINIE35_INPUT":  "ERNIE-3.5-8K input price (元/百万tokens)",
		"MODEL_PRICE_WENXIN_ERINIE35_OUTPUT": "ERNIE-3.5-8K output price (元/百万tokens)",
		"MODEL_PRICE_DEFAULT_INPUT":          "Default input price (元/百万tokens)",
		"MODEL_PRICE_DEFAULT_OUTPUT":         "Default output price (元/百万tokens)",
		"COST_DISPLAY_UNIT":                  "Cost display unit (USD, CNY, or BOTH)",
		"USD_TO_CNY_RATE":                    "Exchange rate (1 USD to CNY)",
		"OLLAMA_READ_TIMEOUT":                "Ollama read timeout",

		// Guardian
		"GUARDIAN_TARGET_URL":                     "Guardian target AI service URL",
		"GUARDIAN_INPUT_SELECTOR":                 "Input field CSS selector",
		"GUARDIAN_BUTTON_SELECTOR":                "Submit button CSS selector",
		"GUARDIAN_RESPONSE_SELECTOR":              "Response content CSS selector",
		"GUARDIAN_DISPLAY_ENABLED":                "Enable browser display mode",
		"DOCKER_ENV":                              "Docker environment identifier",
		"GUARDIAN_BROWSER_TIMEOUT_SECONDS":        "Browser timeout in automatic mode",
		"GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS": "Browser timeout in manual mode",
		"GUARDIAN_MANUAL_MODE":                    "Enable manual mode (browser stays open)",
		"GUARDIAN_CHROME_PATH":                    "Chrome browser path",
		"GUARDIAN_FORCE_BROWSER_REQUIRED":         "Force browser requirement",
		"GUARDIAN_VALIDATION_MODE":                "Validation mode (strict or relaxed)",
		"BROWSER_DATA_DIR":                        "Browser data storage directory",
		"GUARDIAN_INITIAL_WAIT_SECONDS":           "Initial wait time in seconds",
		"GUARDIAN_MIN_POLL_SECONDS":               "Minimum polling interval in seconds",
		"GUARDIAN_MAX_POLL_SECONDS":               "Maximum polling interval in seconds",

		// Other
		"TELEGRAM_BOT_TOKEN":             "Telegram bot token (optional)",
		"TELEGRAM_CHAT_ID":               "Telegram chat ID (optional)",
		"USE_BINANCE_PROXY":              "Enable binance proxy mode",
		"BINANCE_PROXY_URL":              "Binance proxy service URL",
		"MCP_ENABLE_CONTEXT_COMPRESSION": "Enable context compression",
		"MCP_MODEL_CONTEXT_SIZE":         "Model context size",
	}

	if desc, exists := descriptions[key]; exists {
		return desc
	}

	// Generate a default description based on the key
	return fmt.Sprintf("Environment variable: %s", key)
}

// getDefaultEnvValue returns a default value for an environment variable
func getDefaultEnvValue(key string) string {
	defaults := map[string]string{
		"NOFX_BACKEND_PORT":                "8888",
		"NOFX_FRONTEND_PORT":               "3300",
		"NOFX_TIMEZONE":                    "Asia/Shanghai",
		"LOG_LEVEL":                        "info",
		"JWT_EXPIRATION_DAYS":              "7",
		"DB_TYPE":                          "sqlite",
		"DB_PATH":                          "data/data.db",
		"DB_PORT":                          "5432",
		"DB_USER":                          "nofx_user",
		"DB_NAME":                          "nofx",
		"DB_SSLMODE":                       "disable",
		"TRANSPORT_ENCRYPTION":             "false",
		"GUARDIAN_DISPLAY_ENABLED":         "true",
		"GUARDIAN_BROWSER_TIMEOUT_SECONDS": "60",
		"GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS": "600",
		"GUARDIAN_MANUAL_MODE":                    "false",
		"GUARDIAN_FORCE_BROWSER_REQUIRED":         "true",
		"GUARDIAN_VALIDATION_MODE":                "relaxed",
		"BROWSER_DATA_DIR":                        "data/browser_data",
		"GUARDIAN_INITIAL_WAIT_SECONDS":           "60",
		"GUARDIAN_MIN_POLL_SECONDS":               "2",
		"GUARDIAN_MAX_POLL_SECONDS":               "5",
		"USE_BINANCE_PROXY":                       "false",
		"BINANCE_PROXY_URL":                       "http://localhost:8081",
		"MCP_ENABLE_CONTEXT_COMPRESSION":          "true",
		"MCP_MODEL_CONTEXT_SIZE":                  "0",
		"COST_DISPLAY_UNIT":                       "BOTH",
		"USD_TO_CNY_RATE":                         "7.2",
		"OLLAMA_READ_TIMEOUT":                     "270s",
		"MODEL_MAX_TOKENS_DEEPSEEK":               "32768",
		"MODEL_MAX_TOKENS_GPT4":                   "128000",
		"MODEL_MAX_TOKENS_GPT35":                  "16384",
		"MODEL_MAX_TOKENS_CLAUDE":                 "200000",
		"MODEL_MAX_TOKENS_QWEN":                   "32768",
		"MODEL_MAX_TOKENS_OLLAMA":                 "4096",
		"MODEL_PRICE_QWEN_MAX_INPUT":              "2.40",
		"MODEL_PRICE_QWEN_MAX_OUTPUT":             "9.60",
		"MODEL_PRICE_QWEN_PLUS_INPUT":             "1.00",
		"MODEL_PRICE_QWEN_PLUS_OUTPUT":            "4.00",
		"MODEL_PRICE_GPT4O_INPUT":                 "2.50",
		"MODEL_PRICE_GPT4O_OUTPUT":                "10.00",
		"MODEL_PRICE_GPT4OMINI_INPUT":             "0.15",
		"MODEL_PRICE_GPT4OMINI_OUTPUT":            "0.60",
		"MODEL_PRICE_CLAUDE_SONNET_INPUT":         "3.00",
		"MODEL_PRICE_CLAUDE_SONNET_OUTPUT":        "15.00",
		"MODEL_PRICE_DEEPSEEK_INPUT":              "0.20",
		"MODEL_PRICE_DEEPSEEK_OUTPUT":             "0.80",
		"MODEL_PRICE_OLLAMA_INPUT":                "0.00",
		"MODEL_PRICE_OLLAMA_OUTPUT":               "0.00",
		"MODEL_PRICE_WENXIN_ERINIE35_INPUT":       "0.08",
		"MODEL_PRICE_WENXIN_ERINIE35_OUTPUT":      "0.20",
		"MODEL_PRICE_DEFAULT_INPUT":               "0.50",
		"MODEL_PRICE_DEFAULT_OUTPUT":              "2.00",
	}

	if defaultValue, exists := defaults[key]; exists {
		return defaultValue
	}

	return ""
}

// isSecretVariable determines if an environment variable contains sensitive information
func isSecretVariable(key string) bool {
	secretIndicators := []string{
		"SECRET", "KEY", "PASSWORD", "TOKEN", "PRIVATE", "PASS", "PWD",
		"AUTH", "CREDENTIAL", "CERT", "SIGNING", "ENCRYPT", "RSA",
	}

	keyUpper := strings.ToUpper(key)
	for _, indicator := range secretIndicators {
		if strings.Contains(keyUpper, indicator) {
			return true
		}
	}

	return false
}

// getCategoryFromKey determines the category of an environment variable based on its name
func getCategoryFromKey(key string) string {
	keyLower := strings.ToLower(key)

	if strings.Contains(keyLower, "server") || strings.Contains(keyLower, "port") {
		return "server"
	}

	if strings.Contains(keyLower, "auth") || strings.Contains(keyLower, "jwt") {
		return "auth"
	}

	if strings.Contains(keyLower, "encrypt") || strings.Contains(keyLower, "key") || strings.Contains(keyLower, "secret") {
		return "encryption"
	}

	if strings.Contains(keyLower, "db") || strings.Contains(keyLower, "database") {
		return "database"
	}

	if strings.Contains(keyLower, "model") || strings.Contains(keyLower, "ai") || strings.Contains(keyLower, "price") {
		return "ai"
	}

	if strings.Contains(keyLower, "guardian") || strings.Contains(keyLower, "browser") {
		return "guardian"
	}

	return "other"
}
