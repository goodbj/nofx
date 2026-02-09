package models

import (
	"fmt"
	"log"
	"nofx/admin/config"
	
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库实例
type Database struct {
	DB *gorm.DB
}

var dbInstance *Database

// GetDB 获取数据库实例
func GetDB() *Database {
	if dbInstance == nil {
		panic("数据库未初始化")
	}
	return dbInstance
}

// InitDB 初始化数据库
func InitDB(cfg *config.Config) error {
	var dialector gorm.Dialector
	
	switch cfg.DBType {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.DBPath)
	default:
		return fmt.Errorf("不支持的数据库类型: %s", cfg.DBType)
	}
	
	// 创建数据库连接
	newDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 根据配置调整日志级别
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}
	
	// 迁移数据库表
	if err := migrateTables(newDB); err != nil {
		return fmt.Errorf("迁移数据库表失败: %v", err)
	}
	
	dbInstance = &Database{
		DB: newDB,
	}
	
	log.Println("✅ 数据库初始化成功")
	return nil
}

// migrateTables 迁移数据库表
func migrateTables(db *gorm.DB) error {
	// 按依赖顺序创建表
	tables := []interface{}{
		&AdminUser{},
		&PermissionAssignment{},
		&LoginLog{},
		&AuditLog{},
		&SystemConfig{},
	}
	
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return fmt.Errorf("迁移表失败: %v", err)
		}
	}
	
	// 创建默认超级管理员账户（如果不存在）
	if err := createDefaultSuperAdmin(db); err != nil {
		return fmt.Errorf("创建默认管理员失败: %v", err)
	}
	
	// 创建默认系统配置（如果不存在）
	if err := createDefaultSystemConfigs(db); err != nil {
		return fmt.Errorf("创建默认系统配置失败: %v", err)
	}
	
	return nil
}

// createDefaultSuperAdmin 创建默认超级管理员账户
func createDefaultSuperAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&AdminUser{}).Count(&count).Error; err != nil {
		return err
	}
	
	// 如果已有管理员账户，则跳过创建
	if count > 0 {
		log.Println("ℹ️  管理员账户已存在，跳过创建默认账户")
		return nil
	}
	
	// 创建默认超级管理员
	superAdmin := &AdminUser{
		ID:       "00000000-0000-0000-0000-000000000001", // 使用固定UUID便于识别
		Username: "superadmin",
		Email:    "superadmin@nofx.local",
		Role:     SuperAdmin,
		Status:   ActiveStatus,
	}
	
	// 设置默认密码（需要先加密）
	password := "SuperAdmin123!" // 默认密码，生产环境需要修改
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	superAdmin.Password = hashedPassword
	
	if err := db.Create(superAdmin).Error; err != nil {
		return fmt.Errorf("创建超级管理员失败: %v", err)
	}
	
	log.Printf("🔐 已创建默认超级管理员账户: %s / %s", superAdmin.Username, password)
	return nil
}

// createDefaultSystemConfigs 创建默认系统配置
func createDefaultSystemConfigs(db *gorm.DB) error {
	// 检查是否已有配置
	var count int64
	if err := db.Model(&SystemConfig{}).Count(&count).Error; err != nil {
		return err
	}
	
	if count > 0 {
		log.Println("ℹ️  系统配置已存在，跳过创建默认配置")
		return nil
	}
	
	// 创建默认配置
	configs := []*SystemConfig{
		{
			Key:         "general_settings",
			Category:    GeneralConfig,
			Description: "常规系统设置",
			DataType:    "json",
			IsRequired:  false,
			IsEncrypted: false,
		},
		{
			Key:         "security_settings",
			Category:    SecurityConfig,
			Description: "安全系统设置",
			DataType:    "json",
			IsRequired:  false,
			IsEncrypted: false,
		},
		{
			Key:         "notification_settings",
			Category:    NotificationConfig,
			Description: "通知系统设置",
			DataType:    "json",
			IsRequired:  false,
			IsEncrypted: false,
		},
		{
			Key:         "performance_settings",
			Category:    PerformanceConfig,
			Description: "性能系统设置",
			DataType:    "json",
			IsRequired:  false,
			IsEncrypted: false,
		},
	}
	
	for _, config := range configs {
		// 根据配置类型设置默认值
		switch config.Key {
		case "general_settings":
			settings := &GeneralSettings{
				SiteName:        "NOFX Admin Panel",
				SiteDescription: "NOFX Trading System Administration Panel",
				Theme:           "dark",
				Language:        "zh-CN",
				Timezone:        "Asia/Shanghai",
			}
			config.SetValueFromJSON(settings)
		case "security_settings":
			settings := &SecuritySettings{
				SessionTimeout:        60,
				MaxLoginAttempts:      5,
				LockoutDuration:       30,
				PasswordMinLength:     8,
				PasswordExpiryDays:    90,
			}
			config.SetValueFromJSON(settings)
		case "notification_settings":
			settings := &NotificationSettings{
				EmailNotifications:  true,
				SecurityAlerts:      true,
				SystemAlerts:        true,
				UserActivityAlerts:  true,
			}
			config.SetValueFromJSON(settings)
		case "performance_settings":
			settings := &PerformanceSettings{
				CacheEnabled:       true,
				CacheTTL:           300,
				QueryTimeout:       30,
				MaxConnections:     100,
				LogLevel:           "info",
				LogRetentionDays:   30,
				CompressionEnabled: true,
			}
			config.SetValueFromJSON(settings)
		}
		
		if err := db.Create(config).Error; err != nil {
			return fmt.Errorf("创建系统配置 %s 失败: %v", config.Key, err)
		}
	}
	
	log.Println("⚙️  已创建默认系统配置")
	return nil
}

// HashPassword 密码加密函数（将在auth包中实现）
func HashPassword(password string) (string, error) {
	// 这里只是占位符，实际实现会在auth包中
	return password, nil // 仅为演示，实际需要使用bcrypt等加密
}