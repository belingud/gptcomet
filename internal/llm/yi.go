package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultModel = "yi-lightning"
)

// YiLLM implements the LLM interface for Yi
type YiLLM struct {
	*BaseLLM
}

// NewYiLLM creates a new YiLLM
func NewYiLLM(config *types.ClientConfig) *YiLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.lingyiwanwu.com/v1"
	}
	if config.Model == "" {
		config.Model = DefaultModel
	}

	return &YiLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (y *YiLLM) Name() string {
	return "yi"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (y *YiLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.lingyiwanwu.com/v1",
			PromptMessage: "Enter Yi API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Yi API, formats the response, and returns the result as a string.
func (y *YiLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return y.BaseLLM.MakeRequest(ctx, client, y, message, stream)
}
