package binance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response Binance API response wrapper
type Response struct {
	Data       interface{}   `json:"data,omitempty"`
	RateLimits []*RateLimit  `json:"rate_limits,omitempty"`
	Metadata   *ResponseMeta `json:"metadata,omitempty"`
	Error      *BinanceError `json:"error,omitempty"`
}

// ResponseMeta contains metadata about the response
type ResponseMeta struct {
	Timestamp    time.Time           `json:"timestamp"`
	Latency      int64               `json:"latency_ms"` // Response time in milliseconds
	RequestID    string              `json:"request_id,omitempty"`
	Endpoint     string              `json:"endpoint"`
	StatusCode   int                 `json:"status_code"`
	Headers      map[string][]string `json:"headers,omitempty"`
	RetryCount   int                 `json:"retry_count,omitempty"`
	RetryHistory []RetryInfo         `json:"retry_history,omitempty"`
}

// RateLimit represents Binance API rate limit information
type RateLimit struct {
	RateLimitType string `json:"rate_limit_type"` // REQUEST_WEIGHT, ORDERS, RAW_REQUESTS
	Interval      string `json:"interval"`        // MINUTE, SECOND
	IntervalNum   int    `json:"interval_num"`
	Limit         int    `json:"limit"`
	Count         int    `json:"count,omitempty"` // Current usage
}

// RetryInfo contains information about retry attempts
type RetryInfo struct {
	Attempt   int           `json:"attempt"`
	Timestamp time.Time     `json:"timestamp"`
	Delay     time.Duration `json:"delay"`
	Error     string        `json:"error,omitempty"`
}

// NewResponse creates a new response wrapper
func NewResponse() *Response {
	return &Response{
		Metadata: &ResponseMeta{
			Timestamp: time.Now(),
		},
	}
}

// WithData sets response data
func (r *Response) WithData(data interface{}) *Response {
	r.Data = data
	return r
}

// WithError sets response error
func (r *Response) WithError(err *BinanceError) *Response {
	r.Error = err
	if err != nil && r.Metadata != nil {
		r.Metadata.StatusCode = err.StatusCode
	}
	return r
}

// WithRateLimits sets rate limit information
func (r *Response) WithRateLimits(limits []*RateLimit) *Response {
	r.RateLimits = limits
	return r
}

// WithMetadata sets response metadata
func (r *Response) WithMetadata(meta *ResponseMeta) *Response {
	r.Metadata = meta
	return r
}

// WithLatency sets response latency
func (r *Response) WithLatency(latency time.Duration) *Response {
	if r.Metadata != nil {
		r.Metadata.Latency = latency.Milliseconds()
	}
	return r
}

// WithRequestID sets request ID
func (r *Response) WithRequestID(requestID string) *Response {
	if r.Metadata != nil {
		r.Metadata.RequestID = requestID
	}
	return r
}

// WithEndpoint sets endpoint information
func (r *Response) WithEndpoint(endpoint string) *Response {
	if r.Metadata != nil {
		r.Metadata.Endpoint = endpoint
	}
	return r
}

// WithStatusCode sets HTTP status code
func (r *Response) WithStatusCode(code int) *Response {
	if r.Metadata != nil {
		r.Metadata.StatusCode = code
	}
	return r
}

// WithHeaders sets response headers
func (r *Response) WithHeaders(headers map[string][]string) *Response {
	if r.Metadata != nil {
		r.Metadata.Headers = headers
	}
	return r
}

// WithRetryInfo sets retry information
func (r *Response) WithRetryInfo(count int, history []RetryInfo) *Response {
	if r.Metadata != nil {
		r.Metadata.RetryCount = count
		r.Metadata.RetryHistory = history
	}
	return r
}

// IsSuccess returns whether the response was successful
func (r *Response) IsSuccess() bool {
	return r.Error == nil && r.Metadata != nil && r.Metadata.StatusCode >= 200 && r.Metadata.StatusCode < 300
}

// IsError returns whether the response contains an error
func (r *Response) IsError() bool {
	return r.Error != nil
}

// GetError returns the error if present
func (r *Response) GetError() *BinanceError {
	return r.Error
}

// GetData returns the response data
func (r *Response) GetData() interface{} {
	return r.Data
}

// GetRateLimits returns rate limit information
func (r *Response) GetRateLimits() []*RateLimit {
	return r.RateLimits
}

// GetMetadata returns response metadata
func (r *Response) GetMetadata() *ResponseMeta {
	return r.Metadata
}

// ToJSON converts response to JSON
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromJSON parses response from JSON
func (r *Response) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// ExtractRateLimitsFromHeaders extracts rate limit information from HTTP headers
func ExtractRateLimitsFromHeaders(headers http.Header) []*RateLimit {
	var limits []*RateLimit

	// Extract rate limit headers
	for key, values := range headers {
		key = http.CanonicalHeaderKey(key)

		switch key {
		case "X-Mbx-Used-Weight-1m":
			if len(values) > 0 {
				limits = append(limits, &RateLimit{
					RateLimitType: "REQUEST_WEIGHT",
					Interval:      "MINUTE",
					IntervalNum:   1,
					Count:         parseIntHeader(values[0], 0),
				})
			}
		case "X-Mbx-Order-Count-10s":
			if len(values) > 0 {
				limits = append(limits, &RateLimit{
					RateLimitType: "ORDERS",
					Interval:      "SECOND",
					IntervalNum:   10,
					Count:         parseIntHeader(values[0], 0),
				})
			}
		case "X-Mbx-Order-Count-1m":
			if len(values) > 0 {
				limits = append(limits, &RateLimit{
					RateLimitType: "ORDERS",
					Interval:      "MINUTE",
					IntervalNum:   1,
					Count:         parseIntHeader(values[0], 0),
				})
			}
		case "X-Mbx-Order-Count-1h":
			if len(values) > 0 {
				limits = append(limits, &RateLimit{
					RateLimitType: "ORDERS",
					Interval:      "HOUR",
					IntervalNum:   1,
					Count:         parseIntHeader(values[0], 0),
				})
			}
		case "X-Mbx-Order-Count-1d":
			if len(values) > 0 {
				limits = append(limits, &RateLimit{
					RateLimitType: "ORDERS",
					Interval:      "DAY",
					IntervalNum:   1,
					Count:         parseIntHeader(values[0], 0),
				})
			}
		}
	}

	return limits
}

// CreateSuccessResponse creates a successful response
func CreateSuccessResponse(data interface{}, headers http.Header, latency time.Duration) *Response {
	response := NewResponse().
		WithData(data).
		WithRateLimits(ExtractRateLimitsFromHeaders(headers)).
		WithLatency(latency)

	if response.Metadata != nil {
		response.Metadata.StatusCode = http.StatusOK
		response.Metadata.Headers = make(map[string][]string)
		for k, v := range headers {
			response.Metadata.Headers[k] = v
		}
	}

	return response
}

// CreateErrorResponse creates an error response
func CreateErrorResponse(err *BinanceError, headers http.Header, latency time.Duration) *Response {
	response := NewResponse().
		WithError(err).
		WithRateLimits(ExtractRateLimitsFromHeaders(headers)).
		WithLatency(latency)

	if response.Metadata != nil {
		response.Metadata.Headers = make(map[string][]string)
		for k, v := range headers {
			response.Metadata.Headers[k] = v
		}
	}

	return response
}

// Helper function for parsing integers from strings
func parseIntHeader(s string, defaultValue int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
