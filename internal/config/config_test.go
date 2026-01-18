package config

import (
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		configData  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "Success with empty config",
			configData: "",
		},
		{
			name: "Success with valid config",
			configData: `
provider: openai
openai:
  api_key: test-key
  api_base: https://api.openai.com/v1
  model: gpt-4
`,
		},
		{
			name: "Invalid YAML",
			configData: `
provider: openai
openai:
  api_key: [invalid
`,
			wantErr:     true,
			errContains: "failed to parse config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, configFile, cfg.GetPath())
		})
	}
}

func TestConfig_Set(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		key         string
		value       interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:       "Set provider",
			configData: "",
			key:        "provider",
			value:      "openai",
			wantErr:    false,
		},
		{
			name:       "Set invalid provider - unknown provider",
			configData: "",
			key:        "provider",
			value:      "invalid",
			wantErr:    false,
		},
		{
			name:       "Set invalid provider - empty string",
			configData: "",
			key:        "provider",
			value:      "",
			wantErr:    false,
		},
		{
			name:       "Set invalid provider - whitespace only",
			configData: "",
			key:        "provider",
			value:      "   ",
			wantErr:    false,
		},
		{
			name: "Set value in existing config",
			configData: `
provider: openai
openai:
  api_key: test-key
`,
			key:     "openai.model",
			value:   "gpt-4",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			err = cfg.Set(tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				val, ok := cfg.Get(tt.key)
				require.True(t, ok)
				assert.Equal(t, tt.value, val)
			}
		})
	}
}

func TestConfig_GetOutputTranslateTitle(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		want       bool
	}{
		{
			name:       "Empty config returns false",
			configData: "",
			want:       false,
		},
		{
			name: "Explicitly set to true",
			configData: `
output:
  translate_title: true
`,
			want: true,
		},
		{
			name: "Explicitly set to false",
			configData: `
output:
  translate_title: false
`,
			want: false,
		},
		{
			name: "Invalid value returns false",
			configData: `
output:
  translate_title: "invalid"
`,
			want: false,
		},
		{
			name: "Missing output section returns false",
			configData: `
provider: openai
`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			got := cfg.GetOutputTranslateTitle()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_GetClientConfig(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		initProvider string
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, cfg *types.ClientConfig)
	}{
		{
			name:         "Provider not set",
			configData:   `{}`,
			initProvider: "",
			wantErr:      true,
			errContains:  "provider not set",
		},
		{
			name:         "Provider config not found",
			configData:   `{"provider": "openai"}`,
			initProvider: "",
			wantErr:      true,
			errContains:  "provider config not found",
		},
		{
			name:         "API key not found",
			configData:   `{"provider": "openai", "openai": {}}`,
			initProvider: "",
			wantErr:      true,
			errContains:  "api_key not found",
		},
		{
			name: "Basic config with default values",
			configData: `
provider: openai
openai:
  api_key: test-key
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "openai", cfg.Provider)
				assert.Equal(t, "test-key", cfg.APIKey)
				// Check defaults
				assert.Equal(t, "https://api.openai.com/v1", cfg.APIBase)
				assert.Equal(t, "gpt-4o", cfg.Model)
				assert.Equal(t, 1024, cfg.MaxTokens)
				assert.Equal(t, 1.0, cfg.TopP)
				assert.Equal(t, 0.3, cfg.Temperature)
				assert.Equal(t, 0.0, cfg.FrequencyPenalty)
				assert.Equal(t, 3, cfg.Retries)
				assert.Equal(t, "", cfg.Proxy)
			},
		},
		{
			name: "Override provider with initProvider",
			configData: `
provider: openai
openai:
  api_key: openai-key
anthropic:
  api_key: anthropic-key
`,
			initProvider: "anthropic",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "anthropic", cfg.Provider)
				assert.Equal(t, "anthropic-key", cfg.APIKey)
			},
		},
		{
			name: "Ollama provider (no API key required)",
			configData: `
provider: ollama
ollama:
  model: llama2
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "ollama", cfg.Provider)
				assert.Equal(t, "", cfg.APIKey)
				assert.Equal(t, "llama2", cfg.Model)
			},
		},
		{
			name: "Custom API base",
			configData: `
provider: openai
openai:
  api_key: test-key
  api_base: https://custom-openai.example.com/v1
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "https://custom-openai.example.com/v1", cfg.APIBase)
			},
		},
		{
			name: "Custom model",
			configData: `
provider: openai
openai:
  api_key: test-key
  model: gpt-4
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "gpt-4", cfg.Model)
			},
		},
		{
			name: "Proxy setting",
			configData: `
provider: openai
openai:
  api_key: test-key
  proxy: http://proxy.example.com:8080
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "http://proxy.example.com:8080", cfg.Proxy)
			},
		},
		{
			name: "Max tokens",
			configData: `
provider: openai
openai:
  api_key: test-key
  max_tokens: 2048
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 2048, cfg.MaxTokens)
			},
		},
		{
			name: "Top P",
			configData: `
provider: openai
openai:
  api_key: test-key
  top_p: 0.8
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 0.8, cfg.TopP)
			},
		},
		{
			name: "Temperature",
			configData: `
provider: openai
openai:
  api_key: test-key
  temperature: 0.5
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 0.5, cfg.Temperature)
			},
		},
		{
			name: "Frequency penalty",
			configData: `
provider: openai
openai:
  api_key: test-key
  frequency_penalty: 0.3
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 0.3, cfg.FrequencyPenalty)
			},
		},
		{
			name: "Retries",
			configData: `
provider: openai
openai:
  api_key: test-key
  retries: 5
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 3, cfg.Retries)
			},
		},
		{
			name: "Answer path",
			configData: `
provider: openai
openai:
  api_key: test-key
  answer_path: choices.0.message.content
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "choices.0.message.content", cfg.AnswerPath)
			},
		},
		{
			name: "Completion path",
			configData: `
provider: openai
openai:
  api_key: test-key
  completion_path: /v1/chat/completions
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.NotNil(t, cfg.CompletionPath)
				assert.Equal(t, "/v1/chat/completions", *cfg.CompletionPath)
			},
		},
		{
			name: "Extra headers - valid JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_headers: '{"X-Custom-Header": "custom-value", "Authorization": "Bearer token"}'
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.NotNil(t, cfg.ExtraHeaders)
				assert.Equal(t, "custom-value", cfg.ExtraHeaders["X-Custom-Header"])
				assert.Equal(t, "Bearer token", cfg.ExtraHeaders["Authorization"])
			},
		},
		{
			name: "Extra headers - empty JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_headers: '{}'
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Nil(t, cfg.ExtraHeaders)
			},
		},
		{
			name: "Extra headers - invalid JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_headers: '{invalid json}'
`,
			initProvider: "",
			wantErr:      true,
			errContains:  "failed to parse extra_headers",
		},
		{
			name: "Extra body - valid JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_body: '{"custom_field": "custom-value", "options": {"stream": true}}'
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.NotNil(t, cfg.ExtraBody)
				assert.Equal(t, "custom-value", cfg.ExtraBody["custom_field"])
				options, ok := cfg.ExtraBody["options"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, true, options["stream"])
			},
		},
		{
			name: "Extra body - empty JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_body: '{}'
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Nil(t, cfg.ExtraBody)
			},
		},
		{
			name: "Extra body - invalid JSON",
			configData: `
provider: openai
openai:
  api_key: test-key
  extra_body: '{invalid json}'
`,
			initProvider: "",
			wantErr:      true,
			errContains:  "failed to parse extra_body",
		},
		{
			name: "All configuration options",
			configData: `
provider: openai
openai:
  api_key: test-key
  api_base: https://custom-api.example.com/v1
  model: gpt-4-turbo
  max_tokens: 4096
  top_p: 0.95
  temperature: 0.8
  frequency_penalty: 0.2
  retries: 4
  proxy: http://proxy.example.com:8080
  answer_path: choices.0.message.content
  completion_path: /v1/custom/completions
  extra_headers: '{"X-Custom-Header": "custom-value"}'
  extra_body: '{"custom_field": "custom-value"}'
`,
			initProvider: "",
			wantErr:      false,
			validateFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "openai", cfg.Provider)
				assert.Equal(t, "test-key", cfg.APIKey)
				assert.Equal(t, "https://custom-api.example.com/v1", cfg.APIBase)
				assert.Equal(t, "gpt-4-turbo", cfg.Model)
				assert.Equal(t, 4096, cfg.MaxTokens)
				assert.Equal(t, 0.95, cfg.TopP)
				assert.Equal(t, 0.8, cfg.Temperature)
				assert.Equal(t, 0.2, cfg.FrequencyPenalty)
				assert.Equal(t, 3, cfg.Retries)
				assert.Equal(t, "http://proxy.example.com:8080", cfg.Proxy)
				assert.Equal(t, "choices.0.message.content", cfg.AnswerPath)
				assert.NotNil(t, cfg.CompletionPath)
				assert.Equal(t, "/v1/custom/completions", *cfg.CompletionPath)
				assert.NotNil(t, cfg.ExtraHeaders)
				assert.Equal(t, "custom-value", cfg.ExtraHeaders["X-Custom-Header"])
				assert.NotNil(t, cfg.ExtraBody)
				assert.Equal(t, "custom-value", cfg.ExtraBody["custom_field"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			clientConfig, err := cfg.GetClientConfig(tt.initProvider)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, clientConfig)

			if tt.validateFunc != nil {
				tt.validateFunc(t, clientConfig)
			}
		})
	}
}
