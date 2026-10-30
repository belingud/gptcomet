package errors

import (
	"errors"
	"strings"
	"testing"
	"time"
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

func TestHookErrors(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		err := HookAlreadyExistsError("/repo/.git/hooks/prepare-commit-msg", "/repo/.git/hooks/prepare-commit-msg.gptcomet-backup")
		if err.Type != ErrTypeGit {
			t.Errorf("HookAlreadyExistsError() Type = %v, want %v", err.Type, ErrTypeGit)
		}
		if !strings.Contains(err.Message, "prepare-commit-msg") {
			t.Errorf("HookAlreadyExistsError() should contain hook path, got: %s", err.Message)
		}
		if !strings.Contains(err.Error(), "--force") {
			t.Errorf("HookAlreadyExistsError() should suggest --force, got: %s", err.Error())
		}
	})

	t.Run("backup exists", func(t *testing.T) {
		err := HookBackupExistsError("/repo/.git/hooks/prepare-commit-msg.gptcomet-backup")
		if err.Type != ErrTypeValidation {
			t.Errorf("HookBackupExistsError() Type = %v, want %v", err.Type, ErrTypeValidation)
		}
	})

	t.Run("install wraps cause", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := HookInstallError("/repo/.git/hooks/prepare-commit-msg", cause)
		if err.Type != ErrTypeGit {
			t.Errorf("HookInstallError() Type = %v, want %v", err.Type, ErrTypeGit)
		}
		if !errors.Is(err, cause) {
			t.Errorf("HookInstallError() should wrap cause")
		}
	})

	t.Run("uninstall wraps cause", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := HookUninstallError("/repo/.git/hooks/prepare-commit-msg", cause)
		if err.Type != ErrTypeGit {
			t.Errorf("HookUninstallError() Type = %v, want %v", err.Type, ErrTypeGit)
		}
		if !errors.Is(err, cause) {
			t.Errorf("HookUninstallError() should wrap cause")
		}
	})

	t.Run("status wraps cause", func(t *testing.T) {
		cause := errors.New("permission denied")
		err := HookStatusError("/repo/.git/hooks/prepare-commit-msg", cause)
		if err.Type != ErrTypeGit {
			t.Errorf("HookStatusError() Type = %v, want %v", err.Type, ErrTypeGit)
		}
		if !errors.Is(err, cause) {
			t.Errorf("HookStatusError() should wrap cause")
		}
	})
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
		{"bedrock", "AWS_BEARER_TOKEN_BEDROCK"},
		{"cloudflare", "CLOUDFLARE_API_TOKEN"},
		{"doubao", "ARK_API_KEY"},
		{"unknown", "UNKNOWN_API_KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			if got := GetAPIKeyEnvVar(tt.provider); got != tt.want {
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

func TestProviderNotFoundError(t *testing.T) {
	err := ProviderNotFoundError("not-a-provider")

	if err.Type != ErrTypeValidation {
		t.Errorf("ProviderNotFoundError type = %q, want %q", err.Type, ErrTypeValidation)
	}
	if err.Title != ErrTitleUnknownProvider {
		t.Errorf("ProviderNotFoundError title = %q, want %q", err.Title, ErrTitleUnknownProvider)
	}

	rendered := err.Error()
	for _, want := range []string{
		"not-a-provider",
		"gmsg provider list",
		"gmsg provider new",
		"config set not-a-provider.api_key",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("ProviderNotFoundError() missing %q in:\n%s", want, rendered)
		}
	}
}

func TestNoProviderSelectedError(t *testing.T) {
	err := NoProviderSelectedError()

	if err.Type != ErrTypeConfig {
		t.Errorf("NoProviderSelectedError type = %q, want %q", err.Type, ErrTypeConfig)
	}

	rendered := err.Error()
	for _, want := range []string{
		"No Provider Selected",
		"gmsg provider list",
		"gmsg provider use <name>",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("NoProviderSelectedError() missing %q in:\n%s", want, rendered)
		}
	}
}

func TestProviderTestError(t *testing.T) {
	tests := []struct {
		name        string
		cause       error
		wantSubs    []string
		notWantSubs []string
	}{
		{
			name: "authentication failure",
			cause: NewAPIError(
				ErrTitleRequestRetry,
				"Request failed after 1 attempts.",
				errors.New("request failed with status 401: Incorrect API key provided"),
				nil,
			),
			wantSubs: []string{
				"did not respond successfully after 1.25s",
				"status 401",
				"Verify API key",
				"Set new API key",
			},
		},
		{
			name:     "rate limit",
			cause:    errors.New("request failed with status 429: rate limit exceeded"),
			wantSubs: []string{"Wait a few minutes and retry", "Check your API quota"},
		},
		{
			name:     "provider outage",
			cause:    errors.New("request failed with status 503: service unavailable"),
			wantSubs: []string{"provider's API service is operational"},
		},
		{
			name:     "network with credential in url",
			cause:    errors.New(`Get "https://generativelanguage.googleapis.com/v1beta/models?key=SECRETKEY123": dial tcp: no such host`),
			wantSubs: []string{"Check internet connection", "key=***", "no such host"},
			notWantSubs: []string{
				"SECRETKEY123",
			},
		},
		{
			name:     "generic failure",
			cause:    errors.New("something unexpected happened"),
			wantSubs: []string{"Check if the provider's API service is operational", "gmsg provider test openai"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ProviderTestError("openai", 1250*time.Millisecond, tt.cause)
			if err.Type != ErrTypeAPI {
				t.Errorf("ProviderTestError type = %q, want %q", err.Type, ErrTypeAPI)
			}

			rendered := err.Error()
			for _, want := range tt.wantSubs {
				if !strings.Contains(rendered, want) {
					t.Errorf("ProviderTestError() missing %q in:\n%s", want, rendered)
				}
			}
			for _, notWant := range tt.notWantSubs {
				if strings.Contains(rendered, notWant) {
					t.Errorf("ProviderTestError() leaked %q in:\n%s", notWant, rendered)
				}
			}

			var got *GPTCometError
			if !errors.As(err, &got) {
				t.Errorf("ProviderTestError() is not a *GPTCometError")
			}
		})
	}
}

func TestSanitizeErrorDetail(t *testing.T) {
	long := strings.Repeat("detail ", 100)

	tests := []struct {
		name     string
		cause    error
		want     string
		wantCaps bool
	}{
		{
			name:  "nil cause",
			cause: nil,
			want:  "",
		},
		{
			name:  "query secret is redacted",
			cause: errors.New("Get https://example.com/v1?key=abc123&model=x: boom"),
			want:  "Get https://example.com/v1?key=***&model=x: boom",
		},
		{
			name:  "newlines collapse",
			cause: errors.New("line one\nline two"),
			want:  "line one line two",
		},
		{
			name:     "long detail is capped",
			cause:    errors.New(long),
			wantCaps: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeErrorDetail(tt.cause)
			if tt.wantCaps {
				if len([]rune(got)) != providerTestDetailLimit+1 {
					t.Errorf("sanitizeErrorDetail() rune length = %d, want %d", len([]rune(got)), providerTestDetailLimit+1)
				}
				return
			}
			if got != tt.want {
				t.Errorf("sanitizeErrorDetail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProviderNotConfiguredError(t *testing.T) {
	err := ProviderNotConfiguredError("deepseek")

	if err.Type != ErrTypeConfig {
		t.Errorf("ProviderNotConfiguredError type = %q, want %q", err.Type, ErrTypeConfig)
	}

	rendered := err.Error()
	for _, want := range []string{
		"deepseek",
		"gmsg provider use deepseek",
		"gmsg provider new",
		"gmsg config get deepseek.api_base",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("ProviderNotConfiguredError() missing %q in:\n%s", want, rendered)
		}
	}
}
