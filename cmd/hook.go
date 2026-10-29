package cmd

import (
	"context"
	"fmt"

	"github.com/belingud/gptcomet/internal/hook"
	"github.com/belingud/gptcomet/internal/logger"
	"github.com/spf13/cobra"
)

// NewHookCmd creates the parent command for managing the Git
// prepare-commit-msg hook.
func NewHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage the Git prepare-commit-msg hook",
	}

	cmd.AddCommand(
		newHookInstallCmd(),
		newHookUninstallCmd(),
		newHookStatusCmd(),
		newHookRunCmd(),
	)

	return cmd
}

func newHookManager() *hook.Manager {
	return hook.NewManager(func(_ context.Context, repoPath, configPath string) (string, error) {
		service, err := NewCommitService(CommitOptions{
			RepoPath:   repoPath,
			ConfigPath: configPath,
		})
		if err != nil {
			return "", err
		}
		return service.GenerateMessage()
	})
}

func newHookInstallCmd() *cobra.Command {
	var (
		repoPath string
		force    bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the prepare-commit-msg hook in a Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			hookPath, err := newHookManager().Install(hook.InstallOptions{
				RepoPath:   repoPath,
				ConfigPath: configPath,
				Force:      force,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Installed prepare-commit-msg hook at %s\n", hookPath)
			fmt.Fprintln(cmd.OutOrStdout(), "Run `git commit` to generate a message automatically.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use `gmsg hook uninstall` to remove it.")
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "repo", ".", "Repository path")
	cmd.Flags().BoolVar(&force, "force", false, "Back up and replace an existing prepare-commit-msg hook")
	return cmd
}

func newHookUninstallCmd() *cobra.Command {
	var repoPath string

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the gptcomet-managed prepare-commit-msg hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := newHookManager().Uninstall(repoPath)
			if err != nil {
				return err
			}

			if !result.Removed && !result.Restored {
				fmt.Fprintf(cmd.OutOrStdout(), "No gptcomet-managed hook found at %s; left it untouched.\n", result.HookPath)
				return nil
			}

			if result.Removed {
				fmt.Fprintf(cmd.OutOrStdout(), "Removed managed hook at %s\n", result.HookPath)
			}
			if result.Restored {
				fmt.Fprintf(cmd.OutOrStdout(), "Restored the previous hook from %s%s\n", result.HookPath, hook.BackupSuffix)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "repo", ".", "Repository path")
	return cmd
}

func newHookStatusCmd() *cobra.Command {
	var repoPath string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether the prepare-commit-msg hook is installed",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := newHookManager().Status(repoPath)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Hooks directory: %s\n", status.HooksDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Hook: %s (%s)\n", status.HookPath, status.State)
			if status.BackupPath == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Backup: none")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Backup: %s\n", status.BackupPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "repo", ".", "Repository path")
	return cmd
}

func newHookRunCmd() *cobra.Command {
	var repoPath string

	cmd := &cobra.Command{
		Use:    "run <message-file> [source] [sha]",
		Short:  "Run the prepare-commit-msg hook logic",
		Hidden: true,
		Args:   cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				logger.Debug("hook run: failed to get config path: %v", err)
				return nil
			}

			source := ""
			if len(args) > 1 {
				source = args[1]
			}

			err = newHookManager().Run(context.Background(), hook.RunInput{
				MessageFile: args[0],
				Source:      source,
				RepoPath:    repoPath,
				ConfigPath:  configPath,
			})
			if err != nil {
				// The hook must never block a commit. Log only in debug mode.
				logger.Debug("hook run skipped: %v", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&repoPath, "repo", ".", "Repository path")
	return cmd
}
