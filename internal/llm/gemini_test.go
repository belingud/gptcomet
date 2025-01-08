package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewGeminiLLM(t *testing.T) {
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
				apiBase:        "https://generativelanguage.googleapis.com/v1beta/models",
				model:          "gemini-1.5-flash",
				completionPath: "generateContent",
				answerPath:     "candidates.0.content.parts.0.text",
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
			got := NewGeminiLLM(tt.config)
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
		})
	}
}

func TestGeminiLLM_Name(t *testing.T) {
	llm := NewGeminiLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "gemini" {
		t.Errorf("Name() = %v, want %v", got, "gemini")
	}
}

func TestGeminiLLM_GetRequiredConfig(t *testing.T) {
	llm := NewGeminiLLM(&types.ClientConfig{})
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
	if got["api_base"].DefaultValue != "https://generativelanguage.googleapis.com/v1beta/models" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["model"].DefaultValue != "gemini-1.5-flash" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestGeminiLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase: "https://generativelanguage.googleapis.com/v1beta/models",
				Model:   "gemini-1.5-flash",
				APIKey:  "test-key",
			},
			want: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=test-key",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase: "https://generativelanguage.googleapis.com/v1beta/models/",
				Model:   "gemini-1.5-flash",
				APIKey:  "test-key",
			},
			want: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewGeminiLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeminiLLM_FormatMessages(t *testing.T) {
	llm := NewGeminiLLM(&types.ClientConfig{
		MaxTokens:        1024,
		Temperature:      0.7,
		TopP:             0.9,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.5,
	})

	message := "test message"
	got, err := llm.FormatMessages(message, nil)
	if err != nil {
		t.Errorf("FormatMessages() error = %v", err)
		return
	}

	payload, ok := got.(map[string]interface{})
	if !ok {
		t.Errorf("FormatMessages() returned wrong type")
		return
	}

	// 验证消息格式
	contents, ok := payload["contents"].([]map[string]interface{})
	if !ok || len(contents) != 1 {
		t.Errorf("FormatMessages() wrong contents format")
		return
	}

	// 验证生成配置
	genConfig, ok := payload["generationConfig"].(map[string]interface{})
	if !ok {
		t.Errorf("FormatMessages() missing generationConfig")
		return
	}

	if genConfig["maxOutputTokens"] != 1024 {
		t.Errorf("maxOutputTokens = %v, want 1024", genConfig["maxOutputTokens"])
	}
	if genConfig["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7", genConfig["temperature"])
	}
	if genConfig["topP"] != 0.9 {
		t.Errorf("topP = %v, want 0.9", genConfig["topP"])
	}
}

func TestGeminiLLM_GetUsage(t *testing.T) {
	llm := NewGeminiLLM(&types.ClientConfig{})
	testData := []byte(`{
		"usageMetadata": {
			"promptTokenCount": 10,
			"candidatesTokenCount": 20,
			"totalTokenCount": 30
		}
	}`)

	usage, err := llm.GetUsage(testData)
	if err != nil {
		t.Errorf("GetUsage() error = %v", err)
		return
	}

	expected := "Token usage> promptTokenCount: 10, candidatesTokenCount: 20, totalTokenCount: 30"
	if usage != expected {
		t.Errorf("GetUsage() = %v, want %v", usage, expected)
	}
}
