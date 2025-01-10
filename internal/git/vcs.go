package git

import "github.com/belingud/gptcomet/internal/config"

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
	GetStagedDiffFiltered(repoPath string, cfgManager config.ManagerInterface) (string, error)
	GetCurrentBranch(repoPath string) (string, error)
	GetCommitInfo(repoPath, commitHash string) (string, error)
	GetLastCommitHash(repoPath string) (string, error)
	CreateCommit(repoPath, message string) error
}

// NewVCS creates a new VCS object based on the given type.
//
// Parameters:
//   - vcsType: The type of VCS to create
//
// Returns:
//   - A VCS object of the specified type
//   - An error if the type is not recognized
func NewVCS(vcsType VCSType) (VCS, error) {
	switch vcsType {
	case Git:
		return &GitVCS{}, nil
	case SVN:
		return &SVNVCS{}, nil
	default:
		return &GitVCS{}, nil
	}
}
