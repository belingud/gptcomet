package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewCohereLLM(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   struct {
			apiBase string
			model   string
		}
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			want: struct {
				apiBase string
				model   string
			}{
				apiBase: "https://api.cohere.com/v2",
				model:   "command-r-plus",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://custom.api.com",
				Model:   "custom-model",
			},
			want: struct {
				apiBase string
				model   string
			}{
				apiBase: "https://custom.api.com",
				model:   "custom-model",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCohereLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("NewCohereLLM().Config.APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("NewCohereLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestCohereLLM_Name(t *testing.T) {
	llm := NewCohereLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "cohere" {
		t.Errorf("CohereLLM.Name() = %v, want %v", got, "cohere")
	}
}

func TestCohereLLM_GetRequiredConfig(t *testing.T) {
	llm := NewCohereLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"api_key",
		"model",
		"max_tokens",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %v", key)
		}
	}

	// 验证默认值
	if got["api_base"].DefaultValue != "https://api.cohere.com/v2" {
		t.Errorf("GetRequiredConfig() api_base default value = %v, want %v", got["api_base"].DefaultValue, "https://api.cohere.com/v2")
	}
	if got["model"].DefaultValue != "command-r-plus" {
		t.Errorf("GetRequiredConfig() model default value = %v, want %v", got["model"].DefaultValue, "command-r-plus")
	}
}
