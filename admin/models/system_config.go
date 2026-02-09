package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	
	"gorm.io/gorm"
)

// ConfigCategory 配置类别
type ConfigCategory string

const (
	GeneralConfig    ConfigCategory = "general"    // 常规配置
	SecurityConfig   ConfigCategory = "security"   // 安全配置
	AuthenticationConfig ConfigCategory = "authentication" // 认证配置
	NotificationConfig ConfigCategory = "notification" // 通知配置
	PerformanceConfig ConfigCategory = "performance" // 性能配置
	IntegrationConfig ConfigCategory = "integration" // 集成配置
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string         `gorm:"uniqueIndex;not null;size:100" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`           // 配置值，JSON格式存储
	Category    ConfigCategory `gorm:"size:20;index" json:"category"`   // 配置类别
	Description string         `gorm:"size:500" json:"description"`     // 配置描述
	DataType    string         `gorm:"size:20;default:'string'" json:"data_type"` // 数据类型
	IsRequired  bool           `gorm:"default:false" json:"is_required"` // 是否必需
	IsEncrypted bool           `gorm:"default:false" json:"is_encrypted"` // 是否加密存储
	DefaultValue string        `gorm:"type:text" json:"default_value"`   // 默认值
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   string         `gorm:"size:36" json:"updated_by"`      // 最后更新者ID
	IsPublic    bool           `gorm:"default:false" json:"is_public"`   // 是否对普通用户可见
	
	// 验证规则
	ValidationRules string `gorm:"type:text" json:"validation_rules"` // 验证规则(JSON格式)
}

// GeneralSettings 常规设置
type GeneralSettings struct {
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	SiteURL         string `json:"site_url"`
	ContactEmail    string `json:"contact_email"`
	Copyright       string `json:"copyright"`
	MaintenanceMode bool   `json:"maintenance_mode"`
	Theme           string `json:"theme"`
	Language        string `json:"language"`
	Timezone        string `json:"timezone"`
	Dateformat      string `json:"date_format"`
	Timeformat      string `json:"time_format"`
}

// SecuritySettings 安全设置
type SecuritySettings struct {
	ForceHTTPS            bool   `json:"force_https"`
	SessionTimeout        int    `json:"session_timeout"`        // 会话超时(分钟)
	MaxLoginAttempts      int    `json:"max_login_attempts"`     // 最大登录尝试次数
	LockoutDuration       int    `json:"lockout_duration"`       // 锁定持续时间(分钟)
	PasswordMinLength     int    `json:"password_min_length"`    // 密码最小长度
	PasswordRequireUpper  bool   `json:"password_require_upper"` // 要求大写字母
	PasswordRequireLower  bool   `json:"password_require_lower"` // 要求小写字母
	PasswordRequireNumber bool   `json:"password_require_number"` // 要求数字
	PasswordRequireSymbol bool   `json:"password_require_symbol"` // 要求符号
	PasswordExpiryDays    int    `json:"password_expiry_days"`   // 密码过期天数
	IPWhitelist         string `json:"ip_whitelist"`           // IP白名单(逗号分隔)
	IPBlacklist         string `json:"ip_blacklist"`           // IP黑名单(逗号分隔)
	EnableTwoFactor     bool   `json:"enable_two_factor"`      // 启用双因素认证
}

// NotificationSettings 通知设置
type NotificationSettings struct {
	EmailNotifications  bool     `json:"email_notifications"`   // 启用邮件通知
	SMSNotifications    bool     `json:"sms_notifications"`     // 启用短信通知
	PushNotifications   bool     `json:"push_notifications"`    // 启用推送通知
	AdminEmails         []string `json:"admin_emails"`          // 管理员邮箱列表
	SecurityAlerts      bool     `json:"security_alerts"`       // 安全警报
	SystemAlerts        bool     `json:"system_alerts"`         // 系统警报
	UserActivityAlerts  bool     `json:"user_activity_alerts"`  // 用户活动警报
}

// PerformanceSettings 性能设置
type PerformanceSettings struct {
	CacheEnabled        bool `json:"cache_enabled"`         // 启用缓存
	CacheTTL            int  `json:"cache_ttl"`             // 缓存TTL(秒)
	QueryTimeout        int  `json:"query_timeout"`         // 查询超时(秒)
	MaxConnections      int  `json:"max_connections"`       // 最大连接数
	LogLevel            string `json:"log_level"`           // 日志级别
	LogRetentionDays    int  `json:"log_retention_days"`    // 日志保留天数
	CompressionEnabled  bool `json:"compression_enabled"`   // 启用压缩
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "admin_system_configs"
}

// GetValueAsJSON 将配置值解析为JSON
func (sc *SystemConfig) GetValueAsJSON(target interface{}) error {
	if sc.Value == "" {
		return nil
	}
	return json.Unmarshal([]byte(sc.Value), target)
}

// SetValueFromJSON 将对象序列化为JSON并设置为配置值
func (sc *SystemConfig) SetValueFromJSON(value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	sc.Value = string(data)
	return nil
}

// GetGeneralSettings 获取常规设置
func GetGeneralSettings(configs []SystemConfig) *GeneralSettings {
	settings := &GeneralSettings{
		SiteName:        "NOFX Admin Panel",
		SiteDescription: "NOFX Trading System Administration Panel",
		Theme:           "dark",
		Language:        "zh-CN",
		Timezone:        "Asia/Shanghai",
		Dateformat:      "YYYY-MM-DD",
		Timeformat:      "HH:mm:ss",
	}
	
	for _, config := range configs {
		if config.Key == "general_settings" {
			config.GetValueAsJSON(settings)
			break
		}
	}
	return settings
}

// GetSecuritySettings 获取安全设置
func GetSecuritySettings(configs []SystemConfig) *SecuritySettings {
	settings := &SecuritySettings{
		SessionTimeout:        60,
		MaxLoginAttempts:      5,
		LockoutDuration:       30,
		PasswordMinLength:     8,
		PasswordExpiryDays:    90,
	}
	
	for _, config := range configs {
		if config.Key == "security_settings" {
			config.GetValueAsJSON(settings)
			break
		}
	}
	return settings
}

// GetNotificationSettings 获取通知设置
func GetNotificationSettings(configs []SystemConfig) *NotificationSettings {
	settings := &NotificationSettings{
		EmailNotifications:  true,
		SecurityAlerts:      true,
		SystemAlerts:        true,
		UserActivityAlerts:  true,
	}
	
	for _, config := range configs {
		if config.Key == "notification_settings" {
			config.GetValueAsJSON(settings)
			break
		}
	}
	return settings
}

// GetPerformanceSettings 获取性能设置
func GetPerformanceSettings(configs []SystemConfig) *PerformanceSettings {
	settings := &PerformanceSettings{
		CacheEnabled:       true,
		CacheTTL:           300,
		QueryTimeout:       30,
		MaxConnections:     100,
		LogLevel:           "info",
		LogRetentionDays:   30,
		CompressionEnabled: true,
	}
	
	for _, config := range configs {
		if config.Key == "performance_settings" {
			config.GetValueAsJSON(settings)
			break
		}
	}
	return settings
}

// GetConfigByKey 根据键获取配置
func GetConfigByKey(key string) (*SystemConfig, error) {
	db := GetDB()
	var config SystemConfig
	err := db.DB.Where("key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("配置项 '%s' 不存在", key)
		}
		return nil, err
	}
	return &config, nil
}

// GetAllConfigs 获取所有配置
func GetAllConfigs() ([]SystemConfig, error) {
	db := GetDB()
	var configs []SystemConfig
	err := db.DB.Order("category ASC, key ASC").Find(&configs).Error
	return configs, err
}

// SetConfig 设置配置
func SetConfig(key, value, description string, category ConfigCategory, dataType string, isRequired, isEncrypted, isPublic bool, defaultValue, validationRules string) error {
	db := GetDB()
	config := SystemConfig{
		Key:             key,
		Value:           value,
		Category:        category,
		Description:     description,
		DataType:        dataType,
		IsRequired:      isRequired,
		IsEncrypted:     isEncrypted,
		IsPublic:        isPublic,
		DefaultValue:    defaultValue,
		ValidationRules: validationRules,
	}

	// 尝试查找现有配置，如果存在则更新，否则创建
	err := db.DB.Where(SystemConfig{Key: key}).Assign(config).FirstOrCreate(&config).Error
	return err
}

// DeleteConfig 删除配置
func DeleteConfig(key string) error {
	db := GetDB()
	result := db.DB.Where("key = ?", key).Delete(&SystemConfig{})
	return result.Error
}

// GetConfigsByCategory 根据分类获取配置
func GetConfigsByCategory(category ConfigCategory) ([]SystemConfig, error) {
	db := GetDB()
	var configs []SystemConfig
	err := db.DB.Where("category = ?", category).Order("key ASC").Find(&configs).Error
	return configs, err
}

// BulkUpdateConfigs 批量更新配置
func BulkUpdateConfigs(configs []SystemConfig) error {
	db := GetDB()
	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, config := range configs {
		if err := tx.Where(SystemConfig{Key: config.Key}).Assign(config).FirstOrCreate(&config).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新配置项 '%s' 失败: %v", config.Key, err)
		}
	}

	return tx.Commit().Error
}

// InitializeDefaultConfigs 初始化默认配置
func InitializeDefaultConfigs() error {
	defaultConfigs := []struct {
		Key             string
		Value           string
		Category        ConfigCategory
		Description     string
		DataType        string
		IsRequired      bool
		IsEncrypted     bool
		IsPublic        bool
		DefaultValue    string
		ValidationRules string
	}{
		{
			Key:         "app.name",
			Value:       "NOFX Admin System",
			Category:    GeneralConfig,
			Description: "应用名称",
			DataType:    "string",
		},
		{
			Key:         "app.version",
			Value:       "1.0.0",
			Category:    GeneralConfig,
			Description: "应用版本",
			DataType:    "string",
		},
		{
			Key:         "security.rate_limit.enabled",
			Value:       "true",
			Category:    SecurityConfig,
			Description: "启用速率限制",
			DataType:    "boolean",
			DefaultValue: "true",
		},
		{
			Key:         "security.rate_limit.requests_per_minute",
			Value:       "100",
			Category:    SecurityConfig,
			Description: "每分钟最大请求数",
			DataType:    "number",
			DefaultValue: "100",
		},
		{
			Key:         "security.session.timeout_minutes",
			Value:       "60",
			Category:    SecurityConfig,
			Description: "会话超时时间(分钟)",
			DataType:    "number",
			DefaultValue: "60",
		},
		{
			Key:         "logging.level",
			Value:       "info",
			Category:    PerformanceConfig,
			Description: "日志级别",
			DataType:    "string",
			DefaultValue: "info",
		},
	}

	for _, cfg := range defaultConfigs {
		if err := SetConfig(
			cfg.Key, cfg.Value, cfg.Description, cfg.Category,
			cfg.DataType, cfg.IsRequired, cfg.IsEncrypted,
			cfg.IsPublic, cfg.DefaultValue, cfg.ValidationRules,
		); err != nil {
			return fmt.Errorf("设置默认配置 '%s' 失败: %v", cfg.Key, err)
		}
	}

	return nil
}