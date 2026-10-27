package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
)

func TestNewQianfanLLM(t *testing.T) {
	tests := []struct {
		name      string
		config    *types.ClientConfig
		wantBase  string
		wantModel string
	}{
		{
			name:      "default config",
			config:    &types.ClientConfig{},
			wantBase:  DefaultQianfanAPIBase,
			wantModel: DefaultQianfanModel,
		},
		{
			name: "custom config",
			config: &types.ClientConfig{
				APIBase: "https://custom.qianfan.example.com/v2",
				Model:   "ernie-custom",
			},
			wantBase:  "https://custom.qianfan.example.com/v2",
			wantModel: "ernie-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewQianfanLLM(tt.config)
			if got.Config.APIBase != tt.wantBase {
				t.Errorf("APIBase = %s, want %s", got.Config.APIBase, tt.wantBase)
			}
			if got.Config.Model != tt.wantModel {
				t.Errorf("Model = %s, want %s", got.Config.Model, tt.wantModel)
			}
		})
	}
}

func TestQianfanLLM_Name(t *testing.T) {
	llm := NewQianfanLLM(&types.ClientConfig{})
	if got := llm.Name(); got != "qianfan" {
		t.Errorf("Name() = %s, want qianfan", got)
	}
}

func TestQianfanLLM_GetRequiredConfig(t *testing.T) {
	llm := NewQianfanLLM(&types.ClientConfig{})
	got := llm.GetRequiredConfig()

	for _, key := range []string{"api_base", "api_key", "model", "max_tokens"} {
		if _, exists := got[key]; !exists {
			t.Errorf("GetRequiredConfig() missing key %s", key)
		}
	}
	if got["api_base"].DefaultValue != DefaultQianfanAPIBase {
		t.Errorf("api_base default = %s, want %s", got["api_base"].DefaultValue, DefaultQianfanAPIBase)
	}
	if got["model"].DefaultValue != DefaultQianfanModel {
		t.Errorf("model default = %s, want %s", got["model"].DefaultValue, DefaultQianfanModel)
	}
}

func TestQianfanLLM_BuildURL(t *testing.T) {
	tests := []struct {
		name   string
		config *types.ClientConfig
		want   string
	}{
		{
			name:   "default url",
			config: &types.ClientConfig{},
			want:   "https://qianfan.baidubce.com/v2/chat/completions",
		},
		{
			name: "trailing slash",
			config: &types.ClientConfig{
				APIBase: "https://qianfan.baidubce.com/v2/",
			},
			want: "https://qianfan.baidubce.com/v2/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewQianfanLLM(tt.config)
			if got := llm.BuildURL(); got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestQianfanLLM_BuildHeaders(t *testing.T) {
	llm := NewQianfanLLM(&types.ClientConfig{APIKey: "test-key"})
	got := llm.BuildHeaders()

	if got["Authorization"] != "Bearer test-key" {
		t.Errorf("Authorization = %s, want Bearer test-key", got["Authorization"])
	}
	if got["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", got["Content-Type"])
	}
}
