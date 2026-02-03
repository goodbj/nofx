package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// 全局HTTP客户端以提高性能和复用连接
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	},
}

// 从请求头获取目标API URL - 完全依赖nofx提供
func getTargetAPIURL(r *http.Request) string {
	customURL := r.Header.Get("X-Custom-API-URL")
	return customURL
}

// 代理请求函数 - 只转发，不修改
func handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	targetURL := getTargetAPIURL(r)

	if targetURL == "" {
		http.Error(w, "Missing X-Custom-API-URL header", http.StatusBadRequest)
		return
	}

	forwardRequest(w, r, targetURL)
}

// 转发请求函数
func forwardRequest(w http.ResponseWriter, r *http.Request, targetBaseURL string) {
	// 验证目标基础URL是否为空
	if targetBaseURL == "" {
		http.Error(w, "Target API URL cannot be empty", http.StatusBadRequest)
		log.Printf("❌ Empty target URL provided for request from %s", r.RemoteAddr)
		return
	}

	// 获取当前服务器的地址用于循环检测
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8081"
	}
	ownAddress := "http://localhost:" + serverPort

	// 检查是否试图转发到自身，防止循环
	if strings.HasPrefix(targetBaseURL, ownAddress) {
		http.Error(w, "Prevented infinite loop: trying to forward to self", http.StatusBadRequest)
		log.Printf("❌ Blocked request that would cause infinite loop: %s", targetBaseURL)
		return
	}

	// 确保目标基础URL以http://或https://开头
	finalBaseURL := targetBaseURL
	if !strings.HasPrefix(finalBaseURL, "http://") && !strings.HasPrefix(finalBaseURL, "https://") {
		finalBaseURL = "https://" + finalBaseURL
	}

	// 确保基础URL末尾没有斜杠
	finalBaseURL = strings.TrimSuffix(finalBaseURL, "/")

	// 构建目标URL
	targetURL := finalBaseURL + r.URL.Path
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
	req.Header.Set("User-Agent", "NOFX-Transparent-Proxy/1.0")

	// 移除hop-by-hop headers
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Te")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")

	// 使用全局HTTP客户端以提高性能

	resp, err := httpClient.Do(req)
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

// Redact URL to hide sensitive information
func redactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If parsing fails, return the prefix of the URL
		if len(rawURL) > 100 {
			return rawURL[:100] + "..."
		}
		return rawURL
	}

	// Remove query parameters that may contain sensitive information
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

func main() {
	// Create router
	r := mux.NewRouter()

	// Add health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Add catch-all handler for proxy requests
	r.PathPrefix("/").HandlerFunc(handleProxyRequest)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port
	}

	// Create HTTP server with timeout settings
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Independent Transparent Proxy Server starting on port %s", port)
	log.Printf("📊 Health check available at: http://localhost:%s/health", port)

	// Start server
	log.Fatal(server.ListenAndServe())
}
