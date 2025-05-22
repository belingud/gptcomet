package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewOllamaLLM(t *testing.T) {
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
				apiBase:        "http://localhost:11434/api",
				model:          "llama2",
				completionPath: "generate",
				answerPath:     "response",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:        "http://custom.api.com",
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
				apiBase:        "http://custom.api.com",
				model:          "custom-model",
				completionPath: "custom/path",
				answerPath:     "custom.path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOllamaLLM(tt.config)
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

func TestOllamaLLM_Name(t *testing.T) {
	llm := NewOllamaLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "ollama" {
		t.Errorf("Name() = %s, want %s", got, "ollama")
	}
}

func TestOllamaLLM_GetRequiredConfig(t *testing.T) {
	llm := NewOllamaLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"model",
		"max_tokens",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}

	if got["api_base"].DefaultValue != "http://localhost:11434/api" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "llama2" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestOllamaLLM_FormatMessages(t *testing.T) {
	llm := NewOllamaLLM(&types.ClientConfig{
		Model:             "llama2",
		MaxTokens:         1024,
		Temperature:       0.7,
		TopP:              0.9,
		TopK:              40,
		RepetitionPenalty: 1.1,
		FrequencyPenalty:  0.5,
		PresencePenalty:   0.5,
		Seed:              42,
		NumGPU:            1,
		MainGPU:           0,
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

	if payload["model"] != "llama2" {
		t.Errorf("model = %v, want llama2", payload["model"])
	}
	if payload["prompt"] != message {
		t.Errorf("prompt = %v, want %v", payload["prompt"], message)
	}

	options, ok := payload["options"].(map[string]interface{})
	if !ok {
		t.Errorf("options wrong type")
		return
	}

	expectedOptions := map[string]interface{}{
		"num_predict":        1024,
		"temperature":        0.7,
		"top_p":              0.9,
		"top_k":              40,
		"repetition_penalty": 1.1,
		"frequency_penalty":  0.5,
		"presence_penalty":   0.5,
		"seed":               42,
		"num_gpu":            1,
	}

	for k, expected := range expectedOptions {
		got := options[k]
		if got == nil {
			t.Errorf("option %s is missing", k)
			continue
		}
		switch expected.(type) {
		case int:
			gotInt, ok := got.(int)
			if !ok {
				t.Errorf("option %s = %v (%T), want int type", k, got, got)
				continue
			}
			if gotInt != expected.(int) {
				t.Errorf("option %s = %v, want %v", k, gotInt, expected)
			}
		default:
			if got != expected {
				t.Errorf("option %s = %v, want %v", k, got, expected)
			}
		}
	}
}

func TestOllamaLLM_BuildHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   map[string]string
	}{
		{
			name:   "default headers",
			config: &types.ClientConfig{},
			want: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "with extra headers",
			config: &types.ClientConfig{
				ExtraHeaders: map[string]string{
					"X-Custom": "custom-value",
				},
			},
			want: map[string]string{
				"Content-Type": "application/json",
				"X-Custom":     "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewOllamaLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("BuildHeaders()[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}
