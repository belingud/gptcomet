package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLM implements the LLM interface for testing
type MockLLM struct {
	makeRequestFunc       func(ctx context.Context, client *http.Client, message string, stream bool) (string, error)
	buildHeadersFunc      func() map[string]string
	buildURLFunc          func() string
	formatMessagesFunc    func(message string) (interface{}, error)
	getRequiredConfigFunc func() map[string]config.ConfigRequirement
	getUsageFunc          func(data []byte) (string, error)
	parseResponseFunc     func(response []byte) (string, error)
	name                  string
}

func (m *MockLLM) Name() string {
	return m.name
}

func (m *MockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return m.makeRequestFunc(ctx, client, message, stream)
}

func (m *MockLLM) BuildHeaders() map[string]string {
	if m.buildHeadersFunc != nil {
		return m.buildHeadersFunc()
	}
	return map[string]string{}
}

func (m *MockLLM) BuildURL() string {
	if m.buildURLFunc != nil {
		return m.buildURLFunc()
	}
	return ""
}

func (m *MockLLM) FormatMessages(message string) (interface{}, error) {
	if m.formatMessagesFunc != nil {
		return m.formatMessagesFunc(message)
	}
	return message, nil
}

func (m *MockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	if m.getRequiredConfigFunc != nil {
		return m.getRequiredConfigFunc()
	}
	return map[string]config.ConfigRequirement{}
}

func (m *MockLLM) GetUsage(data []byte) (string, error) {
	if m.getUsageFunc != nil {
		return m.getUsageFunc(data)
	}
	return "", nil
}

func (m *MockLLM) ParseResponse(response []byte) (string, error) {
	if m.parseResponseFunc != nil {
		return m.parseResponseFunc(response)
	}
	return string(response), nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantType llm.LLM
	}{
		{"OpenAI", "openai", &llm.OpenAILLM{}},
		{"Claude", "claude", &llm.ClaudeLLM{}},
		{"Default", "", &llm.DefaultLLM{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.ClientConfig{Provider: tt.provider}
			client := New(config)
			assert.IsType(t, tt.wantType, client.llm)
		})
	}
}

func TestCreateProxyTransport(t *testing.T) {
	tests := []struct {
		name       string
		proxy      string
		wantErr    bool
		wantScheme string
	}{
		{"No Proxy", "", false, ""},
		{"HTTP Proxy", "http://proxy.example.com", false, "http"},
		{"SOCKS5 Proxy", "socks5://user:pass@proxy.example.com", false, "socks5"},
		{"Invalid Proxy", "invalid://proxy.example.com", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.ClientConfig{Proxy: tt.proxy}
			client := &Client{config: config}

			transport, err := client.createProxyTransport()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.proxy != "" {
				if tt.wantScheme == "socks5" {
					assert.NotNil(t, transport.DialContext)
				} else {
					assert.NotNil(t, transport.Proxy)
				}
			}
		})
	}
}

func TestChat(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			if !stream {
				return "mock response", nil
			}
			return "", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10, Retries: 3},
		llm:    mockLLM,
	}

	resp, err := client.Chat(context.Background(), "test message", nil)
	require.NoError(t, err)
	assert.Equal(t, "mock response", resp.Content)
}

func TestChatWithRetries(t *testing.T) {
	var attempt int
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			if attempt < 2 {
				attempt++
				return "", errors.New("temporary error")
			}
			return "mock response after retries", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10, Retries: 3},
		llm:    mockLLM,
	}

	resp, err := client.Chat(context.Background(), "test message", nil)
	require.NoError(t, err)
	assert.Equal(t, "mock response after retries", resp.Content)
	assert.Equal(t, 2, attempt)
}

func TestChatErrorAfterRetries(t *testing.T) {
	var attempt int
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			attempt++
			return "", errors.New("persistent error")
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10, Retries: 3},
		llm:    mockLLM,
	}

	_, err := client.Chat(context.Background(), "test message", nil)
	assert.Error(t, err)
	assert.Equal(t, 3, attempt) // 1 initial + 2 retries
	assert.Contains(t, err.Error(), "after 3 attempts")
}

func TestChatContextCancellation(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			return "", context.Canceled
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10, Retries: 3},
		llm:    mockLLM,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Chat(ctx, "test message", nil)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestTranslateMessage(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			return "translated message", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	_, err := client.TranslateMessage("translate to %s: %s", "hello", "fr")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "after 0 attempts")
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		mockError    error
		wantMessage  string
		wantError    bool
	}{
		{
			name:         "success",
			mockResponse: "commit message",
			mockError:    nil,
			wantMessage:  "commit message",
			wantError:    false,
		},
		{
			name:         "error",
			mockResponse: "",
			mockError:    errors.New("api error"),
			wantMessage:  "",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &MockLLM{
				makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
					return tt.mockResponse, tt.mockError
				},
				name: "mock",
			}

			client := &Client{
				config: &types.ClientConfig{Timeout: 10, Retries: 3},
				llm:    mockLLM,
			}

			msg, err := client.GenerateCommitMessage("diff", "generate commit message for: %s")
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, msg)
		})
	}
}

func TestGenerateReviewComment(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		mockError    error
		wantComment  string
		wantError    bool
	}{
		{
			name:         "success",
			mockResponse: "review comment",
			mockError:    nil,
			wantComment:  "review comment",
			wantError:    false,
		},
		{
			name:         "error",
			mockResponse: "",
			mockError:    errors.New("api error"),
			wantComment:  "",
			wantError:    true,
		},
		{
			name:         "retry success",
			mockResponse: "review comment after retry",
			mockError:    nil,
			wantComment:  "review comment after retry",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempt int
			mockLLM := &MockLLM{
				makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
					if tt.name == "retry success" && attempt < 2 {
						attempt++
						return "", errors.New("temporary error")
					}
					return tt.mockResponse, tt.mockError
				},
				name: "mock",
			}

			client := &Client{
				config: &types.ClientConfig{Timeout: 10, Retries: 3},
				llm:    mockLLM,
			}

			comment, err := client.GenerateReviewComment("diff", "generate review comment for: %s")
			if tt.wantError {
				assert.Error(t, err)
				if tt.name == "retry success" {
					assert.Equal(t, 2, attempt)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantComment, comment)
		})
	}
}

// MockRoundTripper is a testable implementation of http.RoundTripper
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip implements the http.RoundTripper interface
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return nil, errors.New("mock RoundTrip function not implemented")
}

func TestStream(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		setupMock      func() *MockLLM
		message        string
		expectedChunks []string
		wantErr        bool
		errorContains  string
	}{
		{
			name: "Success",
			setupMock: func() *MockLLM {
				return &MockLLM{
					name: "mock",
					formatMessagesFunc: func(message string) (interface{}, error) {
						return message, nil
					},
					buildURLFunc: func() string {
						return "https://api.example.com/chat"
					},
					buildHeadersFunc: func() map[string]string {
						return map[string]string{"Authorization": "Bearer test-token"}
					},
					makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
						return "", nil // Not needed for streaming
					},
				}
			},
			message:        "test message",
			expectedChunks: []string{"hello", " world", "!", "\n"},
			wantErr:        false,
		},
		{
			name: "FormatMessages Error",
			setupMock: func() *MockLLM {
				return &MockLLM{
					name: "mock",
					formatMessagesFunc: func(message string) (interface{}, error) {
						return nil, errors.New("format error")
					},
				}
			},
			message:        "test message",
			expectedChunks: nil,
			wantErr:        true,
			errorContains:  "failed to format messages",
		},
		{
			name: "GetClient Error",
			setupMock: func() *MockLLM {
				return &MockLLM{
					name: "mock",
					formatMessagesFunc: func(message string) (interface{}, error) {
						return message, nil
					},
				}
			},
			message:        "test message",
			expectedChunks: nil,
			wantErr:        true,
			errorContains:  "unsupported proxy scheme", // Fix error message expectation
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock LLM
			mockLLM := tt.setupMock()

			// Create client with appropriate configuration
			var client *Client
			if tt.name == "GetClient Error" {
				// Configure invalid proxy to trigger getClient error
				client = &Client{
					config: &types.ClientConfig{Proxy: "invalid://proxy.example.com"},
					llm:    mockLLM,
				}
			} else {
				client = &Client{
					config: &types.ClientConfig{},
					llm:    mockLLM,
				}
			}

			// For success case, create a test server that returns our mock response
			var testServer *httptest.Server
			if tt.name == "Success" {
				// Create a test server that returns our mock streaming response
				testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request properties
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

					// Set necessary headers for SSE
					w.Header().Set("Content-Type", "text/event-stream")
					w.Header().Set("Cache-Control", "no-cache")
					w.Header().Set("Connection", "keep-alive")
					w.WriteHeader(http.StatusOK)

					// Write the chunks as SSE events in the format expected by the Stream method
					flusher, ok := w.(http.Flusher)
					if !ok {
						t.Fatal("ResponseWriter does not implement Flusher")
					}

					// Send each chunk as a properly formatted SSE event
					for _, chunk := range tt.expectedChunks {
						// Create a compact JSON structure expected by the Stream method
						responseData := fmt.Sprintf(`{"choices":[{"delta":{"content":"%s"}}]}`, chunk)

						// Write as SSE data event
						fmt.Fprintf(w, "data: %s\n\n", responseData)
						flusher.Flush()
					}

					// Send a [DONE] event to signal completion
					fmt.Fprintf(w, "data: [DONE]\n\n")
					flusher.Flush()
				}))
				defer testServer.Close()

				// Override the buildURL method to use our test server instead
				mockLLM.buildURLFunc = func() string {
					return testServer.URL
				}
			}

			// Call Stream with our test callback
			var receivedChunks []string
			err := client.Stream(context.Background(), tt.message, func(resp *types.CompletionResponse) error {
				// Record the received chunk
				if tt.name == "Success" {
					// Check that Raw is populated
					assert.NotNil(t, resp.Raw)
				}

				// Add to received chunks
				receivedChunks = append(receivedChunks, resp.Content)

				return nil
			})

			// Check if error matches expectation
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			// If successful, check that all expected chunks were received
			require.NoError(t, err)
			assert.Equal(t, tt.expectedChunks, receivedChunks)
		})
	}
}
