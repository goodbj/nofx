package trader

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ValidationRule defines validation rules for a field
type ValidationRule struct {
	Field      string
	Required   bool
	MinValue   *float64
	MaxValue   *float64
	Validators []func(interface{}) error
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// OrderParams represents order parameters for validation
type OrderParams struct {
	Symbol      string   `json:"symbol"`
	Side        string   `json:"side"`
	Type        string   `json:"type"`
	Quantity    float64  `json:"quantity"`
	Price       *float64 `json:"price,omitempty"`
	StopPrice   *float64 `json:"stopPrice,omitempty"`
	Leverage    *int     `json:"leverage,omitempty"`
	TimeInForce string   `json:"timeInForce,omitempty"`
}

// SymbolInfo contains symbol-specific information for validation
type SymbolInfo struct {
	Symbol            string
	PricePrecision    int
	QuantityPrecision int
	MinQty            float64
	MaxQty            float64
	StepSize          float64
	MinPrice          float64
	MaxPrice          float64
	TickSize          float64
	MinNotional       float64
	Timestamp         time.Time // 添加时间戳字段用于缓存
}

// Validator functions
func validateRequired(value interface{}) error {
	if value == nil || value == "" || (reflect.TypeOf(value).Kind() == reflect.Float64 && value.(float64) == 0) {
		return fmt.Errorf("field is required")
	}
	return nil
}

func validateMinValue(min float64) func(interface{}) error {
	return func(value interface{}) error {
		if f, ok := value.(float64); ok {
			if f < min {
				return fmt.Errorf("value %.8f is less than minimum %.8f", f, min)
			}
		}
		return nil
	}
}

func validateMaxValue(max float64) func(interface{}) error {
	return func(value interface{}) error {
		if f, ok := value.(float64); ok {
			if f > max {
				return fmt.Errorf("value %.8f is greater than maximum %.8f", f, max)
			}
		}
		return nil
	}
}

func validateSymbolFormat(value interface{}) error {
	if symbol, ok := value.(string); ok {
		if len(symbol) < 3 || len(symbol) > 20 {
			return fmt.Errorf("invalid symbol format: %s", symbol)
		}
		// Check for valid characters
		for _, char := range symbol {
			if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
				return fmt.Errorf("invalid character '%c' in symbol: %s", char, symbol)
			}
		}
	}
	return nil
}

func validateOrderSide(value interface{}) error {
	if side, ok := value.(string); ok {
		validSides := map[string]bool{
			"BUY":  true,
			"SELL": true,
		}
		if !validSides[strings.ToUpper(side)] {
			return fmt.Errorf("invalid order side: %s", side)
		}
	}
	return nil
}

func validateOrderType(value interface{}) error {
	if orderType, ok := value.(string); ok {
		validTypes := map[string]bool{
			"LIMIT":                true,
			"MARKET":               true,
			"STOP":                 true,
			"STOP_MARKET":          true,
			"TAKE_PROFIT":          true,
			"TAKE_PROFIT_MARKET":   true,
			"TRAILING_STOP_MARKET": true,
		}
		if !validTypes[strings.ToUpper(orderType)] {
			return fmt.Errorf("invalid order type: %s", orderType)
		}
	}
	return nil
}

func validateStepSizeAlignment(stepSize float64) func(interface{}) error {
	return func(value interface{}) error {
		if qty, ok := value.(float64); ok {
			// Check if quantity is a multiple of stepSize
			remainder := qty - float64(int(qty/stepSize))*stepSize
			if remainder > 1e-10 { // Account for floating point precision
				alignedQty := float64(int(qty/stepSize)) * stepSize
				return fmt.Errorf("quantity %.8f is not aligned with step size %.8f, should be %.8f",
					qty, stepSize, alignedQty)
			}
		}
		return nil
	}
}

func validatePricePrecision(precision int) func(interface{}) error {
	return func(value interface{}) error {
		if price, ok := value.(float64); ok {
			// Convert to string and check decimal places
			priceStr := fmt.Sprintf("%.10f", price)
			priceStr = strings.TrimRight(priceStr, "0")
			priceStr = strings.TrimRight(priceStr, ".")

			if dotIndex := strings.Index(priceStr, "."); dotIndex != -1 {
				decimalPlaces := len(priceStr) - dotIndex - 1
				if decimalPlaces > precision {
					format := fmt.Sprintf("%%.%df", precision)
					roundedPrice := fmt.Sprintf(format, price)
					return fmt.Errorf("price %.8f exceeds maximum precision of %d decimal places, should be %s",
						price, precision, roundedPrice)
				}
			}
		}
		return nil
	}
}

func validateMinNotional(minNotional float64) func(interface{}) error {
	return func(value interface{}) error {
		if qty, ok := value.(float64); ok {
			// This is a simplified check - in reality, notional = price * quantity
			// For now, we'll assume a minimum price of 1 USDT for validation
			notional := qty * 1.0
			if notional < minNotional {
				return fmt.Errorf("order notional %.2f USDT is less than minimum required %.2f USDT",
					notional, minNotional)
			}
		}
		return nil
	}
}

// validateParams validates parameters against rules
func validateParams(params interface{}, rules []ValidationRule) error {
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("params must be a struct")
	}

	t := v.Type()

	for _, rule := range rules {
		_, found := t.FieldByName(rule.Field)
		if !found {
			if rule.Required {
				return &ValidationError{
					Field:   rule.Field,
					Value:   nil,
					Message: "required field not found",
				}
			}
			continue
		}

		fieldValue := v.FieldByName(rule.Field).Interface()

		// Required field check
		if rule.Required {
			if err := validateRequired(fieldValue); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Value:   fieldValue,
					Message: err.Error(),
				}
			}
		}

		// Skip validation for empty optional fields
		if !rule.Required && (fieldValue == nil || fieldValue == "") {
			continue
		}

		// Min/Max value checks
		if rule.MinValue != nil {
			if err := validateMinValue(*rule.MinValue)(fieldValue); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Value:   fieldValue,
					Message: err.Error(),
				}
			}
		}

		if rule.MaxValue != nil {
			if err := validateMaxValue(*rule.MaxValue)(fieldValue); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Value:   fieldValue,
					Message: err.Error(),
				}
			}
		}

		// Custom validators
		for _, validator := range rule.Validators {
			if err := validator(fieldValue); err != nil {
				return &ValidationError{
					Field:   rule.Field,
					Value:   fieldValue,
					Message: err.Error(),
				}
			}
		}
	}

	return nil
}

// ValidateOrder validates order parameters
func (t *FuturesTrader) ValidateOrder(params OrderParams, symbolInfo *SymbolInfo) error {
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

	// Add price validation for limit orders
	if strings.Contains(strings.ToUpper(params.Type), "LIMIT") || params.Price != nil {
		priceRules := ValidationRule{
			Field:    "Price",
			Required: true,
			MinValue: &symbolInfo.MinPrice,
			MaxValue: &symbolInfo.MaxPrice,
			Validators: []func(interface{}) error{
				validatePricePrecision(symbolInfo.PricePrecision),
			},
		}
		rules = append(rules, priceRules)
	}

	// Add stop price validation for stop orders
	if strings.Contains(strings.ToUpper(params.Type), "STOP") || params.StopPrice != nil {
		stopPriceRules := ValidationRule{
			Field:    "StopPrice",
			Required: true,
			MinValue: &symbolInfo.MinPrice,
			MaxValue: &symbolInfo.MaxPrice,
			Validators: []func(interface{}) error{
				validatePricePrecision(symbolInfo.PricePrecision),
			},
		}
		rules = append(rules, stopPriceRules)
	}

	return validateParams(params, rules)
}

// Helper functions
func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
