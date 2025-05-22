package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const DEFAULT_MODELSCOPE_MODEL = "deepseek-ai/DeepSeek-V3-0324"

// ModelScopeLLM implements the LLM interface for ModelScope
// It extends OpenAILLM since ModelScope API is compatible with OpenAI
type ModelScopeLLM struct {
	*OpenAILLM
}

// NewModelScopeLLM creates a new ModelScopeLLM
func NewModelScopeLLM(config *types.ClientConfig) *ModelScopeLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api-inference.modelscope.cn/v1"
	}
	if config.Model == "" {
		config.Model = DEFAULT_MODELSCOPE_MODEL
	}
	return &ModelScopeLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (m *ModelScopeLLM) Name() string {
	return "modelscope"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (m *ModelScopeLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	requirements := m.OpenAILLM.GetRequiredConfig()
	// update API base
	requirements["api_base"] = config.ConfigRequirement{
		DefaultValue:  "https://api-inference.modelscope.cn/v1",
		PromptMessage: "Enter ModelScope API base URL",
	}
	// update model
	requirements["model"] = config.ConfigRequirement{
		DefaultValue:  DEFAULT_MODELSCOPE_MODEL,
		PromptMessage: "Enter model name",
	}
	return requirements
}

// MakeRequest makes a request to the API
func (m *ModelScopeLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return m.OpenAILLM.MakeRequest(ctx, client, message, stream)
}
