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

	message := fmt.Sprintf("Cannot find configuration file at: %s", configPath)
	if configPath == "" {
		message = fmt.Sprintf("Cannot find configuration file at: %s", defaultConfigPath)
	}

	return NewConfigError(
		"Configuration File Not Found",
		message,
		nil,
		[]string{
			"Run 'gptcomet config init' to create a default configuration",
			"Run 'gptcomet config path' to see expected config location",
		},
	)
}

// APIKeyNotSetError is returned when the API key is not configured
func APIKeyNotSetError(provider string) *GPTCometError {
	envVar := getAPIKeyEnvVar(provider)

	return NewConfigError(
		"API Key Not Configured",
		fmt.Sprintf("Provider '%s' requires an API key, but none was found.", provider),
		nil,
		[]string{
			fmt.Sprintf("Set API key: gptcomet config set %s.api_key <your-key>", provider),
			fmt.Sprintf("Or set env var: export %s=<your-key>", envVar),
			fmt.Sprintf("Check provider: gptcomet config get %s", provider),
		},
	)
}

// NoStagedChangesError is returned when there are no staged changes to commit
func NoStagedChangesError() *GPTCometError {
	return NewGitError(
		"No Staged Changes Found",
		"There are no staged changes to generate a commit message for.",
		nil,
		[]string{
			"Stage files: git add <files>",
			"Check status: git status",
			"Stage all: git add -A",
		},
	)
}

// NetworkConnectionError is returned when network connection fails
func NetworkConnectionError(endpoint string, cause error) *GPTCometError {
	return NewNetworkError(
		"Network Connection Failed",
		fmt.Sprintf("Cannot connect to API endpoint '%s'", endpoint),
		cause,
		[]string{
			"Check internet connection",
			fmt.Sprintf("Try: curl -I %s", endpoint),
			"Check proxy settings: gptcomet config get proxy",
			"Use --proxy flag to set proxy if needed",
		},
	)
}

// APIRateLimitError is returned when API rate limit is exceeded
func APIRateLimitError(statusCode int, cause error) *GPTCometError {
	return NewAPIError(
		"API Request Failed",
		fmt.Sprintf("The API returned an error: %d Rate limit exceeded. Please try again later.", statusCode),
		cause,
		[]string{
			"Wait a few minutes and retry",
			"Check your API quota and usage",
			"Consider upgrading your API plan",
			"Or use a different provider",
		},
	)
}

// APIAuthenticationError is returned when API authentication fails
func APIAuthenticationError(provider string, cause error) *GPTCometError {
	return NewAPIError(
		"API Authentication Failed",
		fmt.Sprintf("Authentication with provider '%s' failed. Please check your API key.", provider),
		cause,
		[]string{
			fmt.Sprintf("Verify API key: gptcomet config get %s.api_key", provider),
			fmt.Sprintf("Set new API key: gptcomet config set %s.api_key <your-key>", provider),
			"Check if the API key is valid and not expired",
			"Ensure the API key has the required permissions",
		},
	)
}

// InvalidConfigValueError is returned when a configuration value is invalid
func InvalidConfigValueError(key, value, reason string) *GPTCometError {
	return NewConfigError(
		"Invalid Configuration Value",
		fmt.Sprintf("Configuration key '%s' has an invalid value: %s", key, value),
		nil,
		[]string{
			fmt.Sprintf("Reason: %s", reason),
			fmt.Sprintf("Check current value: gptcomet config get %s", key),
			fmt.Sprintf("Set a valid value: gptcomet config set %s <valid-value>", key),
		},
	)
}

// GitRepositoryNotFoundError is returned when not in a git repository
func GitRepositoryNotFoundError() *GPTCometError {
	return NewGitError(
		"Not a Git Repository",
		"The current directory is not a git repository.",
		nil,
		[]string{
			"Initialize a git repository: git init",
			"Or navigate to a git repository",
		},
	)
}

// GitCommandFailedError is returned when a git command fails
func GitCommandFailedError(command string, cause error) *GPTCometError {
	return NewGitError(
		"Git Command Failed",
		fmt.Sprintf("Git command '%s' failed.", command),
		cause,
		[]string{
			"Check if git is installed: git --version",
			"Ensure you're in a git repository",
			"Check the git command output for more details",
		},
	)
}

// ModelNotFoundError is returned when the specified model is not found
func ModelNotFoundError(provider, model string) *GPTCometError {
	return NewConfigError(
		"Model Not Found",
		fmt.Sprintf("Model '%s' is not available for provider '%s'.", model, provider),
		nil,
		[]string{
			fmt.Sprintf("Check available models for %s", provider),
			fmt.Sprintf("Set a valid model: gptcomet config set %s.model <valid-model>", provider),
			"Visit provider documentation for model list",
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
