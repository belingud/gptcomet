package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrTypeNetwork    ErrorType = "network"
	ErrTypeConfig     ErrorType = "config"
	ErrTypeGit        ErrorType = "git"
	ErrTypeAPI        ErrorType = "api"
	ErrTypeValidation ErrorType = "validation"
	ErrTypeUnknown    ErrorType = "unknown"
)

// GPTCometError is a structured error type that provides detailed error information
// including the error type, message, underlying cause, suggestions for fixing it,
// and optional documentation URL.
type GPTCometError struct {
	Type        ErrorType
	Title       string
	Message     string
	Cause       error
	Suggestions []string
	DocsURL     string
}

// Error implements the error interface
func (e *GPTCometError) Error() string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("\n%s %s\n\n", getErrorIcon(e.Type), e.Title))

	// Message
	sb.WriteString(fmt.Sprintf("%s\n\n", e.Message))

	// Suggestions
	if len(e.Suggestions) > 0 {
		sb.WriteString("What to do:\n")
		for i, suggestion := range e.Suggestions {
			sb.WriteString(fmt.Sprintf("  %s %s\n", getSuggestionIcon(i), suggestion))
		}
		sb.WriteString("\n")
	}

	// Docs URL
	if e.DocsURL != "" {
		sb.WriteString(fmt.Sprintf("Docs: %s\n", e.DocsURL))
	}

	return sb.String()
}

// Unwrap returns the underlying cause
func (e *GPTCometError) Unwrap() error {
	return e.Cause
}

// WrapError creates a new GPTCometError wrapping an existing error
func WrapError(err error, title, message string) *GPTCometError {
	if err == nil {
		return nil
	}
	return &GPTCometError{
		Type:    ErrTypeUnknown,
		Title:   title,
		Message: message,
		Cause:   err,
	}
}

// Helper functions to create specific error types

// NewConfigError creates a new configuration-related error
func NewConfigError(title, message string, cause error, suggestions []string) *GPTCometError {
	return &GPTCometError{
		Type:        ErrTypeConfig,
		Title:       title,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/belingud/gptcomet#configuration",
	}
}

// NewNetworkError creates a new network-related error
func NewNetworkError(title, message string, cause error, suggestions []string) *GPTCometError {
	return &GPTCometError{
		Type:        ErrTypeNetwork,
		Title:       title,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/belingud/gptcomet#network-configuration",
	}
}

// NewGitError creates a new git-related error
func NewGitError(title, message string, cause error, suggestions []string) *GPTCometError {
	return &GPTCometError{
		Type:        ErrTypeGit,
		Title:       title,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/belingud/gptcomet#git-requirements",
	}
}

// NewAPIError creates a new API-related error
func NewAPIError(title, message string, cause error, suggestions []string) *GPTCometError {
	return &GPTCometError{
		Type:        ErrTypeAPI,
		Title:       title,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/belingud/gptcomet#api-configuration",
	}
}

// NewValidationError creates a new validation-related error
func NewValidationError(title, message string, cause error, suggestions []string) *GPTCometError {
	return &GPTCometError{
		Type:        ErrTypeValidation,
		Title:       title,
		Message:     message,
		Cause:       cause,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/belingud/gptcomet#usage",
	}
}

// getErrorIcon returns an appropriate icon for the error type
func getErrorIcon(errorType ErrorType) string {
	switch errorType {
	case ErrTypeNetwork:
		return "🌐"
	case ErrTypeConfig:
		return "⚙️"
	case ErrTypeGit:
		return "📦"
	case ErrTypeAPI:
		return "🔑"
	case ErrTypeValidation:
		return "⚠️"
	case ErrTypeUnknown:
		return "❓"
	default:
		return "❌"
	}
}

// getSuggestionIcon returns an appropriate icon for the suggestion
func getSuggestionIcon(index int) string {
	// Use different icons for variety
	icons := []string{"•", "•", "•", "•"}
	if index < len(icons) {
		return icons[index]
	}
	return "•"
}
