package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"nofx/binance"
)

// Example usage of the improved Binance modules
func main() {
	fmt.Println("🚀 Binance API改进模块示例")
	fmt.Println("========================")

	// Example 1: Configuration Management
	fmt.Println("\n1. 配置管理示例:")
	configExample()

	// Example 2: Error Handling
	fmt.Println("\n2. 错误处理示例:")
	errorHandlingExample()

	// Example 3: Response Wrapper
	fmt.Println("\n3. 响应包装器示例:")
	responseWrapperExample()

	// Example 4: Rate Limiting
	fmt.Println("\n4. 限频控制示例:")
	rateLimitExample()

	// Example 5: Testnet Support
	fmt.Println("\n5. 测试网支持示例:")
	testnetSupportExample()
}

func configExample() {
	// Create configuration with options
	config := binance.NewConfig(
		"your-api-key-here",
		"your-secret-key-here",
		binance.WithTestnet(true),
		binance.WithTimeout(30*time.Second),
		binance.WithRetries(3, 2*time.Second),
		binance.WithRateLimit(1200, 10),
	)

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Printf("❌ 配置验证失败: %v", err)
		return
	}

	fmt.Printf("✅ 配置创建成功\n")
	fmt.Printf("   API Key: %s...\n", config.APIKey[:8])
	fmt.Printf("   Base URL: %s\n", config.BaseURL)
	fmt.Printf("   Testnet: %t\n", config.Testnet)
	fmt.Printf("   Timeout: %v\n", config.Timeout)
	fmt.Printf("   重试次数: %d\n", config.MaxRetries)
	fmt.Printf("   限频设置: %d 请求/分钟\n", config.RateLimit)
}

func errorHandlingExample() {
	// Create different types of errors
	unauthorizedErr := binance.NewUnauthorizedError("Invalid API key")
	forbiddenErr := binance.NewForbiddenError("Insufficient permissions")
	rateLimitErr := binance.NewTooManyRequestsError("Rate limit exceeded", 60*time.Second)

	fmt.Printf("❌ 认证错误: %v\n", unauthorizedErr)
	fmt.Printf("❌ 权限错误: %v\n", forbiddenErr)
	fmt.Printf("❌ 限频错误: %v (重试延迟: %v)\n", rateLimitErr, rateLimitErr.GetRetryAfter())

	// Check error types
	fmt.Printf("   认证错误可重试: %t\n", unauthorizedErr.IsRetryable())
	fmt.Printf("   限频错误可重试: %t\n", rateLimitErr.IsRetryable())
}

func responseWrapperExample() {
	// Create a successful response
	response := binance.NewResponse().
		WithData(map[string]interface{}{
			"balance": 10000.00,
			"assets":  []string{"BTC", "ETH", "USDT"},
		}).
		WithRateLimits([]*binance.RateLimit{
			{RateLimitType: "REQUEST_WEIGHT", Interval: "MINUTE", Limit: 2400, Count: 1200},
			{RateLimitType: "ORDERS", Interval: "MINUTE", Limit: 1200, Count: 600},
		}).
		WithLatency(150 * time.Millisecond).
		WithRequestID("req-12345").
		WithEndpoint("/fapi/v2/account")

	fmt.Printf("✅ 响应创建成功\n")
	fmt.Printf("   数据: %+v\n", response.GetData())
	fmt.Printf("   延迟: %d ms\n", response.GetMetadata().Latency)
	fmt.Printf("   请求ID: %s\n", response.GetMetadata().RequestID)
	fmt.Printf("   限频信息: %d 个限频规则\n", len(response.GetRateLimits()))

	// Convert to JSON
	jsonData, err := response.ToJSON()
	if err != nil {
		log.Printf("❌ JSON转换失败: %v", err)
		return
	}
	fmt.Printf("   JSON响应: %s\n", string(jsonData))
}

func rateLimitExample() {
	// Create rate limiter
	config := binance.NewConfig("test-key", "test-secret")
	limiter := binance.NewRateLimiter(config)

	fmt.Printf("✅ 限频器创建成功\n")
	fmt.Printf("   最大令牌数: %.1f\n", limiter.(*binance.RateLimiter).(*struct {
		tokens         float64
		maxTokens      float64
		refillRate     float64
		lastRefillTime time.Time
		config         *binance.Config
		stats          *binance.RateLimitStats
		mu             sync.RWMutex
	}).maxTokens)

	// Test rate limiting
	ctx := context.Background()
	start := time.Now()

	// Try to make 5 requests
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(ctx); err != nil {
			log.Printf("❌ 请求 %d 限频失败: %v", i+1, err)
			break
		}
		fmt.Printf("   请求 %d: 通过 (耗时: %v)\n", i+1, time.Since(start))
	}

	// Show statistics
	stats := limiter.GetStats()
	fmt.Printf("   统计信息:\n")
	fmt.Printf("     总请求数: %d\n", stats.TotalRequests)
	fmt.Printf("     成功请求数: %d\n", stats.SuccessfulRequests)
	fmt.Printf("     拒绝请求数: %d\n", stats.RejectedRequests)
}

func testnetSupportExample() {
	// Create network validator
	validator := binance.NewNetworkValidator()

	// Test different URLs
	testUrls := []string{
		"https://fapi.binance.com",
		"https://testnet.binancefuture.com",
		"https://stream.binancefuture.com",
	}

	fmt.Printf("✅ 网络验证示例:\n")
	for _, url := range testUrls {
		isTestnet := validator.(*struct {
			config *binance.TestnetConfig
		}).config.IsTestnetURL(url)
		fmt.Printf("   URL: %s\n", url)
		fmt.Printf("   是否测试网: %t\n", isTestnet)

		// Get recommended rate limits
		rpm, burst := validator.GetRecommendedRateLimits(isTestnet)
		fmt.Printf("   推荐限频: %d 请求/分钟 (突发: %d)\n", rpm, burst)
		fmt.Println()
	}

	// Create network-aware client
	config := binance.NewConfig("test-key", "test-secret", binance.WithBaseURL("https://testnet.binancefuture.com"))
	client := binance.NewNetworkAwareClient(config)

	// Validate configuration
	if err := client.Validate(); err != nil {
		fmt.Printf("❌ 客户端验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 客户端验证通过\n")
	}

	// Get network information
	networkInfo := client.GetNetworkInfo()
	fmt.Printf("   网络信息: %+v\n", networkInfo)
}
