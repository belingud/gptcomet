# Design — Programmable output: stdout/stderr contract + exit codes

## Boundary

Two layers change:

1. **Leaf packages stop writing to stdout.** `internal/llm` (request line, usage), and
   `internal/config/parser.go` (config warnings) write to `os.Stderr` directly. `internal/ui/progress`
   writes to `os.Stderr` directly. These packages are leaves in the output sense: they are not
   given writers today, and threading a writer through every provider would be a much larger change
   than the feature is worth.
2. **`cmd` owns the result stream.** `CommitService` and `ReviewService` gain a `stdout io.Writer`
   field (mirroring the existing `stderr` field) and route every status line through their
   `stderr` writer. `main.go` owns the exit code.

`internal/errors` gains the exit-code policy and the usage-error marker so both `main.go` and
tests can share it. No new packages, no new module dependencies.

```
main.go ── Execute() → errors.ExitCode(err) → os.Exit
   │                       ▲
   │ SetFlagErrorFunc ──────┘ (*errors.UsageError → 2)
   ▼
cmd/commit.go, cmd/review.go, cmd/generate.go
   │  service.stdout = cmd.OutOrStdout()
   │  service.stderr = cmd.ErrOrStderr()
   ▼
CommitService / ReviewService ── result → stdout
   │                            └─ status/prompts/reminders → stderr
   ▼
internal/llm, internal/config/parser, internal/ui/progress ── status → os.Stderr
```

## Output contract

| Command / mode | stdout | stderr |
| --- | --- | --- |
| `generate` | commit message only | status, request/usage lines, config warnings, secret-scan warnings, errors |
| `review` (non-stream) | rendered review (glamour + border) | provider banner, "Reviewing…", progress, warnings, errors |
| `review --raw` | model Markdown verbatim (no ANSI, no border) | same as above |
| `review --stream` | streamed review chunks + final newline | reminder, progress, warnings, errors |
| `review --stream --raw` | stream chunks (raw already) | same as above; `--raw` is a documented no-op |
| `commit` (interactive / `-y`) | empty | provider banner, progress, message previews, prompts, errors, commit summary |
| `commit --dry-run` | current boxed "Generated commit message" block | provider banner, progress, warnings, errors |
| `hook install/uninstall/status` | unchanged (human commands) | unchanged |

`generate`'s stdout was already pure; this change removes the mechanism that made it pure
(`cmd/generate.go` `generateWithCleanStdout`) and instead guarantees purity at the source.

## Exit code contract (`internal/errors/exit.go`)

```go
const (
    ExitSuccess     = 0
    ExitUnknown     = 1 // unexpected / unclassified failure
    ExitUsage       = 2 // unknown flag, invalid flag value
    ExitConfig      = 3 // configuration errors
    ExitNetwork     = 4 // network/proxy/connection errors
    ExitGit         = 5 // git/VCS errors (including "no staged changes")
    ExitAPI         = 6 // API/LLM request, response, parsing errors
    ExitValidation  = 7 // invalid input/config value; sensitive.mode=block
)

func ExitCode(err error) int
```

Mapping rules, in order:

1. `err == nil` → `ExitSuccess`.
2. `errors.As(err, *UsageError)` → `ExitUsage`.
3. `errors.As(err, *GPTCometError)` → map `Type`: config → 3, network → 4, git → 5, api → 6,
   validation → 7, unknown → 1. Because it uses `errors.As`, `fmt.Errorf("...: %w", err)` wrappers
   (used throughout `cmd`) still classify correctly.
4. Everything else → `ExitUnknown`.

`UsageError` is a tiny wrapper (`Error()` delegates to the inner error, `Unwrap()` exposes it) so
Cobra prints the normal message but `ExitCode` can recognise it. `main.go` installs
`rootCmd.SetFlagErrorFunc(...)`; Cobra inherits the func on subcommands. Argument-count errors from
`cobra.Args` are not flagged as usage errors in this iteration (they fall back to `1`); the contract
documents code `2` as "flag parsing/usage".

`main.go` changes from `os.Exit(1)` to `os.Exit(gptcometerrors.ExitCode(err))`. Cobra continues to
print the error to stderr exactly once; `Execute` output is not duplicated.

## Code changes by file

- `internal/errors/exit.go` (new): constants, `UsageError`, `NewUsageError`, `ExitCode`.
- `internal/errors/exit_test.go` (new): table-driven mapping tests.
- `main.go`: flag error func + exit-code call; import `internal/errors`.
- `internal/llm/llm.go`: `📤 Sending request...` and usage → `os.Stderr`.
- `internal/llm/gemini.go`, `internal/llm/groq.go`: usage → `os.Stderr`.
- `internal/config/parser.go`: `Warning:` lines → `os.Stderr`.
- `internal/ui/progress.go`: every write → `os.Stderr`; tests updated accordingly.
- `cmd/generate.go`: delete `generateWithCleanStdout`; plain `service.GenerateMessage()` then
  `fmt.Fprintln(cmd.OutOrStdout(), message)`.
- `cmd/review.go`:
  - `ReviewService.stdout io.Writer` (default `os.Stdout` in constructor; tests can inject).
  - `Execute`: wire `stdout`/`stderr` from Cobra in `RunE`; banner + reminders → `stderr`;
    progress already goes to stderr; final review → `stdout`.
  - `ExecuteStream`: reminder → `stderr`; chunks + trailing newline → `stdout`.
  - `ReviewOptions.Raw bool` + `--raw` flag; `formatReviewComment` short-circuits in raw mode and
    returns the original Markdown (no glamour, no `formatHighlightedMessage`).
- `cmd/commit_action.go` / `cmd/commit.go`:
  - `CommitService.stdout io.Writer` (default `os.Stdout`; wire from `cmd.OutOrStdout()` in
    `RunE`).
  - Banner, generation errors, interactive previews/prompts/cancel/invalid-option messages, and
    `createCommit` summary → `stderr`.
  - `--dry-run` keeps `fmt.Fprintf(s.stdout, "\nGenerated commit message:\n%s\n", boxed)`.
- Docs: README.md, README.zh.md (new "Scripting / exit codes" subsection near the Git hook /
  generate docs), CLAUDE.md (output contract note), `.trellis/spec/go/cli-and-output.md`
  (contract + remove Known Debt; Terminal Output: progress → stderr), FUTURE_FEATURES.md item 5.

## Compatibility and risks

- **Exit codes** change from always `1` to classified codes. `!= 0` checks keep working; exact
  `== 1` checks may see `2-7` instead. Called out in README and release notes.
- **Combined-output ordering** (`2>&1`) may interleave differently because stdout is now only the
  result. Single-stream terminal UX is unchanged (both streams hit the same TTY). Tests that
  captured `os.Stdout` for progress must capture `os.Stderr`.
- **`formatReviewComment` fallback**: when glamour fails, the existing behavior returns the
  original comment plus the render error. `--raw` bypasses that path entirely.
- **Hook** is unaffected: it calls `GenerateMessage()` in-process and writes the message file;
  it never parses stdout. `hook run` still swallows generation errors (fail-open).
- Removing the `os.Stdout` swap is safe only after the leaf writers are converted; the
  implementation order below deliberately converts leaves first, then removes the swap.

## Test seams

- `cmd` package tests already construct `ReviewService`/`CommitService` literals with `MockVCS`,
  `MockClient`, and `testutils.MockConfigManager`. New tests set `stdout`/`stderr` to
  `bytes.Buffer` and assert stream contents; no network, no TTY.
- Exit-code tests live in `internal/errors` and need no Cobra.
- The `--raw` behavior is tested through `formatReviewComment` (renderer must not be called) and
  through `Execute` (stdout equals raw Markdown).
- `internal/ui` tests keep the existing `os.Pipe` capture style, switched to `os.Stderr`.
