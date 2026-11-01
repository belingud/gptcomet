package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GenerateOptions contains the settings for the generate command.
type GenerateOptions struct {
	CommonOptions
	RepoPath    string
	Rich        bool
	ConfigPath  string
	SecretsMode string
}

// commitMessageGenerator is the slice of CommitService that `generate` needs.
// It exists so tests can inject a fake without a provider or a Git repo.
type commitMessageGenerator interface {
	GenerateMessage() (string, error)
}

// newGenerateService builds the generator used by the generate command. It is a
// package-level variable so tests can replace it.
var newGenerateService = func(options CommitOptions) (commitMessageGenerator, error) {
	return NewCommitService(options)
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

			service, err := newGenerateService(CommitOptions{
				CommonOptions: options.CommonOptions,
				RepoPath:      options.RepoPath,
				Rich:          options.Rich,
				ConfigPath:    options.ConfigPath,
				SecretsMode:   options.SecretsMode,
			})
			if err != nil {
				return err
			}

			message, err := service.GenerateMessage()
			if err != nil {
				return err
			}

			// stdout carries only the message; status/usage output is written
			// to stderr by the lower layers.
			fmt.Fprintln(cmd.OutOrStdout(), message)
			return nil
		},
	}

	generalFlags := pflag.NewFlagSet("General Flag", pflag.ExitOnError)
	generalFlags.StringVar(&options.RepoPath, "repo", ".", "Repository path")
	generalFlags.BoolVarP(&options.Rich, "rich", "r", false, "Generate rich commit message with details")
	generalFlags.StringVar(&options.SecretsMode, "secrets", "", "Sensitive data scan mode: off, warn, redact, block (default from config)")

	advancedFlags := pflag.NewFlagSet("Overwrite Flag", pflag.ExitOnError)
	AddAdvancedAPIFlags(advancedFlags, &options.CommonOptions)

	cmd.Flags().AddFlagSet(generalFlags)
	cmd.Flags().AddFlagSet(advancedFlags)
	cmd.Flags().SetInterspersed(false)
	SetAdvancedHelpFunc(cmd, generalFlags, advancedFlags)

	return cmd
}
