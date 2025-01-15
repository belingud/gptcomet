package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// MistralLLM implements the LLM interface for Mistral
type MistralLLM struct {
	*BaseLLM
}

// NewMistralLLM creates a new MistralLLM
func NewMistralLLM(config *types.ClientConfig) *MistralLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.mistral.ai/v1"
	}
	if config.Model == "" {
		config.Model = "mistral-large-latest"
	}
	return &MistralLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (m *MistralLLM) Name() string {
	return "mistral"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (m *MistralLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.mistral.ai/v1",
			PromptMessage: "Enter Mistral API base",
		},
		"model": {
			DefaultValue:  "mistral-large-latest",
			PromptMessage: "Enter model name",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Mistral API
func (m *MistralLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return m.BaseLLM.MakeRequest(ctx, client, m, message, stream)
}
