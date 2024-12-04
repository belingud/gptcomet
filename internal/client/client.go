package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/pkg/types"
)

// Client represents an LLM client
type Client struct {
	config *types.ClientConfig
}

// New creates a new LLM client
func New(config *types.ClientConfig) *Client {
	return &Client{
		config: config,
	}
}

// RawChat sends a chat completion request and returns the raw JSON response
func (c *Client) RawChat(messages []types.Message) (string, error) {
	debug.Printf("Discovered model `%s` with provider `%s`.", c.config.Model, c.config.Provider)

	req := &types.CompletionRequest{
		Model:    c.config.Model,
		Messages: messages,
	}

	var jsonStr string
	var err error

	for i := 1; i <= c.config.Retries; i++ {
		jsonStr, err = c.sendRawRequest(req)
		if err == nil {
			break
		}
		fmt.Printf(" Retrying (%d/%d) in %d seconds...\n", i, c.config.Retries, i)
		time.Sleep(time.Duration(i) * time.Second)
	}

	if err != nil {
		return "", fmt.Errorf("failed after %d retries: %w", c.config.Retries, err)
	}

	if jsonStr == "" {
		return "", fmt.Errorf("empty response")
	}

	return jsonStr, nil
}

// Chat sends a chat completion request and returns the processed response
func (c *Client) Chat(messages []types.Message) (*types.CompletionResponse, error) {
	jsonStr, err := c.RawChat(messages)
	if err != nil {
		return nil, err
	}

	var result types.CompletionResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// sendRawRequest sends a completion request to the LLM provider and returns the raw JSON response
func (c *Client) sendRawRequest(req *types.CompletionRequest) (string, error) {
	// Set provider-specific parameters
	if c.config.MaxTokens > 0 {
		req.MaxTokens = &c.config.MaxTokens
	}
	if c.config.Temperature > 0 {
		req.Temperature = &c.config.Temperature
	}
	if c.config.TopP > 0 {
		req.TopP = &c.config.TopP
	}
	if c.config.FrequencyPenalty > 0 {
		req.FrequencyPenalty = &c.config.FrequencyPenalty
	}

	// Marshal the request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the request URL
	u, err := url.Parse(c.config.APIBase)
	if err != nil {
		return "", fmt.Errorf("failed to parse API base: %w", err)
	}
	u.Path = path.Join(u.Path, c.config.CompletionPath)

	// Create the HTTP request
	httpReq, err := http.NewRequest("POST", u.String(), bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	// Add any extra headers
	if c.config.ExtraHeaders != nil {
		for k, v := range c.config.ExtraHeaders {
			httpReq.Header.Set(k, v)
		}
	}

	// Create a transport with proxy if configured
	transport := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
	}

	if c.config.Proxy != "" {
		proxyURL, err := url.Parse(c.config.Proxy)
		if err != nil {
			return "", fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// Create a client with the configured transport and timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(c.config.Timeout) * time.Second,
	}

	// Send the request
	debug.Printf("Sending request to %s", u.String())
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	debug.Printf("Response: %s", string(body))
	return string(body), nil
}

// getAnswerPath returns the configured answer path or the default value
func (c *Client) getAnswerPath() string {
	if c.config.AnswerPath == "" {
		return "choices.0.message.content"
	}
	return c.config.AnswerPath
}

// TranslateMessage translates the given message to the specified language
func (c *Client) TranslateMessage(prompt string, message string, lang string) (string, error) {
	fmt.Printf("Translating message into %s: %s\n", lang, message)
	prompt = strings.Replace(prompt, "{{ placeholder }}", message, 1)
	prompt = strings.Replace(prompt, "{{ output_language }}", config.OutputLanguageMap[lang], 1)
	messages := []types.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	jsonStr, err := c.RawChat(messages)
	if err != nil {
		return "", err
	}

	// Use gjson to extract the answer using the configured answer_path
	result := gjson.Get(jsonStr, c.getAnswerPath())
	if !result.Exists() {
		return "", fmt.Errorf("answer path '%s' not found in response", c.getAnswerPath())
	}

	return result.String(), nil
}

// GenerateCommitMessage generates a commit message for the given diff
func (c *Client) GenerateCommitMessage(diff string, prompt string) (string, error) {
	prompt = strings.Replace(prompt, "{{ placeholder }}", diff, 1)
	messages := []types.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	jsonStr, err := c.RawChat(messages)
	if err != nil {
		return "", err
	}

	// Use gjson to extract the answer using the configured answer_path
	result := gjson.Get(jsonStr, c.getAnswerPath())
	if !result.Exists() {
		return "", fmt.Errorf("answer path '%s' not found in response", c.getAnswerPath())
	}

	return result.String(), nil
}
