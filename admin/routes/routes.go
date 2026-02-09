package routes

import (
	"nofx/admin/auth"
	"nofx/admin/config"
	"nofx/admin/controllers"
	"nofx/admin/models"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由器
func SetupRouter(cfg *config.Config) *gin.Engine {
	// 初始化数据库
	if err := models.InitDB(cfg); err != nil {
		panic(err)
	}

	// 创建认证服务
	authService := auth.NewAuthService(cfg)

	// 创建控制器
	adminCtrl := controllers.NewAdminController(authService)
	userCtrl := controllers.NewUserController(authService)
	permCtrl := controllers.NewPermissionController(authService)
	auditCtrl := controllers.NewAuditController(authService)
	configCtrl := controllers.NewSystemConfigController(authService)

	// 创建路由器
	r := gin.Default()

	// 应用中间件
	r.Use(authService.CORSMiddleware())
	r.Use(authService.SecurityHeadersMiddleware())
	r.Use(authService.IPWhitelistMiddleware())

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "admin-panel",
		})
	})

	// 认证相关路由 - 不需要认证
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", func(c *gin.Context) {
			var req auth.LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			response, err := authService.Login(c, req)
			if err != nil {
				c.JSON(401, gin.H{"error": err.Error()})
				return
			}

			c.JSON(200, response)
		})

		authGroup.POST("/refresh", func(c *gin.Context) {
			// 实现令牌刷新逻辑
			c.JSON(200, gin.H{"message": "refresh endpoint"})
		})
	}

	// 需要认证的路由组
	protected := r.Group("/api")
	protected.Use(authService.RequireAuth())
	{
		// 管理员仪表板和用户资料
		protected.GET("/dashboard", adminCtrl.GetDashboard)
		protected.GET("/profile", adminCtrl.GetProfile)
		protected.PUT("/profile", adminCtrl.UpdateProfile)
		protected.PUT("/profile/change-password", adminCtrl.ChangePassword)

		// 用户管理 - 需要用户管理权限
		users := protected.Group("/users")
		users.Use(authService.RequirePermission(models.UserManagePermission))
		{
			users.GET("", userCtrl.GetUsers)
			users.POST("", userCtrl.CreateUser)
			users.GET("/:id", userCtrl.GetUser)
			users.PUT("/:id", userCtrl.UpdateUser)
			users.DELETE("/:id", userCtrl.DeleteUser)
			users.POST("/:id/lock", userCtrl.LockUser)
			users.POST("/:id/unlock", userCtrl.UnlockUser)
		}

		// 权限管理 - 需要权限管理权限
		permissions := protected.Group("/permissions")
		permissions.Use(authService.RequirePermission(models.AuditManagePermission))
		{
			permissions.GET("", permCtrl.GetPermissions)
			permissions.POST("", permCtrl.GrantPermission)
			permissions.DELETE("/:id", permCtrl.RevokePermission)
			permissions.POST("/assign-role", permCtrl.AssignRole)
			permissions.DELETE("/:id/remove-role", permCtrl.RemoveRole)
		}

		// 系统配置 - 需要系统管理权限
		system := protected.Group("/system")
		system.Use(authService.RequirePermission(models.SystemManagePermission))
		{
			// 系统配置相关路由
			system.GET("/config", configCtrl.GetSystemConfig)
			system.PUT("/config", configCtrl.UpdateSystemConfig)
			system.GET("/config/:key", configCtrl.GetConfigByKey)
			system.POST("/config", configCtrl.CreateOrUpdateConfig)
			system.DELETE("/config/:key", configCtrl.DeleteConfig)
			system.GET("/config/category/:category", configCtrl.GetConfigByCategory)
			system.GET("/config/general", configCtrl.GetGeneralSettings)
			system.GET("/config/security", configCtrl.GetSecuritySettings)
			system.GET("/config/notification", configCtrl.GetNotificationSettings)
			system.GET("/config/performance", configCtrl.GetPerformanceSettings)
		}

		// 监控相关 - 需要监控权限
		monitoring := protected.Group("/monitoring")
		monitoring.Use(authService.RequirePermission(models.MonitorReadPermission))
		{
			monitoring.GET("/status", adminCtrl.GetSystemStatus)
			monitoring.GET("/metrics", adminCtrl.GetSystemMetrics)
		}

		// 审计日志 - 需要审计权限
		audit := protected.Group("/audit")
		audit.Use(authService.RequirePermission(models.AuditReadPermission))
		{
			audit.GET("", auditCtrl.GetAuditLogs)
			audit.GET("/:id", auditCtrl.GetAuditLogByID)
			audit.POST("", authService.RequirePermission(models.AuditWritePermission), auditCtrl.CreateAuditLog)
		}
	}

	return r
}