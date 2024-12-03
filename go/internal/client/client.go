package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"gptcomet/pkg/types"
)

// Client represents an LLM client
type Client struct {
	config     *types.ClientConfig
	httpClient *http.Client
}

// New creates a new LLM client
func New(config *types.ClientConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// RawChat sends a chat completion request and returns the raw JSON response
func (c *Client) RawChat(messages []types.Message) (string, error) {
	fmt.Println(" Hang tight, I'm cooking up something good!")
	req := &types.CompletionRequest{
		Model:    c.config.Model,
		Messages: messages,
	}

	var jsonStr string
	var err error

	for i := 0; i <= c.config.Retries; i++ {
		jsonStr, err = c.sendRawRequest(req)
		if err == nil {
			break
		}
		fmt.Printf(" Retrying (%d/%d) in %d seconds...\n", i+1, c.config.Retries, i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
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
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the request URL
	u, err := url.Parse(c.config.APIBase)
	if err != nil {
		return "", fmt.Errorf("failed to parse API base: %w", err)
	}
	u.Path = c.config.CompletionPath

	// Create the HTTP request
	httpReq, err := http.NewRequest("POST", u.String(), bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	// Add any extra headers
	if c.config.ExtraHeaders != nil {
		for k, v := range c.config.ExtraHeaders {
			httpReq.Header.Set(k, v)
		}
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

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
	prompt = strings.Replace(prompt, "{{ placeholder }}", message, 1)
	prompt = strings.Replace(prompt, "{{ output_language }}", lang, 1)
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
