package hook

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/belingud/gptcomet/internal/git"
)

// RunInput describes one prepare-commit-msg invocation.
type RunInput struct {
	MessageFile string
	Source      string
	RepoPath    string
	ConfigPath  string
}

// Run applies the hook skip rules, generates a message, and writes it to the
// commit message file. Errors are returned for the command layer to swallow;
// callers must treat every error as non-fatal so commits are never blocked.
func (m *Manager) Run(ctx context.Context, in RunInput) error {
	if isTruthy(os.Getenv("GMSG_SKIP")) {
		return nil
	}

	force := isTruthy(os.Getenv("GMSG_FORCE"))
	if !force && isSkippedSource(in.Source) {
		return nil
	}

	content, err := os.ReadFile(in.MessageFile)
	if err != nil {
		return fmt.Errorf("read commit message file: %w", err)
	}
	original := string(content)
	if !force && hasNonCommentContent(original) {
		return nil
	}

	repoPath := strings.TrimSpace(in.RepoPath)
	if repoPath == "" {
		repoPath = "."
	}
	repoRoot, err := git.GetRepoRoot(repoPath)
	if err != nil {
		return err
	}

	message, err := m.generate(ctx, repoRoot, in.ConfigPath)
	if err != nil {
		return err
	}
	message = strings.TrimRight(message, "\r\n")
	if strings.TrimSpace(message) == "" {
		return nil
	}

	preserved := original
	if force {
		preserved = commentLines(original)
	}

	var b strings.Builder
	b.WriteString(message)
	b.WriteString("\n")
	if strings.TrimSpace(preserved) != "" {
		b.WriteString("\n")
		b.WriteString(strings.TrimRight(preserved, "\r\n"))
		b.WriteString("\n")
	}

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(in.MessageFile); statErr == nil {
		mode = info.Mode().Perm()
	}

	if err := os.WriteFile(in.MessageFile, []byte(b.String()), mode); err != nil {
		return fmt.Errorf("write commit message file: %w", err)
	}
	return nil
}

func (m *Manager) generate(ctx context.Context, repoRoot, configPath string) (string, error) {
	if m.Generate == nil {
		return "", fmt.Errorf("hook generator is not configured")
	}
	return m.Generate(ctx, repoRoot, configPath)
}

func isSkippedSource(source string) bool {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "merge", "squash", "commit", "message":
		return true
	default:
		return false
	}
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func hasNonCommentContent(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		return true
	}
	return false
}

func commentLines(content string) string {
	lines := make([]string, 0)
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			lines = append(lines, line)
		}
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\r\n")
}
