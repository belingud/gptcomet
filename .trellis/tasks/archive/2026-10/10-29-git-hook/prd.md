# Add Git hook integration for automatic commit messages

## Goal

After a one-time `gmsg hook install`, a normal `git commit` in that repository automatically
generates a commit message from the staged diff and leaves it in the commit message editor. Hook
execution is fail-open: generation, configuration, or network errors leave the original message
file untouched and never block the commit.

## Background

- `gmsg commit` already generates a message from the staged diff, but only when invoked manually.
- Git executes `prepare-commit-msg` before opening the commit editor, which is the correct point to
  fill the message file automatically.
- `BaseLLM.MakeRequest` and a few providers print request/usage lines to stdout today. A command
  that captures stdout for scripting must not receive those lines.
- Repository convention: only `internal/git` shells out to Git; commands are thin and business
  logic belongs under `internal/`.

## In scope

- `gmsg hook install`, `gmsg hook uninstall`, and `gmsg hook status` for Git repositories.
- A managed, executable `prepare-commit-msg` hook script protected by a marker line.
- A hidden `gmsg hook run <file> [source] [sha]` command that applies skip rules, generates the
  message, and writes the message file.
- A public `gmsg generate` command that prints only the generated commit message to stdout.
- Skip rules: `GMSG_SKIP=1`; source `merge`, `squash`, `commit`, or `message`; and a message file
  that already contains non-comment content. The all-comment template Git supplies for the first
  commit is filled; a template with real content is left untouched.
- Escape hatches: `GMSG_FORCE=1` to override the source/content skips, `GMSG_BIN` to choose the
  binary, and `GMSG_CONFIG` / install-time `--config` to choose the configuration file.
- Preserve existing comment and verbose-diff blocks in the message file.
- Resolve the hooks directory through `git rev-parse`, respecting `core.hooksPath` and linked
  worktrees.
- Idempotent install, foreign-hook protection, `--force` backup, and uninstall restore.
- English and Chinese README documentation, plus the `CLAUDE.md` command list.

## Out of scope

- SVN hooks.
- Global hook installation or modifying `core.hooksPath`.
- Auto-staging, auto-committing, commit splitting, or post-generation commit-message validation.
- New providers or changes to LLM request/response behavior.
- Version bump or release.
- The separate `FUTURE_FEATURES.md` proposal document (created but not committed).

## Requirements

- **R1**: `gmsg hook install` resolves the repository hooks directory, writes
  `prepare-commit-msg`, sets mode `0755`, and is idempotent when the existing hook is
  gptcomet-managed.
- **R2**: If a foreign `prepare-commit-msg` exists, install fails with a Git-typed error and
  actionable suggestions. With `--force`, the original is backed up to
  `prepare-commit-msg.gptcomet-backup` before the managed hook replaces it. An existing backup
  aborts rather than overwriting it.
- **R3**: `gmsg hook uninstall` removes only a gptcomet-managed hook. If a backup exists, it
  restores the backup. Foreign hooks are left untouched.
- **R4**: `gmsg hook status` reports the resolved hooks path, `installed` / `foreign` / `missing`
  state, and whether a backup exists.
- **R5**: The installed shell script invokes `gmsg hook run ...`, suppresses output by default, and
  exits `0` for every failure mode.
- **R6**: `gmsg hook run` skips when `GMSG_SKIP=1`, when source is
  `merge` / `squash` / `commit` / `message`, or when the message file contains non-comment
  content. `GMSG_FORCE=1` overrides the source and content checks.
- **R7**: On success, hook run writes the generated message, preserves existing comment and
  verbose-diff blocks, and keeps the file mode.
- **R8**: On generation error, timeout, empty diff, empty message, or write failure, hook run
  leaves the message file unchanged and does not block the commit.
- **R9**: `gmsg generate` writes only the message to stdout (no provider/status/usage text) and
  exits non-zero on errors. It supports `--repo`, `--rich`, and the existing API override flags.
- **R10**: The hook uses the config path passed to install (resolved to an absolute path) unless
  `GMSG_CONFIG` overrides it. It works when `gmsg` is in `PATH` or `GMSG_BIN` is set.
- **R11**: `README.md`, `README.zh.md`, and `CLAUDE.md` document the commands, behavior, escape
  hatches, and limitations.

## Acceptance Criteria

- [ ] In a temporary Git repo, `gmsg hook install` creates an executable
      `prepare-commit-msg` containing the managed marker; rerunning is idempotent.
- [ ] With a foreign hook present, install fails without `--force` and the original bytes are
      unchanged; with `--force`, the original is backed up, the managed hook is installed, and
      uninstall restores the backup.
- [ ] `gmsg hook status` reports installed/missing/foreign and the resolved hooks path.
- [ ] Hook run with each skip condition leaves the message file byte-identical and does not call
      the generator (verified with an injected fake generator).
- [ ] Hook run with `GMSG_FORCE=1` generates and writes the message despite
      `source=message` or non-comment content.
- [ ] Hook run preserves comment lines and writes a trailing newline; generation failure leaves
      the file unchanged.
- [ ] The installed script passes the message file and its arguments to `gmsg hook run`,
      redirects output, and exits `0` when gmsg is missing or fails (tested with a fake `gmsg`).
- [ ] `gmsg generate` output contains only the message; status/usage prints go to stderr (tested
      through a captured writer/redirect helper).
- [ ] A repository with a custom `core.hooksPath` receives the hook in that path.
- [ ] `go test ./...`, `go test -race ./...`, `go vet ./...`, and `gofmt` are clean.
- [ ] `README.md`, `README.zh.md`, and `CLAUDE.md` are updated; `FUTURE_FEATURES.md` remains
      untracked.

## Decisions

- Skip `git commit -m` and other explicit-message sources by default; only `GMSG_FORCE=1`
  overrides.
- Foreign hook: fail with an error; `--force` backs it up and replaces; uninstall restores.
- Repo-local installation only; respect but never modify `core.hooksPath`.
- Use a thin POSIX shell wrapper plus Go `hook run` logic, keeping behavior testable.
- `gmsg generate` is public; `gmsg hook run` is hidden.
- Solve stdout purity by temporarily redirecting `os.Stdout` during generation rather than
  changing provider output in this task.

## Risks & limitations

- Generation happens synchronously before the editor opens; the existing request timeout bounds
  the wait.
- `git commit --no-verify` does not skip `prepare-commit-msg`; document `GMSG_SKIP=1`.
- Linked worktrees may share a hooks directory, so installing in one can affect others.
- Comment preservation assumes the default `#` comment character in v1.
- If gmsg is absent from `PATH` and `GMSG_BIN` is unset, the hook silently does nothing.
