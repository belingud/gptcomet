# Go Core — Development Guidelines

Scope: `main.go`, `cmd/`, `internal/`, `pkg/`. The Python package under `py/` is only a
launcher shell — see [../py/index.md](../py/index.md). Build/release wiring lives in
[../repo/index.md](../repo/index.md).

The binary is `gmsg` (alias `gptcomet`), built with Cobra. A run reads a Git/SVN diff,
renders a prompt template, calls one of 27 LLM providers through a shared HTTP client,
and prints or commits the result.

## Pre-Development Checklist

Read these before writing Go code:

1. [architecture.md](./architecture.md) — package boundaries, dependency direction, where new code belongs.
2. [provider-guidelines.md](./provider-guidelines.md) — **required** for any `internal/llm/*` change (new provider, URL/header/response shape, default model).
3. [config-and-defaults.md](./config-and-defaults.md) — **required** for any config key, prompt, env var or default-value change.
4. [error-handling.md](./error-handling.md) — **required** when an error can reach the user.
5. [cli-and-output.md](./cli-and-output.md) — **required** for commands, flags, progress bars or terminal output.
6. [quality-guidelines.md](./quality-guidelines.md) — tests, mocks, lint rules.
7. Shared thinking guides: [../guides/index.md](../guides/index.md).

Then state the change boundary before editing: what behavior changes, which files must
change, which providers/commands are affected, and what you are deliberately not touching.

## Quality Check

Run from the repo root. `just` is the only entry point you need (`just --list`):

| Command | What it runs |
| --- | --- |
| `just format` | `go fmt ./...` + `goimports -w .` + `ruff format py/` |
| `just check` | `go vet ./...` + `staticcheck ./...` |
| `just test` | `go test ./...` |
| `just test-coverage` | coverage profile + total percentage |
| `just ci-check` | `go vet` + `staticcheck` + `ruff check py/` |

CI also runs `go test -v -race ./...` (`.github/workflows/quality.yml`), so new tests must be
race-clean and must not depend on wall-clock timing or real network access.

Do not report a Go change as done while `just check` or `just test` fails.

## Package Map

| Path | Responsibility | Notes |
| --- | --- | --- |
| `main.go` | Root Cobra command, `--debug/-d`, `--config/-c`, `var version` | Registered in `.bumpversion.cfg` |
| `cmd/` | Command constructors `New<Name>Cmd()`, flag helpers, lipgloss styling, interactive editor | One `*_test.go` per file |
| `internal/factory/` | `ServiceDependencies` wiring: VCS + config manager + `types.ClientConfig` + API client | Only place that composes all layers |
| `internal/client/` | HTTP client, retry/backoff, proxy (HTTP/SOCKS5), SSE streaming, prompt rendering | Imports `internal/llm`, never the reverse |
| `internal/llm/` | `LLM` interface, `BaseLLM` defaults, 28 provider implementations, registry | Largest package (~64 files) |
| `internal/config/` | YAML manager, dot-key accessors, `GetClientConfig`, `GetSupportedKeys` | `ManagerInterface` is mocked in tests |
| `internal/git/` | `VCS` interface, Git + SVN implementations, gitignore-style `file_ignore` matching | No `os/exec` outside this package |
| `internal/hook/` | `prepare-commit-msg` install/status/uninstall/run logic | Generator is injected by `cmd`; no client/config imports |
| `internal/provider/` | Provider inventory, effective model/API-base display data, switch + default-block seeding | Imports `config.ManagerInterface`/`llm`/`errors`; no network I/O |
| `internal/secretscan/` | Pre-send secret scanning: diff added-line parser, vendored gitleaks rule data (`rules/`), rule matcher, warn/redact/block modes | Leaf package; imports only `internal/errors`; update rules with `just sync-secret-rules` |
| `internal/ui/` | Progress stages, provider selector (bubbletea) | Also markdown rendering via glamour in `cmd/review.go` |
| `internal/errors/` | `GPTCometError` type + one template function per user-facing scenario | Every user-visible failure has a template here |
| `internal/logger/`, `internal/debug/` | Leveled logger; opt-in debug printer to stderr | Leaf packages, no internal deps |
| `internal/prompt/` | `{{ var }}` template rendering | Used by `client` for prompts |
| `internal/constants/` | Timeouts, retry/backoff, endpoints, config keys | Changing a value here changes runtime behavior globally |
| `pkg/types/` | `ClientConfig`, `Message`, `CompletionResponse`, `APIKeySource` | Shared, importable outside `internal/` |
| `pkg/config/` | `ConfigRequirement` type + `defaults` package (default config, prompt defaults) | Defaults are the single source of truth |
