package testutils

import (
	"context"
	"net/http"

	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
)

// Mock LLM implementation for testing
type MockLLM struct {
	name                  string
	generateCommitMessage func(diff string, prompt string) (string, error)
	translateMessage      func(prompt string, message string, lang string) (string, error)
	makeRequest           func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error)
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
	if m.generateCommitMessage != nil {
		return m.generateCommitMessage(diff, prompt)
	}
	return "Test commit message", nil
}

func (m *MockLLM) TranslateMessage(prompt string, message string, lang string) (string, error) {
	if m.translateMessage != nil {
		return m.translateMessage(prompt, message, lang)
	}
	return message, nil
}

func (m *MockLLM) Name() string {
	return m.name
}
