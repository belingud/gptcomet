package llm

import (
	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/pkg/types"
)

// NewProvider creates a new LLM provider instance based on the given provider name and configuration.
// It returns an LLM interface and an error if any occurs during the creation process.
//
// Parameters:
//   - providerName: A string representing the name of the desired LLM provider.
//   - config: A pointer to types.ClientConfig containing the configuration for the provider.
//
// Returns:
//   - LLM: An interface representing the created LLM provider instance.
//   - error: An error if the provider creation fails, or nil if successful.
//
// If the specified provider is not registered, it returns a DefaultLLM instance.
// If the config parameter is nil, it returns an error.
func NewProvider(providerName string, config *types.ClientConfig) (LLM, error) {
	if config == nil {
		return nil, gptErrors.NewValidationError(
			"Invalid Configuration",
			"Configuration object is nil",
			nil,
			[]string{"Ensure valid configuration is provided"},
		)
	}

	constructor, ok := GetProviderConstructor(providerName)
	if !ok {
		return &DefaultLLM{}, nil
	}

	return constructor(config), nil
}

// CreateProvider creates a new provider instance with given config
func CreateProvider(config *types.ClientConfig) (LLM, error) {
	if config == nil {
		return nil, gptErrors.NewValidationError(
			"Invalid Configuration",
			"Configuration object is nil",
			nil,
			[]string{"Ensure valid configuration is provided"},
		)
	}

	providerName := config.Provider
	if providerName == "" {
		providerName = "openai"
	}
	constructor, ok := GetProviderConstructor(providerName)
	if !ok {
		constructor, _ = GetProviderConstructor("openai")
	}

	return constructor(config), nil
}

// init initializes the LLM providers by registering them with their respective constructors.
// This function is automatically called when the package is imported.
//
// It registers the following providers:
// - AI21
// - Azure
// - ChatGLM
// - Claude
// - Cohere
// - Deepseek
// - Gemini
// - Kimi
// - Mistral
// - MiniMax
// - Ollama
// - OpenAI
// - Vertex
// - XAI
// - Sambanova
// - Silicon
// - Tongyi
// - Groq
// - OpenRouter
//
// Each provider is registered with a constructor function that creates a new instance
// of the corresponding LLM implementation.
func init() {
	// AI21
	RegisterProvider("ai21", func(config *types.ClientConfig) LLM {
		return NewAI21LLM(config)
	})

	// Azure
	RegisterProvider("azure", func(config *types.ClientConfig) LLM {
		return NewAzureLLM(config)
	})

	// ChatGLM
	RegisterProvider("chatglm", func(config *types.ClientConfig) LLM {
		return NewChatGLMLLM(config)
	})

	// Claude
	RegisterProvider("claude", func(config *types.ClientConfig) LLM {
		return NewClaudeLLM(config)
	})

	// Cohere
	RegisterProvider("cohere", func(config *types.ClientConfig) LLM {
		return NewCohereLLM(config)
	})

	// Deepseek
	RegisterProvider("deepseek", func(config *types.ClientConfig) LLM {
		return NewDeepSeekLLM(config)
	})

	// Gemini
	RegisterProvider("gemini", func(config *types.ClientConfig) LLM {
		return NewGeminiLLM(config)
	})

	// Groq
	RegisterProvider("groq", func(config *types.ClientConfig) LLM {
		return NewGroqLLM(config)
	})

	// Hunyuan
	RegisterProvider("hunyuan", func(config *types.ClientConfig) LLM {
		return NewHunyuanLLM(config)
	})

	// Kimi
	RegisterProvider("kimi", func(config *types.ClientConfig) LLM {
		return NewKimiLLM(config)
	})

	// LongCat
	RegisterProvider("longcat", func(config *types.ClientConfig) LLM {
		return NewLongCatLLM(config)
	})

	// MiniMax
	RegisterProvider("minimax", func(config *types.ClientConfig) LLM {
		return NewMinimaxLLM(config)
	})

	// Mistral
	RegisterProvider("mistral", func(config *types.ClientConfig) LLM {
		return NewMistralLLM(config)
	})

	// ModelScope
	RegisterProvider("modelscope", func(config *types.ClientConfig) LLM {
		return NewModelScopeLLM(config)
	})

	// Ollama
	RegisterProvider("ollama", func(config *types.ClientConfig) LLM {
		return NewOllamaLLM(config)
	})

	// OpenAI
	RegisterProvider("openai", func(config *types.ClientConfig) LLM {
		return NewOpenAILLM(config)
	})

	// OpenRouter
	RegisterProvider("openrouter", func(config *types.ClientConfig) LLM {
		return NewOpenRouterLLM(config)
	})

	// Sambanova
	RegisterProvider("sambanova", func(config *types.ClientConfig) LLM {
		return NewSambanovaLLM(config)
	})

	// Silicon
	RegisterProvider("silicon", func(config *types.ClientConfig) LLM {
		return NewSiliconLLM(config)
	})

	// Tongyi
	RegisterProvider("tongyi", func(config *types.ClientConfig) LLM {
		return NewTongyiLLM(config)
	})

	// Vertex
	RegisterProvider("vertex", func(config *types.ClientConfig) LLM {
		return NewVertexLLM(config)
	})

	// XAI
	RegisterProvider("xai", func(config *types.ClientConfig) LLM {
		return NewXAILLM(config)
	})

	// Yi
	RegisterProvider("yi", func(config *types.ClientConfig) LLM {
		return NewYiLLM(config)
	})
}
