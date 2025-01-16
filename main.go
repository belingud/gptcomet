package main

import (
	"os"

	"github.com/belingud/gptcomet/cmd"
	"github.com/belingud/gptcomet/internal/debug"

	"github.com/spf13/cobra"
)

var version = "0.3.0"

func main() {
	var (
		debugEnabled bool
		configPath   string
	)

	var rootCmd = &cobra.Command{
		Use:          "gptcomet",
		Short:        "GPTComet - AI-powered Git commit message generator and reviewer",
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

	rootCmd.AddCommand(cmd.NewProviderCmd())      // newprovider
	rootCmd.AddCommand(cmd.NewCommitCmd())        // commit
	rootCmd.AddCommand(cmd.NewConfigCmd())        // config
	rootCmd.AddCommand(cmd.NewUpdateCmd(version)) // update
	rootCmd.AddCommand(cmd.NewReviewCmd())        // review

	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
