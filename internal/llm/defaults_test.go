package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/config/defaults"
)

func TestDefaultModelsMatchGlobalConfig(t *testing.T) {
	if DefaultOpenAIModel != defaults.DefaultModel {
		t.Errorf("DefaultOpenAIModel = %q, defaults.DefaultModel = %q", DefaultOpenAIModel, defaults.DefaultModel)
	}

	openaiConfig, ok := defaults.DefaultConfig["openai"].(map[string]interface{})
	if !ok {
		t.Fatal("defaults.DefaultConfig[openai] is not a map")
	}
	if got := openaiConfig["model"]; got != DefaultOpenAIModel {
		t.Errorf("default openai model = %v, want %q", got, DefaultOpenAIModel)
	}

	claudeConfig, ok := defaults.DefaultConfig["claude"].(map[string]interface{})
	if !ok {
		t.Fatal("defaults.DefaultConfig[claude] is not a map")
	}
	if got := claudeConfig["model"]; got != DefaultClaudeModel {
		t.Errorf("default claude model = %v, want %q", got, DefaultClaudeModel)
	}
}
