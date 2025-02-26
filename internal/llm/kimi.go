package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// KimiLLM implements the LLM interface for Kimi
type KimiLLM struct {
	*OpenAILLM
}

// NewKimiLLM creates a new KimiLLM
func NewKimiLLM(config *types.ClientConfig) *KimiLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.moonshot.cn/v1"
	}
	if config.Model == "" {
		config.Model = "moonshot-v1-8k"
	}

	return &KimiLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (k *KimiLLM) Name() string {
	return "kimi"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (k *KimiLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.moonshot.cn/v1",
			PromptMessage: "Enter Kimi API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "moonshot-v1-8k",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Kimi API
func (k *KimiLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return k.BaseLLM.MakeRequest(ctx, client, k, message, stream)
}
