package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	// DefaultQianfanAPIBase is the OpenAI-compatible Baidu Qianfan v2 endpoint.
	DefaultQianfanAPIBase = "https://qianfan.baidubce.com/v2"
	// DefaultQianfanModel is the default ERNIE model used with Qianfan.
	DefaultQianfanModel = "ernie-5.0"
)

// QianfanLLM implements the LLM interface for Baidu Qianfan (ERNIE).
type QianfanLLM struct {
	*BaseLLM
}

// NewQianfanLLM creates a new QianfanLLM.
func NewQianfanLLM(config *types.ClientConfig) *QianfanLLM {
	BuildStandardConfigSimple(config, DefaultQianfanAPIBase, DefaultQianfanModel)

	return &QianfanLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the provider name.
func (q *QianfanLLM) Name() string {
	return "qianfan"
}

// GetRequiredConfig returns provider-specific configuration requirements.
func (q *QianfanLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultQianfanAPIBase,
			PromptMessage: "Enter Qianfan API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter Qianfan API key",
		},
		"model": {
			DefaultValue:  DefaultQianfanModel,
			PromptMessage: "Enter ERNIE model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Qianfan OpenAI-compatible API.
func (q *QianfanLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return q.BaseLLM.MakeRequest(ctx, client, q, message, stream)
}
