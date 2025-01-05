package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// ClaudeLLM is the Claude LLM provider implementation
type ClaudeLLM struct {
	*BaseLLM
}

// NewClaudeLLM creates a new ClaudeLLM
func NewClaudeLLM(config *types.ClientConfig) *ClaudeLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.anthropic.com/v1"
	}
	if config.Model == "" {
		config.Model = "claude-3-sonnet"
	}
	if config.CompletionPath == "" {
		config.CompletionPath = "messages"
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "content.0.text"
	}
	if config.AnthropicVersion == "" {
		config.AnthropicVersion = "2024-01-01" // 使用最新的稳定版本
	}

	return &ClaudeLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (c *ClaudeLLM) Name() string {
	return "Claude"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (c *ClaudeLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.anthropic.com/v1",
			PromptMessage: "Enter Claude API base",
		},
		"model": {
			DefaultValue:  "claude-3-sonnet",
			PromptMessage: "Enter model name",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"anthropic_version": {
			DefaultValue:  "2024-01-01",
			PromptMessage: "Enter Anthropic API version",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// FormatMessages formats messages for Claude API
func (c *ClaudeLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	messages := make([]map[string]interface{}, 0, len(history)+1)
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": message,
	})

	payload := map[string]interface{}{
		"model":             c.Config.Model,
		"messages":          messages,
		"max_tokens":        c.Config.MaxTokens,
		"temperature":       c.Config.Temperature,
		"top_p":             c.Config.TopP,
		"frequency_penalty": c.Config.FrequencyPenalty,
		"presence_penalty":  c.Config.PresencePenalty,
	}

	return payload, nil
}

// BuildURL builds the API URL
func (c *ClaudeLLM) BuildURL() string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(c.Config.APIBase, "/"), strings.TrimPrefix(c.Config.CompletionPath, "/"))
}

// BuildHeaders builds request headers
func (c *ClaudeLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":      "application/json",
		"anthropic-version": c.Config.AnthropicVersion,
		"x-api-key":         c.Config.APIKey,
	}
	for k, v := range c.Config.ExtraHeaders {
		headers[k] = v
	}
	return headers
}

// ParseResponse parses the response from the API
func (c *ClaudeLLM) ParseResponse(response []byte) (string, error) {
	text := gjson.GetBytes(response, c.Config.AnswerPath).String()
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
	}
	return strings.TrimSpace(text), nil
}

// GetUsage returns usage information for the provider
func (c *ClaudeLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usage")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> input tokens: %d, output tokens: %d",
		usage.Get("input_tokens").Int(),
		usage.Get("output_tokens").Int(),
	), nil
}

// MakeRequest makes a request to the API
func (c *ClaudeLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	return c.BaseLLM.MakeRequest(ctx, client, c, message, history)
}
