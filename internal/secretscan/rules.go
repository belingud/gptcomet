package secretscan

// rule mirrors the subset of a gitleaks rule that the default gitleaks rule set
// actually uses. The generated rules in rules_gen.go are built from
// internal/secretscan/rules/gitleaks.toml; local rules for sensitive file names
// are declared in this file.
type rule struct {
	ID          string
	Regex       string
	SecretGroup int
	Entropy     float64
	Keywords    []string
	Path        string
	Allowlists  []allowlist
}

// allowlistCondition selects whether all or any configured checks must match.
type allowlistCondition int

const (
	allowlistOR allowlistCondition = iota
	allowlistAND
)

// allowlist mirrors a gitleaks rule/global allowlist entry. Commit allowlists do
// not apply because the scanner only sees a diff, not commit metadata.
type allowlist struct {
	Description string
	Condition   allowlistCondition
	Paths       []string
	Regexes     []string
	RegexTarget string
	Stopwords   []string
}

// localRules are maintained in this repository, not generated. They flag files
// whose whole content is sensitive even when no individual line matches a
// credential regex (for example a .env file with an opaque value).
var localRules = []rule{
	{
		ID:   "sensitive-env-file",
		Path: `(?i)(?:^|/)\.env(?:\.[^/]+)?$`,
		Allowlists: []allowlist{
			{
				Description: "checked-in templates are not secrets",
				Paths:       []string{`(?i)(?:^|/)\.env\.(?:example|sample|template|dist)$`},
			},
		},
	},
	{
		ID:   "sensitive-private-key-file",
		Path: `(?i)(?:^|/)(?:id_(?:rsa|ed25519|ecdsa|dsa)|[^/]+\.(?:pem|key|p12|pfx|jks|keystore))$`,
	},
	{
		ID:   "sensitive-credential-file",
		Path: `(?i)(?:^|/)(?:credentials(?:\.json)?|\.npmrc|\.pypirc|\.netrc)$`,
	},
}
