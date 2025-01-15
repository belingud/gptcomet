package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultAI21Model = "jamba-1.5-large"
)

// AI21LLM implements the LLM interface for Groq
type AI21LLM struct {
	*OpenAILLM
}

// NewAI21LLM creates a new AI21LLM
func NewAI21LLM(config *types.ClientConfig) *AI21LLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.ai21.com/studio/v1"
	}
	if config.Model == "" {
		config.Model = DefaultAI21Model
	}

	return &AI21LLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (a *AI21LLM) Name() string {
	return "ai21"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (a *AI21LLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.ai21.com/studio/v1",
			PromptMessage: "Enter AI21 API base URL",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter AI21 API key",
		},
		"model": {
			DefaultValue:  DefaultAI21Model,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// BuildHeaders builds request headers for Groq
func (a *AI21LLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", a.Config.APIKey),
	}
	return headers
}

func (a *AI21LLM) FormatMessages(message string) (interface{}, error) {
	messages := []interface{}{types.Message{
		Role:    "user",
		Content: message,
	}}

	payload := map[string]interface{}{
		"model":      a.Config.Model,
		"messages":   messages,
		"max_tokens": a.Config.MaxTokens,
	}
	if a.Config.Temperature != 0 {
		payload["temperature"] = a.Config.Temperature
	}
	if a.Config.TopP != 0 {
		payload["topP"] = a.Config.TopP
	}
	return payload, nil
}

// MakeRequest implements the LLM interface for AI21
func (a *AI21LLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return a.BaseLLM.MakeRequest(ctx, client, a, message, stream)
}
