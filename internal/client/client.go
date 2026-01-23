package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"github.com/belingud/gptcomet/internal/debug"
	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/pkg/types"
)

type ClientInterface interface {
	Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error)
	TranslateMessage(prompt string, message string, lang string) (string, error)
	GenerateCommitMessage(diff string, prompt string) (string, error)
	GenerateReviewComment(diff string, prompt string) (string, error)
	GenerateReviewCommentStream(diff string, prompt string, callback func(string) error) error
}

// Client represents an LLM client
type Client struct {
	config *types.ClientConfig
	llm    llm.LLM
}

// New creates a new client with the given config
func New(config *types.ClientConfig) (*Client, error) {
	if config == nil {
		return nil, gptErrors.NewValidationError(
			"Invalid Configuration",
			"Client configuration is nil",
			nil,
			[]string{"Ensure valid client configuration is provided"},
		)
	}

	provider, err := llm.CreateProvider(config)
	if err != nil {
		return nil, gptErrors.ProviderCreationError(config.Provider, err)
	}

	return &Client{
		config: config,
		llm:    provider,
	}, nil
}

// Chat sends a chat message to the LLM provider with retry logic
func (c *Client) Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error) {
	client, err := c.getClient()
	if err != nil {
		debug.Printf("❌ Get client failed: %v", err)
		return nil, err
	}

	debug.Printf("🔌 Using proxy: %s", c.config.Proxy)

	var lastErr error
	baseDelay := 500 * time.Millisecond
	maxRetries := c.config.Retries

	for i := 0; i < maxRetries; i++ {
		content, err := c.llm.MakeRequest(ctx, client, message, false)
		if err == nil {
			debug.Printf("✅ Request succeeded after %d retries", i)
			return &types.CompletionResponse{
				Content: content,
				Raw:     make(map[string]interface{}),
			}, nil
		}

		lastErr = err
		fmt.Printf("⚠️ Request failed (attempt %d/%d): %v\n", i+1, maxRetries, err)

		// Don't retry on context cancellation or deadline exceeded
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		// Exponential backoff with jitter
		if i < maxRetries {
			delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(i)))
			jitter := time.Duration(rand.Int63n(int64(delay / 2))) // Add up to 50% jitter
			sleepDuration := delay + jitter

			fmt.Printf("⏳ Retrying in %v...\n", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	return nil, gptErrors.RequestRetryError(maxRetries, lastErr)
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
		return nil, gptErrors.ProxyURLParseError(err)
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
			return nil, gptErrors.ProxyConfigurationError(err)
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
		return nil, gptErrors.UnsupportedProxySchemeError(proxyURL.Scheme)
	}
}

// getClient returns an HTTP client configured with proxy settings if specified
func (c *Client) getClient() (*http.Client, error) {
	// Create a transport with proxy if configured
	transport, err := c.createProxyTransport()
	if err != nil {
		debug.Printf("❌ Create proxy transport failed: %v", err)
		return nil, err
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
	formattedPrompt := strings.Replace(prompt, "{{ placeholder }}", diff, 1)

	// Send the request
	resp, err := c.Chat(context.Background(), formattedPrompt, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// GenerateReviewComment generates a review comment for the given diff
func (c *Client) GenerateReviewComment(diff string, prompt string) (string, error) {
	formattedPrompt := strings.Replace(prompt, "{{ placeholder }}", diff, 1)

	// Send the request
	resp, err := c.Chat(context.Background(), formattedPrompt, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// GenerateReviewCommentStream generates a review comment for the given diff
func (c *Client) GenerateReviewCommentStream(diff string, prompt string, callback func(string) error) error {
	formattedPrompt := strings.Replace(prompt, "{{ placeholder }}", diff, 1)

	// Send the request
	return c.Stream(context.Background(), formattedPrompt, func(resp *types.CompletionResponse) error {
		return callback(resp.Content)
	})
}

// Stream sends a streaming request to the LLM provider and processes the response
// using the provided callback function.
//
// Parameters:
//   - ctx: the context for the request, used for cancellation and timeouts
//   - message: the message to send to the LLM provider
//   - callback: a function that processes the CompletionResponse received from the LLM provider
//
// Returns an error if the client cannot be obtained, the request fails, or the callback function
// returns an error.
func (c *Client) Stream(ctx context.Context, message string, callback func(*types.CompletionResponse) error) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	// Format the message for the provider
	payload, err := c.llm.FormatMessages(message)
	if err != nil {
		return gptErrors.MessageFormattingError(err)
	}

	// Set stream to true for streaming response
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		payloadMap["stream"] = true
	}

	// Marshal the payload
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return gptErrors.RequestMarshalingError(err)
	}

	// Build the URL and headers
	url := c.llm.BuildURL()
	debug.Printf("🔗 URL: %s", url)
	headers := c.llm.BuildHeaders()

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return gptErrors.RequestCreationError(err)
	}

	// Set the headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	fmt.Printf("📤 Sending streaming request to %s...\n", c.llm.Name())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return gptErrors.RequestExecutionError(err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return gptErrors.APIStatusError(resp.StatusCode, string(respBody), nil)
	}

	debug.Printf("✅ Request succeeded, processing streaming response")

	// Process the streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and SSE comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Check for data prefix
		if strings.HasPrefix(line, "data: ") {
			// Extract the data
			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] message
			if data == "[DONE]" {
				// Send a newline before breaking to avoid % prompt appearing right after output
				callback(&types.CompletionResponse{
					Content: "\n",
					Raw:     make(map[string]interface{}),
				})
				break
			}

			// Parse the JSON data
			var streamResp map[string]interface{}
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				debug.Printf("Error parsing SSE data: %v", err)
				continue
			}

			// Extract the content using the provider's answer path
			content := ""
			if choices, ok := streamResp["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if c, ok := delta["content"].(string); ok {
							content = c
						}
					}
				}
			}

			// Skip empty content
			if content == "" {
				continue
			}

			// Call the callback with the content
			err := callback(&types.CompletionResponse{
				Content: content,
				Raw:     streamResp,
			})
			if err != nil {
				return gptErrors.CallbackError(err)
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return gptErrors.WrapError(err, "Response Reading Failed", "Error occurred while reading streaming response")
	}

	return nil
}
