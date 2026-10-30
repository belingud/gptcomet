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
- Flags are applied **after** `GetClientConfig`: services call
  `ApplyCommonOptions(&options.CommonOptions, clientConfig)` (`cmd/flags.go:63`), which only
  overwrites when the flag is non-empty. Booleans like `--stream/-s` (`cmd/review.go:368`) are the
  exception — `false` is meaningful, so they are not part of `CommonOptions`.
- `SilenceUsage` is set once on the root command in `main.go`; only `config validate` opts in
  again explicitly (`cmd/config.go:350`).
- `RunE` returns the error; Cobra prints `Error: ...`. `main.go` exits with code 1 on failure.
- Deprecated commands stay runnable and visible: `gmsg newprovider` prints a yellow `⚠` notice
  pointing at `gmsg provider new` from its own `RunE` (`cmd/newprovider.go`). Do not set Cobra's
  `Deprecated` field while the command must remain discoverable — Cobra hides deprecated commands
  from help.
- Read output through `cmd.OutOrStdout()` in helpers that tests capture (see `cmd/newprovider.go`).
- `generate` has a strict stdout contract: only the final message. `generateWithCleanStdout`
  temporarily redirects `os.Stdout` to stderr because some providers still print status/usage
  lines there (known debt). `hook` commands delegate to `internal/hook`; only `internal/git`
  resolves repository/hooks paths.

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
- Long, staged operations use `ui.Progress` (`internal/ui/progress.go`):
  `ui.NewProgress(verbose)`, `AddStages(...)`, `Start(name)`, `Complete(name)`, `Error(name, err)`.
  `progress.render()` prints the current stage; `StartWithNewLine` / `CompleteInNewLine` exist for
  output that must stay on its own line.
- Markdown review output is rendered with `glamour` behind the `MarkdownRenderer` interface
  (`cmd/review.go:34`), with the style from config `output.markdown_theme`.
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

## Known Debt

`internal/llm/llm.go` and `internal/client/client.go` still print progress lines directly
(`📤 Sending request to %s...`, token usage). Do not extend this pattern — put new output in
`cmd/`, `internal/ui`, or `internal/logger`.
