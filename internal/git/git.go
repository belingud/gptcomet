package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetDiff returns the git diff for staged changes
func GetDiff(repoPath string) (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	cmd.Dir = repoPath
	
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	
	return out.String(), nil
}

// Commit commits the staged changes with the given message
func Commit(repoPath, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	
	return nil
}

// GetRepoPath returns the git repository root path
func GetRepoPath(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	
	return strings.TrimSpace(out.String()), nil
}

// HasStagedChanges checks if there are any staged changes
func HasStagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	cmd.Dir = repoPath
	
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means there are staged changes
			return exitErr.ExitCode() == 1, nil
		}
		return false, fmt.Errorf("failed to check staged changes: %w", err)
	}
	
	// Exit code 0 means no staged changes
	return false, nil
}
