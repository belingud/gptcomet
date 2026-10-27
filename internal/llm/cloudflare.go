package llm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	// DefaultCloudflareAPIBase is the base URL for Cloudflare Workers AI.
	DefaultCloudflareAPIBase = "https://api.cloudflare.com/client/v4/accounts"
	// DefaultCloudflareModel is the default Workers AI model used by GPTComet.
	DefaultCloudflareModel = "@cf/openai/gpt-oss-120b"
)

// CloudflareLLM implements the LLM interface for Cloudflare Workers AI.
type CloudflareLLM struct {
	*BaseLLM
}

// CloudflareAPIBase returns the OpenAI-compatible Workers AI endpoint for an account.
func CloudflareAPIBase(accountID string) string {
	if accountID == "" {
		return DefaultCloudflareAPIBase
	}
	return fmt.Sprintf("%s/%s/ai/v1", DefaultCloudflareAPIBase, accountID)
}

// NewCloudflareLLM creates a new CloudflareLLM.
func NewCloudflareLLM(config *types.ClientConfig) *CloudflareLLM {
	if config.APIBase == "" {
		config.APIBase = CloudflareAPIBase(config.AccountID)
	}

	BuildStandardConfigSimple(config, CloudflareAPIBase(config.AccountID), DefaultCloudflareModel)

	return &CloudflareLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

// Name returns the provider name.
func (c *CloudflareLLM) Name() string {
	return "cloudflare"
}

// GetRequiredConfig returns provider-specific configuration requirements.
func (c *CloudflareLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"account_id": {
			DefaultValue:  "",
			PromptMessage: "Enter Cloudflare account ID",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter Cloudflare API token",
		},
		"model": {
			DefaultValue:  DefaultCloudflareModel,
			PromptMessage: "Enter Workers AI model ID",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the Cloudflare Workers AI OpenAI-compatible API.
func (c *CloudflareLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return c.BaseLLM.MakeRequest(ctx, client, c, message, stream)
}
