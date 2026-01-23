package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/belingud/gptcomet/internal/debug"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/ui"
)

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

// removeThinkTags removes all <thinking> tags and their content from the input string.
//
// Parameters:
//   - input: The input string containing <thinking> tags
//
// Returns:
//   - string: The input string with all <thinking> tags and their content removed
func removeThinkTags(input string) (string, error) {
	if !strings.HasPrefix(input, "<thinking>") {
		return input, nil
	}
	if strings.Contains(input, "<thinking>") && !strings.Contains(input, "</thinking>") {
		return input, fmt.Errorf("thinking tag is not closed! The value of max_token may be too small")
	}
	re := regexp.MustCompile(`(?is)<thinking>.*?</thinking>`)
	cleaned := re.ReplaceAllString(input, "")
	return strings.TrimSpace(cleaned), nil
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
	// Get verbose setting
	verbose := s.getVerboseSetting()

	// Initialize progress tracking if verbose
	var progress *ui.Progress
	if verbose {
		progress = ui.NewProgress(true)
		progress.AddStages("Fetching git diff", "Generating message")
	}

	// check for staged changes
	if progress != nil {
		progress.Start("Fetching git diff")
	}
	hasStagedChanges, err := s.vcs.HasStagedChanges(s.options.RepoPath)
	if err != nil {
		if progress != nil {
			progress.Error("Fetching git diff", err)
		}
		return err
	}
	if !hasStagedChanges {
		if progress != nil {
			progress.Error("Fetching git diff", gptcometerrors.NoStagedChangesError())
		}
		return gptcometerrors.NoStagedChangesError()
	}

	// get diff of staged changes after filtering with file_ignore patterns
	diff, err := s.vcs.GetStagedDiffFiltered(s.options.RepoPath, s.cfgManager)
	debug.Printf("Got diff length: %d\n", len(diff))
	if err != nil {
		if progress != nil {
			progress.Error("Fetching git diff", err)
		}
		return err
	}
	if diff == "" {
		if progress != nil {
			progress.Error("Fetching git diff", gptcometerrors.NoStagedChangesError())
		}
		return gptcometerrors.NoStagedChangesError()
	}

	if progress != nil {
		progress.Complete("Fetching git diff")
	}

	fmt.Printf("Discovered provider: %s, model: %s\n", s.clientConfig.Provider, s.clientConfig.Model)

	if progress != nil {
		progress.StartWithNewLine("Generating message")
	}

	// generate commit message
	commitMsg, err := s.generateCommitMessage(diff)
	if err != nil {
		if progress != nil {
			progress.Error("Generating message", err)
		}
		return err
	}

	commitMsg, err = removeThinkTags(commitMsg)
	if err != nil {
		if progress != nil {
			progress.Error("Generating message", err)
		} else {
			fmt.Printf("Error in generating: %v\n", err)
		}
		return err
	}

	if progress != nil {
		progress.CompleteInNewLine("Generating message")
	}

	if s.options.DryRun {
		fmt.Printf("\nGenerated commit message:\n%s\n", formatBoxedMessage(commitMsg))
		return nil
	}

	return s.handleCommitInteraction(commitMsg)
}

// getVerboseSetting retrieves the console.verbose configuration
func (s *CommitService) getVerboseSetting() bool {
	if val, ok := s.cfgManager.GetNestedValue([]string{"console", "verbose"}); ok {
		if verbose, ok := val.(bool); ok {
			return verbose
		}
	}
	return false
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
			commitMsg, err = removeThinkTags(commitMsg)
			if err != nil {
				fmt.Printf("Error in generating: %v\n", err)
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
	err := s.vcs.CreateCommit(s.options.RepoPath, msg, s.options.NoVerify)
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
