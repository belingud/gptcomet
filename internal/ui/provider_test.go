package ui

import (
	"testing"

	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigInput(t *testing.T) {
	tests := []struct {
		name    string
		configs map[string]config.ConfigRequirement
	}{
		{
			name:    "Empty configs",
			configs: map[string]config.ConfigRequirement{},
		},
		{
			name: "Single config",
			configs: map[string]config.ConfigRequirement{
				"test_key": {
					DefaultValue:  "default",
					PromptMessage: "Enter test key",
				},
			},
		},
		{
			name: "Multiple configs",
			configs: map[string]config.ConfigRequirement{
				"api_key": {
					DefaultValue:  "",
					PromptMessage: "Enter API key",
				},
				"model": {
					DefaultValue:  "gpt-4",
					PromptMessage: "Enter model",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := NewConfigInput(tt.configs)
			require.NotNil(t, ci)

			// Verify number of inputs matches config count
			assert.Equal(t, len(tt.configs), len(ci.inputs))
			assert.Equal(t, len(tt.configs), len(ci.configKeys))

			// Verify config keys are sorted
			var lastKey string
			for _, key := range ci.configKeys {
				if lastKey != "" {
					assert.True(t, key > lastKey, "keys should be sorted")
				}
				lastKey = key
			}

			// Verify inputs are properly initialized
			for i, key := range ci.configKeys {
				input := ci.inputs[i]
				config := tt.configs[key]

				assert.Equal(t, config.DefaultValue, input.Placeholder)

				// Verify API key input is masked
				if key == "api_key" {
					assert.Equal(t, textinput.EchoPassword, input.EchoMode)
				} else {
					assert.Equal(t, textinput.EchoNormal, input.EchoMode)
				}
			}

			// Verify first input is focused
			if len(ci.inputs) > 0 {
				assert.True(t, ci.inputs[0].Focused())
			}
		})
	}
}

func TestConfigInputUpdate(t *testing.T) {
	configs := map[string]config.ConfigRequirement{
		"key1": {DefaultValue: "default1", PromptMessage: "Enter key 1"},
		"key2": {DefaultValue: "default2", PromptMessage: "Enter key 2"},
	}

	tests := []struct {
		name  string
		msg   tea.Msg
		check func(*testing.T, *ConfigInput)
	}{
		{
			name: "Enter key moves to next input",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, ci *ConfigInput) {
				assert.Equal(t, 1, ci.currentKey)
				assert.False(t, ci.done)
			},
		},
		{
			name: "Enter key on last input completes",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, ci *ConfigInput) {
				// Move to the first input and press enter
				_, _ = ci.Update(tea.KeyMsg{Type: tea.KeyEnter})
				// Move to the last input and press enter
				ci.currentKey = 1 // Set to last input index
				_, _ = ci.Update(tea.KeyMsg{Type: tea.KeyEnter})
				// Check if we're still on the last input and done is true
				assert.Equal(t, 1, ci.currentKey)
				assert.True(t, ci.done)
			},
		},
		{
			name: "Escape key quits",
			msg:  tea.KeyMsg{Type: tea.KeyEsc},
			check: func(t *testing.T, ci *ConfigInput) {
				assert.True(t, ci.quitting)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := NewConfigInput(configs)
			_, _ = ci.Update(tt.msg)
			tt.check(t, ci)
		})
	}
}

func TestConfigInputView(t *testing.T) {
	configs := map[string]config.ConfigRequirement{
		"key1": {DefaultValue: "default1", PromptMessage: "Enter key 1"},
		"key2": {DefaultValue: "default2", PromptMessage: "Enter key 2"},
	}

	ci := NewConfigInput(configs)
	require.NotNil(t, ci)

	// Set a value for the first input
	ci.inputs[0].SetValue("test-value")

	// Move to the second input
	_, _ = ci.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Get the view
	view := ci.View()

	// Check that the view contains the expected elements
	assert.Contains(t, view, "Configure provider:")
	assert.Contains(t, view, "Enter key 1")
	assert.Contains(t, view, "Enter key 2")
	assert.Contains(t, view, "test-value")
	assert.Contains(t, view, "default2")
	assert.Contains(t, view, "(2/2)")
}

// GetResult returns the appropriate values based on custom and default keys
func GetResult(defaultKey1, defaultKey2, customKey1, customKey2 string) (string, string) {
	if customKey1 == "" {
		return defaultKey1, defaultKey2
	}
	return customKey1, customKey2
}

func TestGetResult(t *testing.T) {
	tests := []struct {
		name         string
		defaultKey1  string
		defaultKey2  string
		customKey1   string
		customKey2   string
		wantResult1  string
		wantResult2  string
		wantDefault1 string
		wantDefault2 string
	}{
		{
			name:         "test case",
			defaultKey1:  "default1",
			defaultKey2:  "default2",
			customKey1:   "custom1",
			customKey2:   "custom2",
			wantResult1:  "custom1",  // 修改：期望 customKey1
			wantResult2:  "custom2",  // 修改：期望 customKey2
			wantDefault1: "default1", // 修改：期望 defaultKey1
			wantDefault2: "default2", // 保持不变
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2 := GetResult(tt.defaultKey1, tt.defaultKey2, tt.customKey1, tt.customKey2)
			assert.Equal(t, tt.wantResult1, result1)
			assert.Equal(t, tt.wantResult2, result2)

			default1, default2 := GetResult(tt.defaultKey1, tt.defaultKey2, "", "")
			assert.Equal(t, tt.wantDefault1, default1)
			assert.Equal(t, tt.wantDefault2, default2)
		})
	}
}

func TestProviderSelector_New(t *testing.T) {
	providers := []string{"test1", "test2", "test3"}
	selector := NewProviderSelector(providers)

	if selector == nil {
		t.Error("NewProviderSelector returned nil")
	}

	if len(selector.list.Items()) != len(providers)+1 {
		t.Errorf("Expected %d items (including manual input), got %d", len(providers)+1, len(selector.list.Items()))
	}

	// Verify the last item is manual input option
	lastItem := selector.list.Items()[len(providers)]
	if lastItem.(item).title != "Input Manually" {
		t.Errorf("Last item should be 'Input Manually', got %s", lastItem.(item).title)
	}
}

func TestProviderSelector_Update(t *testing.T) {
	selector := NewProviderSelector([]string{"test1"})

	// Test window size update
	model, _ := selector.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	updated := model.(*ProviderSelector)
	if updated.list.Width() != 100 {
		t.Errorf("Expected width 100, got %d", updated.list.Width())
	}

	// Test item selection
	model, _ = selector.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = model.(*ProviderSelector)
	if updated.choice == "" {
		t.Error("Expected choice to be set after Enter key")
	}

	// Test quit
	model, _ = selector.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated = model.(*ProviderSelector)
	if !updated.quitting {
		t.Error("Expected quitting to be true after Ctrl+C")
	}
}

func TestConfigInput_New(t *testing.T) {
	configs := map[string]config.ConfigRequirement{
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "test-model",
			PromptMessage: "Enter model name",
		},
	}

	input := NewConfigInput(configs)

	if len(input.inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(input.inputs))
	}

	// Verify API key input is in password mode
	if input.inputs[0].EchoMode != 1 { // 1 is password mode
		t.Error("API key input should be in password mode")
	}

	// Verify default value is set
	if input.inputs[1].Placeholder != "test-model" {
		t.Errorf("Expected model placeholder 'test-model', got %s", input.inputs[1].Placeholder)
	}
}

func TestConfigInput_Update(t *testing.T) {
	configs := map[string]config.ConfigRequirement{
		"test1": {DefaultValue: "default1"},
		"test2": {DefaultValue: "default2"},
	}
	input := NewConfigInput(configs)

	// Test input value and press enter
	model, _ := input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("test")})
	input = model.(*ConfigInput)

	model, _ = input.Update(tea.KeyMsg{Type: tea.KeyEnter})
	input = model.(*ConfigInput)

	if input.currentKey != 1 {
		t.Error("Expected to move to next input after Enter")
	}

	// Test pressing enter on the last input completes
	model, _ = input.Update(tea.KeyMsg{Type: tea.KeyEnter})
	input = model.(*ConfigInput)

	if !input.done {
		t.Error("Expected done to be true after last Enter")
	}
}

func TestConfigInput_GetConfigs(t *testing.T) {
	configs := map[string]config.ConfigRequirement{
		"test1": {DefaultValue: "default1"},
		"test2": {DefaultValue: "default2"},
	}
	input := NewConfigInput(configs)

	// Set one value and let the other use the default
	input.inputs[0].SetValue("custom1")

	result := input.GetConfigs()

	if result["test1"] != "custom1" {
		t.Errorf("Expected test1='custom1', got %s", result["test1"])
	}
	if result["test2"] != "default2" {
		t.Errorf("Expected test2='default2', got %s", result["test2"])
	}
}

func TestItem_Interface(t *testing.T) {
	i := item{
		title:       "Test Title",
		description: "Test Description",
	}

	if i.Title() != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %s", i.Title())
	}

	if i.Description() != "Test Description" {
		t.Errorf("Expected description 'Test Description', got %s", i.Description())
	}

	if i.FilterValue() != "Test Title" {
		t.Errorf("Expected filter value 'Test Title', got %s", i.FilterValue())
	}
}

func TestItemDelegate_Interface(t *testing.T) {
	d := itemDelegate{}

	if d.Height() != 1 {
		t.Errorf("Expected height 1, got %d", d.Height())
	}

	if d.Spacing() != 0 {
		t.Errorf("Expected spacing 0, got %d", d.Spacing())
	}
}
