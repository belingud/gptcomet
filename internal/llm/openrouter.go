package llm

import (
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

var DefaultOpenrouterModel = "meta-llama/llama-3.1-70b-instruct:free"

// OpenRouterLLM implements the LLM interface for OpenRouter
type OpenRouterLLM struct {
	*OpenAILLM
}

// NewOpenRouterLLM creates a new OpenRouterLLM
func NewOpenRouterLLM(config *types.ClientConfig) *OpenRouterLLM {
	if config.APIBase == "" {
		config.APIBase = "https://openrouter.ai/api/v1"
	}
	if config.Model == "" {
		config.Model = DefaultOpenrouterModel
	}

	return &OpenRouterLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (o *OpenRouterLLM) Name() string {
	return "openrouter"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (o *OpenRouterLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://openrouter.ai/api/v1",
			PromptMessage: "Enter OpenRouter API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultOpenrouterModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// BuildHeaders overrides the parent's BuildHeaders to add OpenRouter specific headers
func (o *OpenRouterLLM) BuildHeaders() map[string]string {
	headers := o.OpenAILLM.BuildHeaders()
	headers["HTTP-Referer"] = "https://github.com/belingud/gptcomet"
	headers["X-Title"] = "GPTComet"
	return headers
}
