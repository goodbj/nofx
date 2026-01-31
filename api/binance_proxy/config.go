package main

import (
	"os"
)

// Config 代理服务配置
type Config struct {
	Port             string
	BinanceAPIKey    string
	BinanceSecretKey string
	CustomAPIURL     string
	Debug            bool
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{
		Port:             getEnvOrDefault("PORT", "8082"),
		BinanceAPIKey:    os.Getenv("BINANCE_API_KEY"),
		BinanceSecretKey: os.Getenv("BINANCE_SECRET_KEY"),
		CustomAPIURL:     os.Getenv("BINANCE_CUSTOM_API_URL"),
		Debug:            os.Getenv("DEBUG") == "true",
	}

	return config
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
