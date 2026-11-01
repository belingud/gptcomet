# Bootstrap Task: Fill Project Development Guidelines

**You (the AI) are running this task. The developer does not read this file.**

The developer just ran `trellis init` on this project for the first time.
`.trellis/` now exists with empty spec scaffolding, and this bootstrap task
exists under `.trellis/tasks/`. When they want to work on it, they should start
this task from a session that provides Trellis session identity.

**Your job**: help them populate `.trellis/spec/` with the team's real
coding conventions. Every future AI session — this project's
`trellis-implement` and `trellis-check` sub-agents — auto-loads spec files
listed in per-task jsonl manifests. Empty spec = sub-agents write generic
code. Real spec = sub-agents match the team's actual patterns.

Don't dump instructions. Open with a short greeting, figure out if the repo
has any existing convention docs (CLAUDE.md, .cursorrules, etc.), and drive
the rest conversationally.

---

## Status — DONE (2026-10-28)

- [x] Fill Go core guidelines (`go/`)
- [x] Fill Python wrapper guidelines (`py/`)
- [x] Fill repository build/release guidelines (`repo/`)
- [x] Add code examples taken from real source files
- [x] Remove non-applicable template scaffolding

The scaffold assumed a backend web app (`routes`, `ORM`, `database migrations`). This repository
is a Go CLI plus a Python distribution shell, so the layer set was **reshaped**, not filled:

- deleted `.trellis/spec/backend/{directory-structure,database-guidelines,error-handling,logging-guidelines,quality-guidelines,index}.md`
  — no routes, no ORM, no database exists in this project;
- created three layers that match the real tree: `go/`, `py/`, `repo/`;
- kept `guides/` unchanged (already applicable).

---

## Delivered spec set

### `go/` — Go CLI core (`main.go`, `cmd/`, `internal/`, `pkg/`)

| File | Documents |
|------|-----------|
| `go/index.md` | Scope, pre-development checklist, `just` quality gate, package map |
| `go/architecture.md` | Layer diagram, dependency rules (`pkg/` → no `internal/`, `llm` → no `client`), where-new-code-goes table, known debt |
| `go/provider-guidelines.md` | `LLM` interface + `BaseLLM` override matrix, provider anatomy, 7-step add-a-provider checklist, registration, anti-patterns |
| `go/config-and-defaults.md` | `~/.config/gptcomet/gptcomet.yaml`, `ManagerInterface`, `GetClientConfig` rules, supported key families, `defaults` package, add-a-key checklist |
| `go/error-handling.md` | `GPTCometError` shape/rendering, typed constructors vs `templates.go` scenarios, raise-per-layer table |
| `go/cli-and-output.md` | Cobra command shape, flag helpers, test seams, lipgloss styles, `ui.Progress`, `GPTCOMET_TEST` guard |
| `go/quality-guidelines.md` | `just` commands, table-driven test pattern, `internal/testutils` mocks/fakes, review checklist |

### `py/` — PyPI distribution shell

| File | Documents |
|------|-----------|
| `py/index.md` | Binary discovery table, `main()` dispatch contract (execvpe vs subprocess, `update` blocked), pdm packaging, rules |
| `py/quality-guidelines.md` | ruff config + rule families, pytest fixtures/mocks, skip guards |

### `repo/` — cross-cutting

| File | Documents |
|------|-----------|
| `repo/index.md` | Justfile-as-task-runner rule, Conventional Commits requirement, docs map (`README.zh` sync, `AGENTS.md` symlink) |
| `repo/build-and-release.md` | Every `just` recipe, `quality.yml` / `release.yml` job flow, wheel platform tags, three-file version bump, release checklist |


### Thinking guides (already populated)

`.trellis/spec/guides/` contains general thinking guides pre-filled with
best practices. Customize only if something clearly doesn't fit this project.

---

## How to fill the spec

### Step 1: Import from existing convention files first (preferred)

Search the repo for existing convention docs. If any exist, read them and
extract the relevant rules into the matching `.trellis/spec/` files —
usually much faster than documenting from scratch.

| File / Directory | Tool |
|------|------|
| `CLAUDE.md` / `CLAUDE.local.md` | Claude Code |
| `AGENTS.md` | Codex / Claude Code / agent-compatible tools |
| `.cursorrules` | Cursor |
| `.cursor/rules/*.mdc` | Cursor (rules directory) |
| `.windsurfrules` | Windsurf |
| `.clinerules` | Cline |
| `.roomodes` | Roo Code |
| `.github/copilot-instructions.md` | GitHub Copilot |
| `.vscode/settings.json` → `github.copilot.chat.codeGeneration.instructions` | VS Code Copilot |
| `CONVENTIONS.md` / `.aider.conf.yml` | aider |
| `CONTRIBUTING.md` | General project conventions |
| `.editorconfig` | Editor formatting rules |

### Step 2: Analyze the codebase for anything not covered by existing docs

Scan real code to discover patterns. Before writing each spec file:
- Find 2-3 real examples of each pattern in the codebase.
- Reference real file paths (not hypothetical ones).
- Document anti-patterns the team clearly avoids.

### Step 3: Document reality, not ideals

**Critical**: write what the code *actually does*, not what it should do.
Sub-agents match the spec, so aspirational patterns that don't exist in the
codebase will cause sub-agents to write code that looks out of place.

If the team has known tech debt, document the current state — improvement
is a separate conversation, not a bootstrap concern.

---

## Evidence & Verification (2026-10-28)

Every rule in the spec traces back to these sources:

- `CLAUDE.md` / `AGENTS.md` (existing conventions, architecture, commands)
- `Justfile`, `.github/workflows/{quality,release}.yml`, `.goreleaser.yaml`,
  `.bumpversion.cfg`, `cliff.toml`, `pyproject.toml`
- `internal/llm/{llm.go,provider.go,registry.go,builder.go,openai.go,claude.go}`
- `internal/client/client.go`, `internal/config/{config.go,accessor.go,client_config}`,
  `internal/git/{vcs.go,ignore.go}`
- `internal/errors/{errors.go,templates.go}`, `internal/ui/progress.go`,
  `internal/debug/debug.go`, `internal/logger/logger.go`
- `internal/testutils/*`, `cmd/{flags.go,config.go,review.go,provider.go,common.go}`, `main.go`
- `py/gptcomet/__init__.py`, `tests/py_tests/*`, `scripts/get_binary_path.py`

Verification performed:

- `find .trellis/spec -type f` → 14 files (7 go, 2 py, 2 repo, 3 guides); no `backend/` scaffolding.
- Placeholder scan (`To fill|TODO|Fill in each|Best practices for`) → no hits.
- Relative-link check across all spec files → no broken links.
- `python3 .trellis/scripts/get_context.py --mode packages` → `Spec layers: go, py, repo`.

---

## Quick explainer of the runtime (share when they ask "why do we need spec at all")

- Every AI coding task spawns two sub-agents: `trellis-implement` (writes
  code) and `trellis-check` (verifies quality).
- Each task has `implement.jsonl` / `check.jsonl` manifests listing which
  spec files to load.
- The platform hook auto-injects those spec files + the task's `prd.md`
  into every sub-agent prompt, so the sub-agent codes/reviews per team
  conventions without anyone pasting them manually.
- Source of truth: `.trellis/spec/`. That's why filling it well now pays
  off forever.

---

## Completion

When the developer confirms the checklist items above are done with real
examples (not placeholders), guide them to run:

```bash
python3 ./.trellis/scripts/task.py finish
python3 ./.trellis/scripts/task.py archive 00-bootstrap-guidelines
```

After archive, every new developer who joins this project will get a
`00-join-<slug>` onboarding task instead of this bootstrap task.

---

## Suggested opening line

"Welcome to Trellis! Your init just set me up to help you fill the project
spec — a one-time setup so every future AI session follows the team's
conventions instead of writing generic code. Before we start, do you have
any existing convention docs (CLAUDE.md, .cursorrules, CONTRIBUTING.md,
etc.) I can pull from, or should I scan the codebase from scratch?"
