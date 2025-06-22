package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewMistralLLM(t *testing.T) {
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
				apiBase: "https://api.mistral.ai/v1",
				model:   "mistral-large-latest",
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
			got := NewMistralLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestMistralLLM_Name(t *testing.T) {
	llm := NewMistralLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "mistral" {
		t.Errorf("Name() = %s, want %s", got, "mistral")
	}
}

func TestMistralLLM_GetRequiredConfig(t *testing.T) {
	llm := NewMistralLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://api.mistral.ai/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "mistral-large-latest" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestMistralLLM_BuildURL(t *testing.T) {
	completionPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.mistral.ai/v1",
				CompletionPath: &completionPath,
			},
			want: "https://api.mistral.ai/v1/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.mistral.ai/v1/",
				CompletionPath: &completionPath,
			},
			want: "https://api.mistral.ai/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewMistralLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestMistralLLM_BuildHeaders(t *testing.T) {
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
			llm := NewMistralLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestMistralLLM_FormatMessages(t *testing.T) {
	llm := NewMistralLLM(&types.ClientConfig{
		Model:            "mistral-large-latest",
		MaxTokens:        1024,
		Temperature:      0.7,
		TopP:             0.9,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.5,
	})

	message := "test message"

	got, err := llm.FormatMessages(message)
	if err != nil {
		t.Errorf("FormatMessages() error = %v", err)
		return
	}

	payload, ok := got.(map[string]interface{})
	if !ok {
		t.Errorf("FormatMessages() returned wrong type")
		return
	}

	// Verify required parameters
	if payload["model"] != "mistral-large-latest" {
		t.Errorf("model = %v, want mistral-large-latest", payload["model"])
	}
	if payload["max_tokens"] != 1024 {
		t.Errorf("max_tokens = %v, want 1024", payload["max_tokens"])
	}

	// Verify that max_completion_tokens is not present
	if _, exists := payload["max_completion_tokens"]; exists {
		t.Errorf("max_completion_tokens should not be present in payload")
	}

	// Verify optional parameters
	if payload["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7", payload["temperature"])
	}
	if payload["top_p"] != 0.9 {
		t.Errorf("top_p = %v, want 0.9", payload["top_p"])
	}
	if payload["frequency_penalty"] != 0.5 {
		t.Errorf("frequency_penalty = %v, want 0.5", payload["frequency_penalty"])
	}
	if payload["presence_penalty"] != 0.5 {
		t.Errorf("presence_penalty = %v, want 0.5", payload["presence_penalty"])
	}

	// Verify message format
	messages, ok := payload["messages"].([]types.Message)
	if !ok {
		t.Errorf("messages wrong type")
		return
	}
	if len(messages) != 1 {
		t.Errorf("got %d messages, want 1", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != message {
		t.Errorf("message content or role does not match expected values")
	}
}
