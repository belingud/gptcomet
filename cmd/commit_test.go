package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTempConfig(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "gptcomet-test-config")
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(`
provider: openai
openai:
  api_key: test-key
output:
  lang: en
`), 0600)
	require.NoError(t, err)

	return configPath, func() {
		os.RemoveAll(tmpDir)
	}
}

// MockTextEditor implements TextEditor interface for testing
type MockTextEditor struct {
	mock.Mock
}

func (m *MockTextEditor) Edit(initialText string) (string, error) {
	args := m.Called(initialText)
	return args.String(0), args.Error(1)
}

// MockVCS implements VCS interface for testing
type MockVCS struct {
	mock.Mock
}

func (m *MockVCS) HasStagedChanges(repoPath string) (bool, error) {
	args := m.Called(repoPath)
	return args.Bool(0), args.Error(1)
}

func (m *MockVCS) GetStagedDiff() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Update the mock CreateCommit to accept the noverify parameter
func (m *MockVCS) CreateCommit(repoPath, message string, noverify bool) error {
	args := m.Called(repoPath, message, noverify)
	return args.Error(0)
}

func (m *MockVCS) GetLastCommitHash(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetCommitDetails(hash string) (string, error) {
	args := m.Called(hash)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetCommitInfo(repoPath string, hash string) (string, error) {
	args := m.Called(repoPath, hash)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetCurrentBranch(repoPath string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetDiff(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetStagedDiffFiltered(repoPath string, cfgManager config.ManagerInterface) (string, error) {
	args := m.Called(repoPath, cfgManager)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetStagedFiles(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

// MockClient mocks the client.Client interface
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error) {
	args := m.Called(ctx, message, history)
	if resp, ok := args.Get(0).(*types.CompletionResponse); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClient) TranslateMessage(prompt string, message string, lang string) (string, error) {
	args := m.Called(prompt, message, lang)
	return args.String(0), args.Error(1)
}

func (m *MockClient) GenerateCommitMessage(diff string, prompt string) (string, error) {
	args := m.Called(diff, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockClient) GenerateReviewComment(diff string, prompt string) (string, error) {
	args := m.Called(diff, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockClient) GenerateReviewCommentStream(diff string, prompt string, callback func(string) error) error {
	args := m.Called(diff, prompt, callback)
	return args.Error(0)
}

func TestNewCommitService(t *testing.T) {
	testCases := []struct {
		name        string
		options     CommitOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "success_git",
			options: CommitOptions{
				ConfigPath: "test_config.yaml",
			},
			wantErr: false,
		},
		{
			name: "success_svn",
			options: CommitOptions{
				ConfigPath: "test_config.yaml",
				UseSVN:     true,
			},
			wantErr: false,
		},
		{
			name: "invalid_config_path",
			options: CommitOptions{
				ConfigPath: "/nonexistent/config.yaml",
			},
			wantErr:     true,
			errContains: "failed to create config manager",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.wantErr {
				configPath, cleanupConfig := setupTempConfig(t)
				defer cleanupConfig()

				cfg, err := config.New(configPath)
				assert.NoError(t, err)

				err = cfg.Set("openai.api_key", "test-key")
				assert.NoError(t, err)

				tc.options.ConfigPath = configPath
			}

			service, err := NewCommitService(tc.options)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestCommitService_Execute(t *testing.T) {
	testCases := []struct {
		name        string
		setupMocks  func(*MockVCS, *MockTextEditor, *MockClient) (string, string)
		options     CommitOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "no_staged_changes",
			setupMocks: func(vcs *MockVCS, editor *MockTextEditor, client *MockClient) (string, string) {
				vcs.On("HasStagedChanges", mock.Anything).Return(false, nil)
				return "", ""
			},
			wantErr:     true,
			errContains: "no staged changes",
		},
		{
			name: "success_auto_yes_no_skip_hook",
			options: CommitOptions{
				AutoYes:  true,
				NoVerify: false,
			},
			setupMocks: func(vcs *MockVCS, editor *MockTextEditor, client *MockClient) (string, string) {
				diff := "test diff no skip"
				commitMsg := "feat: test commit no skip"

				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return(diff, nil)
				vcs.On("CreateCommit", mock.Anything, commitMsg, false).Return(nil)
				vcs.On("GetLastCommitHash", mock.Anything).Return("abc1234", nil)
				vcs.On("GetCommitInfo", mock.Anything, "abc1234").Return("commit abc1234\nAuthor: Test User\nDate: Thu Jan 1 00:00:00 1970 +0000\n\nfeat: test commit no skip", nil)
				client.On("GenerateCommitMessage", diff, mock.Anything).Return(commitMsg, nil)

				return diff, commitMsg
			},
			wantErr: false,
		},
		{
			name: "success_auto_yes_with_skip_hook",
			options: CommitOptions{
				AutoYes:  true,
				NoVerify: true,
			},
			setupMocks: func(vcs *MockVCS, editor *MockTextEditor, client *MockClient) (string, string) {
				diff := "test diff skip hook"
				commitMsg := "feat: test commit skip hook"

				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return(diff, nil)
				vcs.On("CreateCommit", mock.Anything, commitMsg, true).Return(nil)
				vcs.On("GetLastCommitHash", mock.Anything).Return("def4567", nil)
				vcs.On("GetCommitInfo", mock.Anything, "def4567").Return("commit def4567\nAuthor: Test User\nDate: Fri Jan 2 00:00:00 1970 +0000\n\nfeat: test commit skip hook", nil)
				client.On("GenerateCommitMessage", diff, mock.Anything).Return(commitMsg, nil)

				return diff, commitMsg
			},
			wantErr: false,
		},
		/* // Skip interactive test case as it requires mocking stdin
		{
			name: "success_interactive_yes_no_skip_hook",
			options: CommitOptions{
				SkipHook: false,
			},
			setupMocks: func(vcs *MockVCS, editor *MockTextEditor, client *MockClient) (string, string) {
				diff := "interactive diff no skip"
				commitMsg := "feat: interactive commit no skip"

				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return(diff, nil)
				vcs.On("CreateCommit", mock.Anything, commitMsg, false).Return(nil)
				vcs.On("GetLastCommitHash", mock.Anything).Return("ghi7890", nil)
				vcs.On("GetCommitInfo", mock.Anything, "ghi7890").Return("commit ghi7890\nAuthor: Test User\nDate: Sat Jan 3 00:00:00 1970 +0000\n\nfeat: interactive commit no skip", nil)
				client.On("GenerateCommitMessage", diff, mock.Anything).Return(commitMsg, nil)

				return diff, commitMsg
			},
			wantErr: false,
		},
		*/
		{
			name: "error_getting_diff",
			setupMocks: func(vcs *MockVCS, editor *MockTextEditor, client *MockClient) (string, string) {
				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return("", errors.New("diff error"))
				return "", ""
			},
			wantErr:     true,
			errContains: "diff error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath, cleanupConfig := setupTempConfig(t)
			defer cleanupConfig()

			cfg, err := config.New(configPath)
			assert.NoError(t, err)

			err = cfg.Set("openai.api_key", "test-key")
			assert.NoError(t, err)
			err = cfg.Set("prompt.brief_commit_message", "test prompt")
			assert.NoError(t, err)

			mockVCS := new(MockVCS)
			mockEditor := new(MockTextEditor)
			mockClient := new(MockClient)
			_, _ = tc.setupMocks(mockVCS, mockEditor, mockClient)

			service := &CommitService{
				vcs:          mockVCS,
				client:       mockClient,
				cfgManager:   cfg,
				options:      tc.options,
				editor:       mockEditor,
				clientConfig: &types.ClientConfig{Provider: "test-provider", Model: "test-model"},
			}

			err = service.Execute()

			mockVCS.AssertExpectations(t)
			mockEditor.AssertExpectations(t)
			mockClient.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommitService_generateCommitMessage(t *testing.T) {
	testCases := []struct {
		name        string
		diff        string
		setupConfig func(*config.Manager)
		wantMessage string
		wantErr     bool
		errContains string
	}{
		{
			name: "success_brief",
			diff: "test diff",
			setupConfig: func(cfg *config.Manager) {
				err := cfg.Set("prompt.brief_commit_message", "test prompt")
				assert.NoError(t, err)
			},
			wantMessage: "feat: test commit",
			wantErr:     false,
		},
		{
			name: "success_rich",
			diff: "test diff",
			setupConfig: func(cfg *config.Manager) {
				err := cfg.Set("prompt.rich_commit_message", "test rich prompt")
				assert.NoError(t, err)
			},
			wantMessage: "feat: test commit\n\nDetailed description",
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath, cleanupConfig := setupTempConfig(t)
			defer cleanupConfig()

			cfg, err := config.New(configPath)
			assert.NoError(t, err)

			if tc.setupConfig != nil {
				tc.setupConfig(cfg)
			}

			mockClient := new(MockClient)
			mockClient.On("GenerateCommitMessage", tc.diff, mock.Anything).Return(tc.wantMessage, nil)

			service := &CommitService{
				vcs:        &MockVCS{},
				client:     mockClient,
				cfgManager: cfg,
				options:    CommitOptions{},
			}

			message, err := service.generateCommitMessage(tc.diff)

			mockClient.AssertExpectations(t)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantMessage, message)
			}
		})
	}
}
