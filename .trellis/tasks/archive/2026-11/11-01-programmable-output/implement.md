# Implement — Programmable output: stdout/stderr contract + exit codes

Ordered checklist. Do not start a step while the previous step's validation is red.

## 0. Baseline

- [x] `GOCACHE=/tmp/gptcomet-gocache GOMODCACHE=/tmp/gptcomet-go-mod GOTMPDIR=/tmp GOSUMDB=off go test ./...`
      (baseline is green before any edit; staticcheck binary is
      `/tmp/gptcomet-bin/staticcheck` because system staticcheck 2026.1 cannot read Go 1.27.3
      export data).

## 1. Exit-code policy (`internal/errors`)

- [x] Add `internal/errors/exit.go`: `ExitSuccess/Unknown/Usage/Config/Network/Git/API/Validation`
      constants, `UsageError` wrapper, `NewUsageError`, `ExitCode(err)`.
- [x] Add `internal/errors/exit_test.go`: nil → 0; config/network/git/api/validation/unknown types
      → 3/4/5/6/7/1; `fmt.Errorf("%w")` wrapper still classified; `NewUsageError` → 2; plain
      error → 1; wrapped usage error → 2.
- [x] Validate: `go test ./internal/errors/`.

## 2. `main.go` wiring

- [x] `rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return gptcometerrors.NewUsageError(err) })`.
- [x] Replace `os.Exit(1)` with `os.Exit(gptcometerrors.ExitCode(err))`; keep the existing
      commented-out `fmt.Fprintln` note removed/updated.
- [x] Smoke: `go run . commit --bogus-flag` → exit 2 (check `echo $?`); `go run . config get` with
      a missing config still exits 1/3 as classified.

## 3. Leaf packages: status → stderr

- [x] `internal/llm/llm.go`: request line + usage → `fmt.Fprintf(os.Stderr, ...)`.
- [x] `internal/llm/gemini.go`, `internal/llm/groq.go`: usage → stderr.
- [x] `internal/config/parser.go`: `Warning:` lines → stderr.
- [x] `internal/ui/progress.go`: all `fmt.Printf/Println` → `fmt.Fprintf/Fprintln(os.Stderr, ...)`.
- [x] Update `internal/ui/progress_test.go`: capture `os.Stderr` instead of `os.Stdout` in every
      pipe-based test.
- [x] Validate: `go test ./internal/llm/ ./internal/config/ ./internal/ui/`.

## 4. `generate` cleanup

- [x] `cmd/generate.go`: delete `generateWithCleanStdout`, call `service.GenerateMessage()`
      directly, keep `fmt.Fprintln(cmd.OutOrStdout(), message)`.
- [x] `cmd/generate_test.go`: remove the two `generateWithCleanStdout` tests; replace with a test
      that the command's stdout writer receives the message (add a small `newGenerateService`
      seam if needed) or keep `TestNewGenerateCmd` plus a comment-level contract test.
- [x] Validate: `go test ./cmd/ -run Generate`.

## 5. `review` service contract + `--raw`

- [x] `ReviewOptions.Raw bool`; `--raw` general flag with help text (mention stream no-op).
- [x] `ReviewService.stdout io.Writer` (default `os.Stdout`), wiring in `NewReviewCmd` RunE:
      `service.stdout = cmd.OutOrStdout(); service.stderr = cmd.ErrOrStderr()`.
- [x] `Execute`: banner + reminders → stderr; final output → stdout.
- [x] `ExecuteStream`: reminder → stderr; chunks/trailing newline → stdout.
- [x] `formatReviewComment`: raw mode returns the comment unchanged (no glamour, no border).
- [x] Tests:
  - [x] `TestReviewService_Execute_OutputContract` (buffers; stdout has review, stderr has banner
        and reminder; no provider banner in stdout).
  - [x] `TestReviewService_Execute_Raw` (renderer not called; stdout is verbatim Markdown).
  - [x] `TestReviewService_ExecuteStream_OutputContract` (chunks stdout, reminder stderr).
  - [x] Existing stdout-capturing tests still pass (or migrate them to the service writers).
- [x] Validate: `go test ./cmd/ -run Review`.

## 6. `commit` service contract

- [x] `CommitService.stdout io.Writer` (default `os.Stdout`), wiring in `NewCommitCmd` RunE.
- [x] `Execute`: banner, generation errors, `handleCommitInteraction` previews/prompts/cancel/
      invalid-option/error lines, and `createCommit` summary → stderr.
- [x] `--dry-run` keeps boxed message on stdout.
- [x] Tests: `TestCommitService_Execute_DryRunOutputContract` (stdout has boxed message, stderr has
      banner); interactive path writes nothing to stdout (use `AutoYes` + mock VCS).
- [x] Validate: `go test ./cmd/ -run Commit`.

## 7. Docs

- [x] README.md: new "Scripting / automation" subsection (stream contract, exit table, examples),
      update the `generate` paragraph, mention `review --raw`.
- [x] README.zh.md: same in Chinese.
- [x] CLAUDE.md (AGENTS.md symlink): output-contract bullet under Key Patterns; command list note
      for `review --raw`.
- [x] `.trellis/spec/go/cli-and-output.md`: add the contract + exit-code table; remove Known Debt;
      Terminal Output notes progress → stderr; command-shape note for `--raw`.
- [x] FUTURE_FEATURES.md item 5 → ✅ with version placeholder.
- [x] Validate: `grep -rn "stdout" README.md README.zh.md` sanity pass; no contradictions with the
      "status and token-usage output goes to stderr" claim.

## 8. Full verification

- [x] `just format` (or `go fmt ./... && goimports -w .`).
- [x] `go vet ./...`.
- [x] `/tmp/gptcomet-bin/staticcheck ./...`.
- [x] `go test ./...` (full suite) and `go test -race ./cmd/ ./internal/errors/ ./internal/ui/`.
- [x] Manual smoke in a throwaway repo:
  - [x] `gmsg generate >/tmp/msg 2>/tmp/err` → `/tmp/msg` has only the message; `/tmp/err` has the
        request line/usage.
  - [x] `gmsg review >/tmp/out 2>/tmp/err` → `/tmp/out` has review only; `/tmp/err` has banner.
  - [x] `gmsg review --raw` output has no ANSI/border.
  - [x] exit codes: `--bogus-flag` → 2; no staged changes → 5; invalid `--secrets` value → 7.
- [x] Record results in the task journal / final report.

## 9. Wrap-up (Phase 3)

- [x] Update `.trellis/spec/go/cli-and-output.md` and, if warranted, `error-handling.md` with the
      exit-code convention.
- [ ] One-shot commit plan for user approval (feat + docs + tests; no version bump/release unless
      asked separately).
- [ ] Archive task after commit (Trellis finish-work).

## Rollback points

- Steps 1-3 are independent and revert trivially (stream routing only).
- Step 4 depends on step 3; if any direct stdout write is missed, keep `generateWithCleanStdout`
  for one more commit rather than shipping a leaky stdout contract.
- Steps 5-6 are service-local; revert by removing the writer fields and the `--raw` flag.
- Docs and FUTURE_FEATURES are last so no doc claims an unshipped behavior.

## Risky files

- `main.go` exit path (affects every command).
- `cmd/review.go` / `cmd/commit_action.go` (interactive UX; terminal users must see the same
  lines).
- `internal/ui/progress.go` tests (pipe capture moved from stdout to stderr).
- `cmd/generate.go` (removing the safety net; requires step 3 fully done).

## Execution Evidence (2026-11-01)

- `gofmt -l .` / `goimports -l .`: clean.
- `go vet ./...`: clean.
- `/tmp/gptcomet-bin/staticcheck ./...` with `STATICCHECK_CACHE=/tmp/staticcheck-cache`: clean.
- `go test ./...`: all packages ok.
- `go test -race ./...`: all packages ok (exit 0).
- `ruff check py/`: all checks passed.
- End-to-end smoke against a local mock OpenAI-compatible HTTP server (no real network):
  - `gmsg generate >msg 2>err`: `msg` contains only the commit message; `err` contains
    `📤 Sending request...` and the token-usage line.
  - `gmsg review >out 2>err`: `out` contains only the rendered review; `err` contains the
    provider banner, reminder, request line, and usage.
  - `gmsg review --raw`: `out` is the model Markdown verbatim (`# Review` + bullets), no border/ANSI.
  - `gmsg commit --dry-run`: stdout holds only the boxed "Generated commit message" block; the
    banner/request/usage are on stderr.
  - `gmsg commit -y`: stdout is empty; preview, status, and commit summary are on stderr.
- Exit-code smoke: unknown flag → 2; missing `--config` path → 3; no staged changes → 5; unroutable
  provider endpoint (retries exhausted) → 6; invalid `--secrets bogus` → 7; successful
  generate/review/commit → 0.
