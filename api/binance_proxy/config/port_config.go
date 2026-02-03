package config

// PortConfig 端口配置
type PortConfig struct {
	ProxyPort string // 代理服务端口
}

// DefaultPortConfig 默认端口配置
var DefaultPortConfig = &PortConfig{
	ProxyPort: "8081", // 默认代理端口
}

// GetProxyPort 获取代理服务端口
func GetProxyPort() string {
	return DefaultPortConfig.ProxyPort
}
