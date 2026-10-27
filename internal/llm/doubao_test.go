package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewDoubaoLLM(t *testing.T) {
	tests := []struct {
		name      string
		config    *types.ClientConfig
		wantBase  string
		wantModel string
	}{
		{
			name:      "default config",
			config:    &types.ClientConfig{},
			wantBase:  DefaultDoubaoAPIBase,
			wantModel: DefaultDoubaoModel,
		},
		{
			name: "custom endpoint id",
			config: &types.ClientConfig{
				Model: "ep-20261027-example",
			},
			wantBase:  DefaultDoubaoAPIBase,
			wantModel: "ep-20261027-example",
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://ark.example.com/api/v3",
				Model:   "doubao-custom",
			},
			wantBase:  "https://ark.example.com/api/v3",
			wantModel: "doubao-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDoubaoLLM(tt.config)
			if got.Config.APIBase != tt.wantBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.wantBase)
			}
			if got.Config.Model != tt.wantModel {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.wantModel)
			}
		})
	}
}

func TestDoubaoLLM_Name(t *testing.T) {
	llm := NewDoubaoLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "doubao" {
		t.Errorf("Name() = %s, want doubao", got)
	}
}

func TestDoubaoLLM_GetRequiredConfig(t *testing.T) {
	llm := NewDoubaoLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	for _, key := range []string{"api_base", "api_key", "model", "max_tokens"} {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}
	if got["api_base"].DefaultValue != DefaultDoubaoAPIBase {
		t.Errorf("api_base default = %s, want %s", got["api_base"].DefaultValue, DefaultDoubaoAPIBase)
	}
	if got["model"].DefaultValue != DefaultDoubaoModel {
		t.Errorf("model default = %s, want %s", got["model"].DefaultValue, DefaultDoubaoModel)
	}
}

func TestDoubaoLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name:   "default url",
			config: &types.ClientConfig{},
			want:   "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
		},
		{
			name: "trailing slash",
			config: &types.ClientConfig{
				APIBase: "https://ark.cn-beijing.volces.com/api/v3/",
			},
			want: "https://ark.cn-beijing.volces.com/api/v3/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewDoubaoLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDoubaoLLM_BuildHeaders(t *testing.T) {
	llm := NewDoubaoLLM(&types.ClientConfig{APIKey: "test-key"})
	got := llm.BuildHeaders()

	if got["Authorization"] != "Bearer test-key" {
		t.Errorf("Authorization = %s, want Bearer test-key", got["Authorization"])
	}
	if got["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got["Content-Type"])
	}
}
