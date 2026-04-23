package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient mocks the HTTP client for testing
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	if resp, ok := args.Get(0).(*http.Response); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

// MockFileSystem mocks the file system for testing
type MockFileSystem struct {
	mock.Mock
}

func (m *MockFileSystem) MkdirTemp(dir, pattern string) (string, error) {
	args := m.Called(dir, pattern)
	return args.String(0), args.Error(1)
}

func (m *MockFileSystem) RemoveAll(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockFileSystem) UserHomeDir() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

func (m *MockFileSystem) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	args := m.Called(oldpath, newpath)
	return args.Error(0)
}

func (m *MockFileSystem) Symlink(oldname, newname string) error {
	args := m.Called(oldname, newname)
	return args.Error(0)
}

// MockDownloader mocks the downloader for testing
type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) Download(url, dst string) error {
	args := m.Called(url, dst)
	return args.Error(0)
}

// MockExtractor mocks the extractor for testing
type MockExtractor struct {
	mock.Mock
}

func (m *MockExtractor) Extract(src, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

// MockFileCopier mocks the file copier for testing
type MockFileCopier struct {
	mock.Mock
}

func (m *MockFileCopier) Copy(src, dst string) error {
	args := m.Called(src, dst)
	return args.Error(0)
}

type trackingHTTPClient struct {
	called   bool
	response *http.Response
	err      error
}

func (c *trackingHTTPClient) Get(url string) (*http.Response, error) {
	c.called = true
	return c.response, c.err
}

func resetUpdateGlobals(t *testing.T) {
	t.Helper()

	oldInstallationSource := InstallationSource
	oldExecutablePath := executablePath
	oldEvalSymlinks := evalSymlinks
	oldDefaultHTTPClient := DefaultHTTPClient

	t.Cleanup(func() {
		InstallationSource = oldInstallationSource
		executablePath = oldExecutablePath
		evalSymlinks = oldEvalSymlinks
		DefaultHTTPClient = oldDefaultHTTPClient
	})
}

func TestHomebrewInstallationDetectionBySource(t *testing.T) {
	resetUpdateGlobals(t)

	InstallationSource = installSourceHomebrew
	executablePath = func() (string, error) {
		return "/home/user/.local/bin/gmsg", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return path, nil
	}

	assert.True(t, isHomebrewInstallation())
}

func TestHomebrewInstallationDetectionByResolvedPath(t *testing.T) {
	resetUpdateGlobals(t)

	InstallationSource = installSourceStandalone
	executablePath = func() (string, error) {
		return "/opt/homebrew/bin/gmsg", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return "/opt/homebrew/Cellar/gptcomet/2.4.1/bin/gmsg", nil
	}

	assert.True(t, isHomebrewInstallation())
}

func TestHomebrewCellarPathRequiresProjectSegment(t *testing.T) {
	assert.True(t, isHomebrewCellarPath("/usr/local/Cellar/gptcomet/2.4.1/bin/gptcomet"))
	assert.False(t, isHomebrewCellarPath("/usr/local/Cellar/other/2.4.1/bin/gptcomet"))
}

func TestNewUpdateCmdRejectsHomebrewBeforeReleaseCheck(t *testing.T) {
	resetUpdateGlobals(t)

	InstallationSource = installSourceHomebrew
	client := &trackingHTTPClient{}
	DefaultHTTPClient = client

	err := NewUpdateCmd("1.0.0").Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "brew upgrade gptcomet")
	assert.False(t, client.called)
}

func TestNewUpdateCmdAllowsStandaloneUpdateFlow(t *testing.T) {
	resetUpdateGlobals(t)

	InstallationSource = installSourceStandalone
	executablePath = func() (string, error) {
		return "/home/user/.local/bin/gmsg", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return path, nil
	}

	responseBody, _ := json.Marshal(&GithubRelease{TagName: "v1.0.0"})
	client := &trackingHTTPClient{
		response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		},
	}
	DefaultHTTPClient = client

	err := NewUpdateCmd("1.0.0").Execute()

	assert.NoError(t, err)
	assert.True(t, client.called)
}

// TestCheckUpdate tests the checkUpdate function
func TestCheckUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		mockResponse   *GithubRelease
		mockError      error
		expectError    bool
	}{
		{
			name:           "Same version - no update needed",
			currentVersion: "1.0.0",
			mockResponse: &GithubRelease{
				TagName: "v1.0.0",
			},
			expectError: false,
		},
		{
			name:           "API error",
			currentVersion: "1.0.0",
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)

			// Setup mock response
			if tt.mockError != nil {
				mockClient.On("Get", mock.Anything).Return(nil, tt.mockError)
			} else {
				responseBody, _ := json.Marshal(tt.mockResponse)
				mockClient.On("Get", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(responseBody)),
				}, nil)
			}

			err := CheckUpdateWithClient(tt.currentVersion, mockClient)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestInstallUpdate tests the installUpdate function
func TestInstallUpdate(t *testing.T) {
	mockFS := new(MockFileSystem)
	mockClient := new(MockHTTPClient)
	mockDownloader := new(MockDownloader)
	mockExtractor := new(MockExtractor)
	mockCopier := new(MockFileCopier)

	tests := []struct {
		name        string
		version     string
		tag         string
		setupMocks  func(tempDir string)
		expectError bool
	}{
		{
			name:    "Success - Unix system",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				// Mock file system operations
				mockFS.On("MkdirTemp", "", mock.Anything).Return(tempDir, nil)
				mockFS.On("RemoveAll", tempDir).Return(nil)
				mockFS.On("UserHomeDir").Return("/home/user", nil)
				mockFS.On("MkdirAll", mock.Anything, os.FileMode(0755)).Return(nil)
				mockFS.On("Remove", mock.Anything).Return(nil)
				mockFS.On("Rename", mock.Anything, mock.Anything).Return(nil)
				mockFS.On("Symlink", mock.Anything, mock.Anything).Return(nil)

				// Mock downloader
				mockDownloader.On("Download", mock.Anything, mock.Anything).Return(nil)

				// Mock extractor
				mockExtractor.On("Extract", mock.Anything, mock.Anything).Return(nil)

				// Mock copier
				mockCopier.On("Copy", mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "Error - MkdirTemp fails",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				mockFS.On("MkdirTemp", "", mock.Anything).Return("", assert.AnError)
			},
			expectError: true,
		},
		{
			name:    "Error - UserHomeDir fails",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				mockFS.On("MkdirTemp", "", mock.Anything).Return(tempDir, nil)
				mockFS.On("RemoveAll", tempDir).Return(nil)
				mockFS.On("UserHomeDir").Return("", assert.AnError)

				// Mock downloader
				mockDownloader.On("Download", mock.Anything, mock.Anything).Return(nil)

				// Mock extractor
				mockExtractor.On("Extract", mock.Anything, mock.Anything).Return(nil)
			},
			expectError: true,
		},
		{
			name:    "Error - Download fails",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				mockFS.On("MkdirTemp", "", mock.Anything).Return(tempDir, nil)
				mockFS.On("RemoveAll", tempDir).Return(nil)
				mockDownloader.On("Download", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectError: true,
		},
		{
			name:    "Error - Extract fails",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				mockFS.On("MkdirTemp", "", mock.Anything).Return(tempDir, nil)
				mockFS.On("RemoveAll", tempDir).Return(nil)
				mockDownloader.On("Download", mock.Anything, mock.Anything).Return(nil)
				mockExtractor.On("Extract", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectError: true,
		},
		{
			name:    "Error - Copy fails",
			version: "1.1.0",
			tag:     "v1.1.0",
			setupMocks: func(tempDir string) {
				mockFS.On("MkdirTemp", "", mock.Anything).Return(tempDir, nil)
				mockFS.On("RemoveAll", tempDir).Return(nil)
				mockDownloader.On("Download", mock.Anything, mock.Anything).Return(nil)
				mockExtractor.On("Extract", mock.Anything, mock.Anything).Return(nil)
				mockFS.On("UserHomeDir").Return("/home/user", nil)
				mockFS.On("MkdirAll", mock.Anything, os.FileMode(0755)).Return(nil)
				mockCopier.On("Copy", mock.Anything, mock.Anything).Return(assert.AnError)
				mockFS.On("Remove", mock.Anything).Return(nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockFS = new(MockFileSystem)
			mockClient = new(MockHTTPClient)
			mockDownloader = new(MockDownloader)
			mockExtractor = new(MockExtractor)
			mockCopier = new(MockFileCopier)

			// Setup mocks with a fixed temp dir
			tempDir := "/tmp/gptcomet-test"
			tt.setupMocks(tempDir)

			// Run test
			err := InstallUpdateWithAllDeps(tt.version, tt.tag, mockClient, mockFS, mockDownloader, mockExtractor, mockCopier)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expected calls were made
			mockFS.AssertExpectations(t)
			mockClient.AssertExpectations(t)
			mockDownloader.AssertExpectations(t)
			mockExtractor.AssertExpectations(t)
			mockCopier.AssertExpectations(t)
		})
	}
}
