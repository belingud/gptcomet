package secretscan

import (
	"regexp"
	"strings"
	"testing"
)

const (
	fakeAWSAccessKey = "AKIA" + "Q7X2M4P6R5T2V4W6"
	fakeGitHubPAT    = "ghp_" + "5xJ8kLm2Np4Qr6St8Uv0Wx2Yz4Ab6Cd8Ef0G"
	fakeOpenAIKey    = "sk-" + "abcdefghij1234567890" + "T3BlbkFJ" + "abcdefghij1234567890"
	fakePEMPrivate   = "-----BEGIN RSA " + "PRIVATE KEY-----\n" +
		"MIIEowIBAAKCAQEA0f1e2d3c4b5a697887766554433221100ffeeddccbbaa9988776655443322\n" +
		"-----END RSA " + "PRIVATE KEY-----"
)

func findingRuleIDs(findings []rawFinding) []string {
	ids := make([]string, 0, len(findings))
	for _, finding := range findings {
		ids = append(ids, finding.ruleID)
	}
	return ids
}

func hasRule(findings []rawFinding, ruleID string) bool {
	for _, finding := range findings {
		if finding.ruleID == ruleID {
			return true
		}
	}
	return false
}

func TestScanContentDetectsSecrets(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    []string
	}{
		{
			name:    "aws access key",
			path:    "config/settings.go",
			content: "aws_access_key_id = " + fakeAWSAccessKey,
			want:    []string{"aws-access-token"},
		},
		{
			name:    "github personal access token",
			path:    "cmd/deploy.go",
			content: `token := "` + fakeGitHubPAT + `"`,
			want:    []string{"github-pat"},
		},
		{
			name:    "openai api key",
			path:    "config.yaml",
			content: "api_key: " + fakeOpenAIKey,
			want:    []string{"openai-api-key"},
		},
		{
			name:    "pem private key block",
			path:    "deploy/key.txt",
			content: fakePEMPrivate,
			want:    []string{"private-key"},
		},
		{
			name:    "sensitive env file is flagged even without a value match",
			path:    ".env",
			content: "SOME_OPAQUE_SETTING=abc123",
			want:    []string{"sensitive-env-file"},
		},
		{
			name:    "sensitive credential file",
			path:    "config/credentials.json",
			content: "some random content",
			want:    []string{"sensitive-credential-file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := scanContent(tt.path, tt.content)
			for _, ruleID := range tt.want {
				if !hasRule(findings, ruleID) {
					t.Fatalf("rule %q not found, got %v", ruleID, findingRuleIDs(findings))
				}
			}
		})
	}
}

func TestScanContentIgnoresExpectedFalsePositives(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
	}{
		{
			name:    "aws documentation example key",
			path:    "README.md",
			content: "aws_access_key_id = " + "AKIA" + "IOSFODNN7EXAMPLE",
		},
		{
			name:    "lockfile hash",
			path:    "package-lock.json",
			content: "integrity sha512-9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			name:    "uuid stopword",
			path:    "config.yaml",
			content: "id: 014df517-39d1-4453-b7b3-9930c563627c",
		},
		{
			name:    "template placeholder",
			path:    "deploy.yaml",
			content: "password: {{ secrets.DB_PASSWORD }}",
		},
		{
			name:    "gitleaks allow comment",
			path:    "cmd/fixture.go",
			content: `token := "` + fakeGitHubPAT + `" // gitleaks:allow`,
		},
		{
			name:    "env template file",
			path:    ".env.example",
			content: "OPENAI_API_KEY=replace-me",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := scanContent(tt.path, tt.content)
			if len(findings) != 0 {
				t.Fatalf("unexpected findings: %v", findingRuleIDs(findings))
			}
		})
	}
}

func TestScanContentMultiLineSpan(t *testing.T) {
	blob := fakePEMPrivate
	findings := scanContent("deploy/key.txt", blob)
	var pem *rawFinding
	for i := range findings {
		if findings[i].ruleID == "private-key" {
			pem = &findings[i]
			break
		}
	}
	if pem == nil {
		t.Fatalf("private-key finding missing: %v", findingRuleIDs(findings))
	}
	if !strings.Contains(pem.match, "BEGIN RSA "+"PRIVATE KEY") {
		t.Errorf("match should include PEM header: %q", pem.match)
	}
	if pem.start != 0 {
		t.Errorf("start = %d, want 0", pem.start)
	}
	if pem.end <= len("-----BEGIN RSA "+"PRIVATE KEY-----") {
		t.Errorf("end = %d should cover the block", pem.end)
	}
}

func TestAllBundledRulesCompile(t *testing.T) {
	rules, globalAllowlists := compiledRuleSet()
	wantRules := len(gitleaksRules) + len(localRules)
	if wantRules == 0 {
		t.Fatal("no bundled rules")
	}
	if len(rules) != wantRules {
		t.Fatalf("compiled %d rules, want %d (a regex failed to compile)", len(rules), wantRules)
	}
	seen := make(map[string]struct{}, len(rules))
	for _, compiled := range rules {
		if compiled.rule.ID == "" {
			t.Error("rule with empty id")
		}
		if _, ok := seen[compiled.rule.ID]; ok {
			t.Errorf("duplicate rule id %q", compiled.rule.ID)
		}
		seen[compiled.rule.ID] = struct{}{}
		if compiled.regex == nil && compiled.path == nil {
			t.Errorf("rule %q has neither regex nor path", compiled.rule.ID)
		}
	}
	if len(globalAllowlists) == 0 {
		t.Error("global allowlist was not compiled")
	}
}

func TestGitleaksRuleCountMatchesUpstream(t *testing.T) {
	// The vendored gitleaks v8.30.1 default config contains 222 rules. Update
	// this assertion together with the vendored TOML and UPSTREAM.md.
	if got := len(gitleaksRules); got != 222 {
		t.Fatalf("gitleaksRules = %d, want 222", got)
	}
	if got := len(gitleaksGlobalAllowlists); got != 1 {
		t.Fatalf("gitleaksGlobalAllowlists = %d, want 1", got)
	}
}

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "empty", input: "", want: 0},
		{name: "single char repeated", input: "aaaa", want: 0},
		{name: "two equal chars", input: "abab", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shannonEntropy(tt.input)
			if got != tt.want {
				t.Fatalf("shannonEntropy(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAllowlistPathSkipsRule(t *testing.T) {
	if !allowlistSkipsPath(compileAllowlists([]allowlist{{Paths: []string{`\.png$`}}}), "x.png") {
		t.Fatal("global allowlist path should skip")
	}
	if allowlistSkipsPath(compileAllowlists([]allowlist{{Paths: []string{`\.png$`}}}), "x.go") {
		t.Fatal("unexpected skip for .go")
	}
}

func TestAllowlistsMatchFindingTargets(t *testing.T) {
	allow := compileAllowlists([]allowlist{{
		Regexes:     []string{`EXAMPLE$`},
		RegexTarget: "secret",
	}})
	if !allowlistsMatchFinding(allow, "a.go", "line", "match", "AKIAEXAMPLE") {
		t.Fatal("secret target should match")
	}

	lineOnly := compileAllowlists([]allowlist{{
		Regexes:     []string{`placeholder`},
		RegexTarget: "line",
	}})
	if !allowlistsMatchFinding(lineOnly, "a.go", "value = placeholder", "match", "secret") {
		t.Fatal("line target should match")
	}
	if allowlistsMatchFinding(lineOnly, "a.go", "value = real", "match", "secret") {
		t.Fatal("line target should not match")
	}

	and := compileAllowlists([]allowlist{{
		Condition: allowlistAND,
		Paths:     []string{`\.env$`},
		Regexes:   []string{`abc`},
	}})
	if !allowlistsMatchFinding(and, "x.env", "line", "match", "abc") {
		t.Fatal("AND allowlist should match when both checks pass")
	}
	if allowlistsMatchFinding(and, "x.go", "line", "match", "abc") {
		t.Fatal("AND allowlist should not match when path fails")
	}
}

func TestCompileAllow(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantLen  int
		wantErr  bool
	}{
		{name: "valid", patterns: []string{`^dummy-`, `EXAMPLE`}, wantLen: 2},
		{name: "skips blank", patterns: []string{"", "  "}, wantLen: 0},
		{name: "invalid", patterns: []string{`(`}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := CompileAllow(tt.patterns)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(compiled) != tt.wantLen {
				t.Fatalf("got %d regexes, want %d", len(compiled), tt.wantLen)
			}
			if len(compiled) > 0 && !compiled[0].MatchString("dummy-value") {
				t.Error("first allow pattern did not compile as expected")
			}
		})
	}
}

func TestSecretFromMatchPrefersCaptureGroup(t *testing.T) {
	re := regexp.MustCompile(`key=([A-Za-z0-9]+)`)
	secret, ok := secretFromMatch("key=abc123", re, 0)
	if !ok || secret != "abc123" {
		t.Fatalf("secretFromMatch = %q, %v", secret, ok)
	}
	secret, ok = secretFromMatch("key=abc123", re, 1)
	if !ok || secret != "abc123" {
		t.Fatalf("explicit group = %q, %v", secret, ok)
	}
	if _, ok := secretFromMatch("key=abc123", re, 2); ok {
		t.Fatal("missing group should report not ok")
	}
}
