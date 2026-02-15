package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"nofx/kernel"
	"nofx/logger"
	"nofx/market"
	"nofx/mcp"
	"nofx/store"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// validateStrategyConfig validates strategy configuration and returns warnings
func validateStrategyConfig(config *store.StrategyConfig) []string {
	var warnings []string

	// Validate NofxOS API key if any NofxOS feature is enabled
	if (config.Indicators.EnableQuantData || config.Indicators.EnableOIRanking ||
		config.Indicators.EnableNetFlowRanking || config.Indicators.EnablePriceRanking) &&
		config.Indicators.NofxOSAPIKey == "" {
		warnings = append(warnings, "NofxOS API key is not configured. NofxOS data sources may not work properly.")
	}

	return warnings
}

// handlePublicStrategies Get public strategies for strategy market (no auth required)
func (s *Server) handlePublicStrategies(c *gin.Context) {
	strategies, err := s.store.Strategy().ListPublic()
	if err != nil {
		SafeInternalError(c, "Failed to get public strategies", err)
		return
	}

	// Convert to frontend format with visibility control
	result := make([]gin.H, 0, len(strategies))
	for _, st := range strategies {
		item := gin.H{
			"id":             st.ID,
			"name":           st.Name,
			"description":    st.Description,
			"author_email":   "", // Will be filled if we have user info
			"is_public":      st.IsPublic,
			"config_visible": st.ConfigVisible,
			"created_at":     st.CreatedAt,
			"updated_at":     st.UpdatedAt,
		}

		// Only include config if config_visible is true
		if st.ConfigVisible {
			var config store.StrategyConfig
			json.Unmarshal([]byte(st.Config), &config)
			item["config"] = config
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": result,
	})
}

// handleGetStrategies Get strategy list
func (s *Server) handleGetStrategies(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	strategies, err := s.store.Strategy().List(userID)
	if err != nil {
		SafeInternalError(c, "Failed to get strategy list", err)
		return
	}

	// Convert to frontend format
	result := make([]gin.H, 0, len(strategies))
	for _, st := range strategies {
		var config store.StrategyConfig
		json.Unmarshal([]byte(st.Config), &config)

		result = append(result, gin.H{
			"id":             st.ID,
			"name":           st.Name,
			"description":    st.Description,
			"is_active":      st.IsActive,
			"is_default":     st.IsDefault,
			"is_public":      st.IsPublic,
			"config_visible": st.ConfigVisible,
			"config":         config,
			"created_at":     st.CreatedAt,
			"updated_at":     st.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": result,
	})
}

// handleGetStrategy Get single strategy
func (s *Server) handleGetStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	strategyID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	strategy, err := s.store.Strategy().Get(userID, strategyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Strategy not found"})
		return
	}

	var config store.StrategyConfig
	json.Unmarshal([]byte(strategy.Config), &config)

	c.JSON(http.StatusOK, gin.H{
		"id":          strategy.ID,
		"name":        strategy.Name,
		"description": strategy.Description,
		"is_active":   strategy.IsActive,
		"is_default":  strategy.IsDefault,
		"config":      config,
		"created_at":  strategy.CreatedAt,
		"updated_at":  strategy.UpdatedAt,
	})
}

// handleCreateStrategy Create strategy
func (s *Server) handleCreateStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name        string               `json:"name" binding:"required"`
		Description string               `json:"description"`
		Config      store.StrategyConfig `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	// Serialize configuration
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		SafeInternalError(c, "Serialize configuration", err)
		return
	}

	strategy := &store.Strategy{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    false,
		IsDefault:   false,
		Config:      string(configJSON),
	}

	if err := s.store.Strategy().Create(strategy); err != nil {
		SafeInternalError(c, "Failed to create strategy", err)
		return
	}

	// Validate configuration and collect warnings
	warnings := validateStrategyConfig(&req.Config)

	response := gin.H{
		"id":      strategy.ID,
		"message": "Strategy created successfully",
	}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	c.JSON(http.StatusOK, response)
}

// handleUpdateStrategy Update strategy
func (s *Server) handleUpdateStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	strategyID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if it's a system default strategy
	existing, err := s.store.Strategy().Get(userID, strategyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Strategy not found"})
		return
	}
	if existing.IsDefault {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system default strategy"})
		return
	}

	var req struct {
		Name          string               `json:"name"`
		Description   string               `json:"description"`
		Config        store.StrategyConfig `json:"config"`
		IsPublic      bool                 `json:"is_public"`
		ConfigVisible bool                 `json:"config_visible"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	// Serialize configuration
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		SafeInternalError(c, "Serialize configuration", err)
		return
	}

	strategy := &store.Strategy{
		ID:            strategyID,
		UserID:        userID,
		Name:          req.Name,
		Description:   req.Description,
		Config:        string(configJSON),
		IsPublic:      req.IsPublic,
		ConfigVisible: req.ConfigVisible,
	}

	if err := s.store.Strategy().Update(strategy); err != nil {
		SafeInternalError(c, "Failed to update strategy", err)
		return
	}

	// Validate configuration and collect warnings
	warnings := validateStrategyConfig(&req.Config)

	response := gin.H{"message": "Strategy updated successfully"}
	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	c.JSON(http.StatusOK, response)
}

// handleDeleteStrategy Delete strategy
func (s *Server) handleDeleteStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	strategyID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := s.store.Strategy().Delete(userID, strategyID); err != nil {
		SafeInternalError(c, "Failed to delete strategy", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Strategy deleted successfully"})
}

// handleActivateStrategy Activate strategy
func (s *Server) handleActivateStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	strategyID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := s.store.Strategy().SetActive(userID, strategyID); err != nil {
		SafeInternalError(c, "Failed to activate strategy", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Strategy activated successfully"})
}

// handleDuplicateStrategy Duplicate strategy
func (s *Server) handleDuplicateStrategy(c *gin.Context) {
	userID := c.GetString("user_id")
	sourceID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	newID := uuid.New().String()
	if err := s.store.Strategy().Duplicate(userID, sourceID, newID, req.Name); err != nil {
		SafeInternalError(c, "Failed to duplicate strategy", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      newID,
		"message": "Strategy duplicated successfully",
	})
}

// handleGetActiveStrategy Get currently active strategy
func (s *Server) handleGetActiveStrategy(c *gin.Context) {
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	strategy, err := s.store.Strategy().GetActive(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active strategy"})
		return
	}

	var config store.StrategyConfig
	json.Unmarshal([]byte(strategy.Config), &config)

	c.JSON(http.StatusOK, gin.H{
		"id":          strategy.ID,
		"name":        strategy.Name,
		"description": strategy.Description,
		"is_active":   strategy.IsActive,
		"is_default":  strategy.IsDefault,
		"config":      config,
		"created_at":  strategy.CreatedAt,
		"updated_at":  strategy.UpdatedAt,
	})
}

// handleGetDefaultStrategyConfig Get default strategy configuration template
func (s *Server) handleGetDefaultStrategyConfig(c *gin.Context) {
	// Get language from query parameter, default to "en"
	lang := c.Query("lang")
	if lang != "zh" {
		lang = "en"
	}

	// Return default configuration with i18n support
	defaultConfig := store.GetDefaultStrategyConfig(lang)
	c.JSON(http.StatusOK, defaultConfig)
}

// handleGetAnnotatedStrategyConfig Get annotated strategy configuration template for AI understanding
func (s *Server) handleGetAnnotatedStrategyConfig(c *gin.Context) {
	// Get language from query parameter, default to "en"
	lang := c.Query("lang")
	if lang != "zh" {
		lang = "en"
	}

	// Create annotated configuration template
	annotatedConfig := gin.H{
		"// strategy_config_explanation": "策略配置文件结构说明",
		"// version":                     "1.0",
		"// last_updated":                "2025-12-19",

		"name":        "策略名称",
		"description": "策略描述",
		"config": gin.H{
			"coin_source": gin.H{
				"// coin_source_explanation": "币种来源配置",
				"source_type":                "static|ai500|oi_top|mixed",
				"symbols":                    []string{"BTCUSDT", "ETHUSDT"},
				"limit":                      10,
			},

			"indicators": gin.H{
				"// indicators_explanation": "技术指标配置",
				"enable_ema":                true,
				"enable_macd":               true,
				"enable_rsi":                true,
				"enable_atr":                true,
				"enable_boll":               true,
				"enable_volume":             true,
				"enable_oi":                 true,
				"enable_funding_rate":       true,
				"klines": gin.H{
					"primary_timeframe":      "5m",
					"primary_count":          120,
					"longer_timeframe":       "4h",
					"longer_count":           30,
					"enable_multi_timeframe": true,
					"selected_timeframes":    []string{"5m", "15m", "1h", "4h"},
					"timeframe_counts": gin.H{
						"1m":  120,
						"3m":  120,
						"5m":  120,
						"15m": 80,
						"30m": 60,
						"1h":  50,
						"2h":  40,
						"4h":  30,
						"1d":  20,
						"1w":  10,
					},
					"trading_style_preset": "short",
				},
			},

			"risk_control": gin.H{
				"// risk_control_explanation": "风险控制配置",

				"// position_limits":    "仓位限制（代码强制执行）",
				"max_positions":         3,
				"// max_positions_desc": "最大同时持仓数量",

				"// trading_leverage":          "交易杠杆（AI指导）",
				"btc_eth_max_leverage":         5,
				"// btc_eth_max_leverage_desc": "BTC/ETH最大交易所杠杆倍数",
				"altcoin_max_leverage":         5,
				"// altcoin_max_leverage_desc": "山寨币最大交易所杠杆倍数",

				"// position_value_ratio":                  "仓位价值比例（代码强制执行）",
				"btc_eth_max_position_value_ratio":         5.0,
				"// btc_eth_max_position_value_ratio_desc": "BTC/ETH单仓位最大名义价值与账户净值的比例",
				"altcoin_max_position_value_ratio":         1.0,
				"// altcoin_max_position_value_ratio_desc": "山寨币单仓位最大名义价值与账户净值的比例",

				"// risk_parameters":            "风险参数（AI指导）",
				"max_margin_usage":              0.9,
				"// max_margin_usage_desc":      "最大保证金使用率，如0.9表示90%",
				"min_position_size":             12,
				"// min_position_size_desc":     "最小仓位规模，单位USDT",
				"min_risk_reward_ratio":         3.0,
				"// min_risk_reward_ratio_desc": "最小风险回报比，即止盈/止损的最小比例",
				"min_confidence":                75,
				"// min_confidence_desc":        "AI开仓的最小信心度百分比",

				"// additional_risk_controls":            "附加风险控制（代码强制执行）",
				"max_daily_trades":                       10,
				"// max_daily_trades_desc":               "每日最大交易次数",
				"max_hourly_trades":                      3,
				"// max_hourly_trades_desc":              "每小时最大交易次数",
				"max_trades_per_symbol_per_hour":         1,
				"// max_trades_per_symbol_per_hour_desc": "每小时每交易品种最大交易次数",
				"min_hold_time_minutes":                  8,
				"// min_hold_time_minutes_desc":          "最小持仓时间，单位分钟",
				"max_loss_per_trade_percent":             3.0,
				"// max_loss_per_trade_percent_desc":     "单笔交易最大亏损百分比",
				"daily_loss_limit_percent":               2.0,
				"// daily_loss_limit_percent_desc":       "每日总亏损限制百分比",
			},

			"prompt_sections": gin.H{
				"// prompt_sections_explanation": "AI提示词配置",
				"enable_technical_analysis":      true,
				"enable_market_sentiment":        true,
				"enable_risk_management":         true,
				"custom_prompt_additions":        []string{},
			},

			"publish_settings": gin.H{
				"// publish_settings_explanation": "发布设置",
				"allow_public_sharing":            false,
				"allow_cloning":                   true,
				"share_performance_data":          false,
			},
		},

		"// usage_notes": []string{
			"1. 代码强制执行的参数会在交易引擎中被强制遵守",
			"2. AI指导的参数会作为AI决策的参考依据",
			"3. 所有百分比值使用小数形式，如0.9代表90%",
			"4. 时间相关参数以分钟为单位",
		},
	}

	c.JSON(http.StatusOK, annotatedConfig)
}

// handlePreviewPrompt Preview prompt generated by strategy
func (s *Server) handlePreviewPrompt(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Config        store.StrategyConfig `json:"config" binding:"required"`
		AccountEquity float64              `json:"account_equity"`
		PromptVariant string               `json:"prompt_variant"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	// Use default values
	if req.AccountEquity <= 0 {
		req.AccountEquity = 1000.0 // Default simulated account equity
	}
	if req.PromptVariant == "" {
		req.PromptVariant = "balanced"
	}

	// Create strategy engine to build prompt
	engine := kernel.NewStrategyEngine(&req.Config)

	// Build system prompt (using built-in method from strategy engine)
	systemPrompt := engine.BuildSystemPrompt(
		req.AccountEquity,
		req.PromptVariant,
	)

	c.JSON(http.StatusOK, gin.H{
		"system_prompt":  systemPrompt,
		"prompt_variant": req.PromptVariant,
		"config_summary": gin.H{
			"coin_source":      req.Config.CoinSource.SourceType,
			"primary_tf":       req.Config.Indicators.Klines.PrimaryTimeframe,
			"primary_count":    req.Config.Indicators.Klines.PrimaryCount,
			"selected_tfs":     req.Config.Indicators.Klines.SelectedTimeframes,
			"btc_eth_leverage": req.Config.RiskControl.BTCETHMaxLeverage,
			"altcoin_leverage": req.Config.RiskControl.AltcoinMaxLeverage,
			"max_positions":    req.Config.RiskControl.MaxPositions,
		},
	})
}

// handleStrategyTestRun AI test run (does not execute trades, only returns AI analysis results)
func (s *Server) handleStrategyTestRun(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Config        store.StrategyConfig `json:"config" binding:"required"`
		PromptVariant string               `json:"prompt_variant"`
		AIModelID     string               `json:"ai_model_id"`
		RunRealAI     bool                 `json:"run_real_ai"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	if req.PromptVariant == "" {
		req.PromptVariant = "balanced"
	}

	// Create strategy engine to build prompt
	engine := kernel.NewStrategyEngine(&req.Config)

	// Get candidate coins
	candidates, err := engine.GetCandidateCoins()
	if err != nil {
		logger.Errorf("[API Error] Failed to get candidate coins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       "Failed to get candidate coins",
			"ai_response": "",
		})
		return
	}

	// Get timeframe configuration
	timeframes := req.Config.Indicators.Klines.SelectedTimeframes
	primaryTimeframe := req.Config.Indicators.Klines.PrimaryTimeframe
	klineCount := req.Config.Indicators.Klines.PrimaryCount
	timeframeCounts := req.Config.Indicators.Klines.TimeframeCounts
	if timeframeCounts == nil {
		timeframeCounts = make(map[string]int)
	}

	// If no timeframes selected, use default values
	if len(timeframes) == 0 {
		// Backward compatibility: use primary and longer timeframes
		if primaryTimeframe != "" {
			timeframes = append(timeframes, primaryTimeframe)
		} else {
			timeframes = append(timeframes, "3m")
		}
		if req.Config.Indicators.Klines.LongerTimeframe != "" {
			timeframes = append(timeframes, req.Config.Indicators.Klines.LongerTimeframe)
		}
	}
	if primaryTimeframe == "" {
		primaryTimeframe = timeframes[0]
	}
	if klineCount <= 0 {
		klineCount = 30
	}

	fmt.Printf("📊 Using timeframes: %v, primary: %s, kline count: %d\n", timeframes, primaryTimeframe, klineCount)

	// Get real market data (using multiple timeframes with individual counts)
	marketDataMap := make(map[string]*market.Data)
	for _, coin := range candidates {
		data, err := market.GetWithTimeframesAndCounts(coin.Symbol, timeframes, primaryTimeframe, timeframeCounts)
		if err != nil {
			// If getting data for a coin fails, log but continue
			fmt.Printf("⚠️  Failed to get market data for %s: %v\n", coin.Symbol, err)
			continue
		}
		marketDataMap[coin.Symbol] = data
	}

	// Fetch quantitative data for each candidate coin
	symbols := make([]string, 0, len(candidates))
	for _, c := range candidates {
		symbols = append(symbols, c.Symbol)
	}
	quantDataMap := engine.FetchQuantDataBatch(symbols)

	// Fetch OI ranking data (market-wide position changes)
	oiRankingData := engine.FetchOIRankingData()

	// Fetch NetFlow ranking data (market-wide fund flow)
	netFlowRankingData := engine.FetchNetFlowRankingData()

	// Fetch Price ranking data (market-wide gainers/losers)
	priceRankingData := engine.FetchPriceRankingData()

	// Build real context (for generating User Prompt)
	testContext := &kernel.Context{
		CurrentTime:    time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		RuntimeMinutes: 0,
		CallCount:      1,
		Account: kernel.AccountInfo{
			TotalEquity:      1000.0,
			AvailableBalance: 1000.0,
			UnrealizedPnL:    0,
			TotalPnL:         0,
			TotalPnLPct:      0,
			MarginUsed:       0,
			MarginUsedPct:    0,
			PositionCount:    0,
		},
		Positions:          []kernel.PositionInfo{},
		CandidateCoins:     candidates,
		PromptVariant:      req.PromptVariant,
		MarketDataMap:      marketDataMap,
		QuantDataMap:       quantDataMap,
		OIRankingData:      oiRankingData,
		NetFlowRankingData: netFlowRankingData,
		PriceRankingData:   priceRankingData,
	}

	// Build System Prompt
	systemPrompt := engine.BuildSystemPrompt(1000.0, req.PromptVariant)

	// Build User Prompt (using real market data)
	userPrompt := engine.BuildUserPrompt(testContext)

	// If requesting real AI call
	if req.RunRealAI && req.AIModelID != "" {
		aiResponse, aiErr := s.runRealAITest(userID, req.AIModelID, systemPrompt, userPrompt)
		if aiErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"system_prompt":   systemPrompt,
				"user_prompt":     userPrompt,
				"candidate_count": len(candidates),
				"candidates":      candidates,
				"prompt_variant":  req.PromptVariant,
				"config_summary": gin.H{
					"coin_source":      req.Config.CoinSource.SourceType,
					"primary_tf":       req.Config.Indicators.Klines.PrimaryTimeframe,
					"primary_count":    req.Config.Indicators.Klines.PrimaryCount,
					"selected_tfs":     req.Config.Indicators.Klines.SelectedTimeframes,
					"btc_eth_leverage": req.Config.RiskControl.BTCETHMaxLeverage,
					"altcoin_leverage": req.Config.RiskControl.AltcoinMaxLeverage,
					"max_positions":    req.Config.RiskControl.MaxPositions,
				},
				"ai_response": fmt.Sprintf("❌ AI call failed: %s", aiErr.Error()),
				"ai_error":    aiErr.Error(),
				"note":        "AI call error",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"system_prompt":   systemPrompt,
			"user_prompt":     userPrompt,
			"candidate_count": len(candidates),
			"candidates":      candidates,
			"prompt_variant":  req.PromptVariant,
			"config_summary": gin.H{
				"coin_source":      req.Config.CoinSource.SourceType,
				"primary_tf":       req.Config.Indicators.Klines.PrimaryTimeframe,
				"primary_count":    req.Config.Indicators.Klines.PrimaryCount,
				"selected_tfs":     req.Config.Indicators.Klines.SelectedTimeframes,
				"btc_eth_leverage": req.Config.RiskControl.BTCETHMaxLeverage,
				"altcoin_leverage": req.Config.RiskControl.AltcoinMaxLeverage,
				"max_positions":    req.Config.RiskControl.MaxPositions,
			},
			"ai_response": aiResponse,
			"note":        "✅ Real AI test run successful",
		})
		return
	}

	// Return result (without actually calling AI, only return built prompt)
	c.JSON(http.StatusOK, gin.H{
		"system_prompt":   systemPrompt,
		"user_prompt":     userPrompt,
		"candidate_count": len(candidates),
		"candidates":      candidates,
		"prompt_variant":  req.PromptVariant,
		"config_summary": gin.H{
			"coin_source":      req.Config.CoinSource.SourceType,
			"primary_tf":       req.Config.Indicators.Klines.PrimaryTimeframe,
			"primary_count":    req.Config.Indicators.Klines.PrimaryCount,
			"selected_tfs":     req.Config.Indicators.Klines.SelectedTimeframes,
			"btc_eth_leverage": req.Config.RiskControl.BTCETHMaxLeverage,
			"altcoin_leverage": req.Config.RiskControl.AltcoinMaxLeverage,
			"max_positions":    req.Config.RiskControl.MaxPositions,
		},
		"ai_response": "Please select an AI model and click 'Run Test' to perform real AI analysis.",
		"note":        "AI model not selected or real AI call not enabled",
	})
}

// runRealAITest Execute real AI test call
func (s *Server) runRealAITest(userID, modelID, systemPrompt, userPrompt string) (string, error) {
	// Get AI model configuration
	model, err := s.store.AIModel().Get(userID, modelID)
	if err != nil {
		return "", fmt.Errorf("failed to get AI model: %w", err)
	}

	if !model.Enabled {
		return "", fmt.Errorf("AI model %s is not enabled", model.Name)
	}

	if model.APIKey == "" {
		return "", fmt.Errorf("AI model %s is missing API Key", model.Name)
	}

	// Create AI client
	var aiClient mcp.AIClient
	provider := model.Provider

	// Convert EncryptedString to string for API key
	apiKey := string(model.APIKey)
	switch provider {
	case "qwen":
		aiClient = mcp.NewQwenClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "deepseek":
		aiClient = mcp.NewDeepSeekClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "claude":
		aiClient = mcp.NewClaudeClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "kimi":
		aiClient = mcp.NewKimiClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "gemini":
		aiClient = mcp.NewGeminiClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "grok":
		aiClient = mcp.NewGrokClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "openai":
		aiClient = mcp.NewOpenAIClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	case "ollama":
		// Use OllamaClient for consistent URL handling with trader logic
		aiClient = mcp.NewOllamaClient()
		// Ollama typically doesn't need an API key, but we'll use it if provided
		ollamaAPIKey := apiKey
		if ollamaAPIKey == "" {
			ollamaAPIKey = "ollama"
		}
		aiClient.SetAPIKey(ollamaAPIKey, model.CustomAPIURL, model.CustomModelName)
	case "guardian", "guardian-ai":
		// Handle Guardian AI models (browser automation)
		// For guardian models, we use the custom API URL as the target service
		if model.CustomAPIURL != "" {
			aiClient = mcp.NewGuardianClientWithService(model.CustomAPIURL)
		} else {
			// Default to DeepSeek if no custom URL provided
			aiClient = mcp.NewGuardianClientWithTarget("deepseek", "", "", "")
		}
		// Set API key for base client validation even though Guardian uses browser automation
		aiClient.SetAPIKey("dummy-key-for-browser-automation", model.CustomAPIURL, model.CustomModelName)
	default:
		// Use generic client for unknown providers
		aiClient = mcp.NewClient()
		aiClient.SetAPIKey(apiKey, model.CustomAPIURL, model.CustomModelName)
	}

	// Call AI API
	response, err := aiClient.CallWithMessages(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI API call failed: %w", err)
	}

	return response, nil
}

// handleGenerateFullPrompt 生成完整的AI提示词（System + User + 实时数据）
// 用于手动复制到DeepSeek Web等AI平台
func (s *Server) handleGenerateFullPrompt(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		TraderID string `json:"trader_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters: trader_id is required")
		return
	}

	// 获取交易员配置
	traderConfig, err := s.store.Trader().GetFullConfig(userID, req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found or no access permission"})
		return
	}

	// 获取交易员实例
	trader, err := s.traderManager.GetTrader(req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader instance not found"})
		return
	}

	// 使用交易员实例生成包含真实数据的完整Prompt
	systemPrompt, userPrompt, err := trader.GenerateFullPrompt("balanced")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate real-time prompt: " + err.Error()})
		return
	}

	// 返回真实的完整Prompt
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"trader_id":     req.TraderID,
		"trader_name":   traderConfig.Trader.Name,
		"system_prompt": systemPrompt,
		"user_prompt":   userPrompt,
		"message":       "Real-time prompt generated successfully",
	})
}

// filterAndFormatDecisions 对从AI返回的决策进行过滤和格式化处理
func filterAndFormatDecisions(decisions []kernel.Decision) []kernel.Decision {
	// 如果决策数组为空，可能是需要从AI响应文本中提取的情况
	// 尝试从AI响应文本中提取决策
	if len(decisions) == 0 {
		// 这种情况会在解析JSON失败时发生，我们将在调用处处理
		return decisions
	}

	logger.Infof("🔍 开始过滤和格式化 %d 个决策", len(decisions))
	filtered := make([]kernel.Decision, 0, len(decisions))

	for i, originalDecision := range decisions {
		logger.Debugf("📋 决策 %d 处理前: %+v", i+1, originalDecision)
		decision := originalDecision // 创建副本以进行修改

		// 基础验证
		if decision.Symbol == "" || decision.Action == "" {
			logger.Warnf("Skipping invalid decision: symbol=%s, action=%s", decision.Symbol, decision.Action)
			continue
		}

		// 规范化符号名称（去除空格、转换大小写等）
		oldSymbol := decision.Symbol
		decision.Symbol = strings.TrimSpace(strings.ToUpper(decision.Symbol))
		if oldSymbol != decision.Symbol {
			logger.Infof("🔄 符号标准化: %s -> %s", oldSymbol, decision.Symbol)
		}

		// 验证并限制数值范围
		originalValues := map[string]interface{}{
			"PositionSizeUSD": decision.PositionSizeUSD,
			"Leverage":        decision.Leverage,
			"Confidence":      decision.Confidence,
			"StopLoss":        decision.StopLoss,
			"TakeProfit":      decision.TakeProfit,
		}

		if decision.PositionSizeUSD < 0 {
			decision.PositionSizeUSD = 0
			logger.Infof("🔧 修正 PositionSizeUSD: %.2f -> 0", originalValues["PositionSizeUSD"])
		}
		if decision.PositionSizeUSD > 1000000 { // 设置最大仓位限制为100万美元
			decision.PositionSizeUSD = 1000000
			logger.Infof("🔧 修正 PositionSizeUSD: %.2f -> 1000000", originalValues["PositionSizeUSD"])
		}
		if decision.Leverage < 0 {
			decision.Leverage = 0
			logger.Infof("🔧 修正 Leverage: %d -> 0", originalValues["Leverage"])
		}
		if decision.Leverage > 100 { // 设置最大杠杆限制
			decision.Leverage = 100
			logger.Infof("🔧 修正 Leverage: %d -> 100", originalValues["Leverage"])
		}
		if decision.Confidence < 0 {
			decision.Confidence = 0
			logger.Infof("🔧 修正 Confidence: %d -> 0", originalValues["Confidence"])
		}
		if decision.Confidence > 100 {
			decision.Confidence = 100
			logger.Infof("🔧 修正 Confidence: %d -> 100", originalValues["Confidence"])
		}

		// 验证价格相关字段
		if decision.StopLoss < 0 {
			oldValue := decision.StopLoss
			decision.StopLoss = 0
			if oldValue != decision.StopLoss {
				logger.Infof("🔧 修正 StopLoss: %.4f -> 0", oldValue)
			}
		}
		if decision.TakeProfit < 0 {
			oldValue := decision.TakeProfit
			decision.TakeProfit = 0
			if oldValue != decision.TakeProfit {
				logger.Infof("🔧 修正 TakeProfit: %.4f -> 0", oldValue)
			}
		}

		// 验证动作类型是否合法
		validActions := map[string]bool{
			"open_long": true, "open_short": true,
			"close_long": true, "close_short": true,
			"hold": true, "wait": true,
			"update_stop_loss": true, "update_take_profit": true,
			"partial_close": true, "trailing_stop": true,
			"dynamic_take_profit": true, "oco_order": true,
			"bracket_order": true, "add_to_position": true,
		}
		if !validActions[decision.Action] {
			logger.Warnf("Skipping decision with invalid action: %s", decision.Action)
			continue
		}

		logger.Debugf("✅ 决策 %d 处理后: %+v", i+1, decision)
		// 添加到过滤后的决策列表
		filtered = append(filtered, decision)
	}

	logger.Infof("✅ 过滤和格式化完成，共保留 %d 个有效决策", len(filtered))
	return filtered
}

// extractAndFilterDecisionsFromText 从AI响应文本中提取并过滤决策
func extractAndFilterDecisionsFromText(aiResponseText string) ([]kernel.Decision, error) {
	// 移除不可见字符
	s := strings.TrimSpace(aiResponseText)
	s = regexp.MustCompile("[\u200B\u200C\u200D\uFEFF]").ReplaceAllString(s, "")

	// 尝试从 <decision> 标签中提取JSON
	var jsonPart string
	reDecisionTag := regexp.MustCompile(`(?s)<decision>(.*?)</decision>`)
	if match := reDecisionTag.FindStringSubmatch(s); match != nil && len(match) > 1 {
		jsonPart = strings.TrimSpace(match[1])
		logger.Infof("✓ Extracted JSON using <decision> tag")
	} else {
		jsonPart = s
		logger.Infof("⚠️  <decision> tag not found, searching JSON in full text")
	}

	// 尝试从 ```json 代码块中提取
	reJSONFence := regexp.MustCompile(`(?is)` + "```json\\s*(\\[\\s*\\{.*?\\}\\s*\\])\\s*```")
	if match := reJSONFence.FindStringSubmatch(jsonPart); match != nil && len(match) > 1 {
		jsonContent := strings.TrimSpace(match[1])
		// 修复常见的字符问题
		jsonContent = fixMissingQuotesInJSON(jsonContent)

		var decisions []kernel.Decision
		if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
			return nil, fmt.Errorf("failed to parse JSON from fenced code block: %w", err)
		}
		return filterAndFormatDecisions(decisions), nil
	}

	// 尝试直接查找JSON数组
	reJSONArray := regexp.MustCompile(`(?is)\[\s*\{.*?\}\s*\]`)
	jsonContent := strings.TrimSpace(reJSONArray.FindString(jsonPart))
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	// 修复常见的字符问题
	jsonContent = fixMissingQuotesInJSON(jsonContent)

	var decisions []kernel.Decision
	if err := json.Unmarshal([]byte(jsonContent), &decisions); err != nil {
		return nil, fmt.Errorf("failed to parse extracted JSON: %w", err)
	}

	return filterAndFormatDecisions(decisions), nil
}

// fixMissingQuotesInJSON 修复JSON中的常见字符问题
func fixMissingQuotesInJSON(jsonStr string) string {
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "\u2018", "'")
	jsonStr = strings.ReplaceAll(jsonStr, "\u2019", "'")

	jsonStr = strings.ReplaceAll(jsonStr, "［", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "］", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "｛", "{")
	jsonStr = strings.ReplaceAll(jsonStr, "｝", "}")
	jsonStr = strings.ReplaceAll(jsonStr, "：", ":")
	jsonStr = strings.ReplaceAll(jsonStr, "，", ",")

	jsonStr = strings.ReplaceAll(jsonStr, "【", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "】", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "〔", "[")
	jsonStr = strings.ReplaceAll(jsonStr, "〕", "]")
	jsonStr = strings.ReplaceAll(jsonStr, "、", ",")

	jsonStr = strings.ReplaceAll(jsonStr, "　", " ")

	return jsonStr
}

// handleSubmitAIDecision 提交AI决策JSON（从DeepSeek Web等平台复制回来的）
func (s *Server) handleSubmitAIDecision(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		TraderID     string            `json:"trader_id" binding:"required"`
		DecisionJSON string            `json:"decision_json"`
		Decisions    []kernel.Decision `json:"decisions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	// 验证交易员权限
	_, err := s.store.Trader().GetFullConfig(userID, req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found or no access permission"})
		return
	}

	// 如果没有直接提供decisions，则从JSON字符串解析
	var decisions []kernel.Decision

	if len(req.Decisions) > 0 {
		decisions = req.Decisions
		// 对直接提供的决策进行过滤
		decisions = filterAndFormatDecisions(decisions)
	} else if req.DecisionJSON != "" {
		// 首先尝试直接解析JSON数组
		if err := json.Unmarshal([]byte(req.DecisionJSON), &decisions); err != nil {
			// 如果直接解析失败，尝试从AI响应文本中提取决策
			extractedDecisions, extractErr := extractAndFilterDecisionsFromText(req.DecisionJSON)
			if extractErr != nil {
				logger.Warnf("Failed to parse decision JSON directly: %v, attempting to extract from AI response text: %v", err, extractErr)
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Failed to parse decision JSON or extract from AI response",
					"details": extractErr.Error(),
				})
				return
			}
			decisions = extractedDecisions
		} else {
			// 如果直接解析成功，应用过滤
			decisions = filterAndFormatDecisions(decisions)
		}
	}

	if len(decisions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No decisions provided or extracted"})
		return
	}

	// 获取交易员实例
	trader, err := s.traderManager.GetTrader(req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader instance not found"})
		return
	}

	// 使用已过滤的决策
	filteredDecisions := decisions

	// 执行决策（复用现有的执行逻辑）
	results := make([]map[string]interface{}, 0, len(filteredDecisions))
	successCount := 0
	failCount := 0

	for i, decision := range filteredDecisions {
		result := map[string]interface{}{
			"index":   i + 1,
			"symbol":  decision.Symbol,
			"action":  decision.Action,
			"success": false,
		}

		err := trader.ExecuteDecision(&decision)
		if err != nil {
			result["error"] = err.Error()
			failCount++
		} else {
			result["success"] = true
			result["message"] = "Executed successfully"
			successCount++
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"total":         len(decisions),
		"success_count": successCount,
		"fail_count":    failCount,
		"results":       results,
		"message":       fmt.Sprintf("Executed %d/%d decisions successfully", successCount, len(decisions)),
	})
}

// handleGetScanData 获取手动扫描过程中的数据（在AI调用前截断）
func (s *Server) handleGetScanData(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		TraderID string `json:"trader_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters: trader_id is required")
		return
	}

	// 获取交易员配置
	traderConfig, err := s.store.Trader().GetFullConfig(userID, req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader not found or no access permission"})
		return
	}

	// 获取交易员实例
	autoTrader, err := s.traderManager.GetTrader(req.TraderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trader instance not found"})
		return
	}

	// 使用交易员的内部方法获取扫描数据（但不执行AI决策）
	// 这里我们需要调用AutoTrader内部的buildTradingContext方法
	// 由于我们无法直接访问私有方法，我们将通过现有API获取相关信息

	// 获取当前状态信息
	status := autoTrader.GetStatus()

	// 获取账户信息（使用AutoTrader的GetAccountInfo方法）
	accountInfo, err := autoTrader.GetAccountInfo()
	if err != nil {
		accountInfo = nil
	}

	// 获取持仓信息（使用AutoTrader的GetPositions方法）
	positions, err := autoTrader.GetPositions()
	if err != nil {
		positions = nil
	}

	// 获取策略配置
	var strategyConfig *store.StrategyConfig
	if traderConfig.Trader.StrategyID != "" {
		strategy, err := s.store.Strategy().Get(userID, traderConfig.Trader.StrategyID)
		if err == nil && strategy != nil {
			var config store.StrategyConfig
			json.Unmarshal([]byte(strategy.Config), &config)
			strategyConfig = &config
		}
	}

	// 返回获取到的数据
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"trader_id":       req.TraderID,
		"trader_name":     traderConfig.Trader.Name,
		"status":          status,
		"account_info":    accountInfo,
		"positions":       positions,
		"strategy_config": strategyConfig,
		"message":         "Scan data collected successfully (before AI decision)",
		"timestamp":       time.Now().Unix(),
	})
}
