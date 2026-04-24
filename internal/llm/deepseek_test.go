package llm

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/logger"
	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewDeepSeekLLM(t *testing.T) {
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
				apiBase: "https://api.deepseek.com/v1",
				model:   DefaultDeepSeekModel,
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
			got := NewDeepSeekLLM(tt.config)
			if got.Config.APIBase != tt.want.apiBase {
				t.Errorf("NewDeepSeekLLM().Config.APIBase = %v, want %v", got.Config.APIBase, tt.want.apiBase)
			}
			if got.Config.Model != tt.want.model {
				t.Errorf("NewDeepSeekLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.model)
			}
		})
	}
}

func TestDeepSeekLLM_Name(t *testing.T) {
	llm := NewDeepSeekLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "deepseek" {
		t.Errorf("DeepSeekLLM.Name() = %v, want %v", got, "deepseek")
	}
}

func TestDeepSeekLLM_GetRequiredConfig(t *testing.T) {
	llm := NewDeepSeekLLM(&types.ClientConfig{})
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

	if got["api_base"].DefaultValue != "https://api.deepseek.com/v1" {
		t.Errorf("Unexpected default value for api_base: got %v, want %v",
			got["api_base"].DefaultValue, "https://api.deepseek.com/v1")
	}
	if got["model"].DefaultValue != DefaultDeepSeekModel {
		t.Errorf("Unexpected default value for model: got %v, want %v",
			got["model"].DefaultValue, DefaultDeepSeekModel)
	}
}

func TestDeepSeekLLM_BuildURL(t *testing.T) {
	defaultPath := "chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "default url",
			config: &types.ClientConfig{
				APIBase:        "https://api.deepseek.com/v1",
				CompletionPath: &defaultPath,
			},
			want: "https://api.deepseek.com/v1/chat/completions",
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
				APIBase:        "https://api.deepseek.com/v1/",
				CompletionPath: &defaultPath,
			},
			want: "https://api.deepseek.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewDeepSeekLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("DeepSeekLLM.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepSeekLLM_MakeRequestLogsEmptyContentWarning(t *testing.T) {
	rawResponse := `{"choices":[{"message":{"content":"   "}}]}`
	completionPath := "chat/completions?api_key=url-secret"
	client := newDeepSeekTestClient(rawResponse)
	logOutput := captureDeepSeekLoggerOutput(t)
	llm := NewDeepSeekLLM(&types.ClientConfig{
		APIBase:        "https://proxy-user:proxy-secret@deepseek.test/v1",
		APIKey:         "secret-api-key",
		Model:          "deepseek-test-model",
		CompletionPath: &completionPath,
		MaxTokens:      16,
		AnswerPath:     "choices.0.message.content",
	})

	got, err := llm.MakeRequest(context.Background(), client, "private prompt and diff", false)
	if err != nil {
		t.Fatalf("MakeRequest() error = %v", err)
	}
	if got != "" {
		t.Errorf("MakeRequest() = %q, want empty string", got)
	}

	logs := logOutput.String()
	wantParts := []string{
		"DeepSeek returned empty content",
		"provider=deepseek",
		"model=deepseek-test-model",
		"status=200",
		"url=https://deepseek.test/v1/chat/completions",
		"answer_path=choices.0.message.content",
		"raw_response=" + rawResponse,
	}
	for _, want := range wantParts {
		if !strings.Contains(logs, want) {
			t.Errorf("warning log missing %q in %q", want, logs)
		}
	}

	for _, forbidden := range []string{"secret-api-key", "url-secret", "proxy-user", "proxy-secret", "Authorization", "private prompt and diff"} {
		if strings.Contains(logs, forbidden) {
			t.Errorf("warning log contains sensitive value %q in %q", forbidden, logs)
		}
	}
}

func TestDeepSeekLLM_MakeRequestDoesNotLogWarningForNormalContent(t *testing.T) {
	client := newDeepSeekTestClient(`{"choices":[{"message":{"content":" normal response "}}]}`)
	logOutput := captureDeepSeekLoggerOutput(t)
	llm := NewDeepSeekLLM(&types.ClientConfig{
		APIBase:    "https://deepseek.test/v1",
		APIKey:     "secret-api-key",
		Model:      "deepseek-test-model",
		MaxTokens:  16,
		AnswerPath: "choices.0.message.content",
	})

	got, err := llm.MakeRequest(context.Background(), client, "private prompt and diff", false)
	if err != nil {
		t.Fatalf("MakeRequest() error = %v", err)
	}
	if got != "normal response" {
		t.Errorf("MakeRequest() = %q, want %q", got, "normal response")
	}

	logs := logOutput.String()
	if strings.Contains(logs, "DeepSeek returned empty content") {
		t.Errorf("warning log emitted for normal content: %q", logs)
	}
	if strings.Contains(logs, "secret-api-key") || strings.Contains(logs, "private prompt and diff") {
		t.Errorf("log contains sensitive data: %q", logs)
	}
}

func TestDeepSeekLLM_MakeRequestPreservesParseErrorForMissingAnswerPath(t *testing.T) {
	rawResponse := `{"choices":[{"message":{}}]}`
	client := newDeepSeekTestClient(rawResponse)
	logOutput := captureDeepSeekLoggerOutput(t)
	llm := NewDeepSeekLLM(&types.ClientConfig{
		APIBase:    "https://deepseek.test/v1",
		APIKey:     "secret-api-key",
		Model:      "deepseek-test-model",
		MaxTokens:  16,
		AnswerPath: "choices.0.message.content",
	})

	got, err := llm.MakeRequest(context.Background(), client, "private prompt and diff", false)
	if err == nil {
		t.Fatal("MakeRequest() error = nil, want parse error")
	}
	if got != "" {
		t.Errorf("MakeRequest() = %q, want empty string", got)
	}
	if !strings.Contains(err.Error(), "failed to parse response") || !strings.Contains(err.Error(), rawResponse) {
		t.Errorf("MakeRequest() error = %v, want parse error with raw response", err)
	}

	logs := logOutput.String()
	if strings.Contains(logs, "DeepSeek returned empty content") {
		t.Errorf("warning log emitted for missing answer path: %q", logs)
	}
}

type deepSeekRoundTripFunc func(*http.Request) (*http.Response, error)

func (f deepSeekRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newDeepSeekTestClient(responseBody string) *http.Client {
	return &http.Client{
		Transport: deepSeekRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Request:    req,
			}, nil
		}),
	}
}

func captureDeepSeekLoggerOutput(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetLevel(logger.DebugLevel)
	logger.EnableIcons(false)
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
		logger.SetLevel(logger.DebugLevel)
		logger.EnableIcons(true)
	})

	return &buf
}
