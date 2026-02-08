package main

/*
注意，重要声明

这是代理接口服务必须遵守的原则）
AI改写增减代码时必须按照以下原则进行
代理服务现在的作用是：
1 纯透传原则
2 验证与远程服务的连接性
3 传输认证信息
4 将请求转发给远程服务
5 将响应返回给nofx原生方法进行处理
纯透传架构设计，其中：
1代理服务只做网络层的透明转发
2不处理任何业务逻辑
3不依赖环境变量进行控制
4所有功能都通过原始交易者实现
*/
import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// 日志频率控制结构
type LogCounter struct {
	count     int
	lastReset time.Time
	mutex     sync.Mutex
}

// 全局日志计数器
var (
	logCounters = make(map[string]*LogCounter)
	logMutex    = sync.RWMutex{}
)

// 限制相同URL的日志输出频率
func shouldLogRequest(targetURL string) bool {
	now := time.Now()
	threshold := 30 * time.Second // 30秒内最多记录指定次数
	maxLogsPerURL := 3            // 每个URL每30秒最多3条日志

	logMutex.Lock()
	defer logMutex.Unlock()

	counter, exists := logCounters[targetURL]
	if !exists {
		counter = &LogCounter{
			count:     1,
			lastReset: now,
		}
		logCounters[targetURL] = counter
		return true
	}

	// 如果距离上次重置超过阈值，重置计数器
	if now.Sub(counter.lastReset) > threshold {
		counter.count = 1
		counter.lastReset = now
		return true
	}

	// 如果计数未达到上限，增加计数并记录
	if counter.count < maxLogsPerURL {
		counter.count++
		return true
	}

	// 达到上限，不记录日志
	return false
}

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
		// 使用智能日志控制
		if shouldLogRequest(targetBaseURL) {
			log.Printf("❌ Blocked request that would cause infinite loop: %s", targetBaseURL)
		}
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

// Hot reload test comment
