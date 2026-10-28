package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/belingud/gptcomet/internal/constants"
	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/config/defaults"
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
		{"Default (empty provider)", "", &llm.OpenAILLM{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.ClientConfig{Provider: tt.provider}
			client, err := New(config)
			require.NoError(t, err)
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
	t.Run("default template placeholders", func(t *testing.T) {
		var captured string
		mockLLM := &MockLLM{
			makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
				captured = message
				return "translated message", nil
			},
			name: "mock",
		}

		client := &Client{
			config: &types.ClientConfig{Timeout: 10, Retries: 1},
			llm:    mockLLM,
		}

		got, err := client.TranslateMessage(defaults.PromptDefaults["translation"], "feat: add", "Simplified Chinese")
		require.NoError(t, err)
		assert.Equal(t, "translated message", got)
		assert.Contains(t, captured, "feat: add")
		assert.Contains(t, captured, "Simplified Chinese")
		assert.NotContains(t, captured, "{{ placeholder }}")
		assert.NotContains(t, captured, "{{ output.lang }}")
		assert.NotContains(t, captured, "%!(EXTRA")
	})

	t.Run("template without placeholders still includes message", func(t *testing.T) {
		var captured string
		mockLLM := &MockLLM{
			makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
				captured = message
				return "translated message", nil
			},
			name: "mock",
		}

		client := &Client{
			config: &types.ClientConfig{Timeout: 10, Retries: 1},
			llm:    mockLLM,
		}

		_, err := client.TranslateMessage("translate this", "hello", "fr")
		require.NoError(t, err)
		assert.Contains(t, captured, "translate this")
		assert.Contains(t, captured, "Target language: fr")
		assert.Contains(t, captured, "hello")
	})

	t.Run("legacy percent template", func(t *testing.T) {
		var captured string
		mockLLM := &MockLLM{
			makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
				captured = message
				return "translated message", nil
			},
			name: "mock",
		}

		client := &Client{
			config: &types.ClientConfig{Timeout: 10, Retries: 1},
			llm:    mockLLM,
		}

		_, err := client.TranslateMessage("translate to %s: %s", "hello", "fr")
		require.NoError(t, err)
		assert.Equal(t, "translate to hello: fr", captured)
	})
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
			errorContains:  "Message Formatting Failed",
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
			errorContains:  "Unsupported Proxy Scheme", // Fix error message expectation
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

func TestStreamRetriesBeforeFirstChunk(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				t.Error("response writer does not support hijacking")
				return
			}
			conn, _, err := hijacker.Hijack()
			if err == nil {
				conn.Close()
			}
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	mockLLM := &MockLLM{
		name: "mock",
		buildURLFunc: func() string {
			return server.URL
		},
		formatMessagesFunc: func(message string) (interface{}, error) {
			return map[string]interface{}{"model": "test"}, nil
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries:          3,
			StreamAnswerPath: "choices.0.delta.content",
		},
		llm: mockLLM,
	}

	var chunks []string
	err := client.Stream(context.Background(), "test message", func(resp *types.CompletionResponse) error {
		chunks = append(chunks, resp.Content)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "\n"}, chunks)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestStreamDoesNotRetryAfterPartialOutput(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\n")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		if hijacker, ok := w.(http.Hijacker); ok {
			conn, _, err := hijacker.Hijack()
			if err == nil {
				conn.Close()
			}
		}
	}))
	defer server.Close()

	mockLLM := &MockLLM{
		name: "mock",
		buildURLFunc: func() string {
			return server.URL
		},
		formatMessagesFunc: func(message string) (interface{}, error) {
			return map[string]interface{}{"model": "test"}, nil
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries:          3,
			StreamAnswerPath: "choices.0.delta.content",
		},
		llm: mockLLM,
	}

	var chunks []string
	err := client.Stream(context.Background(), "test message", func(resp *types.CompletionResponse) error {
		chunks = append(chunks, resp.Content)
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
	assert.LessOrEqual(t, len(chunks), 1)
}

func TestIsRetryableStreamError(t *testing.T) {
	assert.True(t, isRetryableStreamError(gptErrors.RequestExecutionError(errors.New("connection reset"))))
	assert.False(t, isRetryableStreamError(gptErrors.APIStatusError(http.StatusInternalServerError, "", nil)))
	assert.False(t, isRetryableStreamError(errors.New("plain error")))
}

func TestStreamRetriesAfterEmpty200(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusOK)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			if hijacker, ok := w.(http.Hijacker); ok {
				conn, _, err := hijacker.Hijack()
				if err == nil {
					conn.Close()
				}
			}
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	mockLLM := &MockLLM{
		name: "mock",
		buildURLFunc: func() string {
			return server.URL
		},
		formatMessagesFunc: func(message string) (interface{}, error) {
			return map[string]interface{}{"model": "test"}, nil
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries:          3,
			StreamAnswerPath: "choices.0.delta.content",
		},
		llm: mockLLM,
	}

	var chunks []string
	err := client.Stream(context.Background(), "test message", func(resp *types.CompletionResponse) error {
		chunks = append(chunks, resp.Content)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "\n"}, chunks)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestStreamNoExtraNewlineWhenContentEndsWithNewline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\\n\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	mockLLM := &MockLLM{
		name: "mock",
		buildURLFunc: func() string {
			return server.URL
		},
		formatMessagesFunc: func(message string) (interface{}, error) {
			return map[string]interface{}{"model": "test"}, nil
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries:          1,
			StreamAnswerPath: "choices.0.delta.content",
		},
		llm: mockLLM,
	}

	var chunks []string
	err := client.Stream(context.Background(), "test message", func(resp *types.CompletionResponse) error {
		chunks = append(chunks, resp.Content)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello\n"}, chunks)
}

func TestClientRequestTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout int64
		want    time.Duration
	}{
		{name: "default when unset", timeout: 0, want: constants.DefaultTimeout},
		{name: "default when negative", timeout: -5, want: constants.DefaultTimeout},
		{name: "configured value", timeout: 12, want: 12 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{config: &types.ClientConfig{Timeout: tt.timeout}}
			assert.Equal(t, tt.want, c.requestTimeout())
		})
	}
}

func TestChatTimeout(t *testing.T) {
	mockLLM := &MockLLM{
		name: "mock",
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
			<-ctx.Done()
			return "", ctx.Err()
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries: 1,
			Timeout: 1,
		},
		llm: mockLLM,
	}

	start := time.Now()
	_, err := client.Chat(context.Background(), "test message", nil)
	require.Error(t, err)
	assert.Less(t, time.Since(start), 3*time.Second)
}

func TestGetClientUsesResponseHeaderTimeout(t *testing.T) {
	c := &Client{config: &types.ClientConfig{Timeout: 5}}

	httpClient, err := c.getClient()
	require.NoError(t, err)

	// The total timeout is enforced per-call via context so streaming
	// responses are not cut off.
	assert.Zero(t, httpClient.Timeout)

	transport, ok := httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 5*time.Second, transport.ResponseHeaderTimeout)
}

func TestStreamIdleTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		// Send no content and keep the connection open until the client
		// cancels it.
		<-r.Context().Done()
	}))
	defer server.Close()

	mockLLM := &MockLLM{
		name: "mock",
		buildURLFunc: func() string {
			return server.URL
		},
		formatMessagesFunc: func(message string) (interface{}, error) {
			return map[string]interface{}{"model": "test"}, nil
		},
	}

	client := &Client{
		config: &types.ClientConfig{
			Retries: 1,
			Timeout: 1,
		},
		llm: mockLLM,
	}

	done := make(chan error, 1)
	go func() {
		done <- client.Stream(context.Background(), "test message", func(resp *types.CompletionResponse) error {
			return nil
		})
	}()

	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Stream did not return after the idle timeout")
	}
}
