package llm

import (
	"fmt"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

const (
	DefaultTongyiModel   = "qwen-turbo"
	DefaultTongyiAPIBase = "https://dashscope.aliyuncs.com/compatible-mode/v1"
)

// TongyiLLM implements the LLM interface for Tongyi (DashScope)
type TongyiLLM struct {
	*OpenAILLM
}

// NewTongyiLLM creates a new TongyiLLM
func NewTongyiLLM(config *types.ClientConfig) *TongyiLLM {
	BuildStandardConfigSimple(config, DefaultTongyiAPIBase, DefaultTongyiModel)

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
			DefaultValue:  DefaultTongyiAPIBase,
			PromptMessage: "Enter Tongyi API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultTongyiModel,
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
