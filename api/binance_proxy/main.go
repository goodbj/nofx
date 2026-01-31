package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config := LoadConfig()

	// 设置Gin模式
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "X-API-Key, X-Secret-Key, X-Custom-API-URL, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 初始化服务
	service := NewBinanceProxyService()

	// 注册路由
	setupRoutes(r, service)

	log.Printf("币安代理服务启动中，监听端口: %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatalf("启动币安代理服务失败: %v", err)
	}
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, service *BinanceProxyService) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "binance-proxy"})
	})

	// 代理API组
	proxy := r.Group("/api/proxy")
	{
		// 获取账户余额
		proxy.GET("/balance", service.GetBalance)

		// 获取持仓信息
		proxy.GET("/positions", service.GetPositions)

		// 获取K线数据
		proxy.GET("/klines", service.GetKlines)

		// 订单管理
		proxy.POST("/orders", service.PlaceOrder)
		proxy.DELETE("/orders", service.CancelOrder)
		proxy.GET("/orders", service.GetOrders)

		// 获取账户信息
		proxy.GET("/account", service.GetAccountInfo)

		// 获取交易历史
		proxy.GET("/trades", service.GetTrades)
	}
}
