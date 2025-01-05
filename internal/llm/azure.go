package llm

import (
	"fmt"
	"strings"

	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// AzureLLM implements the LLM interface for Azure OpenAI
type AzureLLM struct {
	*OpenAILLM
}

// NewAzureLLM creates a new AzureLLM
func NewAzureLLM(config *types.ClientConfig) *AzureLLM {
	if config.Model == "" {
		config.Model = "gpt-4o"
	}
	config.CompletionPath = fmt.Sprintf("deployments/%s/chat/completions", config.DeploymentName)

	return &AzureLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (a *AzureLLM) Name() string {
	return "azure"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (a *AzureLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "",
			PromptMessage: "Enter Azure OpenAI endpoint",
		},
		"deployment_name": {
			DefaultValue:  "",
			PromptMessage: "Enter Azure OpenAI deployment name",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "gpt-4o",
			PromptMessage: "Enter deployment name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
		"api_version": {
			DefaultValue:  "2024-02-15-preview",
			PromptMessage: "Enter API version",
		},
	}
}

func (a *AzureLLM) BuildURL() string {
	if a.Config.DeploymentName == "" {
		// Use model name as deployment name if not specified
		a.Config.DeploymentName = a.Config.Model
	}

	// Clean the base URL by removing trailing slashes
	baseURL := strings.TrimSuffix(a.Config.APIBase, "/")

	// Azure OpenAI URL format: {endpoint}/deployments/{deployment-id}/chat/completions?api-version={api-version}
	return fmt.Sprintf("%s/%s?api-version=%s",
		strings.TrimSuffix(baseURL, "?api-version="+a.Config.APIVersion),
		strings.TrimPrefix(a.Config.CompletionPath, "/"),
		a.Config.APIVersion)
}

// BuildHeaders builds request headers for Azure OpenAI
func (a *AzureLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
		"api-key":      a.Config.APIKey,
		"api-version":  a.Config.APIVersion,
	}
	for k, v := range a.Config.ExtraHeaders {
		headers[k] = v
	}
	return headers
}
