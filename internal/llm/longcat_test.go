package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewLongCatLLM(t *testing.T) {
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
				apiBase: "https://api.longcat.chat/openai",
				model:   "LongCat-Flash-Chat",
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
			got := NewLongCatLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("NewLongCatLLM().Config.APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("NewLongCatLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestLongCatLLM_Name(t *testing.T) {
	llm := NewLongCatLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "longcat" {
		t.Errorf("LongCatLLM.Name() = %v, want %v", got, "longcat")
	}
}

func TestLongCatLLM_GetRequiredConfig(t *testing.T) {
	llm := NewLongCatLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://api.longcat.chat/openai" {
		t.Errorf("Unexpected default value for api_base: got %v, want %v",
			got["api_base"].DefaultValue, "https://api.longcat.chat/openai")
	}
	if got["model"].DefaultValue != "LongCat-Flash-Chat" {
		t.Errorf("Unexpected default value for model: got %v, want %v",
			got["model"].DefaultValue, "LongCat-Flash-Chat")
	}
}

func TestLongCatLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "default url",
			config: &types.ClientConfig{
				APIBase:        "https://api.longcat.chat/openai",
				CompletionPath: &defaultPath,
			},
			want: "https://api.longcat.chat/openai/chat/completions",
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
				APIBase:        "https://api.longcat.chat/openai/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.longcat.chat/openai/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewLongCatLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("LongCatLLM.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
