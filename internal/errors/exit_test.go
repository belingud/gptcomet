package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitCode(t *testing.T) {
	configErr := NewConfigError("config", "missing provider", nil, nil)
	networkErr := NewNetworkError("network", "connection failed", nil, nil)
	gitErr := NewGitError("git", "no staged changes", nil, nil)
	apiErr := NewAPIError("api", "request failed", nil, nil)
	validationErr := NewValidationError("validation", "invalid value", nil, nil)
	unknownErr := &GPTCometError{Type: ErrTypeUnknown, Title: "boom", Message: "boom"}

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil is success", err: nil, want: ExitSuccess},
		{name: "config error", err: configErr, want: ExitConfig},
		{name: "network error", err: networkErr, want: ExitNetwork},
		{name: "git error", err: gitErr, want: ExitGit},
		{name: "api error", err: apiErr, want: ExitAPI},
		{name: "validation error", err: validationErr, want: ExitValidation},
		{name: "dependency wrapper around config", err: DependencyCreationError("config manager", ConfigFileNotFoundError("/missing")), want: ExitConfig},
		{name: "dependency wrapper around network", err: DependencyCreationError("api client", networkErr), want: ExitNetwork},
		{name: "validation wrapping plain error", err: NewValidationError("validation", "bad value", errors.New("bad"), nil), want: ExitValidation},
		{name: "validation wrapping unknown causes", err: NewValidationError("validation", "bad value", unknownErr, nil), want: ExitValidation},
		{name: "unknown typed error", err: unknownErr, want: ExitUnknown},
		{name: "wrapped config error", err: fmt.Errorf("load config: %w", configErr), want: ExitConfig},
		{name: "usage error", err: NewUsageError(errors.New("unknown flag: --bogus")), want: ExitUsage},
		{name: "wrapped usage error", err: fmt.Errorf("execute: %w", NewUsageError(errors.New("bad flag"))), want: ExitUsage},
		{name: "plain error", err: errors.New("boom"), want: ExitUnknown},
		{name: "usage error wrapping a typed error wins", err: NewUsageError(configErr), want: ExitUsage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.err); got != tt.want {
				t.Errorf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestNewUsageError(t *testing.T) {
	if err := NewUsageError(nil); err != nil {
		t.Fatalf("NewUsageError(nil) = %v, want nil", err)
	}

	inner := errors.New("unknown flag: --nope")
	err := NewUsageError(inner)
	if !IsUsageError(err) {
		t.Fatal("IsUsageError() = false, want true")
	}
	if got, want := err.Error(), inner.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, inner) {
		t.Error("usage error does not unwrap to the inner error")
	}
}

func TestIsUsageError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "plain", err: errors.New("boom"), want: false},
		{name: "typed", err: NewConfigError("config", "boom", nil, nil), want: false},
		{name: "usage", err: NewUsageError(errors.New("bad flag")), want: true},
		{name: "wrapped usage", err: fmt.Errorf("run: %w", NewUsageError(errors.New("bad flag"))), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUsageError(tt.err); got != tt.want {
				t.Errorf("IsUsageError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
