package llm

import (
	"github.com/belingud/gptcomet/pkg/config"
)

// ConfigTemplate provides standard configuration requirement templates
// for common provider types.

// StandardConfigTemplate returns a standard OpenAI-compatible configuration template.
// Most providers use this pattern with different default values.
func StandardConfigTemplate(apiBase, model, apiBasePrompt, modelPrompt string) map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{
		"api_base": {
			DefaultValue:  apiBase,
			PromptMessage: apiBasePrompt,
		},
		"model": {
			DefaultValue:  model,
			PromptMessage: modelPrompt,
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

// OpenAICompatibleTemplate is a convenience function for OpenAI-compatible providers.
func OpenAICompatibleTemplate(providerName, apiBase, defaultModel string) map[string]config.ConfigRequirement {
	return StandardConfigTemplate(
		apiBase,
		defaultModel,
		"Enter "+providerName+" API base URL",
		"Enter model name",
	)
}

// CustomConfigTemplate creates a template with custom fields.
// Use this for providers with non-standard configuration requirements.
func CustomConfigTemplate(fields map[string]config.ConfigRequirement) map[string]config.ConfigRequirement {
	return fields
}
