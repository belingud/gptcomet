package config

import (
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_ResolveAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		provider   string
		envVar     string
		envValue   string
		wantKey    string
		wantSource APIKeySource
	}{
		{
			name: "config key wins over environment",
			configData: `
provider: openai
openai:
  api_key: config-key
`,
			provider:   "openai",
			envVar:     "OPENAI_API_KEY",
			envValue:   "env-key",
			wantKey:    "config-key",
			wantSource: APIKeySourceConfig,
		},
		{
			name: "environment fallback when config is empty",
			configData: `
provider: openai
openai:
  api_base: https://api.openai.com/v1
`,
			provider:   "openai",
			envVar:     "OPENAI_API_KEY",
			envValue:   "env-key",
			wantKey:    "env-key",
			wantSource: APIKeySourceEnv,
		},
		{
			name: "missing when neither config nor environment provides a key",
			configData: `
provider: openai
openai:
  api_base: https://api.openai.com/v1
`,
			provider:   "openai",
			envVar:     "OPENAI_API_KEY",
			envValue:   "",
			wantKey:    "",
			wantSource: APIKeySourceMissing,
		},
		{
			name: "ollama does not require a key",
			configData: `
provider: ollama
ollama:
  api_key: ignored-config-key
  api_base: http://localhost:11434/api
`,
			provider:   "ollama",
			envVar:     "OLLAMA_API_KEY",
			envValue:   "ignored-env-key",
			wantKey:    "",
			wantSource: APIKeySourceNotRequired,
		},
		{
			name: "lmstudio does not require a key",
			configData: `
provider: lmstudio
lmstudio:
  api_base: http://localhost:1234/v1
`,
			provider:   "lmstudio",
			envVar:     "LMSTUDIO_API_KEY",
			envValue:   "",
			wantKey:    "",
			wantSource: APIKeySourceNotRequired,
		},
		{
			name: "provider specific environment variable name is used",
			configData: `
provider: bedrock
bedrock:
  region: us-east-1
`,
			provider:   "bedrock",
			envVar:     "AWS_BEARER_TOKEN_BEDROCK",
			envValue:   "bedrock-token",
			wantKey:    "bedrock-token",
			wantSource: APIKeySourceEnv,
		},
		{
			name: "unknown provider uses the generic environment variable",
			configData: `
provider: custom-proxy
custom-proxy:
  api_base: https://example.invalid/v1
`,
			provider:   "custom-proxy",
			envVar:     "CUSTOM_PROXY_API_KEY",
			envValue:   "custom-env-key",
			wantKey:    "custom-env-key",
			wantSource: APIKeySourceEnv,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			t.Setenv(tt.envVar, tt.envValue)

			key, source := cfg.ResolveAPIKey(tt.provider)
			assert.Equal(t, tt.wantKey, key)
			assert.Equal(t, tt.wantSource, source)
		})
	}
}
