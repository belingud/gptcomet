package types

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
	Temperature      *float64  `json:"temperature,omitempty"`
	TopP             *float64  `json:"top_p,omitempty"`
	FrequencyPenalty *float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64  `json:"presence_penalty,omitempty"`
}

// CompletionResponse represents a chat completion response
type CompletionResponse struct {
	Content string                 `json:"content"`
	Raw     map[string]interface{} `json:"raw"`
}

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
	APIBase           string            `json:"api_base"`
	APIKey            string            `json:"api_key,omitempty"`
	Model             string            `json:"model"`
	CompletionPath    *string           `json:"completion_path,omitempty"`
	AnswerPath        string            `json:"answer_path,omitempty"`
	MaxTokens         int               `json:"max_tokens"`
	Temperature       float64           `json:"temperature"`
	TopP              float64           `json:"top_p"`
	TopK              int               `json:"top_k,omitempty"`              // Ollama top k
	RepetitionPenalty float64           `json:"repetition_penalty,omitempty"` // Ollama repetition penalty
	Seed              int               `json:"seed,omitempty"`               // Ollama seed
	NumGPU            int               `json:"num_gpu,omitempty"`            // Ollama number of GPUs
	MainGPU           int               `json:"main_gpu,omitempty"`           // Ollama main GPU index
	FrequencyPenalty  float64           `json:"frequency_penalty"`
	PresencePenalty   float64           `json:"presence_penalty"`
	AnthropicVersion  string            `json:"anthropic_version,omitempty"` // Anthropic API version
	APIVersion        string            `json:"api_version,omitempty"`       // Azure OpenAI API version
	DeploymentName    string            `json:"deployment_name,omitempty"`   // Azure OpenAI deployment name
	Debug             bool              `json:"debug,omitempty"`
	ExtraHeaders      map[string]string `json:"extra_headers,omitempty"`
	Proxy             string            `json:"proxy,omitempty"`
	Retries           int               `json:"retries"`
	Timeout           int64             `json:"timeout"`
	Provider          string            `json:"provider"`
	ProjectID         string            `json:"project_id,omitempty"` // Vertex AI project ID
	Location          string            `json:"location,omitempty"`   // Vertex AI location
}
