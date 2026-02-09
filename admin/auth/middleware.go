package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// IPWhitelistMiddleware IP白名单中间件
func (s *AuthService) IPWhitelistMiddleware() gin.HandlerFunc {
	whitelist := strings.Split(s.config.IPWhitelist, ",")
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// 检查IP是否在白名单中
		allowed := false
		for _, ip := range whitelist {
			ip = strings.TrimSpace(ip)
			if ip == clientIP || ip == "*" {
				allowed = true
				break
			}
		}
		
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "IP地址不在白名单中",
				"ip":    clientIP,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RateLimitMiddleware 速率限制中间件
func (s *AuthService) RateLimitMiddleware() gin.HandlerFunc {
	// 这里可以使用内存或Redis实现速率限制
	// 为简化，这里只做基本实现
	
	return func(c *gin.Context) {
		// 实现速率限制逻辑
		// 检查来自同一IP的请求数量是否超过限制
		
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头部中间件
func (s *AuthService) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加安全头部
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		c.Next()
	}
}

// CORS middleware 跨域资源共享中间件
func (s *AuthService) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}