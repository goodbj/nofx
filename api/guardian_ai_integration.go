package api

import (
	"net/http"

	"nofx/mcp"

	"github.com/gin-gonic/gin"
)

// handleCallGuardianAI handles requests to call Guardian AI services
func (s *Server) handleCallGuardianAI(c *gin.Context) {
	var req struct {
		Provider     string `json:"provider"` // AI provider (e.g., "deepseek-browser", "guardian-ai")
		SystemPrompt string `json:"systemPrompt"`
		UserPrompt   string `json:"userPrompt"`
		TargetURL    string `json:"targetUrl,omitempty"` // Optional target URL for custom services
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
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
		case "deepseek", "deepseek-browser":
			aiClient = mcp.NewGuardianClientWithService("deepseek")
		case "chatgpt", "chatgpt-browser":
			aiClient = mcp.NewGuardianClientWithService("chatgpt")
		case "claude", "claude-browser":
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
