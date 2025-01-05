package cmd

import (
	"bytes"
	"testing"

	"github.com/belingud/go-gptcomet/internal/llm"
	"github.com/belingud/go-gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProviderCmd(t *testing.T) {
	cmd := NewProviderCmd()
	require.NotNil(t, cmd)

	// Test command basic properties
	assert.Equal(t, "newprovider", cmd.Use)
	assert.Equal(t, "Configure a new provider interactively", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Create a buffer to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set test environment
	t.Setenv("GPTCOMET_TEST", "1")

	// Register test provider
	llm.RegisterProvider("test-provider", func(config *types.ClientConfig) llm.LLM {
		return &mockLLM{name: "test-provider"}
	})

	// Execute command in non-interactive mode
	err := cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Available providers:")
	assert.Contains(t, output, "test-provider")
}

func TestProviderCmd_NonInteractive(t *testing.T) {
	cmd := NewProviderCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set test environment
	t.Setenv("GPTCOMET_TEST", "1")

	// Register test provider
	llm.RegisterProvider("test-provider", func(config *types.ClientConfig) llm.LLM {
		return &mockLLM{name: "test-provider"}
	})

	// Execute command
	err := cmd.Execute()
	require.NoError(t, err)

	// Check output
	output := buf.String()
	assert.Contains(t, output, "Available providers:")
	assert.Contains(t, output, "test-provider")
}

func TestProviderCmd_EmptyProvidersList(t *testing.T) {
	cmd := NewProviderCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set test environment
	t.Setenv("GPTCOMET_TEST", "1")

	// Do not register any providers, use the existing providers list
	// Only check if the command can execute successfully, do not check the providers list

	// Execute command
	err := cmd.Execute()
	require.NoError(t, err)

	// Check basic output format
	output := buf.String()
	assert.Contains(t, output, "Available providers:")
}

func TestProviderCmd_MultipleProviders(t *testing.T) {
	cmd := NewProviderCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set test environment
	t.Setenv("GPTCOMET_TEST", "1")

	// Register multiple test providers
	llm.RegisterProvider("test-provider1", func(config *types.ClientConfig) llm.LLM {
		return &mockLLM{name: "test-provider1"}
	})
	llm.RegisterProvider("test-provider2", func(config *types.ClientConfig) llm.LLM {
		return &mockLLM{name: "test-provider2"}
	})

	// Execute command
	err := cmd.Execute()
	require.NoError(t, err)

	// Check output contains all registered providers
	output := buf.String()
	assert.Contains(t, output, "Available providers:")
	assert.Contains(t, output, "test-provider1")
	assert.Contains(t, output, "test-provider2")
}
