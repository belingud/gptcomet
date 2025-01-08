package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewXAILLM(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   struct {
			apiBase        string
			model          string
			completionPath string
			answerPath     string
		}
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			want: struct {
				apiBase        string
				model          string
				completionPath string
				answerPath     string
			}{
				apiBase:        "https://api.x.ai/v1",
				model:          "grok-beta",
				completionPath: "chat/completions",
				answerPath:     "choices.0.message.content",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:        "https://custom.api.com",
				Model:          "custom-model",
				CompletionPath: "custom/path",
				AnswerPath:     "custom.path",
			},
			want: struct {
				apiBase        string
				model          string
				completionPath string
				answerPath     string
			}{
				apiBase:        "https://custom.api.com",
				model:          "custom-model",
				completionPath: "custom/path",
				answerPath:     "custom.path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewXAILLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
			if got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %s, want %s", got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %s, want %s", got.Config.AnswerPath, tt.want.answerPath)
			}
		})
	}
}

func TestXAILLM_Name(t *testing.T) {
	llm := NewXAILLM(&types.ClientConfig{})
	if got := llm.Name(); got != "xai" {
		t.Errorf("Name() = %s, want %s", got, "xai")
	}
}

func TestXAILLM_GetRequiredConfig(t *testing.T) {
	llm := NewXAILLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"model",
		"api_key",
		"max_tokens",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}

	// 验证默认值
	if got["api_base"].DefaultValue != "https://api.x.ai/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "grok-beta" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestXAILLM_BuildHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   map[string]string
	}{
		{
			name: "standard headers",
			config: &types.ClientConfig{
				APIKey: "test-key",
			},
			want: map[string]string{
				"Authorization": "Bearer test-key",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "headers with extra headers",
			config: &types.ClientConfig{
				APIKey: "test-key",
				ExtraHeaders: map[string]string{
					"X-Custom": "custom-value",
				},
			},
			want: map[string]string{
				"Authorization": "Bearer test-key",
				"Content-Type":  "application/json",
				"X-Custom":      "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewXAILLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}
