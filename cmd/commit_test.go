package cmd

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"testing"
	"os/exec"
	"path/filepath"

	"github.com/belingud/go-gptcomet/internal/git"
	"github.com/belingud/go-gptcomet/internal/testutils"
	"github.com/belingud/go-gptcomet/pkg/config"
	"github.com/belingud/go-gptcomet/pkg/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommitCmd(t *testing.T) {
	cmd := NewCommitCmd()
	require.NotNil(t, cmd)

	// Test flags
	flags := map[string]bool{
		"config":  false,
		"dry-run": false,
		"rich":    false,
		"svn":     false,
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if _, ok := flags[flag.Name]; ok {
			flags[flag.Name] = true
		}
	})

	for name, found := range flags {
		assert.True(t, found, "flag %q not found", name)
	}
}

func TestCommitCmd_Git(t *testing.T) {
    gitPath, err := exec.LookPath("git")
    if err != nil || !isExecutable(gitPath) {
        t.Skip("git command not found or not executable")
    }
    testCommitCmd(t, false)
}

func TestCommitCmd_SVN(t *testing.T) {
    if _, err := exec.LookPath("svnadmin"); err != nil {
        t.Skip("svnadmin command not found, skipping test")
    }
    testCommitCmd(t, true)
}

func testCommitCmd(t *testing.T, useSVN bool) {
	var vcsType git.VCSType
	if useSVN {
		vcsType = git.SVN
	} else {
		vcsType = git.Git
	}

	// Create a temporary repository
	_, repoPath, cleanup := setupTestRepo(t, vcsType)
	defer cleanup()

	// Create a test file and stage it
	testFileContent := "test content"
	err := os.WriteFile(repoPath+"/test.txt", []byte(testFileContent), 0644)
	require.NoError(t, err)

	if useSVN {
		err = testutils.RunCommand(t, repoPath, "svn", "add", "test.txt")
	} else {
		err = testutils.RunGitCommand(t, repoPath, "add", "test.txt")
	}
	require.NoError(t, err)

	// Create config file
	configContent := `
provider: openai
openai:
  api_key: "test_api_key"
  model: "gpt-4"
`
	configPath, cleanupConfig := testutils.TestConfig(t, configContent)
	defer cleanupConfig()

	// Create and run command
	cmd := NewCommitCmd()
	args := []string{"--config", configPath, "--dry-run"}
	if useSVN {
		args = append(args, "--svn")
	}
	cmd.SetArgs(args)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Execute command
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "Generated commit message:")
}

func setupTestRepo(t *testing.T, vcsType git.VCSType) (git.VCS, string, func()) {
	t.Helper()
	dir := t.TempDir()

	vcs, err := git.NewVCS(vcsType)
	require.NoError(t, err)

	if vcsType == git.Git {
        gitPath, err := exec.LookPath("git")
        if err != nil || !isExecutable(gitPath) {
            t.Skip("git command not found or not executable")
        }
        
        // 确保使用绝对路径执行git命令
        gitCmd := filepath.Clean(gitPath)
        err = testutils.RunCommand(t, dir, gitCmd, "init")
		require.NoError(t, err)
		err = testutils.RunCommand(t, dir, gitCmd, "config", "user.email", "test@example.com")
		require.NoError(t, err)
		err = testutils.RunCommand(t, dir, gitCmd, "config", "user.name", "Test User")
		require.NoError(t, err)
	} else {
        if _, err := exec.LookPath("svnadmin"); err != nil {
            t.Skip("svnadmin command not found")
        }
		err = testutils.RunCommand(t, dir, "svnadmin", "create", "repo")
		require.NoError(t, err)
		err = testutils.RunCommand(t, dir, "svn", "checkout", "file://"+dir+"/repo", dir)
		require.NoError(t, err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return vcs, dir, cleanup
}

func isExecutable(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return (info.Mode() & 0111) != 0
}

func TestCommitCmd_NoStagedChanges(t *testing.T) {
    testCases := []struct {
        name    string
        vcsType git.VCSType
        setup   func(t *testing.T) bool
    }{
        {
            name:    "Git",
            vcsType: git.Git,
            setup: func(t *testing.T) bool {
                if _, err := exec.LookPath("git"); err != nil {
                    t.Skip("git command not found")
                    return false
                }
                return true
            },
        },
        {
            name:    "SVN",
            vcsType: git.SVN,
            setup: func(t *testing.T) bool {
                if _, err := exec.LookPath("svnadmin"); err != nil {
                    t.Skip("svnadmin command not found")
                    return false
                }
                return true
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            if !tc.setup(t) {
                return
            }
            _, _, cleanup := setupTestRepo(t, tc.vcsType)
            defer cleanup()

            cmd := NewCommitCmd()
            if tc.vcsType == git.SVN {
                cmd.SetArgs([]string{"--svn"})
            }

            var buf bytes.Buffer
            cmd.SetOut(&buf)

            err := cmd.Execute()
            require.Error(t, err)
            assert.Contains(t, err.Error(), "no staged changes found")
        })
    }
}

// Mock LLM implementation for testing
type mockLLM struct {
	name                  string
	generateCommitMessage func(diff string, prompt string) (string, error)
	translateMessage      func(prompt string, message string, lang string) (string, error)
	makeRequest           func(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error)
}

func (m *mockLLM) GetRequiredConfig() map[string]config.ConfigRequirement {
	return map[string]config.ConfigRequirement{}
}

func (m *mockLLM) BuildHeaders() map[string]string {
	return map[string]string{}
}

func (m *mockLLM) BuildURL() string {
	return "https://mock.api"
}

func (m *mockLLM) FormatMessages(message string, history []types.Message) (interface{}, error) {
	return message, nil
}

func (m *mockLLM) ParseResponse(response []byte) (string, error) {
	return string(response), nil
}

func (m *mockLLM) GetUsage(data []byte) (string, error) {
	return "", nil
}

func (m *mockLLM) MakeRequest(ctx context.Context, client *http.Client, message string, history []types.Message) (string, error) {
	if m.makeRequest != nil {
		return m.makeRequest(ctx, client, message, history)
	}
	return "mock response", nil
}

func (m *mockLLM) GenerateCommitMessage(diff string, prompt string) (string, error) {
	if m.generateCommitMessage != nil {
		return m.generateCommitMessage(diff, prompt)
	}
	return "Test commit message", nil
}

func (m *mockLLM) TranslateMessage(prompt string, message string, lang string) (string, error) {
	if m.translateMessage != nil {
		return m.translateMessage(prompt, message, lang)
	}
	return message, nil
}

func (m *mockLLM) Name() string {
	return m.name
}
