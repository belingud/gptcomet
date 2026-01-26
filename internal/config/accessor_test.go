package config

import (
	"sort"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Get(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		key        string
		wantValue  interface{}
		wantOK     bool
	}{
		{
			name: "Get root level key",
			configData: `
provider: openai
`,
			key:       "provider",
			wantValue: "openai",
			wantOK:    true,
		},
		{
			name: "Get nested key",
			configData: `
openai:
  api_key: test-key
  model: gpt-4
`,
			key:       "openai.model",
			wantValue: "gpt-4",
			wantOK:    true,
		},
		{
			name:       "Key not found",
			configData: `{}`,
			key:        "nonexistent.key",
			wantValue:  nil,
			wantOK:     false,
		},
		{
			name: "Get deep nested key",
			configData: `
output:
  lang: en
  translate_title: true
`,
			key:       "output.translate_title",
			wantValue: true,
			wantOK:    true,
		},
		{
			name: "Get list value",
			configData: `
file_ignore:
  - "*.log"
  - "*.tmp"
`,
			key:       "file_ignore",
			wantValue: []interface{}{"*.log", "*.tmp"},
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			value, ok := cfg.Get(tt.key)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestManager_GetWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		key          string
		defaultValue interface{}
		wantValue    interface{}
	}{
		{
			name:         "Key exists - return value",
			configData:   `provider: openai`,
			key:          "provider",
			defaultValue: "default-provider",
			wantValue:    "openai",
		},
		{
			name:         "Key not found - return default",
			configData:   `{}`,
			key:          "nonexistent.key",
			defaultValue: "default-value",
			wantValue:    "default-value",
		},
		{
			name:         "Get nested with default",
			configData:   `output: {lang: en}`,
			key:          "output.translate_title",
			defaultValue: false,
			wantValue:    false,
		},
		{
			name:         "Nil default value",
			configData:   `{}`,
			key:          "missing",
			defaultValue: nil,
			wantValue:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			value := cfg.GetWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestManager_GetNestedValue(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		keys       []string
		wantValue  interface{}
		wantOK     bool
	}{
		{
			name:       "Single level key",
			configData: `provider: openai`,
			keys:       []string{"provider"},
			wantValue:  "openai",
			wantOK:     true,
		},
		{
			name: "Two level key",
			configData: `
openai:
  api_key: test-key
`,
			keys:       []string{"openai", "api_key"},
			wantValue:  "test-key",
			wantOK:     true,
		},
		{
			name:       "Key not found",
			configData: `{}`,
			keys:       []string{"missing", "key"},
			wantValue:  nil,
			wantOK:     false,
		},
		{
			name: "Deep nested key",
			configData: `
level1:
  level2:
    level3:
      value: deep
`,
			keys:       []string{"level1", "level2", "level3", "value"},
			wantValue:  "deep",
			wantOK:     true,
		},
		{
			name: "Partial path exists",
			configData: `
openai:
  api_key: test-key
`,
			keys:   []string{"openai", "missing"},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			value, ok := cfg.GetNestedValue(tt.keys)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestManager_SetNestedValue(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		keys       []string
		value      interface{}
		verifyKeys []string
		wantValue  interface{}
	}{
		{
			name:       "Set single level value",
			configData: `{}`,
			keys:       []string{"newkey"},
			value:      "newvalue",
			verifyKeys: []string{"newkey"},
			wantValue:  "newvalue",
		},
		{
			name:       "Set nested value - creates intermediate maps",
			configData: `{}`,
			keys:       []string{"level1", "level2", "value"},
			value:      "nested",
			verifyKeys: []string{"level1", "level2", "value"},
			wantValue:  "nested",
		},
		{
			name: "Update existing nested value",
			configData: `
openai:
  api_key: old-key
`,
			keys:       []string{"openai", "api_key"},
			value:      "new-key",
			verifyKeys: []string{"openai", "api_key"},
			wantValue:  "new-key",
		},
		{
			name: "Set boolean value",
			configData: `
output:
  lang: en
`,
			keys:       []string{"output", "translate_title"},
			value:      true,
			verifyKeys: []string{"output", "translate_title"},
			wantValue:  true,
		},
		{
			name: "Set list value",
			configData: `{}`,
			keys:       []string{"file_ignore"},
			value:      []interface{}{"*.log", "*.tmp"},
			verifyKeys: []string{"file_ignore"},
			wantValue:  []interface{}{"*.log", "*.tmp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			cfg.SetNestedValue(tt.keys, tt.value)

			// Verify the value was set
			value, ok := cfg.GetNestedValue(tt.verifyKeys)
			require.True(t, ok)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestManager_GetSupportedKeys(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		checkKeys  []string // Keys that must be present
	}{
		{
			name:       "Default provider (openai)",
			configData: `{}`,
			checkKeys: []string{
				"provider",
				"file_ignore",
				"output.lang",
				"output.translate_title",
				"<provider>.api_key",
				"<provider>.model",
				"prompt.brief_commit_message",
				"console.verbose",
			},
		},
		{
			name: "Custom provider",
			configData: `
provider: anthropic
anthropic:
  api_key: test-key
`,
			checkKeys: []string{
				"provider",
				"<provider>.api_key",
				"<provider>.model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			keys := cfg.GetSupportedKeys()

			// Check that expected keys are present
			for _, checkKey := range tt.checkKeys {
				assert.Contains(t, keys, checkKey, "Key %s should be in supported keys", checkKey)
			}

			// Verify keys are sorted
			assert.True(t, sort.IsSorted(sort.StringSlice(keys)))

			// Verify <provider> placeholder is used instead of actual provider name
			if tt.configData != "" {
				for _, key := range keys {
					if key == "<provider>" {
						continue
					}
					// Check that actual provider name is not present
					assert.NotContains(t, key, "anthropic", "Should use <provider> placeholder, not actual provider name")
				}
			}
		})
	}
}

func TestManager_GetPrompt(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		isRich     bool
		wantDefault bool
	}{
		{
			name:       "No prompt config - returns brief default",
			configData: `{}`,
			isRich:     false,
			wantDefault: true,
		},
		{
			name:       "No prompt config - returns rich default",
			configData: `{}`,
			isRich:     true,
			wantDefault: true,
		},
		{
			name: "Custom brief prompt",
			configData: `
prompt:
  brief_commit_message: "Custom brief prompt"
  rich_commit_message: "Custom rich prompt"
`,
			isRich:     false,
			wantDefault: false,
		},
		{
			name: "Custom rich prompt",
			configData: `
prompt:
  brief_commit_message: "Custom brief prompt"
  rich_commit_message: "Custom rich prompt"
`,
			isRich:     true,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			prompt := cfg.GetPrompt(tt.isRich)
			assert.NotEmpty(t, prompt)

			if tt.wantDefault {
				// Should return default prompt
				assert.Contains(t, prompt, "commit")
			} else {
				// Should return custom prompt
				if tt.isRich {
					assert.Equal(t, "Custom rich prompt", prompt)
				} else {
					assert.Equal(t, "Custom brief prompt", prompt)
				}
			}
		})
	}
}

func TestManager_GetReviewPrompt(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		wantDefault bool
	}{
		{
			name:       "No prompt config - returns default",
			configData: `{}`,
			wantDefault: true,
		},
		{
			name: "No review prompt in config",
			configData: `
prompt:
  brief_commit_message: "Brief prompt"
`,
			wantDefault: true,
		},
		{
			name: "Custom review prompt",
			configData: `
prompt:
  review: "Custom review prompt"
`,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			prompt := cfg.GetReviewPrompt()
			assert.NotEmpty(t, prompt)

			if tt.wantDefault {
				// Should return default review prompt
				assert.Contains(t, prompt, "review")
			} else {
				// Should return custom prompt
				assert.Equal(t, "Custom review prompt", prompt)
			}
		})
	}
}

func TestManager_GetTranslationPrompt(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		wantDefault bool
	}{
		{
			name:       "No prompt config - returns default",
			configData: `{}`,
			wantDefault: true,
		},
		{
			name: "No translation prompt in config",
			configData: `
prompt:
  review: "Review prompt"
`,
			wantDefault: true,
		},
		{
			name: "Custom translation prompt",
			configData: `
prompt:
  translation: "Custom translation prompt"
`,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			prompt := cfg.GetTranslationPrompt()
			assert.NotEmpty(t, prompt)

			if tt.wantDefault {
				// Should return default translation prompt
				assert.Contains(t, prompt, "translation")
			} else {
				// Should return custom prompt
				assert.Equal(t, "Custom translation prompt", prompt)
			}
		})
	}
}

func TestManager_GetFileIgnore(t *testing.T) {
	tests := []struct {
		name       string
		configData string
		wantNil    bool
		wantLen    int
		wantContains []string
	}{
		{
			name:       "No file_ignore config",
			configData: `{}`,
			wantNil:    true,
		},
		{
			name: "Empty file_ignore list",
			configData: `
file_ignore: []
`,
			wantNil: false,
			wantLen: 0,
		},
		{
			name: "File ignore with patterns",
			configData: `
file_ignore:
  - "*.log"
  - "*.tmp"
  - "dist/"
`,
			wantNil: false,
			wantLen: 3,
			wantContains: []string{"*.log", "*.tmp", "dist/"},
		},
		{
			name: "File ignore with mixed types (only strings returned)",
			configData: `
file_ignore:
  - "*.log"
  - 123
  - "dist/"
`,
			wantNil: false,
			wantLen: 2,
			wantContains: []string{"*.log", "dist/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			patterns := cfg.GetFileIgnore()

			if tt.wantNil {
				assert.Nil(t, patterns)
			} else {
				require.NotNil(t, patterns)
				assert.Equal(t, tt.wantLen, len(patterns))
				for _, wantPattern := range tt.wantContains {
					assert.Contains(t, patterns, wantPattern)
				}
			}
		})
	}
}
