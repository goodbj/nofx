package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/gin-gonic/gin"
)

// BinanceProxyService 币安代理服务，专门负责绕过币安监管，获取数据和执行命令
type BinanceProxyService struct {
	// 可以在这里添加配置或共享资源
}

// maskString 用于隐藏敏感信息，只显示前3位和后3位，中间用***代替
func maskString(s string) string {
	if len(s) <= 6 {
		return "***"
	}
	return s[:3] + "***" + s[len(s)-3:]
}

// NewBinanceProxyService 创建新的币安代理服务实例
func NewBinanceProxyService() *BinanceProxyService {
	return &BinanceProxyService{}
}

// GetBalance 获取账户余额
func (s *BinanceProxyService) GetBalance(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] GetBalance - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] GetBalance - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	fmt.Printf("[DEBUG] GetBalance - Creating client with API Key: %s, Custom API URL: %s\n",
		maskString(apiKey), customAPIURL)

	client := s.createClient(apiKey, secretKey, customAPIURL)

	fmt.Printf("[DEBUG] GetBalance - Client created, attempting to call Binance API...\n")

	// 调用币安API获取账户信息
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	account, err := client.NewGetAccountService().Do(ctx)
	cancel()

	if err != nil {
		// 即使发生错误，也要返回标准的数据格式
		errorResponse := gin.H{
			"data":  nil,
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// 添加调试日志，输出从币安API获取的原始数据
	fmt.Printf("[DEBUG] GetAccountInfo - Raw API Response - TotalWalletBalance: %s, AvailableBalance: %s, TotalUnrealizedProfit: %s\n",
		account.TotalWalletBalance, account.AvailableBalance, account.TotalUnrealizedProfit)

	// 检查账户资产信息
	fmt.Printf("[DEBUG] GetAccountInfo - Account Assets Count: %d\n", len(account.Assets))
	for i, asset := range account.Assets {
		if i < 5 { // 只打印前5个资产，避免日志过多
			fmt.Printf("[DEBUG] GetAccountInfo - Asset[%d]: %s, WalletBalance: %s, UnrealizedProfit: %s\n",
				i, asset.Asset, asset.WalletBalance, asset.UnrealizedProfit)
		}
	}

	// 构造返回结果
	result := make(map[string]interface{})
	result["totalWalletBalance"], _ = strconv.ParseFloat(account.TotalWalletBalance, 64)
	result["availableBalance"], _ = strconv.ParseFloat(account.AvailableBalance, 64)
	result["totalUnrealizedProfit"], _ = strconv.ParseFloat(account.TotalUnrealizedProfit, 64)

	// 添加调试日志，输出转换后的数据
	fmt.Printf("[DEBUG] GetAccountInfo - Converted Result - TotalWalletBalance: %.2f, AvailableBalance: %.2f, TotalUnrealizedProfit: %.2f\n",
		result["totalWalletBalance"], result["availableBalance"], result["totalUnrealizedProfit"])

	// 按照后端服务期望的格式返回数据
	response := gin.H{
		"data": result,
	}
	c.JSON(http.StatusOK, response)
}

// GetPositions 获取持仓信息
func (s *BinanceProxyService) GetPositions(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] GetPositions - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] GetPositions - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	// 调用币安API获取持仓信息
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	positions, err := client.NewGetPositionRiskService().Do(ctx)
	cancel()

	if err != nil {
		// 即使发生错误，也要返回标准的数据格式
		errorResponse := gin.H{
			"data":  nil, // 或者可以返回空数组 []
			"error": fmt.Sprintf("获取持仓信息失败: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// 添加调试日志，输出从币安API获取的原始持仓数据
	fmt.Printf("[DEBUG] GetPositions - Raw API Response - Positions Count: %d\n", len(positions))
	for i, pos := range positions {
		if i < 5 { // 只打印前5个持仓，避免日志过多
			fmt.Printf("[DEBUG] GetPositions - Position[%d]: Symbol: %s, PositionAmt: %s, EntryPrice: %s, MarkPrice: %s, UnrealizedProfit: %s\n",
				i, pos.Symbol, pos.PositionAmt, pos.EntryPrice, pos.MarkPrice, pos.UnRealizedProfit)
		}
	}

	var result []map[string]interface{}
	for _, pos := range positions {
		posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if posAmt == 0 {
			continue // 跳过零持仓
		}

		posMap := make(map[string]interface{})
		posMap["symbol"] = pos.Symbol
		posMap["positionAmt"], _ = strconv.ParseFloat(pos.PositionAmt, 64)
		posMap["entryPrice"], _ = strconv.ParseFloat(pos.EntryPrice, 64)
		posMap["markPrice"], _ = strconv.ParseFloat(pos.MarkPrice, 64)
		posMap["unRealizedProfit"], _ = strconv.ParseFloat(pos.UnRealizedProfit, 64)
		posMap["leverage"], _ = strconv.ParseFloat(pos.Leverage, 64)
		posMap["liquidationPrice"], _ = strconv.ParseFloat(pos.LiquidationPrice, 64)

		// 确定方向
		if posAmt > 0 {
			posMap["side"] = "long"
		} else {
			posMap["side"] = "short"
		}

		result = append(result, posMap)
	}

	// 添加调试日志，输出处理后的持仓数据
	fmt.Printf("[DEBUG] GetPositions - Processed Result - Positions Count: %d\n", len(result))
	for i, pos := range result {
		if i < 5 { // 只打印前5个处理后的持仓
			fmt.Printf("[DEBUG] GetPositions - Processed Position[%d]: Symbol: %s, PositionAmt: %.4f, EntryPrice: %.4f, MarkPrice: %.4f, UnrealizedProfit: %.4f\n",
				i, pos["symbol"], pos["positionAmt"], pos["entryPrice"], pos["markPrice"], pos["unRealizedProfit"])
		}
	}

	// 按照后端服务期望的格式返回数据
	response := gin.H{
		"data": result,
	}
	c.JSON(http.StatusOK, response)
}

// GetKlines 获取K线数据
func (s *BinanceProxyService) GetKlines(c *gin.Context) {
	// 注意：由于币安API的K线服务可能与库版本不匹配，暂时返回错误信息
	// 实际部署时可根据需要使用CoinAnk API或其他数据源
	c.JSON(http.StatusNotImplemented, gin.H{"error": "K线数据获取功能暂未实现"})
}

// PlaceOrder 下单
func (s *BinanceProxyService) PlaceOrder(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] PlaceOrder - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] PlaceOrder - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	var req struct {
		Symbol      string   `json:"symbol"`
		Side        string   `json:"side"`
		OrderType   string   `json:"order_type"`
		Quantity    float64  `json:"quantity"`
		Price       *float64 `json:"price,omitempty"`
		StopPrice   *float64 `json:"stop_price,omitempty"`
		TimeInForce string   `json:"time_in_force"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求数据格式错误: %v", err)})
		return
	}

	orderService := client.NewCreateOrderService().
		Symbol(req.Symbol).
		Side(futures.SideType(req.Side)).
		Type(futures.OrderType(req.OrderType)).
		Quantity(fmt.Sprintf("%.8f", req.Quantity))

	if req.Price != nil {
		orderService.Price(fmt.Sprintf("%.8f", *req.Price))
	}

	if req.StopPrice != nil {
		orderService.StopPrice(fmt.Sprintf("%.8f", *req.StopPrice))
	}

	if req.TimeInForce != "" {
		orderService.TimeInForce(futures.TimeInForceType(req.TimeInForce))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	order, err := orderService.Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("下单失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder 撤单
func (s *BinanceProxyService) CancelOrder(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] CancelOrder - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] CancelOrder - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	var req struct {
		Symbol            string  `json:"symbol"`
		OrderID           *int64  `json:"order_id,omitempty"`
		OrigClientOrderID *string `json:"orig_client_order_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("请求数据格式错误: %v", err)})
		return
	}

	orderService := client.NewCancelOrderService().
		Symbol(req.Symbol)

	if req.OrderID != nil {
		orderService.OrderID(*req.OrderID)
	}

	if req.OrigClientOrderID != nil {
		orderService.OrigClientOrderID(*req.OrigClientOrderID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	order, err := orderService.Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("撤单失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOrders 获取订单
func (s *BinanceProxyService) GetOrders(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] GetOrders - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] GetOrders - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol参数是必需的"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	orders, err := client.NewListOrdersService().
		Symbol(symbol).
		Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取订单失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetAccountInfo 获取账户信息
func (s *BinanceProxyService) GetAccountInfo(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] GetAccountInfo - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] GetAccountInfo - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	account, err := client.NewGetAccountService().Do(ctx)
	cancel()

	if err != nil {
		// 即使发生错误，也要返回标准的数据格式
		errorResponse := gin.H{
			"data":  nil,
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		}
		c.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	// 按照后端服务期望的格式返回数据
	response := gin.H{
		"data": account,
	}
	c.JSON(http.StatusOK, response)
}

// GetTrades 获取交易历史
func (s *BinanceProxyService) GetTrades(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	fmt.Printf("[DEBUG] GetTrades - Received API Key: %s, Secret Key: %s, Custom API URL: %s\n",
		maskString(apiKey), maskString(secretKey), customAPIURL)

	if apiKey == "" || secretKey == "" {
		fmt.Printf("[DEBUG] GetTrades - Missing API Key or Secret Key\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol参数是必需的"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	trades, err := client.NewListAccountTradeService().
		Symbol(symbol).
		Limit(100). // 默认限制100条
		Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取交易历史失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// createClient 创建币安客户端
func (s *BinanceProxyService) createClient(apiKey, secretKey, customAPIURL string) *futures.Client {
	var client *futures.Client
	if customAPIURL != "" && customAPIURL != "undefined" && customAPIURL != "null" {
		// 使用自定义端点
		client = futures.NewClient(apiKey, secretKey)
		client.BaseURL = strings.TrimSuffix(customAPIURL, "/")
	} else {
		// 使用默认端点
		client = futures.NewClient(apiKey, secretKey)
	}

	// 设置较长的超时时间
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client.HTTPClient = httpClient

	// 同步币安服务器时间以避免时间戳错误
	s.syncBinanceServerTime(client)

	return client
}

// syncBinanceServerTime 同步币安服务器时间以确保请求时间戳有效
func (s *BinanceProxyService) syncBinanceServerTime(client *futures.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	serverTime, err := client.NewServerTimeService().Do(ctx)
	if err != nil {
		fmt.Printf("⚠️ Failed to sync Binance server time: %v\n", err)
		return
	}

	now := time.Now().UnixMilli()
	offset := now - serverTime
	client.TimeOffset = offset
	fmt.Printf("⏱ Binance server time synced, offset %dms\n", offset)
}
