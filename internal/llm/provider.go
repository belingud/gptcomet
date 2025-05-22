package llm

import (
	"fmt"
	"sort"

	"github.com/belingud/gptcomet/pkg/types"
)

// ProviderConstructor is a function that creates a new LLM instance
type ProviderConstructor func(config *types.ClientConfig) LLM

var (
	// providers Stores all registered LLM providers
	providers = make(map[string]ProviderConstructor)
)

// RegisterProvider registers a new LLM provider constructor
func RegisterProvider(name string, constructor ProviderConstructor) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if constructor == nil {
		return fmt.Errorf("constructor cannot be nil")
	}

	providers[name] = constructor
	return nil
}

// GetProviders returns a list of all registered providers
func GetProviders() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

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
		return nil, fmt.Errorf("config cannot be nil")
	}

	constructor, ok := providers[providerName]
	if !ok {
		return &DefaultLLM{}, nil
	}

	return constructor(config), nil
}

// CreateProvider creates a new provider instance with given config
func CreateProvider(config *types.ClientConfig) (LLM, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	constructor, ok := providers[config.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", config.Provider)
	}

	return constructor(config), nil
}

// init initializes the LLM providers by registering them with their respective constructors.
// This function is automatically called when the package is imported.
//
// It registers the following providers:
// - Azure
// - ChatGLM
// - Claude
// - Cohere
// - Deepseek
// - Gemini
// - Kimi
// - Mistral
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
	// Azure
	RegisterProvider("azure", func(config *types.ClientConfig) LLM {
		return &AzureLLM{}
	})

	// ChatGLM
	RegisterProvider("chatglm", func(config *types.ClientConfig) LLM {
		return &ChatGLMLLM{}
	})

	// Claude
	RegisterProvider("claude", func(config *types.ClientConfig) LLM {
		return &ClaudeLLM{}
	})

	// Cohere
	RegisterProvider("cohere", func(config *types.ClientConfig) LLM {
		return &CohereLLM{}
	})

	// Deepseek
	RegisterProvider("deepseek", func(config *types.ClientConfig) LLM {
		return &DeepSeekLLM{}
	})

	// Gemini
	RegisterProvider("gemini", func(config *types.ClientConfig) LLM {
		return &GeminiLLM{}
	})

	// Kimi
	RegisterProvider("kimi", func(config *types.ClientConfig) LLM {
		return &KimiLLM{}
	})

	// Mistral
	RegisterProvider("mistral", func(config *types.ClientConfig) LLM {
		return &MistralLLM{}
	})

	// Ollama
	RegisterProvider("ollama", func(config *types.ClientConfig) LLM {
		return &OllamaLLM{}
	})

	// OpenAI
	RegisterProvider("openai", func(config *types.ClientConfig) LLM {
		return &OpenAILLM{}
	})

	// Sambanova
	RegisterProvider("sambanova", func(config *types.ClientConfig) LLM {
		return &SambanovaLLM{}
	})

	// Silicon
	RegisterProvider("silicon", func(config *types.ClientConfig) LLM {
		return &SiliconLLM{}
	})

	// Tongyi
	RegisterProvider("tongyi", func(config *types.ClientConfig) LLM {
		return &TongyiLLM{}
	})

	// Vertex
	RegisterProvider("vertex", func(config *types.ClientConfig) LLM {
		return &VertexLLM{}
	})

	// XAI
	RegisterProvider("xai", func(config *types.ClientConfig) LLM {
		return &XAILLM{}
	})

	// Groq
	RegisterProvider("groq", func(config *types.ClientConfig) LLM {
		return &GroqLLM{}
	})

	// OpenRouter
	RegisterProvider("openrouter", func(config *types.ClientConfig) LLM {
		return &OpenRouterLLM{}
	})

	// AI21
	RegisterProvider("ai21", func(config *types.ClientConfig) LLM {
		return &AI21LLM{}
	})

	// ModelScope
	RegisterProvider("modelscope", func(config *types.ClientConfig) LLM {
		return NewModelScopeLLM(config)
	})

	// Hunyuan
	RegisterProvider("hunyuan", func(config *types.ClientConfig) LLM {
		return NewHunyuanLLM(config)
	})
}
