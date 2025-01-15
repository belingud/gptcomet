package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/internal/git"

	"github.com/spf13/cobra"
)

// CommitOptions contains the configuration settings for the commit operation.
type CommitOptions struct {
	RepoPath   string
	Rich       bool
	DryRun     bool
	UseSVN     bool
	AutoYes    bool
	ConfigPath string
}

// CommitService handles the logic for committing changes to version control
// while integrating with the GPT service for generating commit messages.
// It manages the interaction between the version control system,
// API client, and configuration settings.
type CommitService struct {
	vcs        git.VCS
	client     client.ClientInterface
	cfgManager config.ManagerInterface
	options    CommitOptions
	editor     TextEditor
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
	vcs, cfgManager, err := createServiceDependencies(struct {
		UseSVN     bool
		ConfigPath string
	}{
		UseSVN:     options.UseSVN,
		ConfigPath: options.ConfigPath,
	})
	if err != nil {
		return nil, err
	}

	clientConfig, err := cfgManager.GetClientConfig()
	if err != nil {
		return nil, err
	}

	return &CommitService{
		vcs:        vcs,
		client:     client.New(clientConfig),
		cfgManager: cfgManager,
		options:    options,
		editor:     &TerminalEditor{},
	}, nil
}

// generateCommitMessage generates a commit message based on the provided git diff.
// It uses the configured prompt template (either rich or standard) to generate the message.
// If the configured output language is not English, it also translates the message
// to the target language using a translation prompt.
//
// Parameters:
//   - diff: The git diff string to generate the commit message from
//
// Returns:
//   - string: The generated (and optionally translated) commit message
//   - error: An error if message generation or translation fails, or if config is invalid
func (s *CommitService) generateCommitMessage(diff string) (string, error) {
	prompt := s.cfgManager.GetPrompt(s.options.Rich)
	msg, err := s.client.GenerateCommitMessage(diff, prompt)
	if err != nil {
		return "", err
	}

	// Translate commit message if output.lang is not en
	langValue, ok := s.cfgManager.Get(LANGUAGE_KEY)
	if !ok {
		return "", fmt.Errorf("failed to get output.lang: configuration key not found")
	}
	lang, ok := langValue.(string)
	if !ok {
		return "", fmt.Errorf("output.lang is not a string: %v", langValue)
	}

	// If language is English, return the original message
	if lang == "en" {
		return msg, nil
	}

	// Get translate_title setting
	translateTitle := s.cfgManager.GetOutputTranslateTitle()
	debug.Printf("Translate title setting: %v\n", translateTitle)

	// Handle translation based on translate_title setting
	translatePrompt := s.cfgManager.GetTranslationPrompt()

	// If translate_title is false, split message and translate only content
	if !translateTitle {
		prefix, content := splitCommitMessage(msg)
		debug.Printf("Split commit message: prefix=%s, content=%s\n", prefix, content)
		if prefix != "" {
			// Translate only the content part
			translatedContent, err := s.client.TranslateMessage(translatePrompt, content, lang)
			if err != nil {
				return "", err
			}
			return prefix + ": " + translatedContent, nil
		}
	}

	// Translate the entire message
	return s.client.TranslateMessage(translatePrompt, msg, lang)
}

// splitCommitMessage splits a commit message into prefix and content parts based on the first colon separator.
// The prefix is the text before the first colon, and content is everything after.
// If no colon is found in the message, prefix will be empty and the entire message becomes the content.
// Both prefix and content are returned with leading and trailing whitespace removed.
//
// Parameters:
//   - message: the commit message string to split
//
// Returns:
//   - prefix: the commit type/scope before the colon, or empty if no colon found
//   - content: the main commit message after the colon, or full message if no colon found
//
// Example: "feat: add new feature" -> "feat", "add new feature"
func splitCommitMessage(message string) (prefix, content string) {
	parts := strings.SplitN(message, ":", 2)
	if len(parts) != 2 {
		return "", message
	}

	prefix = strings.TrimSpace(parts[0])
	content = strings.TrimSpace(parts[1])
	return prefix, content
}

// Execute performs the commit operation with the following steps:
// 1. Checks for staged changes in the repository
// 2. Gets the filtered diff of staged changes
// 3. Generates a commit message using the diff
// 4. If dry-run is enabled, prints the generated message
// 5. Otherwise, handles the commit interaction
//
// It returns an error if:
// - There are no staged changes
// - Failed to check staged changes
// - Failed to get diff
// - No changes found after filtering
// - Failed to generate commit message
// - Failed to handle commit interaction
func (s *CommitService) Execute() error {
	// check for staged changes
	hasStagedChanges, err := s.vcs.HasStagedChanges(s.options.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	if !hasStagedChanges {
		return fmt.Errorf("no staged changes found")
	}

	// get diff of staged changes after filtering with file_ignore patterns
	diff, err := s.vcs.GetStagedDiffFiltered(s.options.RepoPath, s.cfgManager)
	debug.Printf("Got diff length: %d\n", len(diff))
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}
	if diff == "" {
		return fmt.Errorf("no staged changes found after filtering")
	}

	// generate commit message
	commitMsg, err := s.generateCommitMessage(diff)
	if err != nil {
		return err
	}

	if s.options.DryRun {
		fmt.Printf("\nGenerated commit message:\n%s\n", formatBoxedMessage(commitMsg))
		return nil
	}

	return s.handleCommitInteraction(commitMsg)
}

// handleCommitInteraction manages the interactive commit message workflow.
// It displays the current commit message and prompts the user for actions:
// - Yes: Creates the commit with the current message
// - No: Cancels the operation
// - Retry: Regenerates the commit message based on staged changes
// - Edit: Opens an editor to manually modify the commit message
//
// The function loops until the user either confirms the commit or cancels the operation.
// If AutoYes option is enabled, it skips the interaction and creates the commit directly.
//
// Parameters:
//   - initialMsg: The initial commit message to start with
//
// Returns:
//   - error: An error if any operation fails, nil otherwise
func (s *CommitService) handleCommitInteraction(initialMsg string) error {
	commitMsg := initialMsg
	reader := bufio.NewReader(os.Stdin)

	for {
		if commitMsg != "" {
			fmt.Printf("\nCurrent commit message:\n%s\n", formatBoxedMessage(commitMsg))
		}

		if s.options.AutoYes {
			return s.createCommit(commitMsg)
		}

		fmt.Print("\nWould you like to create this commit? ([Y]es/[n]o/[r]etry/[e]dit): ")
		answer, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read answer: %w", err)
		}

		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "" {
			answer = "y"
		}

		switch answer {
		case "y", "yes":
			return s.createCommit(commitMsg)
		case "n", "no":
			fmt.Println("Operation cancelled")
			return nil
		case "r", "retry":
			diff, err := s.vcs.GetStagedDiffFiltered(s.options.RepoPath, s.cfgManager)
			if err != nil {
				return err
			}
			commitMsg, err = s.generateCommitMessage(diff)
			if err != nil {
				return err
			}
		case "e", "edit":
			edited, err := s.editor.Edit(commitMsg)
			if err != nil {
				fmt.Printf("Error editing message: %v\n", err)
				continue
			}
			commitMsg = edited
		default:
			fmt.Println("Invalid option, please try again")
		}
	}
}

// createCommit creates a new git commit with the given message and prints commit information.
// It performs the following steps:
// 1. Creates a commit with the provided message
// 2. Retrieves the hash of the newly created commit
// 3. Gets detailed commit information
// 4. Prints the commit details to stdout
//
// Parameters:
//   - msg: string containing the commit message
//
// Returns:
//   - error: nil if successful, otherwise error details with context
func (s *CommitService) createCommit(msg string) error {
	err := s.vcs.CreateCommit(s.options.RepoPath, msg)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	commitHash, err := s.vcs.GetLastCommitHash(s.options.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}

	// Commit details just like git commit output
	commitInfo, err := s.vcs.GetCommitInfo(s.options.RepoPath, commitHash)
	if err != nil {
		return fmt.Errorf("failed to get commit info: %w", err)
	}

	fmt.Printf("\nSuccessfully created commit:\n%s\n", commitInfo)
	return nil
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
//
// If no repository path is specified, it uses the current working directory.
// The command integrates with the root command's persistent configuration path.
//
// Returns a configured cobra.Command ready to be added to the command hierarchy.
func NewCommitCmd() *cobra.Command {
	options := CommitOptions{}

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate and create a commit with staged changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.RepoPath == "" {
				var err error
				options.RepoPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			// 从根命令获取配置路径
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

	cmd.Flags().StringVarP(&options.RepoPath, "config", "c", "", "Config path")
	cmd.Flags().BoolVarP(&options.Rich, "rich", "r", false, "Generate rich commit message with details")
	cmd.Flags().BoolVarP(&options.AutoYes, "yes", "y", false, "Automatically commit without asking")
	cmd.Flags().BoolVar(&options.DryRun, "dry-run", false, "Print the generated commit message and exit without committing")
	cmd.Flags().BoolVar(&options.UseSVN, "svn", false, "Use SVN instead of Git")

	return cmd
}
