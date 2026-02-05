package api

import (
	"net/http"

	"nofx/mcp"
	"nofx/trader"

	"github.com/gin-gonic/gin"
)

// handleCallGuardianAI handles requests to call Guardian AI services
func (s *Server) handleCallGuardianAI(c *gin.Context) {
	var req struct {
		Provider     string `json:"provider"` // AI provider (e.g., "deepseek-browser", "guardian-ai")
		SystemPrompt string `json:"systemPrompt"`
		UserPrompt   string `json:"userPrompt"`
		TargetURL    string `json:"targetUrl,omitempty"`  // Optional target URL for custom services
		CheckLogin   bool   `json:"checkLogin,omitempty"` // Whether to check login status only
		TraderID     string `json:"traderId,omitempty"`   // Trader ID for browser data isolation
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// If only checking login status
	if req.CheckLogin {
		// Determine the target URL based on provider if not provided
		targetURL := req.TargetURL
		if targetURL == "" {
			switch req.Provider {
			case "deepseek":
				targetURL = "https://chat.deepseek.com"
			case "chatgpt":
				targetURL = "https://chat.openai.com"
			case "claude":
				targetURL = "https://claude.ai"
			default:
				targetURL = "https://chat.deepseek.com" // Default
			}
		}

		// Create a temporary client to check login status
		var tempClient mcp.AIClient
		if req.TraderID != "" {
			tempClient = mcp.NewGuardianClientWithOptions(mcp.WithTraderID(req.TraderID))
		} else {
			tempClient = mcp.NewGuardianClient()
		}
		if tempClient == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Guardian client"})
			return
		}

		// Type assert to GuardianClient to access CheckLoginStatus method
		guardianClient, ok := tempClient.(*mcp.GuardianClient)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast client to GuardianClient"})
			return
		}

		isLoggedIn, err := guardianClient.CheckLoginStatus(targetURL)
		if err != nil {
			s.logger.Errorf("Error checking login status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to check login status",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"isLoggedIn":  isLoggedIn,
			"provider":    req.Provider,
			"targetUrl":   targetURL,
			"description": "Login status check completed",
		})
		return
	}

	// Create the appropriate AI client based on the provider
	var aiClient mcp.AIClient
	var err error

	// First, try to create from the browser AI registry
	browserProvider, regErr := mcp.CreateBrowserAIProvider(req.Provider)
	if regErr == nil {
		aiClient = browserProvider.(mcp.AIClient)

		// If target URL is provided and the provider supports dynamic configuration, use it
		if req.TargetURL != "" {
			if guardianProvider, ok := aiClient.(*mcp.GuardianAIProvider); ok {
				guardianProvider.SetDynamicService(req.Provider, req.TargetURL)
			}
		}
	} else {
		// If not found in registry, try to create a standard Guardian client
		switch req.Provider {
		case "deepseek":
			aiClient = mcp.NewGuardianClientWithService("deepseek")
		case "chatgpt":
			aiClient = mcp.NewGuardianClientWithService("chatgpt")
		case "claude":
			aiClient = mcp.NewGuardianClientWithService("claude")
		default:
			// Default to general guardian-ai
			aiClient = mcp.NewGuardianClientWithService("deepseek")
		}
	}

	if aiClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create AI client"})
		return
	}

	// Call the AI service
	response, err := aiClient.CallWithMessages(req.SystemPrompt, req.UserPrompt)
	if err != nil {
		s.logger.Errorf("Error calling Guardian AI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service call failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"response": response,
		"provider": req.Provider,
	})
}

// handleCheckGuardianLoginStatus checks the login status for a given provider
func (s *Server) handleCheckGuardianLoginStatus(c *gin.Context) {
	var req struct {
		Provider  string `json:"provider"`            // AI provider (e.g., "deepseek-browser", "guardian-ai")
		TargetURL string `json:"targetUrl,omitempty"` // Optional target URL for custom services
		AIService string `json:"aiService,omitempty"` // AI service type for configuration
		TraderID  string `json:"traderId,omitempty"`  // Trader ID for browser data isolation
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Determine the target URL based on provider if not provided
	targetURL := req.TargetURL
	if targetURL == "" {
		switch req.Provider {
		case "deepseek":
			targetURL = "https://chat.deepseek.com"
		case "chatgpt":
			targetURL = "https://chat.openai.com"
		case "claude":
			targetURL = "https://claude.ai"
		default:
			targetURL = "https://chat.deepseek.com" // Default
		}
	}

	// Create a temporary client to check login status
	var tempClient mcp.AIClient
	if req.AIService != "" {
		// 如果有交易员ID，使用带交易员ID的创建函数
		if req.TraderID != "" {
			tempClient = trader.GetGuardianClientWithTargetAndTraderID(req.AIService, req.TraderID)
		} else {
			tempClient = mcp.NewGuardianClientForBrowserWithService(req.AIService)
		}
	} else {
		// 如果有交易员ID，使用带交易员ID的创建函数
		if req.TraderID != "" {
			tempClient = mcp.NewGuardianClientWithOptions(mcp.WithTraderID(req.TraderID))
		} else {
			tempClient = mcp.NewGuardianClient()
		}
	}
	if tempClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Guardian client"})
		return
	}

	// Type assert to GuardianClient to access CheckLoginStatus method
	guardianClient, ok := tempClient.(*mcp.GuardianClient)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cast client to GuardianClient"})
		return
	}

	isLoggedIn, err := guardianClient.CheckLoginStatus(targetURL)
	if err != nil {
		s.logger.Errorf("Error checking login status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to check login status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"isLoggedIn":  isLoggedIn,
		"provider":    req.Provider,
		"targetUrl":   targetURL,
		"aiService":   req.AIService,
		"description": "Login status check completed",
	})
}

// handleListGuardianProviders returns a list of available Guardian AI providers
func (s *Server) handleListGuardianProviders(c *gin.Context) {
	providers := mcp.ListBrowserAIProviders()

	// Add standard Guardian providers that might not be in the registry
	standardProviders := []string{"deepseek", "chatgpt", "claude"}

	// Combine and deduplicate
	allProviders := make(map[string]bool)
	for _, p := range providers {
		allProviders[p] = true
	}

	for _, p := range standardProviders {
		if !allProviders[p] {
			providers = append(providers, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
