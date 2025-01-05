package llm

import (
	"testing"

	"github.com/belingud/go-gptcomet/pkg/types"
)

func TestNewBaseLLM(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   struct {
			completionPath string
			answerPath     string
		}
	}{
		{
			name:   "nil config",
			config: nil,
			want: struct {
				completionPath string
				answerPath     string
			}{
				completionPath: "chat/completions",
				answerPath:     "choices.0.message.content",
			},
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				CompletionPath: "custom/path",
				AnswerPath:     "custom.path",
			},
			want: struct {
				completionPath string
				answerPath     string
			}{
				completionPath: "custom/path",
				answerPath:     "custom.path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBaseLLM(tt.config)
			if got.Config.CompletionPath != tt.want.completionPath {
				t.Errorf("CompletionPath = %s, want %s", got.Config.CompletionPath, tt.want.completionPath)
			}
			if got.Config.AnswerPath != tt.want.answerPath {
				t.Errorf("AnswerPath = %s, want %s", got.Config.AnswerPath, tt.want.answerPath)
			}
		})
	}
}

func TestBaseLLM_FormatMessages(t *testing.T) {
	llm := NewBaseLLM(&types.ClientConfig{
		Model:            "test-model",
		MaxTokens:        1024,
		Temperature:      0.7,
		TopP:             0.9,
		FrequencyPenalty: 0.5,
		PresencePenalty:  0.5,
	})

	message := "test message"
	history := []types.Message{
		{Role: "user", Content: "previous message"},
	}

	got, err := llm.FormatMessages(message, history)
	if err != nil {
		t.Errorf("FormatMessages() error = %v", err)
		return
	}

	payload, ok := got.(map[string]interface{})
	if !ok {
		t.Errorf("FormatMessages() returned wrong type")
		return
	}

	// 验证基本字段
	if payload["model"] != "test-model" {
		t.Errorf("model = %v, want test-model", payload["model"])
	}
	if payload["max_tokens"] != 1024 {
		t.Errorf("max_tokens = %v, want 1024", payload["max_tokens"])
	}

	// 验证可选参数
	if payload["temperature"] != 0.7 {
		t.Errorf("temperature = %v, want 0.7", payload["temperature"])
	}
	if payload["top_p"] != 0.9 {
		t.Errorf("top_p = %v, want 0.9", payload["top_p"])
	}

	// 验证消息格式
	messages, ok := payload["messages"].([]types.Message)
	if !ok {
		t.Errorf("messages wrong type")
		return
	}
	if len(messages) != 2 {
		t.Errorf("got %d messages, want 2", len(messages))
	}
}

func TestBaseLLM_BuildHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   map[string]string
	}{
		{
			name: "basic headers",
			config: &types.ClientConfig{
				APIKey: "test-key",
			},
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-key",
			},
		},
		{
			name: "with extra headers",
			config: &types.ClientConfig{
				APIKey: "test-key",
				ExtraHeaders: map[string]string{
					"X-Custom": "custom-value",
				},
			},
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer test-key",
				"X-Custom":      "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewBaseLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("header[%s] = %s, want %s", k, got[k], v)
				}
			}
		})
	}
}

func TestBaseLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://api.example.com",
				CompletionPath: "chat/completions",
			},
			want: "https://api.example.com/chat/completions",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://api.example.com/",
				CompletionPath: "/chat/completions",
			},
			want: "https://api.example.com/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewBaseLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBaseLLM_ParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		response []byte
		config   *types.ClientConfig
		want     string
		wantErr  bool
	}{
		{
			name:     "standard response",
			response: []byte(`{"choices":[{"message":{"content":"test response"}}]}`),
			config: &types.ClientConfig{
				AnswerPath: "choices.0.message.content",
			},
			want:    "test response",
			wantErr: false,
		},
		{
			name:     "response with code blocks",
			response: []byte("{\"choices\":[{\"message\":{\"content\":\"```test response```\"}}]}"),
			config: &types.ClientConfig{
				AnswerPath: "choices.0.message.content",
			},
			want:    "test response",
			wantErr: false,
		},
		{
			name:     "invalid response",
			response: []byte(`{"error": "test error"}`),
			config: &types.ClientConfig{
				AnswerPath: "choices.0.message.content",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewBaseLLM(tt.config)
			got, err := llm.ParseResponse(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseResponse() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBaseLLM_GetUsage(t *testing.T) {
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
			llm := NewBaseLLM(&types.ClientConfig{})
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
