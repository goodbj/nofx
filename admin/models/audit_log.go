package models

import (
	"fmt"
	"time"
	
	"gorm.io/gorm"
)

// AuditAction 审计动作类型
type AuditAction string

const (
	// 用户管理动作
	UserCreatedAudit    AuditAction = "user.created"
	UserUpdatedAudit    AuditAction = "user.updated"
	UserDeletedAudit    AuditAction = "user.deleted"
	UserLoginAudit      AuditAction = "user.login"
	UserLogoutAudit     AuditAction = "user.logout"
	UserLockedAudit     AuditAction = "user.locked"
	UserUnlockedAudit   AuditAction = "user.unlocked"
	
	// 权限管理动作
	PermissionGrantedAudit AuditAction = "permission.granted"
	PermissionRevokedAudit AuditAction = "permission.revoked"
	RoleAssignedAudit      AuditAction = "role.assigned"
	RoleRemovedAudit       AuditAction = "role.removed"
	
	// 系统配置动作
	SystemConfigUpdatedAudit AuditAction = "system.config.updated"
	SystemRestartAudit       AuditAction = "system.restart"
	SystemShutdownAudit      AuditAction = "system.shutdown"
	
	// 数据管理动作
	DataExportedAudit AuditAction = "data.exported"
	DataImportedAudit AuditAction = "data.imported"
	DataDeletedAudit  AuditAction = "data.deleted"
	
	// 安全相关动作
	SecurityAlertAudit AuditAction = "security.alert"
	IPBlockedAudit     AuditAction = "ip.blocked"
	IPUnblockedAudit   AuditAction = "ip.unblocked"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID          uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	Action      AuditAction            `gorm:"not null;size:50;index" json:"action"`
	Resource    string                 `gorm:"size:100;index" json:"resource"`      // 操作的资源
	UserID      *string                `gorm:"size:36;index" json:"user_id"`        // 执行操作的用户ID
	Username    *string                `gorm:"size:50;index" json:"username"`      // 执行操作的用户名
	IPAddress   string                 `gorm:"size:45;index" json:"ip_address"`    // 操作发起的IP地址
	UserAgent   string                 `gorm:"size:500" json:"user_agent"`         // 操作发起的用户代理
	Details     map[string]interface{} `gorm:"type:json" json:"details"`            // 操作详情(JSON格式)
	StatusCode  int                    `json:"status_code"`                          // 操作状态码
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	
	// 关联字段
	User *AdminUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "admin_audit_logs"
}

// AuditFilter 审计日志过滤条件
type AuditFilter struct {
	Page         int           `json:"page"`
	Limit        int           `json:"limit"`
	StartDate    *time.Time    `json:"start_date"`
	EndDate      *time.Time    `json:"end_date"`
	UserID       *string       `json:"user_id"`
	Username     *string       `json:"username"`
	Action       *AuditAction  `json:"action"`
	ResourceType *string       `json:"resource_type"`
	IPAddress    *string       `json:"ip_address"`
	OrderBy      string        `json:"order_by"` // created_at, action, user_id 等
	OrderDir     string        `json:"order_dir"` // asc, desc
}

// CreateAuditLog 创建审计日志
func CreateAuditLog(action AuditAction, resource string, userID *string, username *string, ipAddress, userAgent string, details map[string]interface{}, statusCode int) error {
	db := GetDB()
	log := &AuditLog{
		Action:     action,
		Resource:   resource,
		UserID:     userID,
		Username:   username,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    details,
		StatusCode: statusCode,
	}
	return db.DB.Create(log).Error
}

// GetRecentAuditLogs 获取最近的审计日志
func GetRecentAuditLogs(limit int) ([]AuditLog, error) {
	db := GetDB()
	var logs []AuditLog
	err := db.DB.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

// QueryAuditLogs 查询审计日志
func QueryAuditLogs(filter AuditFilter) ([]AuditLog, int64, error) {
	db := GetDB()
	query := db.DB.Model(&AuditLog{})

	// 应用过滤条件
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Username != nil {
		query = query.Where("username = ?", *filter.Username)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.IPAddress != nil {
		query = query.Where("ip_address = ?", *filter.IPAddress)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 设置排序
	orderBy := "created_at"
	orderDir := "desc"	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))

	// 分页
	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	// 执行查询
	var logs []AuditLog
	err := query.Find(&logs).Error
	return logs, total, err
}