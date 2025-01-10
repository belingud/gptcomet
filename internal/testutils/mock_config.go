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

func (m *MockConfigManager) GetClientConfig() (*types.ClientConfig, error) {
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
