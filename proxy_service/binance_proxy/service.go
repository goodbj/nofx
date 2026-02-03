package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// 从环境变量获取默认目标API URL
var defaultTargetAPIURL = getEnvOrDefault("DEFAULT_TARGET_API_URL", "https://testnet.binancefuture.com")

// 获取环境变量，如果为空则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// 通用代理处理器
func handleGeneric(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理余额查询
func handleBalance(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理账户信息
func handleAccount(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理持仓信息
func handlePositions(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理订单操作
func handleOrder(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理开仓订单
func handleOpenOrders(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理所有订单
func handleAllOrders(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理价格查询
func handleTickerPrice(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理最优买卖档位
func handleBookTicker(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 处理K线数据
func handleKlines(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, defaultTargetAPIURL)
}

// 代理请求函数
func proxyRequest(w http.ResponseWriter, r *http.Request, targetBaseURL string) {
	// 构建目标URL
	targetURL := targetBaseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // 确保原始请求体被关闭

	// 创建新的请求
	req, err := http.NewRequest(r.Method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for header, values := range r.Header {
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}

	// 设置必要的头部
	req.Header.Set("User-Agent", "NOFX-Binance-Proxy/1.0")

	// 移除hop-by-hop headers
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Te")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")

	// 创建具有更好网络配置的HTTP客户端
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   120 * time.Second, // 增加总超时时间
	}

	// 记录请求信息（不包含敏感信息）
	log.Printf("[PROXY] Forwarding %s request to %s", r.Method, redactURL(targetURL))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Proxy request failed: %v", err)
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[ERROR] Error copying response body: %v", err)
	}
}

// 红acted URL以隐藏敏感信息
func redactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// 如果解析失败，返回原URL的前缀部分
		if len(rawURL) > 100 {
			return rawURL[:100] + "..."
		}
		return rawURL
	}

	// 移除可能包含敏感信息的查询参数
	query := parsed.Query()
	redactedParams := []string{"signature", "timestamp", "recvWindow"}

	for _, param := range redactedParams {
		if query.Has(param) {
			query.Set(param, "***REDACTED***")
		}
	}

	parsed.RawQuery = query.Encode()
	return parsed.String()
}
