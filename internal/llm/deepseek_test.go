package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewDeepSeekLLM(t *testing.T) {
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
				apiBase: "https://api.deepseek.com/v1",
				model:   "deepseek-chat",
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
			got := NewDeepSeekLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("NewDeepSeekLLM().Config.APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("NewDeepSeekLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestDeepSeekLLM_Name(t *testing.T) {
	llm := NewDeepSeekLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "deepseek" {
		t.Errorf("DeepSeekLLM.Name() = %v, want %v", got, "deepseek")
	}
}

func TestDeepSeekLLM_GetRequiredConfig(t *testing.T) {
	llm := NewDeepSeekLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	// 检查必需的配置键
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
	if got["api_base"].DefaultValue != "https://api.deepseek.com/v1" {
		t.Errorf("Unexpected default value for api_base: got %v, want %v",
			got["api_base"].DefaultValue, "https://api.deepseek.com/v1")
	}
	if got["model"].DefaultValue != "deepseek-chat" {
		t.Errorf("Unexpected default value for model: got %v, want %v",
			got["model"].DefaultValue, "deepseek-chat")
	}
}

func TestDeepSeekLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "default url",
			config: &types.ClientConfig{
				APIBase:        "https://api.deepseek.com/v1",
				CompletionPath: &defaultPath,
			},
			want: "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name: "custom url",
			config: &types.ClientConfig{
				APIBase:        "https://custom.api.com",
				CompletionPath: &defaultPath,
			},
			want: "https://custom.api.com/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.deepseek.com/v1/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.deepseek.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewDeepSeekLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("DeepSeekLLM.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
