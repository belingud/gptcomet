package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/belingud/go-gptcomet/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigCmd(t *testing.T) {
	// Create a new root command for testing
	rootCmd := &cobra.Command{
		Use: "root",
	}

	// Create a new config command
	configCmd := NewConfigCmd()
	rootCmd.AddCommand(configCmd)

	// Create a dummy config file for testing
	configContent := `
provider: openai
openai:
  api_key: "test_api_key"
`
	configPath, cleanup := testutils.TestConfig(t, configContent)
	defer cleanup()

	// Set the config flag for the root command
	rootCmd.PersistentFlags().StringP("config", "c", configPath, "Config file path")

	// Test cases for config subcommands
	testCases := []struct {
		name        string
		args        []string
		expectedOut string
		expectedErr string
	}{
		{
			name:        "get existing key",
			args:        []string{"config", "get", "provider"},
			expectedOut: "Value for key 'provider':\n\"openai\"\n",
		},
		{
			name:        "get non-existing key",
			args:        []string{"config", "get", "nonexistent"},
			expectedErr: "config key not found: nonexistent",
		},
		{
			name:        "list config",
			args:        []string{"config", "list"},
			expectedOut: "provider: openai\nopenai:\n    api_base: https://api.openai.com/v1\n    api_key: sk-***\n    model: gpt-4o\n    retries: 2\n    proxy: \"\"\n    max_tokens: 2048\n    top_p: 0.7\n    temperature: 0.7\n    frequency_penalty: 0\n    extra_headers: '{}'\n    completion_path: /chat/completions\n    answer_path: choices.0.message.content\nfile_ignore:\n    - bun.lockb\n    - Cargo.lock\n    - composer.lock\n    - Gemfile.lock\n    - package-lock.json\n    - pnpm-lock.yaml\n    - poetry.lock\n    - yarn.lock\n    - pdm.lock\n    - Pipfile.lock\n    - '*.py[cod]'\n    - go.mod\n    - go.sum\n    - uv.lock\n    - README.md\n    - README.MD\n    - '*.md'\n    - '*.MD'\noutput:\n    lang: en\n    rich_template: <title>:<summary>\\n\\n<detail>\nconsole:\n    verbose: true\nanthropic:\n    api_base: https://api.anthropic.com\n    api_key: \"\"\n    model: claude-3.5-sonnet\n    retries: 2\n    proxy: \"\"\n    max_tokens: 2048\n    top_p: 0.7\n    temperature: 0.7\n    frequency_penalty: 0\n    extra_headers: '{}'\n    completion_path: /v1/messages\n    answer_path: content.0.text\n",
		},
		{
			name:        "reset config",
			args:        []string{"config", "reset"},
			expectedOut: "Configuration has been reset to default values\n",
		},
		{
			name:        "reset prompt config",
			args:        []string{"config", "reset", "--prompt"},
			expectedOut: "Prompt configuration has been reset to default values\n",
		},
		{
			name:        "set config value",
			args:        []string{"config", "set", "provider", "testprovider"},
			expectedOut: "Successfully set 'provider' to: testprovider\n",
		},
		{
			name:        "set nested config value",
			args:        []string{"config", "set", "output.lang", "zh-cn"},
			expectedOut: "Successfully set 'output.lang' to: zh-cn\n",
		},
		{
			name:        "set invalid nested config value",
			args:        []string{"config", "set", "output.lang", "invalid-lang"},
			expectedErr: "invalid language code: invalid-lang",
		},
		{
			name:        "get config path",
			args:        []string{"config", "path"},
			expectedOut: fmt.Sprintf("Configuration file path: %s\n", configPath),
		},
		{
			name:        "remove config key",
			args:        []string{"config", "remove", "provider"},
			expectedOut: "Successfully removed key 'provider'\n",
		},
		{
			name:        "remove non-existing config key",
			args:        []string{"config", "remove", "nonexistent"},
			expectedOut: "Key 'nonexistent' not found in configuration\n",
		},
		{
			name:        "remove value from list",
			args:        []string{"config", "remove", "file_ignore", "README.md"},
			expectedOut: "Successfully removed 'README.md' from 'file_ignore'\n",
		},
		{
			name:        "remove non-existing value from list",
			args:        []string{"config", "remove", "file_ignore", "nonexistent"},
			expectedOut: "Successfully removed 'nonexistent' from 'file_ignore'\n",
		},
		{
			name:        "append value to list",
			args:        []string{"config", "append", "file_ignore", "new_ignore.txt"},
			expectedOut: "Successfully appended 'new_ignore.txt' to 'file_ignore'\n",
		},
		{
			name:        "append value to non-list",
			args:        []string{"config", "append", "provider", "new_value"},
			expectedOut: "Warning: Key 'provider' exists but is not a list. It will be converted to a list.\nSuccessfully appended 'new_value' to 'provider'\n",
		},
		{
			name: "list supported config keys",
			args: []string{"config", "keys"},
			expectedOut: `Supported configuration keys:
  <provider>.api_base
  <provider>.api_key
  <provider>.answer_path
  <provider>.completion_path
  <provider>.extra_headers
  <provider>.frequency_penalty
  <provider>.max_tokens
  <provider>.model
  <provider>.proxy
  <provider>.retries
  <provider>.temperature
  <provider>.top_p
  console.verbose
  file_ignore
  output.lang
  output.rich_template
  prompt.brief_commit_message
  prompt.rich_commit_message
  prompt.translation
  provider
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Redirect stdout to capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// Set command arguments
			rootCmd.SetArgs(tc.args)

			// Execute the command
			err := rootCmd.Execute()

			// Check for expected error
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			// Check for expected output
			if tc.expectedOut != "" {
				assert.Equal(t, tc.expectedOut, buf.String())
			}
		})
	}
}

func TestConfigGet_MaskAPIKey(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
provider: openai
openai:
  api_key: "sk-or-v1-abcdefg"
`
	configPath, cleanup := testutils.TestConfig(t, configContent)
	defer cleanup()

	// Create a new config command
	cmd := NewConfigCmd()
	cmd.SetContext(cmd.Context())

	// Set the config flag for the root command
	cmd.Root().PersistentFlags().StringP("config", "c", configPath, "Config file path")

	// Redirect stdout to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set command arguments
	cmd.SetArgs([]string{"get", "openai.api_key"})

	// Execute the command
	err := cmd.Execute()
	require.NoError(t, err)

	// Check if the API key is masked
	expectedOutput := "Value for key 'openai.api_key':\n\"sk-or-v1-abc****\"\n"
	assert.Equal(t, expectedOutput, buf.String())
}

func TestConfigList_MaskAPIKeys(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
provider: openai
openai:
  api_key: "sk-or-v1-abcdefg"
`
	configPath, cleanup := testutils.TestConfig(t, configContent)
	defer cleanup()

	// Create a new config command
	cmd := NewConfigCmd()

	// Set the config flag for the root command
	cmd.Root().PersistentFlags().StringP("config", "c", configPath, "Config file path")

	// Redirect stdout to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Set command arguments
	cmd.SetArgs([]string{"list"})

	// Execute the command
	err := cmd.Execute()
	require.NoError(t, err)
	s := buf.String()

	// Check if the API key is masked in the output
	assert.Contains(t, s, "api_key: sk-or-v1-abc**")
}
