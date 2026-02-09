package controllers

import (
	"fmt"
	"net/http"
	"nofx/admin/auth"
	"nofx/admin/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserController 用户管理控制器
type UserController struct {
	authService *auth.AuthService
}

// NewUserController 创建新的用户管理控制器
func NewUserController(authService *auth.AuthService) *UserController {
	return &UserController{
		authService: authService,
	}
}

// GetUsers 获取用户列表
func (ctrl *UserController) GetUsers(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查权限
	if !c.MustGet("permissions").([]models.PermissionAssignment) {
		// 这里需要修复权限检查逻辑
		// 简化处理：检查用户是否有用户管理权限
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")
	role := c.Query("role")

	offset := (page - 1) * limit

	db := models.GetDB()
	var users []models.AdminUser
	query := db.DB.Model(&models.AdminUser{})

	// 应用过滤条件
	if search != "" {
		search = "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", search, search)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 获取用户列表
	query.Offset(offset).Limit(limit).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetUser 获取单个用户
func (ctrl *UserController) GetUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID := c.Param("id")

	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Preload("Permissions").Preload("LoginLogs").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// CreateUser 创建用户
func (ctrl *UserController) CreateUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required"`
		Status   string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色
	validRoles := map[string]bool{
		"super_admin":    true,
		"system_admin":   true,
		"user_admin":     true,
		"monitor_admin":  true,
		"view_only":      true,
	}

	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}

	// 验证状态
	validStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"locked":    true,
		"suspended": true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态"})
		return
	}

	db := models.GetDB()

	// 检查用户名和邮箱是否已存在
	var existingUser models.AdminUser
	if db.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或邮箱已存在"})
		return
	}

	// 生成用户ID
	userID := uuid.New().String()

	// 加密密码
	hashedPassword, err := ctrl.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// 创建用户
	user := &models.AdminUser{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.Role(req.Role),
		Status:   models.UserStatus(req.Status),
	}

	if err := db.DB.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 记录审计日志
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)
	auditDetails := map[string]interface{}{
		"action": "user_created",
		"target_user_id": user.ID,
		"target_username": user.Username,
		"creator_id": currentUserID,
		"creator_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.UserCreatedAudit,
		"admin_user",
		&currentUserID,
		&currentUsername,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		auditDetails,
		201,
	)
	if err != nil {
		// 记录审计日志失败不应该影响主要操作
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user":    user,
	})
}

// UpdateUser 更新用户
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID := c.Param("id")

	var req struct {
		Username *string `json:"username" binding:"omitempty,min=3,max=50"`
		Email    *string `json:"email" binding:"omitempty,email"`
		Password *string `json:"password" binding:"omitempty,min=6"`
		Role     *string `json:"role"`
		Status   *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	
	if req.Username != nil {
		updateData["username"] = *req.Username
	}
	
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	
	if req.Password != nil {
		hashedPassword, err := ctrl.authService.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		updateData["password"] = hashedPassword
	}
	
	if req.Role != nil {
		// 验证角色
		validRoles := map[string]bool{
			"super_admin":    true,
			"system_admin":   true,
			"user_admin":     true,
			"monitor_admin":  true,
			"view_only":      true,
		}
		
		if !validRoles[*req.Role] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
			return
		}
		updateData["role"] = *req.Role
	}
	
	if req.Status != nil {
		// 验证状态
		validStatuses := map[string]bool{
			"active":    true,
			"inactive":  true,
			"locked":    true,
			"suspended": true,
		}
		
		if !validStatuses[*req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态"})
			return
		}
		updateData["status"] = *req.Status
	}

	// 执行更新
	if err := db.DB.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	// 记录审计日志
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)
	auditDetails := map[string]interface{}{
		"action": "user_updated",
		"target_user_id": user.ID,
		"target_username": user.Username,
		"updater_id": currentUserID,
		"updater_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.UserUpdatedAudit,
		"admin_user",
		&currentUserID,
		&currentUsername,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		auditDetails,
		200,
	)
	if err != nil {
		// 记录审计日志失败不应该影响主要操作
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"user":    user,
	})
}

// DeleteUser 删除用户
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID := c.Param("id")

	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 防止删除自己
	currentUserID, _, _, _ := ctrl.authService.GetUserFromContext(c)
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能删除自己"})
		return
	}

	// 执行软删除
	if err := db.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	// 记录审计日志
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)
	auditDetails := map[string]interface{}{
		"action": "user_deleted",
		"target_user_id": user.ID,
		"target_username": user.Username,
		"deleter_id": currentUserID,
		"deleter_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.UserDeletedAudit,
		"admin_user",
		&currentUserID,
		&currentUsername,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		auditDetails,
		200,
	)
	if err != nil {
		// 记录审计日志失败不应该影响主要操作
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户删除成功",
	})
}

// LockUser 锁定用户
func (ctrl *UserController) LockUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID := c.Param("id")

	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新用户状态为锁定
	if err := db.DB.Model(&user).Update("status", models.LockedStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "锁定用户失败"})
		return
	}

	// 记录审计日志
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)
	auditDetails := map[string]interface{}{
		"action": "user_locked",
		"target_user_id": user.ID,
		"target_username": user.Username,
		"locker_id": currentUserID,
		"locker_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.UserLockedAudit,
		"admin_user",
		&currentUserID,
		&currentUsername,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		auditDetails,
		200,
	)
	if err != nil {
		// 记录审计日志失败不应该影响主要操作
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户锁定成功",
		"user":    user,
	})
}

// UnlockUser 解锁用户
func (ctrl *UserController) UnlockUser(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	userID := c.Param("id")

	db := models.GetDB()
	var user models.AdminUser
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新用户状态为活跃
	if err := db.DB.Model(&user).Update("status", models.ActiveStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解锁用户失败"})
		return
	}

	// 记录审计日志
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)
	auditDetails := map[string]interface{}{
		"action": "user_unlocked",
		"target_user_id": user.ID,
		"target_username": user.Username,
		"unlocker_id": currentUserID,
		"unlocker_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.UserUnlockedAudit,
		"admin_user",
		&currentUserID,
		&currentUsername,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		auditDetails,
		200,
	)
	if err != nil {
		// 记录审计日志失败不应该影响主要操作
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户解锁成功",
		"user":    user,
	})
}