package llm

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

const (
	DefaultLongCatAPIBase = "https://api.longcat.chat/openai"
	DefaultLongCatModel   = "LongCat-Flash-Chat"
)

// LongCatLLM implements the LLM interface for LongCat
type LongCatLLM struct {
	*BaseLLM
}

// NewLongCatLLM creates a new LongCatLLM
func NewLongCatLLM(config *types.ClientConfig) *LongCatLLM {
	BuildStandardConfigSimple(config, DefaultLongCatAPIBase, DefaultLongCatModel)
	return &LongCatLLM{
		BaseLLM: NewBaseLLM(config),
	}
}

func (l *LongCatLLM) Name() string {
	return "longcat"
}

// GetRequiredConfig returns provider-specific configuration requirements
func (l *LongCatLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  DefaultLongCatAPIBase,
			PromptMessage: "Enter LongCat API base",
		},
		"api_key": {
			DefaultValue:  "",
			PromptMessage: "Enter API key",
		},
		"model": {
			DefaultValue:  DefaultLongCatModel,
			PromptMessage: "Enter model name",
		},
		"max_tokens": {
			DefaultValue:  "1024",
			PromptMessage: "Enter max tokens",
		},
	}
}

// MakeRequest makes a request to the LongCat API
func (l *LongCatLLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
	return l.BaseLLM.MakeRequest(ctx, client, l, message, stream)
}
