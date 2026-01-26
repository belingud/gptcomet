package cmd

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAdvancedAPIFlags(t *testing.T) {
	tests := []struct {
		name         string
		setupOpts    *CommonOptions
		flagName     string
		flagValue    string
		wantSet      bool
		description  string
	}{
		{
			name:        "AddFlagsToEmptyOptions",
			setupOpts:   &CommonOptions{},
			flagName:    "api-base",
			flagValue:   "https://api.example.com",
			wantSet:     true,
			description: "Should set APIBase flag successfully",
		},
		{
			name:        "AddFlagsToExistingOptions",
			setupOpts:   &CommonOptions{Model: "gpt-4"},
			flagName:    "model",
			flagValue:   "gpt-3.5-turbo",
			wantSet:     true,
			description: "Should override existing model value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			AddAdvancedAPIFlags(flags, tt.setupOpts)

			err := flags.Set(tt.flagName, tt.flagValue)
			require.NoError(t, err, "Should be able to set flag")

			flag := flags.Lookup(tt.flagName)
			require.NotNil(t, flag, "Flag should exist")
			assert.Equal(t, tt.flagValue, flag.Value.String(), "Flag value should match")
		})
	}
}

func TestAddGeneralFlags(t *testing.T) {
	tests := []struct {
		name        string
		repoPath    string
		useSVN      bool
		description string
	}{
		{
			name:        "SetRepoPath",
			repoPath:    "/path/to/repo",
			useSVN:      false,
			description: "Should set repository path",
		},
		{
			name:        "SetSVNFlag",
			repoPath:    ".",
			useSVN:      true,
			description: "Should set SVN flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var repoPath string
			var useSVN bool
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			AddGeneralFlags(flags, &repoPath, &useSVN)

			err := flags.Set("repo", tt.repoPath)
			require.NoError(t, err, "Should be able to set repo flag")
			assert.Equal(t, tt.repoPath, repoPath, "Repo path should match")

			if tt.useSVN {
				err = flags.Set("svn", "true")
				require.NoError(t, err, "Should be able to set svn flag")
				assert.True(t, useSVN, "SVN flag should be true")
			}
		})
	}
}

func TestAddConfigFlag(t *testing.T) {
	var configPath string
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	AddConfigFlag(flags, &configPath)

	err := flags.Set("config", "/path/to/config.yaml")
	require.NoError(t, err, "Should be able to set config flag")
	assert.Equal(t, "/path/to/config.yaml", configPath, "Config path should match")
}

func TestApplyCommonOptions(t *testing.T) {
	tests := []struct {
		name         string
		opts         CommonOptions
		initialCfg   types.ClientConfig
		wantModified bool
		verifyFunc   func(t *testing.T, cfg *types.ClientConfig)
	}{
		{
			name: "ApplyAllOptions",
			opts: CommonOptions{
				APIBase:          "https://custom.api.com",
				APIKey:           "custom-key",
				MaxTokens:        2048,
				Retries:          5,
				Model:            "custom-model",
				AnswerPath:       "custom.answer",
				CompletionPath:   "/custom/completion",
				Proxy:            "http://proxy.com",
				FrequencyPenalty: 0.5,
				Temperature:      0.8,
				TopP:             0.95,
			},
			initialCfg: types.ClientConfig{
				APIBase:          "https://default.api.com",
				APIKey:           "default-key",
				MaxTokens:        1024,
				Retries:          3,
				Model:            "default-model",
				AnswerPath:       "default.answer",
				CompletionPath:   nil,
				Proxy:            "",
				FrequencyPenalty: 0.0,
				Temperature:      0.3,
				TopP:             1.0,
			},
			wantModified: true,
			verifyFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "https://custom.api.com", cfg.APIBase, "APIBase should be updated")
				assert.Equal(t, "custom-key", cfg.APIKey, "APIKey should be updated")
				assert.Equal(t, 2048, cfg.MaxTokens, "MaxTokens should be updated")
				assert.Equal(t, 5, cfg.Retries, "Retries should be updated")
				assert.Equal(t, "custom-model", cfg.Model, "Model should be updated")
				assert.Equal(t, "custom.answer", cfg.AnswerPath, "AnswerPath should be updated")
				assert.Equal(t, "/custom/completion", *cfg.CompletionPath, "CompletionPath should be updated")
				assert.Equal(t, "http://proxy.com", cfg.Proxy, "Proxy should be updated")
				assert.Equal(t, 0.5, cfg.FrequencyPenalty, "FrequencyPenalty should be updated")
				assert.Equal(t, 0.8, cfg.Temperature, "Temperature should be updated")
				assert.Equal(t, 0.95, cfg.TopP, "TopP should be updated")
			},
		},
		{
			name: "ApplyOnlyNonZeroValues",
			opts: CommonOptions{
				APIBase: "https://custom.api.com",
				// All other fields are zero values
			},
			initialCfg: types.ClientConfig{
				APIBase:    "https://default.api.com",
				APIKey:     "default-key",
				MaxTokens:  1024,
				Retries:    3,
				Model:      "default-model",
				Temperature: 0.7,
			},
			wantModified: true,
			verifyFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, "https://custom.api.com", cfg.APIBase, "APIBase should be updated")
				assert.Equal(t, "default-key", cfg.APIKey, "APIKey should remain unchanged")
				assert.Equal(t, 1024, cfg.MaxTokens, "MaxTokens should remain unchanged")
				assert.Equal(t, 3, cfg.Retries, "Retries should remain unchanged")
				assert.Equal(t, "default-model", cfg.Model, "Model should remain unchanged")
				assert.Equal(t, 0.7, cfg.Temperature, "Temperature should remain unchanged")
			},
		},
		{
			name: "ApplyWithZeroMaxTokens",
			opts: CommonOptions{
				MaxTokens: 0, // Zero value should not be applied
			},
			initialCfg: types.ClientConfig{
				MaxTokens: 1024,
			},
			wantModified: false,
			verifyFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 1024, cfg.MaxTokens, "MaxTokens should remain unchanged when zero")
			},
		},
		{
			name: "ApplyWithZeroTemperature",
			opts: CommonOptions{
				Temperature: 0, // Zero value should not be applied
			},
			initialCfg: types.ClientConfig{
				Temperature: 0.7,
			},
			wantModified: false,
			verifyFunc: func(t *testing.T, cfg *types.ClientConfig) {
				assert.Equal(t, 0.7, cfg.Temperature, "Temperature should remain unchanged when zero")
			},
		},
		{
			name: "ApplyWithCompletionPath",
			opts: CommonOptions{
				CompletionPath: "/custom/completion",
			},
			initialCfg: types.ClientConfig{
				CompletionPath: nil,
			},
			wantModified: true,
			verifyFunc: func(t *testing.T, cfg *types.ClientConfig) {
				require.NotNil(t, cfg.CompletionPath, "CompletionPath should be set")
				assert.Equal(t, "/custom/completion", *cfg.CompletionPath, "CompletionPath should match")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.initialCfg
			ApplyCommonOptions(&tt.opts, &cfg)

			if tt.verifyFunc != nil {
				tt.verifyFunc(t, &cfg)
			}
		})
	}
}

func TestSetAdvancedHelpFunc(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a test command for help output",
	}

	generalFlags := pflag.NewFlagSet("general", pflag.ContinueOnError)
	generalFlags.String("output", "json", "Output format")

	advancedFlags := pflag.NewFlagSet("advanced", pflag.ContinueOnError)
	advancedFlags.String("api-key", "", "API key")

	SetAdvancedHelpFunc(cmd, generalFlags, advancedFlags)

	// Verify that help function is set
	assert.NotNil(t, cmd.HelpFunc(), "Help function should be set")
}

func TestCommonOptions_DefaultValues(t *testing.T) {
	opts := CommonOptions{}

	// All fields should have zero values
	assert.Equal(t, "", opts.APIBase, "APIBase should be empty by default")
	assert.Equal(t, "", opts.APIKey, "APIKey should be empty by default")
	assert.Equal(t, 0, opts.MaxTokens, "MaxTokens should be 0 by default")
	assert.Equal(t, 0, opts.Retries, "Retries should be 0 by default")
	assert.Equal(t, "", opts.Model, "Model should be empty by default")
	assert.Equal(t, "", opts.AnswerPath, "AnswerPath should be empty by default")
	assert.Equal(t, "", opts.CompletionPath, "CompletionPath should be empty by default")
	assert.Equal(t, "", opts.Proxy, "Proxy should be empty by default")
	assert.Equal(t, float64(0), opts.FrequencyPenalty, "FrequencyPenalty should be 0 by default")
	assert.Equal(t, float64(0), opts.Temperature, "Temperature should be 0 by default")
	assert.Equal(t, float64(0), opts.TopP, "TopP should be 0 by default")
	assert.Equal(t, "", opts.Provider, "Provider should be empty by default")
}
