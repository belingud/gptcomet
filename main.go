package main

import (
	"fmt"
	"os"

	"github.com/belingud/gptcomet/cmd"
	"github.com/belingud/gptcomet/internal/debug"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	var debugEnabled bool

	var rootCmd = &cobra.Command{
		Use:     "gptcomet",
		Short:   "GPTComet - AI-powered Git commit message generator",
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug.Enable(debugEnabled)
			debug.Printf("Debug mode enabled")
		},
	}

	// Add debug flag to root command
	rootCmd.PersistentFlags().BoolVarP(&debugEnabled, "debug", "d", false, "Enable debug mode")

	rootCmd.AddCommand(cmd.NewProviderCmd())
	rootCmd.AddCommand(cmd.NewCommitCmd())
	rootCmd.AddCommand(cmd.NewConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
