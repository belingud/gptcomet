package llm

import (
	"testing"

	"github.com/belingud/go-gptcomet/pkg/types"
)

func TestNewVertexLLM(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   struct {
			apiBase        string
			model          string
			projectID      string
			location       string
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
				projectID      string
				location       string
				completionPath string
				answerPath     string
			}{
				apiBase:        "https://us-central1-aiplatform.googleapis.com/v1",
				model:          "gemini-1.5-flash",
				projectID:      "default-project",
				location:       "us-central1",
				completionPath: "projects/default-project/locations/us-central1/publishers/google/models/gemini-1.5-flash:generateContent",
				answerPath:     "candidates.0.content.parts.0.text",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:        "https://custom.api.com",
				Model:          "custom-model",
				ProjectID:      "test-project",
				Location:       "europe-west1",
				CompletionPath: "custom/path",
				AnswerPath:     "custom.path",
			},
			want: struct {
				apiBase        string
				model          string
				projectID      string
				location       string
				completionPath string
				answerPath     string
			}{
				apiBase:        "https://custom.api.com",
				model:          "custom-model",
				projectID:      "test-project",
				location:       "europe-west1",
				completionPath: "custom/path",
				answerPath:     "custom.path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewVertexLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.want.model)
			}
			if got.Config.ProjectID != tt.want.projectID {
				t.Errorf("ProjectID = %s, want %s", got.Config.ProjectID, tt.want.projectID)
			}
			if got.Config.Location != tt.want.location {
				t.Errorf("Location = %s, want %s", got.Config.Location, tt.want.location)
			}
			if got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %s, want %s", got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %s, want %s", got.Config.AnswerPath, tt.want.answerPath)
			}
		})
	}
}

func TestVertexLLM_Name(t *testing.T) {
	llm := NewVertexLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "vertex" {
		t.Errorf("Name() = %s, want %s", got, "vertex")
	}
}

func TestVertexLLM_GetRequiredConfig(t *testing.T) {
	llm := NewVertexLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"project_id",
		"location",
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
	if got["api_base"].DefaultValue != "https://us-central1-aiplatform.googleapis.com/v1" {
		t.Errorf("Unexpected default value for api_base, got %s", got["api_base"].DefaultValue)
	}
	if got["location"].DefaultValue != "us-central1" {
		t.Errorf("Unexpected default value for location, got %s", got["location"].DefaultValue)
	}
	if got["model"].DefaultValue != "gemini-pro" {
		t.Errorf("Unexpected default value for model, got %s", got["model"].DefaultValue)
	}
}

func TestVertexLLM_FormatMessages(t *testing.T) {
	llm := NewVertexLLM(&types.ClientConfig{
		MaxTokens:   1024,
		Temperature: 0.7,
		TopP:        0.9,
		TopK:        40,
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
	genConfig, ok := payload["generation_config"].(map[string]interface{})
	if !ok {
		t.Errorf("FormatMessages() missing generation_config")
		return
	}

	expectedConfig := map[string]interface{}{
		"max_output_tokens": 1024,
		"temperature":       0.7,
		"top_p":             0.9,
		"top_k":             40,
	}

	for k, v := range expectedConfig {
		if genConfig[k] != v {
			t.Errorf("generation_config[%s] = %v, want %v", k, genConfig[k], v)
		}
	}
}

func TestVertexLLM_GetUsage(t *testing.T) {
	llm := NewVertexLLM(&types.ClientConfig{})
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "standard usage",
			data: []byte(`{
                "metadata": {
                    "tokenMetadata": {
                        "inputTokenCount": 10,
                        "outputTokenCount": 20,
                        "totalTokenCount": 30
                    }
                }
            }`),
			want:    "Token usage> input: 10, output: 20, total: 30",
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
