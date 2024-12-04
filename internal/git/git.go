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
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	return out.String(), nil
}

// CreateCommit creates a git commit with the given message
func CreateCommit(repoPath string, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %s, %w", stderr.String(), err)
	}

	return nil
}

// HasStagedChanges checks if there are any staged changes
func HasStagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	cmd.Dir = repoPath

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means there are staged changes
			if exitError.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("failed to check staged changes: %w", err)
	}

	// Exit code 0 means no staged changes
	return false, nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(out.String()), nil
}

// GetCommitInfo returns formatted information about the last commit
func GetCommitInfo(repoPath string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--stat", "--pretty=format:Author: %an <%ae>%n%D(%H)%n%n%s%n")
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get commit info: %w", err)
	}

	// Get branch name
	branch, err := GetCurrentBranch(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}

	// Replace the ref info with just the branch name
	output := out.String()
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		// Replace the second line (which contains ref info) with just the branch name
		lines[1] = strings.Split(lines[1], "(")[0] + lines[1][strings.LastIndex(lines[1], "("):]
		lines[1] = branch + lines[1][strings.LastIndex(lines[1], "("):]
		output = strings.Join(lines, "\n")
	}

	return output, nil
}
