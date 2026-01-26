package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultMinimaxAPIBase = "https://api.minimaxi.com/v1"
	DefaultMinimaxModel   = "MiniMax-M1"
)

// MinimaxLLM implements the LLM interface for Minimax
type MinimaxLLM struct {
	*BaseLLM
}

// NewMinimaxLLM creates a new MinimaxLLM
func NewMinimaxLLM(config *types.ClientConfig) *MinimaxLLM {
	BuildStandardConfigSimple(config, DefaultMinimaxAPIBase, DefaultMinimaxModel)

	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}

	return &MinimaxLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (d *MinimaxLLM) Name() string {
	return "minimax"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (d *MinimaxLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultMinimaxAPIBase,
			PromptMessage: "Enter Minimax API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultMinimaxModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Minimax API
func (d *MinimaxLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return d.BaseLLM.MakeRequest(ctx, client, d, message, stream)
}
