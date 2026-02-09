package controllers

import (
	"net/http"
	"nofx/admin/auth"
	"nofx/admin/models"

	"github.com/gin-gonic/gin"
)

// AdminController 管理员控制器
type AdminController struct {
	authService *auth.AuthService
}

// NewAdminController 创建新的管理员控制器
func NewAdminController(authService *auth.AuthService) *AdminController {
	return &AdminController{
		authService: authService,
	}
}

// GetDashboard 获取管理员仪表板数据
func (ctrl *AdminController) GetDashboard(c *gin.Context) {
	// 获取当前用户信息
	_, username, role, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取系统统计信息
	db := models.GetDB()
	
	var userCount int64
	db.DB.Model(&models.AdminUser{}).Count(&userCount)
	
	var auditCount int64
	db.DB.Model(&models.AuditLog{}).Count(&auditCount)

	// 返回仪表板数据
	c.JSON(http.StatusOK, gin.H{
		"message": "管理员仪表板",
		"data": gin.H{
			"user": gin.H{
				"username": username,
				"role":     string(role),
			},
			"statistics": gin.H{
				"total_users":      userCount,
				"total_audits":     auditCount,
				"active_sessions":  0, // 需要实现会话管理
				"system_health":    "healthy", // 需要实现健康检查
			},
		},
	})
}

// GetProfile 获取当前用户资料
func (ctrl *AdminController) GetProfile(c *gin.Context) {
	userID, username, role, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 从数据库获取用户详细信息
	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Preload("Permissions").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"role":        string(user.Role),
			"status":      string(user.Status),
			"last_login":  user.LastLoginAt,
			"created_at":  user.CreatedAt,
			"permissions": user.Permissions,
		},
	})
}

// UpdateProfile 更新当前用户资料
func (ctrl *AdminController) UpdateProfile(c *gin.Context) {
	userID, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Email    string `json:"email" binding:"omitempty,email"`
		Password string `json:"password" binding:"omitempty,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := models.GetDB()
	updateData := make(map[string]interface{})
	
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	
	if req.Password != "" {
		hashedPassword, err := ctrl.authService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		updateData["password"] = hashedPassword
	}

	if len(updateData) > 0 {
		if err := db.DB.Model(&models.AdminUser{}).Where("id = ?", userID).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户信息更新成功",
	})
}

// ChangePassword 更改密码
func (ctrl *AdminController) ChangePassword(c *gin.Context) {
	userID, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证旧密码
	if !ctrl.authService.VerifyPassword(req.OldPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "旧密码不正确"})
		return
	}

	// 更新密码
	hashedPassword, err := ctrl.authService.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	if err := db.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	// 记录审计日志
	auditLog := &models.AuditLog{
		Action:     models.UserUpdatedAudit,
		UserID:     &userID,
		Username:   &user.Username,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Details:    "Password changed",
		StatusCode: 200,
	}
	db.DB.Create(auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "密码更改成功",
	})
}

// GetSystemStatus 获取系统状态
func (ctrl *AdminController) GetSystemStatus(c *gin.Context) {
	// 获取基本系统信息
	db := models.GetDB()
	
	var userCount int64
	db.DB.Model(&models.AdminUser{}).Count(&userCount)
	
	var auditCount int64
	db.DB.Model(&models.AuditLog{}).Count(&auditCount)
	
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"uptime": "0 days 0 hours 0 minutes", // 实际应用中需要跟踪服务器启动时间
		"total_users": userCount,
		"total_audits": auditCount,
		"version": "1.0.0",
		"environment": "production",
	})
}

// GetSystemMetrics 获取系统性能指标
func (ctrl *AdminController) GetSystemMetrics(c *gin.Context) {
	// 获取系统性能指标
	c.JSON(http.StatusOK, gin.H{
		"system_status": gin.H{
			"uptime": "0天 0小时 0分钟",
			"cpu_usage": 0.0, // 实际应用中需要从系统获取
			"memory_usage": 0.0, // 实际应用中需要从系统获取
			"disk_usage": 0.0, // 实际应用中需要从系统获取
			"network_io": gin.H{"rx": 0.0, "tx": 0.0}, // 实际应用中需要从系统获取
			"db_connections": 0, // 实际应用中需要从数据库连接池获取
			"active_users": 0, // 实际应用中需要从会话管理获取
		},
		"performance_metrics": gin.H{
			"api_response_time": 0.0,
			"db_query_time": 0.0,
			"error_rate": 0.0,
		},
	})
}