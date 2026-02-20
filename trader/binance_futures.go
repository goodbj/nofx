package trader

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"nofx/hook"
	"nofx/logger"
	"nofx/store"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

// getBrOrderID generates unique order ID (for futures contracts)
// Format: x-{BR_ID}{TIMESTAMP}{RANDOM}
// Futures limit is 32 characters, use this limit consistently
// Uses nanosecond timestamp + random number to ensure global uniqueness (collision probability < 10^-20)
func getBrOrderID() string {
	brID := "KzrpZaP9" // Futures br ID

	// Calculate available space: 32 - len("x-KzrpZaP9") = 32 - 11 = 21 characters
	// Allocation: 13-digit timestamp + 8-digit random = 21 characters (perfect utilization)
	timestamp := time.Now().UnixNano() % 10000000000000 // 13-digit nanosecond timestamp

	// Generate 4-byte random number (8 hex digits)
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	// Format: x-KzrpZaP9{13-digit timestamp}{8-digit random}
	// Example: x-KzrpZaP91234567890123abcdef12 (exactly 31 characters)
	orderID := fmt.Sprintf("x-%s%d%s", brID, timestamp, randomHex)

	// Ensure not exceeding 32-character limit (theoretically exactly 31 characters)
	if len(orderID) > 32 {
		orderID = orderID[:32]
	}

	return orderID
}

// FuturesTrader Binance futures trader
type FuturesTrader struct {
	client *futures.Client

	// Balance cache
	cachedBalance     map[string]interface{}
	balanceCacheTime  time.Time
	balanceCacheMutex sync.RWMutex

	// Position cache
	cachedPositions     []map[string]interface{}
	positionsCacheTime  time.Time
	positionsCacheMutex sync.RWMutex

	// Cache validity period (30 seconds for better performance)
	cacheDuration time.Duration

	// Precision manager for handling quantity/price formatting
	precisionManager PrecisionManagerInterface

	// Trader identification for sync operations
	traderID     string
	exchangeID   string
	exchangeType string
	store        *store.Store
}

// validateFormattedQuantity 验证格式化后的数量是否符合step size要求
func (t *FuturesTrader) validateFormattedQuantity(formatted string, stepSize float64) error {
	// 重新解析验证
	parsedQty, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		return fmt.Errorf("failed to parse formatted quantity: %w", err)
	}

	// 检查是否符合stepSize要求
	if stepSize > 0 {
		remainder := math.Mod(parsedQty, stepSize)
		// 允许很小的浮点数误差
		if math.Abs(remainder) > 1e-10 && math.Abs(remainder-stepSize) > 1e-10 {
			return fmt.Errorf("quantity %s does not align with stepSize %f (remainder: %f)",
				formatted, stepSize, remainder)
		}
	}

	return nil
}

// getDefaultStepSize returns reasonable default step size based on symbol type
func getDefaultStepSize(symbol string) float64 {
	// 根据交易对类型返回合理的默认step size
	baseAsset := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(symbol, "USDT"), "BUSD"), "USDC")

	// 高价值资产（BTC, ETH等）
	highValueAssets := map[string]bool{
		"BTC": true, "ETH": true, "BNB": true,
	}

	// 中等价值资产
	midValueAssets := map[string]bool{
		"SOL": true, "ADA": true, "XRP": true, "DOT": true, "LINK": true,
		"MATIC": true, "AVAX": true, "ATOM": true, "NEAR": true, "APT": true,
		"ARB": true, "OP": true,
	}

	// 低价值大量资产
	lowValueAssets := map[string]bool{
		"SHIB": true, "DOGE": true, "PEPE": true, "FLOKI": true, "BONK": true,
	}

	if highValueAssets[baseAsset] {
		return 0.001 // 高价值资产使用较低精度
	} else if midValueAssets[baseAsset] {
		return 0.1 // 中等价值资产使用适中精度
	} else if lowValueAssets[baseAsset] {
		return 1.0 // 低价值资产使用较高精度
	} else {
		return 0.1 // 默认使用中等精度
	}
}

// NewFuturesTrader creates futures trader
func NewFuturesTrader(apiKey, secretKey, userId, customEndpoint string) *FuturesTrader {
	var client *futures.Client
	if customEndpoint != "" {
		// Use custom endpoint if provided
		client = futures.NewClient(apiKey, secretKey)
		// For testnet, the correct endpoint should be https://testnet.binancefuture.com
		// The library will append the API path (/fapi/v2/account, etc.) automatically
		client.BaseURL = customEndpoint
	} else {
		// Use default endpoint
		client = futures.NewClient(apiKey, secretKey)
	}

	// Enhance HTTP client with robust network configuration to handle network instability, especially for testnet and restricted networks
	// Use similar configuration as proxy service for better connectivity
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

	// 如果customEndpoint是代理URL（localhost或127.0.0.1），则添加自定义头
	// 但如果customEndpoint是代理URL，我们需要告诉代理服务真正的目标URL是什么
	// 注意：这里需要特殊处理，因为在代理场景中，customEndpoint是目标URL，而不是代理URL
	// 实际的代理URL是在ProxyTraderWrapper中处理的
	if customEndpoint != "" && (strings.Contains(customEndpoint, "localhost") || strings.Contains(customEndpoint, "127.0.0.1")) {
		// 创建一个自定义RoundTripper来添加自定义头信息
		customTransport := &CustomTransport{
			Transport:      transport,
			TargetEndpoint: customEndpoint, // 实际的目标端点
		}

		if client.HTTPClient == nil {
			client.HTTPClient = &http.Client{
				Transport: customTransport,
				Timeout:   120 * time.Second, // Increased timeout
			}
		} else {
			client.HTTPClient.Transport = customTransport
			client.HTTPClient.Timeout = 120 * time.Second // Increased timeout
		}
	} else {
		// 非代理情况，使用普通的增强传输配置
		if client.HTTPClient == nil {
			client.HTTPClient = &http.Client{
				Transport: transport,
				Timeout:   120 * time.Second, // Increased timeout
			}
		} else {
			client.HTTPClient.Transport = transport
			client.HTTPClient.Timeout = 120 * time.Second // Increased timeout
		}
	}

	hookRes := hook.HookExec[hook.NewBinanceTraderResult](hook.NEW_BINANCE_TRADER, userId, client)
	if hookRes != nil && hookRes.GetResult() != nil {
		client = hookRes.GetResult()
	}

	// Sync time to avoid "Timestamp ahead" error
	syncBinanceServerTime(client)

	// Create precision manager with file-based cache
	precisionManager := NewFileBasedPrecisionManager(client, filepath.Join("trader", "jingdu.json"))

	trader := &FuturesTrader{
		client:           client,
		cacheDuration:    30 * time.Second, // 30-second cache for better performance
		precisionManager: precisionManager,
	}

	// Set dual-side position mode (Hedge Mode)
	// This is required because the code uses PositionSide (LONG/SHORT)
	if err := trader.setDualSidePosition(); err != nil {
		logger.Infof("⚠️ Failed to set dual-side position mode: %v (ignore this warning if already in dual-side mode)", err)
	}

	return trader
}

// NewFuturesTraderViaProxy 创建通过代理的期货交易者实例
// proxyURL: 代理服务的URL
// targetEndpoint: 真正的目标端点
func NewFuturesTraderViaProxy(apiKey, secretKey, userId, proxyURL, targetEndpoint string) *FuturesTrader {
	var client *futures.Client
	// 连接到代理URL
	client = futures.NewClient(apiKey, secretKey)
	client.BaseURL = proxyURL // 连接到代理

	// Enhance HTTP client with robust network configuration to handle network instability, especially for testnet and restricted networks
	// Use similar configuration as proxy service for better connectivity
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

	// 创建CustomTransport，将真正的目标端点传递给代理
	logger.Debugf("🔍 NewFuturesTraderViaProxy - proxyURL: %s, targetEndpoint: %s", proxyURL, targetEndpoint)
	customTransport := &CustomTransport{
		Transport:      transport,
		TargetEndpoint: targetEndpoint, // 真正的目标URL
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{
			Transport: customTransport,
			Timeout:   120 * time.Second, // Increased timeout
		}
	} else {
		client.HTTPClient.Transport = customTransport
		client.HTTPClient.Timeout = 120 * time.Second // Increased timeout
	}

	hookRes := hook.HookExec[hook.NewBinanceTraderResult](hook.NEW_BINANCE_TRADER, userId, client)
	if hookRes != nil && hookRes.GetResult() != nil {
		client = hookRes.GetResult()
	}

	// Sync time to avoid "Timestamp ahead" error
	syncBinanceServerTime(client)

	// Create precision manager with file-based cache
	precisionManager := NewFileBasedPrecisionManager(client, filepath.Join("trader", "jingdu.json"))

	trader := &FuturesTrader{
		client:           client,
		cacheDuration:    30 * time.Second, // 30-second cache for better performance
		precisionManager: precisionManager,
	}

	// Set dual-side position mode (Hedge Mode)
	// This is required because the code uses PositionSide (LONG/SHORT)
	if err := trader.setDualSidePosition(); err != nil {
		logger.Infof("⚠️ Failed to set dual-side position mode: %v (ignore this warning if already in dual-side mode)", err)
	}

	return trader
}

// setDualSidePosition sets dual-side position mode (called during initialization)
func (t *FuturesTrader) setDualSidePosition() error {
	// Try to set dual-side position mode
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := t.client.NewChangePositionModeService().
		DualSide(true). // true = dual-side position (Hedge Mode)
		Do(ctx)

	if err != nil {
		errMsg := err.Error()

		// If error message contains "No need to change", it means already in dual-side position mode
		if strings.Contains(errMsg, "No need to change position side") {
			logger.Infof("  ✓ Account is already in dual-side position mode (Hedge Mode)")
			return nil
		}

		// If error is code -4067 (Position side cannot be changed if there exists open orders),
		// this is a common scenario and shouldn't interrupt initialization
		if strings.Contains(errMsg, "code=-4067") || strings.Contains(errMsg, "Position side cannot be changed if there exists open orders") {
			logger.Infof("  ⚠️ Account has open orders, cannot change position mode now (this is normal). Will retry after orders are closed.")
			logger.Infof("  ℹ️  Dual-side position mode is typically set once during initial account setup.")
			return nil
		}

		// If error is related to API key issues, log specifically for debugging
		if strings.Contains(errMsg, "API-key") {
			logger.Errorf("  ❌ API key error when setting dual-side position mode: %v", err)
			return err // Return this error as it indicates a fundamental authentication problem
		}

		// For other errors, log but don't necessarily interrupt initialization
		logger.Warnf("  ⚠️ Failed to set dual-side position mode: %v", err)
		return nil // Don't interrupt initialization for non-critical errors
	}

	logger.Infof("  ✓ Account has been switched to dual-side position mode (Hedge Mode)")
	logger.Infof("  ℹ️  Dual-side position mode allows holding both long and short positions simultaneously")
	return nil
}

// syncBinanceServerTime syncs Binance server time to ensure request timestamps are valid
func syncBinanceServerTime(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	serverTime, err := client.NewServerTimeService().Do(ctx)
	if err != nil {
		logger.Infof("⚠️ Failed to sync Binance server time: %v", err)
		return
	}

	now := time.Now().UnixMilli()
	offset := now - serverTime
	client.TimeOffset = offset
	logger.Infof("⏱ Binance server time synced, offset %dms", offset)
}

// GetBalance gets account balance (with cache and smart retry logic)
func (t *FuturesTrader) GetBalance() (map[string]interface{}, error) {
	// First check if cache is valid
	t.balanceCacheMutex.RLock()
	if t.cachedBalance != nil && time.Since(t.balanceCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.balanceCacheTime)
		t.balanceCacheMutex.RUnlock()
		logger.Infof("✓ Using cached account balance (cache age: %.1f seconds ago)", cacheAge.Seconds())
		return t.cachedBalance, nil
	}
	t.balanceCacheMutex.RUnlock()

	// Cache expired or doesn't exist, call API with smart retry
	logger.Infof("🔄 Cache expired, calling Binance API to get account balance...")

	var account *futures.Account
	var err error

	// Try once with shorter timeout for fast response
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	account, err = t.client.NewGetAccountService().Do(ctx)
	cancel()

	if err == nil {
		// Success on first try, no need for retry
		result := make(map[string]interface{})
		result["totalWalletBalance"], _ = strconv.ParseFloat(account.TotalWalletBalance, 64)
		result["availableBalance"], _ = strconv.ParseFloat(account.AvailableBalance, 64)
		result["totalUnrealizedProfit"], _ = strconv.ParseFloat(account.TotalUnrealizedProfit, 64)

		logger.Infof("✓ Binance API returned: total balance=%s, available=%s, unrealized PnL=%s",
			account.TotalWalletBalance,
			account.AvailableBalance,
			account.TotalUnrealizedProfit)

		// Update cache
		t.balanceCacheMutex.Lock()
		t.cachedBalance = result
		t.balanceCacheTime = time.Now()
		t.balanceCacheMutex.Unlock()

		return result, nil
	}

	// If first attempt failed with network error, try with retry logic
	isRetryable := strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "i/o timeout")

	if isRetryable {
		// Perform retry logic only for network-related errors
		maxRetries := 2 // Reduced retry count to improve performance

		for attempt := 1; attempt <= maxRetries; attempt++ {
			logger.Infof("❌ Binance API call failed (attempt %d/%d): %v", attempt, maxRetries, err)

			// Wait before retry (shorter backoff)
			waitTime := time.Duration(attempt) * 500 * time.Millisecond
			logger.Infof("⏱ Retrying in %v...", waitTime)
			time.Sleep(waitTime)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			account, err = t.client.NewGetAccountService().Do(ctx)
			cancel()

			if err == nil {
				break // Success, exit retry loop
			}

			// Check again if error is still retryable
			isRetryable = strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "connection reset") ||
				strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "i/o timeout")

			if !isRetryable {
				// Non-retryable error, break retry loop
				break
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	result := make(map[string]interface{})
	result["totalWalletBalance"], _ = strconv.ParseFloat(account.TotalWalletBalance, 64)
	result["availableBalance"], _ = strconv.ParseFloat(account.AvailableBalance, 64)
	result["totalUnrealizedProfit"], _ = strconv.ParseFloat(account.TotalUnrealizedProfit, 64)

	logger.Infof("✓ Binance API returned: total balance=%s, available=%s, unrealized PnL=%s",
		account.TotalWalletBalance,
		account.AvailableBalance,
		account.TotalUnrealizedProfit)

	// Update cache
	t.balanceCacheMutex.Lock()
	t.cachedBalance = result
	t.balanceCacheTime = time.Now()
	t.balanceCacheMutex.Unlock()

	return result, nil
}

// GetPositions gets all positions (with cache)
func (t *FuturesTrader) GetPositions() ([]map[string]interface{}, error) {
	// 强制刷新缓存用于调试
	forceRefresh := true

	// First check if cache is valid
	t.positionsCacheMutex.RLock()
	if !forceRefresh && t.cachedPositions != nil && time.Since(t.positionsCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.positionsCacheTime)
		t.positionsCacheMutex.RUnlock()
		logger.Infof("✓ Using cached position information (cache age: %.1f seconds ago)", cacheAge.Seconds())
		return t.cachedPositions, nil
	}
	t.positionsCacheMutex.RUnlock()

	// Cache expired or doesn't exist, call API
	logger.Infof("🔄 Cache expired, calling Binance API to get position information...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	positions, err := t.client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		// 检查是否为代理转发错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "Error forwarding request") ||
			strings.Contains(errMsg, "forwarding request") ||
			strings.Contains(errMsg, "proxy error") {
			logger.Warnf("⚠️ 代理转发失败，尝试使用缓存数据: %v", err)

			// 如果有缓存数据，即使过期也返回缓存数据
			t.positionsCacheMutex.RLock()
			if t.cachedPositions != nil {
				cacheAge := time.Since(t.positionsCacheTime)
				t.positionsCacheMutex.RUnlock()
				logger.Warnf("⚠️ 返回过期的缓存数据 (cache age: %.1f seconds ago)", cacheAge.Seconds())
				return t.cachedPositions, nil
			}
			t.positionsCacheMutex.RUnlock()

			// 如果没有缓存数据，返回空数组而不是错误
			logger.Warnf("⚠️ 无缓存数据可用，返回空持仓列表")
			return []map[string]interface{}{}, nil
		}

		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var result []map[string]interface{}
	logger.Debugf("🔍 Binance持仓响应解析开始，原始持仓数: %d", len(positions))
	for i, pos := range positions {
		// 安全解析持仓数量
		posAmt, err := strconv.ParseFloat(pos.PositionAmt, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析持仓数量 %s: %v，跳过该持仓", pos.PositionAmt, err)
			continue
		}

		if posAmt == 0 {
			continue // Skip positions with zero amount
		}

		logger.Debugf("  原始持仓[%d]: Symbol=%s, PositionAmt=%s, EntryPrice=%s, Side=%s",
			i, pos.Symbol, pos.PositionAmt, pos.EntryPrice, pos.PositionSide)

		posMap := make(map[string]interface{})
		posMap["symbol"] = pos.Symbol
		// 安全解析各项数值
		positionAmt, err := strconv.ParseFloat(pos.PositionAmt, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析持仓数量 %s: %v，跳过该持仓", pos.PositionAmt, err)
			continue
		}
		posMap["positionAmt"] = positionAmt

		entryPrice, err := strconv.ParseFloat(pos.EntryPrice, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析入场价格 %s: %v，使用0", pos.EntryPrice, err)
			entryPrice = 0
		}
		posMap["entryPrice"] = entryPrice

		markPrice, err := strconv.ParseFloat(pos.MarkPrice, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析标记价格 %s: %v，使用0", pos.MarkPrice, err)
			markPrice = 0
		}
		posMap["markPrice"] = markPrice

		unRealizedProfit, err := strconv.ParseFloat(pos.UnRealizedProfit, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析未实现盈亏 %s: %v，使用0", pos.UnRealizedProfit, err)
			unRealizedProfit = 0
		}
		posMap["unRealizedProfit"] = unRealizedProfit

		leverage, err := strconv.ParseFloat(pos.Leverage, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析杠杆 %s: %v，使用1", pos.Leverage, err)
			leverage = 1
		}
		posMap["leverage"] = leverage

		liquidationPrice, err := strconv.ParseFloat(pos.LiquidationPrice, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析爆仓价格 %s: %v，使用0", pos.LiquidationPrice, err)
			liquidationPrice = 0
		}
		posMap["liquidationPrice"] = liquidationPrice
		posMap["positionSide"] = pos.PositionSide // 保留原始的positionSide信息
		// Note: Binance SDK doesn't expose updateTime field, will fallback to local tracking

		// Determine direction
		if posAmt > 0 {
			posMap["side"] = "long"
		} else {
			posMap["side"] = "short"
		}

		logger.Debugf("  ✅ 处理后持仓[%d]: symbol=%s, positionAmt=%.6f, side=%s, positionSide=%s",
			i, posMap["symbol"], posMap["positionAmt"], posMap["side"], posMap["positionSide"])
		result = append(result, posMap)
	}
	logger.Debugf("🔍 Binance持仓解析完成，有效持仓数: %d", len(result))

	// Update cache
	t.positionsCacheMutex.Lock()
	t.cachedPositions = result
	t.positionsCacheTime = time.Now()
	t.positionsCacheMutex.Unlock()

	return result, nil
}

// SetMarginMode sets margin mode
func (t *FuturesTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	var marginType futures.MarginType
	if isCrossMargin {
		marginType = futures.MarginTypeCrossed
	} else {
		marginType = futures.MarginTypeIsolated
	}

	// Try to set margin mode
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := t.client.NewChangeMarginTypeService().
		Symbol(symbol).
		MarginType(marginType).
		Do(ctx)

	marginModeStr := "Cross Margin"
	if !isCrossMargin {
		marginModeStr = "Isolated Margin"
	}

	if err != nil {
		// If error message contains "No need to change", margin mode is already set to target value
		if contains(err.Error(), "No need to change margin type") {
			logger.Infof("  ✓ %s margin mode is already %s", symbol, marginModeStr)
			return nil
		}
		// If there is an open position, margin mode cannot be changed, but this doesn't affect trading
		if contains(err.Error(), "Margin type cannot be changed if there exists position") {
			logger.Infof("  ⚠️ %s has open positions, cannot change margin mode, continuing with current mode", symbol)
			return nil
		}
		// Detect Multi-Assets mode (error code -4168)
		if contains(err.Error(), "Multi-Assets mode") || contains(err.Error(), "-4168") || contains(err.Error(), "4168") {
			logger.Infof("  ⚠️ %s detected Multi-Assets mode, forcing Cross Margin mode", symbol)
			logger.Infof("  💡 Tip: To use Isolated Margin mode, please disable Multi-Assets mode in Binance")
			return nil
		}
		// Detect Unified Account API (Portfolio Margin)
		if contains(err.Error(), "unified") || contains(err.Error(), "portfolio") || contains(err.Error(), "Portfolio") {
			logger.Infof("  ❌ %s detected Unified Account API, unable to trade futures", symbol)
			return fmt.Errorf("please use 'Spot & Futures Trading' API permission, do not use 'Unified Account API'")
		}
		logger.Infof("  ⚠️ Failed to set margin mode: %v", err)
		// Don't return error, let trading continue
		return nil
	}

	logger.Infof("  ✓ %s margin mode set to %s", symbol, marginModeStr)
	return nil
}

// SetLeverage sets leverage (with smart detection and cooldown period)
func (t *FuturesTrader) SetLeverage(symbol string, leverage int) error {
	// First try to get current leverage (from position information)
	currentLeverage := 0
	positions, err := t.GetPositions()
	if err == nil {
		for _, pos := range positions {
			if pos["symbol"] == symbol {
				// Handle both float64 and *json.Number types for leverage
				levInterface := pos["leverage"]
				var lev float64
				if levVal, ok := levInterface.(float64); ok {
					lev = levVal
				} else if levNum, ok := levInterface.(*json.Number); ok {
					var err error
					lev, err = levNum.Float64()
					if err != nil {
						logger.Warnf("Failed to convert leverage to float for %s, using 0: %v", symbol, err)
						continue // Skip this position and try the next one
					}
				} else {
					logger.Warnf("Failed to get leverage for %s, using 0", symbol)
					continue // Skip this position and try the next one
				}
				currentLeverage = int(lev)
				break
			}
		}
	}

	// If current leverage is already the target leverage, skip
	if currentLeverage == leverage && currentLeverage > 0 {
		logger.Infof("  ✓ %s leverage is already %dx, no need to change", symbol, leverage)
		return nil
	}

	// Change leverage with optimized timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err = t.client.NewChangeLeverageService().
		Symbol(symbol).
		Leverage(leverage).
		Do(ctx)

	if err != nil {
		errMsg := err.Error()
		logger.Warnf("Failed to set leverage for %s to %dx: %v", symbol, leverage, err)

		// If error message contains "No need to change", leverage is already the target value
		if contains(errMsg, "No need to change") {
			logger.Infof("  ✓ %s leverage is already %dx", symbol, leverage)
			return nil
		}

		// If it's error -1000 (unknown error), check if leverage is already set correctly
		if contains(errMsg, "code=-1000") {
			logger.Infof("  ⚠ Got error -1000 when setting leverage for %s, checking current leverage...", symbol)
			// Try to get current positions to see if leverage is already correct
			positions, posErr := t.GetPositions()
			if posErr == nil {
				for _, pos := range positions {
					if pos["symbol"] == symbol {
						// Handle both float64 and *json.Number types for leverage
						levInterface := pos["leverage"]
						var currentLev float64
						if levVal, ok := levInterface.(float64); ok {
							currentLev = levVal
						} else if levNum, ok := levInterface.(*json.Number); ok {
							var convErr error
							currentLev, convErr = levNum.Float64()
							if convErr != nil {
								logger.Warnf("Could not convert leverage to float, proceeding with setting leverage: %v", convErr)
								break
							}
						} else {
							logger.Warnf("Could not get leverage, proceeding with setting leverage")
							break
						}

						if int(currentLev) == leverage {
							logger.Infof("  ✓ %s leverage is already %dx as requested", symbol, leverage)
							return nil
						}
						logger.Infof("  ℹ %s current leverage is %dx, requested %dx", symbol, int(currentLev), leverage)
						break
					}
				}
			}
		}

		// For other errors, return the original error
		return fmt.Errorf("failed to set leverage: %w", err)
	}

	logger.Infof("  ✓ %s leverage changed to %dx", symbol, leverage)

	// Wait 5 seconds after changing leverage (to avoid cooldown period errors)
	logger.Infof("  ⏱ Waiting 5 seconds for cooldown period...")
	time.Sleep(5 * time.Second)

	return nil
}

// OpenLong opens a long position
func (t *FuturesTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// First cancel all pending orders for this symbol (clean up old stop-loss and take-profit orders)
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel old pending orders (may not have any): %v", err)
	}

	// Set leverage (non-fatal if fails)
	if err := t.SetLeverage(symbol, leverage); err != nil {
		logger.Infof("  ⚠ Failed to set leverage for %s: %v", symbol, err)
		// Continue execution even if leverage setting fails
	}

	// Note: Margin mode should be set by the caller (AutoTrader) before opening position via SetMarginMode

	// Format quantity to correct precision
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// Check if formatted quantity is 0 (prevent rounding errors)
	quantityFloat, parseErr := strconv.ParseFloat(quantityStr, 64)
	if parseErr != nil || quantityFloat <= 0 {
		return nil, fmt.Errorf("position size too small, rounded to 0 (original: %.8f -> formatted: %s). Suggest increasing position amount or selecting a lower-priced coin", quantity, quantityStr)
	}

	// Check minimum notional value (Binance requires at least 10 USDT)
	if err := t.CheckMinNotional(symbol, quantityFloat); err != nil {
		return nil, err
	}

	// Create market buy order (using br ID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeBuy).
		PositionSide(futures.PositionSideTypeLong).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to open long position: %w", err)
	}

	logger.Infof("✓ Opened long position successfully: %s quantity: %s", symbol, quantityStr)
	logger.Infof("  Order ID: %d", order.OrderID)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// OpenShort opens a short position
func (t *FuturesTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// First cancel all pending orders for this symbol (clean up old stop-loss and take-profit orders)
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel old pending orders (may not have any): %v", err)
	}

	// Set leverage (non-fatal if fails)
	if err := t.SetLeverage(symbol, leverage); err != nil {
		logger.Infof("  ⚠ Failed to set leverage for %s: %v", symbol, err)
		// Continue execution even if leverage setting fails
	}

	// Note: Margin mode should be set by the caller (AutoTrader) before opening position via SetMarginMode

	// Format quantity to correct precision
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// Check if formatted quantity is 0 (prevent rounding errors)
	quantityFloat, parseErr := strconv.ParseFloat(quantityStr, 64)
	if parseErr != nil || quantityFloat <= 0 {
		return nil, fmt.Errorf("position size too small, rounded to 0 (original: %.8f -> formatted: %s). Suggest increasing position amount or selecting a lower-priced coin", quantity, quantityStr)
	}

	// Check minimum notional value (Binance requires at least 10 USDT)
	if err := t.CheckMinNotional(symbol, quantityFloat); err != nil {
		return nil, err
	}

	// Create market sell order (using br ID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to open short position: %w", err)
	}

	logger.Infof("✓ Opened short position successfully: %s quantity: %s", symbol, quantityStr)
	logger.Infof("  Order ID: %d", order.OrderID)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// CloseLong closes a long position
func (t *FuturesTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	// If quantity is 0, get current position quantity
	if quantity == 0 {
		positions, err := t.GetPositions()
		if err != nil {
			return nil, err
		}

		for _, pos := range positions {
			if pos["symbol"] == symbol && pos["side"] == "long" {
				// Handle both float64 and *json.Number types for positionAmt
				positionAmt, ok := pos["positionAmt"].(float64)
				if !ok {
					// Try to get as json.Number if it fails as float64
					positionAmtNum, ok2 := pos["positionAmt"].(*json.Number)
					if !ok2 {
						return nil, fmt.Errorf("failed to get position amount for %s", symbol)
					}
					var err error
					positionAmt, err = positionAmtNum.Float64()
					if err != nil {
						return nil, fmt.Errorf("failed to convert position amount to float for %s: %w", symbol, err)
					}
				}
				quantity = positionAmt
				break
			}
		}

		if quantity == 0 {
			return nil, fmt.Errorf("no long position found for %s", symbol)
		}
	}

	// Format quantity
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// Create market sell order (close long, using br ID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeSell).
		PositionSide(futures.PositionSideTypeLong).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to close long position: %w", err)
	}

	logger.Infof("✓ Closed long position successfully: %s quantity: %s", symbol, quantityStr)

	// After closing position, cancel all pending orders for this symbol (stop-loss and take-profit orders)
	// We cancel pending orders here because they are no longer relevant after position is closed
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel pending orders: %v", err)
	}

	// Trigger order sync after successful close
	if t.store != nil && t.traderID != "" && t.exchangeID != "" && t.exchangeType != "" {
		go func() {
			if err := t.SyncOrdersFromBinance(t.traderID, t.exchangeID, t.exchangeType, t.store); err != nil {
				logger.Infof("⚠️ Order sync after close long failed: %v", err)
			} else {
				logger.Infof("✅ Order sync triggered after closing long position for %s", symbol)
			}
		}()
	}

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// CloseShort closes a short position
func (t *FuturesTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	// If quantity is 0, get current position quantity
	if quantity == 0 {
		positions, err := t.GetPositions()
		if err != nil {
			return nil, err
		}

		for _, pos := range positions {
			if pos["symbol"] == symbol && pos["side"] == "short" {
				// Handle both float64 and *json.Number types for positionAmt
				positionAmt, ok := pos["positionAmt"].(float64)
				if !ok {
					// Try to get as json.Number if it fails as float64
					positionAmtNum, ok2 := pos["positionAmt"].(*json.Number)
					if !ok2 {
						return nil, fmt.Errorf("failed to get position amount for %s", symbol)
					}
					var err error
					positionAmt, err = positionAmtNum.Float64()
					if err != nil {
						return nil, fmt.Errorf("failed to convert position amount to float for %s: %w", symbol, err)
					}
				}
				quantity = -positionAmt // Short position quantity is negative, take absolute value
				break
			}
		}

		if quantity == 0 {
			return nil, fmt.Errorf("no short position found for %s", symbol)
		}
	}

	// Format quantity
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, err
	}

	// Create market buy order (close short, using br ID)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(futures.SideTypeBuy).
		PositionSide(futures.PositionSideTypeShort).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		NewClientOrderID(getBrOrderID()).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to close short position: %w", err)
	}

	logger.Infof("✓ Closed short position successfully: %s quantity: %s", symbol, quantityStr)

	// After closing position, cancel all pending orders for this symbol (stop-loss and take-profit orders)
	// We cancel pending orders here because they are no longer relevant after position is closed
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel pending orders: %v", err)
	}

	// Trigger order sync after successful close
	if t.store != nil && t.traderID != "" && t.exchangeID != "" && t.exchangeType != "" {
		go func() {
			if err := t.SyncOrdersFromBinance(t.traderID, t.exchangeID, t.exchangeType, t.store); err != nil {
				logger.Infof("⚠️ Order sync after close short failed: %v", err)
			} else {
				logger.Infof("✅ Order sync triggered after closing short position for %s", symbol)
			}
		}()
	}

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	return result, nil
}

// CancelStopLossOrders cancels only stop-loss orders (doesn't affect take-profit orders)
// Now uses both legacy API and new Algo Order API
func (t *FuturesTrader) CancelStopLossOrders(symbol string) error {
	canceledCount := 0
	var cancelErrors []error

	// 1. Cancel legacy stop-loss orders
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	orders, err := t.client.NewListOpenOrdersService().
		Symbol(symbol).
		Do(ctx)

	if err == nil {
		for _, order := range orders {
			orderType := string(order.Type)

			// Only cancel stop-loss orders (don't cancel take-profit orders)
			// Use string comparison since OrderType constants were removed in v2.8.9
			if orderType == "STOP_MARKET" || orderType == "STOP" {
				ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel1()
				_, err := t.client.NewCancelOrderService().
					Symbol(symbol).
					OrderID(order.OrderID).
					Do(ctx1)

				if err != nil {
					errMsg := fmt.Sprintf("Order ID %d: %v", order.OrderID, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel legacy stop-loss order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled legacy stop-loss order (Order ID: %d, Type: %s, Side: %s)", order.OrderID, orderType, order.PositionSide)
			}
		}
	}

	// 2. Cancel Algo stop-loss orders
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()
	algoOrders, err := t.client.NewListOpenAlgoOrdersService().
		Symbol(symbol).
		Do(ctx2)

	if err == nil {
		for _, algoOrder := range algoOrders {
			// Only cancel stop-loss orders
			if algoOrder.OrderType == futures.AlgoOrderTypeStopMarket || algoOrder.OrderType == futures.AlgoOrderTypeStop {
				ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel3()
				_, err := t.client.NewCancelAlgoOrderService().
					AlgoID(algoOrder.AlgoId).
					Do(ctx3)

				if err != nil {
					errMsg := fmt.Sprintf("Algo ID %d: %v", algoOrder.AlgoId, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel Algo stop-loss order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled Algo stop-loss order (Algo ID: %d, Type: %s)", algoOrder.AlgoId, algoOrder.OrderType)
			}
		}
	}

	if canceledCount == 0 && len(cancelErrors) == 0 {
		logger.Infof("  ℹ %s has no stop-loss orders to cancel", symbol)
	} else if canceledCount > 0 {
		logger.Infof("  ✓ Canceled %d stop-loss order(s) for %s", canceledCount, symbol)
	}

	// If all cancellations failed, return error
	if len(cancelErrors) > 0 && canceledCount == 0 {
		return fmt.Errorf("failed to cancel stop-loss orders: %v", cancelErrors)
	}

	return nil
}

// CancelTakeProfitOrders cancels only take-profit orders (doesn't affect stop-loss orders)
// Now uses both legacy API and new Algo Order API
func (t *FuturesTrader) CancelTakeProfitOrders(symbol string) error {
	canceledCount := 0
	var cancelErrors []error

	// 1. Cancel legacy take-profit orders
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	orders, err := t.client.NewListOpenOrdersService().
		Symbol(symbol).
		Do(ctx)

	if err == nil {
		for _, order := range orders {
			orderType := string(order.Type)

			// Only cancel take-profit orders (don't cancel stop-loss orders)
			// Use string comparison since OrderType constants were removed in v2.8.9
			if orderType == "TAKE_PROFIT_MARKET" || orderType == "TAKE_PROFIT" {
				ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel1()
				_, err := t.client.NewCancelOrderService().
					Symbol(symbol).
					OrderID(order.OrderID).
					Do(ctx1)

				if err != nil {
					errMsg := fmt.Sprintf("Order ID %d: %v", order.OrderID, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel legacy take-profit order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled legacy take-profit order (Order ID: %d, Type: %s, Side: %s)", order.OrderID, orderType, order.PositionSide)
			}
		}
	}

	// 2. Cancel Algo take-profit orders
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()
	algoOrders, err := t.client.NewListOpenAlgoOrdersService().
		Symbol(symbol).
		Do(ctx2)

	if err == nil {
		for _, algoOrder := range algoOrders {
			// Only cancel take-profit orders
			if algoOrder.OrderType == futures.AlgoOrderTypeTakeProfitMarket || algoOrder.OrderType == futures.AlgoOrderTypeTakeProfit {
				ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel3()
				_, err := t.client.NewCancelAlgoOrderService().
					AlgoID(algoOrder.AlgoId).
					Do(ctx3)

				if err != nil {
					errMsg := fmt.Sprintf("Algo ID %d: %v", algoOrder.AlgoId, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel Algo take-profit order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled Algo take-profit order (Algo ID: %d, Type: %s)", algoOrder.AlgoId, algoOrder.OrderType)
			}
		}
	}

	if canceledCount == 0 && len(cancelErrors) == 0 {
		logger.Infof("  ℹ %s has no take-profit orders to cancel", symbol)
	} else if canceledCount > 0 {
		logger.Infof("  ✓ Canceled %d take-profit order(s) for %s", canceledCount, symbol)
	}

	// If all cancellations failed, return error
	if len(cancelErrors) > 0 && canceledCount == 0 {
		return fmt.Errorf("failed to cancel take-profit orders: %v", cancelErrors)
	}

	return nil
}

// CancelAllOrders cancels all pending orders for this symbol
// Now uses both legacy API and new Algo Order API
func (t *FuturesTrader) CancelAllOrders(symbol string) error {
	// 1. Cancel all legacy orders
	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()
	err := t.client.NewCancelAllOpenOrdersService().
		Symbol(symbol).
		Do(ctx1)

	if err != nil {
		logger.Infof("  ⚠ Failed to cancel legacy orders: %v", err)
	} else {
		logger.Infof("  ✓ Canceled all legacy pending orders for %s", symbol)
	}

	// 2. Cancel all Algo orders
	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()
	err = t.client.NewCancelAllAlgoOpenOrdersService().
		Symbol(symbol).
		Do(ctx2)

	if err != nil {
		// Ignore "no algo orders" error
		if !contains(err.Error(), "no algo") && !contains(err.Error(), "No algo") {
			logger.Infof("  ⚠ Failed to cancel Algo orders: %v", err)
		}
	} else {
		logger.Infof("  ✓ Canceled all Algo orders for %s", symbol)
	}

	return nil
}

// CancelStopOrders cancels stop-loss and take-profit orders for this symbol (for adjusting stop-loss/take-profit positions)
// Uses both legacy API and new Algo Order API
func (t *FuturesTrader) CancelStopOrders(symbol string) error {
	canceledCount := 0
	var cancelErrors []error

	// 1. Cancel legacy stop-loss and take-profit orders
	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()
	orders, err := t.client.NewListOpenOrdersService().
		Symbol(symbol).
		Do(ctx1)

	if err == nil {
		for _, order := range orders {
			orderType := string(order.Type)

			// Cancel both stop-loss and take-profit orders
			if orderType == "STOP_MARKET" || orderType == "STOP" || orderType == "TAKE_PROFIT_MARKET" || orderType == "TAKE_PROFIT" {
				ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel2()
				_, err := t.client.NewCancelOrderService().
					Symbol(symbol).
					OrderID(order.OrderID).
					Do(ctx2)

				if err != nil {
					errMsg := fmt.Sprintf("Order ID %d: %v", order.OrderID, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel legacy stop/take-profit order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled legacy stop/take-profit order (Order ID: %d, Type: %s, Side: %s)", order.OrderID, orderType, order.PositionSide)
			}
		}
	}

	// 2. Cancel Algo stop-loss and take-profit orders
	ctx3, cancel3 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel3()
	algoOrders, err := t.client.NewListOpenAlgoOrdersService().
		Symbol(symbol).
		Do(ctx3)

	if err == nil {
		for _, algoOrder := range algoOrders {
			// Cancel both stop-loss and take-profit orders
			if algoOrder.OrderType == futures.AlgoOrderTypeStopMarket ||
				algoOrder.OrderType == futures.AlgoOrderTypeStop ||
				algoOrder.OrderType == futures.AlgoOrderTypeTakeProfitMarket ||
				algoOrder.OrderType == futures.AlgoOrderTypeTakeProfit {

				ctx4, cancel4 := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel4()
				_, err := t.client.NewCancelAlgoOrderService().
					AlgoID(algoOrder.AlgoId).
					Do(ctx4)

				if err != nil {
					errMsg := fmt.Sprintf("Algo ID %d: %v", algoOrder.AlgoId, err)
					cancelErrors = append(cancelErrors, fmt.Errorf("%s", errMsg))
					logger.Infof("  ⚠ Failed to cancel Algo stop/take-profit order: %s", errMsg)
					continue
				}

				canceledCount++
				logger.Infof("  ✓ Canceled Algo stop/take-profit order (Algo ID: %d, Type: %s)", algoOrder.AlgoId, algoOrder.OrderType)
			}
		}
	}

	if canceledCount == 0 && len(cancelErrors) == 0 {
		logger.Infof("  ℹ %s has no stop/take-profit orders to cancel", symbol)
	} else if canceledCount > 0 {
		logger.Infof("  ✓ Canceled %d stop/take-profit order(s) for %s", canceledCount, symbol)
	}

	// If all cancellations failed, return error
	if len(cancelErrors) > 0 && canceledCount == 0 {
		return fmt.Errorf("failed to cancel stop/take-profit orders: %v", cancelErrors)
	}

	return nil
}

// PartialClose 部分平仓
func (t *FuturesTrader) PartialClose(symbol string, side string, percentage float64) (map[string]interface{}, error) {
	if percentage <= 0 || percentage > 100 {
		return nil, fmt.Errorf("平仓百分比必须在0-100之间: %.2f", percentage)
	}

	// 获取当前持仓
	positions, err := t.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var currentPos *map[string]interface{}
	for i, pos := range positions {
		if pos["symbol"] == symbol {
			if (side == "long" && pos["side"] == "long") || (side == "short" && pos["side"] == "short") {
				currentPos = &positions[i]
				break
			}
		}
	}

	if currentPos == nil {
		return nil, fmt.Errorf("未找到 %s 的 %s 仓位", symbol, side)
	}

	// 计算部分平仓数量
	// Handle both float64 and *json.Number types for positionAmt
	positionAmt, ok := (*currentPos)["positionAmt"].(float64)
	if !ok {
		// Try to get as json.Number if it fails as float64
		positionAmtNum, ok2 := (*currentPos)["positionAmt"].(*json.Number)
		if !ok2 {
			return nil, fmt.Errorf("failed to get position amount for %s", symbol)
		}
		var err error
		positionAmt, err = positionAmtNum.Float64()
		if err != nil {
			return nil, fmt.Errorf("failed to convert position amount to float for %s: %w", symbol, err)
		}
	}
	absCurrentQty := math.Abs(positionAmt)
	closeQty := absCurrentQty * (percentage / 100.0)

	// 格式化数量
	quantityStr, err := t.FormatQuantity(symbol, closeQty)
	if err != nil {
		return nil, err
	}

	// 执行平仓操作
	var order *futures.CreateOrderResponse
	var sideType futures.SideType
	var posSideType futures.PositionSideType

	if side == "long" {
		sideType = futures.SideTypeSell
		posSideType = futures.PositionSideTypeLong
	} else {
		sideType = futures.SideTypeBuy
		posSideType = futures.PositionSideTypeShort
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err = t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(sideType).
		PositionSide(posSideType).
		Type(futures.OrderTypeMarket).
		Quantity(quantityStr).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("部分平仓失败: %w", err)
	}

	log.Printf("✓ 部分平仓成功: %s %s %.2f%% 数量: %s", symbol, side, percentage, quantityStr)

	result := make(map[string]interface{})
	result["orderId"] = order.OrderID
	result["symbol"] = order.Symbol
	result["status"] = order.Status
	result["quantity"] = quantityStr
	result["percentage"] = percentage
	return result, nil
}

// UpdateStopLoss 更新止损单
func (t *FuturesTrader) UpdateStopLoss(symbol string, positionSide string, newStopPrice float64) error {
	// 获取当前持仓数量
	positions, err := t.GetPositions()
	if err != nil {
		return fmt.Errorf("获取持仓失败: %w", err)
	}

	var currentQty float64
	for _, pos := range positions {
		if pos["symbol"] == symbol {
			if (positionSide == "LONG" && pos["side"] == "long") || (positionSide == "SHORT" && pos["side"] == "short") {
				// Handle both float64 and *json.Number types for positionAmt
				positionAmt, ok := pos["positionAmt"].(float64)
				if !ok {
					// Try to get as json.Number if it fails as float64
					positionAmtNum, ok2 := pos["positionAmt"].(*json.Number)
					if !ok2 {
						return fmt.Errorf("failed to get position amount for %s", symbol)
					}
					var err error
					positionAmt, err = positionAmtNum.Float64()
					if err != nil {
						return fmt.Errorf("failed to convert position amount to float for %s: %w", symbol, err)
					}
				}
				currentQty = math.Abs(positionAmt)
				break
			}
		}
	}

	if currentQty == 0 {
		return fmt.Errorf("未找到 %s 的 %s 仓位", symbol, positionSide)
	}

	// 首先尝试设置新的止损单，而不先取消旧的止损单
	err = t.SetStopLoss(symbol, positionSide, currentQty, newStopPrice)
	if err != nil {
		// 如果新止损设置失败，返回错误，不取消旧的止损单
		// 这样可以保持原有的止损保护
		return fmt.Errorf("设置新止损失败，保持原有止损保护: %w", err)
	}

	// 如果新止损设置成功，再取消旧的止损单
	if err := t.CancelStopLossOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消旧止损单失败（新止损已成功设置）: %v", err)
		// 不返回错误，因为新的止损已经成功设置
	}

	return nil
}

// UpdateTakeProfit 更新止盈单
func (t *FuturesTrader) UpdateTakeProfit(symbol string, positionSide string, newTakeProfitPrice float64) error {
	// 首先取消当前的止盈单
	if err := t.CancelTakeProfitOrders(symbol); err != nil {
		log.Printf("  ⚠ 取消旧止盈单失败（可能没有旧单）: %v", err)
	}

	// 获取当前持仓数量
	positions, err := t.GetPositions()
	if err != nil {
		return fmt.Errorf("获取持仓失败: %w", err)
	}

	var currentQty float64
	for _, pos := range positions {
		if pos["symbol"] == symbol {
			if (positionSide == "LONG" && pos["side"] == "long") || (positionSide == "SHORT" && pos["side"] == "short") {
				// Handle both float64 and *json.Number types for positionAmt
				positionAmt, ok := pos["positionAmt"].(float64)
				if !ok {
					// Try to get as json.Number if it fails as float64
					positionAmtNum, ok2 := pos["positionAmt"].(*json.Number)
					if !ok2 {
						return fmt.Errorf("failed to get position amount for %s", symbol)
					}
					var err error
					positionAmt, err = positionAmtNum.Float64()
					if err != nil {
						return fmt.Errorf("failed to convert position amount to float for %s: %w", symbol, err)
					}
				}
				currentQty = math.Abs(positionAmt)
				break
			}
		}
	}

	if currentQty == 0 {
		return fmt.Errorf("未找到 %s 的 %s 仓位", symbol, positionSide)
	}

	// 设置新的止盈单
	return t.SetTakeProfit(symbol, positionSide, currentQty, newTakeProfitPrice)
}

// GetMarketPrice 获取市场价格
func (t *FuturesTrader) GetMarketPrice(symbol string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	prices, err := t.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get price: %w", err)
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("price not found")
	}

	// 检查价格字符串是否为空
	if prices[0].Price == "" {
		return 0, fmt.Errorf("received empty price string for symbol %s", symbol)
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price '%s' for symbol %s: %w", prices[0].Price, symbol, err)
	}

	return price, nil
}

// CalculatePositionSize calculates position size
func (t *FuturesTrader) CalculatePositionSize(balance, riskPercent, price float64, leverage int) float64 {
	riskAmount := balance * (riskPercent / 100.0)
	positionValue := riskAmount * float64(leverage)
	quantity := positionValue / price
	return quantity
}

// SetStopLoss sets stop-loss order using traditional order API
// If algo order fails, fallback to traditional stop market order
func (t *FuturesTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	var side futures.SideType
	var posSide futures.PositionSideType

	if positionSide == "LONG" {
		side = futures.SideTypeSell
		posSide = futures.PositionSideTypeLong
	} else {
		side = futures.SideTypeBuy
		posSide = futures.PositionSideTypeShort
	}

	// 🔍 获取当前市场价格用于验证
	currentPrice, err := t.GetMarketPrice(symbol)
	if err != nil {
		logger.Warnf("⚠️ 无法获取 %s 当前价格用于止损验证: %v", symbol, err)
	} else {
		// 🔍 验证止损价格合理性
		priceDiff := math.Abs(stopPrice - currentPrice)
		priceDiffPercent := (priceDiff / currentPrice) * 100

		logger.Infof("📊 止损价格验证: %s 当前价格=%.4f, 止损价格=%.4f, 价差=%.4f (%.2f%%)",
			symbol, currentPrice, stopPrice, priceDiff, priceDiffPercent)

		// 做多时止损应该低于当前价格，做空时止损应该高于当前价格
		if positionSide == "LONG" && stopPrice >= currentPrice {
			logger.Errorf("❌ 做多止损价格设置错误: 止损价(%.4f) >= 当前价(%.4f)", stopPrice, currentPrice)
			return fmt.Errorf("long position stop loss price must be below current price: stop=%.4f, current=%.4f", stopPrice, currentPrice)
		}
		if positionSide == "SHORT" && stopPrice <= currentPrice {
			logger.Errorf("❌ 做空止损价格设置错误: 止损价(%.4f) <= 当前价(%.4f)", stopPrice, currentPrice)
			return fmt.Errorf("short position stop loss price must be above current price: stop=%.4f, current=%.4f", stopPrice, currentPrice)
		}

		// 检查价差是否过小（小于0.1%可能触发立即执行）
		if priceDiffPercent < 0.1 {
			logger.Warnf("⚠️ 止损价格与当前价格过于接近 (%.2f%% < 0.1%%)，可能导致订单立即触发", priceDiffPercent)
		}
	}

	// Format price to correct precision
	priceStr, err := t.FormatPrice(symbol, stopPrice)
	if err != nil {
		return fmt.Errorf("failed to format price: %w", err)
	}

	// Format quantity to correct precision
	qtyStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return fmt.Errorf("failed to format quantity: %w", err)
	}

	// Validate formatted strings are not empty
	if priceStr == "" {
		logger.Warnf("⚠️ formatted price is empty for symbol %s, stopPrice: %f, clearing cache and retrying", symbol, stopPrice)
		// Clear the cache to force reloading from file
		t.precisionManager.ClearCache()
		// Try formatting again after clearing cache
		priceStr, err = t.FormatPrice(symbol, stopPrice)
		if err != nil {
			return fmt.Errorf("failed to format price after cache clear: %w", err)
		}
		if priceStr == "" {
			return fmt.Errorf("formatted price is still empty for symbol %s, stopPrice: %f", symbol, stopPrice)
		}
		logger.Infof("✅ Successfully formatted price after cache clear: %s", priceStr)
	}
	if qtyStr == "" {
		logger.Warnf("⚠️ formatted quantity is empty for symbol %s, quantity: %f, clearing cache and retrying", symbol, quantity)
		// Clear the cache to force reloading from file
		t.precisionManager.ClearCache()
		// Try formatting again after clearing cache
		qtyStr, err = t.FormatQuantity(symbol, quantity)
		if err != nil {
			return fmt.Errorf("failed to format quantity after cache clear: %w", err)
		}
		if qtyStr == "" {
			return fmt.Errorf("formatted quantity is still empty for symbol %s, quantity: %f", symbol, quantity)
		}
		logger.Infof("✅ Successfully formatted quantity after cache clear: %s", qtyStr)
	}

	// First, try the traditional stop market order (this should work for most cases)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Try traditional order first
	order := t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		PositionSide(posSide).
		Type("STOP_MARKET"). // Use string constant for stop market order
		StopPrice(priceStr). // Use StopPrice for traditional API
		Quantity(qtyStr).    // Use quantity for traditional API
		WorkingType(futures.WorkingTypeContractPrice)
		// Remove ReduceOnly parameter to avoid -1106 error

	_, err = order.Do(ctx)

	if err != nil {
		// If traditional order fails, try the algo order as fallback
		logger.Infof("  ℹ️ Traditional stop-market order failed: %v, trying algo order", err)

		// Generate a new client algo ID for the algo order
		algoID := getBrOrderID()

		_, algoErr := t.client.NewCreateAlgoOrderService().
			Symbol(symbol).
			Side(side).
			PositionSide(posSide).
			Type(futures.AlgoOrderTypeStopMarket).
			TriggerPrice(priceStr). // Use TriggerPrice for algo API
			WorkingType(futures.WorkingTypeContractPrice).
			ClosePosition(true). // Close entire position for stop loss
			// Remove ReduceOnly parameter to avoid -1106 error
			ClientAlgoId(algoID).
			Do(ctx)

		if algoErr != nil {
			// 🔍 详细错误信息输出
			logger.Errorf("❌ 止损设置完全失败:")
			logger.Errorf("   交易对: %s", symbol)
			logger.Errorf("   持仓方向: %s", positionSide)
			logger.Errorf("   止损价格: %.4f", stopPrice)
			if currentPrice > 0 {
				logger.Errorf("   当前价格: %.4f", currentPrice)
				logger.Errorf("   价格差异: %.4f (%.2f%%)", math.Abs(stopPrice-currentPrice), (math.Abs(stopPrice-currentPrice)/currentPrice)*100)
			}
			logger.Errorf("   传统订单错误: %v", err)
			logger.Errorf("   算法订单错误: %v", algoErr)

			return fmt.Errorf("failed to set stop-loss with both traditional (%v) and algo (%v) orders - check price settings", err, algoErr)
		}

		logger.Infof("  Stop-loss price set (Algo Order): %s", priceStr)
		return nil
	}

	logger.Infof("  Stop-loss price set (Traditional Order): %s", priceStr)
	return nil
}

// SetTakeProfit sets take-profit order using new Algo Order API
// Binance has migrated stop orders to Algo Order system (error -4120 STOP_ORDER_SWITCH_ALGO)
// SetTakeProfit sets take-profit order using traditional order API
// If algo order fails, fallback to traditional take-profit market order
func (t *FuturesTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	var side futures.SideType
	var posSide futures.PositionSideType

	if positionSide == "LONG" {
		side = futures.SideTypeSell
		posSide = futures.PositionSideTypeLong
	} else {
		side = futures.SideTypeBuy
		posSide = futures.PositionSideTypeShort
	}

	// Format quantity to correct precision
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return fmt.Errorf("failed to format quantity: %w", err)
	}

	// Format price to correct precision
	priceStr, err := t.FormatPrice(symbol, takeProfitPrice)
	if err != nil {
		return fmt.Errorf("failed to format price: %w", err)
	}

	// First, try the traditional take-profit market order
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = t.client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		PositionSide(posSide).
		Type("TAKE_PROFIT_MARKET"). // Use string constant for take-profit market order
		StopPrice(priceStr).        // Use StopPrice for traditional API
		Quantity(quantityStr).      // Use quantity for traditional API
		WorkingType(futures.WorkingTypeContractPrice).
		Do(ctx)

	if err != nil {
		// If traditional order fails, try the algo order as fallback
		logger.Infof("  ℹ️ Traditional take-profit-market order failed: %v, trying algo order", err)

		// Generate a new client algo ID for the algo order
		algoID := getBrOrderID()

		_, algoErr := t.client.NewCreateAlgoOrderService().
			Symbol(symbol).
			Side(side).
			PositionSide(posSide).
			Type(futures.AlgoOrderTypeTakeProfitMarket).
			TriggerPrice(priceStr). // Use TriggerPrice for algo API
			WorkingType(futures.WorkingTypeContractPrice).
			Quantity(quantityStr). // Use quantity instead of ClosePosition
			ClientAlgoId(algoID).
			Do(ctx)

		if algoErr != nil {
			return fmt.Errorf("failed to set take-profit with both traditional (%v) and algo (%v) orders", err, algoErr)
		}

		logger.Infof("  Take-profit price set (Algo Order): %s, quantity: %s", priceStr, quantityStr)
		return nil
	}

	logger.Infof("  Take-profit price set (Traditional Order): %s, quantity: %s", priceStr, quantityStr)
	return nil
}

// SetTrailingStop sets trailing stop-loss order using new Algo Order API
// Binance has migrated stop orders to Algo Order system (error -4120 STOP_ORDER_SWITCH_ALGO)
func (t *FuturesTrader) SetTrailingStop(symbol string, positionSide string, quantity, callbackRate, activationPrice float64) error {
	var side futures.SideType
	var posSide futures.PositionSideType

	if positionSide == "LONG" {
		side = futures.SideTypeSell
		posSide = futures.PositionSideTypeLong
	} else {
		side = futures.SideTypeBuy
		posSide = futures.PositionSideTypeShort
	}

	// Calculate activation price based on current market price if activationPrice is 0
	currentPrice := activationPrice
	if currentPrice == 0 {
		price, err := t.GetMarketPrice(symbol)
		if err != nil {
			return fmt.Errorf("failed to get market price for trailing stop: %w", err)
		}
		currentPrice = price
	}

	// Format quantity to correct precision
	quantityStr, err := t.FormatQuantity(symbol, quantity)
	if err != nil {
		return fmt.Errorf("failed to format quantity: %w", err)
	}

	// Format activation price to correct precision
	activationPriceStr, err := t.FormatPrice(symbol, currentPrice)
	if err != nil {
		return fmt.Errorf("failed to format activation price: %w", err)
	}

	// Binance API expects callbackRate in range [0.1, 10] where 1 = 1%
	// If callbackRate is 0 or invalid, use a default value of 0.5 (0.5%)
	if callbackRate <= 0 {
		logger.Infof("  ⚠️ Callback rate is %.2f, using default value of 0.5 (0.5%%)", callbackRate)
		callbackRate = 0.5
	}

	// Validate range (now that we have a positive value)
	if callbackRate > 10.0 {
		logger.Infof("  ⚠️ Callback rate %.2f exceeds maximum of 10.0, capping at 10.0", callbackRate)
		callbackRate = 10.0
	}

	// Use new Algo Order API with trailing stop parameters
	// Note: TRAILING_STOP_MARKET does not support ClosePosition parameter
	// Must use Quantity instead
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err = t.client.NewCreateAlgoOrderService().
		Symbol(symbol).
		Side(side).
		PositionSide(posSide).
		Type(futures.AlgoOrderTypeTrailingStopMarket).   // Use trailing stop market type
		ActivationPrice(activationPriceStr).             // Use formatted price
		CallbackRate(fmt.Sprintf("%.1f", callbackRate)). // Callback rate: 1 = 1%, range [0.1, 10]
		Quantity(quantityStr).                           // Use quantity instead of ClosePosition
		ClientAlgoId(getBrOrderID()).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to set trailing stop: %w", err)
	}

	logger.Infof("  Trailing stop set (Algo Order): callback rate: %.1f%%, activation price: %s, quantity: %s",
		callbackRate, activationPriceStr, quantityStr)
	return nil
}

// GetMinNotional gets minimum notional value (Binance requirement)
func (t *FuturesTrader) GetMinNotional(symbol string) float64 {
	// Use conservative default value of 10 USDT to ensure order passes exchange validation
	return 10.0
}

// CheckMinNotional checks if order meets minimum notional value requirement
func (t *FuturesTrader) CheckMinNotional(symbol string, quantity float64) error {
	price, err := t.GetMarketPrice(symbol)
	if err != nil {
		return fmt.Errorf("failed to get market price: %w", err)
	}

	notionalValue := quantity * price
	minNotional := t.GetMinNotional(symbol)

	if notionalValue < minNotional {
		return fmt.Errorf(
			"order amount %.2f USDT is below minimum requirement %.2f USDT (quantity: %.4f, price: %.4f)",
			notionalValue, minNotional, quantity, price,
		)
	}

	return nil
}

// GetSymbolPrecision gets the quantity precision for a trading pair
// Enhanced version with improved network handling for large data responses
func (t *FuturesTrader) GetSymbolPrecision(symbol string) (int, error) {
	// Try to get exchange info with enhanced retry logic for network errors
	var exchangeInfo *futures.ExchangeInfo
	var err error

	// First try with optimized timeout and transport settings for better reliability
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)

	// Enhance HTTP client configuration for large data handling
	if t.client.HTTPClient != nil {
		transport, ok := t.client.HTTPClient.Transport.(*http.Transport)
		if ok {
			// Optimize transport settings for large responses
			transport.ResponseHeaderTimeout = 30 * time.Second
			transport.TLSHandshakeTimeout = 15 * time.Second
			transport.IdleConnTimeout = 120 * time.Second
			transport.MaxIdleConns = 100
			transport.MaxIdleConnsPerHost = 10
		}
	}

	exchangeInfo, err = t.client.NewExchangeInfoService().Do(ctx)
	cancel()

	if err == nil {
		// Success on first try, no need for retry
		for _, s := range exchangeInfo.Symbols {
			if s.Symbol == symbol {
				// Get precision from LOT_SIZE filter
				for _, filter := range s.Filters {
					if filter["filterType"] == "LOT_SIZE" {
						stepSize := filter["stepSize"].(string)
						precision := calculatePrecision(stepSize)
						logger.Infof("  %s quantity precision: %d (stepSize: %s)", symbol, precision, stepSize)
						return precision, nil
					}
				}
			}
		}

		logger.Errorf("❌ %s precision information not found in exchange info", symbol)
		return 0, fmt.Errorf("precision information not found for symbol %s", symbol)
	}

	// Enhanced retry logic for network-related errors including EOF
	isRetryable := strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "i/o timeout") ||
		strings.Contains(err.Error(), "unexpected EOF") ||
		strings.Contains(err.Error(), "connection closed") ||
		strings.Contains(err.Error(), "broken pipe")

	if isRetryable {
		// Perform retry logic with exponential backoff for network-related errors
		maxRetries := 3
		baseDelay := 1 * time.Second

		for attempt := 1; attempt <= maxRetries; attempt++ {
			logger.Infof("❌ Exchange info API call failed (attempt %d/%d): %v", attempt, maxRetries, err)

			// Exponential backoff with jitter
			waitTime := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			// Add simple jitter to prevent thundering herd
			jitter := time.Duration((time.Now().UnixNano() % 500) * int64(time.Millisecond))
			waitTime += jitter

			logger.Infof("⏳ Retrying in %v...", waitTime)
			time.Sleep(waitTime)

			// Use shorter timeout for retries to be more responsive
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			exchangeInfo, err = t.client.NewExchangeInfoService().Do(ctx)
			cancel()

			if err == nil {
				// Success, exit retry loop
				logger.Infof("✅ Exchange info API call succeeded on attempt %d", attempt)
				break
			}

			// Check again if error is still retryable
			isRetryable = strings.Contains(err.Error(), "EOF") ||
				strings.Contains(err.Error(), "connection reset") ||
				strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "i/o timeout") ||
				strings.Contains(err.Error(), "unexpected EOF") ||
				strings.Contains(err.Error(), "connection closed") ||
				strings.Contains(err.Error(), "broken pipe")

			if !isRetryable {
				// Non-retryable error, break retry loop
				logger.Warnf("⚠️ Non-retryable error encountered: %v", err)
				break
			}
		}
	}

	if err != nil {
		logger.Errorf("❌ Failed to get exchange info for %s after all retries: %v", symbol, err)
		// 禁止备选方案 - 直接返回错误
		return 0, fmt.Errorf("precision retrieval failed for %s: %w", symbol, err)
	}

	// Process successful response
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			// Get precision from LOT_SIZE filter
			for _, filter := range s.Filters {
				if filter["filterType"] == "LOT_SIZE" {
					stepSize := filter["stepSize"].(string)
					precision := calculatePrecision(stepSize)
					logger.Infof("  %s quantity precision: %d (stepSize: %s)", symbol, precision, stepSize)
					return precision, nil
				}
			}
		}
	}

	logger.Errorf("❌ %s precision information not found in exchange info after processing", symbol)
	return 0, fmt.Errorf("precision information not found for symbol %s after full processing", symbol)
}

// calculatePrecision calculates precision from stepSize
func calculatePrecision(stepSize string) int {
	// Remove trailing zeros
	stepSize = trimTrailingZeros(stepSize)

	// Find decimal point
	dotIndex := -1
	for i := 0; i < len(stepSize); i++ {
		if stepSize[i] == '.' {
			dotIndex = i
			break
		}
	}

	// If no decimal point or decimal point is at the end, precision is 0
	if dotIndex == -1 || dotIndex == len(stepSize)-1 {
		return 0
	}

	// Return number of digits after decimal point
	return len(stepSize) - dotIndex - 1
}

// trimTrailingZeros removes trailing zeros
func trimTrailingZeros(s string) string {
	// If no decimal point, return directly
	if !stringContains(s, ".") {
		return s
	}

	// Iterate backwards to remove trailing zeros
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}

	// If last character is decimal point, remove it too
	if len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}

	return s
}

// FormatQuantity formats quantity to correct precision with step size alignment
// Now uses the unified PrecisionManager exclusively
func (t *FuturesTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	logger.Debugf("🔍 BinanceTrader - FormatQuantity: symbol=%s, raw_quantity=%.8f", symbol, quantity)

	// 直接使用统一的精度管理器
	formatted, err := t.precisionManager.FormatQuantityWithValidation(symbol, quantity)
	if err != nil {
		logger.Errorf("❌ FormatQuantity failed with PrecisionManager: %v", err)
		return "", err
	}

	logger.Debugf("✅ BinanceTrader - Final formatted quantity: %s", formatted)
	return formatted, nil
}

// formatQuantityLegacy 旧的格式化方法，已弃用
func (t *FuturesTrader) formatQuantityLegacy(symbol string, quantity float64) (string, error) {
	logger.Warnf("⚠️ formatQuantityLegacy 已弃用，使用PrecisionManager替代: symbol=%s", symbol)
	// 直接使用PrecisionManager进行格式化
	return t.precisionManager.FormatQuantityWithValidation(symbol, quantity)
}

// GetSymbolStepSize 获取交易对的step size（已弃用，请使用PrecisionManager）
func (t *FuturesTrader) GetSymbolStepSize(symbol string) (float64, error) {
	logger.Warnf("⚠️ GetSymbolStepSize 已弃用，使用PrecisionManager替代: symbol=%s", symbol)
	// 获取精度信息
	info, err := t.precisionManager.GetPrecisionInfo(symbol)
	if err != nil {
		logger.Errorf("❌ GetSymbolStepSize fallback error: %v", err)
		// 返回默认值
		return 0.001, nil
	}
	return info.StepSize, nil
}

// GetPricePrecision 获取交易对的价格精度（已弃用，请使用PrecisionManager）
func (t *FuturesTrader) GetPricePrecision(symbol string) (int, error) {
	logger.Warnf("⚠️ GetPricePrecision 已弃用，使用PrecisionManager替代: symbol=%s", symbol)
	// 获取精度信息
	info, err := t.precisionManager.GetPrecisionInfo(symbol)
	if err != nil {
		logger.Errorf("❌ GetPricePrecision fallback error: %v", err)
		// 返回默认值
		return 2, nil
	}
	return info.Precision, nil
}

// FormatPrice formats price to correct precision
// Now uses the unified PrecisionManager exclusively
func (t *FuturesTrader) FormatPrice(symbol string, price float64) (string, error) {
	logger.Debugf("🔍 BinanceTrader - FormatPrice: symbol=%s, raw_price=%.8f", symbol, price)

	// 直接使用统一的精度管理器
	formatted, err := t.precisionManager.FormatPriceWithValidation(symbol, price)
	if err != nil {
		logger.Errorf("❌ FormatPrice failed with PrecisionManager: %v", err)
		return "", err
	}

	logger.Debugf("✅ BinanceTrader - Final formatted price: %s", formatted)
	return formatted, nil
}

// formatPriceLegacy 旧的价格格式化方法，已弃用
func (t *FuturesTrader) formatPriceLegacy(symbol string, price float64) (string, error) {
	logger.Warnf("⚠️ formatPriceLegacy 已弃用，使用PrecisionManager替代: symbol=%s", symbol)
	// 直接使用PrecisionManager进行格式化
	return t.precisionManager.FormatPriceWithValidation(symbol, price)
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetOrderStatus gets order status
func (t *FuturesTrader) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	// Convert orderID to int64
	orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %s", orderID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	order, err := t.client.NewGetOrderService().
		Symbol(symbol).
		OrderID(orderIDInt).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status: %w", err)
	}

	// Parse execution price
	avgPrice := 0.0
	if order.AvgPrice != "" {
		var err error
		avgPrice, err = strconv.ParseFloat(order.AvgPrice, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析平均成交价 %s: %v，使用0", order.AvgPrice, err)
			avgPrice = 0
		}
	} else {
		logger.Debugf("📝 订单平均成交价为空，设置为0")
	}
	executedQty := 0.0
	if order.ExecutedQuantity != "" {
		var err error
		executedQty, err = strconv.ParseFloat(order.ExecutedQuantity, 64)
		if err != nil {
			logger.Warnf("⚠️ 无法解析已成交量 %s: %v，使用0", order.ExecutedQuantity, err)
			executedQty = 0
		}
	} else {
		logger.Debugf("📝 订单已成交量为空，设置为0")
	}

	result := map[string]interface{}{
		"orderId":     order.OrderID,
		"symbol":      order.Symbol,
		"status":      string(order.Status),
		"avgPrice":    avgPrice,
		"executedQty": executedQty,
		"side":        string(order.Side),
		"type":        string(order.Type),
		"time":        order.Time,
		"updateTime":  order.UpdateTime,
	}

	// Binance futures commission fee needs to be obtained through GetUserTrades, not retrieved here for now
	// Can be obtained later through WebSocket or separate query
	result["commission"] = 0.0

	return result, nil
}

// GetClosedPnL retrieves recent closing trades from Binance Futures
// Note: Binance does NOT have a position history API, only trade history.
// This returns individual closing trades (realizedPnl != 0) for real-time position closure detection.
// NOT suitable for historical position reconstruction - use only for matching recent closures.
func (t *FuturesTrader) GetClosedPnL(startTime time.Time, limit int) ([]ClosedPnLRecord, error) {
	trades, err := t.GetTrades(startTime, limit)
	if err != nil {
		return nil, err
	}

	// Filter only closing trades (realizedPnl != 0) and convert to ClosedPnLRecord
	var records []ClosedPnLRecord
	for _, trade := range trades {
		if trade.RealizedPnL == 0 {
			continue // Skip opening trades
		}

		// Determine side from trade
		side := "long"
		if trade.PositionSide == "SHORT" || trade.PositionSide == "short" {
			side = "short"
		} else if trade.PositionSide == "BOTH" || trade.PositionSide == "" {
			// One-way mode: selling closes long, buying closes short
			if trade.Side == "SELL" || trade.Side == "Sell" {
				side = "long"
			} else {
				side = "short"
			}
		}

		// Calculate entry price from PnL (mathematically accurate for this trade)
		var entryPrice float64
		if trade.Quantity > 0 {
			if side == "long" {
				entryPrice = trade.Price - trade.RealizedPnL/trade.Quantity
			} else {
				entryPrice = trade.Price + trade.RealizedPnL/trade.Quantity
			}
		}

		records = append(records, ClosedPnLRecord{
			Symbol:      trade.Symbol,
			Side:        side,
			EntryPrice:  entryPrice,
			ExitPrice:   trade.Price,
			Quantity:    trade.Quantity,
			RealizedPnL: trade.RealizedPnL,
			Fee:         trade.Fee,
			ExitTime:    trade.Time,
			EntryTime:   trade.Time, // Approximate
			OrderID:     trade.TradeID,
			ExchangeID:  trade.TradeID,
			CloseType:   "unknown",
		})
	}

	return records, nil
}

// GetTrades retrieves trade history from Binance Futures using Income API
// Note: Income API has delays (~minutes), for real-time use GetTradesForSymbol instead
func (t *FuturesTrader) GetTrades(startTime time.Time, limit int) ([]TradeRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Use Income API to get REALIZED_PNL records (all symbols)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	incomes, err := t.client.NewGetIncomeHistoryService().
		IncomeType("REALIZED_PNL").
		StartTime(startTime.UnixMilli()).
		Limit(int64(limit)).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get income history: %w", err)
	}

	var trades []TradeRecord
	for _, income := range incomes {
		pnl, _ := strconv.ParseFloat(income.Income, 64)
		if pnl == 0 {
			continue // Skip zero PnL records
		}

		// Income API doesn't provide full trade details, create a minimal record
		// This is mainly used for detecting recent closures, not historical reconstruction
		trade := TradeRecord{
			TradeID:     strconv.FormatInt(income.TranID, 10),
			Symbol:      income.Symbol,
			RealizedPnL: pnl,
			Time:        time.UnixMilli(income.Time).UTC(),
			// Note: Income API doesn't provide price, quantity, side, fee
			// For accurate data, use GetTradesForSymbol with specific symbol
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradesForSymbol retrieves trade history for a specific symbol
// This is more reliable than using Income API which may have delays
func (t *FuturesTrader) GetTradesForSymbol(symbol string, startTime time.Time, limit int) ([]TradeRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	accountTrades, err := t.client.NewListAccountTradeService().
		Symbol(symbol).
		StartTime(startTime.UnixMilli()).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade history for %s: %w", symbol, err)
	}

	var trades []TradeRecord
	for _, at := range accountTrades {
		price, _ := strconv.ParseFloat(at.Price, 64)
		qty, _ := strconv.ParseFloat(at.Quantity, 64)
		fee, _ := strconv.ParseFloat(at.Commission, 64)
		pnl, _ := strconv.ParseFloat(at.RealizedPnl, 64)

		trade := TradeRecord{
			TradeID:      strconv.FormatInt(at.ID, 10),
			Symbol:       at.Symbol,
			Side:         string(at.Side),
			PositionSide: string(at.PositionSide),
			Price:        price,
			Quantity:     qty,
			RealizedPnL:  pnl,
			Fee:          fee,
			Time:         time.UnixMilli(at.Time).UTC(),
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetTradesForSymbolFromID retrieves trade history for a specific symbol starting from a given trade ID
// This is used for incremental sync - only fetch new trades since last sync
func (t *FuturesTrader) GetTradesForSymbolFromID(symbol string, fromID int64, limit int) ([]TradeRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	accountTrades, err := t.client.NewListAccountTradeService().
		Symbol(symbol).
		FromID(fromID).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade history for %s from ID %d: %w", symbol, fromID, err)
	}

	var trades []TradeRecord
	for _, at := range accountTrades {
		price, _ := strconv.ParseFloat(at.Price, 64)
		qty, _ := strconv.ParseFloat(at.Quantity, 64)
		fee, _ := strconv.ParseFloat(at.Commission, 64)
		pnl, _ := strconv.ParseFloat(at.RealizedPnl, 64)

		trade := TradeRecord{
			TradeID:      strconv.FormatInt(at.ID, 10),
			Symbol:       at.Symbol,
			Side:         string(at.Side),
			PositionSide: string(at.PositionSide),
			Price:        price,
			Quantity:     qty,
			RealizedPnL:  pnl,
			Fee:          fee,
			Time:         time.UnixMilli(at.Time).UTC(),
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// GetCommissionSymbols returns symbols that have new commission records since lastSyncTime
// COMMISSION income is generated for every trade, so this is more reliable than REALIZED_PNL
func (t *FuturesTrader) GetCommissionSymbols(lastSyncTime time.Time) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	incomes, err := t.client.NewGetIncomeHistoryService().
		IncomeType("COMMISSION").
		StartTime(lastSyncTime.UnixMilli()).
		Limit(1000).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get commission history: %w", err)
	}

	symbolMap := make(map[string]bool)
	for _, income := range incomes {
		if income.Symbol != "" {
			symbolMap[income.Symbol] = true
		}
	}

	var symbols []string
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// GetPnLSymbols returns symbols that have REALIZED_PNL records since lastSyncTime
// This is a fallback when COMMISSION detection fails (VIP users, BNB fee discount)
func (t *FuturesTrader) GetPnLSymbols(lastSyncTime time.Time) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	incomes, err := t.client.NewGetIncomeHistoryService().
		IncomeType("REALIZED_PNL").
		StartTime(lastSyncTime.UnixMilli()).
		Limit(1000).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get PnL history: %w", err)
	}

	symbolMap := make(map[string]bool)
	for _, income := range incomes {
		if income.Symbol != "" {
			symbolMap[income.Symbol] = true
		}
	}

	var symbols []string
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

// GetOpenOrders gets all open/pending orders for a symbol
func (t *FuturesTrader) GetOpenOrders(symbol string) ([]OpenOrder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	orders, err := t.client.NewListOpenOrdersService().
		Symbol(symbol).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	var result []OpenOrder
	for _, order := range orders {
		// Parse quantity
		quantity := 0.0
		if order.OrigQuantity != "" {
			var err error
			quantity, err = strconv.ParseFloat(order.OrigQuantity, 64)
			if err != nil {
				logger.Warnf("⚠️ 无法解析原始数量 %s: %v，使用0", order.OrigQuantity, err)
				quantity = 0
			}
		} else {
			logger.Debugf("📝 订单原始数量为空，设置为0")
		}
		// Parse price (may be empty for market orders)
		price := 0.0
		if order.Price != "" {
			var err error
			price, err = strconv.ParseFloat(order.Price, 64)
			if err != nil {
				logger.Warnf("⚠️ 无法解析订单价格 %s: %v，使用0", order.Price, err)
				price = 0
			}
		} else {
			logger.Debugf("📝 订单价格为空，设置为0")
		}
		// Parse stop price (may be empty)
		stopPrice := 0.0
		if order.StopPrice != "" {
			var err error
			stopPrice, err = strconv.ParseFloat(order.StopPrice, 64)
			if err != nil {
				logger.Warnf("⚠️ 无法解析止损价格 %s: %v，使用0", order.StopPrice, err)
				stopPrice = 0
			}
		} else {
			logger.Debugf("📝 订单止损价格为空，设置为0")
		}

		result = append(result, OpenOrder{
			OrderID:      fmt.Sprintf("%d", order.OrderID),
			Symbol:       order.Symbol,
			Side:         string(order.Side),
			PositionSide: string(order.PositionSide),
			Type:         string(order.Type),
			Price:        price,
			StopPrice:    stopPrice,
			Quantity:     quantity,
			Status:       string(order.Status),
		})
	}

	return result, nil
}

// CustomTransport 自定义传输层，用于在请求头中添加目标端点信息
type CustomTransport struct {
	Transport      http.RoundTripper
	TargetEndpoint string // 真正的目标端点
}

// RoundTrip 实现RoundTripper接口
func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 如果我们正在使用代理，将真实的目标端点添加到请求头中
	// 这样代理就知道应该将请求转发到哪里
	if ct.TargetEndpoint != "" {
		originalTarget := req.Header.Get("X-Target-URL")
		if originalTarget != "" && originalTarget != ct.TargetEndpoint {
			logger.Warnf("⚠️ X-Target-URL header already exists with different value: %s, replacing with: %s", originalTarget, ct.TargetEndpoint)
		}
		req.Header.Set("X-Target-URL", ct.TargetEndpoint)
	}

	resp, err := ct.Transport.RoundTrip(req)
	if err != nil {
		logger.Errorf("❌ Request failed: %v", err)
		return resp, err
	}
	return resp, err
}

// SetTraderInfo sets trader identification info for sync operations
func (t *FuturesTrader) SetTraderInfo(traderID, exchangeID, exchangeType string, store *store.Store) {
	t.traderID = traderID
	t.exchangeID = exchangeID
	t.exchangeType = exchangeType
	t.store = store
}
