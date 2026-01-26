package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultOpenRouterAPIBase = "https://openrouter.ai/api/v1"
	DefaultOpenRouterModel   = "meta-llama/llama-3.1-70b-instruct:free"
)

// OpenRouterLLM implements the LLM interface for OpenRouter
type OpenRouterLLM struct {
	*BaseLLM
}

// NewOpenRouterLLM creates a new OpenRouterLLM
func NewOpenRouterLLM(config *types.ClientConfig) *OpenRouterLLM {
	BuildStandardConfigSimple(config, DefaultOpenRouterAPIBase, DefaultOpenRouterModel)

	return &OpenRouterLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (o *OpenRouterLLM) Name() string {
	return "openrouter"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (o *OpenRouterLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultOpenRouterAPIBase,
			PromptMessage: "Enter OpenRouter API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultOpenRouterModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// BuildHeaders builds request headers for OpenRouter API.
// It includes the default headers from BuildHeaders() and adds
// HTTP-Referer and X-Title headers.
func (o *OpenRouterLLM) BuildHeaders() map[string]string {
	headers := o.BaseLLM.BuildHeaders()
	headers["HTTP-Referer"] = "https://github.com/belingud/gptcomet"
	headers["X-Title"] = "GPTComet"
	return headers
}

// MakeRequest makes a request to the OpenRouter API
func (o *OpenRouterLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return o.BaseLLM.MakeRequest(ctx, client, o, message, stream)
}
