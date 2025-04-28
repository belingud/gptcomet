package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// OpenAILLM implements the LLM interface for OpenAI
type OpenAILLM struct {
	*BaseLLM
}

// NewOpenAILLM creates a new OpenAILLM
func NewOpenAILLM(config *types.ClientConfig) *OpenAILLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-4o"
	}
	return &OpenAILLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (o *OpenAILLM) Name() string {
	return "openai"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (o *OpenAILLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.openai.com/v1",
			PromptMessage: "Enter OpenAI API base URL",
		},
		"model": {
			DefaultValue:  "gpt-4o",
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

// FormatMessages formats messages for OpenAI API
func (o *OpenAILLM) FormatMessages(message string) (interface{}, error) {
	messages := []types.Message{}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	payload := map[string]interface{}{
		"model":    o.Config.Model,
		"messages": messages,
		// "max_tokens":            o.Config.MaxTokens, // OpenAI: This value is now deprecated in favor of max_completion_tokens
		"max_completion_tokens": o.Config.MaxTokens,
	}
	if o.Config.Temperature != 0 {
		payload["temperature"] = o.Config.Temperature
	}
	if o.Config.TopP != 0 {
		payload["top_p"] = o.Config.TopP
	}
	if o.Config.FrequencyPenalty != 0 {
		payload["frequency_penalty"] = o.Config.FrequencyPenalty
	}
	if o.Config.PresencePenalty != 0 {
		payload["presence_penalty"] = o.Config.PresencePenalty
	}

	return payload, nil
}

// BuildURL builds the API URL
func (o *OpenAILLM) BuildURL() string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(o.Config.APIBase, "/"),
		strings.TrimPrefix(*o.Config.CompletionPath, "/"),
	)
}

// ParseResponse parses the response from the API
func (o *OpenAILLM) ParseResponse(response []byte) (string, error) {
	text := gjson.GetBytes(response, o.Config.AnswerPath).String()
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
	}
	return strings.TrimSpace(text), nil
}

// MakeRequest makes a request to the API
func (o *OpenAILLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return o.BaseLLM.MakeRequest(ctx, client, o, message, stream)
}
