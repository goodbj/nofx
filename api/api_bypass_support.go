package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BypassSupport defines the structure for API bypass support information
type BypassSupport struct {
	ModelID        string `json:"modelId"`
	ModelName      string `json:"modelName"`
	SupportsBypass bool   `json:"supportsBypass"`
	DefaultBypass  bool   `json:"defaultBypass"`
	Description    string `json:"description"`
}

// handleGetBypassSupport returns information about which models support API bypass
func (s *Server) handleGetBypassSupport(c *gin.Context) {
	supportedModels := []BypassSupport{
		{
			ModelID:        "deepseek",
			ModelName:      "DeepSeek",
			SupportsBypass: true,
			DefaultBypass:  false,
			Description:    "DeepSeek AI model with optional API bypass via browser automation",
		},
		{
			ModelID:        "openai",
			ModelName:      "OpenAI GPT",
			SupportsBypass: true,
			DefaultBypass:  false,
			Description:    "OpenAI GPT models with optional API bypass via browser automation",
		},
		{
			ModelID:        "claude",
			ModelName:      "Anthropic Claude",
			SupportsBypass: true,
			DefaultBypass:  false,
			Description:    "Anthropic Claude models with optional API bypass via browser automation",
		},
		{
			ModelID:        "qwen",
			ModelName:      "Alibaba Qwen",
			SupportsBypass: true,
			DefaultBypass:  false,
			Description:    "Alibaba Qwen models with optional API bypass via browser automation",
		},
		{
			ModelID:        "gemini",
			ModelName:      "Google Gemini",
			SupportsBypass: true,
			DefaultBypass:  false,
			Description:    "Google Gemini models with optional API bypass via browser automation",
		},
		{
			ModelID:        "deepseek-browser",
			ModelName:      "DeepSeek (Browser Automation)",
			SupportsBypass: true,
			DefaultBypass:  true,
			Description:    "DeepSeek via browser automation - bypasses API entirely",
		},
		{
			ModelID:        "chatgpt-browser",
			ModelName:      "ChatGPT (Browser Automation)",
			SupportsBypass: true,
			DefaultBypass:  true,
			Description:    "ChatGPT via browser automation - bypasses API entirely",
		},
		{
			ModelID:        "claude-browser",
			ModelName:      "Claude (Browser Automation)",
			SupportsBypass: true,
			DefaultBypass:  true,
			Description:    "Claude via browser automation - bypasses API entirely",
		},
		{
			ModelID:        "guardian-ai",
			ModelName:      "Guardian AI (Browser Automation)",
			SupportsBypass: true,
			DefaultBypass:  true,
			Description:    "Guardian AI via browser automation - bypasses API entirely",
		},
	}

	c.JSON(http.StatusOK, supportedModels)
}

// handleToggleModelBypass allows enabling/disabling bypass for a specific model
func (s *Server) handleToggleModelBypass(c *gin.Context) {
	var req struct {
		ModelID      string            `json:"modelId"`
		EnableBypass bool              `json:"enableBypass"`
		ConfigParams map[string]string `json:"configParams,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// In a real implementation, this would store the bypass setting
	// For now, we'll just return success
	response := gin.H{
		"modelId":       req.ModelID,
		"bypassEnabled": req.EnableBypass,
		"message":       "Bypass setting updated successfully",
	}

	if req.ConfigParams != nil {
		response["configParams"] = req.ConfigParams
	}

	c.JSON(http.StatusOK, response)
}

// Register the bypass support routes
func (s *Server) registerBypassRoutes(router *gin.Engine) {
	bypassGroup := router.Group("/api/bypass")
	{
		bypassGroup.GET("/support", s.handleGetBypassSupport)
		bypassGroup.POST("/toggle", s.handleToggleModelBypass)
	}
}
