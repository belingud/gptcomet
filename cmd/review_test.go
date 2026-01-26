package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewReviewService(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockVCS, *testutils.MockConfigManager)
		wantErr     bool
		errContains string
	}{
		{
			name: "success_git",
			setupMocks: func(vcs *MockVCS, cfg *testutils.MockConfigManager) {
				cfg.On("GetClientConfig").Return(&types.ClientConfig{}, nil)
				cfg.On("Get", "openai.api_key").Return("dummy-key", true)
				cfg.On("Get", "output.lang").Return("en", true)
			},
			wantErr: false,
		},
		{
			name:        "invalid_config_path",
			setupMocks:  func(vcs *MockVCS, cfg *testutils.MockConfigManager) {},
			wantErr:     true,
			errContains: "Dependency Creation Failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVCS := new(MockVCS)
			mockCfg := new(testutils.MockConfigManager)
			tt.setupMocks(mockVCS, mockCfg)

			var options ReviewOptions
			if tt.name == "invalid_config_path" {
				options = ReviewOptions{
					RepoPath:   "test-repo",
					ConfigPath: "/invalid/path",
				}
			} else {
				configPath, cleanup := setupTempConfig(t)
				defer cleanup()
				options = ReviewOptions{
					RepoPath:   "test-repo",
					ConfigPath: configPath,
				}
			}

			service, err := NewReviewService(options)
			fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>err = %v", err)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, options, service.options)
				assert.IsType(t, &GlamourRenderer{}, service.markdownRenderer)
			}
		})
	}
}

func TestReviewService_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockVCS, *testutils.MockConfigManager, *MockClient)
		wantErr     bool
		errContains string
		isPipeInput bool
	}{
		{
			name: "success_with_staged_changes",
			setupMocks: func(vcs *MockVCS, cfg *testutils.MockConfigManager, client *MockClient) {
				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return("test-diff", nil)
				cfg.On("GetReviewPrompt").Return("test-prompt")
				cfg.On("Get", REVIEW_LANG_KEY).Return("en", true)
				cfg.On("GetWithDefault", "output.markdown_theme", mock.Anything).Return("auto")
				cfg.On("GetNestedValue", []string{"console", "verbose"}).Return(false, true)
				client.On("GenerateReviewComment", "test-diff", "test-prompt").Return("test-comment", nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return("staged-diff", nil)
			},
			wantErr:     false,
			isPipeInput: false,
		},
		{
			name: "no_staged_changes",
			setupMocks: func(vcs *MockVCS, cfg *testutils.MockConfigManager, client *MockClient) {
				vcs.On("HasStagedChanges", mock.Anything).Return(false, nil)
				cfg.On("GetNestedValue", []string{"console", "verbose"}).Return(false, true)
			},
			wantErr:     true,
			errContains: "no staged changes found",
			isPipeInput: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockVCS := new(MockVCS)
			mockCfg := new(testutils.MockConfigManager)
			mockClient := new(MockClient)
			tt.setupMocks(mockVCS, mockCfg, mockClient)

			service := &ReviewService{
				vcs:        mockVCS,
				client:     mockClient,
				cfgManager: mockCfg,
				options: ReviewOptions{
					RepoPath: "test-repo",
				},
				markdownRenderer: &GlamourRenderer{},
				clientConfig:     &types.ClientConfig{Provider: "test-provider", Model: "test-model"},
			}

			err := service.Execute()

			mockVCS.AssertExpectations(t)
			mockCfg.AssertExpectations(t)
			mockClient.AssertExpectations(t)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReviewService_getDiff(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockVCS)
		pipeInput   bool
		wantDiff    string
		wantErr     bool
		errContains string
	}{
		{
			name: "success_piped_input",
			setupMocks: func(vcs *MockVCS) {
				// No VCS calls expected for piped input
			},
			pipeInput: true,
			wantDiff:  "piped-diff",
			wantErr:   false,
		},
		{
			name: "success_staged_changes",
			setupMocks: func(vcs *MockVCS) {
				vcs.On("HasStagedChanges", mock.Anything).Return(true, nil)
				vcs.On("GetStagedDiffFiltered", mock.Anything, mock.Anything).Return("staged-diff", nil)
			},
			pipeInput: false,
			wantDiff:  "staged-diff",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVCS := new(MockVCS)
			tt.setupMocks(mockVCS)

			service := &ReviewService{
				vcs: mockVCS,
				options: ReviewOptions{
					RepoPath: "test-repo",
				},
			}

			// Mock pipe input if needed
			if tt.pipeInput {
				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()
				defer w.Close()
				errChan := make(chan error)
				go func() {
					_, err := io.Copy(w, strings.NewReader(tt.wantDiff))
					if err != nil {
						errChan <- err
					} else {
						errChan <- nil
					}
					w.Close()
				}()
				if err := <-errChan; err != nil {
					t.Fatal(err)
				}
				os.Stdin = r
			}

			diff, err := service.getDiff()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDiff, diff)
			}
		})
	}
}

func TestReviewService_generateReviewComment(t *testing.T) {
	tests := []struct {
		name        string
		diff        string
		setupMocks  func(*testutils.MockConfigManager, *MockClient)
		wantComment string
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			diff: "test-diff",
			setupMocks: func(cfg *testutils.MockConfigManager, client *MockClient) {
				cfg.On("GetReviewPrompt").Return("test-prompt")
				cfg.On("Get", REVIEW_LANG_KEY).Return("en", true)
				client.On("GenerateReviewComment", "test-diff", "test-prompt").Return("test-comment", nil)
			},
			wantComment: "test-comment",
			wantErr:     false,
		},
		{
			name:        "empty_diff",
			diff:        "",
			setupMocks:  func(cfg *testutils.MockConfigManager, client *MockClient) {},
			wantErr:     true,
			errContains: "empty diff provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCfg := new(testutils.MockConfigManager)
			mockClient := new(MockClient)
			tt.setupMocks(mockCfg, mockClient)

			service := &ReviewService{
				client:     mockClient,
				cfgManager: mockCfg,
			}

			comment, err := service.generateReviewComment(tt.diff)

			mockCfg.AssertExpectations(t)
			mockClient.AssertExpectations(t)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantComment, comment)
			}
		})
	}
}

func TestReviewService_formatReviewComment(t *testing.T) {
	tests := []struct {
		name          string
		comment       string
		setupMocks    func(*testutils.MockConfigManager)
		wantFormatted string
		wantErr       bool
	}{
		{
			name:    "success",
			comment: "test-comment",
			setupMocks: func(cfg *testutils.MockConfigManager) {
				cfg.On("GetWithDefault", MARKDOWN_THEME, mock.Anything).Return("auto")
			},
			wantFormatted: "test-comment", // Actual formatting would depend on Glamour
			wantErr:       false,
		},
		{
			name:    "render_error",
			comment: "test-comment",
			setupMocks: func(cfg *testutils.MockConfigManager) {
				cfg.On("GetWithDefault", MARKDOWN_THEME, mock.Anything).Return("invalid-style")
			},
			wantFormatted: "test-comment", // Should return original on error
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCfg := new(testutils.MockConfigManager)
			tt.setupMocks(mockCfg)

			service := &ReviewService{
				cfgManager:       mockCfg,
				markdownRenderer: &GlamourRenderer{},
			}

			formatted, err := service.formatReviewComment(tt.comment)

			mockCfg.AssertExpectations(t)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Contains(t, formatted, tt.wantFormatted)
		})
	}
}
