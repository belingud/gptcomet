package llm

import (
	"fmt"

	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// CohereLLM implements the LLM interface for Cohere
type CohereLLM struct {
	*OpenAILLM
}

// NewCohereLLM creates a new CohereLLM
func NewCohereLLM(config *types.ClientConfig) *CohereLLM {

	if config.APIBase == "" {
		config.APIBase = "https://api.cohere.com/v2"
	}
	if config.Model == "" {
		config.Model = "command-r-plus"
	}

	return &CohereLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (c *CohereLLM) Name() string {
	return "cohere"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (c *CohereLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.cohere.com/v2",
			PromptMessage: "Enter Cohere API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "command-r-plus",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// FormatMessages formats messages for Cohere API
func (c *CohereLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {

	messages := []map[string]string{
		{
			"role":    "user",
			"content": message,
		},
	}

	payload := map[string]interface{}{
		"messages":    messages,
		"model":       c.Config.Model,
		"stream":      false,
		"max_tokens":  c.Config.MaxTokens,
		"temperature": c.Config.Temperature,
	}

	return payload, nil
}

// GetUsage returns usage information for the provider
func (c *CohereLLM) GetUsage(response []byte) (string, error) {
	usage := gjson.GetBytes(response, "usage")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> input: %d, output: %d",
		usage.Get("input_tokens").Int(),
		usage.Get("output_tokens").Int(),
	), nil
}
