package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"nofx/market"
	"nofx/trader"
)

// DataProvider 数据提供者接口，用于实现低耦合设计
type DataProvider interface {
	GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error)
	GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error)
	GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]market.Kline, error)
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
	// 使用现有的Binance期货交易者实现
	binanceTrader := trader.NewFuturesTrader(apiKey, secretKey, "", customAPIURL)
	return binanceTrader.GetBalance()
}

// GetPositions 通过直接API调用获取持仓
func (d *DirectDataProvider) GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error) {
	// 使用现有的Binance期货交易者实现
	binanceTrader := trader.NewFuturesTrader(apiKey, secretKey, "", customAPIURL)
	return binanceTrader.GetPositions()
}

// GetKlines 通过直接API调用获取K线数据
func (d *DirectDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]market.Kline, error) {
	// 由于现有实现中没有直接的K线获取方法，这里返回错误
	// 实际项目中可以从CoinAnk API或其他数据源获取
	return nil, fmt.Errorf("direct K线数据获取功能暂未实现")
}

// GetAccountInfo 通过直接API调用获取账户信息
func (d *DirectDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 使用现有的Binance期货交易者实现
	binanceTrader := trader.NewFuturesTrader(apiKey, secretKey, "", customAPIURL)

	// 获取账户信息
	balance, err := binanceTrader.GetBalance()
	if err != nil {
		return nil, err
	}

	positions, err := binanceTrader.GetPositions()
	if err != nil {
		return nil, err
	}

	accountInfo := map[string]interface{}{
		"balance":   balance,
		"positions": positions,
		"timestamp": time.Now().Unix(),
	}

	return accountInfo, nil
}

// GetTrades 通过直接API调用获取交易历史
func (d *DirectDataProvider) GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error) {
	// 这里需要实现具体的交易历史获取逻辑
	// 由于现有代码中没有直接方法，暂时返回错误
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
	url := fmt.Sprintf("%s/api/proxy/positions", p.ProxyURL)

	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		return nil, err
	}

	// 将通用map转换为[]map[string]interface{}
	positionsData, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("response does not contain 'data' field")
	}

	positions, ok := positionsData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("positions data is not an array")
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		if posMap, ok := pos.(map[string]interface{}); ok {
			result = append(result, posMap)
		}
	}

	return result, nil
}

// GetKlines 通过代理服务获取K线数据
func (p *ProxyDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]market.Kline, error) {
	url := fmt.Sprintf("%s/api/proxy/klines?symbol=%s&interval=%s&limit=%d", p.ProxyURL, symbol, interval, limit)

	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		return nil, err
	}

	// 将响应转换为market.Kline数组
	klinesData, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("response does not contain 'data' field")
	}

	klines, ok := klinesData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("klines data is not an array")
	}

	var result []market.Kline
	for _, k := range klines {
		if kMap, ok := k.(map[string]interface{}); ok {
			kline := market.Kline{
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

// GetAccountInfo 通过代理服务获取账户信息
func (p *ProxyDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/proxy/account", p.ProxyURL)
	return p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
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
