package testutils

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
)

// Mock LLM implementation for testing
type MockLLM struct {
	name                      string
	GenerateCommitMessageFunc func(diff string, prompt string) (string, error)
	TranslateMessageFunc      func(prompt string, message string, targetLang string) (string, error)
	makeRequest               func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error)
}

func (m *MockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{}
}

func (m *MockLLM) BuildHeaders() map[string]string {
	return map[string]string{}
}

func (m *MockLLM) BuildURL() string {
	return "https://mock.api"
}

func (m *MockLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	return message, nil
}

func (m *MockLLM) ParseResponse(response []byte) (string, error) {
	return string(response), nil
}

func (m *MockLLM) GetUsage(data []byte) (string, error) {
	return "", nil
}

func (m *MockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	if m.makeRequest != nil {
		return m.makeRequest(ctx, client, message, history)
	}
	return "mock response", nil
}

func (m *MockLLM) GenerateCommitMessage(diff string, prompt string) (string, error) {
	if m.GenerateCommitMessageFunc != nil {
		return m.GenerateCommitMessageFunc(diff, prompt)
	}
	return "Test commit message", nil
}

func (m *MockLLM) TranslateMessage(prompt string, message string, targetLang string) (string, error) {
	if m.TranslateMessageFunc != nil {
		return m.TranslateMessageFunc(prompt, message, targetLang)
	}
	return message, nil
}

func (m *MockLLM) Name() string {
	return m.name
}
