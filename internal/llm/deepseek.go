package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// DeepSeekLLM implements the LLM interface for DeepSeek
type DeepSeekLLM struct {
	*BaseLLM
}

// NewDeepSeekLLM creates a new DeepSeekLLM
func NewDeepSeekLLM(config *types.ClientConfig) *DeepSeekLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.deepseek.com/v1"
	}
	if config.Model == "" {
		config.Model = "deepseek-chat"
	}

	if config.CompletionPath == nil {
		defaultPath := "chat/completions"
		config.CompletionPath = &defaultPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
	}

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
			DefaultValue:  "https://api.deepseek.com/v1",
			PromptMessage: "Enter DeepSeek API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "deepseek-chat",
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
