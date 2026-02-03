package main

import (
	"os"
)

// ProxyConfig 代理服务配置
type ProxyConfig struct {
	Port             string
	BinanceAPIKey    string
	BinanceSecretKey string
	CustomAPIURL     string
	Debug            bool
	DefaultTargetURL string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *ProxyConfig {
	config := &ProxyConfig{
		Port:             getEnvOrDefault("PORT", "8081"),
		BinanceAPIKey:    os.Getenv("BINANCE_API_KEY"),
		BinanceSecretKey: os.Getenv("BINANCE_SECRET_KEY"),
		CustomAPIURL:     os.Getenv("BINANCE_CUSTOM_API_URL"),
		Debug:            os.Getenv("DEBUG") == "true",
		DefaultTargetURL: getEnvOrDefault("DEFAULT_TARGET_API_URL", "https://testnet.binancefuture.com"),
	}

	return config
}

