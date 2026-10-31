package secretscan

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
)

// redactedMarker is the placeholder written into the transmitted diff.
func redactedMarker(ruleID string) string {
	return "[REDACTED:" + ruleID + "]"
}

type redactionEdit struct {
	start       int
	end         int
	replacement string
}

// Apply scans the added lines of a diff and enforces the configured mode:
//
//	off    → no scan, diff unchanged
//	warn   → diff unchanged, findings returned
//	redact → finding spans replaced in the returned diff
//	block  → structured error returned, caller must not send the diff
//
// A plain-text input without unified-diff headers (for example a piped review
// payload) is scanned as if every line were added under the "stdin" pseudo-file.
func Apply(diff string, opts Options) (Result, error) {
	switch opts.Mode {
	case ModeOff:
		return Result{Diff: diff}, nil
	case ModeWarn, ModeRedact, ModeBlock:
	default:
		return Result{}, fmt.Errorf("invalid sensitive scan mode %q", opts.Mode)
	}

	lines := splitDiffLines(diff)
	added, looksLikeDiff := parseAddedLines(lines)
	prefixLen := 1
	if len(added) == 0 {
		if looksLikeDiff || strings.TrimSpace(diff) == "" {
			return Result{Diff: diff}, nil
		}
		added = plainTextAddedLines(lines)
		prefixLen = 0
	}

	edits := make(map[int][]redactionEdit)
	findings := make([]Finding, 0, 4)

	for _, file := range groupAddedLines(added) {
		blob, starts := buildBlob(file.lines)
		for _, raw := range scanContent(file.path, blob) {
			lineText := ""
			if !raw.fileLevel {
				lineText = lineAt(blob, raw.start)
			}
			if allowedByUser(raw, lineText, opts.Allow) {
				continue
			}

			if raw.fileLevel {
				findings = append(findings, Finding{
					RuleID:    raw.ruleID,
					Path:      file.path,
					Line:      file.lines[0].newLine,
					FileLevel: true,
				})
				if opts.Mode == ModeRedact {
					for _, line := range file.lines {
						edits[line.diffIdx] = append(edits[line.diffIdx], redactionEdit{
							start:       0,
							end:         len(line.content),
							replacement: redactedMarker(raw.ruleID),
						})
					}
				}
				continue
			}

			startLine := lineIndex(starts, raw.start)
			endLine := lineIndex(starts, raw.end-1)
			line := file.lines[startLine]
			findings = append(findings, Finding{
				RuleID:   raw.ruleID,
				Path:     file.path,
				Line:     line.newLine,
				Column:   raw.start - starts[startLine] + 1,
				Secret:   raw.secret,
				LineText: lineText,
			})

			if opts.Mode != ModeRedact {
				continue
			}
			for idx := startLine; idx <= endLine && idx < len(file.lines); idx++ {
				lineStart := starts[idx]
				lineLen := len(file.lines[idx].content)
				segmentStart := 0
				if idx == startLine {
					segmentStart = raw.start - lineStart
				}
				segmentEnd := lineLen
				if idx == endLine {
					segmentEnd = raw.end - lineStart
					if segmentEnd > lineLen {
						segmentEnd = lineLen
					}
				}
				if segmentStart < 0 {
					segmentStart = 0
				}
				if segmentEnd <= segmentStart {
					continue
				}
				edits[file.lines[idx].diffIdx] = append(edits[file.lines[idx].diffIdx], redactionEdit{
					start:       segmentStart,
					end:         segmentEnd,
					replacement: redactedMarker(raw.ruleID),
				})
			}
		}
	}

	findings = dedupeFindings(findings)
	sortFindings(findings)

	result := Result{Diff: diff, Findings: findings}
	switch opts.Mode {
	case ModeWarn:
		return result, nil
	case ModeBlock:
		if len(findings) == 0 {
			return result, nil
		}
		// Make it impossible for a caller to accidentally send the diff.
		result.Diff = ""
		return result, gptcometerrors.SensitiveContentError(errorFindings(findings))
	default:
		result.Diff = applyRedactions(lines, edits, prefixLen)
		return result, nil
	}
}

func allowedByUser(raw rawFinding, lineText string, allow []*regexp.Regexp) bool {
	if len(allow) == 0 {
		return false
	}
	candidates := []string{raw.secret, raw.match, lineText}
	for _, re := range allow {
		for _, candidate := range candidates {
			if candidate != "" && re.MatchString(candidate) {
				return true
			}
		}
	}
	return false
}

func dedupeFindings(findings []Finding) []Finding {
	if len(findings) < 2 {
		return findings
	}
	seen := make(map[string]struct{}, len(findings))
	deduped := findings[:0]
	for _, finding := range findings {
		key := finding.RuleID + "\x00" + finding.Path + "\x00" + fmt.Sprint(finding.Line, ":", finding.Column)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, finding)
	}
	return deduped
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Column != right.Column {
			return left.Column < right.Column
		}
		return left.RuleID < right.RuleID
	})
}

func errorFindings(findings []Finding) []gptcometerrors.SensitiveFinding {
	converted := make([]gptcometerrors.SensitiveFinding, 0, len(findings))
	for _, finding := range findings {
		converted = append(converted, gptcometerrors.SensitiveFinding{
			Path:    finding.Path,
			Line:    finding.Line,
			RuleID:  finding.RuleID,
			Preview: finding.Preview(),
		})
	}
	return converted
}

// applyRedactions rewrites only the added lines, right-to-left, so line counts
// and hunk headers stay valid.
func applyRedactions(lines []diffLine, edits map[int][]redactionEdit, prefixLen int) string {
	redacted := make([]diffLine, len(lines))
	copy(redacted, lines)
	for diffIdx, lineEdits := range edits {
		if diffIdx < 0 || diffIdx >= len(redacted) {
			continue
		}
		text := redacted[diffIdx].text
		if prefixLen > len(text) {
			continue
		}
		content := text[prefixLen:]
		hadCarriageReturn := strings.HasSuffix(content, "\r")
		if hadCarriageReturn {
			content = strings.TrimSuffix(content, "\r")
		}
		content = applyEdits(content, lineEdits)
		newText := text[:prefixLen] + content
		if hadCarriageReturn {
			newText += "\r"
		}
		redacted[diffIdx].text = newText
	}
	return joinDiffLines(redacted)
}

func applyEdits(content string, edits []redactionEdit) string {
	if len(edits) == 0 {
		return content
	}
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start < edits[j].start
		}
		return edits[i].end < edits[j].end
	})

	merged := make([]redactionEdit, 0, len(edits))
	for _, edit := range edits {
		if edit.start < 0 {
			edit.start = 0
		}
		if edit.end > len(content) {
			edit.end = len(content)
		}
		if edit.end <= edit.start {
			continue
		}
		if len(merged) > 0 && edit.start <= merged[len(merged)-1].end {
			if edit.end > merged[len(merged)-1].end {
				merged[len(merged)-1].end = edit.end
			}
			continue
		}
		merged = append(merged, edit)
	}

	var builder strings.Builder
	builder.Grow(len(content))
	last := 0
	for _, edit := range merged {
		builder.WriteString(content[last:edit.start])
		builder.WriteString(edit.replacement)
		last = edit.end
	}
	builder.WriteString(content[last:])
	return builder.String()
}
