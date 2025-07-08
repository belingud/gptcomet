package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// MinimaxLLM implements the LLM interface for Minimax
type MinimaxLLM struct {
	*BaseLLM
}

// NewMinimaxLLM creates a new MinimaxLLM
func NewMinimaxLLM(config *types.ClientConfig) *MinimaxLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.minimaxi.com/v1"
	}
	if config.Model == "" {
		config.Model = "MiniMax-M1"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1024
	}

	if config.CompletionPath == nil {
		defaultPath := "chat/completions"
		config.CompletionPath = &defaultPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
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
			DefaultValue:  "https://api.minimaxi.com/v1",
			PromptMessage: "Enter Minimax API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "MiniMax-M1",
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
