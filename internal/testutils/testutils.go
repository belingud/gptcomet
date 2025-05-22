package testutils

import (
	"fmt"
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

	configPath := CreateTestConfig(t, content)
	cleanup := func() {
		os.Remove(configPath)
	}

	return configPath, cleanup
}

// TestGitRepo creates a temporary git repository for testing
func TestGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	repoPath, cleanup := CreateTestRepo(t)
	if err := RunGitCommand(t, repoPath, "init"); err != nil {
		cleanup()
		t.Fatal(err)
	}
	return repoPath, cleanup
}

// RunGitCommand runs a git command in the specified directory
func RunGitCommand(t *testing.T, dir string, args ...string) error {
	t.Helper()

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git command not found")
	}

	// create environment variables
	env := append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_AUTHOR_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
	)

	cmd := exec.Command(gitPath, args...)
	cmd.Dir = dir
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Git command failed: git %v\nOutput: %s", args, output)
		return err
	}

	return nil
}

// RunCommand executes any command and returns the result
func RunCommand(t *testing.T, dir string, name string, args ...string) error {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed: %s %v\nOutput: %s", name, args, output)
		return err
	}

	return nil
}

// CreateTestConfig creates a temporary config file for testing
func CreateTestConfig(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err, "failed to write test config file")

	return configPath
}

// CreateTestDir creates a temporary directory for testing
func CreateTestDir(t *testing.T) (string, func()) {
	t.Helper()

	dir := t.TempDir()
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func InitGitRepo(t *testing.T, dir string) error {
	t.Helper()

	gitCmd := "git"

	// set environment variables
	env := []string{
		"HOME=" + dir,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_AUTHOR_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
	}

	// init git repository
	cmds := [][]string{
		{"init"},
		{"config", "--local", "user.email", "test@example.com"},
		{"config", "--local", "user.name", "Test User"},
	}

	for _, args := range cmds {
		cmd := exec.Command(gitCmd, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %v failed: %v\nOutput: %s", args, err, out)
		}
	}

	return nil
}

func StageFile(t *testing.T, dir string, file string) error {
	t.Helper()

	gitPath, err := exec.Command("which", "git").Output()
	if err != nil {
		return fmt.Errorf("git not found: %w", err)
	}
	gitCmd := strings.TrimSpace(string(gitPath))

	cmd := exec.Command(gitCmd, "add", file)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"HOME="+dir,
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %v\nOutput: %s", err, out)
	}

	return nil
}

// CreateTestRepo creates a temporary repository for testing
func CreateTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "gptcomet-test-*")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}
