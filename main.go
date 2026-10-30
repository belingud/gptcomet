package main

import (
	"os"

	"github.com/belingud/gptcomet/cmd"
	"github.com/belingud/gptcomet/internal/logger"

	"github.com/spf13/cobra"
)

var version = "2.8.0"

// main is the entry point of the GPTComet application. It initializes and configures the command-line interface
// using cobra. The following commands are available:
//
// - provider: List, switch, and test AI providers
// - newprovider: Configure a new provider interactively
// - commit: Generate commit messages using AI
// - config: Manage application configuration
// - update: Check and update to latest version
// - review: Review git changes and commit messages
// - generate: Generate a commit message and print it to stdout
// - hook: Install, remove, or run the prepare-commit-msg hook
//
// The root command supports the following persistent flags:
//
//	--debug, -d: Enable debug mode for verbose logging
//	--config, -c: Specify a custom config file path
//
// If command execution fails, the program exits with status code 1.
func main() {
	var (
		debugEnabled bool
		configPath   string
	)

	var rootCmd = &cobra.Command{
		Use:          "gmsg",
		Aliases:      []string{"gptcomet"},
		Short:        "GPTComet - AI-powered Git commit message generator and reviewer",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debugEnabled {
				logger.SetLevel(logger.DebugLevel)
				logger.Debug("Debug mode enabled")
			} else {
				logger.SetLevel(logger.InfoLevel)
			}
			if configPath != "" {
				logger.Debug("Using config file: %s", configPath)
			}
		},
	}

	// Add persistent flags to root command
	rootCmd.PersistentFlags().BoolVarP(&debugEnabled, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path")

	rootCmd.AddCommand(cmd.NewProviderCmd())      // provider
	rootCmd.AddCommand(cmd.NewProviderSetupCmd()) // newprovider
	rootCmd.AddCommand(cmd.NewCommitCmd())        // commit
	rootCmd.AddCommand(cmd.NewConfigCmd())        // config
	rootCmd.AddCommand(cmd.NewUpdateCmd(version)) // update
	rootCmd.AddCommand(cmd.NewReviewCmd())        // review
	rootCmd.AddCommand(cmd.NewGenerateCmd())      // generate
	rootCmd.AddCommand(cmd.NewHookCmd())          // hook

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
