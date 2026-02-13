package trader

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BinanceAPIError represents a standardized Binance API error
type BinanceAPIError struct {
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	StatusCode int    `json:"-"`
	RequestID  string `json:"-"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

func (e *BinanceAPIError) Error() string {
	return fmt.Sprintf("Binance API Error %d: %s", e.Code, e.Message)
}

// IsRetryable checks if the error is retryable
func (e *BinanceAPIError) IsRetryable() bool {
	// Network errors and server errors are retryable
	retryableCodes := []int{
		-1003, // Too many requests
		-1021, // Timestamp for this request is outside of the recvWindow
		-1022, // Signature for this request is not valid
	}

	for _, code := range retryableCodes {
		if e.Code == code {
			return true
		}
	}

	// Server errors (5xx) are retryable
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsFatal checks if the error is fatal and should not be retried
func (e *BinanceAPIError) IsFatal() bool {
	// These errors indicate client-side issues that won't be fixed by retrying
	fatalCodes := []int{
		-1100, // Illegal characters found in parameter
		-1101, // Too many parameters sent
		-1102, // Parameter is empty
		-1103, // Parameter is not required
		-1111, // Precision is over the maximum defined for this asset
		-1112, // No orders on book for symbol
		-2010, // Account has insufficient balance
		-2011, // Account has insufficient balance for requested action
		-2013, // Order does not exist
		-2014, // API key does not exist
		-2015, // Invalid API key or IP
	}

	for _, code := range fatalCodes {
		if e.Code == code {
			return true
		}
	}

	// Client errors (4xx except 429) are fatal
	return e.StatusCode >= 400 && e.StatusCode < 500 && e.StatusCode != 429
}

// Error handlers for specific error types
func handlePrecisionError(err *BinanceAPIError) error {
	return fmt.Errorf("precision error: %s - please check quantity and price formatting", err.Message)
}

func handleAPIKeyError(err *BinanceAPIError) error {
	return fmt.Errorf("API key error: %s - please verify your API credentials", err.Message)
}

func handleTimestampError(err *BinanceAPIError) error {
	return fmt.Errorf("timestamp error: %s - system time may be out of sync", err.Message)
}

func handleInsufficientBalanceError(err *BinanceAPIError) error {
	return fmt.Errorf("insufficient balance: %s - please check your account balance", err.Message)
}

func handleInvalidOrderError(err *BinanceAPIError) error {
	return fmt.Errorf("invalid order: %s - please check order parameters", err.Message)
}

func handleGenericError(err *BinanceAPIError) error {
	if err.IsFatal() {
		return fmt.Errorf("fatal API error %d: %s", err.Code, err.Message)
	}
	return fmt.Errorf("API error %d: %s", err.Code, err.Message)
}

// HandleBinanceError processes Binance API errors with appropriate handling
func HandleBinanceError(err error) error {
	if binanceErr, ok := err.(*BinanceAPIError); ok {
		switch binanceErr.Code {
		case -1111:
			return handlePrecisionError(binanceErr)
		case -2014, -2015:
			return handleAPIKeyError(binanceErr)
		case -1021, -1022:
			return handleTimestampError(binanceErr)
		case -2010, -2011:
			return handleInsufficientBalanceError(binanceErr)
		case -1100, -1101, -1102, -1103, -1112, -2013:
			return handleInvalidOrderError(binanceErr)
		default:
			return handleGenericError(binanceErr)
		}
	}
	return err
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	RetryableCodes []int
	ShouldRetry    func(error) bool
}

// DefaultRetryConfig provides sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		RetryableCodes: []int{
			-1003, // Too many requests
			-1021, // Timestamp error
			-1022, // Signature error
		},
		ShouldRetry: func(err error) bool {
			if binanceErr, ok := err.(*BinanceAPIError); ok {
				return binanceErr.IsRetryable()
			}
			// Retry on network errors
			if strings.Contains(err.Error(), "connection") ||
				strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "EOF") {
				return true
			}
			return false
		},
	}
}

// calculateBackoff implements exponential backoff with jitter
func calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := baseDelay * time.Duration(1<<uint(attempt))

	// Add jitter to prevent thundering herd
	jitter := time.Duration(float64(delay) * 0.1 * (0.5 - float64(time.Now().UnixNano()%1000000000)/1000000000.0))
	delay += jitter

	// Cap at maximum delay
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// isRetryableError determines if an error should be retried
func isRetryableError(resp *http.Response, err error, retryableCodes []int) bool {
	// Network errors are retryable
	if err != nil {
		if strings.Contains(err.Error(), "connection") ||
			strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "EOF") {
			return true
		}
		return false
	}

	// Check status codes
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return true
	}

	return false
}
