package testutils

import (
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/mock"
)

// MockConfigManager is a mock implementation of config.Manager
type MockConfigManager struct {
	mock.Mock
}

func (m *MockConfigManager) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

func (m *MockConfigManager) Set(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockConfigManager) List() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockConfigManager) Reset(promptOnly bool) error {
	args := m.Called(promptOnly)
	return args.Error(0)
}

func (m *MockConfigManager) Remove(key string, value string) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockConfigManager) Append(key string, value interface{}) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockConfigManager) GetPath() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigManager) GetClientConfig(initProvider string) (*types.ClientConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ClientConfig), args.Error(1)
}

func (m *MockConfigManager) GetSupportedKeys() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockConfigManager) GetFileIgnore() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockConfigManager) GetOutputTranslateTitle() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConfigManager) GetPrompt(rich bool) string {
	args := m.Called(rich)
	return args.String(0)
}

func (m *MockConfigManager) GetReviewPrompt() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigManager) GetTranslationPrompt() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfigManager) GetWithDefault(key string, defaultValue interface{}) interface{} {
	args := m.Called(key, defaultValue)
	return args.Get(0)
}

func (m *MockConfigManager) ListWithoutPrompt() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockConfigManager) UpdateProviderConfig(provider string, config map[string]string) error {
	args := m.Called(provider, config)
	return args.Error(0)
}

func (m *MockConfigManager) GetNestedValue(keys []string) (interface{}, bool) {
	args := m.Called(keys)
	return args.Get(0), args.Bool(1)
}

func (m *MockConfigManager) SetNestedValue(keys []string, value interface{}) {
	m.Called(keys, value)
}

func (m *MockConfigManager) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConfigManager) Save() error {
	args := m.Called()
	return args.Error(0)
}
