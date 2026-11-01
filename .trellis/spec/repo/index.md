# Repository Conventions — Index

Cross-cutting rules for the repository as a whole: task runner, CI, versioning, release and
documentation. Language-specific rules live in [../go/index.md](../go/index.md) and
[../py/index.md](../py/index.md).

## Pre-Development Checklist

1. Read [build-and-release.md](./build-and-release.md) if you touch `Justfile`, `.github/`,
   `.goreleaser.yaml`, `.bumpversion.cfg`, `cliff.toml`, `pyproject.toml` packaging, or a version
   string.
2. Confirm which language layer you are changing and open that layer's index first.
3. Shared thinking guides: [../guides/index.md](../guides/index.md).

## Ground Rules

- **`Justfile` is the single task runner.** Every recurring command has a recipe; CI and
  contributors should not need memorized `go`/`uv` invocations. `just --list` is the source of
  truth.
- **Conventional Commits are mandatory.** `cliff.toml` groups the changelog by commit type
  (`feat`, `fix`, `chore`, `docs`, …) and `git-cliff` filters `docs:`/`test:`/`ci:` out of
  release notes. A non-conventional subject silently disappears from `CHANGELOG.md`.
- **Documentation ships with the change.** User-visible behavior updates `README.md`,
  `README.zh.md` (keep the pair in sync — the Chinese file is a full translation, not a stub),
  and `CLAUDE.md` when agent-facing instructions change. `AGENTS.md` is a symlink to
  `CLAUDE.md`; never write to it separately.
- **Generated artifacts stay out of git**: `dist/`, `bin/`, `coverage.out`, `coverage.xml`,
  `.coverage*`, `py/gptcomet/bin/gptcomet_*` (only `.gitkeep` is tracked).
- **Trellis files are committed** except the ones `.trellis/.gitignore` excludes
  (`.developer`, `.current-task`, `.runtime/`, `__pycache__`). Do not edit `.trellis/spec/guides/`
  unless a guide is factually wrong for this repository.
- **Task artifacts are the plan of record.** Non-trivial work follows the phases in
  `.trellis/workflow.md`: PRD (and `design.md`/`implement.md` for complex tasks) before code,
  spec update before commit.

## Documentation Map

| File | Audience | Update when |
| --- | --- | --- |
| `README.md` / `README.zh.md` | Users | commands, flags, config keys, providers, install |
| `docs/provider-models.md` | Users + maintainers | provider default models (has its own refresh checklist) |
| `CLAUDE.md` (↔ `AGENTS.md`) | AI agents | architecture, conventions, commands |
| `CONTRIBUTING.rst` | Contributors | dev setup, contribution flow |
| `ROADMAP.md`, `IMPLEMENTATION_PLAN.md`, `VSCODE_EXTENSION_PLAN.md` | Maintainers | planning docs, untracked in git |
| `.trellis/spec/**` | AI agents | coding standards (this directory) |
