package config

// PortConfig 集中管理所有服务端口配置
type PortConfig struct {
	// 前端服务端口
	FrontendPort string

	// 后端服务端口
	BackendPort string

	// 币安代理服务端口
	BinanceProxyPort string

	// 币安代理容器内部端口
	BinanceProxyInternalPort string

	// Guardian服务端口
	GuardianPort string

	// 数据库端口
	DBPort string

	// Redis端口
	RedisPort string
}

// GetPortConfig 获取端口配置
func GetPortConfig() *PortConfig {
	return &PortConfig{
		FrontendPort:             "3300", // 前端服务
		BackendPort:              "8888", // 后端服务
		BinanceProxyPort:         "8081", // 币安代理服务对外端口
		BinanceProxyInternalPort: "8082", // 币安代理服务容器内端口
		GuardianPort:             "8083", // Guardian服务
		DBPort:                   "5432", // PostgreSQL默认端口
		RedisPort:                "6379", // Redis默认端口
	}
}

// GetBinanceProxyURL 获取币安代理服务URL
func GetBinanceProxyURL() string {
	portConfig := GetPortConfig()
	return "http://localhost:" + portConfig.BinanceProxyPort
}
