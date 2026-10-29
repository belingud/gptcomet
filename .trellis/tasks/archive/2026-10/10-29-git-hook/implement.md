# Implementation — Git hook integration

## Preconditions

- PRD and design reviewed and approved; `task.py start` has not run yet.
- Before editing code, load `trellis-before-dev` and re-read the Go guidelines that apply:
  `cli-and-output.md`, `error-handling.md`, `config-and-defaults.md` (for docs only; no new keys),
  `quality-guidelines.md`.
- Work on a feature branch or `master` as the developer prefers; do not mix unrelated `.claude/`,
  `.agents/`, or `.dsh/` files into the feature commits.

## Ordered checklist

### 0. Proposal document (separate from the hook feature)

- [ ] Create `FUTURE_FEATURES.md` at the repo root with the P0/P1/P2 feature proposals already
      discussed (hook, repo config, style validation, review ranges/JSON, provider UX, large diff,
      PR/changelog, i18n, secret scan, refine, cache/cost, local API, GitHub Action, split, style
      learning, fallback, not-recommended items).
- [ ] Do **not** `git add` this file. It must remain untracked.

### 1. Git path helpers

- [ ] Add `internal/git/hooks.go`:
  - `GetRepoRoot(repoPath string) (string, error)` via `git rev-parse --show-toplevel`.
  - `GetHooksDir(repoPath string) (string, error)` via
    `git rev-parse --path-format=absolute --git-path hooks`, with a fallback for older Git.
  - Wrap failures with `GitRepositoryNotFoundError` / `GitCommandFailedError`.
- [ ] Add `internal/git/hooks_test.go` with `testutils.TestGitRepo`:
  - normal repo resolves `.git/hooks`;
  - `git config core.hooksPath custom-hooks` is honored;
  - non-repo returns an error.
- [ ] Keep all `exec.Command("git", ...)` calls inside `internal/git`.

### 2. Hook error templates

- [ ] Add titles/messages/suggestions to `internal/errors/constants.go`.
- [ ] Add `HookAlreadyExistsError`, `HookBackupExistsError`, `HookInstallError`,
      `HookUninstallError` to `internal/errors/templates.go` using `%w` for causes.
- [ ] Extend `internal/errors/errors_test.go` with substring/suggestion assertions.

### 3. Hook domain — install / status / uninstall

- [ ] Add `internal/hook/manager.go` with `Manager`, `InstallOptions`, `Status`,
      `UninstallResult` from the design.
- [ ] Add `internal/hook/script.go` with the managed marker, script template, `shellQuote`, and
      `isManagedHook`.
- [ ] Implement:
  - `Install`: resolve hooks dir, mkdir, classify existing file, write executable script;
  - `Uninstall`: remove only managed files, restore backup;
  - `Status`: hooks dir/path/state/backup.
- [ ] Add `internal/hook/manager_test.go` for create/idempotent/foreign/`--force`/restore/status.

### 4. Hook run rules

- [ ] Add `internal/hook/run.go`:
  - `RunInput`, skip matrix, `GMSG_SKIP` / `GMSG_FORCE` truthiness;
  - source skip for merge/squash/commit/message; `template` is handled by the non-comment content check;
  - non-comment content detection and comment preservation;
  - injected `GenerateFunc`, repo-root resolution, file-mode preservation;
  - generation errors returned for the command layer to swallow.
- [ ] Add `internal/hook/run_test.go` with a stub generator:
  - every skip condition leaves bytes unchanged and does not call the generator;
  - `GMSG_FORCE=1` overrides;
  - message + comments composition;
  - empty message, generator error, missing file;
  - install script contains marker/config and calls `hook run`.
- [ ] Run `go test ./internal/hook ./internal/git` before moving on.

### 5. Shared generation

- [ ] In `cmd/commit_action.go`, extract `getStagedDiff()` and add `GenerateMessage()`.
- [ ] Point `Execute` and `handleCommitInteraction` retry at `getStagedDiff()`; keep progress and
      error text unchanged.
- [ ] Add `cmd/generate.go`:
  - `NewGenerateCmd()` with `--repo`, `--rich`, advanced API flags (Git only);
  - generation while `os.Stdout` is redirected to `os.Stderr`, restored with `defer`;
  - output only the message through `cmd.OutOrStdout()`.
- [ ] Add `cmd/generate_test.go` for flag construction and the stdout-redirect helper; avoid real
      API calls.
- [ ] Add `cmd/hook.go`:
  - `NewHookCmd()` parent plus `install`, `uninstall`, `status`, hidden `run`;
  - `--repo` and `--force` flags; root `--config` propagation;
  - `hook run` logs errors at debug and returns nil (fail-open);
  - generator closure builds `CommitService` and calls `GenerateMessage`.
- [ ] Register `NewHookCmd()` and `NewGenerateCmd()` in `main.go`; update the main command list.

### 6. Documentation

- [ ] `README.md`: add `gmsg hook` and `gmsg generate` to the command list and a "Git hook"
      section covering install/status/uninstall, skip behavior, `GMSG_SKIP`, `GMSG_FORCE`,
      `GMSG_BIN`, `GMSG_CONFIG`, `--no-verify` caveat, worktree/`core.hooksPath` behavior, and
      synchronous generation latency.
- [ ] `README.zh.md`: full translation, not a stub.
- [ ] `CLAUDE.md`: add the two commands to "Core Commands" and mention the hook flow.
- [ ] Update `.trellis/spec/go/architecture.md` package map with `internal/hook` and
      `.trellis/spec/go/cli-and-output.md` with the hook command shape if the implementation
      introduces a reusable pattern.

### 7. Quality gate

- [ ] `gofmt -l cmd internal pkg main.go` prints nothing.
- [ ] `go test ./...` green.
- [ ] `go test -race ./...` green.
- [ ] `go vet ./...` green.
- [ ] `staticcheck ./...` green (if available; CI pins 2026.2.2).
- [ ] `go build ./...` green.
- [ ] Manual temporary-repo check:
  - `gmsg hook install`, `gmsg hook status`, `gmsg hook uninstall`;
  - foreign hook + `--force` backup/restore;
  - installed script with a fake `gmsg` in `PATH` for args and fail-open.
- [ ] Verify `FUTURE_FEATURES.md` is still untracked.

### 8. Commit / finish

- [ ] Commit 1: `feat(generate): add message-only commit generation` (service extraction + command
      + tests).
- [ ] Commit 2: `feat(hook): install and manage prepare-commit-msg hook` (git helpers, hook
      package, commands, error templates, README/CLAUDE docs, tests).
- [ ] Keep `FUTURE_FEATURES.md` out of both commits.
- [ ] Add the Trellis task artifacts (`.trellis/tasks/10-29-git-hook/`) in a `chore(trellis)` commit
      only if the developer agrees; do not add unrelated `.claude/`, `.agents/`, `.dsh/`, or
      `.trellis` scaffolding.
- [ ] Run `trellis-check`, update the journal, and archive the task with
      `python3 ./.trellis/scripts/task.py archive`.

## Risk / rollback points

| Risk | Mitigation / rollback |
| --- | --- |
| `CommitService` refactor breaks commit flow | Keep `getStagedDiff` error text identical; run existing `cmd` commit tests; revert the service commit independently |
| `os.Stdout` mutation hides output or races | Restore with `defer`; never run generation concurrently; test helper directly |
| Hook script quoting breaks paths with spaces | `shellQuote` + table test with spaces and single quotes |
| Foreign hook overwritten | Default error; backup required; uninstall restore test |
| Hook blocks commits | Every script path exits 0; hook run errors are swallowed; existing request timeout |
| Worktree/shared hooks surprise | Document; tests cover `core.hooksPath`; no global writes |
| Trellis files muddy the feature diff | Commit task artifacts separately or leave untracked; never `git add .trellis` wholesale |

## Definition of done

- All acceptance criteria in `prd.md` are checked.
- `just format`, `just test`, `just check`/`go vet`, and race tests pass.
- README/README.zh/CLAUDE.md updated and consistent.
- Hook install/uninstall/status and the fail-open path verified in a temporary repo.
- Task artifacts updated; `FUTURE_FEATURES.md` remains uncommitted.
