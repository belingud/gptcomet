package llm

import (
	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// MistralLLM implements the LLM interface for Mistral
type MistralLLM struct {
	*OpenAILLM
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
		OpenAILLM: NewOpenAILLM(config),
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
