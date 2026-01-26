package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

var hyCustomPath = "custom/path"

func TestNewHunyuanLLM(t *testing.T) {
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
				apiBase:        "https://api.hunyuan.cloud.tencent.com/v1",
				model:          DefaultHunyuanModel,
				completionPath: "chat/completions",
				answerPath:     "choices.0.message.content",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:        "https://custom.api.com",
				Model:          "custom-model",
				CompletionPath: &hyCustomPath,
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
			got := NewHunyuanLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
			if *got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %s, want %s", *got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %s, want %s", got.Config.AnswerPath, tt.want.answerPath)
			}
		})
	}
}

func TestHunyuanLLM_Name(t *testing.T) {
	llm := NewHunyuanLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "hunyuan" {
		t.Errorf("Name() = %s, want %s", got, "hunyuan")
	}
}

func TestHunyuanLLM_GetRequiredConfig(t *testing.T) {
	llm := NewHunyuanLLM(&types.ClientConfig{})
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

	// Verify default values
	if got["api_base"].DefaultValue != "https://api.hunyuan.cloud.tencent.com/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != DefaultHunyuanModel {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}
