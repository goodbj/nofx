package dataprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"nofx/config"
)

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
	// 直接模式下，暂时返回错误，因为需要完整实现
	return nil, fmt.Errorf("direct模式需要完整实现")
}

// GetPositions 通过直接API调用获取持仓
func (d *DirectDataProvider) GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error) {
	// 直接模式下，暂时返回错误，因为需要完整实现
	return nil, fmt.Errorf("direct模式需要完整实现")
}

// GetKlines 通过直接API调用获取K线数据
func (d *DirectDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error) {
	// 直接模式下，暂时返回错误，因为需要完整实现
	return nil, fmt.Errorf("direct K线数据获取功能暂未实现")
}

// GetAccountInfo 通过直接API调用获取账户信息
func (d *DirectDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 直接模式下，暂时返回错误，因为需要完整实现
	return nil, fmt.Errorf("direct模式需要完整实现")
}

// GetTrades 通过直接API调用获取交易历史
func (d *DirectDataProvider) GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 直接模式下，暂时返回错误，因为需要完整实现
	return nil, fmt.Errorf("direct 交易历史获取功能暂未实现")
}

// ProxyDataProvider 代理数据提供者（新实现）
type ProxyDataProvider struct {
	ProxyURL string
}

// NewProxyDataProvider 创建代理数据提供者
func NewProxyDataProvider(proxyURL string) *ProxyDataProvider {
	return &ProxyDataProvider{
		ProxyURL: proxyURL,
	}
}

// GetBalance 通过代理服务获取余额
func (p *ProxyDataProvider) GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/proxy/balance", p.ProxyURL)
	return p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
}

// GetPositions 通过代理服务获取持仓
func (p *ProxyDataProvider) GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error) {
	// 脱敏处理API Key
	maskedAPIKey := ""
	if len(apiKey) > 6 {
		maskedAPIKey = apiKey[:3] + "***" + apiKey[len(apiKey)-3:]
	} else {
		maskedAPIKey = apiKey
	}

	fmt.Printf("[DEBUG] GetPositions - Request to proxy - API Key: %s, Custom API URL: %s\n",
		maskedAPIKey, customAPIURL)

	url := fmt.Sprintf("%s/api/proxy/positions", p.ProxyURL)

	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		fmt.Printf("[DEBUG] GetPositions - Error making request: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] GetPositions - Raw response from proxy (truncated): %s\n", truncateString(fmt.Sprintf("%+v", response), 1000))

	// 直接返回原始响应中的数据，让后端服务使用其原有的处理逻辑
	positionsData, ok := response["data"]
	if !ok {
		fmt.Printf("[DEBUG] GetPositions - Response does not contain 'data' field: %+v\n", response)
		return nil, fmt.Errorf("response does not contain 'data' field, response: %+v", response)
	}

	positions, ok := positionsData.([]interface{})
	if !ok {
		// 添加调试信息，输出实际的数据类型
		fmt.Printf("[DEBUG] GetPositions - Positions data is not an array, actual type: %T, value: %+v\n", positionsData, positionsData)
		return nil, fmt.Errorf("positions data is not an array, actual type: %T, value: %+v", positionsData, positionsData)
	}

	// 检查是否有错误信息
	if errorMsg, exists := response["error"]; exists {
		return nil, fmt.Errorf("binance proxy error: %v", errorMsg)
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		if posMap, ok := pos.(map[string]interface{}); ok {
			result = append(result, posMap)
		}
	}

	fmt.Printf("[DEBUG] GetPositions - Final result count: %d, result: %+v\n", len(result), result)
	return result, nil
}

// GetKlines 通过代理服务获取K线数据
func (p *ProxyDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error) {
	url := fmt.Sprintf("%s/api/proxy/klines?symbol=%s&interval=%s&limit=%d", p.ProxyURL, symbol, interval, limit)

	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		return nil, err
	}

	// 将响应转换为Kline数组
	klinesData, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("response does not contain 'data' field, response: %+v", response)
	}

	klines, ok := klinesData.([]interface{})
	if !ok {
		// 添加调试信息，输出实际的数据类型
		return nil, fmt.Errorf("klines data is not an array, actual type: %T, value: %+v", klinesData, klinesData)
	}

	var result []Kline
	for _, k := range klines {
		if kMap, ok := k.(map[string]interface{}); ok {
			kline := Kline{
				OpenTime:  int64(kMap["openTime"].(float64)),
				Open:      kMap["open"].(float64),
				High:      kMap["high"].(float64),
				Low:       kMap["low"].(float64),
				Close:     kMap["close"].(float64),
				Volume:    kMap["volume"].(float64),
				CloseTime: int64(kMap["closeTime"].(float64)),
			}
			result = append(result, kline)
		}
	}

	return result, nil
}

// truncateString 用于截断长字符串，只显示前n个字符
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetAccountInfo 通过代理服务获取账户信息
func (p *ProxyDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 脱敏处理API Key
	maskedAPIKey := ""
	if len(apiKey) > 6 {
		maskedAPIKey = apiKey[:3] + "***" + apiKey[len(apiKey)-3:]
	} else {
		maskedAPIKey = apiKey
	}

	fmt.Printf("[DEBUG] GetAccountInfo - Request to proxy - API Key: %s, Custom API URL: %s\n",
		maskedAPIKey, customAPIURL)

	url := fmt.Sprintf("%s/api/proxy/account", p.ProxyURL)
	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		fmt.Printf("[DEBUG] GetAccountInfo - Error making request: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] GetAccountInfo - Raw response from proxy (truncated): %s\n", truncateString(fmt.Sprintf("%+v", response), 1000))

	// 提取data字段，与GetPositions等其他方法保持一致
	accountData, ok := response["data"]
	if !ok {
		fmt.Printf("[DEBUG] GetAccountInfo - Response does not contain 'data' field: %+v\n", response)
		return nil, fmt.Errorf("response does not contain 'data' field, response: %+v", response)
	}

	// 检查是否有错误信息
	if errorMsg, exists := response["error"]; exists {
		return nil, fmt.Errorf("binance proxy error: %v", errorMsg)
	}

	// 返回data部分的内容，让后端服务使用其原有的处理逻辑
	// 这样可以完全模仿nofx原生的处理逻辑
	accountInfo, ok := accountData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("account data is not an object, actual type: %T, value: %+v", accountData, accountData)
	}

	return accountInfo, nil
}

// GetTrades 通过代理服务获取交易历史
func (p *ProxyDataProvider) GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/proxy/trades?symbol=%s", p.ProxyURL, symbol)
	return p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
}

// makeRequest 发起HTTP请求到代理服务
func (p *ProxyDataProvider) makeRequest(url string, queryParams map[string]string, apiKey, secretKey, customAPIURL string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 添加认证头
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-Secret-Key", secretKey)
	if customAPIURL != "" {
		req.Header.Set("X-Custom-API-URL", customAPIURL)
	}

	// 添加查询参数
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// PostRequest 发起POST请求到代理服务
func (p *ProxyDataProvider) PostRequest(url string, payload interface{}, apiKey, secretKey, customAPIURL string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 添加认证头
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("X-Secret-Key", secretKey)
	if customAPIURL != "" {
		req.Header.Set("X-Custom-API-URL", customAPIURL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// Factory 函数，根据环境变量返回对应的数据提供者
func NewDataProvider(useProxy bool, proxyURL string) DataProvider {
	if useProxy {
		return NewProxyDataProvider(proxyURL)
	}
	return NewDirectDataProvider()
}

// GetDataProviderFromEnv 从环境变量获取数据提供者实例
func GetDataProviderFromEnv() DataProvider {
	useProxy := os.Getenv("USE_BINANCE_PROXY") == "true"
	proxyURL := os.Getenv("BINANCE_PROXY_URL")
	if proxyURL == "" {
		// 使用集中的端口配置
		proxyURL = config.GetBinanceProxyURLWithEnv()
	}

	return NewDataProvider(useProxy, proxyURL)
}
