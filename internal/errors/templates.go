package errors

import (
	"fmt"
	"os"
	"path/filepath"
)

// Common error templates and constructors

// ConfigFileNotFoundError is returned when the configuration file cannot be found
func ConfigFileNotFoundError(configPath string) *GPTCometError {
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".config", "gptcomet", "gptcomet.yaml")

	message := fmt.Sprintf(ErrMsgConfigNotFound, configPath)
	if configPath == "" {
		message = fmt.Sprintf(ErrMsgConfigNotFound, defaultConfigPath)
	}

	return NewConfigError(
		ErrTitleConfigNotFound,
		message,
		nil,
		[]string{
			SuggInitConfig,
			SuggCheckConfigPath,
		},
	)
}

// APIKeyNotSetError is returned when the API key is not configured
func APIKeyNotSetError(provider string) *GPTCometError {
	envVar := getAPIKeyEnvVar(provider)

	return NewConfigError(
		ErrTitleAPIKeyNotSet,
		fmt.Sprintf(ErrMsgAPIKeyNotSet, provider),
		nil,
		[]string{
			fmt.Sprintf(SuggSetAPIKey, provider),
			fmt.Sprintf(SuggSetEnvVar, envVar),
			fmt.Sprintf(SuggCheckProvider, provider),
		},
	)
}

// NoStagedChangesError is returned when there are no staged changes to commit
func NoStagedChangesError() *GPTCometError {
	return NewGitError(
		ErrTitleNoStagedChanges,
		ErrMsgNoStagedChanges,
		nil,
		[]string{
			SuggStageFiles,
			SuggCheckStatus,
			SuggStageAll,
		},
	)
}

// NetworkConnectionError is returned when network connection fails
func NetworkConnectionError(endpoint string, cause error) *GPTCometError {
	return NewNetworkError(
		ErrTitleNetworkConnection,
		fmt.Sprintf(ErrMsgNetworkConnection, endpoint),
		cause,
		[]string{
			SuggCheckInternet,
			fmt.Sprintf(SuggTryCurl, endpoint),
			SuggCheckProxy,
			SuggUseProxyFlag,
		},
	)
}

// APIRateLimitError is returned when API rate limit is exceeded
func APIRateLimitError(statusCode int, cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleAPIRequest,
		fmt.Sprintf(ErrMsgAPIRateLimit, statusCode),
		cause,
		[]string{
			SuggWaitRetry,
			SuggCheckQuota,
			SuggUpgradePlan,
			SuggUseDifferentProvider,
		},
	)
}

// APIAuthenticationError is returned when API authentication fails
func APIAuthenticationError(provider string, cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleAPIAuth,
		fmt.Sprintf(ErrMsgAPIAuth, provider),
		cause,
		[]string{
			fmt.Sprintf(SuggVerifyAPIKey, provider),
			fmt.Sprintf(SuggSetNewAPIKey, provider),
			SuggCheckKeyValid,
			SuggEnsurePermissions,
		},
	)
}

// InvalidConfigValueError is returned when a configuration value is invalid
func InvalidConfigValueError(key, value, reason string) *GPTCometError {
	return NewConfigError(
		ErrTitleInvalidConfig,
		fmt.Sprintf(ErrMsgInvalidConfig, key, value),
		nil,
		[]string{
			fmt.Sprintf("Reason: %s", reason),
			fmt.Sprintf(SuggCheckCurrentValue, key),
			fmt.Sprintf(SuggSetValidValue, key),
		},
	)
}

// GitRepositoryNotFoundError is returned when not in a git repository
func GitRepositoryNotFoundError() *GPTCometError {
	return NewGitError(
		ErrTitleNotGitRepo,
		ErrMsgNotGitRepo,
		nil,
		[]string{
			SuggInitGit,
			SuggNavigateGit,
		},
	)
}

// GitCommandFailedError is returned when a git command fails
func GitCommandFailedError(command string, cause error) *GPTCometError {
	return NewGitError(
		ErrTitleGitCommand,
		fmt.Sprintf(ErrMsgGitCommand, command),
		cause,
		[]string{
			SuggCheckGitInstalled,
			SuggEnsureInGit,
			SuggCheckGitOutput,
		},
	)
}

// ModelNotFoundError is returned when the specified model is not found
func ModelNotFoundError(provider, model string) *GPTCometError {
	return NewConfigError(
		ErrTitleModelNotFound,
		fmt.Sprintf(ErrMsgModelNotFound, model, provider),
		nil,
		[]string{
			fmt.Sprintf(SuggCheckAvailableModels, provider),
			fmt.Sprintf(SuggSetValidModel, provider),
			SuggVisitDocs,
		},
	)
}

// ProviderCreationError is returned when provider initialization fails
func ProviderCreationError(provider string, cause error) *GPTCometError {
	return NewValidationError(
		ErrTitleProviderCreation,
		fmt.Sprintf(ErrMsgProviderCreation, provider),
		cause,
		[]string{
			SuggCheckProviderConfig,
			fmt.Sprintf(SuggCheckProvider, provider),
			SuggReportIssue,
		},
	)
}

// VCSCreationError is returned when VCS client creation fails
func VCSCreationError(vcsType string, cause error) *GPTCometError {
	return NewGitError(
		ErrTitleVCSCreation,
		fmt.Sprintf(ErrMsgVCSCreation, vcsType),
		cause,
		[]string{
			SuggCheckVCSType,
			SuggCheckGitInstalled,
			SuggReportIssue,
		},
	)
}

// ProxyConfigurationError is returned when proxy configuration fails
func ProxyConfigurationError(cause error) *GPTCometError {
	return NewNetworkError(
		ErrTitleProxyConfiguration,
		ErrMsgProxyConfiguration,
		cause,
		[]string{
			SuggCheckProxyFormat,
			SuggCheckProxy,
			SuggUseProxyFlag,
		},
	)
}

// RequestCreationError is returned when HTTP request creation fails
func RequestCreationError(cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleRequestCreation,
		ErrMsgRequestCreation,
		cause,
		[]string{
			SuggCheckAPIEndpoint,
			SuggReportIssue,
		},
	)
}

// RequestExecutionError is returned when HTTP request execution fails
func RequestExecutionError(cause error) *GPTCometError {
	return NewNetworkError(
		ErrTitleRequestExecution,
		ErrMsgRequestExecution,
		cause,
		[]string{
			SuggCheckInternet,
			SuggCheckProxy,
			SuggWaitRetry,
		},
	)
}

// ResponseParsingError is returned when API response parsing fails
func ResponseParsingError(cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleResponseParsing,
		ErrMsgResponseParsing,
		cause,
		[]string{
			SuggCheckResponseFormat,
			SuggReportIssue,
		},
	)
}

// DependencyCreationError is returned when dependency creation fails
func DependencyCreationError(dependency string, cause error) *GPTCometError {
	return NewValidationError(
		ErrTitleDependencyCreation,
		ErrMsgDependencyCreation,
		cause,
		[]string{
			fmt.Sprintf("Failed to create: %s", dependency),
			SuggReportIssue,
		},
	)
}

// InternalError is returned for unexpected internal errors
func InternalError(message string, cause error) *GPTCometError {
	return &GPTCometError{
		Type:    ErrTypeUnknown,
		Title:   ErrTitleInternalError,
		Message: fmt.Sprintf(ErrMsgInternalError, message),
		Cause:   cause,
		Suggestions: []string{
			SuggReportIssue,
		},
	}
}

// RequestRetryError is returned when a request fails after multiple retry attempts
func RequestRetryError(attempts int, cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleRequestRetry,
		fmt.Sprintf(ErrMsgRequestRetry, attempts),
		cause,
		[]string{
			SuggCheckInternet,
			SuggCheckProviderStatus,
			SuggWaitRetry,
			SuggCheckErrorDetails,
		},
	)
}

// MessageFormattingError is returned when message formatting fails
func MessageFormattingError(cause error) *GPTCometError {
	return NewValidationError(
		ErrTitleMessageFormatting,
		ErrMsgMessageFormatting,
		cause,
		[]string{
			SuggVerifyRequestPayload,
			SuggReportIssue,
		},
	)
}

// RequestMarshalingError is returned when request marshaling fails
func RequestMarshalingError(cause error) *GPTCometError {
	return NewAPIError(
		ErrTitleRequestMarshaling,
		ErrMsgRequestMarshaling,
		cause,
		[]string{
			SuggVerifyRequestPayload,
			SuggReportIssue,
		},
	)
}

// APIStatusError is returned when API returns non-OK status code
func APIStatusError(statusCode int, responseBody string, cause error) *GPTCometError {
	message := fmt.Sprintf(ErrMsgAPIStatusError, statusCode)
	if responseBody != "" {
		message += fmt.Sprintf("\nResponse: %s", responseBody)
	}

	// Determine suggestions based on status code
	suggestions := []string{
		SuggCheckProviderStatus,
		SuggCheckErrorDetails,
	}

	// Add specific suggestions based on status code
	switch {
	case statusCode == 401 || statusCode == 403:
		suggestions = append([]string{
			"Verify your API key is valid and has required permissions",
			SuggCheckKeyValid,
		}, suggestions...)
	case statusCode == 429:
		suggestions = append([]string{
			SuggWaitRetry,
			SuggCheckQuota,
		}, suggestions...)
	case statusCode >= 500:
		suggestions = append([]string{
			"Provider service may be experiencing issues",
			SuggWaitRetry,
		}, suggestions...)
	}

	return NewAPIError(
		ErrTitleAPIStatusError,
		message,
		cause,
		suggestions,
	)
}

// CallbackError is returned when a callback function fails
func CallbackError(cause error) *GPTCometError {
	return NewValidationError(
		ErrTitleCallbackError,
		ErrMsgCallbackError,
		cause,
		[]string{
			SuggCheckErrorDetails,
			SuggReportIssue,
		},
	)
}

// UnsupportedProxySchemeError is returned when an unsupported proxy scheme is used
func UnsupportedProxySchemeError(scheme string) *GPTCometError {
	return NewNetworkError(
		ErrTitleUnsupportedProxy,
		fmt.Sprintf(ErrMsgUnsupportedProxy, scheme),
		nil,
		[]string{
			SuggSupportedProxySchemes,
			SuggCheckProxyFormat,
		},
	)
}

// ProxyURLParseError is returned when proxy URL parsing fails
func ProxyURLParseError(cause error) *GPTCometError {
	return NewNetworkError(
		ErrTitleProxyConfiguration,
		"Failed to parse proxy URL.",
		cause,
		[]string{
			SuggCheckProxyFormat,
			SuggCheckProxy,
		},
	)
}

// getAPIKeyEnvVar returns the environment variable name for the provider's API key
func getAPIKeyEnvVar(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "azure":
		return "AZURE_API_KEY"
	case "claude":
		return "ANTHROPIC_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "cohere":
		return "COHERE_API_KEY"
	default:
		return fmt.Sprintf("%s_API_KEY", provider)
	}
}
