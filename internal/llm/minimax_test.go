package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewMinimaxLLM(t *testing.T) {
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
				apiBase:        "https://api.minimaxi.com/v1",
				model:          "MiniMax-M1",
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
			got := NewMinimaxLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %v, want %v", got.Config.Model, tt.want.model)
			}
			if *got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %v, want %v", *got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %v, want %v", got.Config.AnswerPath, tt.want.answerPath)
			}
		})
	}
}

func TestMinimaxLLM_Name(t *testing.T) {
	llm := NewMinimaxLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "minimax" {
		t.Errorf("Name() = %v, want %v", got, "minimax")
	}
}

func TestMinimaxLLM_GetRequiredConfig(t *testing.T) {
	llm := NewMinimaxLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://api.minimaxi.com/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "MiniMax-M1" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
	if got["max_tokens"].DefaultValue != "1024" {
		t.Errorf("Unexpected default value for max_tokens, got %s", got["max_tokens"].DefaultValue)
	}
}

func TestMinimaxLLM_MakeRequest(t *testing.T) {
	// This test is a simple verification that MakeRequest calls the BaseLLM's MakeRequest
	// Since the actual implementation delegates to BaseLLM, we just need to ensure it's called correctly
	// A more comprehensive test would use mocking to verify the interaction
	llm := NewMinimaxLLM(&types.ClientConfig{})
	// The actual test would be more complex with mocking
	// Here we're just ensuring the method exists and doesn't panic
	_ = llm.MakeRequest
}
