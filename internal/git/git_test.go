package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupVCSTest creates a test repository and returns the path and cleanup function
func setupVCSTest(t *testing.T, vcsType VCSType) (vcs VCS, dir string, cleanup func()) {
	t.Helper()
	if vcsType == SVN {
		skipIfSVNNotAvailable(t)
	}
	dir = t.TempDir()

	v, err := NewVCS(vcsType)
	require.NoError(t, err)

	if vcsType == Git {
		// Setup Git repository
		err = testutils.RunGitCommand(t, dir, "init")
		require.NoError(t, err)

		err = testutils.RunGitCommand(t, dir, "config", "user.email", "test@example.com")
		require.NoError(t, err)
		err = testutils.RunGitCommand(t, dir, "config", "user.name", "Test User")
		require.NoError(t, err)
	} else if vcsType == SVN {
		// Setup SVN repository
		err = testutils.RunCommand(t, dir, "svnadmin", "create", "repo")
		require.NoError(t, err)
		err = testutils.RunCommand(t, dir, "svn", "checkout", "file://"+filepath.Join(dir, "repo"), dir)
		require.NoError(t, err)
	}

	cleanup = func() {
		os.RemoveAll(dir)
	}

	return v, dir, cleanup
}

func TestVCSImplementations(t *testing.T) {
	testCases := []struct {
		name    string
		vcsType VCSType
		skip    func(t *testing.T)
	}{
		{
			name:    "Git VCS",
			vcsType: Git,
			skip:    func(t *testing.T) {}, // Git tests always run
		},
		{
			name:    "SVN VCS",
			vcsType: SVN,
			skip:    skipIfSVNNotAvailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.skip(t) // Skip if necessary
			vcs, dir, cleanup := setupVCSTest(t, tc.vcsType)
			defer cleanup()

			// Test file addition and diff retrieval
			t.Run("GetDiff", func(t *testing.T) {
				// Create test file
				err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test content"), 0644)
				require.NoError(t, err)

				if tc.vcsType == Git {
					err = testutils.RunGitCommand(t, dir, "add", "test.txt")
				} else {
					err = testutils.RunCommand(t, dir, "svn", "add", "test.txt")
				}
				require.NoError(t, err)

				// Test GetDiff
				diff, err := vcs.GetDiff(dir)
				require.NoError(t, err)
				assert.Contains(t, diff, "test content")
			})

			// Test change detection
			t.Run("HasStagedChanges", func(t *testing.T) {
				hasChanges, err := vcs.HasStagedChanges(dir)
				require.NoError(t, err)
				assert.True(t, hasChanges)
			})

			// Test getting list of changed files
			t.Run("GetStagedFiles", func(t *testing.T) {
				files, err := vcs.GetStagedFiles(dir)
				require.NoError(t, err)
				assert.Contains(t, files, "test.txt")
			})

			// Test commit creation
			t.Run("CreateCommit", func(t *testing.T) {
				err := vcs.CreateCommit(dir, "test commit")
				require.NoError(t, err)

				// Verify commit success
				hash, err := vcs.GetLastCommitHash(dir)
				require.NoError(t, err)
				assert.NotEmpty(t, hash)

				info, err := vcs.GetCommitInfo(dir, hash)
				require.NoError(t, err)
				assert.Contains(t, info, "test commit")
			})

			// Test getting current branch
			t.Run("GetCurrentBranch", func(t *testing.T) {
				branch, err := vcs.GetCurrentBranch(dir)
				require.NoError(t, err)
				if tc.vcsType == Git {
					assert.Equal(t, "master", branch)
				} else {
					assert.Contains(t, branch, "file://")
				}
			})
		})
	}
}

func TestNewVCS(t *testing.T) {
	testCases := []struct {
		name     string
		vcsType  VCSType
		expected VCS
	}{
		{
			name:     "Git VCS",
			vcsType:  Git,
			expected: &GitVCS{},
		},
		{
			name:     "SVN VCS",
			vcsType:  SVN,
			expected: &SVNVCS{},
		},
		{
			name:     "Default VCS",
			vcsType:  "unknown",
			expected: &GitVCS{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vcs, err := NewVCS(tc.vcsType)
			require.NoError(t, err)
			assert.IsType(t, tc.expected, vcs)
		})
	}
}
