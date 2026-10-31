# Add sensitive data pre-send scanning with gitleaks rules

## Goal

Before a diff is sent to any LLM provider (`commit`, `generate`, the Git hook, and `review`
including piped stdin diffs), scan the added lines locally for credentials and secrets. Default
behavior warns; users can opt into `redact` or `block`. The scanner reuses the MIT-licensed
gitleaks v8.30.1 rule data but ships a lightweight in-repo matcher — no gitleaks library or CLI
dependency, and no new Go module dependencies.

## Requirements

1. **Only added lines are scanned.** The unified diff is parsed into per-file added-line content;
   deleted/context lines are ignored, and the real diff/commit content is never modified.
2. **Rule data comes from gitleaks v8.30.1** (`config/gitleaks.toml`, 222 rules), vendored under
   `internal/secretscan/rules/` with its MIT license text and upstream version/commit recorded.
   A maintainer script generates `rules_gen.go`; the runtime carries no TOML parser dependency.
3. **Matcher supports the features the default rule set actually uses:** regex + `secretGroup`,
   keyword prefilter, Shannon entropy, path rules, and per-rule allowlists (regexes/stopwords/
   paths, AND/OR). Not ported: base64/hex decoding, archive/history/directory scanning,
   `.gitleaksignore`/baselines, `required` proximity rules (zero rules in the default set).
4. **Local path rules catch sensitive files** (`.env*`, `id_rsa`/`id_ed25519`, `*.pem`, `*.key`,
   `*.p12`, `*.pfx`, `credentials*`, `.npmrc`, `.pypirc`, `.netrc`) even when no credential regex
   matches. This must not depend on `file_ignore` and must not change its defaults.
5. **Modes:** `off`/`warn`/`redact`/`block`; default `warn`.
   - `warn`: send unchanged, print masked findings to stderr.
   - `redact`: replace every finding's span (or whole added content for sensitive-file findings)
     in the diff that is sent, keeping line structure intact.
   - `block`: send nothing; return a structured error listing findings.
6. **Configuration:** `sensitive.mode` and `sensitive.allow` (list of regexes matched against the
   finding secret or line; matching findings are ignored). Keys are added to `DefaultConfig`,
   `GetSupportedKeys`, `config validate`, README/README.zh/CLAUDE docs, and the Go spec.
7. **CLI override:** `--secrets <mode>` on `commit`, `generate`, and `review` overrides the config
   for that invocation. `off` is the documented one-shot bypass.
8. **Output contract:** findings are printed to stderr with masked previews only — never the full
   secret, never stdout. `gmsg generate` keeps stdout limited to the message. `block` errors are
   `*errors.GPTCometError` from a new template.
9. **Hook semantics unchanged:** the hook remains fail-open; in `block` mode it does not send the
   diff and does not block `git commit`, but the user still sees the warning on stderr.
10. **Docs and tracking:** update README.md, README.zh.md, CLAUDE.md, the Go config spec, and mark
    FUTURE_FEATURES.md item 8 as implemented.

## Non-Goals

- Embedding the gitleaks library or shelling out to a `gitleaks` binary.
- Base64/hex encoded-secret decoding, git history/archive scanning, baseline or
  `.gitleaksignore` files.
- Changing default `file_ignore` (for example adding `.env`) or blocking commits themselves.
- Automatic release/publish (version bump and release stay a separate user-approved step).

## Acceptance Criteria

- [x] `internal/secretscan` has table-driven tests for diff parsing, each rule family, entropy,
      allowlists, all four modes, and redaction preserving diff line counts.
- [x] Known fixtures are detected (AWS key, GitHub PAT, OpenAI key, PEM header, `.env` line);
      documented false positives are not (AWS `...EXAMPLE` sample key, git SHA/lockfile hash,
      UUID, placeholder password, `{{ ... }}` template values).
- [x] `commit` / `generate` / hook / `review` (staged) / `review` (piped) all pass through the
      scanner; tests cover the commit and review service paths with mocks and no network.
- [x] `sensitive.mode` invalid values are rejected by `config validate` and by command execution
      with an actionable error.
- [x] No test or runtime output contains a full secret value.
- [x] `just format`, `just check`, `just test` and `go test -race ./...` pass.
- [x] Docs and FUTURE_FEATURES.md updated; gitleaks MIT attribution present in the repo.

## Notes

- Task artifacts: `design.md` for the technical design, `implement.md` for the ordered execution
  and validation plan.
- Write scopes: `internal/secretscan/`, `internal/errors/`, `internal/config/`,
  `pkg/config/defaults/`, `cmd/`, `scripts/`, docs, `FUTURE_FEATURES.md`.
- gitleaks license: MIT, Copyright (c) 2019 Zachary Rice — the vendored rules and generated file
  must keep the copyright/permission notice.
