package dataprovider

// Kline 表示K线数据结构
type Kline struct {
	OpenTime                 int64   `json:"openTime"`
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	Volume                   float64 `json:"volume"`
	CloseTime                int64   `json:"closeTime"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume"`
	NumberOfTrades           int64   `json:"numberOfTrades"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume"`
}

// DataProvider 数据提供者接口定义，用于实现低耦合设计
// 币安代理服务插件，专门负责绕过币安监管，从主程序获取连接参数，专门负责获取数据、发送数据和执行命令
type DataProvider interface {
	GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error)
	GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error)
	GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error)
	GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error)
	GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error)
}

// DirectDataProvider 直接数据提供者（原实现）
type DirectDataProvider struct {
}

// NewDirectDataProvider 创建直接数据提供者
func NewDirectDataProvider() *DirectDataProvider {
	return &DirectDataProvider{}
}

// GetBalance 通过直接API调用获取余额
func (d *DirectDataProvider) GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error) {
	// 在纯透传架构中，这个方法应该由原始交易者实现
	// 这里保留接口定义但不实现具体逻辑
	return nil, nil
}

// GetPositions 通过直接API调用获取持仓
func (d *DirectDataProvider) GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error) {
	// 在纯透传架构中，这个方法应该由原始交易者实现
	// 这里保留接口定义但不实现具体逻辑
	return nil, nil
}

// GetKlines 通过直接API调用获取K线数据
func (d *DirectDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error) {
	// 在纯透传架构中，这个方法应该由原始交易者实现
	// 这里保留接口定义但不实现具体逻辑
	return nil, nil
}

// GetAccountInfo 通过直接API调用获取账户信息
func (d *DirectDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 在纯透传架构中，这个方法应该由原始交易者实现
	// 这里保留接口定义但不实现具体逻辑
	return nil, nil
}

// GetTrades 通过直接API调用获取交易历史
func (d *DirectDataProvider) GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 在纯透传架构中，这个方法应该由原始交易者实现
	// 这里保留接口定义但不实现具体逻辑
	return nil, nil
}

// GetDataProviderFromEnv 从环境变量获取数据提供者实例
func GetDataProviderFromEnv() DataProvider {
	// 在纯透传架构中，我们不再使用环境变量控制代理模式
	// 所有数据提供者逻辑由原始交易者和ProxyTraderWrapper处理
	return NewDirectDataProvider()
}
