package git

import (
	"os/exec"
	"path/filepath"
	"strings"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
)

// GetRepoRoot resolves the working-tree root for repoPath.
//
// It is used by the hook runtime because Git invokes hooks from the repository
// root, but a user can also call `gmsg hook run` manually from a subdirectory.
func GetRepoRoot(repoPath string) (string, error) {
	out, err := runGitCommand(repoPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	root := strings.TrimSpace(out)
	if root == "" {
		return "", gptcometerrors.GitRepositoryNotFoundError()
	}

	return filepath.Clean(root), nil
}

// GetHooksDir resolves the hooks directory for repoPath. Asking Git instead of
// assuming `.git/hooks` keeps `core.hooksPath` and linked worktrees working.
func GetHooksDir(repoPath string) (string, error) {
	out, err := runGitCommand(repoPath, "rev-parse", "--path-format=absolute", "--git-path", "hooks")
	if err != nil {
		// --path-format was added in Git 2.31. Fall back to a relative path
		// and resolve it against the working-tree root.
		out, err = runGitCommand(repoPath, "rev-parse", "--git-path", "hooks")
		if err != nil {
			return "", err
		}
	}

	hooksDir := strings.TrimSpace(out)
	if hooksDir == "" {
		return "", gptcometerrors.GitRepositoryNotFoundError()
	}
	if !filepath.IsAbs(hooksDir) {
		root, rootErr := GetRepoRoot(repoPath)
		if rootErr != nil {
			return "", rootErr
		}
		hooksDir = filepath.Join(root, hooksDir)
	}

	return filepath.Clean(hooksDir), nil
}

func runGitCommand(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	return (&GitVCS{}).runCommand(cmd, repoPath)
}
