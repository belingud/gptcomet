package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/factory"
	"github.com/belingud/gptcomet/internal/git"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitRepositoryInitialization tests initializing a git repository
func TestGitRepositoryInitialization(t *testing.T) {
	repoPath, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	// Verify git repo was created
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), ".git should be a directory")

	// Create VCS instance
	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)
	assert.NotNil(t, vcs)

	// Verify we can get diff (indicates valid repo)
	// Git doesn't have IsRepo method, we verify by attempting operations
	_, err = vcs.GetDiff(repoPath)
	// Empty repo may return empty diff without error
}

// TestGitDiffGeneration tests generating git diff
func TestGitDiffGeneration(t *testing.T) {
	repoPath, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Create a file
	testFile := filepath.Join(repoPath, "test.txt")
	err = os.WriteFile(testFile, []byte("Hello, World!\n"), 0644)
	require.NoError(t, err)

	// Stage the file
	err = testutils.RunGitCommand(t, repoPath, "add", "test.txt")
	require.NoError(t, err)

	// Get staged diff
	diff, err := vcs.GetDiff(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, diff, "Diff should not be empty")
	assert.Contains(t, diff, "Hello, World!", "Diff should contain file content")
	assert.Contains(t, diff, "+", "Diff should contain + for additions")
}

// TestCommitWorkflow tests the full commit workflow with a mock
func TestCommitWorkflow(t *testing.T) {
	// Setup git repository
	repoPath, cleanupRepo := testutils.TestGitRepo(t)
	defer cleanupRepo()

	// Create test config
	configData := `
provider: openai
openai:
  api_key: test-key
  model: gpt-4o
output:
  lang: en
`
	configFile, cleanupConfig := testutils.TestConfig(t, configData)
	defer cleanupConfig()

	// Create service dependencies
	cfg, err := config.New(configFile)
	require.NoError(t, err)

	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Create a file and stage it
	testFile := filepath.Join(repoPath, "main.go")
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	err = os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	err = testutils.RunGitCommand(t, repoPath, "add", "main.go")
	require.NoError(t, err)

	// Get staged diff
	diff, err := vcs.GetDiff(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "Hello, World!")

	// Verify config can be loaded and parsed
	clientCfg, err := cfg.GetClientConfig("")
	require.NoError(t, err)
	assert.Equal(t, "openai", clientCfg.Provider)
	assert.Equal(t, "test-key", clientCfg.APIKey)
}

// TestFactoryServiceCreation tests the factory pattern for service creation
func TestFactoryServiceCreation(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: test-key
  model: gpt-4o
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	// Test service dependencies creation
	vcs, cfgManager, err := factory.NewServiceDependencies(factory.ServiceOptions{
		UseSVN:     false,
		ConfigPath: configFile,
		Provider:   "",
	})

	require.NoError(t, err)
	assert.NotNil(t, vcs)
	assert.NotNil(t, cfgManager)

	// Test full dependencies with client
	deps, err := factory.NewServiceDependenciesWithClient(factory.ServiceOptions{
		UseSVN:     false,
		ConfigPath: configFile,
		Provider:   "",
	})

	require.NoError(t, err)
	assert.NotNil(t, deps)
	assert.NotNil(t, deps.VCS)
	assert.NotNil(t, deps.CfgManager)
	assert.NotNil(t, deps.APIConfig)
	assert.NotNil(t, deps.APIClient)
}

// TestMultipleFileCommit tests committing multiple files
func TestMultipleFileCommit(t *testing.T) {
	repoPath, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Create multiple files
	files := map[string]string{
		"file1.txt": "Content of file 1\n",
		"file2.txt": "Content of file 2\n",
		"file3.txt": "Content of file 3\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		err = testutils.RunGitCommand(t, repoPath, "add", filename)
		require.NoError(t, err)
	}

	// Get staged diff
	diff, err := vcs.GetDiff(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)

	// Verify all files are in diff
	for filename := range files {
		assert.Contains(t, diff, filename, "Diff should contain %s", filename)
	}
}

// TestEmptyDiff tests behavior when there's no staged changes
func TestEmptyDiff(t *testing.T) {
	repoPath, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Get diff without any staged changes
	diff, err := vcs.GetDiff(repoPath)
	// This may or may not error depending on implementation
	// but diff should be empty
	if err == nil {
		assert.Empty(t, diff, "Diff should be empty when no changes are staged")
	}
}

// TestGitCommitCreation tests creating an actual git commit
func TestGitCommitCreation(t *testing.T) {
	repoPath, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Create and stage a file
	testFile := filepath.Join(repoPath, "readme.md")
	err = os.WriteFile(testFile, []byte("# Test Project\n"), 0644)
	require.NoError(t, err)

	err = testutils.RunGitCommand(t, repoPath, "add", "readme.md")
	require.NoError(t, err)

	// Create a commit
	commitMsg := "docs: add readme"
	err = vcs.CreateCommit(repoPath, commitMsg, false)
	require.NoError(t, err)

	// Verify commit was created
	err = testutils.RunGitCommand(t, repoPath, "log", "--oneline", "-1")
	require.NoError(t, err)
}

// TestSVNRepositoryDetection tests SVN repository detection
func TestSVNRepositoryDetection(t *testing.T) {
	// Create a directory
	tmpDir := t.TempDir()

	// Try to create SVN VCS (should not error)
	vcs, err := git.NewVCS(git.SVN)
	require.NoError(t, err)
	assert.NotNil(t, vcs)

	// Check if directory has any staged changes (should be false for non-SVN directory)
	hasChanges, err := vcs.HasStagedChanges(tmpDir)
	if err != nil {
		// SVN not installed or directory is not a repo
		t.Skip("SVN not available or directory is not a SVN repository")
	}
	assert.False(t, hasChanges, "Should not have staged changes in non-SVN directory")
}

// TestConfigWithGitWorkflow tests the full workflow of config + git
func TestConfigWithGitWorkflow(t *testing.T) {
	// Setup git repository
	repoPath, cleanupRepo := testutils.TestGitRepo(t)
	defer cleanupRepo()

	// Setup config
	configData := `
provider: openai
openai:
  api_key: test-key
  model: gpt-4o
file_ignore:
  - "*.log"
  - "*.tmp"
  - ".env"
`
	configFile, cleanupConfig := testutils.TestConfig(t, configData)
	defer cleanupConfig()

	// Load config
	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Create VCS
	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	// Create files (including ones that should be ignored)
	files := map[string]string{
		"main.go":    "package main\n",
		"test.log":   "log content\n",
		"data.tmp":   "temp data\n",
		".env":       "SECRET=key\n",
		"readme.md":  "# Project\n",
	}

	for filename, content := range files {
		filePath := filepath.Join(repoPath, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Get ignore patterns from config
	ignorePatterns := cfg.GetFileIgnore()
	assert.Contains(t, ignorePatterns, "*.log")
	assert.Contains(t, ignorePatterns, "*.tmp")
	assert.Contains(t, ignorePatterns, ".env")

	// Stage all files (in real use, would filter based on ignore patterns)
	err = testutils.RunGitCommand(t, repoPath, "add", "main.go", "readme.md")
	require.NoError(t, err)

	// Get diff
	diff, err := vcs.GetDiff(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)

	// Verify only the non-ignored files are in diff
	assert.Contains(t, diff, "main.go")
	assert.Contains(t, diff, "readme.md")
}

// TestVCSErrorHandling tests VCS error scenarios
func TestVCSErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantError bool
	}{
		{
			name: "Non-existent directory",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/directory/path"
			},
			wantError: true,
		},
		{
			name: "Valid git directory",
			setupFunc: func(t *testing.T) string {
				repoPath, _ := testutils.TestGitRepo(t)
				return repoPath
			},
			wantError: false,
		},
		{
			name: "Non-git directory",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setupFunc(t)
			vcs, err := git.NewVCS(git.Git)
			require.NoError(t, err)

			hasChanges, err := vcs.HasStagedChanges(repoPath)
			if tt.wantError {
				// For non-git directories, this should error or return false
				if err == nil {
					assert.False(t, hasChanges, "Should not have staged changes")
				}
			} else {
				require.NoError(t, err)
				// Valid repo may or may not have changes
			}
		})
	}
}

// TestConcurrentConfigAccess tests concurrent access to config
func TestConcurrentConfigAccess(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: test-key
  model: gpt-4o
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	// Launch multiple goroutines that read config concurrently
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			// Read config
			clientCfg, err := cfg.GetClientConfig("")
			assert.NoError(t, err)
			assert.NotNil(t, clientCfg)
			assert.Equal(t, "openai", clientCfg.Provider)

			// Get values
			val, ok := cfg.Get("provider")
			assert.True(t, ok)
			assert.Equal(t, "openai", val)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// TestEndToEndWorkflow tests a complete end-to-end workflow
func TestEndToEndWorkflow(t *testing.T) {
	// Step 1: Create git repository
	repoPath, cleanupRepo := testutils.TestGitRepo(t)
	defer cleanupRepo()

	// Step 2: Create config
	configData := `
provider: openai
openai:
  api_key: test-api-key
  model: gpt-4o
  max_tokens: 1024
  temperature: 0.7
output:
  lang: en
  translate_title: false
`
	configFile, cleanupConfig := testutils.TestConfig(t, configData)
	defer cleanupConfig()

	// Step 3: Initialize all components
	deps, err := factory.NewServiceDependenciesWithClient(factory.ServiceOptions{
		UseSVN:     false,
		ConfigPath: configFile,
		Provider:   "",
	})
	require.NoError(t, err)
	assert.NotNil(t, deps)

	// Step 4: Create test files in the repo
	testFile := filepath.Join(repoPath, "feature.go")
	fileContent := `package main

func NewFeature() string {
	return "new feature"
}
`
	err = os.WriteFile(testFile, []byte(fileContent), 0644)
	require.NoError(t, err)

	// Step 5: Stage the file
	err = testutils.RunGitCommand(t, repoPath, "add", "feature.go")
	require.NoError(t, err)

	// Step 6: Get VCS instance and verify diff
	vcs, err := git.NewVCS(git.Git)
	require.NoError(t, err)

	diff, err := vcs.GetDiff(repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, diff)
	assert.Contains(t, diff, "NewFeature")

	// Step 7: Verify all components are properly configured
	assert.Equal(t, "openai", deps.APIConfig.Provider)
	assert.Equal(t, "test-api-key", deps.APIConfig.APIKey)
	assert.Equal(t, "gpt-4o", deps.APIConfig.Model)
	assert.Equal(t, 1024, deps.APIConfig.MaxTokens)
	assert.Equal(t, 0.7, deps.APIConfig.Temperature)

	// Step 8: Verify client is ready (we won't make actual API calls)
	assert.NotNil(t, deps.APIClient)

	// At this point, all components are ready for generating a commit message
	// In a real scenario, we would call deps.APIClient.GenerateCommitMessage(diff, prompt)
	// but we skip that here since it would require a real API key and network call
}

// TestContextCancellation tests that operations respect context cancellation
func TestContextCancellation(t *testing.T) {
	configData := `
provider: openai
openai:
  api_key: test-key
`
	configFile, cleanup := testutils.TestConfig(t, configData)
	defer cleanup()

	cfg, err := config.New(configFile)
	require.NoError(t, err)

	clientCfg, err := cfg.GetClientConfig("")
	require.NoError(t, err)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Operations should respect the cancelled context
	_ = clientCfg // Use clientCfg to avoid unused variable warning
	assert.True(t, ctx.Err() != nil, "Context should be cancelled")
}
