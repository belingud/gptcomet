package types

const (
	DefaultAPIBase = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o"
	DefaultRetries = 3
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a chat completion request
type CompletionRequest struct {
	Model            string    `json:"model"`
	Messages         []Message `json:"messages"`
	MaxTokens        *int      `json:"max_tokens,omitempty"`
	Temperature      *float32  `json:"temperature,omitempty"`
	TopP             *float32  `json:"top_p,omitempty"`
	FrequencyPenalty *float32  `json:"frequency_penalty,omitempty"`
}

// CompletionResponse represents a chat completion response
type CompletionResponse map[string]interface{}

// Choice represents a completion choice
type Choice struct {
	Message Message `json:"message"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ClientConfig represents the configuration for an LLM client
type ClientConfig struct {
	APIBase        string            `json:"api_base"`
	APIKey         string            `json:"api_key"`
	Model          string            `json:"model"`
	Provider       string            `json:"provider"`
	Retries        int               `json:"retries"`
	Timeout        int64             `json:"timeout"`
	Proxy          string            `json:"proxy,omitempty"`
	ExtraHeaders   map[string]string `json:"extra_headers,omitempty"`
	CompletionPath string            `json:"completion_path"`
	AnswerPath     string            `json:"answer_path"`
	Debug          bool              `json:"debug,omitempty"`
}
