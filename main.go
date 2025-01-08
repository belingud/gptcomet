package main

import (
	"os"

	"github.com/belingud/gptcomet/cmd"
	"github.com/belingud/gptcomet/internal/debug"

	"github.com/spf13/cobra"
)

var version = "0.1.9"

func main() {
	var (
		debugEnabled bool
		configPath   string
	)

	var rootCmd = &cobra.Command{
		Use:          "gptcomet",
		Short:        "GPTComet - AI-powered Git commit message generator",
		Version:      version,
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug.Enable(debugEnabled)
			debug.Printf("Debug mode enabled")
			if configPath != "" {
				debug.Printf("Using config file: %s", configPath)
			}
		},
	}

	// Add persistent flags to root command
	rootCmd.PersistentFlags().BoolVarP(&debugEnabled, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Config file path")

	rootCmd.AddCommand(cmd.NewProviderCmd())
	rootCmd.AddCommand(cmd.NewCommitCmd())
	rootCmd.AddCommand(cmd.NewConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
