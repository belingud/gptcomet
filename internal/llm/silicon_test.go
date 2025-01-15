package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewSiliconLLM(t *testing.T) {
	customPath := "custom/path"
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
				apiBase:        "https://api.siliconflow.cn/v1",
				model:          "Qwen/Qwen2.5-7B-Instruct",
				completionPath: "chat/completions",
				answerPath:     "choices.0.message.content",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:        "https://custom.api.com",
				Model:          "custom-model",
				CompletionPath: &customPath,
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
			got := NewSiliconLLM(tt.config)
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

func TestSiliconLLM_Name(t *testing.T) {
	llm := NewSiliconLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "silicon" {
		t.Errorf("Name() = %s, want %s", got, "silicon")
	}
}

func TestSiliconLLM_GetRequiredConfig(t *testing.T) {
	llm := NewSiliconLLM(&types.ClientConfig{})
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

	// check default values
	if got["api_base"].DefaultValue != "https://api.siliconflow.cn/v1" {
		t.Errorf("Unexpected default value for api_base, wanted https://api.siliconflow.cn/v1, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "Qwen/Qwen2.5-7B-Instruct" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestSiliconLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.siliconflow.cn/v1",
				CompletionPath: &defaultPath,
			},
			want: "https://api.siliconflow.cn/v1/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.siliconflow.cn/v1/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.siliconflow.cn/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewSiliconLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSiliconLLM_BuildHeaders(t *testing.T) {
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
			llm := NewSiliconLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestSiliconLLM_GetUsage(t *testing.T) {
	llm := NewSiliconLLM(&types.ClientConfig{})
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
			want:    "Token usage> prompt: 10, completion: 20, total: 30",
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
