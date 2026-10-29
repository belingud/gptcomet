package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRepoRoot(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	root, err := GetRepoRoot(repo)
	require.NoError(t, err)

	wantRoot, err := filepath.EvalSymlinks(repo)
	require.NoError(t, err)
	gotRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	assert.Equal(t, wantRoot, gotRoot)

	subdir := filepath.Join(repo, "nested")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	rootFromSubdir, err := GetRepoRoot(subdir)
	require.NoError(t, err)
	gotFromSubdir, err := filepath.EvalSymlinks(rootFromSubdir)
	require.NoError(t, err)
	assert.Equal(t, wantRoot, gotFromSubdir)
}

func TestGetHooksDir(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	root, err := GetRepoRoot(repo)
	require.NoError(t, err)

	hooksDir, err := GetHooksDir(repo)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".git", "hooks"), hooksDir)

	customDir := filepath.Join(repo, "custom-hooks")
	require.NoError(t, os.MkdirAll(customDir, 0o755))
	require.NoError(t, testutils.RunGitCommand(t, repo, "config", "core.hooksPath", customDir))

	customHooksDir, err := GetHooksDir(repo)
	require.NoError(t, err)
	wantCustomDir, err := filepath.EvalSymlinks(customDir)
	require.NoError(t, err)
	gotCustomHooksDir, err := filepath.EvalSymlinks(customHooksDir)
	require.NoError(t, err)
	assert.Equal(t, wantCustomDir, gotCustomHooksDir)
}

func TestGetRepoRoot_NotARepository(t *testing.T) {
	_, err := GetRepoRoot(t.TempDir())
	require.Error(t, err)
}

func TestGetHooksDir_NotARepository(t *testing.T) {
	_, err := GetHooksDir(t.TempDir())
	require.Error(t, err)
}
