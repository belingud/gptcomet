package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	// DefaultLMStudioAPIBase is the default local LM Studio OpenAI-compatible endpoint.
	DefaultLMStudioAPIBase = "http://localhost:1234/v1"
	// DefaultLMStudioModel is a placeholder for the model loaded in LM Studio.
	DefaultLMStudioModel = "local-model"
)

// LMStudioLLM implements the LLM interface for a local LM Studio server.
type LMStudioLLM struct {
	*BaseLLM
}

// NewLMStudioLLM creates a new LMStudioLLM.
func NewLMStudioLLM(config *types.ClientConfig) *LMStudioLLM {
	BuildStandardConfigSimple(config, DefaultLMStudioAPIBase, DefaultLMStudioModel)

	return &LMStudioLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the provider name.
func (l *LMStudioLLM) Name() string {
	return "lmstudio"
}

// GetRequiredConfig returns provider-specific configuration requirements.
func (l *LMStudioLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultLMStudioAPIBase,
			PromptMessage: "Enter LM Studio API base",
		},
		"model": {
			DefaultValue:  DefaultLMStudioModel,
			PromptMessage: "Enter the model loaded in LM Studio",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key (optional, press Enter to skip)",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the LM Studio OpenAI-compatible API.
func (l *LMStudioLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return l.BaseLLM.MakeRequest(ctx, client, l, message, stream)
}
