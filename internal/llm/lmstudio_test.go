package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewLMStudioLLM(t *testing.T) {
	tests := []struct {
		name      string
		config    *types.ClientConfig
		wantBase  string
		wantModel string
	}{
		{
			name:      "default config",
			config:    &types.ClientConfig{},
			wantBase:  DefaultLMStudioAPIBase,
			wantModel: DefaultLMStudioModel,
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "http://127.0.0.1:1234/v1",
				Model:   "qwen3.8-8b",
			},
			wantBase:  "http://127.0.0.1:1234/v1",
			wantModel: "qwen3.8-8b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLMStudioLLM(tt.config)
			if got.Config.APIBase != tt.wantBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.wantBase)
			}
			if got.Config.Model != tt.wantModel {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.wantModel)
			}
		})
	}
}

func TestLMStudioLLM_Name(t *testing.T) {
	llm := NewLMStudioLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "lmstudio" {
		t.Errorf("Name() = %s, want lmstudio", got)
	}
}

func TestLMStudioLLM_GetRequiredConfig(t *testing.T) {
	llm := NewLMStudioLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	for _, key := range []string{"api_base", "model", "api_key", "max_tokens"} {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}
	if got["api_base"].DefaultValue != DefaultLMStudioAPIBase {
		t.Errorf("api_base default = %s, want %s", got["api_base"].DefaultValue, DefaultLMStudioAPIBase)
	}
}

func TestLMStudioLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name:   "default url",
			config: &types.ClientConfig{},
			want:   "http://localhost:1234/v1/chat/completions",
		},
		{
			name: "trailing slash",
			config: &types.ClientConfig{
				APIBase: "http://localhost:1234/v1/",
			},
			want: "http://localhost:1234/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewLMStudioLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLMStudioLLM_BuildHeaders(t *testing.T) {
	t.Run("no api key", func(t *testing.T) {
		llm := NewLMStudioLLM(&types.ClientConfig{})
		got := llm.BuildHeaders()

		if _, exists := got["Authorization"]; exists {
			t.Errorf("Authorization header should be omitted without an API key")
		}
		if got["Content-Type"] != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", got["Content-Type"])
		}
	})

	t.Run("with api key", func(t *testing.T) {
		llm := NewLMStudioLLM(&types.ClientConfig{APIKey: "test-key"})
		got := llm.BuildHeaders()

		if got["Authorization"] != "Bearer test-key" {
			t.Errorf("Authorization = %s, want Bearer test-key", got["Authorization"])
		}
	})
}
