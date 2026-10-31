package secretscan

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
)

func makeDiff(path string, lines ...string) string {
	var builder strings.Builder
	builder.WriteString("diff --git a/" + path + " b/" + path + "\n")
	builder.WriteString("index 1111111..2222222 100644\n")
	builder.WriteString("--- a/" + path + "\n")
	builder.WriteString("+++ b/" + path + "\n")
	builder.WriteString("@@ -0,0 +1," + strconv.Itoa(len(lines)) + " @@\n")
	for _, line := range lines {
		builder.WriteString("+" + line + "\n")
	}
	return builder.String()
}

func TestApplyWarnKeepsDiffUnchanged(t *testing.T) {
	diff := makeDiff("cmd/deploy.go", fakeGitHubPAT)

	result, err := Apply(diff, Options{Mode: ModeWarn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != diff {
		t.Fatal("warn mode must not modify the diff")
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].RuleID != "github-pat" {
		t.Fatalf("rule = %q, want github-pat", result.Findings[0].RuleID)
	}
	if result.Findings[0].Path != "cmd/deploy.go" || result.Findings[0].Line != 1 {
		t.Fatalf("location = %s:%d", result.Findings[0].Path, result.Findings[0].Line)
	}
	if result.Findings[0].Column <= 0 {
		t.Fatalf("column = %d, want > 0", result.Findings[0].Column)
	}
	if !strings.Contains(result.Findings[0].LineText, "ghp_") {
		t.Fatalf("line text = %q", result.Findings[0].LineText)
	}
}

func TestApplyRedactReplacesSecretOnly(t *testing.T) {
	diff := makeDiff("cmd/deploy.go", fakeGitHubPAT, "other := 1")

	result, err := Apply(diff, Options{Mode: ModeRedact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Diff, fakeGitHubPAT) {
		t.Fatal("redacted diff still contains the secret")
	}
	if !strings.Contains(result.Diff, "[REDACTED:github-pat]") {
		t.Fatalf("redacted diff missing marker:\n%s", result.Diff)
	}
	if !strings.Contains(result.Diff, "other := 1") {
		t.Fatal("non-sensitive line changed")
	}
	if got, want := strings.Count(result.Diff, "\n"), strings.Count(diff, "\n"); got != want {
		t.Fatalf("line count changed: got %d, want %d", got, want)
	}
	if strings.Count(result.Diff, "@@") != strings.Count(diff, "@@") {
		t.Fatal("hunk header changed")
	}
}

func TestApplyRedactMultiLinePrivateKey(t *testing.T) {
	diff := makeDiff("deploy/key.txt", strings.Split(fakePEMPrivate, "\n")...)

	result, err := Apply(diff, Options{Mode: ModeRedact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Diff, "PRIVATE KEY") || strings.Contains(result.Diff, "MIIEow") {
		t.Fatalf("redacted diff leaks key material:\n%s", result.Diff)
	}
	if got, want := strings.Count(result.Diff, "\n"), strings.Count(diff, "\n"); got != want {
		t.Fatalf("line count changed: got %d, want %d", got, want)
	}
	if !strings.Contains(result.Diff, "[REDACTED:private-key]") {
		t.Fatalf("redacted diff missing marker:\n%s", result.Diff)
	}
}

func TestApplyRedactSensitiveFile(t *testing.T) {
	diff := makeDiff(".env", "OPENAI_API_KEY=opaque-value", "DATABASE_URL=postgres://localhost/db")

	result, err := Apply(diff, Options{Mode: ModeRedact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Diff, "opaque-value") || strings.Contains(result.Diff, "DATABASE_URL") {
		t.Fatalf("sensitive file content leaked:\n%s", result.Diff)
	}
	if !strings.Contains(result.Diff, "[REDACTED:sensitive-env-file]") {
		t.Fatalf("missing file-level marker:\n%s", result.Diff)
	}
	// Both added lines keep their "+" prefix and newline.
	if got, want := strings.Count(result.Diff, "\n"), strings.Count(diff, "\n"); got != want {
		t.Fatalf("line count changed: got %d, want %d", got, want)
	}
}

func TestApplyBlockReturnsStructuredError(t *testing.T) {
	diff := makeDiff("cmd/deploy.go", fakeGitHubPAT)

	result, err := Apply(diff, Options{Mode: ModeBlock})
	if err == nil {
		t.Fatal("expected block error")
	}
	var gptErr *gptcometerrors.GPTCometError
	if !errors.As(err, &gptErr) {
		t.Fatalf("error is %T, want *GPTCometError", err)
	}
	if gptErr.Type != gptcometerrors.ErrTypeValidation {
		t.Fatalf("error type = %q, want validation", gptErr.Type)
	}
	if result.Diff != "" {
		t.Fatal("block mode must not return a diff to send")
	}
	if len(result.Findings) == 0 {
		t.Fatal("block mode should still report findings")
	}
	message := err.Error()
	if strings.Contains(message, fakeGitHubPAT) {
		t.Fatal("block error leaks the full secret")
	}
	if !strings.Contains(message, "cmd/deploy.go:1") || !strings.Contains(message, "github-pat") {
		t.Fatalf("block error missing location/rule:\n%s", message)
	}
	if !strings.Contains(message, "--secrets off") {
		t.Fatalf("block error missing bypass suggestion:\n%s", message)
	}
}

func TestApplyOffDoesNothing(t *testing.T) {
	diff := makeDiff("cmd/deploy.go", fakeGitHubPAT)

	result, err := Apply(diff, Options{Mode: ModeOff})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != diff {
		t.Fatal("off mode must not modify the diff")
	}
	if len(result.Findings) != 0 {
		t.Fatalf("off mode returned findings: %+v", result.Findings)
	}
}

func TestApplyUserAllowlistSilencesFinding(t *testing.T) {
	diff := makeDiff("cmd/deploy.go", fakeGitHubPAT)
	allow := []*regexp.Regexp{regexp.MustCompile(`^ghp_`)}

	result, err := Apply(diff, Options{Mode: ModeBlock, Allow: allow})
	if err != nil {
		t.Fatalf("allowlisted finding should not block: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("allowlisted finding still reported: %+v", result.Findings)
	}
	if result.Diff != diff {
		t.Fatal("diff should be unchanged in block mode without findings")
	}
}

func TestApplyPlainTextFallback(t *testing.T) {
	text := "token=" + fakeGitHubPAT + "\n"

	result, err := Apply(text, Options{Mode: ModeWarn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	found := false
	for _, finding := range result.Findings {
		if finding.Path != stdinPath || finding.Line != 1 {
			t.Fatalf("finding location = %s:%d", finding.Path, finding.Line)
		}
		if finding.RuleID == "github-pat" {
			found = true
		}
	}
	if !found {
		t.Fatalf("github-pat finding missing: %+v", result.Findings)
	}

	redacted, err := Apply(text, Options{Mode: ModeRedact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(redacted.Diff, fakeGitHubPAT) {
		t.Fatal("plain-text redaction leaked the secret")
	}
	if !strings.Contains(redacted.Diff, "[REDACTED:") {
		t.Fatalf("redacted plain text missing marker: %q", redacted.Diff)
	}
	if strings.Count(redacted.Diff, "\n") != 1 {
		t.Fatalf("line count changed: %q", redacted.Diff)
	}
}

func TestApplyPreservesCRLF(t *testing.T) {
	diff := strings.ReplaceAll(makeDiff("cmd/deploy.go", fakeGitHubPAT), "\n", "\r\n")

	result, err := Apply(diff, Options{Mode: ModeRedact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result.Diff, fakeGitHubPAT) {
		t.Fatal("redacted diff leaked the secret")
	}
	if !strings.Contains(result.Diff, "[REDACTED:github-pat]\r\n") {
		t.Fatalf("CRLF not preserved:\n%q", result.Diff)
	}
	if got, want := strings.Count(result.Diff, "\r\n"), strings.Count(diff, "\r\n"); got != want {
		t.Fatalf("line count changed: got %d, want %d", got, want)
	}
}

func TestApplyEmptyDiff(t *testing.T) {
	result, err := Apply("", Options{Mode: ModeWarn})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Diff != "" || len(result.Findings) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestApplyInvalidMode(t *testing.T) {
	if _, err := Apply("x", Options{Mode: Mode("bogus")}); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    Mode
		wantErr bool
	}{
		{name: "off", value: "off", want: ModeOff},
		{name: "warn", value: "warn", want: ModeWarn},
		{name: "redact uppercase", value: " REDACT ", want: ModeRedact},
		{name: "block", value: "block", want: ModeBlock},
		{name: "invalid", value: "yes", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseMode(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFindingPreviewMasksSecret(t *testing.T) {
	tests := []struct {
		name      string
		finding   Finding
		want      string
		wantNotIn string
	}{
		{
			name:      "file level",
			finding:   Finding{FileLevel: true, Secret: ""},
			want:      "(file-level)",
			wantNotIn: "",
		},
		{
			name:      "long secret",
			finding:   Finding{Secret: fakeGitHubPAT},
			want:      "ghp_…Ef0G",
			wantNotIn: fakeGitHubPAT,
		},
		{
			name:      "short secret",
			finding:   Finding{Secret: "abc123"},
			want:      "******",
			wantNotIn: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.finding.Preview()
			if got != tt.want {
				t.Fatalf("Preview() = %q, want %q", got, tt.want)
			}
			if tt.wantNotIn != "" && strings.Contains(got, tt.wantNotIn) {
				t.Fatalf("Preview() leaks secret: %q", got)
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	if maskSecret("") != "" {
		t.Fatal("empty secret should mask to empty")
	}
	if maskSecret("12345678") != "********" {
		t.Fatal("short secret should be fully masked")
	}
	if got := maskSecret("ghp_abcdefghijklmnopqrstuvwxyz"); got != "ghp_…wxyz" {
		t.Fatalf("maskSecret() = %q", got)
	}
}

func BenchmarkApplyLargeDiff(b *testing.B) {
	lines := make([]string, 5000)
	for i := range lines {
		lines[i] = fmt.Sprintf("func helper%d(x int) int { return x + %d }", i, i)
	}
	diff := makeDiff("big.go", lines...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Apply(diff, Options{Mode: ModeWarn}); err != nil {
			b.Fatal(err)
		}
	}
}
