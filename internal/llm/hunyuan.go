package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const DEFAULT_HUNYUAN_MODEL = "hunyuan-lite"

// HunyuanLLM implements the LLM interface for Tencent Hunyuan
// It extends OpenAILLM since Hunyuan API is compatible with OpenAI
type HunyuanLLM struct {
	*OpenAILLM
}

// NewHunyuanLLM creates a new HunyuanLLM
func NewHunyuanLLM(config *types.ClientConfig) *HunyuanLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.hunyuan.cloud.tencent.com/v1"
	}
	if config.Model == "" {
		config.Model = DEFAULT_HUNYUAN_MODEL
	}
	return &HunyuanLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (h *HunyuanLLM) Name() string {
	return "hunyuan"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (h *HunyuanLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	requirements := h.OpenAILLM.GetRequiredConfig()
	// update API base
	requirements["api_base"] = config.ConfigRequirement{
		DefaultValue:  "https://api.hunyuan.cloud.tencent.com/v1",
		PromptMessage: "Enter Hunyuan API base URL",
	}
	// update model
	requirements["model"] = config.ConfigRequirement{
		DefaultValue:  DEFAULT_HUNYUAN_MODEL,
		PromptMessage: "Enter model name",
	}
	return requirements
}

// MakeRequest makes a request to the API
func (h *HunyuanLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return h.OpenAILLM.MakeRequest(ctx, client, message, stream)
}
