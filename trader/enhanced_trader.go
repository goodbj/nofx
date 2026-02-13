package trader

import (
	"context"
	"fmt"
	"nofx/logger"
	"strconv"
	"time"
)

// EnhancedFuturesTrader extends FuturesTrader with improved error handling and validation
type EnhancedFuturesTrader struct {
	*FuturesTrader
	validator *OrderValidator
}

// OrderValidator handles order validation
type OrderValidator struct {
	symbolCache map[string]*SymbolInfo
}

// NewEnhancedFuturesTrader creates a trader with enhanced capabilities
func NewEnhancedFuturesTrader(original *FuturesTrader) *EnhancedFuturesTrader {
	return &EnhancedFuturesTrader{
		FuturesTrader: original,
		validator: &OrderValidator{
			symbolCache: make(map[string]*SymbolInfo),
		},
	}
}

// EnhancedOpenLong with improved error handling and validation
func (et *EnhancedFuturesTrader) EnhancedOpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 1. Validate parameters first
	params := OrderParams{
		Symbol:   symbol,
		Side:     "BUY",
		Type:     "MARKET",
		Quantity: quantity,
		Leverage: &leverage,
	}

	symbolInfo, err := et.getSymbolInfo(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol info: %w", err)
	}

	if err := et.validator.ValidateOrder(params, symbolInfo); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// 2. Format quantity with proper precision
	formattedQty, err := et.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to format quantity: %w", err)
	}

	// 3. Execute with retry logic
	result, err := et.executeWithRetry(func() (map[string]interface{}, error) {
		return et.FuturesTrader.OpenLong(symbol, quantity, leverage)
	})

	if err != nil {
		// 4. Handle specific error types
		return nil, HandleBinanceError(err)
	}

	// 5. Add enhanced logging
	logger.Infof("✅ Successfully opened long position for %s: quantity=%s, leverage=%d",
		symbol, formattedQty, leverage)

	return result, nil
}

// EnhancedOpenShort with improved error handling and validation
func (et *EnhancedFuturesTrader) EnhancedOpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// 1. Validate parameters
	params := OrderParams{
		Symbol:   symbol,
		Side:     "SELL",
		Type:     "MARKET",
		Quantity: quantity,
		Leverage: &leverage,
	}

	symbolInfo, err := et.getSymbolInfo(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol info: %w", err)
	}

	if err := et.validator.ValidateOrder(params, symbolInfo); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// 2. Format quantity with proper precision
	formattedQty, err := et.FormatQuantity(symbol, quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to format quantity: %w", err)
	}

	// 3. Execute with retry logic
	result, err := et.executeWithRetry(func() (map[string]interface{}, error) {
		return et.FuturesTrader.OpenShort(symbol, quantity, leverage)
	})

	if err != nil {
		// 4. Handle specific error types
		return nil, HandleBinanceError(err)
	}

	// 5. Add enhanced logging
	logger.Infof("✅ Successfully opened short position for %s: quantity=%s, leverage=%d",
		symbol, formattedQty, leverage)

	return result, nil
}

// executeWithRetry executes a function with retry logic
func (et *EnhancedFuturesTrader) executeWithRetry(operation func() (map[string]interface{}, error)) (map[string]interface{}, error) {
	config := DefaultRetryConfig()

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !config.ShouldRetry(err) {
			logger.Warnf("❌ Non-retryable error on attempt %d: %v", attempt+1, err)
			break
		}

		if attempt < config.MaxRetries {
			delay := calculateBackoff(attempt, config.BaseDelay, config.MaxDelay)
			logger.Infof("⏳ Retryable error encountered, retrying in %v (attempt %d/%d): %v",
				delay, attempt+1, config.MaxRetries+1, err)

			// Use context for cancellation support
			ctx, cancel := context.WithTimeout(context.Background(), delay)
			<-ctx.Done()
			cancel()
		}
	}

	return nil, fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// getSymbolInfo retrieves symbol information with caching
func (et *EnhancedFuturesTrader) getSymbolInfo(symbol string) (*SymbolInfo, error) {
	// Check cache first
	if info, exists := et.validator.symbolCache[symbol]; exists {
		// Cache invalidation - refresh if older than 1 hour
		if time.Since(info.Timestamp) < time.Hour {
			return info, nil
		}
	}

	// Fetch fresh data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exchangeInfo, err := et.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange info: %w", err)
	}

	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == symbol {
			info := &SymbolInfo{
				Symbol:            symbol,
				PricePrecision:    s.QuotePrecision,
				QuantityPrecision: s.BaseAssetPrecision,
				Timestamp:         time.Now(),
			}

			// Parse filters
			for _, filter := range s.Filters {
				switch filter["filterType"] {
				case "LOT_SIZE":
					if minQtyStr, ok := filter["minQty"].(string); ok {
						info.MinQty, _ = strconv.ParseFloat(minQtyStr, 64)
					}
					if maxQtyStr, ok := filter["maxQty"].(string); ok {
						info.MaxQty, _ = strconv.ParseFloat(maxQtyStr, 64)
					}
					if stepSizeStr, ok := filter["stepSize"].(string); ok {
						info.StepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
					}
				case "PRICE_FILTER":
					if minPriceStr, ok := filter["minPrice"].(string); ok {
						info.MinPrice, _ = strconv.ParseFloat(minPriceStr, 64)
					}
					if maxPriceStr, ok := filter["maxPrice"].(string); ok {
						info.MaxPrice, _ = strconv.ParseFloat(maxPriceStr, 64)
					}
					if tickSizeStr, ok := filter["tickSize"].(string); ok {
						info.TickSize, _ = strconv.ParseFloat(tickSizeStr, 64)
					}
				case "MIN_NOTIONAL":
					if notionalStr, ok := filter["notional"].(string); ok {
						info.MinNotional, _ = strconv.ParseFloat(notionalStr, 64)
					}
				}
			}

			// Cache the information
			et.validator.symbolCache[symbol] = info
			return info, nil
		}
	}

	return nil, fmt.Errorf("symbol %s not found in exchange info", symbol)
}

// ValidateOrder is a convenience method for external use
func (et *EnhancedFuturesTrader) ValidateOrder(params OrderParams) error {
	symbolInfo, err := et.getSymbolInfo(params.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get symbol info: %w", err)
	}

	return et.validator.ValidateOrder(params, symbolInfo)
}

// OrderValidator methods
func (ov *OrderValidator) ValidateOrder(params OrderParams, symbolInfo *SymbolInfo) error {
	return validateParams(params, ov.buildValidationRules(symbolInfo))
}

func (ov *OrderValidator) buildValidationRules(symbolInfo *SymbolInfo) []ValidationRule {
	rules := []ValidationRule{
		{
			Field:    "Symbol",
			Required: true,
			Validators: []func(interface{}) error{
				validateSymbolFormat,
			},
		},
		{
			Field:    "Side",
			Required: true,
			Validators: []func(interface{}) error{
				validateOrderSide,
			},
		},
		{
			Field:    "Type",
			Required: true,
			Validators: []func(interface{}) error{
				validateOrderType,
			},
		},
		{
			Field:    "Quantity",
			Required: true,
			MinValue: &symbolInfo.MinQty,
			MaxValue: &symbolInfo.MaxQty,
			Validators: []func(interface{}) error{
				validateStepSizeAlignment(symbolInfo.StepSize),
				validateMinNotional(symbolInfo.MinNotional),
			},
		},
	}

	return rules
}
