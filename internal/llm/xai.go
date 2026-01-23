package llm

import (
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultXAIAPIBase = "https://api.x.ai/v1"
	DefaultXAIModel   = "grok-beta"
)

// XAILLM implements the LLM interface for XAI
type XAILLM struct {
	*OpenAILLM
}

// NewXAILLM creates a new XAILLM
func NewXAILLM(config *types.ClientConfig) *XAILLM {
	BuildStandardConfigSimple(config, DefaultXAIAPIBase, DefaultXAIModel)

	return &XAILLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (x *XAILLM) Name() string {
	return "xai"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (x *XAILLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultXAIAPIBase,
			PromptMessage: "Enter XAI API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultXAIModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}
