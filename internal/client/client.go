package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/pkg/types"
)

// Client represents an LLM client
type Client struct {
	config *types.ClientConfig
	llm    llm.LLM
}

// New creates a new client with the given config
func New(config *types.ClientConfig) *Client {
	var provider llm.LLM
	switch config.Provider {
	case "openai":
		provider = llm.NewOpenAILLM(config)
	case "claude":
		provider = llm.NewClaudeLLM(config)
	case "gemini":
		provider = llm.NewGeminiLLM(config)
	case "mistral":
		provider = llm.NewMistralLLM(config)
	case "xai":
		provider = llm.NewXAILLM(config)
	case "cohere":
		provider = llm.NewCohereLLM(config)
	case "tongyi":
		provider = llm.NewTongyiLLM(config)
	case "deepseek":
		provider = llm.NewDeepSeekLLM(config)
	case "chatglm":
		provider = llm.NewChatGLMLLM(config)
	case "azure":
		provider = llm.NewAzureLLM(config)
	case "vertex":
		provider = llm.NewVertexLLM(config)
	case "kimi":
		provider = llm.NewKimiLLM(config)
	case "ollama":
		provider = llm.NewOllamaLLM(config)
	case "silicon":
		provider = llm.NewSiliconLLM(config)
	case "sambanova":
		provider = llm.NewSambanovaLLM(config)
	case "groq":
		provider = llm.NewGroqLLM(config)
	default:
		// Default to OpenAI if provider is not specified
		provider = llm.NewOpenAILLM(config)
	}

	return &Client{
		config: config,
		llm:    provider,
	}
}

// Chat sends a chat message to the LLM provider
func (c *Client) Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error) {
	client, err := c.getClient()
	if err != nil {
		debug.Printf("❌ Get client failed: %v", err)
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	debug.Printf("🔌 Using proxy: %s", c.config.Proxy)

	content, err := c.llm.MakeRequest(ctx, client, message, history)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	debug.Printf("✅ Request succeeded")
	return &types.CompletionResponse{
		Content: content,
		Raw:     make(map[string]interface{}),
	}, nil
}

// createProxyTransport creates an http.Transport with proxy settings based on the configuration
func (c *Client) createProxyTransport() (*http.Transport, error) {
	debug.Printf("Starting proxy configuration with URL: %s", c.config.Proxy)
	var (
		MaxIdleConns       = 100
		IdleConnTimeout    = 90 * time.Second
		DisableCompression = true
	)

	// Return default transport if no proxy configured
	if c.config.Proxy == "" {
		debug.Printf("No proxy configured, using direct connection")
		return &http.Transport{
			MaxIdleConns:       MaxIdleConns,
			IdleConnTimeout:    IdleConnTimeout,
			DisableCompression: DisableCompression,
		}, nil
	}

	fmt.Printf("Using proxy: %s\n", c.config.Proxy)

	proxyURL, err := url.Parse(c.config.Proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	switch proxyURL.Scheme {
	case "http", "https":
		debug.Printf("Configuring HTTP/HTTPS proxy: %s", proxyURL.String())
		transport := &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			MaxIdleConns:       MaxIdleConns,
			IdleConnTimeout:    IdleConnTimeout,
			DisableCompression: DisableCompression,
		}

		// Add proxy authentication if provided
		if proxyURL.User != nil {
			username := proxyURL.User.Username()
			password, hasPassword := proxyURL.User.Password()

			if hasPassword {
				auth := username + ":" + password
				basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
				transport.ProxyConnectHeader = http.Header{
					"Proxy-Authorization": []string{basicAuth},
				}
				debug.Printf("Added proxy authentication for user: %s", username)
			}
		}
		return transport, nil

	case "socks5":
		debug.Printf("Configuring SOCKS5 proxy: %s", proxyURL.String())

		// Configure SOCKS5 authentication
		var auth *proxy.Auth
		if proxyURL.User != nil {
			auth = &proxy.Auth{
				User: proxyURL.User.Username(),
			}
			if password, ok := proxyURL.User.Password(); ok {
				auth.Password = password
			}
		}

		// Create SOCKS5 dialer
		dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		return &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				debug.Printf("Attempting SOCKS5 connection to: %s", addr)
				return dialer.Dial(network, addr)
			},
			MaxIdleConns:       MaxIdleConns,
			IdleConnTimeout:    IdleConnTimeout,
			DisableCompression: DisableCompression,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
	}
}

// getClient returns an HTTP client configured with proxy settings if specified
func (c *Client) getClient() (*http.Client, error) {
	// Create a transport with proxy if configured
	transport, err := c.createProxyTransport()
	if err != nil {
		debug.Printf("❌ Create proxy transport failed: %v", err)
		return nil, fmt.Errorf("failed to create proxy transport: %w", err)
	}

	// Create a client with the configured transport and timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(c.config.Timeout) * time.Second,
	}

	return client, nil
}

// TranslateMessage translates the given message to the specified language
func (c *Client) TranslateMessage(prompt string, message string, lang string) (string, error) {
	// Format the prompt
	formattedPrompt := fmt.Sprintf(prompt, message, lang)

	// Send the request
	resp, err := c.Chat(context.Background(), formattedPrompt, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// GenerateCommitMessage generates a commit message for the given diff
func (c *Client) GenerateCommitMessage(diff string, prompt string) (string, error) {
	formattedPrompt := fmt.Sprintf(prompt, diff)

	// Send the request
	resp, err := c.Chat(context.Background(), formattedPrompt, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}
