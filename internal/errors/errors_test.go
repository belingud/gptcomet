package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestGPTCometError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *GPTCometError
		wantSub []string
	}{
		{
			name: "Config error with suggestions",
			err: NewConfigError(
				"Test Config Error",
				"Test configuration failed",
				nil,
				[]string{"Suggestion 1", "Suggestion 2"},
			),
			wantSub: []string{"⚙️", "Test Config Error", "Test configuration failed", "Suggestion 1", "Suggestion 2"},
		},
		{
			name: "Network error without suggestions",
			err: NewNetworkError(
				"Test Network Error",
				"Network connection failed",
				nil,
				nil,
			),
			wantSub: []string{"🌐", "Test Network Error", "Network connection failed"},
		},
		{
			name: "Git error with cause",
			err: NewGitError(
				"Test Git Error",
				"Git command failed",
				errors.New("underlying error"),
				[]string{"Check git installation"},
			),
			wantSub: []string{"📦", "Test Git Error", "Git command failed", "Check git installation"},
		},
		{
			name: "API error",
			err: NewAPIError(
				"Test API Error",
				"API request failed",
				nil,
				[]string{"Retry later"},
			),
			wantSub: []string{"🔑", "Test API Error", "API request failed", "Retry later"},
		},
		{
			name: "Validation error",
			err: NewValidationError(
				"Test Validation Error",
				"Invalid input",
				nil,
				[]string{"Check input format"},
			),
			wantSub: []string{"⚠️", "Test Validation Error", "Invalid input", "Check input format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, substr := range tt.wantSub {
				if !strings.Contains(got, substr) {
					t.Errorf("Error() should contain %q, but got:\n%s", substr, got)
				}
			}
		})
	}
}

func TestGPTCometError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewConfigError("Test", "Message", cause, nil)

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestConfigFileNotFoundError(t *testing.T) {
	err := ConfigFileNotFoundError("/path/to/config.yaml")

	if err.Type != ErrTypeConfig {
		t.Errorf("ConfigFileNotFoundError() Type = %v, want %v", err.Type, ErrTypeConfig)
	}

	if !strings.Contains(err.Message, "/path/to/config.yaml") {
		t.Errorf("ConfigFileNotFoundError() should contain path, got: %s", err.Message)
	}

	if len(err.Suggestions) == 0 {
		t.Errorf("ConfigFileNotFoundError() should have suggestions")
	}
}

func TestAPIKeyNotSetError(t *testing.T) {
	providers := []string{"openai", "azure", "claude", "gemini", "cohere"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			err := APIKeyNotSetError(provider)

			if err.Type != ErrTypeConfig {
				t.Errorf("APIKeyNotSetError() Type = %v, want %v", err.Type, ErrTypeConfig)
			}

			if !strings.Contains(err.Message, provider) {
				t.Errorf("APIKeyNotSetError() should mention provider, got: %s", err.Message)
			}

			if len(err.Suggestions) == 0 {
				t.Errorf("APIKeyNotSetError() should have suggestions for %s", provider)
			}
		})
	}
}

func TestNoStagedChangesError(t *testing.T) {
	err := NoStagedChangesError()

	if err.Type != ErrTypeGit {
		t.Errorf("NoStagedChangesError() Type = %v, want %v", err.Type, ErrTypeGit)
	}

	if len(err.Suggestions) == 0 {
		t.Errorf("NoStagedChangesError() should have suggestions")
	}

	// Check for git add suggestion
	hasGitAddSuggestion := false
	for _, s := range err.Suggestions {
		if strings.Contains(s, "git add") {
			hasGitAddSuggestion = true
			break
		}
	}
	if !hasGitAddSuggestion {
		t.Errorf("NoStagedChangesError() should have git add suggestion")
	}
}

func TestNetworkConnectionError(t *testing.T) {
	endpoint := "https://api.example.com"
	cause := errors.New("connection refused")
	err := NetworkConnectionError(endpoint, cause)

	if err.Type != ErrTypeNetwork {
		t.Errorf("NetworkConnectionError() Type = %v, want %v", err.Type, ErrTypeNetwork)
	}

	if !strings.Contains(err.Message, endpoint) {
		t.Errorf("NetworkConnectionError() should contain endpoint, got: %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("NetworkConnectionError() Cause = %v, want %v", err.Cause, cause)
	}

	if len(err.Suggestions) == 0 {
		t.Errorf("NetworkConnectionError() should have suggestions")
	}
}

func TestAPIRateLimitError(t *testing.T) {
	statusCode := 429
	cause := errors.New("rate limit exceeded")
	err := APIRateLimitError(statusCode, cause)

	if err.Type != ErrTypeAPI {
		t.Errorf("APIRateLimitError() Type = %v, want %v", err.Type, ErrTypeAPI)
	}

	if !strings.Contains(err.Message, "429") {
		t.Errorf("APIRateLimitError() should contain status code, got: %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("APIRateLimitError() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestAPIAuthenticationError(t *testing.T) {
	provider := "openai"
	cause := errors.New("unauthorized")
	err := APIAuthenticationError(provider, cause)

	if err.Type != ErrTypeAPI {
		t.Errorf("APIAuthenticationError() Type = %v, want %v", err.Type, ErrTypeAPI)
	}

	if !strings.Contains(err.Message, provider) {
		t.Errorf("APIAuthenticationError() should mention provider, got: %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("APIAuthenticationError() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestInvalidConfigValueError(t *testing.T) {
	key := "model"
	value := "invalid-model"
	reason := "model not found"
	err := InvalidConfigValueError(key, value, reason)

	if err.Type != ErrTypeConfig {
		t.Errorf("InvalidConfigValueError() Type = %v, want %v", err.Type, ErrTypeConfig)
	}

	if !strings.Contains(err.Message, key) || !strings.Contains(err.Message, value) {
		t.Errorf("InvalidConfigValueError() should contain key and value, got: %s", err.Message)
	}

	// Check if suggestions contain the reason
	hasReason := false
	for _, s := range err.Suggestions {
		if strings.Contains(s, reason) {
			hasReason = true
			break
		}
	}
	if !hasReason {
		t.Errorf("InvalidConfigValueError() suggestions should contain reason")
	}
}

func TestGitRepositoryNotFoundError(t *testing.T) {
	err := GitRepositoryNotFoundError()

	if err.Type != ErrTypeGit {
		t.Errorf("GitRepositoryNotFoundError() Type = %v, want %v", err.Type, ErrTypeGit)
	}

	if len(err.Suggestions) == 0 {
		t.Errorf("GitRepositoryNotFoundError() should have suggestions")
	}
}

func TestGitCommandFailedError(t *testing.T) {
	command := "git status"
	cause := errors.New("not a git repository")
	err := GitCommandFailedError(command, cause)

	if err.Type != ErrTypeGit {
		t.Errorf("GitCommandFailedError() Type = %v, want %v", err.Type, ErrTypeGit)
	}

	if !strings.Contains(err.Message, command) {
		t.Errorf("GitCommandFailedError() should contain command, got: %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("GitCommandFailedError() Cause = %v, want %v", err.Cause, cause)
	}
}

func TestModelNotFoundError(t *testing.T) {
	provider := "openai"
	model := "gpt-5"
	err := ModelNotFoundError(provider, model)

	if err.Type != ErrTypeConfig {
		t.Errorf("ModelNotFoundError() Type = %v, want %v", err.Type, ErrTypeConfig)
	}

	if !strings.Contains(err.Message, model) || !strings.Contains(err.Message, provider) {
		t.Errorf("ModelNotFoundError() should contain model and provider, got: %s", err.Message)
	}
}

func TestGetAPIKeyEnvVar(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"openai", "OPENAI_API_KEY"},
		{"azure", "AZURE_API_KEY"},
		{"claude", "ANTHROPIC_API_KEY"},
		{"gemini", "GEMINI_API_KEY"},
		{"cohere", "COHERE_API_KEY"},
		{"unknown", "unknown_API_KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			if got := getAPIKeyEnvVar(tt.provider); got != tt.want {
				t.Errorf("getAPIKeyEnvVar(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		})
	}
}

func TestGetErrorIcon(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      string
	}{
		{ErrTypeNetwork, "🌐"},
		{ErrTypeConfig, "⚙️"},
		{ErrTypeGit, "📦"},
		{ErrTypeAPI, "🔑"},
		{ErrTypeValidation, "⚠️"},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			if got := getErrorIcon(tt.errorType); got != tt.want {
				t.Errorf("getErrorIcon(%v) = %q, want %q", tt.errorType, got, tt.want)
			}
		})
	}
}

func TestGetSuggestionIcon(t *testing.T) {
	// Test first few icons
	for i := 0; i < 10; i++ {
		icon := getSuggestionIcon(i)
		if icon == "" {
			t.Errorf("getSuggestionIcon(%d) should return non-empty string", i)
		}
	}
}
