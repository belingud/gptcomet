package llm

import (
	"fmt"
	"sort"

	"github.com/belingud/go-gptcomet/pkg/types"
)

// ProviderConstructor 是创建 LLM 实例的函数类型
type ProviderConstructor func(config *types.ClientConfig) LLM

var (
	// providers 存储所有注册的 provider
	providers = make(map[string]ProviderConstructor)
)

// RegisterProvider 注册一个新的 provider
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

// GetProviders 返回所有已注册的 provider 名称
func GetProviders() []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// NewProvider 根据名称创建对应的 LLM 实例
func NewProvider(name string, config *types.ClientConfig) (LLM, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	constructor, ok := providers[name]
	if !ok {
		return &DefaultLLM{}, nil
	}
	return constructor(config), nil
}

// 在 init 函数中注册所有 provider
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
}
