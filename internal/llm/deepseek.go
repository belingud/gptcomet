package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultDeepSeekAPIBase = "https://api.deepseek.com/v1"
	DefaultDeepSeekModel   = "deepseek-chat"
)

// DeepSeekLLM implements the LLM interface for DeepSeek
type DeepSeekLLM struct {
	*BaseLLM
}

// NewDeepSeekLLM creates a new DeepSeekLLM
func NewDeepSeekLLM(config *types.ClientConfig) *DeepSeekLLM {
	BuildStandardConfigSimple(config, DefaultDeepSeekAPIBase, DefaultDeepSeekModel)
	return &DeepSeekLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (d *DeepSeekLLM) Name() string {
	return "deepseek"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (d *DeepSeekLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultDeepSeekAPIBase,
			PromptMessage: "Enter DeepSeek API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultDeepSeekModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the DeepSeek API
func (d *DeepSeekLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return d.BaseLLM.MakeRequest(ctx, client, d, message, stream)
}
