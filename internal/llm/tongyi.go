package llm

import (
	"fmt"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// TongyiLLM implements the LLM interface for Tongyi (DashScope)
type TongyiLLM struct {
	*OpenAILLM
}

// NewTongyiLLM creates a new TongyiLLM
func NewTongyiLLM(config *types.ClientConfig) *TongyiLLM {
	if config.APIBase == "" {
		config.APIBase = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if config.CompletionPath == "" {
		config.CompletionPath = "chat/completions"
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
	}
	if config.Model == "" {
		config.Model = "qwen-turbo"
	}

	return &TongyiLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (t *TongyiLLM) Name() string {
	return "tongyi"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (t *TongyiLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://dashscope.aliyuncs.com/compatible-mode/v1",
			PromptMessage: "Enter Tongyi API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "qwen-turbo",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// BuildHeaders builds request headers
func (t *TongyiLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", t.Config.APIKey),
	}
	for k, v := range t.Config.ExtraHeaders {
		headers[k] = v
	}
	return headers
}

// GetUsage returns usage information for the provider
func (t *TongyiLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usage")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> input: %d, output: %d, total: %d",
		usage.Get("prompt_tokens").Int(),
		usage.Get("completion_tokens").Int(),
		usage.Get("total_tokens").Int(),
	), nil
}
