package controllers

import (
	"net/http"
	"strconv"
	"strings"
	
	"nofx/admin/auth"
	"nofx/admin/models"

	"github.com/gin-gonic/gin"
)

// AuditController 审计日志控制器
type AuditController struct {
	authService *auth.AuthService
}

// NewAuditController 创建新的审计日志控制器
func NewAuditController(authService *auth.AuthService) *AuditController {
	return &AuditController{
		authService: authService,
	}
}

// GetAuditLogs 获取审计日志列表
func (ctrl *AuditController) GetAuditLogs(c *gin.Context) {
	// 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	userID := c.Query("user_id")
	username := c.Query("username")
	action := c.Query("action")
	ipAddress := c.Query("ip_address")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 构建查询条件
	db := models.GetDB()
	query := db.DB.Model(&models.AuditLog{})

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate+" 00:00:00")
	}

	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if ipAddress != "" {
		query = query.Where("ip_address = ?", ipAddress)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取日志列表
	var logs []models.AuditLog
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审计日志失败"})
		return
	}

	// 计算总页数
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"pagination": gin.H{
			"current": page,
			"pageSize": limit,
			"total":   total,
			"totalPages": totalPages,
		},
	})
}

// GetAuditLogByID 根据ID获取审计日志详情
func (ctrl *AuditController) GetAuditLogByID(c *gin.Context) {
	id := c.Param("id")

	db := models.GetDB()
	var log models.AuditLog
	err := db.DB.Preload("User").Where("id = ?", id).First(&log).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "审计日志不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"log":     log,
	})
}

// CreateAuditLog 手动创建审计日志（通常用于内部调用）
func (ctrl *AuditController) CreateAuditLog(c *gin.Context) {
	var req struct {
		Action    string                 `json:"action" binding:"required"`
		Resource  string                 `json:"resource" binding:"required"`
		IPAddress string                 `json:"ip_address"`
		UserAgent string                 `json:"user_agent"`
		Details   map[string]interface{} `json:"details"`
		Username  string                 `json:"username"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户信息
	currentUserID, currentUsername, _, err := ctrl.authService.GetUserFromContext(c)
	if err == nil {
		// 如果能获取到当前用户，则优先使用当前用户信息
		req.Username = currentUsername
	}

	db := models.GetDB()
	auditLog := &models.AuditLog{
		Action:    models.AuditAction(req.Action),
		Resource:  req.Resource,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Details:   req.Details,
		Username:  req.Username,
		UserID:    &currentUserID,
	}

	err = db.DB.Create(auditLog).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建审计日志失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"log":     auditLog,
		"message": "审计日志创建成功",
	})
}