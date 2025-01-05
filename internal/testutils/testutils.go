package testutils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDir creates a temporary directory for testing
func TestDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "gptcomet-test-*")
	require.NoError(t, err, "Failed to create test directory")

	cleanup := func() {
		err := os.RemoveAll(dir)
		assert.NoError(t, err, "Failed to cleanup test directory")
	}

	return dir, cleanup
}

// TestFile creates a temporary file with content for testing
func TestFile(t *testing.T, content string) (string, func()) {
	t.Helper()

	dir, cleanupDir := TestDir(t)

	file := filepath.Join(dir, "test.txt")
	err := os.WriteFile(file, []byte(content), 0644)
	require.NoError(t, err, "Failed to write test file")

	cleanup := func() {
		cleanupDir()
	}

	return file, cleanup
}

// TestConfig creates a temporary config file for testing
func TestConfig(t *testing.T, content string) (string, func()) {
	t.Helper()

	dir, cleanupDir := TestDir(t)

	configFile := filepath.Join(dir, "gptcomet.yaml")
	err := os.WriteFile(configFile, []byte(content), 0644)
	require.NoError(t, err, "Failed to write config file")

	cleanup := func() {
		cleanupDir()
	}

	return configFile, cleanup
}

// TestGitRepo creates a temporary git repository for testing
func TestGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	dir, cleanupDir := TestDir(t)

	// Initialize git repository
	err := os.Chdir(dir)
	require.NoError(t, err, "Failed to change directory")

	err = RunGitCommand(t, dir, "init")
	require.NoError(t, err, "Failed to initialize git repository")

	err = RunGitCommand(t, dir, "config", "user.email", "test@example.com")
	require.NoError(t, err, "Failed to configure git email")

	err = RunGitCommand(t, dir, "config", "user.name", "Test User")
	require.NoError(t, err, "Failed to configure git username")

	cleanup := func() {
		cleanupDir()
	}

	return dir, cleanup
}

// RunGitCommand runs a git command in the specified directory
func RunGitCommand(t *testing.T, dir string, args ...string) error {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Git command failed: %s\nOutput: %s", strings.Join(args, " "), output)
	}
	return err
}

// RunCommand executes any command and returns the result
func RunCommand(t *testing.T, dir string, name string, args ...string) error {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	return cmd.Run()
}
