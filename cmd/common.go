package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/git"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Style definitions
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	// remind style
	remindStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	boxStyle    = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("2")).
			Padding(0, 1)
	topBottomBorder = lipgloss.Border{
		Top:          "=",
		Bottom:       "=",
		Left:         " ",
		Right:        " ",
		TopLeft:      " ",
		TopRight:     " ",
		BottomLeft:   " ",
		BottomRight:  " ",
		MiddleLeft:   " ",
		MiddleRight:  " ",
		Middle:       " ",
		MiddleTop:    " ",
		MiddleBottom: " ",
	}
	topBottomBorderStyle = lipgloss.NewStyle().
				BorderStyle(topBottomBorder).
				BorderForeground(lipgloss.Color("2"))
)

// CommandError represents specific error types that can occur during command operations
type CommandError struct {
	Type    string
	Message string
	Err     error
}

func (e *CommandError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func formatBoxedMessage(msg string) string {
	return boxStyle.Render(successStyle.Render(msg))
}

func formatHighlightedMessage(msg string) string {
	return topBottomBorderStyle.Render(successStyle.Render(msg))
}

func formatRemindMessage(msg string) string {
	return remindStyle.Render(msg)
}

const (
	LANGUAGE_KEY    = "output.lang"
	REVIEW_LANG_KEY = "output.review_lang"
	MARKDOWN_THEME  = "output.markdown_theme"
)

func createServiceDependencies(options struct {
	UseSVN     bool
	ConfigPath string
}) (git.VCS, config.ManagerInterface, error) {
	vcsType := git.Git
	if options.UseSVN {
		vcsType = git.SVN
	}

	vcs, err := git.NewVCS(vcsType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create VCS (%s): %w", vcsType, err)
	}

	cfgManager, err := config.New(options.ConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	return vcs, cfgManager, nil
}

type textEditor struct {
	textarea textarea.Model
	err      error
}

func (m textEditor) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles key presses and updates the text editor model and command
// accordingly. The following keys are handled:
//
// - Esc: Quit the editor if pressed with the alt key.
// - Ctrl+C: Quit the editor.
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

// View renders the text editor as a string. It formats the text area with
// a message instructing the user how to exit the editor, and then appends
// the rendered text area.
func (m textEditor) View() string {
	return fmt.Sprintf(
		"Edit message (Ctrl+C or Alt+Esc to save and exit):\n\n%s",
		m.textarea.View(),
	)
}

// editText opens an interactive text editor in the terminal for editing the provided text.
// It creates a text area with line wrapping, configures it with the terminal width,
// and allows the user to edit the text in a modal interface.
//
// Parameters:
//   - initialText: string - The text to be pre-populated in the editor
//
// Returns:
//   - string: The edited text after user modifications, with leading/trailing whitespace removed
//   - error: Any error that occurred during the editing process, including initialization
//     or program execution errors
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

// TextEditor represents an interface for text editing operations
type TextEditor interface {
	Edit(initialText string) (string, error)
}

// TerminalEditor implements TextEditor for terminal-based editing
type TerminalEditor struct{}

func (e *TerminalEditor) Edit(initialText string) (string, error) {
	return editText(initialText)
}
