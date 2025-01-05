package llm

import (
	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// ChatGLMLLM implements the LLM interface for ChatGLM
type ChatGLMLLM struct {
	*OpenAILLM
}

// NewChatGLMLLM creates a new ChatGLMLLM
func NewChatGLMLLM(config *types.ClientConfig) *ChatGLMLLM {
	if config.APIBase == "" {
		config.APIBase = "https://open.bigmodel.cn/api/paas/v4"
	}
	if config.Model == "" {
		config.Model = "glm-4-flash"
	}

	return &ChatGLMLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (c *ChatGLMLLM) Name() string {
	return "ChatGLM"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (c *ChatGLMLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://open.bigmodel.cn/api/paas/v4",
			PromptMessage: "Enter ChatGLM API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  "glm-4-flash",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}
