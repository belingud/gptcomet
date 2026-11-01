# Programmable output: stdout/stderr contract + exit codes

## Goal

Make GPTComet safe to drive from scripts, editor plugins, and CI: `stdout` carries only the
machine-consumable result, every status/progress/warning/error line moves to `stderr`, and the
process exit code identifies the failure class. `generate` already behaves this way; `commit` and
`review` do not. Add `review --raw` so callers can capture the model's original Markdown.

This completes FUTURE_FEATURES.md item 5 (输出可编程化), currently marked 🟡 partial.

## Background (confirmed by repo inspection)

- **`generate` works by global stdout hijack, not by clean internals.** `cmd/generate.go:78-85`
  temporarily swaps `os.Stdout` to `os.Stderr` while generating because lower layers print
  directly to stdout: `internal/llm/llm.go:248` (`📤 Sending request...`), `internal/llm/llm.go:272`
  (token usage), `internal/llm/gemini.go:195`, `internal/llm/groq.go:177`,
  `internal/config/parser.go:21-49` (config warnings).
- **`review` mixes result and status on stdout.** `cmd/review.go:129` prints the provider/model
  banner, `:191` / `:300` print "Reviewing…" reminders, `:200` streams result chunks, `:158`
  prints the final review, and `internal/ui/progress.go` prints progress to stdout
  (`cmd/review.go:107`). Non-stream output is glamour-rendered and wrapped in a `=` border by
  `formatHighlightedMessage` (`cmd/review.go:309-325`, `cmd/common.go:59-65`).
- **`commit` prints interactive chrome on stdout.** Provider banner `cmd/commit_action.go:227`,
  generation errors `:247`/`:330`, dry-run preview `:256-258`, interactive preview `:295`,
  prompt `:302`, cancellation/invalid-option messages `:317`/`:341`, final commit info `:375`.
  All go to stdout today; there is no stdout contract at all.
- **Exit codes are always 1 on failure.** `main.go:70-73` calls `os.Exit(1)` for every error even
  though `internal/errors` already has a 6-category `GPTCometError` (`internal/errors/errors.go:10-18`)
  and the config/git/client layers already use it for most failures.
- **`review` output cannot be captured as Markdown.** The default path always runs glamour plus the
  border; there is no way to get the model's original Markdown for a file/CI comment.

## Requirements

1. **stdout contract per command**
   - `generate`: stdout contains exactly the commit message (unchanged behavior, but achieved
     without mutating `os.Stdout`).
   - `review`: stdout contains only the review result. In stream mode the result chunks go to
     stdout; in non-stream mode the rendered review (or raw Markdown with `--raw`) goes to stdout.
   - `commit`: stdout is empty except in `--dry-run`, where it keeps the current boxed message
     output. Interactive previews, prompts, status, warnings, and the final commit summary go to
     stderr. Scripts should use `generate` for a clean message.
   - Hook management commands (`hook install/uninstall/status`) are human-facing and keep their
     current stdout output (out of scope).
2. **Every status line goes to stderr**: provider/model banner, progress stages, "Reviewing…"
   reminders, request/usage lines from the LLM layer, config parse warnings, sensitive-data
   warnings, interactive prompts and errors.
3. **Remove the `generateWithCleanStdout` global `os.Stdout` swap** once the leaf packages no
   longer write to stdout.
4. **Exit codes**
   - `0` success, `1` unexpected/unknown failure, `2` command-line usage error (unknown flag or
     invalid flag value), `3` configuration error, `4` network error, `5` Git/VCS error,
     `6` API/LLM error, `7` validation/content error (invalid input or config value; includes
     `sensitive.mode=block` interception).
   - Mapping derives from `*GPTCometError.Type`; errors that are not structured fall back to `1`.
     Cobra flag-parsing errors are marked as usage errors (`2`).
5. **`review --raw` flag**: non-stream review prints the model's Markdown verbatim to stdout, with
   no glamour rendering, no ANSI styling, and no border. With `--stream`, output is already raw;
   the flag is documented as a no-op there.
6. **Tests**: unit tests for the exit-code mapping; service-level tests asserting the stdout/stderr
   split for `review` (normal + raw + stream) and `commit --dry-run`; progress tests updated to
   stderr.
7. **Docs**: README.md and README.zh.md get a "scripting" section with the stream contract, exit
   code table, and examples (`gmsg generate > msg`, `gmsg review --raw > review.md`,
   `gmsg review --stream | pbcopy`). CLAUDE.md/AGENTS.md and
   `.trellis/spec/go/cli-and-output.md` are updated (remove the "known debt" note, record the new
   contract). FUTURE_FEATURES.md item 5 is marked implemented.

## Acceptance Criteria

- [x] `gmsg generate` writes only the message to stdout; the LLM request line and token usage
      appear on stderr; no code path swaps `os.Stdout`.
- [x] `gmsg review` (mocked service) writes the provider banner and "Reviewing…" reminder to
      stderr and only the review result to stdout; `--stream` chunks go to stdout only.
- [x] `gmsg review --raw` writes the model's Markdown verbatim to stdout without invoking glamour
      or the border helper.
- [x] `gmsg commit --dry-run` writes the boxed message to stdout and the provider banner to
      stderr; the interactive commit flow produces no stdout output.
- [x] Exit codes are covered by tests: nil → 0; each `GPTCometError` type → 3/4/5/6/7; usage error
      → 2; plain error → 1. `main.go` exits with `errors.ExitCode(err)`.
- [x] A smoke check shows `gmsg <cmd> --bogus-flag` exits 2 and a command with no staged changes
      exits 5.
- [x] `go test ./...`, `go vet ./...`, `staticcheck ./...`, and `gofmt` are clean.
- [x] README.md, README.zh.md, CLAUDE.md, `.trellis/spec/go/cli-and-output.md`, and
      FUTURE_FEATURES.md item 5 are updated.

## Key Decisions

- **Exit codes are classified, not just boolean.** Scripts checking `!= 0` are unaffected;
  scripts that depended on the exact previous code `1` may observe other codes. The contract is
  documented as a feature of this change.
- **Status output always goes to stderr, even on a TTY.** Terminal users see the same lines (same
  terminal); only piping separates the streams.
- **`commit --dry-run` keeps its decorated output.** `generate` is the documented scripting entry
  point; changing the dry-run format would break more expectations than it buys. Only the
  interactive chrome moves to stderr.
- **`review --raw` does not alter stream mode.** Stream output is already unformatted; the flag is
  documented as affecting the non-stream rendering path only.
- **Sensitive `block` maps to code 7 (validation)** instead of adding a seventh error type; the
  detailed remediation stays in the structured error on stderr.

## Out of Scope

- Structured review JSON (`--format json` with file/line/severity) and `--fail-on` — FUTURE_FEATURES
  item 2.
- Review range flags (`--base`, `--commit`, `--last`, `--path`) — item 2.
- `generate --json` (considered as option C and deferred).
- Output streams of `config`, `provider`, `update`, and `hook` management subcommands.
- Version bump, changelog, commit, and release — separate user-approved steps.

## Notes

- Write scopes: `main.go`, `cmd/`, `internal/errors/`, `internal/llm/`, `internal/config/parser.go`,
  `internal/ui/progress.go`, README.md, README.zh.md, CLAUDE.md, `.trellis/spec/go/`,
  `FUTURE_FEATURES.md`.
- Tests stay hermetic: service tests use the existing mocks (`MockVCS`, `MockClient`,
  `testutils.MockConfigManager`); no network calls.
