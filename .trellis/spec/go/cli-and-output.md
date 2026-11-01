# CLI & Output Guidelines

The CLI is Cobra-based. Commands are thin: parse flags → build dependencies through
`internal/factory` → call a service → print. Business logic lives in `internal/` (see
`cmd/review.go` `ReviewService` and `cmd/commit_action.go`).

## Command Shape

```go
// cmd/<name>.go
func New<Name>Cmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "<name>",
        Short: "<one line>",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. read flags, 2. factory.NewServiceDependenciesWithClient, 3. run, 4. print
            return nil
        },
    }
    generalFlags := pflag.NewFlagSet("general", pflag.ExitOnError)
    advancedFlags := pflag.NewFlagSet("advanced", pflag.ExitOnError)
    AddGeneralFlags(generalFlags, &options.RepoPath, &options.UseSVN)   // --repo/-r, --svn/-v
    AddAdvancedAPIFlags(advancedFlags, &options.CommonOptions)          // --model, --api-base, ...
    cmd.Flags().AddFlagSet(generalFlags)
    cmd.Flags().AddFlagSet(advancedFlags)
    return cmd
}
```

Rules:

- Register every new command in `main.go`: `rootCmd.AddCommand(cmd.New<Name>Cmd())`.
- Only the root command defines persistent flags: `--debug/-d` and `--config/-c`. API-override
  flags always come from `AddAdvancedAPIFlags` (`cmd/flags.go`). General flags are not fully
  shared today: `review` uses `AddGeneralFlags` (`--repo/-r`, `--svn/-v`) while `commit` declares
  its own (`--rich/-r`, `--yes/-y`, `--dry-run`, `--no-verify`, `--svn`). Follow the closest
  existing command — do not invent a third flag shape.
- `CommonOptions` (`cmd/flags.go:15`) is the shared API-override struct; add new overrides there
  and to `AddAdvancedAPIFlags`, not per-command.
- Commands that send a diff (`commit`, `generate`, `review`) register a local `--secrets
  off|warn|redact|block` flag through `cmd/sensitive.go`'s `resolveSensitiveOptions`. Warnings go
  to stderr (`printSensitiveFindings`), never stdout, so the `generate` contract is preserved.
- Flags are applied **after** `GetClientConfig`: services call
  `ApplyCommonOptions(&options.CommonOptions, clientConfig)` (`cmd/flags.go:63`), which only
  overwrites when the flag is non-empty. Booleans like `--stream/-s` (`cmd/review.go:368`) are the
  exception — `false` is meaningful, so they are not part of `CommonOptions`.
- `SilenceUsage` is set once on the root command in `main.go`; only `config validate` opts in
  again explicitly (`cmd/config.go:350`).
- `RunE` returns the error; Cobra prints `Error: ...` once to stderr. `main.go` exits with
  `errors.ExitCode(err)` (see [Output Contract](#output-contract)); a flag-parsing error becomes a
  `*errors.UsageError` via `rootCmd.SetFlagErrorFunc` and exits with code 2.
- Deprecated commands stay runnable and visible: `gmsg newprovider` prints a yellow `⚠` notice
  with its removal version and replacement from its own `RunE` (`cmd/newprovider.go`). Do not set
  Cobra's `Deprecated` field while the command must remain discoverable — Cobra hides deprecated
  commands from help. The removal target lives in `docs/deprecations.md` and in
  `providerSetupRemovalVersion`.
- Read output through `cmd.OutOrStdout()` in helpers that tests capture (see `cmd/newprovider.go`).
- `generate` has a strict stdout contract: only the final message. There is no `os.Stdout` swap
  anymore: leaf packages report status through `logger.Statusf` and `ui.Progress`, both of which
  write to stderr. `hook` commands delegate to `internal/hook`; only `internal/git` resolves
  repository/hooks paths.
- Service tests inject `stdout`/`stderr` fields (`orStdout`/`orStderr` in `cmd/writers.go`); the
  command wires them from `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` in `RunE`.

## Dependency Seams

- `setConfigManager(cmd, mgr)` / `getConfigManager(cmd)` (`cmd/config.go:33`) store the manager in
  the Cobra context. Tests inject `testutils.MockConfigManager` with this pair instead of building
  a real config file (`cmd/config_test.go` `setupTest`). `getConfigManager` skips commands with a
  nil context; `providerConfigManager` (`cmd/provider.go`) falls back to `config.New` from the
  root `--config` flag when no manager is in the context.
- The `provider test` command uses the `newProviderProber` package variable as its client seam so
  tests inject a fake prober and never hit the network (`cmd/provider_probe.go`).
- Services (`ReviewService`, commit action) take interfaces (`git.VCS`, `client.ClientInterface`,
  `config.ManagerInterface`, `MarkdownRenderer`), never concrete types, so tests can mock them.

## Output Contract

stdout is a machine interface; stderr is for humans and diagnostics. This is a hard rule for every
new command and feature.

| Command / mode | stdout | stderr |
| --- | --- | --- |
| `generate` | commit message only | status, request/usage lines, warnings, errors |
| `review` | rendered review (glamour + border) | provider banner, reminders, progress, warnings, errors |
| `review --raw` | model Markdown verbatim (no rendering, no border) | same as `review` |
| `review --stream` | streamed chunks + trailing newline | reminders, warnings, errors (`--raw` is a no-op; stream is already raw) |
| `commit` (interactive / `--yes`) | empty | banner, progress, previews, prompts, commit summary, errors |
| `commit --dry-run` | boxed "Generated commit message" preview | banner, progress, warnings, errors |

Exit codes (`internal/errors/exit.go`):

| Code | Constant | Meaning |
| --- | --- | --- |
| `0` | `ExitSuccess` | success |
| `1` | `ExitUnknown` | unexpected or unclassified failure |
| `2` | `ExitUsage` | command-line usage error (unknown flag / invalid flag value) |
| `3` | `ExitConfig` | configuration error |
| `4` | `ExitNetwork` | network/proxy/connection error |
| `5` | `ExitGit` | Git/VCS error, including "no staged changes" |
| `6` | `ExitAPI` | API/LLM error (rejection, parsing, retries exhausted) |
| `7` | `ExitValidation` | invalid input/content, including `sensitive.mode=block` |

`errors.ExitCode` classifies wrapped `*GPTCometError` values by `Type` (`errors.As`); plain errors
fall back to `ExitUnknown`. A validation wrapper that carries a classified cause (e.g.
`DependencyCreationError` around a config error) follows the cause's class, so scripts see the
root failure. Wrap with `%w` so the classification survives. Reuse the existing `internal/errors`
template functions instead of ad-hoc `fmt.Errorf` when a failure is user-facing.

## Terminal Output

Styling primitives live in `cmd/common.go`:

| Helper | Look |
| --- | --- |
| `formatBoxedMessage` | rounded green box — major success |
| `formatHighlightedMessage` | `====` bordered line — status banner |
| `formatRemindMessage` | yellow text — hints ("Reviewing, may take a few seconds…") |

- Colors use `lipgloss` with the ANSI palette (`lipgloss.Color("2")` green, `"3"` yellow), not
  hex values, so themes keep working.
- Emoji prefix the message kind, not the sentence: 🚀 starting work, 📤 outbound request,
  📝 writing a file, ⚠️ warning. Keep them at line start.
- Diagnostics go through `internal/logger` (`logger.Debug/Info/Warn/Error`); `--debug` switches
  the level in `main.go`'s `PersistentPreRun`. Do not print debug output with `fmt.Println`.
- User-facing status lines that must never reach stdout (request/usage lines, config warnings) go
  through `logger.Statusf`; it writes raw to stderr without level prefix, icon, or timestamp.
- Long, staged operations use `ui.Progress` (`internal/ui/progress.go`):
  `ui.NewProgress(verbose)`, `AddStages(...)`, `Start(name)`, `Complete(name)`, `Error(name, err)`.
  `progress.render()` prints the current stage to stderr; `StartWithNewLine` / `CompleteInNewLine`
  exist for output that must stay on its own line.
- Markdown review output is rendered with `glamour` behind the `MarkdownRenderer` interface
  (`cmd/review.go:34`), with the style from config `output.markdown_theme`. `review --raw` short
  circuits this path and keeps the model Markdown untouched for editor plugins/CI.
- Interactive prompts (provider picker, commit-message editor) use `bubbletea` models from
  `internal/ui` and `cmd/common.go` (`textEditor`). Interactive paths must honour the
  `GPTCOMET_TEST=1` escape hatch (`cmd/newprovider.go:26`) so tests never block on a TTY.

## What the User Sees (commit flow)

1. `commit` collects the staged diff (`git.VCS.GetStagedDiffFiltered`, applying `file_ignore`).
2. Renders the configured prompt (`prompt.brief_commit_message` / `rich_commit_message`).
3. Calls `client.GenerateCommitMessage` (retries + timeout inside `internal/client`).
4. Prints the message in the configured template (`output.rich_template`) and offers edit /
   confirm / abort.
5. Commits with `git commit [-n]` through `internal/git`.

Any new command should follow the same order: gather → render → call → display → confirm →
mutate the repo last.

## Output Rules

- No `fmt.Print*` / `os.Stdout` writes in `internal/` (non-test) code: put result output in `cmd/`
  services through their `stdout` writer and status output through `logger.Statusf` or
  `internal/ui.Progress` (stderr). This is enforced by review; `grep -rn "fmt.Print\|os.Stdout"
  --include='*.go' internal/` should only match `internal/logger` and tests.
