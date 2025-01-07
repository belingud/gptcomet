package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"github.com/belingud/go-gptcomet/internal/debug"
	"github.com/belingud/go-gptcomet/internal/llm"
	"github.com/belingud/go-gptcomet/pkg/types"
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
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	content, err := c.llm.MakeRequest(ctx, client, message, history)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return &types.CompletionResponse{
		Content: content,
		Raw:     make(map[string]interface{}),
	}, nil
}

// createProxyTransport creates an http.Transport with proxy settings based on the configuration
func (c *Client) createProxyTransport() (*http.Transport, error) {
	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: false}, // default verification
	}

	if c.config.Proxy != "" {
		debug.Printf("Using proxy: %s", c.config.Proxy)
		proxyURL, err := url.Parse(c.config.Proxy)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}

		switch proxyURL.Scheme {
		case "http", "https":
			transport.Proxy = http.ProxyURL(proxyURL)
			debug.Printf("Using HTTP/HTTPS proxy: %s", proxyURL.String())
		case "socks5":
			auth := &proxy.Auth{}
			if proxyURL.User != nil {
				auth.User = proxyURL.User.Username()
				if password, ok := proxyURL.User.Password(); ok {
					auth.Password = password
				}
			}
			dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
			}
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			}
			debug.Printf("Using SOCKS5 proxy: %s", proxyURL.String())
		default:
			return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
		}

		// add proxy authentication
		if proxyURL.User != nil {
			transport.ProxyConnectHeader = http.Header{}
			auth := proxyURL.User.String()
			basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
			transport.ProxyConnectHeader.Add("Proxy-Authorization", "Basic "+basicAuth)
			debug.Printf("Added proxy authentication")
		}
	}

	return transport, nil
}

// getClient returns an HTTP client configured with proxy settings if specified
func (c *Client) getClient() (*http.Client, error) {
	// Create a transport with proxy if configured
	transport, err := c.createProxyTransport()
	if err != nil {
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

// GenerateCodeExplanation generates an explanation for the given code in the specified language
func (c *Client) GenerateCodeExplanation(message, lang string) (string, error) {
	const prompt = "Explain the following %s code:\n\n%s"
	formattedPrompt := fmt.Sprintf(prompt, lang, message)

	// Send the request
	resp, err := c.Chat(context.Background(), formattedPrompt, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// Stream sends a chat message to the LLM provider and streams the response
func (c *Client) Stream(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	content, err := c.llm.MakeRequest(ctx, client, message, history)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return &types.CompletionResponse{
		Content: content,
		Raw:     make(map[string]interface{}),
	}, nil
}
