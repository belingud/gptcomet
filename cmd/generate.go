package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GenerateOptions contains the settings for the generate command.
type GenerateOptions struct {
	CommonOptions
	RepoPath   string
	Rich       bool
	ConfigPath string
}

// NewGenerateCmd creates a command that prints a generated commit message to
// stdout without creating a commit. It is intended for scripts and editor
// integrations; only the message is written to stdout.
func NewGenerateCmd() *cobra.Command {
	options := GenerateOptions{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a commit message and print it to stdout",
		Long:  "Generate a commit message from the staged diff without creating a commit.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}
			options.ConfigPath = configPath

			service, err := NewCommitService(CommitOptions{
				CommonOptions: options.CommonOptions,
				RepoPath:      options.RepoPath,
				Rich:          options.Rich,
				ConfigPath:    options.ConfigPath,
			})
			if err != nil {
				return err
			}

			message, err := generateWithCleanStdout(service.GenerateMessage)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), message)
			return nil
		},
	}

	generalFlags := pflag.NewFlagSet("General Flag", pflag.ExitOnError)
	generalFlags.StringVar(&options.RepoPath, "repo", ".", "Repository path")
	generalFlags.BoolVarP(&options.Rich, "rich", "r", false, "Generate rich commit message with details")

	advancedFlags := pflag.NewFlagSet("Overwrite Flag", pflag.ExitOnError)
	AddAdvancedAPIFlags(advancedFlags, &options.CommonOptions)

	cmd.Flags().AddFlagSet(generalFlags)
	cmd.Flags().AddFlagSet(advancedFlags)
	cmd.Flags().SetInterspersed(false)
	SetAdvancedHelpFunc(cmd, generalFlags, advancedFlags)

	return cmd
}

// generateWithCleanStdout redirects package-level stdout writes to stderr while
// generation runs. Several providers still print request/usage lines directly
// to stdout; the generate contract must expose only the final message.
func generateWithCleanStdout(generate func() (string, error)) (string, error) {
	original := os.Stdout
	os.Stdout = os.Stderr
	defer func() { os.Stdout = original }()

	return generate()
}
