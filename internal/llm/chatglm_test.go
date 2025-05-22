package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewChatGLMLLM(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   *ChatGLMLLM
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			want: &ChatGLMLLM{
				BaseLLM: &BaseLLM{
					Config: &types.ClientConfig{
						APIBase:        "https://open.bigmodel.cn/api/paas/v4",
						Model:          "glm-4-flash",
						CompletionPath: &defaultPath,
						AnswerPath:     "choices.0.message.content",
					},
				},
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://custom.api.com",
				Model:   "custom-model",
			},
			want: &ChatGLMLLM{
				BaseLLM: &BaseLLM{
					Config: &types.ClientConfig{
						APIBase:        "https://custom.api.com",
						Model:          "custom-model",
						CompletionPath: &defaultPath,
						AnswerPath:     "choices.0.message.content",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChatGLMLLM(tt.config)
			if got.Config.APIBase != tt.want.Config.APIBase {
				t.Errorf("NewChatGLMLLM().Config.APIBase = %v, want %v", got.Config.APIBase, tt.want.Config.APIBase)
			}
			if got.Config.Model != tt.want.Config.Model {
				t.Errorf("NewChatGLMLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.Config.Model)
			}
			if *got.Config.CompletionPath != *tt.want.Config.CompletionPath {
				t.Errorf("NewChatGLMLLM().Config.CompletionPath = %v, want %v", *got.Config.CompletionPath, *tt.want.Config.CompletionPath)
			}
		})
	}
}

func TestChatGLMLLM_Name(t *testing.T) {
	llm := NewChatGLMLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "ChatGLM" {
		t.Errorf("ChatGLMLLM.Name() = %v, want %v", got, "ChatGLM")
	}
}

func TestChatGLMLLM_GetRequiredConfig(t *testing.T) {
	llm := NewChatGLMLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://open.bigmodel.cn/api/paas/v4" {
		t.Errorf("Unexpected default value for api_base")
	}
	if got["model"].DefaultValue != "glm-4-flash" {
		t.Errorf("Unexpected default value for model")
	}
}

func TestChatGLMLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "default url",
			config: &types.ClientConfig{
				APIBase:        "https://open.bigmodel.cn/api/paas/v4",
				CompletionPath: &defaultPath,
			},
			want: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
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
				APIBase:        "https://open.bigmodel.cn/api/paas/v4/",
				CompletionPath: &defaultPath,
			},
			want: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewChatGLMLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("ChatGLMLLM.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
