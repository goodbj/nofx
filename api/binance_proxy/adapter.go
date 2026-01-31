package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// DataProvider 数据提供者接口，用于实现低耦合设计
type DataProvider interface {
	GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error)
	GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error)
	GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error)
	GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error)
	GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error)
}

// DirectDataProvider 直接数据提供者（原实现）
type DirectDataProvider struct {
	baseURL string
}

// NewDirectDataProvider 创建直接数据提供者
func NewDirectDataProvider() *DirectDataProvider {
	return &DirectDataProvider{
		baseURL: "https://fapi.binance.com", // 默认使用币安期货API
	}
}

// GetBalance 通过直接API调用获取余额
func (d *DirectDataProvider) GetBalance(apiKey, secretKey, customAPIURL string) (map[string]interface{}, error) {
	baseURL := d.baseURL
	if customAPIURL != "" {
		baseURL = strings.TrimSuffix(customAPIURL, "/")
	}

	endpoint := baseURL + "/fapi/v2/account"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	params := url.Values{}
	params.Add("timestamp", timestamp)

	queryString := params.Encode()
	signature := d.createSignature(queryString, secretKey)
	queryStringWithSig := queryString + "&signature=" + signature

	req, err := http.NewRequest("GET", endpoint+"?"+queryStringWithSig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var accountInfo map[string]interface{}
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 提取余额信息
	balances, ok := accountInfo["assets"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get balances from response")
	}

	balanceMap := make(map[string]interface{})
	for _, assetInterface := range balances {
		if asset, ok := assetInterface.(map[string]interface{}); ok {
			assetName, _ := asset["asset"].(string)
			balanceMap[assetName] = map[string]interface{}{
				"asset":      asset["asset"],
				"walletBal":  asset["walletBalance"],
				"unrealized": asset["unrealizedProfit"],
			}
		}
	}

	return balanceMap, nil
}

// GetPositions 通过直接API调用获取持仓
func (d *DirectDataProvider) GetPositions(apiKey, secretKey, customAPIURL string) ([]map[string]interface{}, error) {
	baseURL := d.baseURL
	if customAPIURL != "" {
		baseURL = strings.TrimSuffix(customAPIURL, "/")
	}

	endpoint := baseURL + "/fapi/v2/positionRisk"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	params := url.Values{}
	params.Add("timestamp", timestamp)

	queryString := params.Encode()
	signature := d.createSignature(queryString, secretKey)
	queryStringWithSig := queryString + "&signature=" + signature

	req, err := http.NewRequest("GET", endpoint+"?"+queryStringWithSig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var positions []map[string]interface{}
	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return positions, nil
}

// GetKlines 通过直接API调用获取K线数据
func (d *DirectDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error) {
	baseURL := d.baseURL
	if customAPIURL != "" {
		baseURL = strings.TrimSuffix(customAPIURL, "/")
	}

	endpoint := fmt.Sprintf("%s/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", baseURL, symbol, interval, limit)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var klineData [][]interface{}
	if err := json.Unmarshal(body, &klineData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var klines []Kline
	for _, k := range klineData {
		if len(k) >= 11 { // K线数据至少需要11个元素
			kline := Kline{
				OpenTime:  int64(k[0].(float64)),
				Open:      k[1].(float64),
				High:      k[2].(float64),
				Low:       k[3].(float64),
				Close:     k[4].(float64),
				Volume:    k[5].(float64),
				CloseTime: int64(k[6].(float64)),
			}
			if k[8] != nil {
				kline.NumberOfTrades = int64(k[8].(float64))
			}
			if k[9] != nil {
				kline.TakerBuyBaseAssetVolume = k[9].(float64)
			}
			if k[10] != nil {
				kline.TakerBuyQuoteAssetVolume = k[10].(float64)
			}
			klines = append(klines, kline)
		}
	}

	return klines, nil
}

// GetAccountInfo 通过直接API调用获取账户信息
func (d *DirectDataProvider) GetAccountInfo(apiKey, secretKey, customAPIURL string) (interface{}, error) {
	baseURL := d.baseURL
	if customAPIURL != "" {
		baseURL = strings.TrimSuffix(customAPIURL, "/")
	}

	endpoint := baseURL + "/fapi/v2/account"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	params := url.Values{}
	params.Add("timestamp", timestamp)

	queryString := params.Encode()
	signature := d.createSignature(queryString, secretKey)
	queryStringWithSig := queryString + "&signature=" + signature

	req, err := http.NewRequest("GET", endpoint+"?"+queryStringWithSig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var accountInfo map[string]interface{}
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return accountInfo, nil
}

// GetTrades 通过直接API调用获取交易历史
func (d *DirectDataProvider) GetTrades(symbol, apiKey, secretKey, customAPIURL string) (interface{}, error) {
	baseURL := d.baseURL
	if customAPIURL != "" {
		baseURL = strings.TrimSuffix(customAPIURL, "/")
	}

	endpoint := baseURL + "/fapi/v1/userTrades"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	params := url.Values{}
	params.Add("symbol", symbol)
	params.Add("timestamp", timestamp)

	queryString := params.Encode()
	signature := d.createSignature(queryString, secretKey)
	queryStringWithSig := queryString + "&signature=" + signature

	req, err := http.NewRequest("GET", endpoint+"?"+queryStringWithSig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	var trades interface{}
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return trades, nil
}

// createSignature 创建签名
func (d *DirectDataProvider) createSignature(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
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
func (p *ProxyDataProvider) GetKlines(symbol, interval string, limit int, apiKey, secretKey, customAPIURL string) ([]Kline, error) {
	url := fmt.Sprintf("%s/api/proxy/klines?symbol=%s&interval=%s&limit=%d", p.ProxyURL, symbol, interval, limit)

	response, err := p.makeRequest(url, map[string]string{}, apiKey, secretKey, customAPIURL)
	if err != nil {
		return nil, err
	}

	// 将响应转换为Kline数组
	klinesData, ok := response["data"]
	if !ok {
		return nil, fmt.Errorf("response does not contain 'data' field")
	}

	klines, ok := klinesData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("klines data is not an array")
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
