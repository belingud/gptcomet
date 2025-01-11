package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// OpenAILLM is the OpenAI LLM provider implementation
type OpenAILLM struct {
	*BaseLLM
}

// NewOpenAILLM creates a new OpenAILLM
func NewOpenAILLM(config *types.ClientConfig) *OpenAILLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-4o"
	}
	if config.CompletionPath == nil {
		completionPath := "chat/completions"
		config.CompletionPath = &completionPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
	}

	return &OpenAILLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// GetRequiredConfig returns provider-specific configuration requirements
func (o *OpenAILLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.openai.com/v1",
			PromptMessage: "Enter OpenAI API base",
		},
		"model": {
			DefaultValue:  "gpt-4o",
			PromptMessage: "Enter model name",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

func (o *OpenAILLM) Name() string {
	return "openai"
}

// FormatMessages formats messages for OpenAI API
func (o *OpenAILLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	return o.BaseLLM.FormatMessages(message, history)
}

// BuildURL builds the API URL
func (o *OpenAILLM) BuildURL() string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(o.Config.APIBase, "/"), strings.TrimPrefix(*o.Config.CompletionPath, "/"))
}

// BuildHeaders builds request headers
func (o *OpenAILLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", o.Config.APIKey),
		"Content-Type":  "application/json",
	}
	for k, v := range o.Config.ExtraHeaders {
		headers[k] = v
	}
	// print headers detail
	for k, v := range headers {
		debug.Printf("Header: %s: %s", k, v)
	}
	return headers
}

// GetUsage returns usage information for the provider
func (o *OpenAILLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usage")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> prompt tokens: %d, completion tokens: %d, total tokens: %d",
		usage.Get("prompt_tokens").Int(),
		usage.Get("completion_tokens").Int(),
		usage.Get("total_tokens").Int(),
	), nil
}

// MakeRequest makes a request to the API
func (o *OpenAILLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	url := o.BuildURL()
	debug.Printf("API URL: %s", url)
	headers := o.BuildHeaders()
	payload, err := o.FormatMessages(message, history)
	if err != nil {
		return "", fmt.Errorf("failed to format messages: %w", err)
	}

	debug.Printf("📤 Sending request...")

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	debug.Printf("Response: %s", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	usage, err := o.GetUsage(respBody)
	if err != nil {
		return "", fmt.Errorf("failed to get usage: %w", err)
	}
	if usage != "" {
		debug.Printf("%s", usage)
	}

	return o.ParseResponse(respBody)
}
