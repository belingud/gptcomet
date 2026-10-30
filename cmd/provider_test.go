package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/internal/provider"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newProviderTestManager creates a config manager backed by a temporary config
// file and registers it on the command's context.
func newProviderTestManager(t *testing.T, content string) (*config.Manager, string) {
	t.Helper()

	configFile, cleanup := testutils.TestConfig(t, content)
	t.Cleanup(cleanup)

	manager, err := config.New(configFile)
	require.NoError(t, err)

	return manager, configFile
}

// withRealProviderRegistry isolates the global LLM registry for tests that
// enumerate providers. Other tests in this package register MockLLM-based
// providers without a GetRequiredConfig expectation; the provider inventory
// asks every registered provider for its defaults, so those mocks would fail.
// The previous registry contents are restored when the test finishes.
func withRealProviderRegistry(t *testing.T) {
	t.Helper()

	type registration struct {
		name        string
		constructor llm.ProviderConstructor
	}
	existing := make([]registration, 0, len(llm.GetProviders()))
	for _, name := range llm.GetProviders() {
		if constructor, ok := llm.GetProviderConstructor(name); ok {
			existing = append(existing, registration{name: name, constructor: constructor})
		}
	}

	llm.ResetRegistry()
	t.Cleanup(func() {
		llm.ResetRegistry()
		for _, entry := range existing {
			_ = llm.RegisterProvider(entry.name, entry.constructor)
		}
	})

	builtins := map[string]llm.ProviderConstructor{
		"openai":   func(c *types.ClientConfig) llm.LLM { return llm.NewOpenAILLM(c) },
		"claude":   func(c *types.ClientConfig) llm.LLM { return llm.NewClaudeLLM(c) },
		"deepseek": func(c *types.ClientConfig) llm.LLM { return llm.NewDeepSeekLLM(c) },
		"ollama":   func(c *types.ClientConfig) llm.LLM { return llm.NewOllamaLLM(c) },
	}
	for name, constructor := range builtins {
		require.NoError(t, llm.RegisterProvider(name, constructor))
	}
}

func TestNewProviderCmd(t *testing.T) {
	cmd := NewProviderCmd()
	require.NotNil(t, cmd)

	assert.Equal(t, "provider", cmd.Use)
	assert.Nil(t, cmd.RunE)

	wantSubcommands := map[string]bool{
		"list": false,
		"use":  false,
		"test": false,
	}
	for _, sub := range cmd.Commands() {
		if _, ok := wantSubcommands[sub.Name()]; ok {
			wantSubcommands[sub.Name()] = true
		}
	}
	for name, found := range wantSubcommands {
		assert.True(t, found, "expected subcommand %q", name)
	}
}

func TestProviderListCmd(t *testing.T) {
	withRealProviderRegistry(t)
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("MYPROXY_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: super-secret-key-123456
  model: gpt-custom
myproxy:
  api_base: https://proxy.example/v1
  model: proxy-model
`)

	cmd := newProviderListCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, cmd.Execute())

	output := buf.String()
	for _, want := range []string{
		"Current provider: openai",
		"model:    gpt-custom (config)",
		"api key:  config (supe***",
		"PROVIDER",
		"* openai",
		"myproxy",
		"Switch providers with 'gmsg provider use <name>'.",
		"gmsg newprovider",
	} {
		assert.Contains(t, output, want)
	}
	assert.NotContains(t, output, "super-secret-key-123456", "full API key must never be printed")
}

func TestProviderListCmd_NoProviderSelected(t *testing.T) {
	withRealProviderRegistry(t)
	t.Setenv("OPENAI_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
output:
  lang: en
`)

	cmd := newProviderListCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "Current provider: (not set)")
}

func TestProviderUseCmd_SwitchesAndSeeds(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")

	manager, configPath := newProviderTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	cmd := newProviderUseCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"deepseek"})

	require.NoError(t, cmd.Execute())

	output := buf.String()
	assert.Contains(t, output, "Active provider: deepseek")
	assert.Contains(t, output, "Created default configuration for 'deepseek'")
	assert.Contains(t, output, "API key for 'deepseek' is not configured")

	reloaded, err := config.New(configPath)
	require.NoError(t, err)
	current, ok := reloaded.Get("provider")
	require.True(t, ok)
	assert.Equal(t, "deepseek", current)

	raw, ok := reloaded.GetNestedValue([]string{"deepseek"})
	require.True(t, ok)
	_, ok = raw.(map[string]interface{})
	assert.True(t, ok, "deepseek configuration block should be persisted")
}

func TestProviderUseCmd_UnknownProvider(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	cmd := newProviderUseCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"no-such-provider"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown Provider")

	current, ok := manager.Get("provider")
	require.True(t, ok)
	assert.Equal(t, "openai", current, "provider must stay unchanged on failure")
}

func TestProviderUseCmd_CustomProviderWarning(t *testing.T) {
	t.Setenv("MYPROXY_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: openai-key
myproxy:
  api_base: https://proxy.example/v1
  model: proxy-model
`)

	cmd := newProviderUseCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"myproxy"})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "OpenAI-compatible fallback")
}

func TestProviderUseCmd_AlreadyActive(t *testing.T) {
	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	cmd := newProviderUseCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"openai"})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "already active")
}

func TestProviderUseCmd_ReportsMissingFields(t *testing.T) {
	t.Setenv("AZURE_API_KEY", "")

	manager, _ := newProviderTestManager(t, `
provider: openai
openai:
  api_key: openai-key
`)

	cmd := newProviderUseCmd()
	cmd.SetContext(context.Background())
	setConfigManager(cmd, manager)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"azure"})

	require.NoError(t, cmd.Execute())
	output := buf.String()
	assert.Contains(t, output, "still needs:")
	assert.Contains(t, output, "deployment_name")
}

func TestFormatKeySource(t *testing.T) {
	tests := []struct {
		name string
		info provider.Info
		want string
	}{
		{
			name: "config with masked key",
			info: provider.Info{KeySource: config.APIKeySourceConfig, KeyMasked: "sk-ab***"},
			want: "config (sk-ab***)",
		},
		{
			name: "environment",
			info: provider.Info{KeySource: config.APIKeySourceEnv, KeyEnvVar: "OPENAI_API_KEY"},
			want: "env OPENAI_API_KEY",
		},
		{
			name: "not required",
			info: provider.Info{KeySource: config.APIKeySourceNotRequired},
			want: "not required",
		},
		{
			name: "missing",
			info: provider.Info{KeySource: config.APIKeySourceMissing},
			want: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatKeySource(tt.info))
		})
	}
}
