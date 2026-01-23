package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultSambanovaAPIBase = "https://api.sambanova.ai/v1"
	DefaultSambanovaModel   = "Meta-Llama-3.3-70B-Instruct"
)

// SambanovaLLM implements the LLM interface for SambaNova
type SambanovaLLM struct {
	*OpenAILLM
}

// NewSambanovaLLM creates a new SambanovaLLM
func NewSambanovaLLM(config *types.ClientConfig) *SambanovaLLM {
	BuildStandardConfigSimple(config, DefaultSambanovaAPIBase, DefaultSambanovaModel)

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
			DefaultValue:  DefaultSambanovaAPIBase,
			PromptMessage: "Enter SambaNova API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultSambanovaModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the SambaNova API
func (s *SambanovaLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return s.BaseLLM.MakeRequest(ctx, client, s, message, stream)
}
