package errors

import (
	stderrors "errors"
)

// Process exit codes returned by GPTComet commands.
//
// The contract is documented in README.md ("Scripting and exit codes"): stdout
// only carries the result, while the exit code classifies the failure. Scripts
// checking "non-zero means failure" keep working; scripts can additionally
// branch on the concrete code.
const (
	// ExitSuccess means the command completed successfully.
	ExitSuccess = 0
	// ExitUnknown is the fallback for failures that are not classified.
	ExitUnknown = 1
	// ExitUsage means the command line itself was invalid (unknown flag,
	// invalid flag value).
	ExitUsage = 2
	// ExitConfig means the configuration is missing or invalid.
	ExitConfig = 3
	// ExitNetwork means a network, proxy, or connection failure.
	ExitNetwork = 4
	// ExitGit means a Git/VCS failure, including "no staged changes".
	ExitGit = 5
	// ExitAPI means an API/LLM request, response, or parsing failure.
	ExitAPI = 6
	// ExitValidation means invalid input or content, including a diff blocked
	// by sensitive.mode=block.
	ExitValidation = 7
)

// UsageError marks an error as a command-line usage problem so ExitCode can map
// it to ExitUsage. The wrapper only adds classification: Error is the wrapped
// message, so cobra prints the usual text.
type UsageError struct {
	Err error
}

// Error implements the error interface.
func (e *UsageError) Error() string {
	if e == nil || e.Err == nil {
		return "invalid command usage"
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e *UsageError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewUsageError classifies err as a command-line usage error. It returns nil
// when err is nil so it can wrap cobra callback results directly.
func NewUsageError(err error) error {
	if err == nil {
		return nil
	}
	return &UsageError{Err: err}
}

// IsUsageError reports whether err is or wraps a UsageError.
func IsUsageError(err error) bool {
	var usageErr *UsageError
	return stderrors.As(err, &usageErr)
}

// ExitCode maps err to the process exit code described by the constants above.
//
// Structured *GPTCometError values are classified by their Type; wrapped
// errors (created with fmt.Errorf("...: %w", err)) are still recognised.
// Anything else falls back to ExitUnknown.
func ExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if IsUsageError(err) {
		return ExitUsage
	}

	var gptErr *GPTCometError
	if stderrors.As(err, &gptErr) {
		return exitCodeForType(gptErr)
	}

	return ExitUnknown
}

// exitCodeForType maps a structured error to an exit code.
//
// A validation error may only describe which step failed while wrapping a more
// specific classified cause (for example DependencyCreationError wrapping
// ConfigFileNotFoundError). In that case the cause's class is what a script
// needs, so it wins over the generic wrapper type.
func exitCodeForType(gptErr *GPTCometError) int {
	if gptErr.Type == ErrTypeValidation && gptErr.Cause != nil {
		var causeErr *GPTCometError
		if stderrors.As(gptErr.Cause, &causeErr) && causeErr != gptErr {
			switch causeErr.Type {
			case ErrTypeConfig, ErrTypeNetwork, ErrTypeGit, ErrTypeAPI:
				return exitCodeForType(causeErr)
			}
		}
	}

	switch gptErr.Type {
	case ErrTypeConfig:
		return ExitConfig
	case ErrTypeNetwork:
		return ExitNetwork
	case ErrTypeGit:
		return ExitGit
	case ErrTypeAPI:
		return ExitAPI
	case ErrTypeValidation:
		return ExitValidation
	default:
		return ExitUnknown
	}
}
