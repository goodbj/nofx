package config

import (
	"fmt"
	"os"
)

// GetPortFromEnvOrConfig 从环境变量获取端口，如果不存在则使用配置
func GetPortFromEnvOrConfig(envVarName, defaultPort string) string {
	envPort := os.Getenv(envVarName)
	if envPort != "" {
		return envPort
	}
	return defaultPort
}

// GetAllServicePorts 获取所有服务端口，支持环境变量覆盖
func GetAllServicePorts() map[string]string {
	config := GetPortConfig()

	return map[string]string{
		"FRONTEND_PORT":               GetPortFromEnvOrConfig("FRONTEND_PORT", config.FrontendPort),
		"BACKEND_PORT":                GetPortFromEnvOrConfig("BACKEND_PORT", config.BackendPort),
		"BINANCE_PROXY_PORT":          GetPortFromEnvOrConfig("BINANCE_PROXY_PORT", config.BinanceProxyPort),
		"BINANCE_PROXY_INTERNAL_PORT": GetPortFromEnvOrConfig("BINANCE_PROXY_INTERNAL_PORT", config.BinanceProxyInternalPort),
		"GUARDIAN_PORT":               GetPortFromEnvOrConfig("GUARDIAN_PORT", config.GuardianPort),
		"DB_PORT":                     GetPortFromEnvOrConfig("DB_PORT", config.DBPort),
		"REDIS_PORT":                  GetPortFromEnvOrConfig("REDIS_PORT", config.RedisPort),
	}
}

// GetBinanceProxyURLWithEnv 获取币安代理服务URL，支持环境变量覆盖
func GetBinanceProxyURLWithEnv() string {
	ports := GetAllServicePorts()
	return fmt.Sprintf("http://localhost:%s", ports["BINANCE_PROXY_PORT"])
}

// GetBackendURL 获取后端服务URL
func GetBackendURL() string {
	ports := GetAllServicePorts()
	return fmt.Sprintf("http://localhost:%s", ports["BACKEND_PORT"])
}

// GetFrontendURL 获取前端服务URL
func GetFrontendURL() string {
	ports := GetAllServicePorts()
	return fmt.Sprintf("http://localhost:%s", ports["FRONTEND_PORT"])
}
