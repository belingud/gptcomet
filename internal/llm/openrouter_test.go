package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewOpenRouterLLM(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.ClientConfig
		expected *OpenRouterLLM
	}{
		{
			name:   "empty config should set defaults",
			config: &types.ClientConfig{},
			expected: &OpenRouterLLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://openrouter.ai/api/v1",
					Model:   "meta-llama/llama-3.1-70b-instruct:free",
				}),
			},
		},
		{
			name: "custom config should be respected",
			config: &types.ClientConfig{
				APIBase: "https://custom.openrouter.ai/v1",
				Model:   "anthropic/claude-3",
				APIKey:  "test-key",
			},
			expected: &OpenRouterLLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://custom.openrouter.ai/v1",
					Model:   "anthropic/claude-3",
					APIKey:  "test-key",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewOpenRouterLLM(tt.config)
			assert.Equal(t, tt.expected.Config.APIBase, result.Config.APIBase)
			assert.Equal(t, tt.expected.Config.Model, result.Config.Model)
			assert.Equal(t, tt.expected.Config.APIKey, result.Config.APIKey)
		})
	}
}

func TestOpenRouterLLM_Name(t *testing.T) {
	llm := NewOpenRouterLLM(&types.ClientConfig{})
	assert.Equal(t, "openrouter", llm.Name())
}

func TestOpenRouterLLM_GetRequiredConfig(t *testing.T) {
	llm := NewOpenRouterLLM(&types.ClientConfig{})
	config := llm.GetRequiredConfig()

	assert.Equal(t, "https://openrouter.ai/api/v1", config["api_base"].DefaultValue)
	assert.Equal(t, "meta-llama/llama-3.1-70b-instruct:free", config["model"].DefaultValue)
	assert.Equal(t, "1024", config["max_tokens"].DefaultValue)
	assert.Empty(t, config["api_key"].DefaultValue)
}

func TestOpenRouterLLM_BuildHeaders(t *testing.T) {
	llm := NewOpenRouterLLM(&types.ClientConfig{
		APIKey: "test-key",
	})

	headers := llm.BuildHeaders()
	assert.Equal(t, "Bearer test-key", headers["Authorization"])
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "https://github.com/belingud/gptcomet", headers["HTTP-Referer"])
	assert.Equal(t, "GPTComet", headers["X-Title"])
}
