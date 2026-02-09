package controllers

import (
	"net/http"
	"strconv"

	"nofx/admin/auth"
	"nofx/admin/models"

	"github.com/gin-gonic/gin"
)

// SystemConfigController 系统配置控制器
type SystemConfigController struct {
	authService *auth.AuthService
}

// NewSystemConfigController 创建新的系统配置控制器
func NewSystemConfigController(authService *auth.AuthService) *SystemConfigController {
	return &SystemConfigController{
		authService: authService,
	}
}

// GetSystemConfig 获取系统配置
func (ctrl *SystemConfigController) GetSystemConfig(c *gin.Context) {
	configs, err := models.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  configs,
	})
}

// GetConfigByCategory 根据分类获取配置
func (ctrl *SystemConfigController) GetConfigByCategory(c *gin.Context) {
	category := c.Param("category")

	configs, err := models.GetConfigsByCategory(models.ConfigCategory(category))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取配置分类失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": category,
		"config":   configs,
	})
}

// UpdateSystemConfig 更新系统配置
func (ctrl *SystemConfigController) UpdateSystemConfig(c *gin.Context) {
	var req struct {
		Config []models.SystemConfig `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	currentUserID, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 更新配置
	for i := range req.Config {
		req.Config[i].UpdatedBy = currentUserID
	}

	if err := models.BulkUpdateConfigs(req.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "系统配置更新成功",
	})
}

// GetConfigByKey 根据键获取特定配置
func (ctrl *SystemConfigController) GetConfigByKey(c *gin.Context) {
	key := c.Param("key")

	config, err := models.GetConfigByKey(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
	})
}

// CreateOrUpdateConfig 创建或更新配置项
func (ctrl *SystemConfigController) CreateOrUpdateConfig(c *gin.Context) {
	var config models.SystemConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	currentUserID, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	config.UpdatedBy = currentUserID

	db := models.GetDB()
	// 使用 FirstOrCreate 方法，如果存在则更新，否则创建
	if err := db.DB.Where(models.SystemConfig{Key: config.Key}).Assign(config).FirstOrCreate(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  config,
		"message": "配置保存成功",
	})
}

// DeleteConfig 删除配置
func (ctrl *SystemConfigController) DeleteConfig(c *gin.Context) {
	key := c.Param("key")

	if err := models.DeleteConfig(key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除配置失败"})
		return
	}

	// 记录审计日志
	userID, username, _, _ := ctrl.authService.GetUserFromContext(c)
	auditLog := &models.AuditLog{
		Action:     models.ConfigDeletedAudit,
		Resource:   "system_config",
		UserID:     &userID,
		Username:   &username,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Details:    map[string]interface{}{"key": key},
		StatusCode: 200,
	}
	db := models.GetDB()
	db.DB.Create(auditLog)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置删除成功",
	})
}

// GetGeneralSettings 获取常规设置
func (ctrl *SystemConfigController) GetGeneralSettings(c *gin.Context) {
	configs, err := models.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取常规设置失败"})
		return
	}

	settings := models.GetGeneralSettings(configs)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// GetSecuritySettings 获取安全设置
func (ctrl *SystemConfigController) GetSecuritySettings(c *gin.Context) {
	configs, err := models.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取安全设置失败"})
		return
	}

	settings := models.GetSecuritySettings(configs)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// GetNotificationSettings 获取通知设置
func (ctrl *SystemConfigController) GetNotificationSettings(c *gin.Context) {
	configs, err := models.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取通知设置失败"})
		return
	}

	settings := models.GetNotificationSettings(configs)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}

// GetPerformanceSettings 获取性能设置
func (ctrl *SystemConfigController) GetPerformanceSettings(c *gin.Context) {
	configs, err := models.GetAllConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取性能设置失败"})
		return
	}

	settings := models.GetPerformanceSettings(configs)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"settings": settings,
	})
}