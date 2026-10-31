// Package secretscan detects credentials in added diff lines before the diff is
// sent to an LLM provider.
//
// The rule data is vendored from gitleaks (MIT, see rules/UPSTREAM.md) and
// matched by a small local engine. The scanner only ever looks at added lines:
// it never modifies the user's commit, only the text that would be transmitted.
package secretscan

import (
	"fmt"
	"regexp"
	"strings"
)

// Mode controls what the scanner does with a finding.
type Mode string

const (
	// ModeOff disables scanning entirely.
	ModeOff Mode = "off"
	// ModeWarn reports findings and sends the diff unchanged.
	ModeWarn Mode = "warn"
	// ModeRedact replaces the finding spans in the transmitted diff.
	ModeRedact Mode = "redact"
	// ModeBlock sends nothing and returns a structured error.
	ModeBlock Mode = "block"
)

// ParseMode validates a user-provided mode value.
func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	switch mode {
	case ModeOff, ModeWarn, ModeRedact, ModeBlock:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid sensitive scan mode %q (expected off, warn, redact or block)", value)
	}
}

// Options controls one scan.
type Options struct {
	Mode Mode
	// Allow silences findings whose secret, match, or full line matches one of
	// these user-configured regular expressions (config key sensitive.allow).
	Allow []*regexp.Regexp
}

// CompileAllow compiles the sensitive.allow config patterns.
func CompileAllow(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if strings.TrimSpace(pattern) == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid sensitive.allow pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

// Finding is one detected secret. Secret always holds the raw value and must
// never be rendered to the user without Preview/maskSecret.
type Finding struct {
	RuleID    string
	Path      string
	Line      int // 1-based line number in the new file (0 for file-level findings)
	Column    int // 1-based byte column in the added line content
	FileLevel bool
	Secret    string
	LineText  string
}

// Preview renders a masked, user-safe description of the finding.
func (f Finding) Preview() string {
	if f.FileLevel || f.Secret == "" {
		return "(file-level)"
	}
	return maskSecret(f.Secret)
}

// maskSecret hides most of a secret. Secrets of 8 runes or fewer are fully
// masked; longer ones keep at most two leading and trailing runes.
func maskSecret(secret string) string {
	runes := []rune(secret)
	switch {
	case len(runes) == 0:
		return ""
	case len(runes) <= 8:
		return strings.Repeat("*", len(runes))
	case len(runes) <= 16:
		return string(runes[:2]) + "…" + string(runes[len(runes)-2:])
	default:
		return string(runes[:4]) + "…" + string(runes[len(runes)-4:])
	}
}

// Result is the outcome of Apply.
type Result struct {
	// Diff is the text that should be sent to the provider (possibly redacted).
	Diff string
	// Findings is always populated, even in block mode, so callers can warn.
	Findings []Finding
}
