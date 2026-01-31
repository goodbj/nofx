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

// NewBinanceProxyService 创建新的币安代理服务实例
func NewBinanceProxyService() *BinanceProxyService {
	return &BinanceProxyService{}
}

// GetBalance 获取账户余额
func (s *BinanceProxyService) GetBalance(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	if apiKey == "" || secretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	// 调用币安API获取账户信息
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	account, err := client.NewGetAccountService().Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取账户信息失败: %v", err)})
		return
	}

	// 构造返回结果
	result := make(map[string]interface{})
	result["totalWalletBalance"], _ = strconv.ParseFloat(account.TotalWalletBalance, 64)
	result["availableBalance"], _ = strconv.ParseFloat(account.AvailableBalance, 64)
	result["totalUnrealizedProfit"], _ = strconv.ParseFloat(account.TotalUnrealizedProfit, 64)

	c.JSON(http.StatusOK, result)
}

// GetPositions 获取持仓信息
func (s *BinanceProxyService) GetPositions(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	if apiKey == "" || secretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	// 调用币安API获取持仓信息
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	positions, err := client.NewGetPositionRiskService().Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取持仓信息失败: %v", err)})
		return
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

	c.JSON(http.StatusOK, result)
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

	if apiKey == "" || secretKey == "" {
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

	if apiKey == "" || secretKey == "" {
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

	if apiKey == "" || secretKey == "" {
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

	if apiKey == "" || secretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API Key和Secret Key是必需的"})
		return
	}

	client := s.createClient(apiKey, secretKey, customAPIURL)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	account, err := client.NewGetAccountService().Do(ctx)
	cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("获取账户信息失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, account)
}

// GetTrades 获取交易历史
func (s *BinanceProxyService) GetTrades(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	secretKey := c.GetHeader("X-Secret-Key")
	customAPIURL := c.GetHeader("X-Custom-API-URL")

	if apiKey == "" || secretKey == "" {
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

	// 增加HTTP客户端超时时间以应对网络不稳定
	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	} else {
		client.HTTPClient.Timeout = 30 * time.Second
	}

	return client
}
