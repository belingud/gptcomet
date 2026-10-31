package secretscan

import (
	"math"
	"regexp"
	"strings"
	"sync"
)

// gitleaksAllowSignature lets a developer opt a single line out of the scan,
// mirroring gitleaks' inline suppression comment.
const gitleaksAllowSignature = "gitleaks:allow"

// rawFinding is an internal match before it is mapped back to diff coordinates.
type rawFinding struct {
	ruleID    string
	fileLevel bool
	start     int // byte offset in the file blob
	end       int // byte offset in the file blob, exclusive
	match     string
	secret    string
}

type compiledRule struct {
	rule       rule
	regex      *regexp.Regexp
	path       *regexp.Regexp
	keywords   []string
	allowlists []compiledAllowlist
}

type compiledAllowlist struct {
	condition allowlistCondition
	paths     []*regexp.Regexp
	regexes   []*regexp.Regexp
	target    string
	stopwords []string
}

var (
	compileOnce    sync.Once
	compiledRules  []compiledRule
	compiledGlobal []compiledAllowlist
)

// compiledRuleSet compiles the generated gitleaks rules, the local sensitive-file
// rules, and the global allowlist exactly once per process.
func compiledRuleSet() ([]compiledRule, []compiledAllowlist) {
	compileOnce.Do(func() {
		rules := make([]rule, 0, len(gitleaksRules)+len(localRules))
		rules = append(rules, gitleaksRules...)
		rules = append(rules, localRules...)

		compiledRules = make([]compiledRule, 0, len(rules))
		for _, rawRule := range rules {
			compiled, ok := compileRule(rawRule)
			if ok {
				compiledRules = append(compiledRules, compiled)
			}
		}
		compiledGlobal = compileAllowlists(gitleaksGlobalAllowlists)
	})
	return compiledRules, compiledGlobal
}

func compileRule(rawRule rule) (compiledRule, bool) {
	compiled := compiledRule{
		rule:       rawRule,
		keywords:   lowerStrings(rawRule.Keywords),
		allowlists: compileAllowlists(rawRule.Allowlists),
	}
	if rawRule.Regex != "" {
		re, err := regexp.Compile(rawRule.Regex)
		if err != nil {
			return compiledRule{}, false
		}
		compiled.regex = re
	}
	if rawRule.Path != "" {
		re, err := regexp.Compile(rawRule.Path)
		if err != nil {
			return compiledRule{}, false
		}
		compiled.path = re
	}
	if compiled.regex == nil && compiled.path == nil {
		return compiledRule{}, false
	}
	return compiled, true
}

func compileAllowlists(allowlists []allowlist) []compiledAllowlist {
	compiled := make([]compiledAllowlist, 0, len(allowlists))
	for _, entry := range allowlists {
		item := compiledAllowlist{
			condition: entry.Condition,
			target:    strings.ToLower(entry.RegexTarget),
			stopwords: lowerStrings(entry.Stopwords),
		}
		for _, pattern := range entry.Paths {
			if re, err := regexp.Compile(pattern); err == nil {
				item.paths = append(item.paths, re)
			}
		}
		for _, pattern := range entry.Regexes {
			if re, err := regexp.Compile(pattern); err == nil {
				item.regexes = append(item.regexes, re)
			}
		}
		compiled = append(compiled, item)
	}
	return compiled
}

func lowerStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	lowered := make([]string, len(values))
	for i, value := range values {
		lowered[i] = strings.ToLower(value)
	}
	return lowered
}

// scanContent runs every rule against one file's added-line blob.
func scanContent(path, blob string) []rawFinding {
	rules, globalAllowlists := compiledRuleSet()

	if allowlistSkipsPath(globalAllowlists, path) {
		return nil
	}

	lowerBlob := strings.ToLower(blob)
	var findings []rawFinding

	for _, compiled := range rules {
		if compiled.path != nil && !compiled.path.MatchString(path) {
			continue
		}
		if ruleAllowlistsSkipPath(compiled.allowlists, path) {
			continue
		}

		if compiled.regex == nil {
			// Path-only rule (for example a PKCS#12 file or .env).
			if compiled.path != nil && compiled.path.MatchString(path) {
				findings = append(findings, rawFinding{
					ruleID:    compiled.rule.ID,
					fileLevel: true,
				})
			}
			continue
		}

		if len(compiled.keywords) > 0 && !containsAny(lowerBlob, compiled.keywords) {
			continue
		}

		for _, matchIndex := range compiled.regex.FindAllStringSubmatchIndex(blob, -1) {
			start := matchIndex[0]
			// Mirror gitleaks: trailing newlines can be part of a regex match
			// because of \s/$ alternatives; trim them before locating the span.
			match := strings.Trim(blob[start:matchIndex[1]], "\n")
			if match == "" {
				continue
			}
			end := start + len(match)
			secret, ok := secretFromMatch(match, compiled.regex, compiled.rule.SecretGroup)
			if !ok {
				continue
			}

			if compiled.rule.Entropy != 0 && shannonEntropy(secret) <= compiled.rule.Entropy {
				continue
			}

			line := lineAt(blob, start)
			if strings.Contains(line, gitleaksAllowSignature) {
				continue
			}
			if allowlistsMatchFinding(globalAllowlists, path, line, match, secret) {
				continue
			}
			if allowlistsMatchFinding(compiled.allowlists, path, line, match, secret) {
				continue
			}

			findings = append(findings, rawFinding{
				ruleID: compiled.rule.ID,
				start:  start,
				end:    end,
				match:  match,
				secret: secret,
			})
		}
	}

	return findings
}

// secretFromMatch mirrors gitleaks: when secretGroup is set use that capture
// group, otherwise use the first non-empty capture group.
func secretFromMatch(match string, re *regexp.Regexp, secretGroup int) (string, bool) {
	groups := re.FindStringSubmatch(match)
	if len(groups) < 2 {
		return match, true
	}
	if secretGroup > 0 {
		if len(groups) <= secretGroup {
			return "", false
		}
		return groups[secretGroup], true
	}
	for _, group := range groups[1:] {
		if group != "" {
			return group, true
		}
	}
	return match, true
}

func containsAny(haystack string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

// allowlistSkipsPath reports whether a global allowlist path check skips the
// whole file (gitleaks semantics for OR allowlists).
func allowlistSkipsPath(allowlists []compiledAllowlist, path string) bool {
	for _, item := range allowlists {
		if item.condition == allowlistAND {
			continue
		}
		if anyRegexMatch(item.paths, path) {
			return true
		}
	}
	return false
}

// ruleAllowlistsSkipPath reports whether a rule's own path allowlist skips the
// file before any content matching happens.
func ruleAllowlistsSkipPath(allowlists []compiledAllowlist, path string) bool {
	for _, item := range allowlists {
		pathAllowed := anyRegexMatch(item.paths, path)
		if item.condition == allowlistAND {
			if len(item.regexes) > 0 || len(item.stopwords) > 0 {
				// Deferred to the per-finding check.
				continue
			}
			if len(item.paths) > 0 && pathAllowed {
				return true
			}
			continue
		}
		if pathAllowed {
			return true
		}
	}
	return false
}

func allowlistsMatchFinding(allowlists []compiledAllowlist, path, line, match, secret string) bool {
	for _, item := range allowlists {
		target := secret
		switch item.target {
		case "match":
			target = match
		case "line":
			target = line
		}

		regexAllowed := anyRegexMatch(item.regexes, target)
		stopwordAllowed := containsAny(strings.ToLower(secret), item.stopwords)
		pathAllowed := anyRegexMatch(item.paths, path)

		if item.condition == allowlistAND {
			checks := make([]bool, 0, 3)
			if len(item.paths) > 0 {
				checks = append(checks, pathAllowed)
			}
			if len(item.regexes) > 0 {
				checks = append(checks, regexAllowed)
			}
			if len(item.stopwords) > 0 {
				checks = append(checks, stopwordAllowed)
			}
			if len(checks) > 0 && allTrue(checks) {
				return true
			}
			continue
		}

		if pathAllowed || regexAllowed || stopwordAllowed {
			return true
		}
	}
	return false
}

func anyRegexMatch(regexes []*regexp.Regexp, value string) bool {
	for _, re := range regexes {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func allTrue(values []bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}

// shannonEntropy matches gitleaks' implementation so the vendored entropy
// thresholds keep the same behavior.
func shannonEntropy(data string) (entropy float64) {
	if data == "" {
		return 0
	}
	charCounts := make(map[rune]int)
	for _, char := range data {
		charCounts[char]++
	}
	invLength := 1.0 / float64(len(data))
	for _, count := range charCounts {
		freq := float64(count) * invLength
		entropy -= freq * math.Log2(freq)
	}
	return entropy
}
