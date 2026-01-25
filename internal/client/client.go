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
	"reflect"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"github.com/belingud/gptcomet/internal/constants"
	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/internal/logger"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/tidwall/gjson"
)

// sanitizeURLForLogging returns a sanitized version of the URL for logging purposes.
// For gemini/vertex providers, it strips query parameters to avoid logging sensitive data like API keys.
// For other providers, it returns the URL as-is.
func sanitizeURLForLogging(provider string, rawURL string) string {
	if provider != "gemini" && provider != "vertex" {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		// If parsing fails, return the original URL as a fallback
		return rawURL
	}

	// Remove query parameters and fragment
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String()
}

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
		logger.Error("Get client failed: %v", err)
		return nil, err
	}

	logger.Debug("Using proxy: %s", c.config.Proxy)

	var lastErr error
	maxRetries := c.config.Retries

	for i := 0; i < maxRetries; i++ {
		content, err := c.llm.MakeRequest(ctx, client, message, false)
		if err == nil {
			logger.Debug("Request succeeded after %d retries", i)
			return &types.CompletionResponse{
				Content: content,
				Raw:     make(map[string]interface{}),
			}, nil
		}

		lastErr = err
		logger.Warn("Request failed (attempt %d/%d): %v", i+1, maxRetries, err)

		// Don't retry on context cancellation or deadline exceeded
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		// Exponential backoff with jitter
		if i < maxRetries {
			delay := time.Duration(float64(constants.BaseRetryDelay) * math.Pow(2, float64(i)))
			jitter := time.Duration(rand.Int63n(int64(float64(delay) * constants.MaxJitterPercent)))
			sleepDuration := delay + jitter

			logger.Info("Retrying in %v...", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}

	return nil, gptErrors.RequestRetryError(maxRetries, lastErr)
}

// createProxyTransport creates an http.Transport with proxy settings based on the configuration
func (c *Client) createProxyTransport() (*http.Transport, error) {
	logger.Debug("Starting proxy configuration with URL: %s", c.config.Proxy)

	// Return default transport if no proxy configured
	if c.config.Proxy == "" {
		logger.Debug("No proxy configured, using direct connection")
		return &http.Transport{
			MaxIdleConns:       constants.MaxIdleConns,
			IdleConnTimeout:    constants.IdleConnTimeout,
			DisableCompression: constants.DisableCompression,
		}, nil
	}

	logger.Info("Using proxy: %s", c.config.Proxy)

	proxyURL, err := url.Parse(c.config.Proxy)
	if err != nil {
		return nil, gptErrors.ProxyURLParseError(err)
	}

	switch proxyURL.Scheme {
	case constants.ProxySchemeHTTP, constants.ProxySchemeHTTPS:
		logger.Debug("Configuring HTTP/HTTPS proxy: %s", proxyURL.String())
		transport := &http.Transport{
			Proxy:              http.ProxyURL(proxyURL),
			MaxIdleConns:       constants.MaxIdleConns,
			IdleConnTimeout:    constants.IdleConnTimeout,
			DisableCompression: constants.DisableCompression,
		}

		// Add proxy authentication if provided
		if proxyURL.User != nil {
			username := proxyURL.User.Username()
			password, hasPassword := proxyURL.User.Password()

			if hasPassword {
				auth := username + ":" + password
				basicAuth := constants.BasicPrefix + base64.StdEncoding.EncodeToString([]byte(auth))
				transport.ProxyConnectHeader = http.Header{
					constants.HeaderProxyAuthorization: []string{basicAuth},
				}
				logger.Debug("Added proxy authentication for user: %s", username)
			}
		}
		return transport, nil

	case constants.ProxySchemeSocks5:
		logger.Debug("Configuring SOCKS5 proxy: %s", proxyURL.String())

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
				logger.Debug("Attempting SOCKS5 connection to: %s", addr)
				return dialer.Dial(network, addr)
			},
			MaxIdleConns:       constants.MaxIdleConns,
			IdleConnTimeout:    constants.IdleConnTimeout,
			DisableCompression: constants.DisableCompression,
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
		logger.Error("Create proxy transport failed: %v", err)
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

	// Check if provider has a custom streaming URL method (e.g., Gemini)
	// If it does, use that URL instead of adding "stream" to the payload
	url := ""
	if buildStreamURLMethod := reflect.ValueOf(c.llm).MethodByName("BuildStreamURL"); buildStreamURLMethod.IsValid() {
		// Provider has custom streaming URL (e.g., Gemini uses :streamGenerateContent)
		results := buildStreamURLMethod.Call(nil)
		if len(results) > 0 && results[0].Kind() == reflect.String {
			url = results[0].String()
			logger.Debug("Using provider's custom streaming URL")
		}
	} else {
		// Provider doesn't have custom streaming URL, use standard URL and add "stream" param
		url = c.llm.BuildURL()
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			payloadMap["stream"] = true
		}
	}

	// Marshal the payload
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return gptErrors.RequestMarshalingError(err)
	}

	logger.Debug("Request URL: %s", sanitizeURLForLogging(c.config.Provider, url))
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

	logger.Info("Sending streaming request to %s...", c.llm.Name())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return gptErrors.RequestExecutionError(err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != constants.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return gptErrors.APIStatusError(resp.StatusCode, string(respBody), nil)
	}

	logger.Debug("Request succeeded, processing streaming response")

	// Process the streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and SSE comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// Check for data prefix - handle both "data: " and "data:" formats (SSE)
		// If line doesn't start with "data:", treat it as NDJSON (Ollama format)
		var data string
		if strings.HasPrefix(line, constants.SSEDataPrefix) {
			// "data: " format (with space)
			data = strings.TrimPrefix(line, constants.SSEDataPrefix)
		} else if strings.HasPrefix(line, "data:") {
			// "data:" format (without space) - also valid per SSE spec
			data = strings.TrimPrefix(line, "data:")
		} else {
			// Not SSE format, treat as direct JSON (NDJSON format like Ollama)
			data = line
		}

		// Check for [DONE] message
		if data == constants.SSEDone {
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
			continue
		}

		// Check for Ollama-specific "done" flag (NDJSON format)
		if done, ok := streamResp["done"].(bool); ok && done {
			logger.Debug("Ollama stream finished (done: true)")
			// Send a newline before breaking to avoid % prompt appearing right after output
			callback(&types.CompletionResponse{
				Content: "\n",
				Raw:     make(map[string]interface{}),
			})
			break
		}

		// Extract the content using the provider's configured stream answer path
		streamAnswerPath := c.config.StreamAnswerPath
		if streamAnswerPath == "" {
			// Fallback: try to convert non-streaming path to streaming path
			// For OpenAI-compatible: "choices.0.message.content" -> "choices.0.delta.content"
			streamAnswerPath = c.config.AnswerPath
			if strings.Contains(streamAnswerPath, "message") {
				streamAnswerPath = strings.Replace(streamAnswerPath, "message", "delta", 1)
			} else {
				// If conversion fails, use OpenAI default
				streamAnswerPath = "choices.0.delta.content"
			}
		}

		// Use gjson to extract the content
		content := gjson.GetBytes([]byte(data), streamAnswerPath).String()

		// For Ollama: if response is empty, try thinking field (some models output thinking process)
		if content == "" && c.config.Provider == "ollama" {
			content = gjson.GetBytes([]byte(data), "thinking").String()
		}

		// Skip empty content
		if content == "" {
			logger.Debug("Empty content from stream chunk")
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

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return gptErrors.WrapError(err, "Response Reading Failed", "Error occurred while reading streaming response")
	}

	return nil
}
