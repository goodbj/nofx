package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"nofx/admin/config"
	"nofx/admin/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	config *config.Config
}

// NewAuthService 创建新的认证服务
func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Remember bool   `json:"remember"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token     string      `json:"token"`
	User      UserInfo    `json:"user"`
	ExpiresAt int64       `json:"expires_at"`
	Message   string      `json:"message"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Login 管理员登录
func (s *AuthService) Login(c *gin.Context, req LoginRequest) (*LoginResponse, error) {
	// 从数据库查询用户
	db := models.GetDB()
	var user models.AdminUser
	
	if err := db.DB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		// 记录登录失败日志
		s.logLoginAttempt(c, &user, req.Username, models.FailedLogin, "用户不存在")
		return nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status == models.LockedStatus {
		s.logLoginAttempt(c, &user, req.Username, models.BlockedLogin, "账户已被锁定")
		return nil, errors.New("账户已被锁定")
	}

	if user.Status == models.SuspendedStatus {
		s.logLoginAttempt(c, &user, req.Username, models.BlockedLogin, "账户已被暂停")
		return nil, errors.New("账户已被暂停")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// 记录登录失败日志
		s.logLoginAttempt(c, &user, req.Username, models.FailedLogin, "密码错误")
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT令牌
	token, err := s.generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成令牌失败: %v", err)
	}

	// 更新最后登录时间和IP
	now := time.Now()
	ip := c.ClientIP()
	db.DB.Model(&user).Updates(map[string]interface{}{
		"last_login_at": &now,
		"last_login_ip": ip,
	})

	// 记录登录成功日志
	s.logLoginAttempt(c, &user, req.Username, models.SuccessLogin, "")

	// 返回登录成功响应
	response := &LoginResponse{
		Token:   token,
		ExpiresAt: time.Now().Add(time.Duration(s.config.JWTExpiration) * time.Hour).Unix(),
		Message: "登录成功",
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     string(user.Role),
			Status:   string(user.Status),
		},
	}

	return response, nil
}

// generateToken 生成JWT令牌
func (s *AuthService) generateToken(user *models.AdminUser) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.config.JWTExpiration) * time.Hour)
	
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nofx-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken 验证JWT令牌
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	return claims, nil
}

// HashPassword 加密密码
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证密码
func (s *AuthService) VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// RequireAuth 中间件：要求认证
func (s *AuthService) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少授权头部",
			})
			c.Abort()
			return
		}

		// 检查授权头部格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的授权头部格式",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证令牌
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌",
			})
			c.Abort()
			return
		}

		// 将用户信息添加到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 继续处理请求
		c.Next()
	}
}

// RequireRole 中间件：要求特定角色
func (s *AuthService) RequireRole(requiredRole models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少授权头部",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的授权头部格式",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌",
			})
			c.Abort()
			return
		}

		// 检查角色权限
		if models.Role(claims.Role) != requiredRole && models.Role(claims.Role) != models.SuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequirePermission 中间件：要求特定权限
func (s *AuthService) RequirePermission(requiredPermission models.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少授权头部",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的授权头部格式",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌",
			})
			c.Abort()
			return
		}

		// 从数据库获取用户详细信息以检查权限
		db := models.GetDB()
		var user models.AdminUser
		
		if err := db.DB.Preload("Permissions").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission := user.HasPermission(requiredPermission)
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("permissions", user.Permissions)
		c.Next()
	}
}

// logLoginAttempt 记录登录尝试
func (s *AuthService) logLoginAttempt(c *gin.Context, user *models.AdminUser, username string, status models.LoginStatus, reason string) {
	loginLog := &models.LoginLog{
		UserID:    user.ID,
		Username:  &username,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Status:    status,
		Reason:    reason,
	}

	db := models.GetDB()
	if err := db.DB.Create(loginLog).Error; err != nil {
		log.Printf("记录登录日志失败: %v", err)
	}

	// 同时记录审计日志
	action := models.UserLoginAudit
	if status == models.FailedLogin {
		action = "user.login.failed"
	} else if status == models.BlockedLogin {
		action = "user.login.blocked"
	}

	auditLog := &models.AuditLog{
		Action:     action,
		UserID:     &user.ID,
		Username:   &username,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Details:    fmt.Sprintf("Login attempt - Status: %s, Reason: %s", status, reason),
		StatusCode: 200,
	}

	if err := db.DB.Create(auditLog).Error; err != nil {
		log.Printf("记录审计日志失败: %v", err)
	}
}

// GetUserFromContext 从上下文获取用户信息
func (s *AuthService) GetUserFromContext(c *gin.Context) (string, string, models.Role, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", "", "", errors.New("用户未认证")
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")

	userIDStr, ok := userID.(string)
	if !ok {
		return "", "", "", errors.New("无效的用户ID类型")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", "", "", errors.New("无效的用户名类型")
	}

	roleStr, ok := role.(string)
	if !ok {
		return "", "", "", errors.New("无效的角色类型")
	}

	return userIDStr, usernameStr, models.Role(roleStr), nil
}