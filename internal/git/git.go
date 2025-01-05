package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/belingud/go-gptcomet/internal/config"
	"github.com/belingud/go-gptcomet/internal/debug"
)

// GitVCS implements the VCS interface for Git
type GitVCS struct{}

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

// GetDiff retrieves the staged git diff for the specified repository path.
// It runs the "git diff --staged -U2" command and filters out lines that start with "index", "---", and "+++".
//
// Parameters:
//   - repoPath: The file path to the git repository.
//
// Returns:
//   - A string containing the filtered diff output.
//   - An error if the command fails or if the specified path is not a git repository.
func (g *GitVCS) GetDiff(repoPath string) (string, error) {
	cmd := exec.Command("git", "diff", "--staged", "-U2")
	return g.runCommand(cmd, repoPath)
}

// HasStagedChanges checks if there are any staged changes in the git repository at the given path.
// It runs "git diff --staged --quiet" command and interprets the exit code to determine if there
// are staged changes.
//
// Parameters:
//   - repoPath: The file system path to the git repository
//
// Returns:
//   - bool: true if there are staged changes, false otherwise
//   - error: nil if the command executed successfully, error otherwise
//
// The function returns true if the git diff command exits with code 1 (staged changes present),
// false if it exits with code 0 (no staged changes), and an error for any other exit code or
// if the command fails to execute.
func (g *GitVCS) HasStagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("git", "diff", "--staged", "--quiet")
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means there are staged changes
			if exitError.ExitCode() == 1 {
				return true, nil
			} else {
				return false, fmt.Errorf("git diff command failed with exit code %d: %w\nGit output: %s", exitError.ExitCode(), err, stderr.String())
			}
		}
		return false, fmt.Errorf("failed to check staged changes: %w\nGit output: %s", err, stderr.String())
	}

	// Exit code 0 means no staged changes
	return false, nil
}

// GetStagedFiles returns a list of files that are currently staged for commit in the git repository
// at the specified path. It executes the 'git diff --staged --name-only' command to get the list
// of staged files.
//
// Parameters:
//   - repoPath: The file system path to the git repository
//
// Returns:
//   - []string: A slice containing the paths of all staged files, or nil if no files are staged
//   - error: An error if the git command fails or if there are issues accessing the repository
//
// The function will return (nil, nil) if there are no staged files in the repository.
// If the git command fails, it returns a detailed error message including the exit code.
func (g *GitVCS) GetStagedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--staged", "--name-only")
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff command failed with exit code %d: %w\nGit output: %s", exitError.ExitCode(), err, stderr.String())
		}
		return nil, fmt.Errorf("failed to get staged files: %w\nGit output: %s", err, stderr.String())
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 0 || (len(files) == 1 && files[0] == "") {
		return nil, nil
	}

	return files, nil
}

// ShouldIgnoreFile checks if a file should be ignored based on patterns
func ShouldIgnoreFile(file string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, file)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// GetStagedDiffFiltered returns the git diff for staged changes, excluding files that match the patterns
// specified in the config manager under the "file_ignore" key.
//
// Parameters:
//   - repoPath: The file system path to the git repository
//   - cfgManager: The config manager to use for retrieving ignore patterns
//
// Returns:
//   - string: The filtered diff output
//   - error: An error if the git command fails or if there are issues accessing the repository
//
// The function will return an empty string if there are no staged files in the repository.
// If the git command fails, it returns a detailed error message including the exit code.
func (g *GitVCS) GetStagedDiffFiltered(repoPath string, cfgManager *config.Manager) (string, error) {
	// get staged files
	stagedFiles, err := g.GetStagedFiles(repoPath)
	if err != nil {
		return "", err
	}
	debug.Printf("Staged files: %v", stagedFiles)

	// get ignore patterns
	var ignorePatterns []string
	if patterns, ok := cfgManager.Get("file_ignore"); ok {
		if patternList, ok := patterns.([]interface{}); ok {
			for _, p := range patternList {
				if str, ok := p.(string); ok {
					ignorePatterns = append(ignorePatterns, str)
				}
			}
		}
	}
	debug.Printf("Ignore patterns: %v", ignorePatterns)

	// if no ignore patterns, return the diff as is
	if len(ignorePatterns) == 0 {
		cmd := exec.Command("git", "diff", "--staged", "-U2")
		return g.runCommand(cmd, repoPath)
	}

	// filter out ignored files
	var excludeFiles []string
	for _, file := range stagedFiles {
		if ShouldIgnoreFile(file, ignorePatterns) {
			// git diff --staged -U2 -- :!file
			excludeFiles = append(excludeFiles, ":!"+file)
		}
	}
	debug.Printf("Files to exclude: %v", excludeFiles)

	// return if all staged files are ignored
	if len(excludeFiles) == len(stagedFiles) {
		fmt.Println("All staged files are ignored")
		return "", nil
	}

	// if there are no ignored files, return the diff directly
	if len(excludeFiles) == 0 {
		cmd := exec.Command("git", "diff", "--staged", "-U2")
		return g.runCommand(cmd, repoPath)
	}

	// git diff --staged -U2 -- :!file1 :!file2
	args := []string{"diff", "--staged", "-U2", "--"}
	args = append(args, excludeFiles...)
	debug.Printf("Diff command: git %v", args)

	cmd := exec.Command("git", args...)
	return g.runCommand(cmd, repoPath)
}

// GetCurrentBranch returns the name of the current branch in the git repository
// at the specified path.
//
// Parameters:
//   - repoPath: The file system path to the git repository
//
// Returns:
//   - string: The name of the current branch
//   - error: An error if the git command fails or if there are issues accessing the repository
func (g *GitVCS) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := g.runCommand(cmd, repoPath)
	return strings.TrimSpace(output), err
}

// GetCommitInfo returns formatted information about the commit
// If commitHash is empty, returns info about the last commit
//
// Parameters:
//   - repoPath: The file system path to the git repository
//   - commitHash: The hash of the commit to get info for (or empty for the last commit)
//
// Returns:
//   - string: The formatted commit info
//   - error: An error if the git command fails or if there are issues accessing the repository
func (g *GitVCS) GetCommitInfo(repoPath string, commitHash string) (string, error) {
	if commitHash == "" {
		// Get last commit hash
		hash, err := g.GetLastCommitHash(repoPath)
		if err != nil {
			return "", err
		}
		commitHash = hash
	}
	commitHash = strings.TrimSpace(commitHash)

	cmd := exec.Command("git", "log", "-1", "--stat",
		"--pretty=format:Author: %an <%ae>%n%D(%H)%n%n%s%n",
		commitHash)
	output, err := g.runCommand(cmd, repoPath)
	if err != nil {
		return "", err
	}
	branch, err := g.GetCurrentBranch(repoPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(output, "\n")
	if len(lines) > 1 {
		// Replace the second line (which contains ref info) with just the branch name
		lines[1] = strings.Split(lines[1], "(")[0] + lines[1][strings.LastIndex(lines[1], "("):]
		lines[1] = branch + lines[1][strings.LastIndex(lines[1], "("):]

		// Add colors to the stats
		for i := 4; i < len(lines); i++ {
			line := lines[i]
			if strings.Contains(line, "|") {
				parts := strings.Split(line, "|")
				if len(parts) == 2 {
					stats := strings.TrimSpace(parts[1])
					coloredStats := strings.ReplaceAll(stats, "+", colorGreen+"+")
					coloredStats = strings.ReplaceAll(coloredStats, "-", colorReset+colorRed+"-")
					lines[i] = parts[0] + "| " + coloredStats + colorReset
				}
			}
		}
		output = strings.Join(lines, "\n")
	}
	return output, nil
}

// GetLastCommitHash returns the hash of the last commit
// Parameters:
//   - repoPath: The file system path to the git repository
//
// Returns:
//   - string: The hash of the last commit
//   - error: An error if the git command fails or if there are issues accessing the repository
func (g *GitVCS) GetLastCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	return g.runCommand(cmd, repoPath)
}

// CreateCommit creates a git commit with the given message
//
// Parameters:
//   - repoPath: The file system path to the git repository
//   - message: The commit message
//
// Returns:
//   - error: An error if the git command fails or if there are issues accessing the repository
func (g *GitVCS) CreateCommit(repoPath string, message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	_, err := g.runCommand(cmd, repoPath)
	return err
}

// runCommand 执行命令并返回输出
func (g *GitVCS) runCommand(cmd *exec.Cmd, repoPath string) (string, error) {
	debug.Printf("Running command: %v", cmd.Args)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
