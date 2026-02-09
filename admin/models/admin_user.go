package models

import (
	"time"
	
	"gorm.io/gorm"
)

// Role 管理员角色枚举
type Role string

const (
	SuperAdmin Role = "super_admin"  // 超级管理员 - 拥有所有权限
	SystemAdmin Role = "system_admin" // 系统管理员 - 管理系统配置
	UserAdmin Role = "user_admin"    // 用户管理员 - 管理用户账户
	MonitorAdmin Role = "monitor_admin" // 监控管理员 - 查看系统监控
	ViewOnly Role = "view_only"      // 只读用户 - 仅查看权限
)

// Permission 权限定义
type Permission string

const (
	// 用户管理权限
	UserReadPermission    Permission = "user:read"    // 读取用户信息
	UserWritePermission   Permission = "user:write"   // 创建/修改用户
	UserDeletePermission  Permission = "user:delete"  // 删除用户
	UserManagePermission  Permission = "user:manage"  // 管理用户权限
	
	// 系统配置权限
	SystemReadPermission  Permission = "system:read"  // 读取系统配置
	SystemWritePermission Permission = "system:write" // 修改系统配置
	SystemManagePermission Permission = "system:manage" // 管理系统设置
	
	// 监控权限
	MonitorReadPermission Permission = "monitor:read" // 查看监控数据
	MonitorWritePermission Permission = "monitor:write" // 修改监控设置
	
	// 审计权限
	AuditReadPermission   Permission = "audit:read"   // 查看审计日志
	AuditWritePermission  Permission = "audit:write"  // 写入审计日志
	AuditManagePermission Permission = "audit:manage" // 管理审计设置
	
	// 所有权限
	AllPermissions Permission = "*" // 拥有所有权限
)

// AdminUser 管理员用户模型
type AdminUser struct {
	ID          string         `gorm:"primaryKey;type:char(36)" json:"id"`
	Username    string         `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email       string         `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password    string         `gorm:"not null;size:255" json:"password"`
	Role        Role           `gorm:"default:user_admin" json:"role"`
	Status      UserStatus     `gorm:"default:active" json:"status"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	LastLoginIP string         `json:"last_login_ip"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联字段
	Permissions []PermissionAssignment `gorm:"foreignKey:UserID" json:"permissions"`
	LoginLogs   []LoginLog            `gorm:"foreignKey:UserID" json:"login_logs"`
}

// UserStatus 用户状态
type UserStatus string

const (
	ActiveStatus   UserStatus = "active"   // 活跃
	InactiveStatus UserStatus = "inactive" // 非活跃
	LockedStatus   UserStatus = "locked"   // 锁定
	SuspendedStatus UserStatus = "suspended" // 暂停
)

// PermissionAssignment 权限分配模型
type PermissionAssignment struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      string       `gorm:"not null;index" json:"user_id"`
	Permission  Permission   `gorm:"not null;size:50" json:"permission"`
	GrantedBy   string       `gorm:"size:36" json:"granted_by"`
	GrantedAt   time.Time    `json:"granted_at"`
	RevokedAt   *time.Time   `json:"revoked_at"`
	RevokedBy   string       `gorm:"size:36" json:"revoked_by"`
	Reason      string       `gorm:"size:255" json:"reason"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// LoginLog 登录日志模型
type LoginLog struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      string       `gorm:"not null;index" json:"user_id"`
	Username    string       `gorm:"size:50" json:"username"`
	IP          string       `gorm:"size:45" json:"ip"` // IPv6 地址最长45字符
	UserAgent   string       `gorm:"size:500" json:"user_agent"`
	Status      LoginStatus  `gorm:"default:success" json:"status"`
	Reason      string       `gorm:"size:255" json:"reason"` // 失败原因
	CreatedAt   time.Time    `json:"created_at"`
}

// LoginStatus 登录状态
type LoginStatus string

const (
	SuccessLogin LoginStatus = "success"
	FailedLogin  LoginStatus = "failed"
	BlockedLogin LoginStatus = "blocked"
)

// TableName 指定表名
func (AdminUser) TableName() string {
	return "admin_users"
}

func (PermissionAssignment) TableName() string {
	return "admin_permissions"
}

func (LoginLog) TableName() string {
	return "admin_login_logs"
}

// HasPermission 检查用户是否有指定权限
func (u *AdminUser) HasPermission(permission Permission) bool {
	// 超级管理员拥有所有权限
	if u.Role == SuperAdmin {
		return true
	}
	
	// 检查是否拥有特定权限
	for _, perm := range u.Permissions {
		if perm.Permission == permission || perm.Permission == AllPermissions {
			if perm.RevokedAt == nil { // 权限未被撤销
				return true
			}
		}
	}
	
	return false
}

// HasRole 检查用户是否具有指定角色
func (u *AdminUser) HasRole(role Role) bool {
	return u.Role == role
}

// IsSuperAdmin 检查是否为超级管理员
func (u *AdminUser) IsSuperAdmin() bool {
	return u.Role == SuperAdmin
}