// Package constants provides shared constants used throughout the GPTComet application.
//
// This package centralizes all magic strings, numbers, and configuration values
// to ensure consistency and ease of maintenance. It includes constants for:
//
//   - HTTP client configuration (timeouts, connection pooling)
//   - Retry logic parameters (delays, backoff strategy)
//   - API endpoints and paths
//   - Configuration keys
//   - HTTP headers and content types
//   - HTTP status codes
//   - Server-Sent Events (SSE) constants
//   - Proxy schemes
//   - Security parameters (API key masking)
//
// Usage:
//
//	transport := &http.Transport{
//	    MaxIdleConns:       constants.MaxIdleConns,
//	    IdleConnTimeout:    constants.IdleConnTimeout,
//	    DisableCompression: constants.DisableCompression,
//	}
//
//	delay := constants.BaseRetryDelay * time.Duration(attempt)
//
//	if resp.StatusCode != constants.StatusOK {
//	    return errors.New("request failed")
//	}
//
// By centralizing these values, the codebase becomes more maintainable and
// changes can be made in a single location rather than scattered throughout
// multiple files.
package constants

import "time"

// HTTP Client Configuration
const (
	// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts
	MaxIdleConns = 100

	// IdleConnTimeout is the maximum amount of time an idle connection will remain idle before closing itself
	IdleConnTimeout = 90 * time.Second

	// DisableCompression prevents the Transport from requesting compression with an "Accept-Encoding: gzip" request header
	DisableCompression = true
)

// Retry Configuration
const (
	// BaseRetryDelay is the base delay for exponential backoff retry logic
	BaseRetryDelay = 500 * time.Millisecond

	// DefaultMaxRetries is the default maximum number of retry attempts
	DefaultMaxRetries = 3

	// MaxJitterPercent is the maximum jitter percentage added to retry delay (50%)
	MaxJitterPercent = 0.5
)

// API Endpoints
const (
	// ChatCompletionsEndpoint is the standard OpenAI-compatible chat completions endpoint
	ChatCompletionsEndpoint = "/chat/completions"

	// EmbeddingsEndpoint is the standard OpenAI-compatible embeddings endpoint
	EmbeddingsEndpoint = "/embeddings"
)

// Configuration Keys
const (
	// ConfigKeyAPIBase is the configuration key for API base URL
	ConfigKeyAPIBase = "api_base"

	// ConfigKeyAPIKey is the configuration key for API key
	ConfigKeyAPIKey = "api_key"

	// ConfigKeyModel is the configuration key for model name
	ConfigKeyModel = "model"

	// ConfigKeyMaxTokens is the configuration key for max tokens
	ConfigKeyMaxTokens = "max_tokens"

	// ConfigKeyTemperature is the configuration key for temperature
	ConfigKeyTemperature = "temperature"

	// ConfigKeyTopP is the configuration key for top_p
	ConfigKeyTopP = "top_p"

	// ConfigKeyFrequencyPenalty is the configuration key for frequency_penalty
	ConfigKeyFrequencyPenalty = "frequency_penalty"

	// ConfigKeyRetries is the configuration key for retry attempts
	ConfigKeyRetries = "retries"

	// ConfigKeyProxy is the configuration key for proxy URL
	ConfigKeyProxy = "proxy"

	// ConfigKeyProvider is the configuration key for LLM provider
	ConfigKeyProvider = "provider"

	// ConfigKeyAnswerPath is the configuration key for answer path
	ConfigKeyAnswerPath = "answer_path"

	// ConfigKeyCompletionPath is the configuration key for completion path
	ConfigKeyCompletionPath = "completion_path"
)

// HTTP Headers
const (
	// HeaderContentType is the Content-Type header name
	HeaderContentType = "Content-Type"

	// HeaderAuthorization is the Authorization header name
	HeaderAuthorization = "Authorization"

	// HeaderAPIKey is a common API key header name
	HeaderAPIKey = "api-key"

	// HeaderProxyAuthorization is the Proxy-Authorization header name
	HeaderProxyAuthorization = "Proxy-Authorization"

	// ContentTypeJSON is the JSON content type
	ContentTypeJSON = "application/json"

	// ContentTypeStream is the SSE stream content type
	ContentTypeStream = "text/event-stream"
)

// HTTP Status Codes (commonly used)
const (
	// StatusOK indicates success
	StatusOK = 200

	// StatusBadRequest indicates invalid request
	StatusBadRequest = 400

	// StatusUnauthorized indicates authentication failure
	StatusUnauthorized = 401

	// StatusForbidden indicates authorization failure
	StatusForbidden = 403

	// StatusNotFound indicates resource not found
	StatusNotFound = 404

	// StatusTooManyRequests indicates rate limiting
	StatusTooManyRequests = 429

	// StatusInternalServerError indicates server error
	StatusInternalServerError = 500

	// StatusBadGateway indicates gateway error
	StatusBadGateway = 502

	// StatusServiceUnavailable indicates service unavailable
	StatusServiceUnavailable = 503

	// StatusGatewayTimeout indicates gateway timeout
	StatusGatewayTimeout = 504
)

// SSE (Server-Sent Events) Constants
const (
	// SSEDataPrefix is the prefix for SSE data lines
	SSEDataPrefix = "data: "

	// SSEDone is the message indicating stream completion
	SSEDone = "[DONE]"
)

// Default Values
const (
	// DefaultOllamaAPIBase is the default Ollama API base URL
	DefaultOllamaAPIBase = "http://localhost:11434/api"

	// DefaultTimeout is the default HTTP request timeout
	DefaultTimeout = 60 * time.Second
)

// Common Strings
const (
	// BearerPrefix is the Bearer token prefix for authorization headers
	BearerPrefix = "Bearer "

	// BasicPrefix is the Basic auth prefix for authorization headers
	BasicPrefix = "Basic "
)

// Proxy Schemes
const (
	// ProxySchemeHTTP indicates HTTP proxy
	ProxySchemeHTTP = "http"

	// ProxySchemeHTTPS indicates HTTPS proxy
	ProxySchemeHTTPS = "https"

	// ProxySchemeSocks5 indicates SOCKS5 proxy
	ProxySchemeSocks5 = "socks5"
)

// OpenRouter specific constants
const (
	// OpenRouterRefererURL is the referer URL for OpenRouter API requests
	OpenRouterRefererURL = "https://github.com/belingud/gptcomet"
)

// API Key Security
const (
	// APIKeyMaskLength is the number of characters to show when masking API keys
	APIKeyMaskLength = 4

	// APIKeyMinLength is the minimum length for a valid API key
	APIKeyMinLength = 8
)
