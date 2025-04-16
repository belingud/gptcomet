package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isSVNAvailable checks if SVN tools are installed on the system
func isSVNAvailable() bool {
	_, err := exec.LookPath("svn")
	return err == nil
}

// skipIfSVNNotAvailable skips the test if SVN is not installed
func skipIfSVNNotAvailable(t *testing.T) {
	t.Helper()
	if !isSVNAvailable() {
		t.Skip("SVN is not installed, skipping SVN tests")
	}
}

// setupSVNTest creates a test SVN repository and returns related information
func setupSVNTest(t *testing.T) (vcs *SVNVCS, dir string, cleanup func()) {
	t.Helper()
	skipIfSVNNotAvailable(t)
	dir = t.TempDir()

	// Create SVN repository
	err := testutils.RunCommand(t, dir, "svnadmin", "create", "repo")
	require.NoError(t, err)

	// Checkout repository
	workDir := filepath.Join(dir, "work")
	err = os.Mkdir(workDir, 0755)
	require.NoError(t, err)

	repoURL := "file://" + filepath.Join(dir, "repo")
	err = testutils.RunCommand(t, dir, "svn", "checkout", repoURL, workDir)
	require.NoError(t, err)

	cleanup = func() {
		os.RemoveAll(dir)
	}

	return &SVNVCS{}, workDir, cleanup
}

func TestSVNVCS(t *testing.T) {
	skipIfSVNNotAvailable(t)
	vcs, dir, cleanup := setupSVNTest(t)
	defer cleanup()

	t.Run("GetDiff with no changes", func(t *testing.T) {
		diff, err := vcs.GetDiff(dir)
		require.NoError(t, err)
		assert.Empty(t, diff)
	})

	t.Run("HasStagedChanges with no changes", func(t *testing.T) {
		hasChanges, err := vcs.HasStagedChanges(dir)
		require.NoError(t, err)
		assert.False(t, hasChanges)
	})

	t.Run("Add and commit file", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(dir, "test.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)

		// Add file to SVN
		err = testutils.RunCommand(t, dir, "svn", "add", testFile)
		require.NoError(t, err)

		// Check for pending changes
		hasChanges, err := vcs.HasStagedChanges(dir)
		require.NoError(t, err)
		assert.True(t, hasChanges)

		// Get differences
		diff, err := vcs.GetDiff(dir)
		require.NoError(t, err)
		assert.Contains(t, diff, "test content")

		// Get staged files
		files, err := vcs.GetStagedFiles(dir)
		require.NoError(t, err)
		assert.Contains(t, files, "test.txt")

		// Create commit
		err = vcs.CreateCommit(dir, "test commit", false) // skipHook is ignored by SVN
		require.NoError(t, err)

		// Verify commit
		hash, err := vcs.GetLastCommitHash(dir)
		require.NoError(t, err)
		assert.Equal(t, "1", hash) // First commit should be revision 1

		// Verify commit message
		info, err := vcs.GetCommitInfo(dir, hash)
		require.NoError(t, err)
		assert.Contains(t, info, "test commit")
	})

	t.Run("GetCurrentBranch", func(t *testing.T) {
		branch, err := vcs.GetCurrentBranch(dir)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(branch, "file://"))
	})

	t.Run("Invalid repository path", func(t *testing.T) {
		invalidDir := filepath.Join(t.TempDir(), "nonexistent")
		_, err := vcs.GetDiff(invalidDir)
		assert.Error(t, err)
	})
}

func TestSVNVCSErrors(t *testing.T) {
	skipIfSVNNotAvailable(t)
	vcs := &SVNVCS{}
	invalidDir := filepath.Join(t.TempDir(), "nonexistent")

	testCases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "GetDiff error",
			fn: func() error {
				_, err := vcs.GetDiff(invalidDir)
				return err
			},
		},
		{
			name: "HasStagedChanges error",
			fn: func() error {
				_, err := vcs.HasStagedChanges(invalidDir)
				return err
			},
		},
		{
			name: "GetStagedFiles error",
			fn: func() error {
				_, err := vcs.GetStagedFiles(invalidDir)
				return err
			},
		},
		{
			name: "CreateCommit error",
			fn: func() error {
				return vcs.CreateCommit(invalidDir, "test commit", false) // skipHook is ignored by SVN
			},
		},
		{
			name: "GetLastCommitHash error",
			fn: func() error {
				_, err := vcs.GetLastCommitHash(invalidDir)
				return err
			},
		},
		{
			name: "GetCommitInfo error",
			fn: func() error {
				_, err := vcs.GetCommitInfo(invalidDir, "1")
				return err
			},
		},
		{
			name: "GetCurrentBranch error",
			fn: func() error {
				_, err := vcs.GetCurrentBranch(invalidDir)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			assert.Error(t, err)
		})
	}
}
