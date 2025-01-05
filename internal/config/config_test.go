package config

import (
	"testing"

	"github.com/belingud/go-gptcomet/internal/testutils"
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
		key         string
		value       interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:    "Set provider",
			key:     "provider",
			value:   "openai",
			wantErr: false,
		},
		{
			name:    "Set invalid provider - unknown provider",
			key:     "provider",
			value:   "invalid",
			wantErr: false,
		},
		{
			name:    "Set invalid provider - empty string",
			key:     "provider",
			value:   "",
			wantErr: false,
		},
		{
			name:    "Set invalid provider - whitespace only",
			key:     "provider",
			value:   "   ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := New("")
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
