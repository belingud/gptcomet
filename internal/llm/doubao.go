package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	// DefaultDoubaoAPIBase is the OpenAI-compatible Volcano Engine Ark endpoint.
	DefaultDoubaoAPIBase = "https://ark.cn-beijing.volces.com/api/v3"
	// DefaultDoubaoModel is the default Doubao model used with Volcano Ark.
	DefaultDoubaoModel = "doubao-seed-2-1-pro-260915"
)

// DoubaoLLM implements the LLM interface for ByteDance Doubao (Volcano Engine Ark).
type DoubaoLLM struct {
	*BaseLLM
}

// NewDoubaoLLM creates a new DoubaoLLM.
func NewDoubaoLLM(config *types.ClientConfig) *DoubaoLLM {
	BuildStandardConfigSimple(config, DefaultDoubaoAPIBase, DefaultDoubaoModel)

	return &DoubaoLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the provider name.
func (d *DoubaoLLM) Name() string {
	return "doubao"
}

// GetRequiredConfig returns provider-specific configuration requirements.
func (d *DoubaoLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultDoubaoAPIBase,
			PromptMessage: "Enter Doubao API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter Volcano Ark API key",
		},
		"model": {
			DefaultValue:  DefaultDoubaoModel,
			PromptMessage: "Enter Doubao model ID or endpoint ID",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Doubao OpenAI-compatible API.
func (d *DoubaoLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return d.BaseLLM.MakeRequest(ctx, client, d, message, stream)
}
