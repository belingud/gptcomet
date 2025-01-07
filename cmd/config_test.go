package cmd

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/cobra"
	"github.com/belingud/go-gptcomet/internal/testutils"
)

// setupTest 创建测试环境
func setupTest() (*cobra.Command, *testutils.MockConfigManager) {
	mock := &testutils.MockConfigManager{
		Data: make(map[string]interface{}),
		Path: "/mock/config/path",
	}
	cmd := &cobra.Command{}
	// 初始化 command 的 context
	cmd.SetContext(context.Background())
	setConfigManager(cmd, mock)
	return cmd, mock
}

func TestConfigManagerContext(t *testing.T) {
	cmd, mock := setupTest()
	
	got, err := getConfigManager(cmd)
	assert.NoError(t, err)
	assert.Equal(t, mock, got)
}

func TestGetConfigCmd(t *testing.T) {
	cmd := newGetConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{
		GetFunc: func(key string) (interface{}, bool) {
			return "test-value", true
		},
	}
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSetConfigCmd(t *testing.T) {
	cmd := newSetConfigCmd()
	cmd.SetContext(context.Background())
	var setCalled bool
	mock := &testutils.MockConfigManager{
		SetFunc: func(key string, value interface{}) error {
			assert.Equal(t, "test.key", key)
			assert.Equal(t, "new-value", value)
			setCalled = true
			return nil
		},
	}
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key", "new-value"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, setCalled)
}

func TestListConfigCmd(t *testing.T) {
	cmd := newListConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{
		ListFunc: func() (string, error) {
			return "{\"test\": \"value\"}", nil
		},
	}
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestResetConfigCmd(t *testing.T) {
	cmd := newResetConfigCmd()
	cmd.SetContext(context.Background())
	var resetCalled bool
	mock := &testutils.MockConfigManager{
		ResetFunc: func(promptOnly bool) error {
			assert.False(t, promptOnly)
			resetCalled = true
			return nil
		},
	}
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"--prompt=false"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, resetCalled)
}

func TestPathConfigCmd(t *testing.T) {
	cmd := newPathConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{
		GetPathFunc: func() string {
			return "/mock/config/path"
		},
	}
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRemoveConfigCmd(t *testing.T) {
	cmd := newRemoveConfigCmd()
	cmd.SetContext(context.Background())
	var removeCalled bool
	mock := &testutils.MockConfigManager{
		GetFunc: func(key string) (interface{}, bool) {
			return "test-value", true
		},
		RemoveFunc: func(key string, value string) error {
			assert.Equal(t, "test.key", key)
			assert.Equal(t, "", value)
			removeCalled = true
			return nil
		},
	}
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, removeCalled)
}

func TestAppendConfigCmd(t *testing.T) {
	cmd := newAppendConfigCmd()
	cmd.SetContext(context.Background())
	var appendCalled bool
	mock := &testutils.MockConfigManager{
		GetFunc: func(key string) (interface{}, bool) {
			return []interface{}{}, true
		},
		AppendFunc: func(key string, value interface{}) error {
			assert.Equal(t, "test.list", key)
			assert.Equal(t, "value1", value)
			appendCalled = true
			return nil
		},
	}
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.list", "value1"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.True(t, appendCalled)
}

func TestKeysConfigCmd(t *testing.T) {
	cmd := newKeysConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{
		GetSupportedKeysFunc: func() []string {
			return []string{"test.key", "openai.api_key"}
		},
	}
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, 8, len(cmd.Commands()))
}
