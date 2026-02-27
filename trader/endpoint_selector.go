package trader

import (
	"nofx/logger"
	"os"
	"strings"
)

// EndpointSelector统一的API端点选择器
type EndpointSelector struct{}

// NewEndpointSelector 创建新的端点选择器
func NewEndpointSelector() *EndpointSelector {
	return &EndpointSelector{}
}

// GetBinanceMainnetEndpoint 为实盘Binance获取正确的API端点
func (es *EndpointSelector) GetBinanceMainnetEndpoint(customAPIURL string) string {
	logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Input customAPIURL: '%s'", customAPIURL)

	// 检查是否为代理URL（不包含testnet相关关键词，避免误判）
	isProxyURL := (strings.Contains(customAPIURL, "localhost") ||
		strings.Contains(customAPIURL, "127.0.0.1")) &&
		!strings.Contains(strings.ToLower(customAPIURL), "testnet") &&
		!strings.Contains(strings.ToLower(customAPIURL), "test")

	logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] isProxyURL (excluding testnet URLs): %t", isProxyURL)

	if isProxyURL {
		logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] URL is proxy (but not testnet), for real account: %s", customAPIURL)
		logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Real account proxy URL does not contain testnet/test, returning mainnet endpoint")
		return "https://fapi.binance.com"
	}

	// 如果有自定义URL，需要根据内容进行标准化
	logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Processing custom URL for real account: %s", customAPIURL)
	if customAPIURL != "" {
		// 检查URL是否包含testnet相关字样 - 对于实盘账户，即使包含也应返回主网
		containsTestnet := strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")

		logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Real account URL contains testnet/test: %t", containsTestnet)

		if containsTestnet {
			// 对于实盘账户，即使URL包含testnet也返回主网
			logger.Warnf("⚠️ Real account URL contains testnet-like string: %s, but returning mainnet endpoint for safety", customAPIURL)
			return "https://fapi.binance.com"
		} else {
			// 对于非testnet的自定义URL，直接返回
			logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Real account non-testnet custom URL, returning as-is: %s", customAPIURL)
			return customAPIURL
		}
	}

	// 默认返回主网API
	logger.Debugf("🔗 [EndpointSelector.GetBinanceMainnetEndpoint] Real account CustomAPIURL is empty, returning default mainnet endpoint")
	return "https://fapi.binance.com"
}

// GetBinanceDemoEndpoint 为虚拟盘Binance获取正确的API端点
func (es *EndpointSelector) GetBinanceDemoEndpoint(customAPIURL string) string {
	logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Input customAPIURL: '%s'", customAPIURL)

	// 检查是否为代理URL（不包含testnet相关关键词，避免误判）
	isProxyURL := (strings.Contains(customAPIURL, "localhost") ||
		strings.Contains(customAPIURL, "127.0.0.1")) &&
		!strings.Contains(strings.ToLower(customAPIURL), "testnet") &&
		!strings.Contains(strings.ToLower(customAPIURL), "test")

	logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo isProxyURL (excluding testnet URLs): %t", isProxyURL)

	if isProxyURL {
		logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo URL is proxy (but not testnet), checking for testnet keywords in: %s", customAPIURL)
		// 如果是代理URL，根据实际的CustomAPIURL来决定目标端点
		// 如果CustomAPIURL本身包含testnet相关字样，则使用测试网，否则使用默认测试网
		if customAPIURL != "" && (strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")) {
			logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo Proxy URL contains testnet/test, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		}
		logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo Proxy URL does not contain testnet/test, returning default testnet endpoint")
		return "https://testnet.binancefuture.com"
	}

	// 如果有自定义URL，需要根据内容进行标准化
	logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Processing custom URL for demo account: %s", customAPIURL)
	if customAPIURL != "" {
		// 检查URL是否包含testnet相关字样
		containsTestnet := strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")

		logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo URL contains testnet/test: %t", containsTestnet)

		// 明确区分是否为官方测试网URL
		isOfficialTestnet := containsTestnet &&
			(strings.Contains(customAPIURL, "binancefuture.com") ||
				strings.Contains(customAPIURL, "binanceus.com"))

		logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Is official demo testnet URL: %t", isOfficialTestnet)

		if isOfficialTestnet {
			logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Official demo testnet URL detected, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		} else if containsTestnet {
			// 如果URL包含testnet但不是官方测试网URL，返回默认测试网
			logger.Warnf("⚠️ Non-official demo testnet-like URL detected: %s, returning default testnet endpoint", customAPIURL)
			return "https://testnet.binancefuture.com"
		} else {
			// 对于非testnet的自定义URL，返回默认测试网（因为这是demo账户）
			logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo account non-testnet custom URL, returning default testnet endpoint")
			return "https://testnet.binancefuture.com"
		}
	}

	// 默认返回测试网API
	logger.Debugf("🔗 [EndpointSelector.GetBinanceDemoEndpoint] Demo account CustomAPIURL is empty, returning default testnet endpoint")
	return "https://testnet.binancefuture.com"
}

// GetBinanceEndpoint 为Binance获取正确的API端点
// 根据新规则：完全忽略Testnet开关，只看CustomAPIURL内容
// 但对某些特定情况进行标准化处理
// 已废弃：请使用 GetBinanceMainnetEndpoint 或 GetBinanceDemoEndpoint
func (es *EndpointSelector) GetBinanceEndpoint(customAPIURL string, isTestnet bool) string {
	logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Input customAPIURL: '%s', Input isTestnet: %t (should be ignored)", customAPIURL, isTestnet)

	// 检查是否为代理URL（不包含testnet相关关键词，避免误判）
	isProxyURL := (strings.Contains(customAPIURL, "localhost") ||
		strings.Contains(customAPIURL, "127.0.0.1")) &&
		!strings.Contains(strings.ToLower(customAPIURL), "testnet") &&
		!strings.Contains(strings.ToLower(customAPIURL), "test")

	logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] isProxyURL (excluding testnet URLs): %t", isProxyURL)

	if isProxyURL {
		logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] URL is proxy (but not testnet), checking for testnet keywords in: %s", customAPIURL)
		// 如果是代理URL，根据实际的CustomAPIURL来决定目标端点，而不是testnet标志
		// 如果CustomAPIURL本身包含testnet相关字样，则使用测试网，否则使用主网
		if customAPIURL != "" && (strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")) {
			logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Proxy URL contains testnet/test, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		}
		logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Proxy URL does not contain testnet/test, returning mainnet endpoint")
		return "https://fapi.binance.com"
	}

	// 如果有自定义URL，需要根据内容进行标准化
	logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Processing custom URL: %s", customAPIURL)
	if customAPIURL != "" {
		// 检查URL是否包含testnet相关字样
		containsTestnet := strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")

		logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] URL contains testnet/test: %t", containsTestnet)

		// 明确区分是否为官方测试网URL
		isOfficialTestnet := containsTestnet &&
			(strings.Contains(customAPIURL, "binancefuture.com") ||
				strings.Contains(customAPIURL, "binanceus.com"))

		logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Is official testnet URL: %t", isOfficialTestnet)

		if isOfficialTestnet {
			logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Official testnet URL detected, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		} else if containsTestnet {
			// 如果URL包含testnet但不是官方测试网URL，可能是误配，返回主网
			logger.Warnf("⚠️ Non-official testnet-like URL detected: %s, returning mainnet endpoint", customAPIURL)
			return "https://fapi.binance.com"
		} else {
			// 对于非testnet的自定义URL，直接返回
			logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] Non-testnet custom URL, returning as-is: %s", customAPIURL)
			return customAPIURL
		}
	}

	// 默认返回主网API，不再考虑isTestnet参数
	logger.Debugf("🔗 [EndpointSelector.GetBinanceEndpoint] CustomAPIURL is empty, returning default mainnet endpoint")
	return "https://fapi.binance.com"
}

// ShouldUseProxy检查是否应该使用代理
func (es *EndpointSelector) ShouldUseProxy() bool {
	return os.Getenv("USE_BINANCE_PROXY") == "true"
}

// GetProxyURL 获取代理URL
func (es *EndpointSelector) GetProxyURL() string {
	proxyURL := os.Getenv("BINANCE_PROXY_URL")
	if proxyURL == "" {
		proxyPort := os.Getenv("BINANCE_PROXY_PORT")
		if proxyPort == "" {
			proxyPort = "8081"
		}
		proxyURL = "http://localhost:" + proxyPort
	}
	return proxyURL
}

// GetUnifiedBinanceEndpoint 统一的端点选择器，智能判断URL类型并返回合适的端点
// 不需要知道是实盘还是虚拟盘，只根据URL内容进行处理
func (es *EndpointSelector) GetUnifiedBinanceEndpoint(customAPIURL string) string {
	logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Input customAPIURL: '%s'", customAPIURL)

	// 检查是否为代理URL（不包含testnet相关关键词，避免误判）
	isProxyURL := (strings.Contains(customAPIURL, "localhost") ||
		strings.Contains(customAPIURL, "127.0.0.1")) &&
		!strings.Contains(strings.ToLower(customAPIURL), "testnet") &&
		!strings.Contains(strings.ToLower(customAPIURL), "test")

	logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] isProxyURL (excluding testnet URLs): %t", isProxyURL)

	if isProxyURL {
		logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] URL is proxy, determining target based on URL content: %s", customAPIURL)
		// 如果是代理URL，根据实际的CustomAPIURL来决定目标端点
		if customAPIURL != "" && (strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")) {
			logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Proxy URL contains testnet/test, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		}
		logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Proxy URL does not contain testnet/test, returning mainnet endpoint")
		return "https://fapi.binance.com"
	}

	// 如果是直接URL，根据URL内容智能判断
	logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Processing direct URL: %s", customAPIURL)
	if customAPIURL != "" {
		// 检查URL是否包含testnet相关字样
		containsTestnet := strings.Contains(strings.ToLower(customAPIURL), "testnet") ||
			strings.Contains(strings.ToLower(customAPIURL), "test")

		logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] URL contains testnet/test: %t", containsTestnet)

		// 明确区分是否为官方测试网URL
		isOfficialTestnet := containsTestnet &&
			(strings.Contains(customAPIURL, "binancefuture.com") ||
				strings.Contains(customAPIURL, "binanceus.com"))

		logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Is official testnet URL: %t", isOfficialTestnet)

		if isOfficialTestnet {
			logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Official testnet URL detected, returning testnet endpoint")
			return "https://testnet.binancefuture.com"
		} else if containsTestnet {
			// 如果URL包含testnet但不是官方测试网URL，返回默认测试网
			logger.Warnf("⚠️ Non-official testnet-like URL detected: %s, returning default testnet endpoint", customAPIURL)
			return "https://testnet.binancefuture.com"
		} else {
			// 对于非testnet的自定义URL，直接返回
			logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] Non-testnet custom URL, returning as-is: %s", customAPIURL)
			return customAPIURL
		}
	}

	// 默认返回主网API
	logger.Debugf("🔗 [EndpointSelector.GetUnifiedBinanceEndpoint] CustomAPIURL is empty, returning default mainnet endpoint")
	return "https://fapi.binance.com"
}
