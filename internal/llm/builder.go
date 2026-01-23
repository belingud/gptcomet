package llm

import (
	"github.com/belingud/gptcomet/pkg/types"
)

// Builder provides helper functions for constructing LLM providers
// with common configuration patterns.

// SetDefaultAPIBase sets the API base URL if it's not already set.
// Returns the modified config.
func SetDefaultAPIBase(config *types.ClientConfig, defaultURL string) {
	if config.APIBase == "" {
		config.APIBase = defaultURL
	}
}

// SetDefaultModel sets the model name if it's not already set.
// Returns the modified config.
func SetDefaultModel(config *types.ClientConfig, defaultModel string) {
	if config.Model == "" {
		config.Model = defaultModel
	}
}

// SetDefaultCompletionPath sets the completion path if it's not already set.
func SetDefaultCompletionPath(config *types.ClientConfig, defaultPath string) {
	if config.CompletionPath == nil {
		config.CompletionPath = &defaultPath
	}
}

// SetDefaultAnswerPath sets the answer path if it's not already set.
func SetDefaultAnswerPath(config *types.ClientConfig, defaultPath string) {
	if config.AnswerPath == "" {
		config.AnswerPath = defaultPath
	}
}

// BuildStandardConfig applies standard defaults to a client config.
// This is useful for OpenAI-compatible providers.
// If completionPath or answerPath are empty, uses OpenAI-compatible defaults.
func BuildStandardConfig(config *types.ClientConfig, apiBase, model, completionPath, answerPath string) {
	SetDefaultAPIBase(config, apiBase)
	SetDefaultModel(config, model)

	// Use OpenAI-compatible defaults if not specified
	if completionPath == "" {
		completionPath = "chat/completions"
	}
	if answerPath == "" {
		answerPath = "choices.0.message.content"
	}

	SetDefaultCompletionPath(config, completionPath)
	SetDefaultAnswerPath(config, answerPath)
}

// BuildStandardConfigSimple is a convenience function that only requires apiBase and model,
// using OpenAI-compatible defaults for completion and answer paths.
func BuildStandardConfigSimple(config *types.ClientConfig, apiBase, model string) {
	BuildStandardConfig(config, apiBase, model, "", "")
}
