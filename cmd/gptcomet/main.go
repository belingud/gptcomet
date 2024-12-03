package main

import (
	"fmt"
	"github.com/belingud/gptcomet/cmd"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	var debug bool

	var rootCmd = &cobra.Command{
		Use:     "gptcomet",
		Short:   "GPTComet - AI-powered Git commit message generator",
		Version: version,
	}

	// Add debug flag to root command
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	rootCmd.AddCommand(cmd.NewProviderCmd())
	rootCmd.AddCommand(cmd.NewCommitCmd(&debug))
	rootCmd.AddCommand(cmd.NewConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
