package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewSambanovaLLM(t *testing.T) {
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
				apiBase: "https://api.sambanova.ai/v1",
				model:   "Meta-Llama-3.3-70B-Instruct",
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
			got := NewSambanovaLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestSambanovaLLM_Name(t *testing.T) {
	llm := NewSambanovaLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "sambanova" {
		t.Errorf("Name() = %s, want %s", got, "sambanova")
	}
}

func TestSambanovaLLM_GetRequiredConfig(t *testing.T) {
	llm := NewSambanovaLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://api.sambanova.ai/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "Meta-Llama-3.3-70B-Instruct" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestSambanovaLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.sambanova.ai/v1",
				CompletionPath: &defaultPath,
			},
			want: "https://api.sambanova.ai/v1/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.sambanova.ai/v1/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.sambanova.ai/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewSambanovaLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSambanovaLLM_BuildHeaders(t *testing.T) {
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
			llm := NewSambanovaLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders() %s = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}
