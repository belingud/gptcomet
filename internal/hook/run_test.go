package hook

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeMessageFile(t *testing.T, content string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
	return path
}

func TestManagerRunSkips(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	tests := []struct {
		name    string
		source  string
		content string
		env     map[string]string
	}{
		{name: "gmsg skip", source: "", content: "", env: map[string]string{"GMSG_SKIP": "1"}},
		{name: "merge source", source: "merge", content: "", env: map[string]string{}},
		{name: "squash source", source: "squash", content: "", env: map[string]string{}},
		{name: "commit source", source: "commit", content: "", env: map[string]string{}},
		{name: "template with user content", source: "template", content: "template message\n", env: map[string]string{}},
		{name: "message source", source: "message", content: "explicit message", env: map[string]string{}},
		{name: "existing content", source: "", content: "existing message\n", env: map[string]string{}},
		{name: "existing content with comments", source: "", content: "existing message\n# comment\n", env: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GMSG_SKIP", "")
			t.Setenv("GMSG_FORCE", "")
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			called := false
			manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
				called = true
				return "feat: generated", nil
			})
			path := writeMessageFile(t, tt.content, 0o644)

			err := manager.Run(context.Background(), RunInput{
				MessageFile: path,
				Source:      tt.source,
				RepoPath:    repo,
			})
			require.NoError(t, err)
			assert.False(t, called, "generator must not be called")

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(got))
		})
	}
}

func TestManagerRunWritesMessage(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	path := writeMessageFile(t, "", 0o600)
	called := false
	manager := NewManager(func(_ context.Context, repoPath, configPath string) (string, error) {
		called = true
		assert.NotEmpty(t, repoPath)
		assert.Empty(t, configPath)
		return "feat: add hook", nil
	})

	err := manager.Run(context.Background(), RunInput{MessageFile: path, Source: "template", RepoPath: repo})
	require.NoError(t, err)
	assert.True(t, called)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "feat: add hook\n", string(got))

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestManagerRunPreservesComments(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	path := writeMessageFile(t, "# Please enter the commit message.\n# On branch master\n", 0o644)
	manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
		return "feat: generated", nil
	})

	require.NoError(t, manager.Run(context.Background(), RunInput{MessageFile: path, RepoPath: repo}))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "feat: generated\n\n# Please enter the commit message.\n# On branch master\n", string(got))
}

func TestManagerRunForceOverrides(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	t.Setenv("GMSG_FORCE", "1")
	path := writeMessageFile(t, "old message\n# keep this comment\n", 0o644)

	manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
		return "feat: generated", nil
	})

	err := manager.Run(context.Background(), RunInput{
		MessageFile: path,
		Source:      "message",
		RepoPath:    repo,
	})
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "feat: generated\n\n# keep this comment\n", string(got))
}

func TestManagerRunGeneratorErrorLeavesFile(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	path := writeMessageFile(t, "", 0o644)
	manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
		return "", errors.New("API unavailable")
	})

	err := manager.Run(context.Background(), RunInput{MessageFile: path, RepoPath: repo})
	require.Error(t, err)

	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Empty(t, string(got))
}

func TestManagerRunEmptyMessageLeavesFile(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	path := writeMessageFile(t, "", 0o644)
	manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
		return "  \n", nil
	})

	require.NoError(t, manager.Run(context.Background(), RunInput{MessageFile: path, RepoPath: repo}))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, string(got))
}

func TestManagerRunMissingFile(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	manager := NewManager(func(_ context.Context, _, _ string) (string, error) {
		return "feat: generated", nil
	})

	err := manager.Run(context.Background(), RunInput{
		MessageFile: filepath.Join(t.TempDir(), "missing"),
		RepoPath:    repo,
	})
	require.Error(t, err)
}

func TestIsSkippedSource(t *testing.T) {
	tests := []struct {
		source string
		want   bool
	}{
		{source: "", want: false},
		{source: "message", want: true},
		{source: "merge", want: true},
		{source: "squash", want: true},
		{source: "commit", want: true},
		{source: "template", want: false},
		{source: "other", want: false},
		{source: " MERGE ", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			assert.Equal(t, tt.want, isSkippedSource(tt.source))
		})
	}
}

func TestIsTruthy(t *testing.T) {
	for _, value := range []string{"1", "true", "TRUE", "yes", "on"} {
		assert.True(t, isTruthy(value), value)
	}
	for _, value := range []string{"", "0", "false", "no", "off", "random"} {
		assert.False(t, isTruthy(value), value)
	}
}

func TestHasNonCommentContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "empty", content: "", want: false},
		{name: "comments only", content: "# one\n# two\n", want: false},
		{name: "blank lines", content: "\n \n", want: false},
		{name: "message", content: "feat: x\n", want: true},
		{name: "message after comments", content: "# one\nfeat: x\n", want: true},
		{name: "indented comment", content: "  # comment\n", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasNonCommentContent(tt.content))
		})
	}
}

func TestCommentLines(t *testing.T) {
	content := "old message\n# comment\n\n# another\n"
	assert.Equal(t, "# comment\n\n# another", commentLines(content))
}
