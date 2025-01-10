package cmd

import (
	"context"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// setupTest creates test environment
func setupTest() (*cobra.Command, *testutils.MockConfigManager) {
	mock := &testutils.MockConfigManager{}
	cmd := &cobra.Command{}
	// Initialize command context
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
	mock := &testutils.MockConfigManager{}
	mock.On("Get", "test.key").Return("test-value", true)
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key"})
	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestSetConfigCmd(t *testing.T) {
	cmd := newSetConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("Set", "test.key", "new-value").Return(nil)
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key", "new-value"})
	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestListConfigCmd(t *testing.T) {
	cmd := newListConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("List").Return("{\"test\": \"value\"}", nil)
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestResetConfigCmd(t *testing.T) {
	cmd := newResetConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("Reset", false).Return(nil)
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"--prompt=false"})
	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestPathConfigCmd(t *testing.T) {
	cmd := newPathConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("GetPath").Return("/mock/config/path")
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestRemoveConfigCmd(t *testing.T) {
	cmd := newRemoveConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("Get", "test.key").Return("test-value", true)
	mock.On("Remove", "test.key", "").Return(nil)
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.key"})
	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestAppendConfigCmd(t *testing.T) {
	cmd := newAppendConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("Get", "test.list").Return([]interface{}{}, true)
	mock.On("Append", "test.list", "value1").Return(nil)
	setConfigManager(cmd, mock)

	cmd.SetArgs([]string{"test.list", "value1"})
	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestKeysConfigCmd(t *testing.T) {
	cmd := newKeysConfigCmd()
	cmd.SetContext(context.Background())
	mock := &testutils.MockConfigManager{}
	mock.On("GetSupportedKeys").Return([]string{"test.key", "openai.api_key"})
	setConfigManager(cmd, mock)

	err := cmd.Execute()
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.Equal(t, 8, len(cmd.Commands()))
}
