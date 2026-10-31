package secretscan

import (
	"strings"
	"testing"
)

func TestSplitJoinDiffLinesRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "no trailing newline", in: "a\nb"},
		{name: "trailing newline", in: "a\nb\n"},
		{name: "blank lines", in: "\n\n"},
		{name: "single newline", in: "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinDiffLines(splitDiffLines(tt.in))
			if got != tt.in {
				t.Fatalf("round trip mismatch: got %q, want %q", got, tt.in)
			}
		})
	}
}

func TestParseAddedLines(t *testing.T) {
	tests := []struct {
		name          string
		diff          string
		wantCount     int
		wantPaths     []string
		wantNewLines  []int
		wantLooksDiff bool
		wantContents  []string
	}{
		{
			name: "single file with context and deletions",
			diff: "diff --git a/a.go b/a.go\n" +
				"index 111..222 100644\n" +
				"--- a/a.go\n" +
				"+++ b/a.go\n" +
				"@@ -1,3 +1,3 @@\n" +
				" context\n" +
				"-old line\n" +
				"+new line\n" +
				"+second new line\n",
			wantCount:     2,
			wantPaths:     []string{"a.go", "a.go"},
			wantNewLines:  []int{2, 3},
			wantLooksDiff: true,
			wantContents:  []string{"new line", "second new line"},
		},
		{
			name: "multiple files and hunks",
			diff: "diff --git a/a.go b/a.go\n" +
				"--- a/a.go\n" +
				"+++ b/a.go\n" +
				"@@ -1,1 +1,1 @@\n" +
				"+first\n" +
				"@@ -10,1 +11,1 @@\n" +
				"+second\n" +
				"diff --git a/b.go b/b.go\n" +
				"--- a/b.go\n" +
				"+++ b/b.go\n" +
				"@@ -0,0 +1,1 @@\n" +
				"+third\n",
			wantCount:     3,
			wantPaths:     []string{"a.go", "a.go", "b.go"},
			wantNewLines:  []int{1, 11, 1},
			wantLooksDiff: true,
			wantContents:  []string{"first", "second", "third"},
		},
		{
			name: "deleted file is skipped",
			diff: "diff --git a/old b/old\n" +
				"--- a/old\n" +
				"+++ /dev/null\n" +
				"@@ -1,1 +0,0 @@\n" +
				"-gone\n",
			wantLooksDiff: true,
		},
		{
			name: "binary diff is skipped",
			diff: "diff --git a/x.png b/x.png\n" +
				"Binary files a/x.png and b/x.png differ\n",
			wantLooksDiff: true,
			// binary detection leaves the header path set but no hunks, so no additions.
		},
		{
			name: "no newline marker does not advance lines",
			diff: "diff --git a/a.txt b/a.txt\n" +
				"--- a/a.txt\n" +
				"+++ b/a.txt\n" +
				"@@ -1,2 +1,2 @@\n" +
				" keep\n" +
				"+added\n" +
				"\\ No newline at end of file\n",
			wantCount:     1,
			wantPaths:     []string{"a.txt"},
			wantNewLines:  []int{2},
			wantLooksDiff: true,
			wantContents:  []string{"added"},
		},
		{
			name:          "plain text is not a diff",
			diff:          "just some text\nwith two lines\n",
			wantLooksDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, looksLikeDiff := parseAddedLines(splitDiffLines(tt.diff))
			if looksLikeDiff != tt.wantLooksDiff {
				t.Fatalf("looksLikeDiff = %v, want %v", looksLikeDiff, tt.wantLooksDiff)
			}
			if len(added) != tt.wantCount {
				t.Fatalf("got %d added lines, want %d: %+v", len(added), tt.wantCount, added)
			}
			for i, line := range added {
				if line.path != tt.wantPaths[i] {
					t.Errorf("line %d path = %q, want %q", i, line.path, tt.wantPaths[i])
				}
				if line.newLine != tt.wantNewLines[i] {
					t.Errorf("line %d newLine = %d, want %d", i, line.newLine, tt.wantNewLines[i])
				}
				if line.content != tt.wantContents[i] {
					t.Errorf("line %d content = %q, want %q", i, line.content, tt.wantContents[i])
				}
			}
		})
	}
}

func TestParseDiffPath(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain", raw: "b/internal/config.yaml", want: "internal/config.yaml"},
		{name: "dev null", raw: "/dev/null", want: "/dev/null"},
		{name: "quoted", raw: `"b/with space/file.go"`, want: "with space/file.go"},
		{name: "empty", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDiffPath(tt.raw); got != tt.want {
				t.Fatalf("parseDiffPath(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestBuildBlobAndLineHelpers(t *testing.T) {
	lines := []addedLine{
		{content: "one"},
		{content: "two"},
		{content: ""},
	}
	blob, starts := buildBlob(lines)
	if blob != "one\ntwo\n" {
		t.Fatalf("blob = %q, want %q", blob, "one\ntwo\n")
	}
	if len(starts) != 3 || starts[0] != 0 || starts[1] != 4 || starts[2] != 8 {
		t.Fatalf("starts = %v", starts)
	}

	if got := lineIndex(starts, 0); got != 0 {
		t.Errorf("lineIndex(0) = %d, want 0", got)
	}
	if got := lineIndex(starts, 5); got != 1 {
		t.Errorf("lineIndex(5) = %d, want 1", got)
	}
	if got := lineIndex(starts, 8); got != 2 {
		t.Errorf("lineIndex(8) = %d, want 2", got)
	}

	if got := lineAt(blob, 5); got != "two" {
		t.Errorf("lineAt(5) = %q, want %q", got, "two")
	}
	if got := lineAt(blob, len(blob)); got != "" {
		t.Errorf("lineAt(end) = %q, want empty", got)
	}
}

func TestGroupAddedLinesPreservesOrder(t *testing.T) {
	added := []addedLine{
		{path: "b.go"},
		{path: "a.go"},
		{path: "b.go"},
	}
	files := groupAddedLines(added)
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if files[0].path != "b.go" || len(files[0].lines) != 2 {
		t.Errorf("first file = %+v", files[0])
	}
	if files[1].path != "a.go" || len(files[1].lines) != 1 {
		t.Errorf("second file = %+v", files[1])
	}
	if !strings.Contains(files[0].path, "b") {
		t.Errorf("unexpected path %q", files[0].path)
	}
}
