package git

import (
	"os"
	"path/filepath"
	"testing"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
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
				err := vcs.CreateCommit(dir, "test commit", false)
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

func TestFilterIgnoredFiles(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		patterns []string
		want     []string
	}{
		{
			name:     "no patterns keeps all files",
			files:    []string{"a.go", "b.lock"},
			patterns: nil,
			want:     []string{"a.go", "b.lock"},
		},
		{
			name:     "exact match",
			files:    []string{"a.go", "go.sum"},
			patterns: []string{"go.sum"},
			want:     []string{"a.go"},
		},
		{
			name:     "wildcard match",
			files:    []string{"a.go", "b.lock", "c.tmp"},
			patterns: []string{"*.lock", "*.tmp"},
			want:     []string{"a.go"},
		},
		{
			name:     "all files ignored",
			files:    []string{"a.lock", "b.lock"},
			patterns: []string{"*.lock"},
			want:     []string{},
		},
		{
			name:     "empty input",
			files:    nil,
			patterns: []string{"*.lock"},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FilterIgnoredFiles(tt.files, tt.patterns))
		})
	}
}

func TestGitVCS_GetStagedDiffFiltered_UsesStringSlicePatterns(t *testing.T) {
	vcs, dir, cleanup := setupVCSTest(t, Git)
	defer cleanup()

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "nested"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "keep.go"), []byte("package main\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skip.lock"), []byte("lock\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nested", "keep.go"), []byte("package nested\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nested", "skip.lock"), []byte("lock\n"), 0644))
	require.NoError(t, testutils.RunGitCommand(t, dir, "add", "."))

	cfg := new(testutils.MockConfigManager)
	cfg.On("GetFileIgnore").Return([]string{"*.lock"})

	diff, err := vcs.GetStagedDiffFiltered(dir, cfg)
	require.NoError(t, err)
	assert.Contains(t, diff, "keep.go")
	assert.NotContains(t, diff, "skip.lock")
	assert.Contains(t, diff, "nested/keep.go")
	assert.NotContains(t, diff, "nested/skip.lock")
	cfg.AssertExpectations(t)
}

func TestRepoDisplayPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := t.TempDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty defaults to cwd", "", wd},
		{"dot resolves to cwd", ".", wd},
		{"absolute path unchanged", dir, dir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, repoDisplayPath(tt.input))
		})
	}
}

// TestGitVCSOutsideRepository verifies that every git command failure outside a
// working tree is classified as "Not a Git Repository" rather than a generic
// command failure: git diff falls back to --no-index mode there and only reports
// an unknown option, which used to leak into a misleading error.
func TestGitVCSOutsideRepository(t *testing.T) {
	dir := t.TempDir()
	vcs := &GitVCS{}

	tests := []struct {
		name string
		call func() error
	}{
		{"GetDiff", func() error { _, err := vcs.GetDiff(dir); return err }},
		{"HasStagedChanges", func() error { _, err := vcs.HasStagedChanges(dir); return err }},
		{"GetStagedFiles", func() error { _, err := vcs.GetStagedFiles(dir); return err }},
		{"GetCurrentBranch", func() error { _, err := vcs.GetCurrentBranch(dir); return err }},
		{"GetLastCommitHash", func() error { _, err := vcs.GetLastCommitHash(dir); return err }},
		{"GetCommitInfo", func() error { _, err := vcs.GetCommitInfo(dir, "abc123"); return err }},
		{"CreateCommit", func() error { return vcs.CreateCommit(dir, "message", false) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			require.Error(t, err)

			var gptErr *gptcometerrors.GPTCometError
			require.ErrorAs(t, err, &gptErr)
			assert.Equal(t, gptcometerrors.ErrTypeGit, gptErr.Type)
			assert.Equal(t, gptcometerrors.ErrTitleNotGitRepo, gptErr.Title)
			assert.Contains(t, gptErr.Message, dir)
			assert.Contains(t, gptErr.Suggestions, gptcometerrors.SuggInitGit)
		})
	}
}

// TestGitCommandFailureIncludesDetails verifies that a genuine in-repository
// failure keeps the "Git Command Failed" classification but exposes git's
// stderr/exit status in Details, with the command label rendered as plain args.
func TestGitCommandFailureIncludesDetails(t *testing.T) {
	vcs, dir, cleanup := setupVCSTest(t, Git)
	defer cleanup()

	_, err := vcs.GetCommitInfo(dir, "no-such-revision")
	require.Error(t, err)

	var gptErr *gptcometerrors.GPTCometError
	require.ErrorAs(t, err, &gptErr)
	assert.Equal(t, gptcometerrors.ErrTypeGit, gptErr.Type)
	assert.Equal(t, gptcometerrors.ErrTitleGitCommand, gptErr.Title)
	assert.Contains(t, gptErr.Message, "git log -1 --stat")
	assert.NotContains(t, gptErr.Message, "[")
	assert.Contains(t, gptErr.Details, "exit status 128")
}

// TestGitCommandNotInstalled verifies that a missing git executable is reported
// as "Git Not Installed" instead of "Not a Git Repository".
func TestGitCommandNotInstalled(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	vcs := &GitVCS{}
	_, err := vcs.GetDiff(t.TempDir())
	require.Error(t, err)

	var gptErr *gptcometerrors.GPTCometError
	require.ErrorAs(t, err, &gptErr)
	assert.Equal(t, gptcometerrors.ErrTypeGit, gptErr.Type)
	assert.Equal(t, gptcometerrors.ErrTitleGitNotInstalled, gptErr.Title)
	assert.Contains(t, gptErr.Details, "executable file not found")
	assert.Contains(t, gptErr.Suggestions, gptcometerrors.SuggCheckGitInstalled)
}
