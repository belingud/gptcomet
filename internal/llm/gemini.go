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

// GeminiLLM implements the LLM interface for Gemini
type GeminiLLM struct {
	*BaseLLM
}

// NewGeminiLLM creates a new GeminiLLM
func NewGeminiLLM(config *types.ClientConfig) *GeminiLLM {
	if config.APIBase == "" {
		config.APIBase = "https://generativelanguage.googleapis.com/v1beta/models"
	}
	if config.Model == "" {
		config.Model = "gemini-1.5-flash"
	}
	if config.CompletionPath == "" {
		config.CompletionPath = "generateContent"
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "candidates.0.content.parts.0.text"
	}

	return &GeminiLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (g *GeminiLLM) Name() string {
	return "gemini"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (g *GeminiLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://generativelanguage.googleapis.com/v1beta/models",
			PromptMessage: "Enter Gemini API base",
		},
		"model": {
			DefaultValue:  "gemini-1.5-flash",
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

// FormatMessages formats messages for Gemini API
func (g *GeminiLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	var contents []map[string]interface{}

	contents = append(contents, map[string]interface{}{
		"role":  "user",
		"parts": []map[string]string{{"text": message}},
	})

	payload := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": g.Config.MaxTokens,
		},
	}

	if g.Config.Temperature > 0 {
		payload["generationConfig"].(map[string]interface{})["temperature"] = g.Config.Temperature
	}
	if g.Config.TopP > 0 {
		payload["generationConfig"].(map[string]interface{})["topP"] = g.Config.TopP
	}
	if g.Config.FrequencyPenalty > 0 {
		payload["generationConfig"].(map[string]interface{})["frequencyPenalty"] = g.Config.FrequencyPenalty
	}
	if g.Config.PresencePenalty > 0 {
		payload["generationConfig"].(map[string]interface{})["presencePenalty"] = g.Config.PresencePenalty
	}
	debug.Printf("generationConfig: %v", payload["generationConfig"])

	return payload, nil
}

// BuildURL builds the API URL
func (g *GeminiLLM) BuildURL() string {
	return fmt.Sprintf("%s/%s:generateContent?key=%s", strings.TrimSuffix(g.Config.APIBase, "/"), g.Config.Model, g.Config.APIKey)
}

// BuildHeaders builds request headers
func (g *GeminiLLM) BuildHeaders() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

// GetUsage returns usage information for the provider
func (g *GeminiLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "usageMetadata")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> promptTokenCount: %d, candidatesTokenCount: %d, totalTokenCount: %d",
		usage.Get("promptTokenCount").Int(),
		usage.Get("candidatesTokenCount").Int(),
		usage.Get("totalTokenCount").Int(),
	), nil
}

// MakeRequest makes a request to the API
func (g *GeminiLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	url := g.BuildURL()
	headers := g.BuildHeaders()
	payload, err := g.FormatMessages(message, history)
	if err != nil {
		return "", fmt.Errorf("failed to format messages: %w", err)
	}

	debug.Printf("Sending request...")

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
		return "", fmt.Errorf("failed to read response: %s", err)
	}

	debug.Printf("Response: %s", string(respBody))

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
