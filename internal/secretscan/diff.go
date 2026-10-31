package secretscan

import "strings"

// stdinPath is the pseudo file path used when the input is plain text rather
// than a unified diff (for example a diff piped into "gmsg review").
const stdinPath = "stdin"

// diffLine is one line of the original input, keeping the information needed to
// rebuild it byte-for-byte after redaction.
type diffLine struct {
	text       string // without the trailing newline
	hadNewline bool
}

func splitDiffLines(text string) []diffLine {
	if text == "" {
		return nil
	}

	parts := strings.Split(text, "\n")
	lines := make([]diffLine, 0, len(parts))
	for i, part := range parts {
		if i == len(parts)-1 && part == "" {
			// Trailing newline: the previous line already recorded it.
			break
		}
		lines = append(lines, diffLine{text: part, hadNewline: i < len(parts)-1})
	}
	return lines
}

func joinDiffLines(lines []diffLine) string {
	var builder strings.Builder
	for _, line := range lines {
		builder.WriteString(line.text)
		if line.hadNewline {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

// addedLine is one added ("+") line parsed out of a unified diff.
type addedLine struct {
	path    string
	newLine int // 1-based line number in the new file
	diffIdx int // index into the split diff lines
	content string
}

type fileAddedLines struct {
	path  string
	lines []addedLine
}

// parseAddedLines extracts every added line from a unified diff. The bool result
// reports whether the input looked like a diff at all, so callers can fall back
// to treating plain text as added content.
func parseAddedLines(lines []diffLine) ([]addedLine, bool) {
	var (
		added         []addedLine
		path          string
		newLine       int
		inHunk        bool
		binary        bool
		looksLikeDiff bool
	)

	for idx, line := range lines {
		text := line.text
		switch {
		case strings.HasPrefix(text, "diff --git "):
			looksLikeDiff = true
			path = ""
			newLine = 0
			inHunk = false
			binary = false
			continue
		case strings.HasPrefix(text, "+++ "):
			looksLikeDiff = true
			path = parseDiffPath(text[len("+++ "):])
			inHunk = false
			continue
		case strings.HasPrefix(text, "@@ "):
			looksLikeDiff = true
			newLine = parseHunkNewStart(text)
			inHunk = true
			continue
		case strings.HasPrefix(text, "Binary files ") && strings.HasSuffix(text, " differ"):
			binary = true
			continue
		case text == "GIT binary patch":
			binary = true
			continue
		}

		if !inHunk || binary || path == "" || path == "/dev/null" {
			continue
		}

		switch {
		case strings.HasPrefix(text, "+"):
			added = append(added, addedLine{
				path:    path,
				newLine: newLine,
				diffIdx: idx,
				content: strings.TrimSuffix(text[1:], "\r"),
			})
			newLine++
		case strings.HasPrefix(text, "-"):
			// Deleted lines do not advance the new-file line number.
		case strings.HasPrefix(text, `\`):
			// "\ No newline at end of file" is not content.
		default:
			// Context lines start with a space; an empty line also counts.
			newLine++
		}
	}

	return added, looksLikeDiff
}

// parseDiffPath converts a "+++ b/path" target into a clean relative path.
func parseDiffPath(raw string) string {
	path := strings.TrimSpace(raw)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, `"`) && strings.HasSuffix(path, `"`) {
		// Git quotes paths with unusual characters; strip the quotes and the
		// common escape sequences. Rare and only cosmetic for the scan.
		path = strings.Trim(path, `"`)
		path = strings.ReplaceAll(path, `\"`, `"`)
		path = strings.ReplaceAll(path, `\\`, `\`)
	}
	if path == "/dev/null" {
		return path
	}
	if strings.HasPrefix(path, "b/") {
		return path[len("b/"):]
	}
	return path
}

// parseHunkNewStart parses the new-file start line out of "@@ -a,b +c,d @@".
func parseHunkNewStart(text string) int {
	plus := strings.IndexByte(text, '+')
	if plus < 0 {
		return 0
	}
	rest := text[plus+1:]
	if comma := strings.IndexAny(rest, ", "); comma >= 0 {
		rest = rest[:comma]
	}
	number := 0
	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if ch < '0' || ch > '9' {
			break
		}
		number = number*10 + int(ch-'0')
	}
	return number
}

// plainTextAddedLines treats every input line as added content of a synthetic
// "stdin" file.
func plainTextAddedLines(lines []diffLine) []addedLine {
	added := make([]addedLine, 0, len(lines))
	for idx, line := range lines {
		added = append(added, addedLine{
			path:    stdinPath,
			newLine: idx + 1,
			diffIdx: idx,
			content: strings.TrimSuffix(line.text, "\r"),
		})
	}
	return added
}

// groupAddedLines preserves first-seen file order, which keeps warnings and
// redaction deterministic.
func groupAddedLines(added []addedLine) []fileAddedLines {
	indexByPath := make(map[string]int)
	files := make([]fileAddedLines, 0, 4)
	for _, line := range added {
		index, ok := indexByPath[line.path]
		if !ok {
			index = len(files)
			indexByPath[line.path] = index
			files = append(files, fileAddedLines{path: line.path})
		}
		files[index].lines = append(files[index].lines, line)
	}
	return files
}

// buildBlob joins the file's added lines with newlines and records the byte
// offset of each line inside the blob.
func buildBlob(lines []addedLine) (string, []int) {
	starts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		starts[i] = offset
		offset += len(line.content) + 1
	}
	size := 0
	if len(lines) > 0 {
		size = starts[len(starts)-1] + len(lines[len(lines)-1].content)
	}

	var builder strings.Builder
	builder.Grow(size)
	for i, line := range lines {
		if i > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(line.content)
	}
	return builder.String(), starts
}

// lineAt returns the full blob line containing offset.
func lineAt(blob string, offset int) string {
	if offset < 0 {
		offset = 0
	}
	if offset > len(blob) {
		offset = len(blob)
	}
	start := strings.LastIndexByte(blob[:offset], '\n') + 1
	end := strings.IndexByte(blob[offset:], '\n')
	if end < 0 {
		end = len(blob)
	} else {
		end += offset
	}
	return blob[start:end]
}

// lineIndex returns the index of the line containing offset.
func lineIndex(starts []int, offset int) int {
	if len(starts) == 0 {
		return 0
	}
	low, high := 0, len(starts)-1
	for low < high {
		mid := (low + high + 1) / 2
		if starts[mid] <= offset {
			low = mid
		} else {
			high = mid - 1
		}
	}
	if low >= len(starts) {
		return len(starts) - 1
	}
	return low
}
