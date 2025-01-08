package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewClaudeLLM(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   struct {
			apiBase          string
			model            string
			completionPath   string
			answerPath       string
			anthropicVersion string
		}
	}{
		{
			name:   "default config",
			config: &types.ClientConfig{},
			want: struct {
				apiBase          string
				model            string
				completionPath   string
				answerPath       string
				anthropicVersion string
			}{
				apiBase:          "https://api.anthropic.com/v1",
				model:            "claude-3-sonnet",
				completionPath:   "messages",
				answerPath:       "content.0.text",
				anthropicVersion: "2024-01-01",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:          "https://custom.api.com",
				Model:            "custom-model",
				CompletionPath:   "custom/path",
				AnswerPath:       "custom.path",
				AnthropicVersion: "2024-02-01",
			},
			want: struct {
				apiBase          string
				model            string
				completionPath   string
				answerPath       string
				anthropicVersion string
			}{
				apiBase:          "https://custom.api.com",
				model:            "custom-model",
				completionPath:   "custom/path",
				answerPath:       "custom.path",
				anthropicVersion: "2024-02-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClaudeLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %v, want %v", got.Config.Model, tt.want.model)
			}
			if got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %v, want %v", got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %v, want %v", got.Config.AnswerPath, tt.want.answerPath)
			}
			if got.Config.AnthropicVersion != tt.want.anthropicVersion {
				t.Errorf("AnthropicVersion = %v, want %v", got.Config.AnthropicVersion, tt.want.anthropicVersion)
			}
		})
	}
}

func TestClaudeLLM_Name(t *testing.T) {
	llm := NewClaudeLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "Claude" {
		t.Errorf("Name() = %v, want %v", got, "Claude")
	}
}

func TestClaudeLLM_GetRequiredConfig(t *testing.T) {
	llm := NewClaudeLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"model",
		"api_key",
		"anthropic_version",
		"max_tokens",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %v", key)
		}
	}

	// Verify default values
	if got["api_base"].DefaultValue != "https://api.anthropic.com/v1" {
		t.Errorf("Unexpected default value for api_base")
	}
	if got["model"].DefaultValue != "claude-3-sonnet" {
		t.Errorf("Unexpected default value for model")
	}
	if got["anthropic_version"].DefaultValue != "2024-01-01" {
		t.Errorf("Unexpected default value for anthropic_version")
	}
}

func TestClaudeLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.anthropic.com/v1",
				CompletionPath: "messages",
			},
			want: "https://api.anthropic.com/v1/messages",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.anthropic.com/v1/",
				CompletionPath: "messages",
			},
			want: "https://api.anthropic.com/v1/messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewClaudeLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaudeLLM_BuildHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   map[string]string
	}{
		{
			name: "standard headers",
			config: &types.ClientConfig{
				APIKey:           "test-key",
				AnthropicVersion: "2024-01-01",
			},
			want: map[string]string{
				"Content-Type":      "application/json",
				"anthropic-version": "2024-01-01",
				"x-api-key":         "test-key",
			},
		},
		{
			name: "headers with extra headers",
			config: &types.ClientConfig{
				APIKey:           "test-key",
				AnthropicVersion: "2024-01-01",
				ExtraHeaders: map[string]string{
					"X-Custom": "custom-value",
				},
			},
			want: map[string]string{
				"Content-Type":      "application/json",
				"anthropic-version": "2024-01-01",
				"x-api-key":         "test-key",
				"X-Custom":          "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewClaudeLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%v] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
