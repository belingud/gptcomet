package testutils

import (
	"errors"
)

var ErrMock = errors.New("mock error")

// VCSType represents the type of version control system
type VCSType string

const (
	Git VCSType = "git"
	SVN VCSType = "svn"
)

// VCS defines the interface for version control operations
type VCS interface {
	GetDiff(repoPath string) (string, error)
	HasStagedChanges(repoPath string) (bool, error)
	GetStagedFiles(repoPath string) ([]string, error)
	GetStagedDiffFiltered(repoPath string, cfgManager *MockConfigManager) (string, error)
	GetCurrentBranch(repoPath string) (string, error)
	GetCommitInfo(repoPath, commitHash string) (string, error)
	GetLastCommitHash(repoPath string) (string, error)
	CreateCommit(repoPath, message string) error
}

type MockVCS struct {
	TypeFunc                  func() VCSType
	NewFunc                   func(vcsType VCSType) (VCS, error)
	HasStagedChangesFunc      func(repoPath string) (bool, error)
	GetStagedDiffFilteredFunc func(repoPath string, cfgManager *MockConfigManager) (string, error)
	CreateCommitFunc          func(repoPath string, message string) error
	GetLastCommitHashFunc     func(repoPath string) (string, error)
	GetCommitInfoFunc         func(repoPath string, commitHash string) (string, error)
}

func (m *MockVCS) Type() VCSType {
	if m.TypeFunc != nil {
		return m.TypeFunc()
	}
	return Git
}

func (m *MockVCS) HasStagedChanges(repoPath string) (bool, error) {
	if m.HasStagedChangesFunc != nil {
		return m.HasStagedChangesFunc(repoPath)
	}
	return false, nil
}

func (m *MockVCS) GetStagedDiffFiltered(repoPath string, cfgManager *MockConfigManager) (string, error) {
	if m.GetStagedDiffFilteredFunc != nil {
		return m.GetStagedDiffFilteredFunc(repoPath, cfgManager)
	}
	return "", nil
}

func (m *MockVCS) CreateCommit(repoPath string, message string) error {
	if m.CreateCommitFunc != nil {
		return m.CreateCommitFunc(repoPath, message)
	}
	return nil
}

func (m *MockVCS) GetLastCommitHash(repoPath string) (string, error) {
	if m.GetLastCommitHashFunc != nil {
		return m.GetLastCommitHashFunc(repoPath)
	}
	return "", nil
}

func (m *MockVCS) GetCommitInfo(repoPath string, commitHash string) (string, error) {
	if m.GetCommitInfoFunc != nil {
		return m.GetCommitInfoFunc(repoPath, commitHash)
	}
	return "", nil
}
