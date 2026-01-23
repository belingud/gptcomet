package integration

import (
	"testing"

	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderInitializationAllProviders tests that all registered providers can be initialized
func TestProviderInitializationAllProviders(t *testing.T) {
	// List of all providers that should be registered
	providers := []struct {
		name        string
		configKey   string
		requiresKey bool
		config      string
	}{
		{
			name:        "OpenAI",
			configKey:   "openai",
			requiresKey: true,
			config: `
provider: openai
openai:
  api_key: test-key
  model: gpt-4o
`,
		},
		{
			name:        "Claude/Anthropic",
			configKey:   "anthropic",
			requiresKey: true,
			config: `
provider: anthropic
anthropic:
  api_key: test-key
  model: claude-3-opus
`,
		},
		{
			name:        "Gemini",
			configKey:   "gemini",
			requiresKey: true,
			config: `
provider: gemini
gemini:
  api_key: test-key
  model: gemini-pro
`,
		},
		{
			name:        "Ollama",
			configKey:   "ollama",
			requiresKey: false,
			config: `
provider: ollama
ollama:
  api_base: http://localhost:11434
  model: llama2
`,
		},
		{
			name:        "DeepSeek",
			configKey:   "deepseek",
			requiresKey: true,
			config: `
provider: deepseek
deepseek:
  api_key: test-key
  model: deepseek-chat
`,
		},
		{
			name:        "Groq",
			configKey:   "groq",
			requiresKey: true,
			config: `
provider: groq
groq:
  api_key: test-key
  model: llama2-70b-4096
`,
		},
		{
			name:        "Mistral",
			configKey:   "mistral",
			requiresKey: true,
			config: `
provider: mistral
mistral:
  api_key: test-key
  model: mistral-medium
`,
		},
		{
			name:        "Cohere",
			configKey:   "cohere",
			requiresKey: true,
			config: `
provider: cohere
cohere:
  api_key: test-key
  model: command
`,
		},
		{
			name:        "AI21",
			configKey:   "ai21",
			requiresKey: true,
			config: `
provider: ai21
ai21:
  api_key: test-key
  model: j2-ultra
`,
		},
		{
			name:        "XAI",
			configKey:   "xai",
			requiresKey: true,
			config: `
provider: xai
xai:
  api_key: test-key
  model: grok-beta
`,
		},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.config)
			defer cleanup()

			cfg, err := config.New(configFile)
			require.NoError(t, err)

			clientCfg, err := cfg.GetClientConfig("")
			require.NoError(t, err)

			// Verify client config has expected values
			assert.Equal(t, tt.configKey, clientCfg.Provider)
			if tt.requiresKey {
				assert.NotEmpty(t, clientCfg.APIKey, "Provider %s requires API key", tt.name)
			}

			// Initialize client with provider
			apiClient, err := client.New(clientCfg)
			require.NoError(t, err)
			assert.NotNil(t, apiClient)
		})
	}
}

// TestProviderRegistry tests that the provider registry works correctly
func TestProviderRegistry(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		expectError  bool
	}{
		{
			name:         "OpenAI provider exists",
			providerName: "openai",
			expectError:  false,
		},
		{
			name:         "Claude provider exists",
			providerName: "claude",
			expectError:  false,
		},
		{
			name:         "Gemini provider exists",
			providerName: "gemini",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal config for the provider
			cfg := &types.ClientConfig{
				Provider: tt.providerName,
				APIKey:   "test-key",
				APIBase:  "https://api.example.com/v1",
				Model:    "test-model",
			}

			// Get provider from registry via NewProvider
			provider, err := llm.NewProvider(tt.providerName, cfg)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, provider, "Provider should not be nil")

			// Verify provider has required methods
			assert.NotEmpty(t, provider.Name())
			assert.NotNil(t, provider.GetRequiredConfig())
		})
	}
}

// TestProviderConfigRequirements tests that each provider correctly defines its config requirements
func TestProviderConfigRequirements(t *testing.T) {
	providerNames := []string{
		"openai", "claude", "gemini", "ollama", "deepseek",
		"groq", "mistral", "cohere", "ai21", "xai",
	}

	for _, providerName := range providerNames {
		t.Run(providerName, func(t *testing.T) {
			// Create a minimal config
			cfg := &types.ClientConfig{
				Provider: providerName,
				APIKey:   "test-key",
				APIBase:  "https://api.example.com/v1",
				Model:    "test-model",
			}

			provider, err := llm.NewProvider(providerName, cfg)
			require.NoError(t, err)
			require.NotNil(t, provider)

			requirements := provider.GetRequiredConfig()
			assert.NotNil(t, requirements, "Provider %s should have config requirements", providerName)

			// Verify common fields are defined
			if providerName != "ollama" && providerName != "vertex" && providerName != "azure" {
				// Most providers require an API key
				_, hasAPIKey := requirements["api_key"]
				assert.True(t, hasAPIKey, "Provider %s should define api_key requirement", providerName)
			}

			// All providers should define model
			_, hasModel := requirements["model"]
			assert.True(t, hasModel, "Provider %s should define model requirement", providerName)
		})
	}
}

// TestProviderSwitching tests switching between providers at runtime
func TestProviderSwitching(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: openai-key
  model: gpt-4
anthropic:
  api_key: anthropic-key
  model: claude-3-opus
gemini:
  api_key: gemini-key
  model: gemini-pro
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Test switching to different providers
	testCases := []struct {
		provider string
		wantKey  string
		wantName string
	}{
		{"openai", "openai-key", "openai"},
		{"anthropic", "anthropic-key", "anthropic"},
		{"gemini", "gemini-key", "gemini"},
	}

	for _, tc := range testCases {
		t.Run(tc.provider, func(t *testing.T) {
			clientCfg, err := cfg.GetClientConfig(tc.provider)
			require.NoError(t, err)

			assert.Equal(t, tc.provider, clientCfg.Provider)
			assert.Equal(t, tc.wantKey, clientCfg.APIKey)

			// Create client with this provider
			apiClient, err := client.New(clientCfg)
			require.NoError(t, err)
			assert.NotNil(t, apiClient)
		})
	}
}

// TestProviderURLBuilding tests that providers build correct URLs
func TestProviderURLBuilding(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		provider    string
		expectedURL string
	}{
		{
			name: "OpenAI default URL",
			config: `
provider: openai
openai:
  api_key: test-key
`,
			provider:    "openai",
			expectedURL: "https://api.openai.com/v1/chat/completions",
		},
		{
			name: "OpenAI custom URL",
			config: `
provider: openai
openai:
  api_key: test-key
  api_base: https://custom.openai.com/v1
`,
			provider:    "openai",
			expectedURL: "https://custom.openai.com/v1/chat/completions",
		},
		{
			name: "Ollama default URL",
			config: `
provider: ollama
ollama:
  model: llama2
  api_base: http://localhost:11434
`,
			provider:    "ollama",
			expectedURL: "http://localhost:11434/generate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.config)
			defer cleanup()

			cfg, err := config.New(configFile)
			require.NoError(t, err)

			clientCfg, err := cfg.GetClientConfig(tt.provider)
			require.NoError(t, err)

			// Create provider using NewProvider
			provider, err := llm.NewProvider(tt.provider, clientCfg)
			require.NoError(t, err)
			require.NotNil(t, provider)

			// Build URL
			url := provider.BuildURL()
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

// TestProviderHeadersBuilding tests that providers build correct headers
func TestProviderHeadersBuilding(t *testing.T) {
	tests := []struct {
		name           string
		config         string
		provider       string
		expectedHeader string
		headerKey      string
	}{
		{
			name: "OpenAI headers",
			config: `
provider: openai
openai:
  api_key: test-key-123
`,
			provider:       "openai",
			expectedHeader: "Bearer test-key-123",
			headerKey:      "Authorization",
		},
		{
			name: "Claude headers",
			config: `
provider: claude
claude:
  api_key: sk-ant-test-key
`,
			provider:       "claude",
			expectedHeader: "sk-ant-test-key",
			headerKey:      "x-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.config)
			defer cleanup()

			cfg, err := config.New(configFile)
			require.NoError(t, err)

			clientCfg, err := cfg.GetClientConfig(tt.provider)
			require.NoError(t, err)

			// Create provider using NewProvider
			provider, err := llm.NewProvider(tt.provider, clientCfg)
			require.NoError(t, err)
			require.NotNil(t, provider)

			// Build headers
			headers := provider.BuildHeaders()
			assert.NotNil(t, headers)

			if tt.headerKey != "" {
				value, exists := headers[tt.headerKey]
				assert.True(t, exists, "Header %s should exist", tt.headerKey)
				assert.Equal(t, tt.expectedHeader, value)
			}
		})
	}
}

// TestProviderWithProxyConfiguration tests that providers work with proxy settings
func TestProviderWithProxyConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		wantProxy string
	}{
		{
			name: "HTTP proxy",
			config: `
provider: openai
openai:
  api_key: test-key
  proxy: http://proxy.example.com:8080
`,
			wantProxy: "http://proxy.example.com:8080",
		},
		{
			name: "SOCKS5 proxy",
			config: `
provider: openai
openai:
  api_key: test-key
  proxy: socks5://user:pass@proxy.example.com:1080
`,
			wantProxy: "socks5://user:pass@proxy.example.com:1080",
		},
		{
			name: "No proxy",
			config: `
provider: openai
openai:
  api_key: test-key
`,
			wantProxy: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.config)
			defer cleanup()

			cfg, err := config.New(configFile)
			require.NoError(t, err)

			clientCfg, err := cfg.GetClientConfig("")
			require.NoError(t, err)

			assert.Equal(t, tt.wantProxy, clientCfg.Proxy)

			// Create client (should not error even with proxy configured)
			apiClient, err := client.New(clientCfg)
			require.NoError(t, err)
			assert.NotNil(t, apiClient)
		})
	}
}
