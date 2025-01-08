package llm

import (
	"context"
	"net/http"
	"testing"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockProvider struct {
	name string
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{}
}

func (p *MockProvider) BuildHeaders() map[string]string {
	return map[string]string{}
}

func (p *MockProvider) BuildURL() string {
	return ""
}

func (p *MockProvider) FormatMessages(message string, history []types.Message) (interface{}, error) {
	return nil, nil
}

func (p *MockProvider) ParseResponse(response []byte) (string, error) {
	return "", nil
}

func (p *MockProvider) GetUsage(data []byte) (string, error) {
	return "", nil
}

func (p *MockProvider) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	return "mock response", nil
}

func (p *MockProvider) Chat(messages []types.Message) (string, error) {
	return "mock response", nil
}

func (p *MockProvider) GetConfig() *types.ClientConfig {
	return &types.ClientConfig{}
}

// Mock LLM implementation for testing
type mockLLM struct {
	name                  string
	generateCommitMessage func(diff string, prompt string) (string, error)
	translateMessage      func(prompt string, message string, lang string) (string, error)
	makeRequest           func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error)
}

func (m *mockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return nil
}

func (m *mockLLM) BuildHeaders() map[string]string {
	return nil
}

func (m *mockLLM) BuildURL() string {
	return ""
}

func (m *mockLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	return nil, nil
}

func (m *mockLLM) ParseResponse(response []byte) (string, error) {
	return "", nil
}

func (m *mockLLM) GetUsage(data []byte) (string, error) {
	return "", nil
}

func (m *mockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	if m.makeRequest != nil {
		return m.makeRequest(ctx, client, message, history)
	}
	return "", nil
}

func (m *mockLLM) GenerateCommitMessage(diff string, prompt string) (string, error) {
	if m.generateCommitMessage != nil {
		return m.generateCommitMessage(diff, prompt)
	}
	return "Test commit message", nil
}

func (m *mockLLM) TranslateMessage(prompt string, message string, lang string) (string, error) {
	if m.translateMessage != nil {
		return m.translateMessage(prompt, message, lang)
	}
	return message, nil
}

func (m *mockLLM) Name() string {
	return m.name
}

func TestRegisterProvider(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		constructor func(config *types.ClientConfig) LLM
		wantErr     bool
		errContains string
	}{
		{
			name:     "Register valid provider",
			provider: "mock",
			constructor: func(config *types.ClientConfig) LLM {
				return &mockLLM{name: "mock"}
			},
			wantErr: false,
		},
		{
			name:     "Register empty provider",
			provider: "",
			constructor: func(config *types.ClientConfig) LLM {
				return &mockLLM{name: "mock"}
			},
			wantErr:     true,
			errContains: "provider name cannot be empty",
		},
		{
			name:        "Register nil constructor",
			provider:    "mock",
			constructor: nil,
			wantErr:     true,
			errContains: "constructor cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset providers before each test
			providers = make(map[string]ProviderConstructor)

			err := RegisterProvider(tt.provider, tt.constructor)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Contains(t, GetProviders(), tt.provider)
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	// Reset providers before test
	providers = make(map[string]ProviderConstructor)

	// Register a mock provider
	err := RegisterProvider("mock", func(config *types.ClientConfig) LLM {
		return &MockProvider{name: "mock"}
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		provider    string
		config      *types.ClientConfig
		wantErr     bool
		errContains string
	}{
		{
			name:     "Create valid provider",
			provider: "mock",
			config:   &types.ClientConfig{},
			wantErr:  false,
		},
		{
			name:        "Create unknown provider",
			provider:    "unknown",
			config:      &types.ClientConfig{},
			wantErr:     true,
			errContains: "unknown provider",
		},
		{
			name:        "Create with nil config",
			provider:    "mock",
			config:      nil,
			wantErr:     true,
			errContains: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.provider, tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, provider)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.provider, provider.Name())
		})
	}
}

func TestListProviders(t *testing.T) {
	// Clear providers
	providers = make(map[string]ProviderConstructor)

	// Register test providers
	testProviders := []string{"mock1", "mock2", "mock3"}
	for _, p := range testProviders {
		err := RegisterProvider(p, func(config *types.ClientConfig) LLM {
			return &mockLLM{name: p}
		})
		require.NoError(t, err)
	}

	// Get list of providers
	list := GetProviders()
	assert.Equal(t, len(testProviders), len(list))
	for _, p := range testProviders {
		assert.Contains(t, list, p)
	}
}

func TestCreateProvider(t *testing.T) {
	// Reset providers before test
	providers = make(map[string]ProviderConstructor)

	// Register a mock provider
	err := RegisterProvider("mock", func(config *types.ClientConfig) LLM {
		return &MockProvider{name: "mock"}
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		config      *types.ClientConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid provider",
			config: &types.ClientConfig{
				Provider: "mock",
			},
			wantErr: false,
		},
		{
			name: "Unknown provider",
			config: &types.ClientConfig{
				Provider: "unknown",
			},
			wantErr:     true,
			errContains: "unknown provider",
		},
		{
			name:        "Nil config",
			config:      nil,
			wantErr:     true,
			errContains: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := CreateProvider(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, provider)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.config.Provider, provider.Name())
		})
	}
}
