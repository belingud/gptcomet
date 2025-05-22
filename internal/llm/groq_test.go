package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGroqLLM(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.ClientConfig
		expected *GroqLLM
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			expected: &GroqLLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://api.groq.com/openai/v1",
					Model:   "llama-3.3-70b-versatile",
				}),
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://custom-api.groq.com/v1",
				Model:   "custom-model",
			},
			expected: &GroqLLM{
				OpenAILLM: NewOpenAILLM(&types.ClientConfig{
					APIBase: "https://custom-api.groq.com/v1",
					Model:   "custom-model",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewGroqLLM(tt.config)
			assert.Equal(t, "groq", llm.Name())
			assert.Equal(t, tt.expected.Config.APIBase, llm.Config.APIBase)
			assert.Equal(t, tt.expected.Config.Model, llm.Config.Model)
		})
	}
}

func TestGroqLLMGetRequiredConfig(t *testing.T) {
	llm := NewGroqLLM(&types.ClientConfig{})
	config := llm.GetRequiredConfig()

	expected := map[string]string{
		"api_base":   "https://api.groq.com/openai/v1",
		"model":      "llama-3.3-70b-versatile",
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
