# Implement — Sensitive data pre-send scanning

Ordered checklist. Each step ends with the listed tests/validation; do not start the next step on a
red test in the previous one.

## 0. Rules data + codegen (no product code yet)

- [x] Vendor `gitleaks.toml` v8.30.1 + `LICENSE` + `UPSTREAM.md` under
      `internal/secretscan/rules/`.
- [x] Write `scripts/gen_secret_rules.py` (stdlib `tomllib`) emitting
      `internal/secretscan/rules_gen.go`.
- [x] Add `sync-secret-rules` recipe to `justfile` (regenerate from vendored TOML + `gofmt`).
- [x] Sanity: 222 generated rules, all regexes compile (`go test` once the package exists).

## 1. `internal/secretscan` engine

- [x] `diff.go`: unified-diff parser → `[]addedLine` (+ plain-text `stdin` fallback).
- [x] `matcher.go`: keyword prefilter, regex/secretGroup, entropy, path rules, allowlists,
      user `Options.Allow`, dedupe/sort.
- [x] `rules.go`: `rule`/`allowlist` types, local sensitive-file rules, lazy rule compilation.
- [x] `modes.go`: `Mode`/`ParseMode`/`Options`/`Result`/`Apply`, masked preview + redaction.
- [x] Tests: `diff_test.go`, `matcher_test.go`, `modes_test.go` (fixtures from gitleaks rules:
      AWS key, GitHub PAT, OpenAI key, PEM header, `.env`; false positives: `...EXAMPLE`,
      git SHA/lockfile hash, UUID, `{{ ... }}`, placeholder password; entropy; allowlists).
- [x] `go test ./internal/secretscan/` green.

## 2. Errors

- [x] `SensitiveFinding` type + `SensitiveContentError`,
      `InvalidSensitiveModeError`, `InvalidSensitiveAllowError` in `internal/errors/`.
- [x] Tests in `internal/errors/errors_test.go` assert masked message, suggestions, docs URL, and
      that a full secret never appears.
- [x] `go test ./internal/errors/` green.

## 3. Config keys + defaults + docs spec

- [x] `pkg/config/defaults`: `DefaultSensitiveMode`, `sensitive` block in `DefaultConfig`.
- [x] `internal/config/accessor.go`: `sensitive.mode`, `sensitive.allow` in `GetSupportedKeys`.
- [x] `internal/config/validate.go`: mode value + allow list type/regex validation.
- [x] Tests: `defaults_test.go`, `accessor_test.go` supported keys, `validate_test.go` cases.
- [x] `.trellis/spec/go/config-and-defaults.md`: add the key family + defaults.
- [x] `go test ./internal/config/ ./pkg/config/defaults/` green.

## 4. Command integration

- [x] `cmd/sensitive.go`: option resolution from config/flag, stderr warning renderer.
- [x] `CommitService.getStagedDiff()`: `Apply`, warn/redact/block handling, `stderr io.Writer`.
- [x] `ReviewService.getDiff()`: same for piped + staged.
- [x] `--secrets` flag on `commit`, `generate`, `review`; options plumbed through.
- [x] Update existing cmd tests that now touch `sensitive.*` mocks (add `GetWithDefault`
      expectations); add new service-path tests for each mode.
- [x] `cmd/flags_test.go` / command tests cover the new flag.
- [x] `just format` then `just test` (cmd + touched packages).

## 5. Docs + tracking

- [x] README.md / README.zh.md: config keys, `--secrets` flags, security/pre-send scanning
      section (what is not detected: encoded secrets, history).
- [x] CLAUDE.md: architecture bullet for `internal/secretscan` + supported keys.
- [x] FUTURE_FEATURES.md item 8 → ✅ with version placeholder for the release.
- [x] `docs/` note if the README cross-links it.

## 6. Quality gate (blocking)

```bash
export GOCACHE=/tmp/gptcomet-gocache GOMODCACHE=/tmp/gptcomet-go-mod GOTMPDIR=/tmp
export STATICCHECK_CACHE=/tmp/gptcomet-staticcheck
just format
PATH=/tmp/gptcomet-bin:$PATH just check
just test
go test -race ./...
```

- [x] All four commands pass; `git diff` shows no formatting churn outside touched files.
- [x] Re-read the final diff for accidental secret fixtures committed in tests (use obviously fake
      keys, e.g. `AKIAIOSFODNN7EXAMPLE`, `ghp_` + 36 filler chars).
- [x] Present commit plan (file list + message) and get user confirmation before committing.

## Rollback points

- Step 4 introduces the only behavior change (request content). If a regression is found after
  merge, `sensitive.mode: off` disables it without reverting code; a full `git revert` of the
  feature commit is safe (no persisted schema or state).
