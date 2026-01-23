package cmd

import (
	"fmt"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommonOptions contains shared API override configuration
// used across multiple commands (commit, review, etc.)
type CommonOptions struct {
	// API Configuration
	APIBase          string
	APIKey           string
	MaxTokens        int
	Retries          int
	Model            string
	AnswerPath       string
	CompletionPath   string
	Proxy            string
	FrequencyPenalty float64
	Temperature      float64
	TopP             float64
	Provider         string
}

// AddAdvancedAPIFlags adds all API override flags to a command
// These flags allow users to override configuration file settings
func AddAdvancedAPIFlags(flags *pflag.FlagSet, opts *CommonOptions) {
	flags.StringVar(&opts.APIBase, "api-base", "", "Override API base URL")
	flags.StringVar(&opts.APIKey, "api-key", "", "Override API key")
	flags.IntVar(&opts.MaxTokens, "max-tokens", 0, "Override maximum tokens")
	flags.IntVar(&opts.Retries, "retries", 0, "Override retry count")
	flags.StringVar(&opts.Model, "model", "", "Override model name")
	flags.StringVar(&opts.AnswerPath, "answer-path", "", "Override answer path")
	flags.StringVar(&opts.CompletionPath, "completion-path", "", "Override completion path")
	flags.StringVar(&opts.Proxy, "proxy", "", "Override proxy URL")
	flags.Float64Var(&opts.FrequencyPenalty, "frequency-penalty", 0, "Override frequency penalty")
	flags.Float64Var(&opts.Temperature, "temperature", 0, "Override temperature")
	flags.Float64Var(&opts.TopP, "top-p", 0, "Override top_p value")
	flags.StringVar(&opts.Provider, "provider", "", "Override AI provider (openai/deepseek)")
}

// AddGeneralFlags adds general operational flags to a command
// repoPath: pointer to repo path variable
// useSVN: pointer to SVN flag variable
func AddGeneralFlags(flags *pflag.FlagSet, repoPath *string, useSVN *bool) {
	flags.StringVarP(repoPath, "repo", "r", ".", "Path to the repository")
	flags.BoolVarP(useSVN, "svn", "v", false, "Use SVN instead of Git")
}

// AddConfigFlag adds the configuration file path flag
func AddConfigFlag(flags *pflag.FlagSet, configPath *string) {
	flags.StringVarP(configPath, "config", "c", "", "Path to the configuration file")
}

// ApplyCommonOptions applies the common API options to a client config
// This function applies non-zero values from opts to clientConfig
func ApplyCommonOptions(opts *CommonOptions, clientConfig *types.ClientConfig) {
	if opts.APIBase != "" {
		clientConfig.APIBase = opts.APIBase
	}
	if opts.APIKey != "" {
		clientConfig.APIKey = opts.APIKey
	}
	if opts.MaxTokens > 0 {
		clientConfig.MaxTokens = opts.MaxTokens
	}
	if opts.Retries > 0 {
		clientConfig.Retries = opts.Retries
	}
	if opts.Model != "" {
		clientConfig.Model = opts.Model
	}
	if opts.AnswerPath != "" {
		clientConfig.AnswerPath = opts.AnswerPath
	}
	if opts.CompletionPath != "" {
		clientConfig.CompletionPath = &opts.CompletionPath
	}
	if opts.Proxy != "" {
		clientConfig.Proxy = opts.Proxy
	}
	if opts.FrequencyPenalty != 0 {
		clientConfig.FrequencyPenalty = opts.FrequencyPenalty
	}
	if opts.Temperature != 0 {
		clientConfig.Temperature = opts.Temperature
	}
	if opts.TopP != 0 {
		clientConfig.TopP = opts.TopP
	}
}

// SetAdvancedHelpFunc sets a custom help function that organizes flags into groups
// generalFlags: primary operational flags
// advancedFlags: API override flags
func SetAdvancedHelpFunc(cmd *cobra.Command, generalFlags, advancedFlags *pflag.FlagSet) {
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		fmt.Println(c.Long)
		fmt.Println("\nUsage:")
		fmt.Printf("  %s\n", c.UseLine())
		fmt.Println("\nGeneral Flags:")
		generalFlags.PrintDefaults()
		fmt.Println("\nOverwrite Flags:")
		advancedFlags.PrintDefaults()
		fmt.Println()
		fmt.Println(`Global Flags:
  -d, --debug           Enable debug mode`)
	})
}
