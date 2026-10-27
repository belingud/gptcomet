package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	// DefaultBedrockRegion is the default AWS region for the Bedrock runtime endpoint.
	DefaultBedrockRegion = "us-east-1"
	// DefaultBedrockAPIBase is the OpenAI-compatible Bedrock runtime endpoint for DefaultBedrockRegion.
	DefaultBedrockAPIBase = "https://bedrock-runtime.us-east-1.amazonaws.com/openai/v1"
	// DefaultBedrockModel is the default model used with Bedrock Chat Completions.
	DefaultBedrockModel = "openai.gpt-oss-120b-1:0"
)

// BedrockLLM implements the LLM interface for Amazon Bedrock's
// OpenAI-compatible Chat Completions API.
type BedrockLLM struct {
	*BaseLLM
}

// BedrockAPIBase returns the OpenAI-compatible Bedrock runtime endpoint for a region.
func BedrockAPIBase(region string) string {
	if region == "" {
		region = DefaultBedrockRegion
	}
	return fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/openai/v1", region)
}

// NewBedrockLLM creates a new BedrockLLM.
func NewBedrockLLM(config *types.ClientConfig) *BedrockLLM {
	if config.Region == "" {
		config.Region = DefaultBedrockRegion
	}
	if config.APIBase == "" {
		config.APIBase = BedrockAPIBase(config.Region)
	}

	BuildStandardConfigSimple(config, BedrockAPIBase(config.Region), DefaultBedrockModel)

	return &BedrockLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the provider name.
func (b *BedrockLLM) Name() string {
	return "bedrock"
}

// GetRequiredConfig returns provider-specific configuration requirements.
func (b *BedrockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"region": {
			DefaultValue:  DefaultBedrockRegion,
			PromptMessage: "Enter AWS region",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter Bedrock API key",
		},
		"model": {
			DefaultValue:  DefaultBedrockModel,
			PromptMessage: "Enter Bedrock model ID",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Bedrock OpenAI-compatible API.
func (b *BedrockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return b.BaseLLM.MakeRequest(ctx, client, b, message, stream)
}
