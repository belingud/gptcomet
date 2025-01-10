package cmd

import (
	"errors"
	"os"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// 将setupTest改为setupCommitTest
func setupCommitTest(t *testing.T) (string, func()) {
	t.Helper()
	repoPath, cleanup := testutils.TestGitRepo(t)
	origStdin := os.Stdin
	return repoPath, func() {
		cleanup()
		os.Stdin = origStdin
	}
}

func TestNewCommitService(t *testing.T) {
	testCases := []struct {
		name        string
		options     CommitOptions
		mockConfig  *testutils.MockConfigManager
		mockVCS     *testutils.MockVCS
		wantErr     bool
		errContains string
	}{
		{
			name: "success_git",
			options: CommitOptions{
				ConfigPath: "test_config.yaml",
			},
			mockConfig: &testutils.MockConfigManager{
				Data: map[string]interface{}{
					"openai.api_key": "test-key",
				},
			},
			wantErr: false,
		},
		{
			name: "success_svn",
			options: CommitOptions{
				ConfigPath: "test_config.yaml",
				UseSVN:     true,
			},
			mockConfig: &testutils.MockConfigManager{
				Data: map[string]interface{}{
					"openai.api_key": "test-key",
				},
			},
			wantErr: false,
		},
		{
			name: "error_vcs_creation",
			options: CommitOptions{
				ConfigPath: "test_config.yaml",
			},
			mockVCS: &testutils.MockVCS{
				NewFunc: func(vcsType testutils.VCSType) (testutils.VCS, error) {
					return nil, errors.New("mock error")
				},
			},
			wantErr:     true,
			errContains: "failed to create VCS",
		},
		// 添加更多测试用例...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ... 测试实现 ...
		})
	}
}

// 在所有使用setupTest的地方替换为setupCommitTest
func TestCommitService_Execute(t *testing.T) {
	repoPath, cleanup := setupCommitTest(t)
	defer cleanup()

	testCases := []struct {
		name           string
		options        CommitOptions
		setupMocks     func(*testutils.MockVCS, *testutils.MockConfigManager, *testutils.MockLLM)
		input          string
		wantErr        bool
		errContains    string
		checkCommitMsg bool
	}{
		{
			name: "success_auto_yes",
			options: CommitOptions{
				RepoPath: repoPath,
				AutoYes:  true,
			},
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				vcs.HasStagedChangesFunc = func(string) (bool, error) { return true, nil }
				vcs.GetStagedDiffFilteredFunc = func(string, *testutils.MockConfigManager) (string, error) { return "test diff", nil }
				llm.GenerateCommitMessageFunc = func(string, string) (string, error) { return "test commit", nil }
				cfg.GetFunc = func(string) (interface{}, bool) { return "en", true }
			},
			wantErr: false,
		},
		{
			name: "no_staged_changes",
			options: CommitOptions{
				RepoPath: repoPath,
			},
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				vcs.HasStagedChangesFunc = func(string) (bool, error) { return false, nil }
			},
			wantErr:     true,
			errContains: "no staged changes",
		},
		// 添加更多测试用例...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ... 测试实现 ...
		})
	}
}

// 在所有使用setupTest的地方替换为setupCommitTest
func TestCommitService_handleCommitInteraction(t *testing.T) {
	repoPath, cleanup := setupCommitTest(t)
	defer cleanup()

	testCases := []struct {
		name        string
		options     CommitOptions
		input       string
		setupMocks  func(*testutils.MockVCS, *testutils.MockConfigManager, *testutils.MockLLM)
		wantErr     bool
		errContains string
	}{
		{
			name: "confirm_yes",
			options: CommitOptions{
				RepoPath: repoPath,
			},
			input: "y\n",
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				vcs.CreateCommitFunc = func(string, string) error { return nil }
				vcs.GetLastCommitHashFunc = func(string) (string, error) { return "hash", nil }
				vcs.GetCommitInfoFunc = func(string, string) (string, error) { return "info", nil }
			},
			wantErr: false,
		},
		{
			name:  "retry_then_yes",
			input: "r\ny\n",
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				vcs.GetStagedDiffFilteredFunc = func(string, *testutils.MockConfigManager) (string, error) { return "new diff", nil }
				llm.GenerateCommitMessageFunc = func(string, string) (string, error) { return "new message", nil }
				cfg.GetFunc = func(string) (interface{}, bool) { return "en", true }
			},
			wantErr: false,
		},
		{
			name:  "edit_then_yes",
			input: "e\nCustom message\ny\n",
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				vcs.CreateCommitFunc = func(string, string) error { return nil }
				vcs.GetLastCommitHashFunc = func(string) (string, error) { return "hash", nil }
				vcs.GetCommitInfoFunc = func(string, string) (string, error) { return "info", nil }
			},
			wantErr: false,
		},
		// 添加更多测试用例...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ... 测试实现 ...
		})
	}
}

func TestCommitService_generateCommitMessage(t *testing.T) {
	testCases := []struct {
		name        string
		setupMocks  func(*testutils.MockVCS, *testutils.MockConfigManager, *testutils.MockLLM)
		diff        string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "success_english",
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				cfg.GetFunc = func(string) (interface{}, bool) { return "en", true }
				llm.GenerateCommitMessageFunc = func(string, string) (string, error) { return "test commit", nil }
			},
			diff:    "test diff",
			want:    "test commit",
			wantErr: false,
		},
		{
			name: "success_translation",
			setupMocks: func(vcs *testutils.MockVCS, cfg *testutils.MockConfigManager, llm *testutils.MockLLM) {
				cfg.GetFunc = func(string) (interface{}, bool) { return "zh", true }
				llm.GenerateCommitMessageFunc = func(string, string) (string, error) { return "test commit", nil }
				llm.TranslateMessageFunc = func(string, string, string) (string, error) { return "测试提交", nil }
			},
			diff:    "test diff",
			want:    "测试提交",
			wantErr: false,
		},
		// 添加更多测试用例...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ... 测试实现 ...
		})
	}
}

func Test_splitCommitMessage(t *testing.T) {
	testCases := []struct {
		name        string
		message     string
		wantPrefix  string
		wantContent string
	}{
		{
			name:        "with_prefix",
			message:     "feat: add feature",
			wantPrefix:  "feat",
			wantContent: "add feature",
		},
		{
			name:        "no_prefix",
			message:     "add feature",
			wantPrefix:  "",
			wantContent: "add feature",
		},
		{
			name:        "multiple_colons",
			message:     "fix: bug: correct issue",
			wantPrefix:  "fix",
			wantContent: "bug: correct issue",
		},
		// 添加更多测试用例...
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prefix, content := splitCommitMessage(tc.message)
			assert.Equal(t, tc.wantPrefix, prefix)
			assert.Equal(t, tc.wantContent, content)
		})
	}
}
