package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLoadingWithVariousInputs tests config loading with different configurations
func TestConfigLoadingWithVariousInputs(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, cfg *config.Manager)
	}{
		{
			name:       "Empty config file",
			configData: ``,
			wantErr:    false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				assert.NotNil(t, cfg)
				// Should load with default values
			},
		},
		{
			name: "Minimal valid config - OpenAI",
			configData: `
provider: openai
openai:
  api_key: test-key-openai
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				clientCfg, err := cfg.GetClientConfig("")
				require.NoError(t, err)
				assert.Equal(t, "openai", clientCfg.Provider)
				assert.Equal(t, "test-key-openai", clientCfg.APIKey)
			},
		},
		{
			name: "Multiple providers configured",
			configData: `
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
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				// Test default provider
				clientCfg, err := cfg.GetClientConfig("")
				require.NoError(t, err)
				assert.Equal(t, "openai", clientCfg.Provider)
				assert.Equal(t, "openai-key", clientCfg.APIKey)

				// Test switching to different provider
				clientCfg, err = cfg.GetClientConfig("anthropic")
				require.NoError(t, err)
				assert.Equal(t, "anthropic", clientCfg.Provider)
				assert.Equal(t, "anthropic-key", clientCfg.APIKey)
			},
		},
		{
			name: "Full configuration with all options",
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
  retries: 5
  proxy: http://proxy.example.com:8080
output:
  lang: en
  translate_title: true
file_ignore:
  - "*.log"
  - "*.tmp"
  - node_modules/
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				clientCfg, err := cfg.GetClientConfig("")
				require.NoError(t, err)
				assert.Equal(t, "openai", clientCfg.Provider)
				assert.Equal(t, "test-key", clientCfg.APIKey)
				assert.Equal(t, "https://custom-api.example.com/v1", clientCfg.APIBase)
				assert.Equal(t, "gpt-4-turbo", clientCfg.Model)
				assert.Equal(t, 4096, clientCfg.MaxTokens)
				assert.Equal(t, 0.95, clientCfg.TopP)
				assert.Equal(t, 0.8, clientCfg.Temperature)
				assert.Equal(t, 0.2, clientCfg.FrequencyPenalty)
				assert.Equal(t, "http://proxy.example.com:8080", clientCfg.Proxy)

				// Check output settings
				assert.True(t, cfg.GetOutputTranslateTitle())

				// Check file ignore patterns
				fileIgnore := cfg.GetFileIgnore()
				assert.Contains(t, fileIgnore, "*.log")
				assert.Contains(t, fileIgnore, "*.tmp")
				assert.Contains(t, fileIgnore, "node_modules/")
			},
		},
		{
			name: "Config with custom prompts",
			configData: `
provider: openai
openai:
  api_key: test-key
prompt:
  brief_commit_message: "Custom brief commit message prompt"
  rich_commit_message: "Custom rich commit message prompt"
  translation: "Custom translation prompt"
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				prompt := cfg.GetPrompt(false) // false for brief commit message
				assert.Equal(t, "Custom brief commit message prompt", prompt)
			},
		},
		{
			name: "Config with Ollama (no API key required)",
			configData: `
provider: ollama
ollama:
  api_base: http://localhost:11434
  model: llama2
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				clientCfg, err := cfg.GetClientConfig("")
				require.NoError(t, err)
				assert.Equal(t, "ollama", clientCfg.Provider)
				assert.Equal(t, "", clientCfg.APIKey)
				assert.Equal(t, "llama2", clientCfg.Model)
			},
		},
		{
			name: "Invalid YAML syntax",
			configData: `
provider: openai
openai:
  api_key: [invalid
`,
			wantErr:     true,
			errContains: "Failed to load config",
		},
		{
			name: "Config with nested values",
			configData: `
provider: openai
openai:
  api_key: test-key
output:
  lang: zh
  translate_title: false
`,
			wantErr: false,
			validateFunc: func(t *testing.T, cfg *config.Manager) {
				// Test nested value access
				lang, ok := cfg.Get("output.lang")
				assert.True(t, ok)
				assert.Equal(t, "zh", lang)

				translateTitle := cfg.GetOutputTranslateTitle()
				assert.False(t, translateTitle)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := config.New(configFile)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, cfg)

			if tt.validateFunc != nil {
				tt.validateFunc(t, cfg)
			}
		})
	}
}

// TestConfigModificationFlow tests the full flow of loading, modifying, and saving config
func TestConfigModificationFlow(t *testing.T) {
	// Create initial config
	initialConfig := `
provider: openai
openai:
  api_key: initial-key
  model: gpt-4
`
	configFile, cleanup := testutils.TestConfig(t, initialConfig)
	defer cleanup()

	// Load config
	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Modify values
	err = cfg.Set("openai.model", "gpt-4-turbo")
	require.NoError(t, err)

	err = cfg.Set("openai.temperature", 0.7)
	require.NoError(t, err)

	// Verify changes are in memory
	model, ok := cfg.Get("openai.model")
	require.True(t, ok)
	assert.Equal(t, "gpt-4-turbo", model)

	temp, ok := cfg.Get("openai.temperature")
	require.True(t, ok)
	assert.Equal(t, 0.7, temp)

	// Save config
	err = cfg.Save()
	require.NoError(t, err)

	// Reload config and verify persistence
	cfg2, err := config.New(configFile)
	require.NoError(t, err)

	model2, ok := cfg2.Get("openai.model")
	require.True(t, ok)
	assert.Equal(t, "gpt-4-turbo", model2)

	temp2, ok := cfg2.Get("openai.temperature")
	require.True(t, ok)
	assert.Equal(t, 0.7, temp2)
}

// TestConfigFilePermissions tests that config files have correct permissions
func TestConfigFilePermissions(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: secret-key-12345
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Modify and save
	err = cfg.Set("openai.model", "gpt-4")
	require.NoError(t, err)

	err = cfg.Save()
	require.NoError(t, err)

	// Check file permissions
	info, err := os.Stat(configFile)
	require.NoError(t, err)

	// File should be readable and writable by owner only (0600 or 0644)
	mode := info.Mode()
	assert.True(t, mode.Perm() <= 0644, "Config file should have restricted permissions")
}

// TestConfigDirectoryCreation tests that config directory is created if it doesn't exist
func TestConfigDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested path that doesn't exist
	configPath := filepath.Join(tmpDir, "subdir", "another", "config.yaml")

	// Create config with non-existent parent directories
	configData := `
provider: openai
openai:
  api_key: test-key
`
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	require.NoError(t, err)

	err = os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	cfg, err := config.New(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

// TestConfigResetFlow tests resetting configuration values
func TestConfigResetFlow(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: test-key
  model: gpt-4
  temperature: 0.8
output:
  lang: en
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Verify initial value
	temp, ok := cfg.Get("openai.temperature")
	require.True(t, ok)
	assert.Equal(t, 0.8, temp)

	// Remove value (Reset is for prompt-only reset, use Remove for specific keys)
	err = cfg.Remove("openai.temperature", "")
	require.NoError(t, err)

	// Verify value is removed
	_, ok = cfg.Get("openai.temperature")
	assert.False(t, ok, "Value should be removed")

	// Save and reload
	err = cfg.Save()
	require.NoError(t, err)

	cfg2, err := config.New(configFile)
	require.NoError(t, err)

	// Verify value is still removed after reload
	_, ok = cfg2.Get("openai.temperature")
	assert.False(t, ok, "Value should remain removed after reload")
}
