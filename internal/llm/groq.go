package llm

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

const (
	DefaultGroqModel = "llama-3.3-70b-versatile"
)

// GroqLLM implements the LLM interface for Groq
type GroqLLM struct {
	*OpenAILLM
}

// NewGroqLLM creates a new GroqLLM
func NewGroqLLM(config *types.ClientConfig) *GroqLLM {
	if config.APIBase == "" {
		config.APIBase = "https://api.groq.com/openai/v1"
	}
	if config.Model == "" {
		config.Model = DefaultGroqModel
	}

	return &GroqLLM{
		OpenAILLM: NewOpenAILLM(config),
	}
}

func (g *GroqLLM) Name() string {
	return "groq"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (g *GroqLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://api.groq.com/openai/v1",
			PromptMessage: "Enter Groq API base URL",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter Groq API key",
		},
		"model": {
			DefaultValue:  DefaultGroqModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// BuildHeaders builds request headers for Groq
func (g *GroqLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", g.Config.APIKey),
	}
	return headers
}

func (g *GroqLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	messages := append(history, types.Message{
		Role:    "user",
		Content: message,
	})

	payload := map[string]interface{}{
		"model":      g.Config.Model,
		"messages":   messages,
		"max_tokens": g.Config.MaxTokens,
	}
	if g.Config.Temperature != 0 {
		payload["temperature"] = g.Config.Temperature
	}
	if g.Config.TopP != 0 {
		payload["top_p"] = g.Config.TopP
	}
	return payload, nil
}

// GetUsage returns usage information for the provider
//
// The function takes the following parameters:
//   - data: the response data from the provider
//
// The function returns a string describing the token usage and an error if the
// usage information is not found.
func (g *GroqLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usage")
	if !usage.Exists() {
		return "", fmt.Errorf("usage not found")
	}

	return fmt.Sprintf(
		"Token usage> prompt tokens: %d, completion tokens: %d, total tokens: %d",
		usage.Get("prompt_tokens").Int(),
		usage.Get("completion_tokens").Int(),
		usage.Get("total_tokens").Int(),
	), nil
}

// MakeRequest makes a request to the API
func (g *GroqLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	url := g.BuildURL()
	debug.Printf("API URL: %s", url)
	headers := g.BuildHeaders()
	payload, err := g.FormatMessages(message, history)
	if err != nil {
		return "", fmt.Errorf("failed to format messages: %w", err)
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
		debug.Printf("Header: %s: %s", k, v)
		req.Header.Set(k, v)
	}

	debug.Printf("📤 Sending request to %s...", g.Name())
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("❌ failed to send request: %w", err)
	}
	defer resp.Body.Close()
	// Handle gzip encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}
	respBody, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	usage, err := g.GetUsage(respBody)
	if err != nil {
		return "", fmt.Errorf("failed to get usage: %w", err)
	}
	if usage != "" {
		fmt.Printf("%s\n", usage)
	}

	return g.ParseResponse(respBody)
}
