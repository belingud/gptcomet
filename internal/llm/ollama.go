package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// OllamaLLM is the Ollama LLM provider implementation
type OllamaLLM struct {
	*BaseLLM
}

// NewOllamaLLM creates a new OllamaLLM
func NewOllamaLLM(config *types.ClientConfig) *OllamaLLM {
	if config.APIBase == "" {
		config.APIBase = "http://localhost:11434/api"
	}
	if config.CompletionPath == nil {
		completionPath := "generate"
		config.CompletionPath = &completionPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "response"
	}
	if config.Model == "" {
		config.Model = "llama2"
	}

	return &OllamaLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (o *OllamaLLM) Name() string {
	return "ollama"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (o *OllamaLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "http://localhost:11434/api",
			PromptMessage: "Enter Ollama API Base URL",
		},
		"model": {
			DefaultValue:  "llama2",
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// FormatMessages formats messages for Ollama API
func (o *OllamaLLM) FormatMessages(message string) (interface{}, error) {
	options := map[string]interface{}{
		"num_predict": o.Config.MaxTokens,
	}

	if o.Config.Temperature != 0 {
		options["temperature"] = o.Config.Temperature
	}
	if o.Config.TopP != 0.0 {
		options["top_p"] = o.Config.TopP
	}
	if o.Config.TopK != 0.0 {
		options["top_k"] = o.Config.TopK
	}
	if o.Config.RepetitionPenalty != 0.0 {
		options["repetition_penalty"] = o.Config.RepetitionPenalty
	}
	if o.Config.Seed != 0 {
		options["seed"] = o.Config.Seed
	}
	if o.Config.NumGPU != 0 {
		options["num_gpu"] = o.Config.NumGPU
	}
	if o.Config.MainGPU != 0 {
		options["main_gpu"] = o.Config.MainGPU
	}
	if o.Config.FrequencyPenalty != 0.0 {
		options["frequency_penalty"] = o.Config.FrequencyPenalty
	}
	if o.Config.PresencePenalty != 0.0 {
		options["presence_penalty"] = o.Config.PresencePenalty
	}

	payload := map[string]interface{}{
		"model":   o.Config.Model,
		"prompt":  message,
		"options": options,
	}

	return payload, nil
}

// GetUsage returns usage information for the provider
func (o *OllamaLLM) GetUsage(data []byte) (string, error) {
	return "", nil
}

// MakeRequest makes a request to the API
func (o *OllamaLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	payload, err := o.FormatMessages(message)
	if err != nil {
		return "", fmt.Errorf("failed to format messages: %w", err)
	}

	if stream {
		payload.(map[string]interface{})["stream"] = true
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s", o.Config.APIBase, *o.Config.CompletionPath)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range o.BuildHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// BuildHeaders builds request headers
func (o *OllamaLLM) BuildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	for k, v := range o.Config.ExtraHeaders {
		headers[k] = v
	}
	return headers
}
