package llm

import (
	"testing"

	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaultAPIBase(t *testing.T) {
	tests := []struct {
		name        string
		config      *types.ClientConfig
		defaultURL  string
		expectedURL string
	}{
		{
			name:        "set default when APIBase is empty",
			config:      &types.ClientConfig{},
			defaultURL:  "https://api.example.com/v1",
			expectedURL: "https://api.example.com/v1",
		},
		{
			name: "keep existing APIBase when already set",
			config: &types.ClientConfig{
				APIBase: "https://existing.api.com",
			},
			defaultURL:  "https://api.example.com/v1",
			expectedURL: "https://existing.api.com",
		},
		{
			name:        "handle empty default URL",
			config:      &types.ClientConfig{},
			defaultURL:  "",
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultAPIBase(tt.config, tt.defaultURL)
			assert.Equal(t, tt.expectedURL, tt.config.APIBase)
		})
	}
}

func TestSetDefaultModel(t *testing.T) {
	tests := []struct {
		name          string
		config        *types.ClientConfig
		defaultModel  string
		expectedModel string
	}{
		{
			name:          "set default when Model is empty",
			config:        &types.ClientConfig{},
			defaultModel:  "gpt-4",
			expectedModel: "gpt-4",
		},
		{
			name: "keep existing Model when already set",
			config: &types.ClientConfig{
				Model: "existing-model",
			},
			defaultModel:  "gpt-4",
			expectedModel: "existing-model",
		},
		{
			name:          "handle empty default model",
			config:        &types.ClientConfig{},
			defaultModel:  "",
			expectedModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultModel(tt.config, tt.defaultModel)
			assert.Equal(t, tt.expectedModel, tt.config.Model)
		})
	}
}

func TestSetDefaultCompletionPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.ClientConfig
		defaultPath  string
		expectedPath string
	}{
		{
			name:         "set default when CompletionPath is nil",
			config:       &types.ClientConfig{},
			defaultPath:  "chat/completions",
			expectedPath: "chat/completions",
		},
		{
			name: "keep existing CompletionPath when already set",
			config: &types.ClientConfig{
				CompletionPath: func() *string { s := "existing/path"; return &s }(),
			},
			defaultPath:  "chat/completions",
			expectedPath: "existing/path",
		},
		{
			name:         "handle empty default path",
			config:       &types.ClientConfig{},
			defaultPath:  "",
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultCompletionPath(tt.config, tt.defaultPath)
			if tt.config.CompletionPath == nil {
				assert.Equal(t, tt.expectedPath, "")
			} else {
				assert.Equal(t, tt.expectedPath, *tt.config.CompletionPath)
			}
		})
	}
}

func TestSetDefaultAnswerPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *types.ClientConfig
		defaultPath  string
		expectedPath string
	}{
		{
			name:         "set default when AnswerPath is empty",
			config:       &types.ClientConfig{},
			defaultPath:  "choices.0.message.content",
			expectedPath: "choices.0.message.content",
		},
		{
			name: "keep existing AnswerPath when already set",
			config: &types.ClientConfig{
				AnswerPath: "existing.path",
			},
			defaultPath:  "choices.0.message.content",
			expectedPath: "existing.path",
		},
		{
			name:         "handle empty default path",
			config:       &types.ClientConfig{},
			defaultPath:  "",
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultAnswerPath(tt.config, tt.defaultPath)
			assert.Equal(t, tt.expectedPath, tt.config.AnswerPath)
		})
	}
}

func TestBuildStandardConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *types.ClientConfig
		apiBase        string
		model          string
		completionPath string
		answerPath     string
		expected       types.ClientConfig
	}{
		{
			name:           "full config with all parameters",
			config:         &types.ClientConfig{},
			apiBase:        "https://api.example.com/v1",
			model:          "gpt-4",
			completionPath: "v1/chat/completions",
			answerPath:     "data.0.content",
			expected: types.ClientConfig{
				APIBase:        "https://api.example.com/v1",
				Model:          "gpt-4",
				CompletionPath: func() *string { s := "v1/chat/completions"; return &s }(),
				AnswerPath:     "data.0.content",
			},
		},
		{
			name:           "empty completion and answer paths use OpenAI defaults",
			config:         &types.ClientConfig{},
			apiBase:        "https://api.openai.com/v1",
			model:          "gpt-4o",
			completionPath: "",
			answerPath:     "",
			expected: types.ClientConfig{
				APIBase:        "https://api.openai.com/v1",
				Model:          "gpt-4o",
				CompletionPath: func() *string { s := "chat/completions"; return &s }(),
				AnswerPath:     "choices.0.message.content",
			},
		},
		{
			name: "keep existing config values",
			config: &types.ClientConfig{
				APIBase:        "https://existing.api.com",
				Model:          "existing-model",
				CompletionPath: func() *string { s := "existing/path"; return &s }(),
				AnswerPath:     "existing.answer",
			},
			apiBase:        "https://api.example.com/v1",
			model:          "gpt-4",
			completionPath: "v1/chat/completions",
			answerPath:     "data.0.content",
			expected: types.ClientConfig{
				APIBase:        "https://existing.api.com",
				Model:          "existing-model",
				CompletionPath: func() *string { s := "existing/path"; return &s }(),
				AnswerPath:     "existing.answer",
			},
		},
		{
			name:           "partial config - only apiBase and model set",
			config:         &types.ClientConfig{},
			apiBase:        "https://api.example.com/v1",
			model:          "gpt-4",
			completionPath: "",
			answerPath:     "",
			expected: types.ClientConfig{
				APIBase:        "https://api.example.com/v1",
				Model:          "gpt-4",
				CompletionPath: func() *string { s := "chat/completions"; return &s }(),
				AnswerPath:     "choices.0.message.content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuildStandardConfig(tt.config, tt.apiBase, tt.model, tt.completionPath, tt.answerPath)
			assert.Equal(t, tt.expected.APIBase, tt.config.APIBase)
			assert.Equal(t, tt.expected.Model, tt.config.Model)

			if tt.expected.CompletionPath == nil {
				assert.Nil(t, tt.config.CompletionPath)
			} else {
				assert.NotNil(t, tt.config.CompletionPath)
				assert.Equal(t, *tt.expected.CompletionPath, *tt.config.CompletionPath)
			}

			assert.Equal(t, tt.expected.AnswerPath, tt.config.AnswerPath)
		})
	}
}

func TestBuildStandardConfigSimple(t *testing.T) {
	tests := []struct {
		name     string
		config   *types.ClientConfig
		apiBase  string
		model    string
		expected types.ClientConfig
	}{
		{
			name:    "simple config with OpenAI defaults",
			config:  &types.ClientConfig{},
			apiBase: "https://api.openai.com/v1",
			model:   "gpt-4o",
			expected: types.ClientConfig{
				APIBase:        "https://api.openai.com/v1",
				Model:          "gpt-4o",
				CompletionPath: func() *string { s := "chat/completions"; return &s }(),
				AnswerPath:     "choices.0.message.content",
			},
		},
		{
			name: "keep existing values",
			config: &types.ClientConfig{
				APIBase:        "https://existing.api.com",
				Model:          "existing-model",
				CompletionPath: func() *string { s := "existing/path"; return &s }(),
				AnswerPath:     "existing.answer",
			},
			apiBase: "https://api.example.com/v1",
			model:   "gpt-4",
			expected: types.ClientConfig{
				APIBase:        "https://existing.api.com",
				Model:          "existing-model",
				CompletionPath: func() *string { s := "existing/path"; return &s }(),
				AnswerPath:     "existing.answer",
			},
		},
		{
			name:    "custom provider with simple config",
			config:  &types.ClientConfig{},
			apiBase: "https://custom.provider.com/api",
			model:   "custom-model",
			expected: types.ClientConfig{
				APIBase:        "https://custom.provider.com/api",
				Model:          "custom-model",
				CompletionPath: func() *string { s := "chat/completions"; return &s }(),
				AnswerPath:     "choices.0.message.content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuildStandardConfigSimple(tt.config, tt.apiBase, tt.model)
			assert.Equal(t, tt.expected.APIBase, tt.config.APIBase)
			assert.Equal(t, tt.expected.Model, tt.config.Model)

			if tt.expected.CompletionPath == nil {
				assert.Nil(t, tt.config.CompletionPath)
			} else {
				assert.NotNil(t, tt.config.CompletionPath)
				assert.Equal(t, *tt.expected.CompletionPath, *tt.config.CompletionPath)
			}

			assert.Equal(t, tt.expected.AnswerPath, tt.config.AnswerPath)
		})
	}
}

func TestBuilderIntegration(t *testing.T) {
	t.Run("multiple setters on same config", func(t *testing.T) {
		config := &types.ClientConfig{}

		SetDefaultAPIBase(config, "https://api.example.com/v1")
		SetDefaultModel(config, "gpt-4")
		SetDefaultCompletionPath(config, "v1/chat/completions")
		SetDefaultAnswerPath(config, "data.0.content")

		assert.Equal(t, "https://api.example.com/v1", config.APIBase)
		assert.Equal(t, "gpt-4", config.Model)
		assert.Equal(t, "v1/chat/completions", *config.CompletionPath)
		assert.Equal(t, "data.0.content", config.AnswerPath)
	})

	t.Run("BuildStandardConfig vs BuildStandardConfigSimple", func(t *testing.T) {
		config1 := &types.ClientConfig{}
		config2 := &types.ClientConfig{}

		BuildStandardConfigSimple(config1, "https://api.openai.com/v1", "gpt-4o")
		BuildStandardConfig(config2, "https://api.openai.com/v1", "gpt-4o", "", "")

		assert.Equal(t, config1.APIBase, config2.APIBase)
		assert.Equal(t, config1.Model, config2.Model)
		assert.Equal(t, *config1.CompletionPath, *config2.CompletionPath)
		assert.Equal(t, config1.AnswerPath, config2.AnswerPath)
	})
}
