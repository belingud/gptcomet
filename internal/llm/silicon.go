package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// SiliconLLM implements the LLM interface for Silicon
type SiliconLLM struct {
	*BaseLLM
}

// NewSiliconLLM creates a new SiliconLLM
func NewSiliconLLM(config *types.ClientConfig) *SiliconLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.siliconflow.cn/v1"
	}
	if config.Model == "" {
		config.Model = "Qwen/Qwen2.5-7B-Instruct"
	}

	return &SiliconLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (s *SiliconLLM) Name() string {
	return "silicon"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (s *SiliconLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.siliconflow.cn/v1",
			PromptMessage: "Enter Silicon API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "Qwen/Qwen2.5-7B-Instruct",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Silicon API, formats the response, and returns the result as a string.
func (s *SiliconLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return s.BaseLLM.MakeRequest(ctx, client, s, message, stream)
}
