package testutils

import (
	"github.com/stretchr/testify/mock"
)

// MockVCS is a mock implementation of VCS interface
type MockVCS struct {
	mock.Mock
}

func (m *MockVCS) HasStagedChanges(repoPath string) (bool, error) {
	args := m.Called(repoPath)
	return args.Bool(0), args.Error(1)
}

// CreateCommit simulates creating a commit, ignoring the skipHook parameter for the mock
func (m *MockVCS) CreateCommit(repoPath string, message string, skipHook bool) error {
	args := m.Called(repoPath, message, skipHook)
	return args.Error(0)
}

func (m *MockVCS) GetLastCommitHash(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetCommitInfo(repoPath string, hash string) (string, error) {
	args := m.Called(repoPath, hash)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetDiff(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetStagedFiles(repoPath string) ([]string, error) {
	args := m.Called(repoPath)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockVCS) GetFileContent(repoPath string, file string) (string, error) {
	args := m.Called(repoPath, file)
	return args.String(0), args.Error(1)
}

func (m *MockVCS) GetCurrentBranch(repoPath string) (string, error) {
	args := m.Called(repoPath)
	return args.String(0), args.Error(1)
}
