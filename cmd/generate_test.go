package cmd

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerateCmd(t *testing.T) {
	cmd := NewGenerateCmd()
	assert.Equal(t, "generate", cmd.Use)

	for _, name := range []string{"repo", "rich", "model", "api-base", "timeout"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag %q should be registered", name)
	}
}

func TestGenerateWithCleanStdout(t *testing.T) {
	oldStdout, oldStderr := os.Stdout, os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)
	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
		_ = stderrReader.Close()
		_ = stderrWriter.Close()
	}()

	message, err := generateWithCleanStdout(func() (string, error) {
		fmt.Print("provider noise")
		fmt.Fprint(os.Stderr, "status")
		return "feat: message", nil
	})
	require.NoError(t, err)

	require.NoError(t, stdoutWriter.Close())
	require.NoError(t, stderrWriter.Close())

	stdoutBytes, err := io.ReadAll(stdoutReader)
	require.NoError(t, err)
	stderrBytes, err := io.ReadAll(stderrReader)
	require.NoError(t, err)

	assert.Empty(t, string(stdoutBytes))
	assert.Contains(t, string(stderrBytes), "provider noise")
	assert.Contains(t, string(stderrBytes), "status")
	assert.Equal(t, "feat: message", message)
}

func TestGenerateWithCleanStdout_RestoresOnError(t *testing.T) {
	originalStdout, originalStderr := os.Stdout, os.Stderr

	_, err := generateWithCleanStdout(func() (string, error) {
		return "", assert.AnError
	})
	require.Error(t, err)
	assert.Equal(t, originalStdout, os.Stdout)
	assert.Equal(t, originalStderr, os.Stderr)
}
