package mcp

import "strings"

// IsValidAIEndpointURL checks if the URL is a valid AI endpoint and not another service
func IsValidAIEndpointURL(url string) bool {
	invalidDomains := []string{
		"binancefuture.com",
		"bybit.com",
		"okx.com",
		"hyperliquid.xyz",
		"astergateway.com",
		"testnet.binancefuture.com",
		"api.bybit.com",
		"www.okx.com",
		"api.hyperliquid.xyz",
	}

	for _, domain := range invalidDomains {
		if strings.Contains(url, domain) {
			return false
		}
	}
	return true
}
