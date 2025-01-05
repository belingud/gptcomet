package llm

import (
	"testing"

	"github.com/belingud/go-gptcomet/pkg/types"
)

func TestNewTongyiLLM(t *testing.T) {
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
				apiBase:        "https://dashscope.aliyuncs.com/compatible-mode/v1",
				model:          "qwen-turbo",
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
			got := NewTongyiLLM(tt.config)
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

func TestTongyiLLM_Name(t *testing.T) {
	llm := NewTongyiLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "tongyi" {
		t.Errorf("Name() = %s, want %s", got, "tongyi")
	}
}

func TestTongyiLLM_GetRequiredConfig(t *testing.T) {
	llm := NewTongyiLLM(&types.ClientConfig{})
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
	if got["api_base"].DefaultValue != "https://dashscope.aliyuncs.com/compatible-mode/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "qwen-turbo" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestTongyiLLM_BuildHeaders(t *testing.T) {
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
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-key",
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
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-key",
				"X-Custom":      "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewTongyiLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestTongyiLLM_GetUsage(t *testing.T) {
	llm := NewTongyiLLM(&types.ClientConfig{})
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "standard usage",
			data: []byte(`{
                "usage": {
                    "prompt_tokens": 10,
                    "completion_tokens": 20,
                    "total_tokens": 30
                }
            }`),
			want:    "Token usage> input: 10, output: 20, total: 30",
			wantErr: false,
		},
		{
			name:    "no usage info",
			data:    []byte(`{}`),
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := llm.GetUsage(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUsage() = %s, want %s", got, tt.want)
			}
		})
	}
}
