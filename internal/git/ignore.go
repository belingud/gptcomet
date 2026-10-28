package git

import (
	"path"
	"strings"
)

// ShouldIgnoreFile reports whether file matches any of the supplied
// gitignore-style patterns.
//
// Supported syntax:
//   - "name" matches a file or directory name at any depth
//   - "dir/file" is anchored to the repository root
//   - "/name" is anchored to the repository root explicitly
//   - "dir/" matches a directory and everything below it
//   - "*", "?", "[]" match within a single path segment
//   - "**" matches zero or more path segments
//   - a leading "!" re-includes a file matched by an earlier pattern
//
// Patterns are evaluated in order, so the last matching pattern wins.
func ShouldIgnoreFile(file string, ignorePatterns []string) bool {
	normalized := normalizeIgnorePath(file)
	if normalized == "" {
		return false
	}

	ignored := false
	for _, raw := range ignorePatterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		negated := strings.HasPrefix(pattern, "!")
		if negated {
			pattern = strings.TrimSpace(strings.TrimPrefix(pattern, "!"))
		}
		if pattern == "" {
			continue
		}

		if matchIgnorePattern(pattern, normalized) {
			ignored = !negated
		}
	}

	return ignored
}

// FilterIgnoredFiles returns files that do not match any of the ignore
// patterns. The input order is preserved.
func FilterIgnoredFiles(files []string, ignorePatterns []string) []string {
	if len(ignorePatterns) == 0 {
		return files
	}

	filtered := make([]string, 0, len(files))
	for _, file := range files {
		if !ShouldIgnoreFile(file, ignorePatterns) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// normalizeIgnorePath converts a repository-relative path to the slash form
// used by ignore patterns.
func normalizeIgnorePath(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "/")
	for strings.HasPrefix(value, "./") {
		value = strings.TrimPrefix(value, "./")
	}
	for strings.HasPrefix(value, "/") {
		value = strings.TrimPrefix(value, "/")
	}
	if value == "" {
		return ""
	}

	cleaned := path.Clean(value)
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func matchIgnorePattern(pattern, file string) bool {
	pattern = strings.TrimSpace(pattern)
	pattern = strings.ReplaceAll(pattern, "\\", "/")
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return false
	}

	anchored := strings.HasPrefix(pattern, "/")
	pattern = strings.TrimPrefix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	if pattern == "" {
		return false
	}

	patternSegments := strings.Split(pattern, "/")
	fileSegments := strings.Split(file, "/")

	// A pattern containing a slash is relative to the repository root, just
	// like a gitignore pattern. A leading slash makes that explicit.
	if anchored || strings.Contains(pattern, "/") {
		return matchSegmentsPrefix(patternSegments, fileSegments)
	}

	// A pattern without a slash matches a file or directory name at any
	// depth, e.g. "go.sum" also matches "vendor/foo/go.sum".
	for _, segment := range fileSegments {
		if matchIgnoreSegment(patternSegments[0], segment) {
			return true
		}
	}
	return false
}

// matchSegmentsPrefix reports whether pattern matches the leading segments of
// file. Returning true once the pattern is exhausted implements gitignore's
// "a matching directory excludes its contents" behavior.
func matchSegmentsPrefix(patternSegments, fileSegments []string) bool {
	if len(patternSegments) == 0 {
		return true
	}

	if patternSegments[0] == "**" {
		for i := 0; i <= len(fileSegments); i++ {
			if matchSegmentsPrefix(patternSegments[1:], fileSegments[i:]) {
				return true
			}
		}
		return false
	}

	if len(fileSegments) == 0 {
		return false
	}
	if !matchIgnoreSegment(patternSegments[0], fileSegments[0]) {
		return false
	}
	return matchSegmentsPrefix(patternSegments[1:], fileSegments[1:])
}

func matchIgnoreSegment(pattern, segment string) bool {
	matched, err := path.Match(pattern, segment)
	return err == nil && matched
}
