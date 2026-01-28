package cmd

import (
	"fmt"
	"os"

	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/factory"
	"github.com/belingud/gptcomet/internal/git"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommitOptions contains the configuration settings for the commit operation.
type CommitOptions struct {
	CommonOptions
	RepoPath   string
	Rich       bool
	DryRun     bool
	UseSVN     bool
	AutoYes    bool
	ConfigPath string
	NoVerify   bool
}

// CommitService handles the logic for committing changes to version control
// while integrating with the GPT service for generating commit messages.
// It manages the interaction between the version control system,
// API client, and configuration settings.
type CommitService struct {
	vcs          git.VCS
	client       client.ClientInterface
	cfgManager   config.ManagerInterface
	options      CommitOptions
	editor       TextEditor
	clientConfig *types.ClientConfig
}

// NewCommitService creates a new CommitService instance with the provided options.
// It initializes the version control system (Git or SVN), configuration manager,
// and API client based on the given options.
//
// Parameters:
//   - options: CommitOptions containing configuration path and VCS preferences
//
// Returns:
//   - *CommitService: A new CommitService instance if successful
//   - error: An error if initialization fails due to VCS or config issues
func NewCommitService(options CommitOptions) (*CommitService, error) {
	vcs, cfgManager, err := factory.NewServiceDependencies(factory.ServiceOptions{
		UseSVN:     options.UseSVN,
		ConfigPath: options.ConfigPath,
		Provider:   options.Provider,
	})
	if err != nil {
		return nil, err
	}

	clientConfig, err := cfgManager.GetClientConfig(options.Provider)
	if err != nil {
		return nil, err
	}

	// Overwrite client config with command line flags
	ApplyCommonOptions(&options.CommonOptions, clientConfig)

	apiClient, err := client.New(clientConfig)
	if err != nil {
		return nil, err
	}

	return &CommitService{
		vcs:          vcs,
		client:       apiClient,
		cfgManager:   cfgManager,
		options:      options,
		editor:       &TerminalEditor{},
		clientConfig: clientConfig,
	}, nil
}

// NewCommitCmd creates and returns a new cobra.Command for the 'commit' subcommand.
// This command generates and creates a commit with staged changes in a Git or SVN repository.
//
// The command supports the following flags:
//   - --config, -c: Config path for the repository (string)
//   - --rich, -r: Generate detailed commit message with more context (bool)
//   - --yes, -y: Skip confirmation prompt and commit automatically (bool)
//   - --dry-run: Preview the generated commit message without actually committing (bool)
//   - --svn: Use SVN instead of Git for version control operations (bool)
//   - --no-verify: Skip git hooks verification, akin to using 'git commit --no-verify' (bool)
//   - --api-base: Override API base URL (string)
//   - --api-key: Override API key (string)
//   - --max-tokens: Override maximum tokens (int)
//   - --retries: Override retry count (int)
//   - --model: Override model name (string)
//   - --answer-path: Override answer path (string)
//   - --completion-path: Override completion path (string)
//   - --proxy: Override proxy URL (string)
//   - --frequency-penalty: Override frequency penalty (float)
//   - --temperature: Override temperature (float)
//   - --top-p: Override top_p value (float)
//   - --provider: Override AI provider (openai/deepseek)
//
// If no repository path is specified, it uses the current working directory.
// The command integrates with the root command's persistent configuration path.
//
// Returns a configured cobra.Command ready to be added to the command hierarchy.
func NewCommitCmd() *cobra.Command {
	options := CommitOptions{}

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate and create a commit with staged changes.",
		Long:  `Generate and create a commit with staged changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.RepoPath == "" {
				var err error
				options.RepoPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			// get config path from root command
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}
			options.ConfigPath = configPath

			service, err := NewCommitService(options)
			if err != nil {
				return err
			}

			return service.Execute()
		},
	}

	var generalFlags = pflag.NewFlagSet("General Flag", pflag.ExitOnError)
	var advancedFlags = pflag.NewFlagSet("Overwrite Flag", pflag.ExitOnError)

	// General Flags
	generalFlags.StringVar(&options.RepoPath, "repo", "", "Repository path")
	generalFlags.BoolVarP(&options.Rich, "rich", "r", false, "Generate rich commit message with details")
	generalFlags.BoolVarP(&options.AutoYes, "yes", "y", false, "Automatically commit without asking")
	generalFlags.BoolVar(&options.NoVerify, "no-verify", false, "Skip git hooks verification, akin to using 'git commit --no-verify'")
	generalFlags.BoolVar(&options.DryRun, "dry-run", false, "Print the generated commit message and exit without committing")
	generalFlags.BoolVar(&options.UseSVN, "svn", false, "Use SVN instead of Git")

	// Advanced API Flags (shared with other commands)
	AddAdvancedAPIFlags(advancedFlags, &options.CommonOptions)

	// Add flag groups to command
	cmd.Flags().AddFlagSet(generalFlags)
	cmd.Flags().AddFlagSet(advancedFlags)

	// Organize flags in help output
	cmd.Flags().SetInterspersed(false)
	SetAdvancedHelpFunc(cmd, generalFlags, advancedFlags)

	return cmd
}
