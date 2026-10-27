package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewBedrockLLM(t *testing.T) {
	tests := []struct {
		name       string
		config     *types.ClientConfig
		wantBase   string
		wantModel  string
		wantRegion string
	}{
		{
			name:       "default config",
			config:     &types.ClientConfig{},
			wantBase:   DefaultBedrockAPIBase,
			wantModel:  DefaultBedrockModel,
			wantRegion: DefaultBedrockRegion,
		},
		{
			name: "custom region",
			config: &types.ClientConfig{
				Region: "eu-west-1",
			},
			wantBase:   "https://bedrock-runtime.eu-west-1.amazonaws.com/openai/v1",
			wantModel:  DefaultBedrockModel,
			wantRegion: "eu-west-1",
		},
		{
			name: "custom api base and model",
			config: &types.ClientConfig{
				APIBase: "https://bedrock.example.com/openai/v1",
				Model:   "custom-model",
				Region:  "us-west-2",
			},
			wantBase:   "https://bedrock.example.com/openai/v1",
			wantModel:  "custom-model",
			wantRegion: "us-west-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBedrockLLM(tt.config)
			if got.Config.APIBase != tt.wantBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.wantBase)
			}
			if got.Config.Model != tt.wantModel {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.wantModel)
			}
			if got.Config.Region != tt.wantRegion {
				t.Errorf("Region = %s, want %s", got.Config.Region, tt.wantRegion)
			}
		})
	}
}

func TestBedrockLLM_Name(t *testing.T) {
	llm := NewBedrockLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "bedrock" {
		t.Errorf("Name() = %s, want bedrock", got)
	}
}

func TestBedrockLLM_GetRequiredConfig(t *testing.T) {
	llm := NewBedrockLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	for _, key := range []string{"region", "api_key", "model", "max_tokens"} {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}
	if got["region"].DefaultValue != DefaultBedrockRegion {
		t.Errorf("region default = %s, want %s", got["region"].DefaultValue, DefaultBedrockRegion)
	}
	if got["model"].DefaultValue != DefaultBedrockModel {
		t.Errorf("model default = %s, want %s", got["model"].DefaultValue, DefaultBedrockModel)
	}
}

func TestBedrockLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "default region",
			config: &types.ClientConfig{
				Region: "us-east-1",
			},
			want: "https://bedrock-runtime.us-east-1.amazonaws.com/openai/v1/chat/completions",
		},
		{
			name: "custom region",
			config: &types.ClientConfig{
				Region: "ap-southeast-1",
			},
			want: "https://bedrock-runtime.ap-southeast-1.amazonaws.com/openai/v1/chat/completions",
		},
		{
			name: "custom base with trailing slash",
			config: &types.ClientConfig{
				APIBase: "https://bedrock.example.com/openai/v1/",
			},
			want: "https://bedrock.example.com/openai/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewBedrockLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBedrockLLM_BuildHeaders(t *testing.T) {
	llm := NewBedrockLLM(&types.ClientConfig{
		APIKey: "test-key",
		ExtraHeaders: map[string]string{
			"X-Custom": "custom-value",
		},
	})
	got := llm.BuildHeaders()

	if got["Authorization"] != "Bearer test-key" {
		t.Errorf("Authorization = %s, want Bearer test-key", got["Authorization"])
	}
	if got["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got["Content-Type"])
	}
	if got["X-Custom"] != "custom-value" {
		t.Errorf("X-Custom = %s, want custom-value", got["X-Custom"])
	}
}
