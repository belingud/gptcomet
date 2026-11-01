# Fix git error reporting outside a repository

## Goal

Running `gmsg commit` / `gmsg generate` / `gmsg review` (and hook commands) outside a
git repository currently reports `Git Command Failed` with suggestions that ask the
user to check whether git is installed. The real cause — the working directory is not
a git repository — is never surfaced:

- Git 2.55 makes `git diff --staged` fall back to `--no-index` mode outside a repo,
  so it fails with `unknown option 'staged'` (exit 129) instead of an explicit
  "not a git repository" error.
- `GPTCometError.Error()` never renders `Cause`, so the captured git stderr is hidden.
- `Git command '... (exit code 129)' failed.` puts the exit code inside the quotes.

Exit code (5) and the stdout/stderr contract are already correct and must not change.

## Requirements

1. When a git command fails because `repoPath` is not inside a git working tree,
   return the existing `Not a Git Repository` error (`GitRepositoryNotFoundError`):
   type git → exit 5, suggestions `git init` / navigate to a repository.
   Detection must not parse localized stderr: probe with
   `git rev-parse --is-inside-work-tree` and use exit status/output.
   Applies to every GitVCS command (diff, staged checks, branch/log/commit) and to
   hook helpers that execute git through `runCommand`.
2. Generic git failures inside a repository (bad revision, commit rejected by a hook,
   …) keep the `Git Command Failed` error, but the rendered error must include a
   capped, flattened excerpt of git's stderr and exit status, so "check the git
   command output" points at something the user can actually see.
3. Command label reads `git diff --staged -U2` (args joined with spaces); the exit
   code is not placed inside the quoted command.
4. Do not change the stdout/stderr contract or exit-code mapping. `--svn` paths keep
   their current behavior (no git probe, no git-specific classification).
5. Tests: unit tests in `internal/git` (non-repo → Not a Git Repository for exported
   GitVCS methods; in-repo failing command → Git Command Failed with stderr excerpt),
   an error-rendering test for the details excerpt, and E2E assertions against the
   built binary (non-repo: exit 5, empty stdout, stderr has `Not a Git Repository` +
   `git init`, no "Check if git is installed"; in-repo success unchanged).

## Constraints

- Probe only on the failure path; successful commands must not spawn an extra process.
- User machine runs git with zh_CN locale: never match on git's message text.
- Reuse the existing 280-rune `sanitizeErrorDetail` cap/redaction for stderr excerpts.
- Must also work on older Git, where non-repo `git diff` exits 128 instead of 129.

## Acceptance Criteria

- [x] `gmsg commit`, `gmsg generate`, `gmsg review` in a non-git dir: exit 5, empty
      stdout, stderr contains `Not a Git Repository` and `git init`, and does not
      contain `Check if git is installed`.
- [x] `gmsg hook status` outside a repo reports the same `Not a Git Repository` error.
- [x] In-repo failing git command: error shows `git <args>` joined with spaces plus a
      stderr excerpt in the rendered error; exit 5.
- [x] Successful in-repo commands behave exactly as before (no output changes).
- [x] `go test ./...`, `go vet ./...`, `staticcheck ./...`, gofmt/goimports clean;
      E2E script passes.

## Notes

- Related: v2.11.0 output-contract work.
- `GitRepositoryNotFoundError` already existed but only hook helpers used it, and only
  when `rev-parse` produced empty output (it does not on failure).
- Repro: non-git dir + `gmsg commit` →
  `Git command 'git diff --staged --name-only (exit code 129)' failed.`
