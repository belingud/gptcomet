package llm

import (
	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// SambanovaLLM implements the LLM interface for SambaNova
type SambanovaLLM struct {
	*OpenAILLM
}

// NewSambanovaLLM creates a new SambanovaLLM
func NewSambanovaLLM(config *types.ClientConfig) *SambanovaLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.sambanova.ai/v1"
	}
	if config.Model == "" {
		config.Model = "Meta-Llama-3.3-70B-Instruct"
	}

	return &SambanovaLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (s *SambanovaLLM) Name() string {
	return "sambanova"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (s *SambanovaLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.sambanova.ai/v1",
			PromptMessage: "Enter SambaNova API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "Meta-Llama-3.3-70B-Instruct",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}
