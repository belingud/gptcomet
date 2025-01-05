package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/belingud/go-gptcomet/internal/llm"
	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLM implements the LLM interface for testing
type MockLLM struct {
	makeRequestFunc       func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error)
	buildHeadersFunc      func() map[string]string
	buildURLFunc          func() string
	formatMessagesFunc    func(model string, messages []types.Message) (interface{}, error)
	getRequiredConfigFunc func() map[string]config.ConfigRequirement
	getUsageFunc          func(data []byte) (string, error)
	parseResponseFunc     func(response []byte) (string, error)
	name                  string
}

func (m *MockLLM) Name() string {
	return m.name
}

func (m *MockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	return m.makeRequestFunc(ctx, client, message, history)
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

func (m *MockLLM) FormatMessages(model string, messages []types.Message) (interface{}, error) {
	if m.formatMessagesFunc != nil {
		return m.formatMessagesFunc(model, messages)
	}
	return messages, nil
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
		{"Default", "", &llm.OpenAILLM{}},
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
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
			return "mock response", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	resp, err := client.Chat(context.Background(), "test message", nil)
	require.NoError(t, err)
	assert.Equal(t, "mock response", resp.Content)
}

func TestChatError(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
			return "", errors.New("mock error")
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	_, err := client.Chat(context.Background(), "test message", nil)
	assert.Error(t, err)
}

func TestTranslateMessage(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
			return "translated message", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	translated, err := client.TranslateMessage("translate to %s: %s", "hello", "fr")
	require.NoError(t, err)
	assert.Equal(t, "translated message", translated)
}

func TestGenerateCommitMessage(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
			return "commit message", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	msg, err := client.GenerateCommitMessage("diff", "generate commit message for: %s")
	require.NoError(t, err)
	assert.Equal(t, "commit message", msg)
}

func TestGenerateCodeExplanation(t *testing.T) {
	mockLLM := &MockLLM{
		makeRequestFunc: func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
			return "code explanation", nil
		},
		name: "mock",
	}

	client := &Client{
		config: &types.ClientConfig{Timeout: 10},
		llm:    mockLLM,
	}

	explanation, err := client.GenerateCodeExplanation("code", "go")
	require.NoError(t, err)
	assert.Equal(t, "code explanation", explanation)
}
