package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/go-gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupVCSTest 创建测试仓库并返回路径和清理函数
func setupVCSTest(t *testing.T, vcsType VCSType) (vcs VCS, dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()

	v, err := NewVCS(vcsType)
	require.NoError(t, err)

	if vcsType == Git {
		err = testutils.RunGitCommand(t, dir, "init")
		require.NoError(t, err)

		err = testutils.RunGitCommand(t, dir, "config", "user.email", "test@example.com")
		require.NoError(t, err)
		err = testutils.RunGitCommand(t, dir, "config", "user.name", "Test User")
		require.NoError(t, err)
	} else if vcsType == SVN {
		// 为 SVN 设置测试仓库
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
	}{
		{
			name:    "Git VCS",
			vcsType: Git,
		},
		{
			name:    "SVN VCS",
			vcsType: SVN,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vcs, dir, cleanup := setupVCSTest(t, tc.vcsType)
			defer cleanup()

			// 测试添加文件和获取差异
			t.Run("GetDiff", func(t *testing.T) {
				// 创建测试文件
				err := os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test content"), 0644)
				require.NoError(t, err)

				if tc.vcsType == Git {
					err = testutils.RunGitCommand(t, dir, "add", "test.txt")
				} else {
					err = testutils.RunCommand(t, dir, "svn", "add", "test.txt")
				}
				require.NoError(t, err)

				// 测试 GetDiff
				diff, err := vcs.GetDiff(dir)
				require.NoError(t, err)
				assert.Contains(t, diff, "test content")
			})

			// 测试检查变更
			t.Run("HasStagedChanges", func(t *testing.T) {
				hasChanges, err := vcs.HasStagedChanges(dir)
				require.NoError(t, err)
				assert.True(t, hasChanges)
			})

			// 测试获取变更文件列表
			t.Run("GetStagedFiles", func(t *testing.T) {
				files, err := vcs.GetStagedFiles(dir)
				require.NoError(t, err)
				assert.Contains(t, files, "test.txt")
			})

			// 测试创建提交
			t.Run("CreateCommit", func(t *testing.T) {
				err := vcs.CreateCommit(dir, "test commit")
				require.NoError(t, err)

				// 验证提交是否成功
				hash, err := vcs.GetLastCommitHash(dir)
				require.NoError(t, err)
				assert.NotEmpty(t, hash)

				info, err := vcs.GetCommitInfo(dir, hash)
				require.NoError(t, err)
				assert.Contains(t, info, "test commit")
			})

			// 测试获取当前分支
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
