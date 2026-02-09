package controllers

import (
	"fmt"
	"net/http"
	"nofx/admin/auth"
	"nofx/admin/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionController 权限管理控制器
type PermissionController struct {
	authService *auth.AuthService
}

// NewPermissionController 创建新的权限管理控制器
func NewPermissionController(authService *auth.AuthService) *PermissionController {
	return &PermissionController{
		authService: authService,
	}
}

// GetPermissions 获取权限列表
func (ctrl *PermissionController) GetPermissions(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	userID := c.Query("user_id")

	offset := (page - 1) * limit

	db := models.GetDB()
	var permissions []models.PermissionAssignment
	query := db.DB.Model(&models.PermissionAssignment{})

	// 应用过滤条件
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 获取权限列表
	query.Offset(offset).Limit(limit).Preload("User").Find(&permissions)

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GrantPermission 授予权限
func (ctrl *PermissionController) GrantPermission(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Permission string `json:"permission" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证权限类型
	validPermissions := map[string]bool{
		"user:read":    true,
		"user:write":   true,
		"user:delete":  true,
		"user:manage":  true,
		"system:read":  true,
		"system:write": true,
		"system:manage": true,
		"monitor:read": true,
		"monitor:write": true,
		"audit:read":   true,
		"audit:write":  true,
		"audit:manage": true,
		"*":            true,
	}

	if !validPermissions[req.Permission] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限"})
		return
	}

	db := models.GetDB()

	// 检查用户是否存在
	var user models.AdminUser
	if err := db.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查权限是否已经存在且未被撤销
	var existingPerm models.PermissionAssignment
	if db.DB.Where("user_id = ? AND permission = ? AND revoked_at IS NULL", req.UserID, req.Permission).First(&existingPerm).RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "权限已存在"})
		return
	}

	// 获取当前用户ID
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)

	// 创建权限分配
	permission := &models.PermissionAssignment{
		UserID:     req.UserID,
		Permission: models.Permission(req.Permission),
		GrantedBy:  currentUserID,
		GrantedAt:  c.Now(),
		Reason:     req.Reason,
	}

	if err := db.DB.Create(permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "授予权限失败"})
		return
	}

	// 记录审计日志
	auditDetails := map[string]interface{}{
		"action": "permission_granted",
		"target_user_id": req.UserID,
		"permission": req.Permission,
		"reason": req.Reason,
		"granted_by": currentUserID,
		"granted_by_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.PermissionGrantedAudit,
		"permission",
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
		"message":    "权限授予成功",
		"permission": permission,
	})
}

// RevokePermission 撤销权限
func (ctrl *PermissionController) RevokePermission(c *gin.Context) {
	idStr := c.Param("id")
	
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := models.GetDB()
	var permission models.PermissionAssignment
	if err := db.DB.Where("id = ?", idStr).First(&permission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "权限分配不存在"})
		return
	}

	// 检查权限是否已经被撤销
	if permission.RevokedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "权限已被撤销"})
		return
	}

	// 获取当前用户ID
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)

	// 撤销权限
	now := c.Now()
	permission.RevokedAt = &now
	permission.RevokedBy = currentUserID

	if err := db.DB.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "撤销权限失败"})
		return
	}

	// 记录审计日志
	auditDetails := map[string]interface{}{
		"action": "permission_revoked",
		"target_user_id": permission.UserID,
		"permission": string(permission.Permission),
		"revoked_by": currentUserID,
		"revoked_by_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.PermissionRevokedAudit,
		"permission",
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
		"message": "权限撤销成功",
		"permission": permission,
	})
}

// AssignRole 分配角色
func (ctrl *PermissionController) AssignRole(c *gin.Context) {
	_, _, _, err := ctrl.authService.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证角色
	validRoles := map[string]bool{
		"super_admin":   true,
		"system_admin":  true,
		"user_admin":    true,
		"monitor_admin": true,
		"view_only":     true,
	}

	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色"})
		return
	}

	db := models.GetDB()

	// 检查用户是否存在
	var user models.AdminUser
	if err := db.DB.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新用户角色
	oldRole := user.Role
	user.Role = models.Role(req.Role)

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配角色失败"})
		return
	}

	// 获取当前用户ID
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)

	// 记录审计日志
	auditDetails := map[string]interface{}{
		"action": "role_assigned",
		"target_user_id": req.UserID,
		"new_role": req.Role,
		"old_role": string(oldRole),
		"assigned_by": currentUserID,
		"assigned_by_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.RoleAssignedAudit,
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
		"message": "角色分配成功",
		"user":    user,
	})
}

// RemoveRole 移除角色（重置为默认角色）
func (ctrl *PermissionController) RemoveRole(c *gin.Context) {
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

	// 获取当前用户ID
	currentUserID, currentUsername, _, _ := ctrl.authService.GetUserFromContext(c)

	// 检查是否尝试移除自己的角色
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不能移除自己的角色"})
		return
	}

	// 保存旧角色
	oldRole := user.Role

	// 重置为默认角色
	user.Role = models.UserAdmin

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除角色失败"})
		return
	}

	// 记录审计日志
	auditDetails := map[string]interface{}{
		"action": "role_removed",
		"target_user_id": userID,
		"old_role": string(oldRole),
		"new_role": "user_admin",
		"removed_by": currentUserID,
		"removed_by_username": currentUsername,
	}
	err := models.CreateAuditLog(
		models.RoleRemovedAudit,
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
		"message": "角色移除成功",
		"user":    user,
	})
}