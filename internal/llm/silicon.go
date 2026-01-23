package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultSiliconAPIBase = "https://api.siliconflow.cn/v1"
	DefaultSiliconModel   = "Qwen/Qwen2.5-7B-Instruct"
)

// SiliconLLM implements the LLM interface for Silicon
type SiliconLLM struct {
	*BaseLLM
}

// NewSiliconLLM creates a new SiliconLLM
func NewSiliconLLM(config *types.ClientConfig) *SiliconLLM {
	BuildStandardConfigSimple(config, DefaultSiliconAPIBase, DefaultSiliconModel)

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
			DefaultValue:  DefaultSiliconAPIBase,
			PromptMessage: "Enter Silicon API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultSiliconModel,
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
