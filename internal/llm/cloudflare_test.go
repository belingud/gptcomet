package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewCloudflareLLM(t *testing.T) {
	tests := []struct {
		name      string
		config    *types.ClientConfig
		wantBase  string
		wantModel string
	}{
		{
			name: "default config with account",
			config: &types.ClientConfig{
				AccountID: "account-123",
			},
			wantBase:  "https://api.cloudflare.com/client/v4/accounts/account-123/ai/v1",
			wantModel: DefaultCloudflareModel,
		},
		{
			name:      "missing account uses base url",
			config:    &types.ClientConfig{},
			wantBase:  DefaultCloudflareAPIBase,
			wantModel: DefaultCloudflareModel,
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase:   "https://proxy.example.com/cloudflare/ai/v1",
				AccountID: "ignored",
				Model:     "@cf/custom/model",
			},
			wantBase:  "https://proxy.example.com/cloudflare/ai/v1",
			wantModel: "@cf/custom/model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCloudflareLLM(tt.config)
			if got.Config.APIBase != tt.wantBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.wantBase)
			}
			if got.Config.Model != tt.wantModel {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.wantModel)
			}
		})
	}
}

func TestCloudflareLLM_Name(t *testing.T) {
	llm := NewCloudflareLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "cloudflare" {
		t.Errorf("Name() = %s, want cloudflare", got)
	}
}

func TestCloudflareLLM_GetRequiredConfig(t *testing.T) {
	llm := NewCloudflareLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	for _, key := range []string{"account_id", "api_key", "model", "max_tokens"} {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}
	if got["model"].DefaultValue != DefaultCloudflareModel {
		t.Errorf("model default = %s, want %s", got["model"].DefaultValue, DefaultCloudflareModel)
	}
}

func TestCloudflareLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name: "account endpoint",
			config: &types.ClientConfig{
				AccountID: "account-123",
			},
			want: "https://api.cloudflare.com/client/v4/accounts/account-123/ai/v1/chat/completions",
		},
		{
			name: "custom base with trailing slash",
			config: &types.ClientConfig{
				APIBase: "https://proxy.example.com/cloudflare/ai/v1/",
			},
			want: "https://proxy.example.com/cloudflare/ai/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewCloudflareLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCloudflareLLM_BuildHeaders(t *testing.T) {
	llm := NewCloudflareLLM(&types.ClientConfig{APIKey: "test-token"})
	got := llm.BuildHeaders()

	if got["Authorization"] != "Bearer test-token" {
		t.Errorf("Authorization = %s, want Bearer test-token", got["Authorization"])
	}
	if got["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got["Content-Type"])
	}
}

func TestCloudflareAPIBase(t *testing.T) {
	if got := CloudflareAPIBase("account-123"); got != "https://api.cloudflare.com/client/v4/accounts/account-123/ai/v1" {
		t.Errorf("CloudflareAPIBase() = %s", got)
	}
	if got := CloudflareAPIBase(""); got != DefaultCloudflareAPIBase {
		t.Errorf("CloudflareAPIBase(\"\") = %s, want %s", got, DefaultCloudflareAPIBase)
	}
}
