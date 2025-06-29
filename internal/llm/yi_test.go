package llm

import (
	"strings"
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewYiLLM(t *testing.T) {
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
				apiBase: "https://api.lingyiwanwu.com/v1",
				model:   "yi-lightning",
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
			got := NewYiLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestYiLLM_Name(t *testing.T) {
	llm := NewYiLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "yi" {
		t.Errorf("Name() = %s, want %s", got, "yi")
	}
}

func TestYiLLM_GetRequiredConfig(t *testing.T) {
	llm := NewYiLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"api_key",
		"model",
		"max_tokens",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}

	if got["api_base"].DefaultValue != "https://api.lingyiwanwu.com/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "yi-lightning" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
	if got["max_tokens"].DefaultValue != "1024" {
		t.Errorf("Unexpected default value for max_tokens, got %s", got["max_tokens"].DefaultValue)
	}
}

func TestYiLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.lingyiwanwu.com/v1",
				CompletionPath: &defaultPath,
			},
			want: "https://api.lingyiwanwu.com/v1/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.lingyiwanwu.com/v1/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.lingyiwanwu.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewYiLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestYiLLM_BuildHeaders(t *testing.T) {
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
			llm := NewYiLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestYiLLM_GetUsage(t *testing.T) {
	llm := NewYiLLM(&types.ClientConfig{})
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

func TestYiLLM_ParseResponse(t *testing.T) {
	defaultAnswerPath := "choices.0.message.content"
	customAnswerPath := "custom.path"

	tests := []struct {
		name       string
		config     *types.ClientConfig
		response   []byte
		want       string
		wantErr    bool
		errMessage string
	}{
		{
			name: "standard response",
			config: &types.ClientConfig{
				AnswerPath: defaultAnswerPath,
			},
			response: []byte(`{
                "choices": [
                    {
                        "message": {
                            "content": "Hello, world!"
                        }
                    }
                ]
            }`),
			want:    "Hello, world!",
			wantErr: false,
		},
		{
			name: "response with code block",
			config: &types.ClientConfig{
				AnswerPath: defaultAnswerPath,
			},
			response: []byte(`{
                "choices": [
                    {
                        "message": {
                            "content": "code example"
                        }
                    }
                ]
            }`),
			want:    "code example",
			wantErr: false,
		},
		{
			name: "custom answer path",
			config: &types.ClientConfig{
				AnswerPath: customAnswerPath,
			},
			response: []byte(`{
                "custom": {
                    "path": "Custom response"
                }
            }`),
			want:    "Custom response",
			wantErr: false,
		},
		{
			name: "invalid response",
			config: &types.ClientConfig{
				AnswerPath: defaultAnswerPath,
			},
			response:   []byte(`{}`),
			want:       "",
			wantErr:    true,
			errMessage: "failed to parse response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewYiLLM(tt.config)
			got, err := llm.ParseResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMessage) {
				t.Errorf("ParseResponse() error = %v, expected to contain %v", err, tt.errMessage)
				return
			}
			if got != tt.want {
				t.Errorf("ParseResponse() = %s, want %s", got, tt.want)
			}
		})
	}
}
