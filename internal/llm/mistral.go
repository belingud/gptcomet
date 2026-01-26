package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultMistralAPIBase = "https://api.mistral.ai/v1"
	DefaultMistralModel   = "mistral-large-latest"
)

// MistralLLM implements the LLM interface for Mistral
type MistralLLM struct {
	*BaseLLM
}

// NewMistralLLM creates a new MistralLLM
func NewMistralLLM(config *types.ClientConfig) *MistralLLM {
	BuildStandardConfigSimple(config, DefaultMistralAPIBase, DefaultMistralModel)

	return &MistralLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (m *MistralLLM) Name() string {
	return "mistral"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (m *MistralLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultMistralAPIBase,
			PromptMessage: "Enter Mistral API base",
		},
		"model": {
			DefaultValue:  DefaultMistralModel,
			PromptMessage: "Enter model name",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// FormatMessages formats messages for Mistral API
func (m *MistralLLM) FormatMessages(message string) (interface{}, error) {
	messages := []types.Message{}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	payload := map[string]interface{}{
		"model":      m.Config.Model,
		"messages":   messages,
		"max_tokens": m.Config.MaxTokens,
	}
	if m.Config.Temperature != 0 {
		payload["temperature"] = m.Config.Temperature
	}
	if m.Config.TopP != 0 {
		payload["top_p"] = m.Config.TopP
	}
	if m.Config.FrequencyPenalty != 0 {
		payload["frequency_penalty"] = m.Config.FrequencyPenalty
	}
	if m.Config.PresencePenalty != 0 {
		payload["presence_penalty"] = m.Config.PresencePenalty
	}

	return payload, nil
}

// MakeRequest makes a request to the Mistral API
func (m *MistralLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return m.BaseLLM.MakeRequest(ctx, client, m, message, stream)
}
