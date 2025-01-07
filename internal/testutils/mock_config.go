package testutils

// MockConfigManager 实现配置管理器接口用于测试
type MockConfigManager struct {
	Data                 map[string]interface{}
	Path                 string
	GetFunc              func(key string) (interface{}, bool)
	ListFunc             func() (string, error)
	ResetFunc            func(promptOnly bool) error
	SetFunc              func(key string, value interface{}) error
	GetPathFunc          func() string
	RemoveFunc           func(key string, value string) error
	AppendFunc           func(key string, value interface{}) error
	GetSupportedKeysFunc func() []string
}

func (m *MockConfigManager) Get(key string) (interface{}, bool) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	val, ok := m.Data[key]
	return val, ok
}

func (m *MockConfigManager) List() (string, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return "{\"test\": \"value\"}", nil
}

func (m *MockConfigManager) Reset(promptOnly bool) error {
	if m.ResetFunc != nil {
		return m.ResetFunc(promptOnly)
	}
	return nil
}

func (m *MockConfigManager) Set(key string, value interface{}) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value)
	}
	return nil
}

func (m *MockConfigManager) GetPath() string {
	if m.GetPathFunc != nil {
		return m.GetPathFunc()
	}
	return m.Path
}

func (m *MockConfigManager) Remove(key string, value string) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(key, value)
	}
	return nil
}

func (m *MockConfigManager) Append(key string, value interface{}) error {
	if m.AppendFunc != nil {
		return m.AppendFunc(key, value)
	}
	return nil
}

func (m *MockConfigManager) GetSupportedKeys() []string {
	if m.GetSupportedKeysFunc != nil {
		return m.GetSupportedKeysFunc()
	}
	return []string{}
}