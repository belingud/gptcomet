package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAI21LLM(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.ClientConfig
		expected *AI21LLM
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			expected: &AI21LLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://api.ai21.com/studio/v1",
					Model:   "jamba-1.5-large",
				}),
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://custom-api.ai21.com/v1",
				Model:   "custom-model",
			},
			expected: &AI21LLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://custom-api.ai21.com/v1",
					Model:   "custom-model",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewAI21LLM(tt.config)
			assert.Equal(t, "ai21", llm.Name())
			assert.Equal(t, tt.expected.Config.APIBase, llm.Config.APIBase)
			assert.Equal(t, tt.expected.Config.Model, llm.Config.Model)
		})
	}
}

func TestAI21LMGetRequiredConfig(t *testing.T) {
	llm := NewAI21LLM(&types.ClientConfig{})
	config := llm.GetRequiredConfig()

	expected := map[string]string{
		"api_base":   "https://api.ai21.com/studio/v1",
		"model":      "jamba-1.5-large",
		"max_tokens": "1024",
	}

	assert.Equal(t, expected["api_base"], config["api_base"].DefaultValue)
	assert.Equal(t, expected["model"], config["model"].DefaultValue)
	assert.Equal(t, expected["max_tokens"], config["max_tokens"].DefaultValue)

	requiredKeys := []string{"api_base", "api_key", "model", "max_tokens"}
	for _, key := range requiredKeys {
		_, exists := config[key]
		assert.True(t, exists, "Missing required config key: "+key)
	}
}
