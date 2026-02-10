package binance

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrorType represents different types of Binance API errors
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeClient
	ErrorTypeUnauthorized
	ErrorTypeForbidden
	ErrorTypeNotFound
	ErrorTypeBadRequest
	ErrorTypeTooManyRequests
	ErrorTypeRateLimitBan
	ErrorTypeServer
	ErrorTypeNetwork
	ErrorTypeTimeout
	ErrorTypeInvalidSignature
	ErrorTypeInvalidTimestamp
)

// BinanceError represents a Binance API error
type BinanceError struct {
	Type       ErrorType     `json:"type"`
	Code       int           `json:"code"`
	Message    string        `json:"message"`
	StatusCode int           `json:"status_code,omitempty"`
	URL        string        `json:"url,omitempty"`
	Retryable  bool          `json:"retryable"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

// Error implements error interface
func (e *BinanceError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("Binance API error [%d]: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("Binance API error: %s", e.Message)
}

// IsRetryable returns whether the error is retryable
func (e *BinanceError) IsRetryable() bool {
	return e.Retryable
}

// GetRetryAfter returns the recommended retry delay
func (e *BinanceError) GetRetryAfter() time.Duration {
	return e.RetryAfter
}

// Error helpers for common error types
func NewUnauthorizedError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeUnauthorized,
		Code:       -2014, // Binance unauthorized code
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewForbiddenError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeForbidden,
		Code:       -2015, // Binance insufficient permission code
		Message:    message,
		StatusCode: http.StatusForbidden,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewTooManyRequestsError(message string, retryAfter time.Duration) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeTooManyRequests,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
		Retryable:  true,
		RetryAfter: retryAfter,
		Timestamp:  time.Now(),
	}
}

func NewRateLimitBanError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeRateLimitBan,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewBadRequestError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewNotFoundError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewServerError(message string, statusCode int) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeServer,
		Message:    message,
		StatusCode: statusCode,
		Retryable:  true,
		RetryAfter: 5 * time.Second,
		Timestamp:  time.Now(),
	}
}

func NewNetworkError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeNetwork,
		Message:    message,
		Retryable:  true,
		RetryAfter: 2 * time.Second,
		Timestamp:  time.Now(),
	}
}

func NewTimeoutError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeTimeout,
		Message:    message,
		Retryable:  true,
		RetryAfter: 3 * time.Second,
		Timestamp:  time.Now(),
	}
}

func NewInvalidSignatureError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeInvalidSignature,
		Code:       -1022, // Binance invalid signature code
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Retryable:  false,
		Timestamp:  time.Now(),
	}
}

func NewInvalidTimestampError(message string) *BinanceError {
	return &BinanceError{
		Type:       ErrorTypeInvalidTimestamp,
		Code:       -1021, // Binance timestamp out of window code
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Retryable:  true,
		RetryAfter: 1 * time.Second,
		Timestamp:  time.Now(),
	}
}

// ParseErrorResponse parses Binance API error response
func ParseErrorResponse(statusCode int, body []byte, url string) *BinanceError {
	// Default error based on status code
	var binanceErr *BinanceError

	switch statusCode {
	case http.StatusUnauthorized:
		binanceErr = NewUnauthorizedError("Invalid API key or signature")
	case http.StatusForbidden:
		binanceErr = NewForbiddenError("Insufficient permissions")
	case http.StatusTooManyRequests:
		binanceErr = NewTooManyRequestsError("Rate limit exceeded", 60*time.Second)
	case http.StatusBadRequest:
		binanceErr = NewBadRequestError("Bad request")
	case http.StatusNotFound:
		binanceErr = NewNotFoundError("Endpoint not found")
	case http.StatusRequestTimeout:
		binanceErr = NewTimeoutError("Request timeout")
	default:
		if statusCode >= 500 {
			binanceErr = NewServerError("Server error", statusCode)
		} else {
			binanceErr = &BinanceError{
				Type:       ErrorTypeUnknown,
				Message:    "Unknown error",
				StatusCode: statusCode,
				Retryable:  statusCode >= 500,
				Timestamp:  time.Now(),
			}
		}
	}

	binanceErr.URL = url

	// Try to parse Binance error code from response body
	bodyStr := strings.ToLower(string(body))

	// Check for common Binance error codes
	if strings.Contains(bodyStr, "-2014") || strings.Contains(bodyStr, "api-key") {
		binanceErr.Type = ErrorTypeUnauthorized
		binanceErr.Code = -2014
		binanceErr.Retryable = false
	} else if strings.Contains(bodyStr, "-2015") || strings.Contains(bodyStr, "permission") {
		binanceErr.Type = ErrorTypeForbidden
		binanceErr.Code = -2015
		binanceErr.Retryable = false
	} else if strings.Contains(bodyStr, "-1021") || strings.Contains(bodyStr, "timestamp") {
		binanceErr.Type = ErrorTypeInvalidTimestamp
		binanceErr.Code = -1021
		binanceErr.Retryable = true
		binanceErr.RetryAfter = 1 * time.Second
	} else if strings.Contains(bodyStr, "-1022") || strings.Contains(bodyStr, "signature") {
		binanceErr.Type = ErrorTypeInvalidSignature
		binanceErr.Code = -1022
		binanceErr.Retryable = false
	}

	return binanceErr
}

// IsBinanceErrorType checks if an error is of a specific Binance error type
func IsBinanceErrorType(err error, errorType ErrorType) bool {
	if binanceErr, ok := err.(*BinanceError); ok {
		return binanceErr.Type == errorType
	}
	return false
}

// GetBinanceError returns the BinanceError if the error is a BinanceError
func GetBinanceError(err error) (*BinanceError, bool) {
	if binanceErr, ok := err.(*BinanceError); ok {
		return binanceErr, true
	}
	return nil, false
}
