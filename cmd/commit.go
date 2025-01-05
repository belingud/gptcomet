package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/belingud/go-gptcomet/internal/client"
	"github.com/belingud/go-gptcomet/internal/config"
	"github.com/belingud/go-gptcomet/internal/debug"
	"github.com/belingud/go-gptcomet/internal/git"

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
	var (
		repoPath string
		rich     bool
		dryRun   bool
		useSVN   bool
		autoYes  bool
	)

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

			if rich {
				debug.Println("Using rich output")
			}

			// Create VCS instance based on flag
			vcsType := git.Git
			if useSVN {
				vcsType = git.SVN
			}

			vcs, err := git.NewVCS(vcsType)
			if err != nil {
				return fmt.Errorf("failed to create VCS (%s): %w", vcsType, err)
			}
			debug.Printf("Using VCS: %s", vcsType)

			// Check for staged changes
			hasStagedChanges, err := vcs.HasStagedChanges(repoPath)
			if err != nil {
				return fmt.Errorf("failed to check staged changes: %w", err)
			}
			if !hasStagedChanges {
				return fmt.Errorf("no staged changes found")
			}
			debug.Println("Found staged changes")

			// Get config path from root command
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Create config manager
			cfgManager, err := config.New(configPath)
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Get filtered diff
			diff, err := vcs.GetStagedDiffFiltered(repoPath, cfgManager)
			if err != nil {
				return fmt.Errorf("failed to get diff: %w", err)
			}
			if diff == "" {
				return fmt.Errorf("no staged changes found after filtering")
			}
			debug.Printf("Got diff length: %d", len(diff))

			// Get client config
			clientConfig, err := cfgManager.GetClientConfig()
			if err != nil {
				return err
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

				// If dry-run is set, exit here without committing
				if dryRun {
					return nil
				}
				var answer string
				if autoYes {
					// Automatically commit without asking
					answer = "y"
				} else {

					fmt.Print("\nWould you like to create this commit? ([Y]es/[n]o/[r]etry/[e]dit): ")
					answer, err = reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read answer: %w", err)
					}
					answer = strings.ToLower(strings.TrimSpace(answer))
				}

				// If empty answer, use default (yes)
				if answer == "" {
					answer = "y"
				}

				switch answer {
				case "y", "yes":
					// Create commit
					err = vcs.CreateCommit(repoPath, commitMsg)
					if err != nil {
						return fmt.Errorf("failed to create commit: %w", err)
					}

					// Get commit hash
					commitHash, err := vcs.GetLastCommitHash(repoPath)
					if err != nil {
						return fmt.Errorf("failed to get commit hash: %w", err)
					}

					// Get commit info
					commitInfo, err := vcs.GetCommitInfo(repoPath, commitHash)
					if err != nil {
						return fmt.Errorf("failed to get commit info: %w", err)
					}

					fmt.Printf("\nSuccessfully created commit:\n%s\n", commitInfo)
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
	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Automatically commit without asking")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the generated commit message and exit without committing")
	cmd.Flags().BoolVar(&useSVN, "svn", false, "Use SVN instead of Git")

	return cmd
}
