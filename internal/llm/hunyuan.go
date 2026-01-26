package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultHunyuanAPIBase = "https://api.hunyuan.cloud.tencent.com/v1"
	DefaultHunyuanModel   = "hunyuan-lite"
)

// HunyuanLLM implements the LLM interface for Tencent Hunyuan
// It extends OpenAILLM since Hunyuan API is compatible with OpenAI
type HunyuanLLM struct {
	*OpenAILLM
}

// NewHunyuanLLM creates a new HunyuanLLM
func NewHunyuanLLM(config *types.ClientConfig) *HunyuanLLM {
	BuildStandardConfigSimple(config, DefaultHunyuanAPIBase, DefaultHunyuanModel)

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
		DefaultValue:  DefaultHunyuanAPIBase,
		PromptMessage: "Enter Hunyuan API base URL",
	}
	// update model
	requirements["model"] = config.ConfigRequirement{
		DefaultValue:  DefaultHunyuanModel,
		PromptMessage: "Enter model name",
	}
	return requirements
}

// MakeRequest makes a request to the API
func (h *HunyuanLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return h.OpenAILLM.MakeRequest(ctx, client, message, stream)
}
