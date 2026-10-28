package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldIgnoreFile(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		patterns []string
		want     bool
	}{
		{
			name:     "exact root file",
			file:     "go.sum",
			patterns: []string{"go.sum"},
			want:     true,
		},
		{
			name:     "exact nested file",
			file:     "vendor/foo/go.sum",
			patterns: []string{"go.sum"},
			want:     true,
		},
		{
			name:     "file name at any depth",
			file:     "web/app/package-lock.json",
			patterns: []string{"package-lock.json"},
			want:     true,
		},
		{
			name:     "wildcard extension at any depth",
			file:     "cmd/a/b.lock",
			patterns: []string{"*.lock"},
			want:     true,
		},
		{
			name:     "character class at any depth",
			file:     "pkg/__pycache__/a.pyc",
			patterns: []string{"*.py[cod]"},
			want:     true,
		},
		{
			name:     "anchored root matches",
			file:     "go.sum",
			patterns: []string{"/go.sum"},
			want:     true,
		},
		{
			name:     "anchored root does not match nested",
			file:     "sub/go.sum",
			patterns: []string{"/go.sum"},
			want:     false,
		},
		{
			name:     "path pattern matches descendants",
			file:     "vendor/a/b.go",
			patterns: []string{"vendor/**"},
			want:     true,
		},
		{
			name:     "path pattern is anchored",
			file:     "other/vendor/a.go",
			patterns: []string{"vendor/**"},
			want:     false,
		},
		{
			name:     "double star in the middle",
			file:     "a/b/c/d.go",
			patterns: []string{"a/**/d.go"},
			want:     true,
		},
		{
			name:     "double star matches zero segments",
			file:     "a/d.go",
			patterns: []string{"a/**/d.go"},
			want:     true,
		},
		{
			name:     "single star does not cross directories",
			file:     "docs/sub/a.md",
			patterns: []string{"docs/*.md"},
			want:     false,
		},
		{
			name:     "single star matches direct child",
			file:     "docs/a.md",
			patterns: []string{"docs/*.md"},
			want:     true,
		},
		{
			name:     "directory pattern at any depth",
			file:     "pkg/build/x.go",
			patterns: []string{"build/"},
			want:     true,
		},
		{
			name:     "directory pattern matches path itself",
			file:     "pkg/build",
			patterns: []string{"build/"},
			want:     true,
		},
		{
			name:     "negation re-includes",
			file:     "keep.log",
			patterns: []string{"*.log", "!keep.log"},
			want:     false,
		},
		{
			name:     "negation order matters",
			file:     "keep.log",
			patterns: []string{"!keep.log", "*.log"},
			want:     true,
		},
		{
			name:     "comment is skipped",
			file:     "a.log",
			patterns: []string{"# comment", "*.log"},
			want:     true,
		},
		{
			name:     "windows separators are normalized",
			file:     `vendor\foo\go.sum`,
			patterns: []string{"go.sum"},
			want:     true,
		},
		{
			name:     "dot slash path is normalized",
			file:     "./cmd/main.go",
			patterns: []string{"cmd/*.go"},
			want:     true,
		},
		{
			name:     "non match",
			file:     "main.go",
			patterns: []string{"*.lock"},
			want:     false,
		},
		{
			name:     "empty patterns",
			file:     "main.go",
			patterns: []string{"", "  "},
			want:     false,
		},
		{
			name:     "nil patterns",
			file:     "main.go",
			patterns: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShouldIgnoreFile(tt.file, tt.patterns))
		})
	}
}

func TestFilterIgnoredFiles_GitignoreSemantics(t *testing.T) {
	files := []string{
		"cmd/main.go",
		"go.sum",
		"vendor/x/go.sum",
		"web/package-lock.json",
	}
	patterns := []string{"go.sum", "package-lock.json"}

	assert.Equal(t, []string{"cmd/main.go"}, FilterIgnoredFiles(files, patterns))
}
