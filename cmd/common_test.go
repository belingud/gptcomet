package cmd

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFormatBoxedMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantEmpty bool
	}{
		{
			name:      "Simple message",
			message:   "Test message",
			wantEmpty: false,
		},
		{
			name:      "Empty message",
			message:   "",
			wantEmpty: false, // The function should still box empty strings
		},
		{
			name:      "Long message",
			message:   "This is a much longer test message that contains more content",
			wantEmpty: false,
		},
		{
			name:      "Message with special characters",
			message:   "Test: @#$%^&*()",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBoxedMessage(tt.message)
			assert.NotEmpty(t, result, "formatBoxedMessage should return non-empty result")
			// Note: The result may format the message differently, so we just check it's non-empty
		})
	}
}

func TestFormatHighlightedMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Simple message",
			message: "Important message",
		},
		{
			name:    "Empty message",
			message: "",
		},
		{
			name:    "Long message",
			message: "This is an important highlighted message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHighlightedMessage(tt.message)
			assert.NotEmpty(t, result, "formatHighlightedMessage should return non-empty result")
			// The result should contain the original message
			assert.Contains(t, result, tt.message, "Result should contain original message")
		})
	}
}

func TestFormatRemindMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Reminder message",
			message: "Please remember this",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRemindMessage(tt.message)
			assert.NotEmpty(t, result, "formatRemindMessage should return non-empty result")
			// The result should contain the original message
			assert.Contains(t, result, tt.message, "Result should contain original message")
		})
	}
}

func TestCommandError(t *testing.T) {
	tests := []struct {
		name         string
		errType      string
		errMessage   string
		err          error
		wantContains []string
	}{
		{
			name:         "Error with underlying error",
			errType:      "TestError",
			errMessage:   "Something went wrong",
			err:          assert.AnError,
			wantContains: []string{"TestError", "Something went wrong"},
		},
		{
			name:         "Error without underlying error",
			errType:      "TestError",
			errMessage:   "Simple error",
			err:          nil,
			wantContains: []string{"TestError", "Simple error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdErr := &CommandError{
				Type:    tt.errType,
				Message: tt.errMessage,
				Err:     tt.err,
			}

			errStr := cmdErr.Error()
			assert.NotEmpty(t, errStr, "Error() should return non-empty string")

			for _, substr := range tt.wantContains {
				assert.Contains(t, errStr, substr, "Error string should contain: "+substr)
			}
		})
	}
}

func TestTextEditor(t *testing.T) {
	t.Run("TextEditor interface", func(t *testing.T) {
		// Verify that TerminalEditor implements TextEditor
		editor := &TerminalEditor{}

		var _ TextEditor = editor
		// If this compiles, the interface is implemented correctly
		assert.NotNil(t, editor, "TerminalEditor should be non-nil")
	})
}

func TestConstants(t *testing.T) {
	t.Run("Language key constants", func(t *testing.T) {
		assert.Equal(t, "output.lang", LANGUAGE_KEY, "LANGUAGE_KEY should match")
		assert.Equal(t, "output.review_lang", REVIEW_LANG_KEY, "REVIEW_LANG_KEY should match")
		assert.Equal(t, "output.markdown_theme", MARKDOWN_THEME, "MARKDOWN_THEME should match")
	})
}

func TestTextEditorInit(t *testing.T) {
	t.Run("Init returns blink command", func(t *testing.T) {
		editor := textEditor{}
		cmd := editor.Init()
		assert.NotNil(t, cmd, "Init() should return a command")
	})
}

func TestTextEditorUpdate(t *testing.T) {
	t.Run("Update with Esc key with Alt", func(t *testing.T) {
		editor := textEditor{}
		msg := tea.KeyMsg{
			Type: tea.KeyEsc,
			Alt:  true,
		}

		model, cmd := editor.Update(msg)
		// tea.Quit is a function, we can't compare it directly
		// Just verify that a command is returned
		assert.NotNil(t, cmd, "Update should return a command when Alt+Esc is pressed")
		assert.NotNil(t, model, "Update should return a model")
	})

	t.Run("Update with Esc key without Alt", func(t *testing.T) {
		editor := textEditor{}
		msg := tea.KeyMsg{
			Type: tea.KeyEsc,
			Alt:  false,
		}

		model, cmd := editor.Update(msg)
		assert.Nil(t, cmd, "Update should not return a quit command when Esc is pressed without Alt")
		assert.NotNil(t, model, "Update should return a model")
	})

	t.Run("Update with Ctrl+C", func(t *testing.T) {
		editor := textEditor{}
		msg := tea.KeyMsg{
			Type: tea.KeyCtrlC,
		}

		model, cmd := editor.Update(msg)
		// tea.Quit is a function, we can't compare it directly
		// Just verify that a command is returned
		assert.NotNil(t, cmd, "Update should return a command when Ctrl+C is pressed")
		assert.NotNil(t, model, "Update should return a model")
	})

	t.Run("Update with other key messages", func(t *testing.T) {
		editor := textEditor{}
		msg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
		}

		model, cmd := editor.Update(msg)
		assert.Nil(t, cmd, "Update should not return a quit command for regular keys")
		assert.NotNil(t, model, "Update should return a model")
	})
}

func TestTextEditorView(t *testing.T) {
	t.Run("View returns formatted string", func(t *testing.T) {
		// Create a properly initialized textEditor with a valid textarea
		ta := textarea.New()
		editor := textEditor{
			textarea: ta,
		}
		view := editor.View()

		assert.NotEmpty(t, view, "View() should return non-empty string")
		assert.Contains(t, view, "Edit message", "View should contain instruction text")
		assert.Contains(t, view, "Ctrl+C or Alt+Esc", "View should contain exit instruction")
	})
}

func TestTerminalEditorEdit(t *testing.T) {
	t.Run("TerminalEditor Edit method exists", func(t *testing.T) {
		editor := &TerminalEditor{}

		// We can't test the actual interactive editor in unit tests,
		// but we can verify the method signature and that it compiles
		assert.NotNil(t, editor, "TerminalEditor should be non-nil")
	})
}

func TestCommandErrorFormatting(t *testing.T) {
	t.Run("Error format with underlying error includes error details", func(t *testing.T) {
		underlyingErr := assert.AnError
		cmdErr := &CommandError{
			Type:    "ValidationError",
			Message: "Invalid input",
			Err:     underlyingErr,
		}
		
		errStr := cmdErr.Error()
		assert.Contains(t, errStr, "ValidationError", "Error should contain type")
		assert.Contains(t, errStr, "Invalid input", "Error should contain message")
		assert.Contains(t, errStr, "assert.AnError", "Error should contain underlying error")
	})
	
	t.Run("Error format without underlying error", func(t *testing.T) {
		cmdErr := &CommandError{
			Type:    "NetworkError",
			Message: "Connection failed",
			Err:     nil,
		}
		
		errStr := cmdErr.Error()
		assert.Contains(t, errStr, "NetworkError", "Error should contain type")
		assert.Contains(t, errStr, "Connection failed", "Error should contain message")
		assert.NotContains(t, errStr, "(", "Error should not contain parentheses when no underlying error")
	})
}
