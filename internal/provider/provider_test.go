package provider

import (
	"errors"
	"sort"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestManager creates a config manager backed by a temporary config file.
func newTestManager(t *testing.T, content string) (*config.Manager, string) {
	t.Helper()

	configFile, cleanup := testutils.TestConfig(t, content)
	t.Cleanup(cleanup)

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	return cfg, configFile
}

func findInfo(t *testing.T, infos []Info, name string) Info {
	t.Helper()

	for _, info := range infos {
		if info.Name == name {
			return info
		}
	}
	t.Fatalf("provider %q not found in list", name)
	return Info{}
}

func TestService_List(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("MYPROXY_API_KEY", "")
	t.Setenv("OLLAMA_API_KEY", "")

	cfg, _ := newTestManager(t, `
provider: claude
file_ignore:
  - "*.lock"
output:
  lang: en
console:
  verbose: true
openai:
  api_key: config-key-123456
  model: gpt-custom
claude:
  api_base: https://anthropic.example/v1
myproxy:
  api_base: https://proxy.example/v1
  model: proxy-model
`)

	svc := NewService(cfg)
	infos := svc.List()
	require.NotEmpty(t, infos)

	names := make([]string, 0, len(infos))
	activeCount := 0
	for _, info := range infos {
		names = append(names, info.Name)
		if info.Active {
			activeCount++
		}
	}

	assert.True(t, sort.StringsAreSorted(names), "provider names should be sorted: %v", names)
	assert.Equal(t, len(llm.GetProviders())+1, len(infos), "expected registered providers plus the custom one")
	assert.Equal(t, 1, activeCount, "exactly one provider should be active")

	for _, reserved := range []string{"output", "console", "file_ignore", "provider"} {
		assert.NotContains(t, names, reserved, "reserved root key %q must not be listed as a provider", reserved)
	}

	openai := findInfo(t, infos, "openai")
	assert.False(t, openai.Active)
	assert.True(t, openai.Registered)
	assert.True(t, openai.Configured)
	assert.Equal(t, "gpt-custom", openai.Model)
	assert.True(t, openai.ModelFromConfig)
	assert.Equal(t, config.APIKeySourceConfig, openai.KeySource)
	assert.Contains(t, openai.KeyMasked, "*")
	assert.NotContains(t, openai.KeyMasked, "config-key-123456")

	claude := findInfo(t, infos, "claude")
	assert.True(t, claude.Active)
	assert.Equal(t, "https://anthropic.example/v1", claude.APIBase)
	assert.True(t, claude.APIBaseFromConfig)
	assert.False(t, claude.ModelFromConfig)
	assert.Equal(t, providerDefault("claude", "model"), claude.Model)
	assert.Equal(t, config.APIKeySourceMissing, claude.KeySource)
	assert.Equal(t, "ANTHROPIC_API_KEY", claude.KeyEnvVar)

	custom := findInfo(t, infos, "myproxy")
	assert.False(t, custom.Registered)
	assert.True(t, custom.Configured)
	assert.Equal(t, "proxy-model", custom.Model)
	assert.Equal(t, config.APIKeySourceMissing, custom.KeySource)
	assert.Equal(t, "MYPROXY_API_KEY", custom.KeyEnvVar)

	ollama := findInfo(t, infos, "ollama")
	assert.Equal(t, config.APIKeySourceNotRequired, ollama.KeySource)
}

func TestService_List_EnvKeySource(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_base: https://api.openai.com/v1
`)
	t.Setenv("OPENAI_API_KEY", "env-only-key-98765")

	info := NewService(cfg).Get("openai")
	assert.Equal(t, config.APIKeySourceEnv, info.KeySource)
	assert.Equal(t, "OPENAI_API_KEY", info.KeyEnvVar)
	assert.Empty(t, info.KeyMasked)
}

func TestService_Use_SeedsRegisteredProviderBlock(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)
	t.Setenv("DEEPSEEK_API_KEY", "env-deepseek-key")

	svc := NewService(cfg)
	result, err := svc.Use("deepseek")
	require.NoError(t, err)

	assert.True(t, result.Seeded)
	assert.True(t, result.Registered)
	assert.False(t, result.AlreadyActive)
	assert.Equal(t, "deepseek", result.Provider)
	assert.Equal(t, config.APIKeySourceEnv, result.KeySource)
	assert.Empty(t, result.MissingFields)
	assert.Equal(t, "deepseek", svc.CurrentName())

	raw, ok := cfg.GetNestedValue([]string{"deepseek"})
	require.True(t, ok)
	section, ok := raw.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, section["model"])
	assert.NotEmpty(t, section["api_base"])
	assert.Equal(t, "", section["api_key"])

	clientConfig, err := cfg.GetClientConfig("")
	require.NoError(t, err)
	assert.Equal(t, "env-deepseek-key", clientConfig.APIKey)
	assert.Equal(t, section["model"], clientConfig.Model)
	assert.Equal(t, section["api_base"], clientConfig.APIBase)
}

func TestService_Use_PersistsAcrossReload(t *testing.T) {
	cfg, configPath := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)
	t.Setenv("DEEPSEEK_API_KEY", "env-deepseek-key")

	_, err := NewService(cfg).Use("deepseek")
	require.NoError(t, err)

	reloaded, err := config.New(configPath)
	require.NoError(t, err)
	assert.Equal(t, "deepseek", NewService(reloaded).CurrentName())

	raw, ok := reloaded.GetNestedValue([]string{"deepseek"})
	require.True(t, ok)
	_, ok = raw.(map[string]interface{})
	assert.True(t, ok, "seeded deepseek block should be persisted")
}

func TestService_Use_RejectsUnknownProvider(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)
	svc := NewService(cfg)

	_, err := svc.Use("no-such-provider")
	require.Error(t, err)

	var gptErr *gptcometerrors.GPTCometError
	require.True(t, errors.As(err, &gptErr))
	assert.Equal(t, gptcometerrors.ErrTitleUnknownProvider, gptErr.Title)
	assert.Contains(t, err.Error(), "gmsg provider list")
	assert.Contains(t, err.Error(), "gmsg newprovider")
	assert.Equal(t, "openai", svc.CurrentName(), "provider must not change on failure")
}

func TestService_Use_AcceptsConfiguredCustomProvider(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
myproxy:
  api_base: https://proxy.example/v1
  model: proxy-model
`)
	svc := NewService(cfg)

	result, err := svc.Use("myproxy")
	require.NoError(t, err)

	assert.Equal(t, "myproxy", result.Provider)
	assert.False(t, result.Registered)
	assert.False(t, result.Seeded)
	assert.Empty(t, result.MissingFields)
	assert.Equal(t, "myproxy", svc.CurrentName())
}

func TestService_Use_DoesNotRewriteExistingBlock(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
deepseek:
  api_base: https://custom.example/v1
  model: deepseek-custom
`)
	t.Setenv("DEEPSEEK_API_KEY", "")

	result, err := NewService(cfg).Use("deepseek")
	require.NoError(t, err)

	assert.False(t, result.Seeded)
	assert.Equal(t, config.APIKeySourceMissing, result.KeySource)

	raw, ok := cfg.GetNestedValue([]string{"deepseek"})
	require.True(t, ok)
	section, ok := raw.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://custom.example/v1", section["api_base"])
	assert.Equal(t, "deepseek-custom", section["model"])
}

func TestService_Use_ReportsMissingRequiredFields(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)
	t.Setenv("AZURE_API_KEY", "")

	result, err := NewService(cfg).Use("azure")
	require.NoError(t, err)

	assert.True(t, result.Seeded)
	assert.Equal(t, config.APIKeySourceMissing, result.KeySource)
	assert.Contains(t, result.MissingFields, "api_base")
	assert.Contains(t, result.MissingFields, "deployment_name")
}

func TestService_Use_KeylessProvider(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)
	svc := NewService(cfg)

	result, err := svc.Use("ollama")
	require.NoError(t, err)
	assert.Equal(t, config.APIKeySourceNotRequired, result.KeySource)

	clientConfig, err := cfg.GetClientConfig("")
	require.NoError(t, err)
	assert.Empty(t, clientConfig.APIKey)
}

func TestService_Use_AlreadyActive(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	result, err := NewService(cfg).Use("openai")
	require.NoError(t, err)

	assert.True(t, result.AlreadyActive)
	assert.False(t, result.Seeded)
	assert.Equal(t, config.APIKeySourceConfig, result.KeySource)
}

func TestService_Use_NormalizesBuiltInProviderCase(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	result, err := NewService(cfg).Use("OpenAI")
	require.NoError(t, err)
	assert.Equal(t, "openai", result.Provider)
}

func TestService_NormalizeName(t *testing.T) {
	cfg, _ := newTestManager(t, `
provider: openai
openai:
  api_key: openai-key
myproxy:
  api_base: https://proxy.example/v1
  model: proxy-model
`)
	svc := NewService(cfg)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "already canonical", input: "openai", want: "openai"},
		{name: "built-in provider is lowercased", input: "OpenAI", want: "openai"},
		{name: "whitespace is trimmed", input: "  deepseek  ", want: "deepseek"},
		{name: "custom provider keeps its spelling when exact", input: "myproxy", want: "myproxy"},
		{name: "custom provider is lowercased when that matches", input: "MyProxy", want: "myproxy"},
		{name: "unknown name is returned trimmed", input: "  mystery  ", want: "mystery"},
		{name: "empty", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, svc.NormalizeName(tt.input))
		})
	}
}
