package config

// GuardianConfig 守护程序配置
type GuardianConfig struct {
	// 浏览器配置
	BrowserConfig *BrowserConfig

	// 数据传输配置
	DataTransferConfig *DataTransferConfig

	// 监听配置
	ListenInterval int // 监听间隔（毫秒）

	// 日志配置
	LogLevel    string
	LogFilePath string

	// AI服务配置
	AIProvider string // AI服务提供商，如 "deepseek", "chatgpt" 等
	AIEndpoint string // AI服务端点URL
	PageURL    string // AI网页界面URL

	// 超时配置
	RequestTimeout  int // 请求超时时间（秒）
	ResponseTimeout int // 响应超时时间（秒）

	// 重试配置
	MaxRetries int // 最大重试次数
	RetryDelay int // 重试延迟（毫秒）
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	ExecutablePath string   // 浏览器执行路径
	Headless       bool     // 是否无头模式
	UserDataDir    string   // 用户数据目录
	WindowSize     string   // 窗口大小，格式："width,height"
	Proxy          string   // 代理设置
	DisableImages  bool     // 是否禁用图片加载
	DisableJS      bool     // 是否禁用JavaScript（通常不推荐）
	AdditionalArgs []string // 额外的浏览器启动参数
}

// DataTransferConfig 数据传输配置
type DataTransferConfig struct {
	// NoFx API配置
	NofxAPIEndpoint string // NoFx API端点
	APIKey          string // API认证密钥
	AuthToken       string // 认证令牌

	// 传输超时
	TransferTimeout int // 数据传输超时（秒）

	// 重试机制
	MaxTransferRetries int // 最大传输重试次数
	TransferRetryDelay int // 传输重试延迟（毫秒）

	// 数据格式
	DataFormat string // 数据格式，如 "json", "text" 等
}

// NewDefaultGuardianConfig 创建默认配置
func NewDefaultGuardianConfig() *GuardianConfig {
	return &GuardianConfig{
		ListenInterval:  5000, // 5秒监听间隔
		LogLevel:        "info",
		LogFilePath:     "./guardian.log",
		AIProvider:      "deepseek",
		AIEndpoint:      "https://chat.deepseek.com",
		PageURL:         "https://chat.deepseek.com",
		RequestTimeout:  30,   // 30秒请求超时
		ResponseTimeout: 120,  // 2分钟响应超时
		MaxRetries:      3,    // 最多重试3次
		RetryDelay:      2000, // 重试延迟2秒

		BrowserConfig: &BrowserConfig{
			ExecutablePath: "", // 空字符串表示使用默认路径
			Headless:       false,
			UserDataDir:    "",
			WindowSize:     "1920,1080",
			Proxy:          "",
			DisableImages:  false,
			DisableJS:      false,
			AdditionalArgs: []string{},
		},

		DataTransferConfig: &DataTransferConfig{
			NofxAPIEndpoint:    "http://localhost:8888/api",
			APIKey:             "",
			AuthToken:          "",
			TransferTimeout:    30,
			MaxTransferRetries: 3,
			TransferRetryDelay: 2000,
			DataFormat:         "json",
		},
	}
}
