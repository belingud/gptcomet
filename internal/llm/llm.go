package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"maps"

	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// LLM is the interface that all LLM providers must implement
type LLM interface {
	// Name returns the name of the provider
	Name() string

	// BuildURL builds the API URL
	BuildURL() string
	// GetRequiredConfig returns provider-specific configuration requirements
	GetRequiredConfig() map[string]config.ConfigRequirement

	// FormatMessages formats messages for the provider's API
	FormatMessages(message string) (interface{}, error)

	// MakeRequest makes a request to the API
	MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error)

	// GetUsage returns usage information for the provider
	GetUsage(data []byte) (string, error)

	// BuildHeaders builds request headers
	BuildHeaders() map[string]string

	// ParseResponse parses the response from the API
	ParseResponse(response []byte) (string, error)
}

// BaseLLM provides common functionality for all LLM providers
type BaseLLM struct {
	Config *types.ClientConfig
}

// NewBaseLLM creates a new BaseLLM.
//
// If config is nil, it sets default values for the required configuration
// options. Otherwise, it uses the values provided in config.
func NewBaseLLM(config *types.ClientConfig) *BaseLLM {
	if config == nil {
		config = &types.ClientConfig{}
	}
	// Set default values if not provided
	if config.CompletionPath == nil {
		defaultPath := "chat/completions"
		config.CompletionPath = &defaultPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
	}
	return &BaseLLM{
		Config: config,
	}
}

// GetRequiredConfig returns provider-specific configuration requirements
//
// The map keys are the configuration option names, and the values are
// config.ConfigRequirement structs that define the default value and
// prompt message for each option.
//
// The default values are only used if the user does not provide a value
// for the option.
func (b *BaseLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "",
			PromptMessage: "Enter API Base URL",
		},
		"model": {
			DefaultValue:  "",
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

// FormatMessages formats messages for the provider's API
//
// This is a default implementation which should be overridden by the
// provider if it needs to format the messages differently.
func (b *BaseLLM) FormatMessages(message string) (interface{}, error) {
	messages := []types.Message{}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	payload := map[string]interface{}{
		"model":                 b.Config.Model,
		"messages":              messages,
		"max_tokens":            b.Config.MaxTokens, // OpenAI: This value is now deprecated in favor of max_completion_tokens
		"max_completion_tokens": b.Config.MaxTokens,
	}
	if b.Config.Temperature != 0 {
		payload["temperature"] = b.Config.Temperature
	}
	if b.Config.TopP != 0 {
		payload["top_p"] = b.Config.TopP
	}
	if b.Config.FrequencyPenalty != 0 {
		payload["frequency_penalty"] = b.Config.FrequencyPenalty
	}
	if b.Config.PresencePenalty != 0 {
		payload["presence_penalty"] = b.Config.PresencePenalty
	}

	return payload, nil
}

// BuildHeaders provides a default implementation for building headers
func (b *BaseLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if b.Config.APIKey != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", b.Config.APIKey)
	}
	maps.Copy(headers, b.Config.ExtraHeaders)
	return headers
}

// BuildURL builds the API URL by trimming and joining the API base and completion path.
func (b *BaseLLM) BuildURL() string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(b.Config.APIBase, "/"), strings.TrimPrefix(*b.Config.CompletionPath, "/"))
}

// ParseResponse parses the response from the API according to the provider's
// configuration. It first tries to extract the answer using the answer path
// specified in the configuration. If the answer is not found, it returns an
// error. If the answer is found, it trims any leading or trailing triple backticks
// if present, and returns the trimmed text.
func (b *BaseLLM) ParseResponse(response []byte) (string, error) {
	result := gjson.GetBytes(response, b.Config.AnswerPath)
	if !result.Exists() {
		return "", fmt.Errorf("failed to parse response: %s", string(response))
	}
	text := result.String()
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
	}
	return strings.TrimSpace(text), nil
}

// GetUsage returns a string representing the token usage of the response.
// It tries to extract the usage information from the response data using the
// following field names: "prompt_tokens", "completion_tokens", and "total_tokens".
// If the information is not found, it returns an empty string.
func (b *BaseLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usage")
	if !usage.Exists() {
		return "", nil
	}

	var promptTokens, completionTokens, totalTokens int64

	// Try different field names used by different providers
	promptTokens = usage.Get("prompt_tokens").Int()
	completionTokens = usage.Get("completion_tokens").Int()
	totalTokens = usage.Get("total_tokens").Int()

	return fmt.Sprintf(
		"Token usage> prompt: %d, completion: %d, total: %d",
		promptTokens,
		completionTokens,
		totalTokens,
	), nil
}

// MakeRequest makes a request to the provider's API, formats the response, and
// returns the result as a string.
//
// If the request fails or the response is invalid, it returns an error.
//
// The function takes the following parameters:
//   - ctx: the context for the request
//   - client: the HTTP client to use for the request
//   - provider: the provider to make the request to
//   - message: the message to send to the provider
//   - history: the message history to send to the provider
//
// The function returns the response from the provider as a string, or an error
// if the request fails.
func (b *BaseLLM) MakeRequest(ctx context.Context, client *http.Client, provider LLM, message string, stream bool) (string, error) {
	url := provider.BuildURL()
	debug.Printf("🔗 URL: %s", url)
	headers := provider.BuildHeaders()
	payload, err := provider.FormatMessages(message)
	if err != nil {
		return "", fmt.Errorf("failed to format messages: %w", err)
	}
	if stream {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			payloadMap["stream"] = true
		}
	}

	// Merge extra body
	if payloadMap, ok := payload.(map[string]interface{}); ok && len(b.Config.ExtraBody) > 0 {
		maps.Copy(payloadMap, b.Config.ExtraBody)
		payload = payloadMap
	}

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

	fmt.Printf("📤 Sending request to %s...\n", provider.Name())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// If not streaming, read the entire response body
	if !stream {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %s", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		usage, err := provider.GetUsage(respBody)
		if err != nil {
			return "", fmt.Errorf("failed to get usage: %w", err)
		}
		if usage != "" {
			fmt.Printf("%s\n", usage)
		}

		return provider.ParseResponse(respBody)
	}

	// For streaming, we'll return the entire response body as a string
	// In a real implementation, this would parse the SSE stream and call a callback
	// But for now, we'll just read the entire response to maintain compatibility
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return provider.ParseResponse(respBody)
}

// DefaultLLM provides default implementation of LLM interface
type DefaultLLM struct {
	*BaseLLM
}

// NewDefaultLLM creates a new DefaultLLM.
//
// If config is nil, it sets default values for the required configuration
// options. Otherwise, it uses the values provided in config.
func NewDefaultLLM(config *types.ClientConfig) *DefaultLLM {
	if config == nil {
		config = &types.ClientConfig{}
	}
	// Set default values if not provided
	if config.CompletionPath == nil {
		defaultPath := "chat/completions"
		config.CompletionPath = &defaultPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "choices.0.message.content"
	}
	return &DefaultLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the name of the provider, which is "default" for DefaultLLM.
func (d *DefaultLLM) Name() string {
	return d.Config.Provider
}

// MakeRequest implements the LLM interface for DefaultLLM.
func (d *DefaultLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return d.BaseLLM.MakeRequest(ctx, client, d, message, stream)
}
