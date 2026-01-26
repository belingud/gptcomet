package constants

import (
	"testing"
	"time"
)

func TestHTTPClientConfiguration(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{
			name:  "MaxIdleConns is 100",
			value: MaxIdleConns,
			want:  100,
		},
		{
			name:  "IdleConnTimeout is 90 seconds",
			value: IdleConnTimeout,
			want:  90 * time.Second,
		},
		{
			name:  "DisableCompression is true",
			value: DisableCompression,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.want {
				t.Errorf("got %v, want %v", tt.value, tt.want)
			}
		})
	}
}

func TestRetryConfiguration(t *testing.T) {
	if BaseRetryDelay != 500*time.Millisecond {
		t.Errorf("BaseRetryDelay = %v, want 500ms", BaseRetryDelay)
	}

	if DefaultMaxRetries != 3 {
		t.Errorf("DefaultMaxRetries = %d, want 3", DefaultMaxRetries)
	}

	if MaxJitterPercent != 0.5 {
		t.Errorf("MaxJitterPercent = %v, want 0.5", MaxJitterPercent)
	}
}

func TestAPIEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{
			name:     "ChatCompletionsEndpoint",
			constant: ChatCompletionsEndpoint,
			want:     "/chat/completions",
		},
		{
			name:     "EmbeddingsEndpoint",
			constant: EmbeddingsEndpoint,
			want:     "/embeddings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestConfigurationKeys(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ConfigKeyAPIBase", ConfigKeyAPIBase, "api_base"},
		{"ConfigKeyAPIKey", ConfigKeyAPIKey, "api_key"},
		{"ConfigKeyModel", ConfigKeyModel, "model"},
		{"ConfigKeyMaxTokens", ConfigKeyMaxTokens, "max_tokens"},
		{"ConfigKeyTemperature", ConfigKeyTemperature, "temperature"},
		{"ConfigKeyTopP", ConfigKeyTopP, "top_p"},
		{"ConfigKeyFrequencyPenalty", ConfigKeyFrequencyPenalty, "frequency_penalty"},
		{"ConfigKeyRetries", ConfigKeyRetries, "retries"},
		{"ConfigKeyProxy", ConfigKeyProxy, "proxy"},
		{"ConfigKeyProvider", ConfigKeyProvider, "provider"},
		{"ConfigKeyAnswerPath", ConfigKeyAnswerPath, "answer_path"},
		{"ConfigKeyCompletionPath", ConfigKeyCompletionPath, "completion_path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestHTTPHeaders(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"HeaderContentType", HeaderContentType, "Content-Type"},
		{"HeaderAuthorization", HeaderAuthorization, "Authorization"},
		{"HeaderAPIKey", HeaderAPIKey, "api-key"},
		{"HeaderProxyAuthorization", HeaderProxyAuthorization, "Proxy-Authorization"},
		{"ContentTypeJSON", ContentTypeJSON, "application/json"},
		{"ContentTypeStream", ContentTypeStream, "text/event-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		want     int
	}{
		{"StatusOK", StatusOK, 200},
		{"StatusBadRequest", StatusBadRequest, 400},
		{"StatusUnauthorized", StatusUnauthorized, 401},
		{"StatusForbidden", StatusForbidden, 403},
		{"StatusNotFound", StatusNotFound, 404},
		{"StatusTooManyRequests", StatusTooManyRequests, 429},
		{"StatusInternalServerError", StatusInternalServerError, 500},
		{"StatusBadGateway", StatusBadGateway, 502},
		{"StatusServiceUnavailable", StatusServiceUnavailable, 503},
		{"StatusGatewayTimeout", StatusGatewayTimeout, 504},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestSSEConstants(t *testing.T) {
	if SSEDataPrefix != "data: " {
		t.Errorf("SSEDataPrefix = %v, want 'data: '", SSEDataPrefix)
	}

	if SSEDone != "[DONE]" {
		t.Errorf("SSEDone = %v, want '[DONE]'", SSEDone)
	}
}

func TestDefaultValues(t *testing.T) {
	if DefaultOllamaAPIBase != "http://localhost:11434/api" {
		t.Errorf("DefaultOllamaAPIBase = %v, want 'http://localhost:11434/api'", DefaultOllamaAPIBase)
	}

	if DefaultTimeout != 60*time.Second {
		t.Errorf("DefaultTimeout = %v, want 60s", DefaultTimeout)
	}
}

func TestCommonStrings(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"BearerPrefix", BearerPrefix, "Bearer "},
		{"BasicPrefix", BasicPrefix, "Basic "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestProxySchemes(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"ProxySchemeHTTP", ProxySchemeHTTP, "http"},
		{"ProxySchemeHTTPS", ProxySchemeHTTPS, "https"},
		{"ProxySchemeSocks5", ProxySchemeSocks5, "socks5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("got %v, want %v", tt.constant, tt.want)
			}
		})
	}
}

func TestOpenRouterConstants(t *testing.T) {
	want := "https://github.com/belingud/gptcomet"
	if OpenRouterRefererURL != want {
		t.Errorf("OpenRouterRefererURL = %v, want %v", OpenRouterRefererURL, want)
	}
}

func TestAPIKeySecurity(t *testing.T) {
	if APIKeyMaskLength != 4 {
		t.Errorf("APIKeyMaskLength = %d, want 4", APIKeyMaskLength)
	}

	if APIKeyMinLength != 8 {
		t.Errorf("APIKeyMinLength = %d, want 8", APIKeyMinLength)
	}
}
