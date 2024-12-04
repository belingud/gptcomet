package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/belingud/gptcomet/internal/client"
	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/internal/git"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Style definitions
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	boxStyle     = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Padding(0, 1)
)

const (
	LANGUAGE_KEY = "output.lang"
)

type textEditor struct {
	textarea textarea.Model
	err      error
}

func (m textEditor) Init() tea.Cmd {
	return textarea.Blink
}

func (m textEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if msg.Alt {
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m textEditor) View() string {
	return fmt.Sprintf(
		"Edit commit message (Ctrl+C or Alt+Esc to save and exit):\n\n%s",
		m.textarea.View(),
	)
}

func editText(initialText string) (string, error) {
	// Get terminal width
	width, _, err := term.GetSize(int(syscall.Stdout))
	if err != nil {
		width = 100 // Default width if unable to get terminal size
	}

	ta := textarea.New()
	ta.SetValue(initialText)
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 4096
	ta.SetWidth(width - 4) // Leave some margin for borders
	ta.SetHeight(10)       // Set a reasonable height

	m := textEditor{
		textarea: ta,
		err:      nil,
	}

	p := tea.NewProgram(m)
	model, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run editor: %w", err)
	}

	finalModel := model.(textEditor)
	if finalModel.err != nil {
		return "", finalModel.err
	}

	return strings.TrimSpace(finalModel.textarea.Value()), nil
}

func formatCommitMessage(msg string) string {
	return boxStyle.Render(successStyle.Render(msg))
}

// NewCommitCmd creates a new commit command
func NewCommitCmd() *cobra.Command {
	var repoPath string
	var rich bool

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Generate and create a commit with staged changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoPath == "" {
				var err error
				repoPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}
			debug.Printf("Using repository path: %s", repoPath)

			// Check for staged changes
			hasStagedChanges, err := git.HasStagedChanges(repoPath)
			if err != nil {
				return fmt.Errorf("failed to check staged changes: %w", err)
			}
			if !hasStagedChanges {
				return fmt.Errorf("no staged changes found")
			}
			debug.Println("Found staged changes")

			// Get diff
			diff, err := git.GetDiff(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get diff: %w", err)
			}
			debug.Printf("Got diff length: %d", len(diff))

			// Create config manager
			cfgManager, err := config.New()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Get client config
			clientConfig, err := cfgManager.GetClientConfig()
			if err != nil {
				return fmt.Errorf("failed to get client config: %w", err)
			}

			// Create client
			client := client.New(clientConfig)

			reader := bufio.NewReader(os.Stdin)
			var commitMsg string
			for {
				if commitMsg != "" {
					fmt.Printf("\nCurrent commit message:\n%s\n", formatCommitMessage(commitMsg))
				}
				fmt.Println("🤖 Hang tight, I'm cooking up something good!")

				// Get prompt based on rich flag
				prompt := cfgManager.GetPrompt(rich)

				if commitMsg == "" {
					// Generate commit message
					var err error
					commitMsg, err = client.GenerateCommitMessage(diff, prompt)
					if err != nil {
						return fmt.Errorf("failed to generate commit message: %w", err)
					}

				}
				// If output.lang is not "en", prompt for translation
				var lang string
				langValue, ok := cfgManager.Get(LANGUAGE_KEY)
				if !ok {
					return fmt.Errorf("failed to get output.lang: configuration key not found")
				}
				lang, ok = langValue.(string)
				if !ok {
					return fmt.Errorf("output.lang is not a string: %v", langValue)
				}
				if lang != "en" {
					translatePrompt := cfgManager.GetTranslationPrompt()
					commitMsg, err = client.TranslateMessage(translatePrompt, commitMsg, lang)
					if err != nil {
						return fmt.Errorf("failed to translate commit message: %w", err)
					}
				}
				fmt.Printf("\nGenerated commit message:\n%s\n", formatCommitMessage(commitMsg))

				fmt.Print("\nWould you like to create this commit? ([Y]es/[n]o/[r]etry/[e]dit): ")
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read answer: %w", err)
				}
				answer = strings.ToLower(strings.TrimSpace(answer))

				// If empty answer, use default (yes)
				if answer == "" {
					answer = "y"
				}

				switch answer {
				case "y", "yes":
					// Create commit
					if err := git.CreateCommit(repoPath, commitMsg); err != nil {
						return fmt.Errorf("failed to create commit: %w", err)
					}
					fmt.Printf("\nSuccessfully created commit with message:\n%s\n", formatCommitMessage(commitMsg))
					return nil
				case "n", "no":
					fmt.Println("Operation cancelled")
					return nil
				case "r", "retry":
					commitMsg = ""
					continue
				case "e", "edit":
					edited, err := editText(commitMsg)
					if err != nil {
						fmt.Printf("Error editing message: %v\n", err)
						continue
					}
					commitMsg = edited
					continue
				default:
					fmt.Println("Invalid option, please try again")
					continue
				}
			}
		},
	}

	cmd.Flags().StringVarP(&repoPath, "config", "c", "", "Config path")
	cmd.Flags().BoolVarP(&rich, "rich", "r", false, "Generate rich commit message with details")

	return cmd
}
