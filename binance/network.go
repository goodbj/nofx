package binance

import (
	"fmt"
	"strings"
	"time"
)

// TestnetConfig testnet configuration utilities
type TestnetConfig struct {
	// Testnet endpoints
	MainnetAPIURL    string `json:"mainnet_api_url"`
	TestnetAPIURL    string `json:"testnet_api_url"`
	MainnetWebsocket string `json:"mainnet_websocket"`
	TestnetWebsocket string `json:"testnet_websocket"`

	// Testnet specific settings
	AllowTestFunds   bool   `json:"allow_test_funds"`
	TestAPIKeyPrefix string `json:"test_api_key_prefix"`
}

// NewTestnetConfig creates default testnet configuration
func NewTestnetConfig() *TestnetConfig {
	return &TestnetConfig{
		MainnetAPIURL:    "https://fapi.binance.com",
		TestnetAPIURL:    "https://testnet.binancefuture.com",
		MainnetWebsocket: "wss://fstream.binance.com/ws",
		TestnetWebsocket: "wss://stream.binancefuture.com/ws",
		AllowTestFunds:   true,
		TestAPIKeyPrefix: "TEST",
	}
}

// GetAPIURL returns appropriate API URL based on testnet setting
func (tc *TestnetConfig) GetAPIURL(testnet bool) string {
	if testnet {
		return tc.TestnetAPIURL
	}
	return tc.MainnetAPIURL
}

// GetWebsocketURL returns appropriate websocket URL based on testnet setting
func (tc *TestnetConfig) GetWebsocketURL(testnet bool) string {
	if testnet {
		return tc.TestnetWebsocket
	}
	return tc.MainnetWebsocket
}

// ValidateAPIKey validates if API key is appropriate for the network
func (tc *TestnetConfig) ValidateAPIKey(apiKey string, testnet bool) error {
	if testnet {
		// In testnet, we're more permissive but can still check format
		if len(apiKey) < 10 {
			return fmt.Errorf("testnet API key appears too short")
		}
		// Testnet keys might have different format requirements
		return nil
	}

	// For mainnet, be more strict
	if len(apiKey) != 64 {
		return fmt.Errorf("mainnet API key must be 64 characters long, got %d", len(apiKey))
	}

	// Check if it looks like a testnet key
	if strings.HasPrefix(apiKey, tc.TestAPIKeyPrefix) || strings.Contains(strings.ToLower(apiKey), "test") {
		return fmt.Errorf("testnet API key used for mainnet connection")
	}

	return nil
}

// ValidateSecretKey validates secret key format
func (tc *TestnetConfig) ValidateSecretKey(secretKey string, testnet bool) error {
	if len(secretKey) != 64 {
		return fmt.Errorf("secret key must be 64 characters long, got %d", len(secretKey))
	}

	// Additional validation could be added here
	return nil
}

// NetworkValidator validates network-specific requirements
type NetworkValidator struct {
	config *TestnetConfig
}

// NewNetworkValidator creates a new network validator
func NewNetworkValidator() *NetworkValidator {
	return &NetworkValidator{
		config: NewTestnetConfig(),
	}
}

// IsTestnetURL determines if a URL is a testnet URL
func (tc *TestnetConfig) IsTestnetURL(url string) bool {
	lowerURL := strings.ToLower(url)
	return strings.Contains(lowerURL, "testnet") ||
		strings.Contains(lowerURL, "test") ||
		url == "https://testnet.binancefuture.com" ||
		url == "https://stream.binancefuture.com"
}

// ValidateExchangeConfig validates exchange configuration for network appropriateness
func (nv *NetworkValidator) ValidateExchangeConfig(config *Config, testnetExpected bool) error {
	// Check if testnet setting matches expected
	isActuallyTestnet := nv.IsTestnetURL(config.BaseURL)

	if testnetExpected && !isActuallyTestnet {
		return fmt.Errorf("exchange configured for mainnet but testnet expected. URL: %s", config.BaseURL)
	}

	if !testnetExpected && isActuallyTestnet {
		return fmt.Errorf("exchange configured for testnet but mainnet expected. URL: %s", config.BaseURL)
	}

	// Validate API keys are appropriate for the network
	if err := nv.config.ValidateAPIKey(config.APIKey, testnetExpected); err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	if err := nv.config.ValidateSecretKey(config.SecretKey, testnetExpected); err != nil {
		return fmt.Errorf("secret key validation failed: %w", err)
	}

	return nil
}

// IsTestnetURL determines if a URL is a testnet URL
func (nv *NetworkValidator) IsTestnetURL(url string) bool {
	lowerURL := strings.ToLower(url)
	return strings.Contains(lowerURL, "testnet") ||
		strings.Contains(lowerURL, "test") ||
		url == "https://testnet.binancefuture.com" ||
		url == "https://stream.binancefuture.com"
}

// GetRecommendedRateLimits returns recommended rate limits for the network
func (nv *NetworkValidator) GetRecommendedRateLimits(testnet bool) (requestsPerMinute, burst int) {
	if testnet {
		// Testnet typically allows higher rate limits
		return 2400, 20 // 2400 requests per minute, burst of 20
	}
	// Mainnet conservative limits
	return 1200, 10 // 1200 requests per minute, burst of 10
}

// NetworkAwareClient network-aware Binance client wrapper
type NetworkAwareClient struct {
	underlyingClient interface{} // This would be the actual Binance client
	config           *Config
	testnetConfig    *TestnetConfig
	validator        *NetworkValidator
}

// NewNetworkAwareClient creates a new network-aware client
func NewNetworkAwareClient(config *Config) *NetworkAwareClient {
	testnetConfig := NewTestnetConfig()
	validator := NewNetworkValidator()

	return &NetworkAwareClient{
		config:        config,
		testnetConfig: testnetConfig,
		validator:     validator,
	}
}

// Validate validates the client configuration
func (nac *NetworkAwareClient) Validate() error {
	expectedTestnet := nac.testnetConfig.IsTestnetURL(nac.config.BaseURL)
	return nac.validator.ValidateExchangeConfig(nac.config, expectedTestnet)
}

// GetNetworkInfo returns network information
func (nac *NetworkAwareClient) GetNetworkInfo() map[string]interface{} {
	isTestnet := nac.testnetConfig.IsTestnetURL(nac.config.BaseURL)

	return map[string]interface{}{
		"is_testnet":    isTestnet,
		"api_url":       nac.config.BaseURL,
		"websocket_url": nac.testnetConfig.GetWebsocketURL(isTestnet),
		"recommended_rate_limit": func() map[string]int {
			rpm, burst := nac.validator.GetRecommendedRateLimits(isTestnet)
			return map[string]int{
				"requests_per_minute": rpm,
				"burst":               burst,
			}
		}(),
		"validation_status": func() string {
			if err := nac.Validate(); err != nil {
				return fmt.Sprintf("INVALID: %v", err)
			}
			return "VALID"
		}(),
		"timestamp": time.Now(),
	}
}

// SwitchNetwork switches the client to a different network
func (nac *NetworkAwareClient) SwitchNetwork(toTestnet bool) error {
	oldNetwork := nac.testnetConfig.IsTestnetURL(nac.config.BaseURL)

	if oldNetwork == toTestnet {
		return fmt.Errorf("already on %s network", map[bool]string{true: "testnet", false: "mainnet"}[toTestnet])
	}

	// Update configuration
	nac.config.BaseURL = nac.testnetConfig.GetAPIURL(toTestnet)
	nac.config.Testnet = toTestnet

	// Revalidate
	return nac.Validate()
}
