package config

import (
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/belingud/gptcomet/pkg/config/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Validate(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	tests := []struct {
		name       string
		configData string
		wantIssues []string
		wantValid  bool
	}{
		{
			name: "valid configuration",
			configData: `
provider: openai
openai:
  api_key: test-key
output:
  lang: en
  review_lang: zh-cn
  markdown_theme: dark
  translate_title: false
file_ignore:
  - "*.lock"
prompt:
  review: "review this"
`,
			wantValid: true,
		},
		{
			name:       "missing provider",
			configData: `{}`,
			wantIssues: []string{"provider is not set"},
		},
		{
			name: "provider section missing",
			configData: `
provider: openai
`,
			wantIssues: []string{`provider "openai" has no configuration section`},
		},
		{
			name: "missing api key",
			configData: `
provider: openai
openai: {}
`,
			wantIssues: []string{"api_key is not set"},
		},
		{
			name: "invalid output language",
			configData: `
provider: openai
openai:
  api_key: test-key
output:
  lang: xx
`,
			wantIssues: []string{"output.lang"},
		},
		{
			name: "invalid markdown theme",
			configData: `
provider: openai
openai:
  api_key: test-key
output:
  markdown_theme: neon
`,
			wantIssues: []string{"output.markdown_theme"},
		},
		{
			name: "invalid translate_title type",
			configData: `
provider: openai
openai:
  api_key: test-key
output:
  translate_title: "yes"
`,
			wantIssues: []string{"output.translate_title"},
		},
		{
			name: "invalid retries",
			configData: `
provider: openai
openai:
  api_key: test-key
  retries: 0
`,
			wantIssues: []string{"retries must be >= 1"},
		},
		{
			name: "invalid top_p",
			configData: `
provider: openai
openai:
  api_key: test-key
  top_p: 1.5
`,
			wantIssues: []string{"top_p must be between"},
		},
		{
			name: "invalid file_ignore type",
			configData: `
provider: openai
openai:
  api_key: test-key
file_ignore: "*.lock"
`,
			wantIssues: []string{"file_ignore must be a list of strings"},
		},
		{
			name: "empty prompt",
			configData: `
provider: openai
openai:
  api_key: test-key
prompt:
  review: ""
`,
			wantIssues: []string{"prompt.review must be a non-empty string"},
		},
		{
			name: "azure missing deployment",
			configData: `
provider: azure
azure:
  api_key: test-key
  api_version: 2024-02-15-preview
`,
			wantIssues: []string{`provider "azure" requires deployment_name`},
		},
		{
			name: "vertex missing project and location",
			configData: `
provider: vertex
vertex:
  api_key: test-key
`,
			wantIssues: []string{`provider "vertex" requires project_id`, `provider "vertex" requires location`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile, cleanup := testutils.TestConfig(t, tt.configData)
			defer cleanup()

			cfg, err := New(configFile)
			require.NoError(t, err)

			got := cfg.Validate()
			if tt.wantValid {
				assert.Empty(t, got)
			}
			issues := strings.Join(got, "\n")
			for _, want := range tt.wantIssues {
				assert.Contains(t, issues, want)
			}
		})
	}
}

func TestManager_Validate_EnvAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")

	configFile, cleanup := testutils.TestConfig(t, `
provider: openai
openai: {}
`)
	defer cleanup()

	cfg, err := New(configFile)
	require.NoError(t, err)
	assert.Empty(t, cfg.Validate())
}

func TestManager_Validate_DefaultConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")

	cfg := &Manager{config: defaults.DefaultConfig}
	assert.Empty(t, cfg.Validate())
}

func TestIsValidMarkdownTheme(t *testing.T) {
	for _, theme := range []string{"auto", "ascii", "dark", "dracula", "light", "tokyo-night", "notty", "pink"} {
		assert.True(t, IsValidMarkdownTheme(theme), theme)
	}
	assert.False(t, IsValidMarkdownTheme("neon"))
	assert.False(t, IsValidMarkdownTheme(""))
}
