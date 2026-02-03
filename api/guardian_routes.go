package api

import (
	"github.com/gin-gonic/gin"
)

// registerGuardianRoutes 注册Guardian相关的API路由
func (s *Server) registerGuardianRoutes(api *gin.RouterGroup) {
	// Guardian AI endpoints for traders page integration
	api.GET("/guardian/providers", s.handleListGuardianProviders)
	api.POST("/guardian/call", s.handleCallGuardianAI)
	api.POST("/guardian/check-login", s.handleCheckGuardianLoginStatus)

	// Guardian browser automation test routes (no authentication required)
	api.POST("/test-guardian", s.handleTestGuardian)
	api.POST("/test-guardian-analysis", s.handleTestGuardianAnalysis)
	api.POST("/open-guardian-browser", s.handleOpenGuardianBrowser)

	// API bypass support endpoints
	api.GET("/bypass/support", s.handleGetBypassSupport)
	api.POST("/bypass/toggle", s.handleToggleModelBypass)
}
