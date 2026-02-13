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
	"fmt"
	"io"
	"log"
	"net"
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

// 异步日志工作器
func startLogWorker() {
	go func() {
		for logMsg := range logQueue {
			log.Printf("[PROXY] %s", logMsg)
		}
	}()
}

// 异步日志记录函数
func logAsync(format string, args ...interface{}) {
	logWorkerStarted.Do(startLogWorker)

	logMsg := fmt.Sprintf(format, args...)

	select {
	case logQueue <- logMsg:
		// 成功入队
	default:
		// 队列满时丢弃日志，避免阻塞
		if len(logQueue) == cap(logQueue) {
			// 只记录队列满的警告，避免过多日志
			select {
			case logQueue <- "[WARNING] Log queue is full, dropping messages":
			default:
			}
		}
	}
}

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
	Timeout: 120 * time.Second, // 增加超时时间到120秒，适应网络波动
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 增加拨号超时时间
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          200,               // 增加最大空闲连接数
		MaxIdleConnsPerHost:   20,                // 增加每主机最大空闲连接数
		MaxConnsPerHost:       50,                // 限制每主机最大连接数
		IdleConnTimeout:       120 * time.Second, // 延长空闲连接超时
		TLSHandshakeTimeout:   30 * time.Second,  // 增加TLS握手超时时间
		ResponseHeaderTimeout: 60 * time.Second,  // 增加响应头超时时间
		ExpectContinueTimeout: 5 * time.Second,   // 增加Expect-continue超时时间
		DisableKeepAlives:     false,
		DisableCompression:    false, // 启用压缩
	},
}

// 并发控制信号量 - 限制最大并发请求数
var semaphore = make(chan struct{}, 100) // 最多100个并发请求

// 异步日志队列
var logQueue = make(chan string, 1000)
var logWorkerStarted sync.Once

// 内存池 - 复用缓冲区
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 32KB缓冲区
	},
}

// 从请求头获取完整的目标URL
func getTargetURL(r *http.Request) string {
	// 尝试多种可能的头部名称，包括与CustomTransport兼容的头部
	headers := []string{"X-Target-URL", "X-Target-Url", "x-target-url", "X-TARGET-URL", "X-Custom-API-URL", "X-Custom-Api-Url", "x-custom-api-url"}

	for _, header := range headers {
		if url := r.Header.Get(header); url != "" {
			log.Printf("🎯 Found target URL in header %s: %s", header, url)
			return url
		}
	}

	// 如果没找到，记录所有头部信息用于调试
	log.Printf("🔍 Request headers:")
	for name, values := range r.Header {
		log.Printf("  %s: %v", name, values)
	}

	return ""
}

// 代理请求函数 - 只转发，不修改
func handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	// 使用信号量控制并发
	semaphore <- struct{}{}        // 获取令牌
	defer func() { <-semaphore }() // 释放令牌

	logAsync("🔄 Proxy request received: %s %s", r.Method, r.URL.Path)

	targetURL := getTargetURL(r)
	logAsync("🎯 Target URL from header: %s", targetURL)

	if targetURL == "" {
		http.Error(w, "Missing X-Target-URL header", http.StatusBadRequest)
		return
	}

	forwardRequest(w, r, targetURL)
}

// 转发请求函数
func forwardRequest(w http.ResponseWriter, r *http.Request, targetBaseURL string) {
	// 验证目标基础URL是否为空
	if targetBaseURL == "" {
		http.Error(w, "Target URL cannot be empty", http.StatusBadRequest)
		logAsync("❌ Empty target URL provided for request from %s", r.RemoteAddr)
		return
	}

	// 获取当前服务器的地址用于循环检测
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	// 增强循环检测逻辑 - 包括多种可能的localhost表示形式
	ownAddresses := []string{
		"http://localhost:" + serverPort,
		"http://127.0.0.1:" + serverPort,
		"http://0.0.0.0:" + serverPort,
		"http://[::1]:" + serverPort,
		"http://localhost:8081",
		"http://127.0.0.1:8081",
		"http://0.0.0.0:8081",
		"http://[::1]:8081",
		"http://192.168.1.1:" + serverPort, // 常见本地IP
		"http://10.0.0.1:" + serverPort,    // 常见私有IP
	}

	// 检查是否试图转发到自身，防止循环
	for _, addr := range ownAddresses {
		if strings.HasPrefix(targetBaseURL, addr) {
			http.Error(w, "Prevented infinite loop: trying to forward to self", http.StatusBadRequest)
			// 使用智能日志控制
			if shouldLogRequest(targetBaseURL) {
				logAsync("❌ Blocked request that would cause infinite loop: %s -> %s", r.RemoteAddr, targetBaseURL)
			}
			return
		}
	}

	// 确保目标基础URL以http://或https://开头
	finalBaseURL := targetBaseURL
	if !strings.HasPrefix(finalBaseURL, "http://") && !strings.HasPrefix(finalBaseURL, "https://") {
		finalBaseURL = "https://" + finalBaseURL
	}

	// 确保基础URL末尾没有斜杠
	finalBaseURL = strings.TrimSuffix(finalBaseURL, "/")

	// 构建目标URL
	// 如果目标URL已经是完整URL，则将其作为基础URL，并附加请求路径和查询参数
	logAsync("🔍 Checking if targetBaseURL has HTTP(S) prefix: %s", targetBaseURL)
	var targetURL string
	if strings.HasPrefix(targetBaseURL, "http://") || strings.HasPrefix(targetBaseURL, "https://") {
		logAsync("🔗 Processing as HTTP(S) URL, base: %s", targetBaseURL)
		logAsync("🔗 Original request path: %s, query: %s", r.URL.Path, r.URL.RawQuery)
		// 将X-Target-URL作为基础URL，然后附加原始请求的路径和查询参数
		baseURL := strings.TrimSuffix(targetBaseURL, "/")
		requestPath := r.URL.Path
		logAsync("🔗 Processed requestPath: %s", requestPath)
		if requestPath == "" || requestPath == "/" {
			requestPath = ""
			logAsync("🔗 Resetting requestPath to empty")
		}
		targetURL = baseURL + requestPath
		logAsync("🔗 URL after adding path: %s", targetURL)
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
			logAsync("🔗 Final URL after adding query: %s", targetURL)
		}
	} else {
		logAsync("🔗 Processing as non-HTTP URL, finalBaseURL: %s, request path: %s", finalBaseURL, r.URL.Path)
		// 否则将基础URL与请求路径组合
		targetURL = finalBaseURL + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}
	}

	logAsync("🎯 Final target URL: %s", targetURL)

	// 使用内存池读取请求体
	buffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(buffer)

	body, err := io.ReadAll(io.LimitReader(r.Body, int64(len(buffer))))
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
	// 移除我们使用的特殊头部，避免循环传递
	req.Header.Del("X-Target-URL")

	// 使用全局HTTP客户端以提高性能
	resp, err := httpClient.Do(req)
	if err != nil {
		logAsync("[ERROR] Proxy request failed: %v", err)
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 记录响应状态码用于调试
	if resp.StatusCode >= 400 {
		logAsync("[DEBUG] Response from target server: status %d for URL %s", resp.StatusCode, targetURL)
	}

	// 复制响应头
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 使用内存池读取响应体
	responseBuffer := bufferPool.Get().([]byte)
	defer bufferPool.Put(responseBuffer)

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, int64(len(responseBuffer))))
	if err != nil {
		logAsync("[ERROR] Error reading response body: %v", err)
		http.Error(w, "Error reading response from target server", http.StatusInternalServerError)
		return
	}

	// 记录响应体大小和内容（如果是错误响应）
	if resp.StatusCode >= 400 {
		bodyLen := len(responseBody)
		truncateLen := bodyLen
		if truncateLen > 200 {
			truncateLen = 200
		}
		logAsync("[DEBUG] Response body size: %d bytes for error status %d", bodyLen, resp.StatusCode)
		if bodyLen > 0 {
			logAsync("[DEBUG] Response body (first 200 chars): %s", string(responseBody)[:truncateLen])
		} else {
			logAsync("[DEBUG] Response body is empty for error status %d", resp.StatusCode)
		}
	}

	// 检查响应内容类型
	contentType := resp.Header.Get("Content-Type")
	logAsync("[DEBUG] Response Content-Type: %s", contentType)

	// 对于Binance API，即使错误响应也应返回JSON格式，但有时可能返回HTML错误页面或纯文本
	// 如果是错误状态码且内容不是JSON格式，尝试返回更友好的错误信息
	if resp.StatusCode >= 400 && len(responseBody) > 0 {
		responseStr := string(responseBody)
		isJSON := strings.HasPrefix(strings.TrimSpace(responseStr), "{") ||
			strings.HasPrefix(strings.TrimSpace(responseStr), "[")

		if !isJSON {
			logAsync("[WARNING] Non-JSON response received from %s", targetURL)
			// 尝试返回一个JSON格式的错误响应，以便客户端正确处理
			errorResponse := fmt.Sprintf(`{"error": "Non-JSON response from upstream server", "original_status": %d, "content_type": %q}`, resp.StatusCode, contentType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			_, writeErr := w.Write([]byte(errorResponse))
			if writeErr != nil {
				logAsync("[ERROR] Error writing error response: %v", writeErr)
			}
		} else {
			// JSON格式的错误响应，直接返回
			_, err = w.Write(responseBody)
			if err != nil {
				logAsync("[ERROR] Error writing response body: %v", err)
			}
		}
	} else {
		// 正常响应或错误响应但有内容
		_, err = w.Write(responseBody)
		if err != nil {
			logAsync("[ERROR] Error writing response body: %v", err)
		}
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

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	log.Printf("⚡ Performance optimizations enabled:")
	log.Printf("   - Connection pooling: MaxIdleConns=200, MaxIdleConnsPerHost=20")
	log.Printf("   - Concurrency control: Max 100 concurrent requests")
	log.Printf("   - Async logging: Non-blocking log processing")
	log.Printf("   - Memory pooling: 32KB buffer reuse")
	log.Printf("   - Response compression: Enabled")

	// Start server
	log.Fatal(server.ListenAndServe())
}

// Hot reload test comment
