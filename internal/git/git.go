package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Repo represents a git repository
type Repo struct {
	path string
}

// Commit represents a git commit
type Commit struct {
	hash    string
	message string
	author  string
	email   string
	repo    *Repo
}

// NewRepo creates a new Repo instance
func NewRepo(path string) (*Repo, error) {
	return &Repo{path: path}, nil
}

// Branch returns the current branch name
func (r *Repo) Branch() string {
	branch, err := GetCurrentBranch(r.path)
	if err != nil {
		return "unknown"
	}
	return branch
}

// Commit creates a new commit with the given message
func (r *Repo) Commit(message string) (*Commit, error) {
	if err := CreateCommit(r.path, message); err != nil {
		return nil, err
	}

	// Get the commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.path
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	hash := strings.TrimSpace(out.String())

	// Get author and email
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%an|%ae")
	cmd.Dir = r.path
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get author info: %w", err)
	}
	parts := strings.Split(strings.TrimSpace(out.String()), "|")
	author := parts[0]
	email := ""
	if len(parts) > 1 {
		email = parts[1]
	}

	return &Commit{
		hash:    hash,
		message: message,
		author:  author,
		email:   email,
		repo:    r,
	}, nil
}

// ShowStat returns the git show --stat output for a commit
func (r *Repo) ShowStat(hash string) string {
	cmd := exec.Command("git", "show", "--pretty=format:", "--stat", hash)
	cmd.Dir = r.path
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

// Hash returns the commit hash
func (c *Commit) Hash() string {
	return c.hash
}

// Message returns the commit message
func (c *Commit) Message() string {
	return c.message
}

// Author returns the commit author
func (c *Commit) Author() string {
	return c.author
}

// Email returns the author's email
func (c *Commit) Email() string {
	return c.email
}

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
