package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// VertexLLM implements the LLM interface for Google Cloud Vertex AI
type VertexLLM struct {
	*BaseLLM
}

// NewVertexLLM creates a new VertexLLM
func NewVertexLLM(config *types.ClientConfig) *VertexLLM {
	if config.APIBase == "" {
		config.APIBase = "https://us-central1-aiplatform.googleapis.com/v1"
	}
	if config.Model == "" {
		config.Model = "gemini-1.5-flash"
	}

	if config.ProjectID == "" {
		config.ProjectID = "default-project"
	}

	if config.Location == "" {
		config.Location = "us-central1"
	}

	if config.CompletionPath == nil {
		completionPath := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s:generateContent",
			config.ProjectID, config.Location, config.Model)
		config.CompletionPath = &completionPath
	}
	if config.AnswerPath == "" {
		config.AnswerPath = "candidates.0.content.parts.0.text"
	}

	return &VertexLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (v *VertexLLM) Name() string {
	return "vertex"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (v *VertexLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  "https://us-central1-aiplatform.googleapis.com/v1",
			PromptMessage: "Enter Vertex AI API Base URL",
		},
		"project_id": {
			DefaultValue:  "",
			PromptMessage: "Enter Google Cloud project ID",
		},
		"location": {
			DefaultValue:  "us-central1",
			PromptMessage: "Enter location (e.g., us-central1)",
		},
		"model": {
			DefaultValue:  "gemini-pro",
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

// BuildURL builds the API URL for Vertex AI
func (v *VertexLLM) BuildURL() string {
	return fmt.Sprintf("%s/%s", v.Config.APIBase, fmt.Sprintf(*v.Config.CompletionPath, v.Config.Model))
}

// FormatMessages formats messages for Vertex AI
func (v *VertexLLM) FormatMessages(message string) (interface{}, error) {
	contents := []map[string]interface{}{
		{
			"role": "user",
			"parts": []map[string]string{
				{
					"text": message,
				},
			},
		},
	}

	payload := map[string]interface{}{
		"contents":          contents,
		"generation_config": map[string]interface{}{},
	}

	if v.Config.MaxTokens != 0 {
		payload["generation_config"].(map[string]interface{})["max_output_tokens"] = v.Config.MaxTokens
	}
	if v.Config.Temperature != 0.0 {
		payload["generation_config"].(map[string]interface{})["temperature"] = v.Config.Temperature
	}
	if v.Config.TopP != 0.0 {
		payload["generation_config"].(map[string]interface{})["top_p"] = v.Config.TopP
	}
	if v.Config.TopK != 0.0 {
		payload["generation_config"].(map[string]interface{})["top_k"] = v.Config.TopK
	}

	return payload, nil
}

// GetUsage returns usage information for the provider
func (v *VertexLLM) GetUsage(data []byte) (string, error) {
	usage := gjson.GetBytes(data, "metadata.tokenMetadata")
	if !usage.Exists() {
		return "", nil
	}

	return fmt.Sprintf(
		"Token usage> input: %d, output: %d, total: %d",
		usage.Get("inputTokenCount").Int(),
		usage.Get("outputTokenCount").Int(),
		usage.Get("totalTokenCount").Int(),
	), nil
}

// MakeRequest makes a request to the API
func (v *VertexLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return v.BaseLLM.MakeRequest(ctx, client, v, message, stream)
}
