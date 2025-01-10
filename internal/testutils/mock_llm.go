package testutils

import (
	"context"
	"net/http"

	"github.com/belingud/gptcomet/pkg/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/mock"
)

// MockLLM is a mock implementation of LLM interface using testify/mock
type MockLLM struct {
	mock.Mock
}

func (m *MockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	args := m.Called()
	return args.Get(0).(map[string]config.ConfigRequirement)
}

func (m *MockLLM) BuildHeaders() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *MockLLM) BuildURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	args := m.Called(message, history)
	return args.Get(0), args.Error(1)
}

func (m *MockLLM) ParseResponse(response []byte) (string, error) {
	args := m.Called(response)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) GetUsage(data []byte) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	args := m.Called(ctx, client, message, history)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) Complete(message string) (string, error) {
	args := m.Called(message)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) GenerateCommitMessage(diff string, prompt string) (string, error) {
	args := m.Called(diff, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) TranslateMessage(prompt string, message string, targetLang string) (string, error) {
	args := m.Called(prompt, message, targetLang)
	return args.String(0), args.Error(1)
}

func (m *MockLLM) Name() string {
	args := m.Called()
	return args.String(0)
}
