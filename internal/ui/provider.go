package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	internal_cfg "github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpTextHeight = 5

var (
	titleStyle = lipgloss.NewStyle().
			MarginLeft(2)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))

	paginationStyle = list.DefaultStyles().
			PaginationStyle.
			PaddingLeft(4)

	helpStyle = list.DefaultStyles().
			HelpStyle.
			PaddingLeft(4).
			PaddingBottom(1)

	quitTextStyle = lipgloss.NewStyle().
			Margin(1, 0, 1, 0)
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type ProviderSelector struct {
	list        list.Model
	choice      string
	quitting    bool
	manualInput textinput.Model
	inputMode   bool
}

func NewProviderSelector(providers []string) *ProviderSelector {
	items := make([]list.Item, len(providers)+1)
	for i, p := range providers {
		items[i] = item{title: p, description: fmt.Sprintf("Configure %s provider", p)}
	}
	items[len(providers)] = item{title: "Input Manually", description: "Enter provider name manually"}

	const defaultWidth = 40

	// Calculate list height based on number of items
	// Add 5 for title, help text and footer
	listHeight := len(items) + helpTextHeight
	if listHeight < 1 {
		listHeight = 1
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select Provider"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	// disable pagination
	l.DisableQuitKeybindings()
	l.SetShowPagination(false)

	// initialize manual input
	input := textinput.New()
	input.Placeholder = "Enter provider name"
	input.Width = defaultWidth

	return &ProviderSelector{
		list:        l,
		manualInput: input,
	}
}

func (m *ProviderSelector) Init() tea.Cmd {
	return nil
}

func (m *ProviderSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Calculate max height based on window size
		// Subtract 4 for margins and borders
		maxHeight := msg.Height - 4
		if maxHeight < 1 {
			maxHeight = 1
		}

		// Get current items count
		// Calculate list height based on number of items
		// Add 5 for title, help text and footer
		itemsHeight := len(m.list.Items()) + helpTextHeight
		if itemsHeight > maxHeight {
			itemsHeight = maxHeight
		}

		m.list.SetHeight(itemsHeight)
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		// if in manual
		if m.inputMode {
			switch msg.Type {
			case tea.KeyEsc:
				m.inputMode = false
				return m, nil
			case tea.KeyEnter:
				if value := m.manualInput.Value(); value != "" {
					m.choice = value
					return m, tea.Quit
				}
			default:
				var cmd tea.Cmd
				m.manualInput, cmd = m.manualInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// if not in manual
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				if i.Title() == "Input Manually" {
					m.inputMode = true
					m.manualInput.Focus()
					return m, nil
				}
				m.choice = i.Title()
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *ProviderSelector) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected provider: %s", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("Configuration cancelled.")
	}
	if m.inputMode {
		return fmt.Sprintf(
			"\nEnter provider name:\n\n%s\n\n(Press Enter to confirm, Esc to go back)\n",
			m.manualInput.View(),
		)
	}
	return "\n" + m.list.View()
}

func (m *ProviderSelector) Selected() string {
	return m.choice
}

type ConfigInput struct {
	// provider   string
	inputs     []textinput.Model
	configs    map[string]config.ConfigRequirement
	configKeys []string
	currentKey int
	done       bool
	quitting   bool
}

func NewConfigInput(configs map[string]config.ConfigRequirement) *ConfigInput {
	var inputs []textinput.Model
	keys := []string{"api_base", "model", "api_key", "max_tokens"}
	var configKeys []string
	processed := make(map[string]bool)

	for _, key := range keys {
		if _, ok := configs[key]; ok {
			configKeys = append(configKeys, key)
			processed[key] = true
		}
	}

	// collect remaining keys and sort them for stable ordering
	var remainingKeys []string
	for k := range configs {
		if !processed[k] {
			remainingKeys = append(remainingKeys, k)
		}
	}
	sort.Strings(remainingKeys)
	configKeys = append(configKeys, remainingKeys...)

	// create input for each config
	for _, key := range configKeys {
		input := textinput.New()
		input.Placeholder = configs[key].DefaultValue
		input.Width = 40
		if key == "api_key" {
			input.EchoMode = textinput.EchoPassword
		}
		inputs = append(inputs, input)
	}

	// activate first input
	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return &ConfigInput{
		inputs:     inputs,
		configs:    configs,
		configKeys: configKeys,
		currentKey: 0,
	}
}

func (m *ConfigInput) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ConfigInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			// if current input is empty and has default value
			if m.inputs[m.currentKey].Value() == "" {
				if def := m.configs[m.configKeys[m.currentKey]].DefaultValue; def != "" {
					m.inputs[m.currentKey].SetValue(def)
				} else {
					// error
					return m, nil
				}
			}

			// move to next input
			if m.currentKey < len(m.inputs)-1 {
				m.inputs[m.currentKey].Blur()
				m.currentKey++
				m.inputs[m.currentKey].Focus()
				return m, nil
			}

			// if last input, done
			m.done = true
			return m, tea.Quit
		}
	}

	// only update current input
	var cmd tea.Cmd
	m.inputs[m.currentKey], cmd = m.inputs[m.currentKey].Update(msg)
	return m, cmd
}

func (m *ConfigInput) View() string {
	var s strings.Builder

	s.WriteString("Configure provider:\n\n")

	// show previous inputs if any
	if m.currentKey > 0 {
		s.WriteString("Previous inputs:\n")
		for i := 0; i < m.currentKey; i++ {
			key := m.configKeys[i]
			config := m.configs[key]
			prompt := config.PromptMessage
			if prompt == "" {
				prompt = key
			}
			value := m.inputs[i].Value()
			if value == "" {
				value = config.DefaultValue
			} else if key == "api_key" {
				value = internal_cfg.MaskAPIKey(value, 3)
			}
			s.WriteString(fmt.Sprintf("  %s: %s\n", prompt, value))
		}
		s.WriteString("\n")
	}

	// only current input field
	key := m.configKeys[m.currentKey]
	config := m.configs[key]

	// use PromptMessage if it exists, otherwise use key
	prompt := config.PromptMessage
	if prompt == "" {
		prompt = key
	}
	s.WriteString(fmt.Sprintf("Enter %s", prompt))
	if config.DefaultValue != "" {
		s.WriteString(fmt.Sprintf(" (default: %s)", config.DefaultValue))
	}
	s.WriteString(":\n")
	s.WriteString(m.inputs[m.currentKey].View())
	s.WriteString("\n\n")

	// show progress
	s.WriteString(fmt.Sprintf("(%d/%d) Press Enter to continue, Esc to quit", m.currentKey+1, len(m.inputs)))

	return s.String()
}

func (m *ConfigInput) Done() bool {
	return m.done
}

func (m *ConfigInput) GetConfigs() map[string]string {
	result := make(map[string]string)
	for i, input := range m.inputs {
		key := m.configKeys[i]
		value := input.Value()
		if value == "" {
			value = m.configs[key].DefaultValue
		}
		result[key] = value
	}
	return result
}
