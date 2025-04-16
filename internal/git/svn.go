package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/belingud/gptcomet/internal/config"
)

// SVNVCS implements the VCS interface for SVN
type SVNVCS struct{}

// GetDiff retrieves the diff of the SVN repository at the specified path.
// It runs the "svn diff" command and returns its output as a string.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//
// Returns:
//   - string: The diff output
//   - error: An error if the SVN command fails or if there are issues accessing the repository
func (s *SVNVCS) GetDiff(repoPath string) (string, error) {
	cmd := exec.Command("svn", "diff")
	return s.runCommand(cmd, repoPath)
}

// HasStagedChanges checks if there are any staged changes in the SVN repository at the given path.
// It runs the "svn status" command and checks if the output contains any lines. If there are lines,
// it returns true, otherwise false.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//
// Returns:
//   - bool: true if there are staged changes, false otherwise
//   - error: nil if the command executed successfully, error otherwise
func (s *SVNVCS) HasStagedChanges(repoPath string) (bool, error) {
	cmd := exec.Command("svn", "status")
	output, err := s.runCommand(cmd, repoPath)
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(output)) > 0, nil
}

// GetStagedFiles returns a list of files that are currently staged for commit in the SVN repository
// at the specified path. It executes the 'svn status' command and filters out lines that start with
// '?' and empty lines. The remaining lines are split after the first 7 characters and the resulting
// paths are returned in a slice.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//
// Returns:
//   - []string: A slice containing the paths of all staged files, or nil if no files are staged
//   - error: An error if the svn command fails or if there are issues accessing the repository
func (s *SVNVCS) GetStagedFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("svn", "status")
	output, err := s.runCommand(cmd, repoPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(output, "\n") {
		if len(line) > 7 && line[0] != '?' {
			files = append(files, strings.TrimSpace(line[7:]))
		}
	}
	return files, nil
}

// GetStagedDiffFiltered returns the diff of staged changes, excluding files that match the patterns
// specified in the config manager under the "file_ignore" key.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//   - cfgManager: The config manager to use for retrieving ignore patterns
//
// Returns:
//   - string: The filtered diff output
//   - error: An error if the svn command fails or if there are issues accessing the repository
func (s *SVNVCS) GetStagedDiffFiltered(repoPath string, cfgManager config.ManagerInterface) (string, error) {
	files, err := s.GetStagedFiles(repoPath)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", nil
	}

	cmd := exec.Command("svn", append([]string{"diff"}, files...)...)
	return s.runCommand(cmd, repoPath)
}

// GetCurrentBranch returns the name of the current branch in the SVN repository
// at the specified path. It runs the "svn info --show-item url" command to get the
// URL of the current branch, and extracts the branch name from it.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//
// Returns:
//   - string: The name of the current branch
//   - error: An error if the SVN command fails or if there are issues accessing the repository
func (s *SVNVCS) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("svn", "info", "--show-item", "url")
	return s.runCommand(cmd, repoPath)
}

// GetCommitInfo returns formatted information about the commit
// If commitHash is empty, returns info about the last commit
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//   - commitHash: The hash of the commit to get info for (or empty for the last commit)
//
// Returns:
//   - string: The formatted commit info
//   - error: An error if the SVN command fails or if there are issues accessing the repository
func (s *SVNVCS) GetCommitInfo(repoPath, commitHash string) (string, error) {
	args := []string{"log", "-r", "HEAD", "-v"}
	if commitHash != "" {
		args = []string{"log", "-r", commitHash, "-v"}
	}
	cmd := exec.Command("svn", args...)
	output, err := s.runCommand(cmd, repoPath)
	if err != nil {
		return "", err
	}

	branch, err := s.GetCurrentBranch(repoPath)
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

// GetLastCommitHash returns the number of the last commit in the SVN repository
// at the specified path.
//
// Parameters:
//   - repoPath: The file system path to the SVN repository
//
// Returns:
//   - string: The number of the last commit (as a string)
//   - error: An error if the SVN command fails or if there are issues accessing the repository
func (s *SVNVCS) GetLastCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("svn", "info", "--show-item", "revision")
	return s.runCommand(cmd, repoPath)
}

// CreateCommit commits changes in the SVN repository with the given message.
// It uses 'svn commit -m <message>' command.
// The skipHook parameter is ignored for SVN.
func (s *SVNVCS) CreateCommit(repoPath, message string, skipHook bool) error {
	cmd := exec.Command("svn", "commit", "-m", message)
	_, err := s.runCommand(cmd, repoPath)
	return err
}

// runCommand executes a given SVN command in the specified repository path and returns its output.
// It captures both stdout and stderr, returning the stdout output as a string if successful.
// If the command fails, it returns an error that includes both the original error and stderr output.
//
// Parameters:
//   - cmd: The prepared exec.Cmd to be executed
//   - repoPath: The directory path where the command should be executed
//
// Returns:
//   - string: The command's stdout output
//   - error: Any error that occurred during command execution
func (s *SVNVCS) runCommand(cmd *exec.Cmd, repoPath string) (string, error) {
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
