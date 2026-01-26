package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultChatGLMAPIBase = "https://open.bigmodel.cn/api/paas/v4"
	DefaultChatGLMModel   = "glm-4-flash"
)

// ChatGLMLLM implements the LLM interface for ChatGLM
type ChatGLMLLM struct {
	*BaseLLM
}

// NewChatGLMLLM creates a new ChatGLMLLM
func NewChatGLMLLM(config *types.ClientConfig) *ChatGLMLLM {
	BuildStandardConfigSimple(config, DefaultChatGLMAPIBase, DefaultChatGLMModel)

	return &ChatGLMLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (c *ChatGLMLLM) Name() string {
	return "ChatGLM"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (c *ChatGLMLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultChatGLMAPIBase,
			PromptMessage: "Enter ChatGLM API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultChatGLMModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the ChatGLM API
func (c *ChatGLMLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return c.BaseLLM.MakeRequest(ctx, client, c, message, stream)
}
