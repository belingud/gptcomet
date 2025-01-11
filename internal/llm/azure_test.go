package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewAzureLLM(t *testing.T) {
	completionPath := "deployments/test-deployment/chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   *AzureLLM
	}{
		{
			name: "default config",
			config: &types.ClientConfig{
				APIBase:        "https://test.openai.azure.com",
				DeploymentName: "test-deployment",
			},
			want: &AzureLLM{
				OpenAILLM: &OpenAILLM{
					BaseLLM: &BaseLLM{
						Config: &types.ClientConfig{
							APIBase:        "https://test.openai.azure.com",
							DeploymentName: "test-deployment",
							Model:          "gpt-4o",
							CompletionPath: &completionPath,
						},
					},
				},
			},
		},
		{
			name: "custom model",
			config: &types.ClientConfig{
				APIBase:        "https://test.openai.azure.com",
				DeploymentName: "test-deployment",
				Model:          "custom-model",
			},
			want: &AzureLLM{
				OpenAILLM: &OpenAILLM{
					BaseLLM: &BaseLLM{
						Config: &types.ClientConfig{
							APIBase:        "https://test.openai.azure.com",
							DeploymentName: "test-deployment",
							Model:          "custom-model",
							CompletionPath: &completionPath,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAzureLLM(tt.config)
			if got.Config.Model != tt.want.Config.Model {
				t.Errorf("NewAzureLLM().Config.Model = %v, want %v", got.Config.Model, tt.want.Config.Model)
			}
			if *got.Config.CompletionPath != *tt.want.Config.CompletionPath {
				t.Errorf("NewAzureLLM().Config.CompletionPath = %v, want %v", *got.Config.CompletionPath, *tt.want.Config.CompletionPath)
			}
		})
	}
}

func TestAzureLLM_Name(t *testing.T) {
	llm := NewAzureLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "azure" {
		t.Errorf("AzureLLM.Name() = %v, want %v", got, "azure")
	}
}

func TestAzureLLM_GetRequiredConfig(t *testing.T) {
	llm := NewAzureLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	requiredKeys := []string{
		"api_base",
		"deployment_name",
		"api_key",
		"model",
		"max_tokens",
		"api_version",
	}

	for _, key := range requiredKeys {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %v", key)
		}
	}
}

func TestAzureLLM_BuildURL(t *testing.T) {
	completionPath := "deployments/test-deployment/chat/completions"
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "standard url",
			config: &types.ClientConfig{
				APIBase:        "https://test.openai.azure.com",
				DeploymentName: "test-deployment",
				APIVersion:     "2024-02-15-preview",
				CompletionPath: &completionPath,
			},
			want: "https://test.openai.azure.com/deployments/test-deployment/chat/completions?api-version=2024-02-15-preview",
		},
		{
			name: "url with trailing slash",
			config: &types.ClientConfig{
				APIBase:        "https://test.openai.azure.com/",
				DeploymentName: "test-deployment",
				APIVersion:     "2024-02-15-preview",
				CompletionPath: &completionPath,
			},
			want: "https://test.openai.azure.com/deployments/test-deployment/chat/completions?api-version=2024-02-15-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewAzureLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("AzureLLM.BuildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAzureLLM_BuildHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   map[string]string
	}{
		{
			name: "standard headers",
			config: &types.ClientConfig{
				APIKey:     "test-key",
				APIVersion: "2024-02-15-preview",
			},
			want: map[string]string{
				"Content-Type": "application/json",
				"api-key":      "test-key",
				"api-version":  "2024-02-15-preview",
			},
		},
		{
			name: "headers with extra headers",
			config: &types.ClientConfig{
				APIKey:     "test-key",
				APIVersion: "2024-02-15-preview",
				ExtraHeaders: map[string]string{
					"X-Custom": "custom-value",
				},
			},
			want: map[string]string{
				"Content-Type": "application/json",
				"api-key":      "test-key",
				"api-version":  "2024-02-15-preview",
				"X-Custom":     "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewAzureLLM(tt.config)
			got := llm.BuildHeaders()

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("AzureLLM.BuildHeaders()[%v] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}
