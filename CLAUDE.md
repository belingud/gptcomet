# GPTComet - AI-Powered Git Commit Message Generator

## Project Overview

GPTComet generates Git commit messages and reviews code using LLMs. Supports 27+ providers (OpenAI, Claude, Gemini, etc.) and both Git/SVN.

**Core Commands**:

- `gmsg commit` - Generate commit message from diff
- `gmsg review` - AI code review (supports streaming with `--stream`)
- `gmsg generate` - Print a generated commit message to stdout (no commit)
- `gmsg hook` - Manage the Git `prepare-commit-msg` hook (install/uninstall/status)
- `gmsg config` - Manage configuration (get/set/list/reset/path/keys/append/remove/validate)
- `gmsg provider` - Inspect/switch/test/add providers (`list`, `new`, `use <name>`, `test [name]`)
- `gmsg newprovider` - Deprecated: use `gmsg provider new`; kept for now, removal scheduled for v2.14.0

## Architecture

```
Go Core                    Python Wrapper
────────                  ────────────────
cmd/           → CLI commands (commit, review, generate, hook, config, provider)
internal/
├── client/    → HTTP client with retry, proxy (HTTP/SOCKS5), streaming
├── config/    → YAML config manager (~/.config/gptcomet/gptcomet.yaml)
├── factory/   → Dependency injection (VCS, config, client creation)
├── git/       → VCS abstraction (git, svn)
├── hook/      → prepare-commit-msg install/run logic
├── llm/       → Provider interface + 27 implementations
├── provider/  → Provider inventory, switch/seeding rules, probe metadata
├── ui/        → Terminal UI (progress, markdown rendering)
├── errors/    → Structured error handling (typed errors with suggestions)
├── debug/     → Debug logging
└── testutils/ → Test mocks
pkg/
├── config/defaults/ → Default config + prompts
└── types/          → Shared types
py/            → Python wrapper (PyPI distribution shell)
```

## LLM Provider System

**Interface** ([internal/llm/llm.go](internal/llm/llm.go)):

```go
type LLM interface {
    Name() string
    BuildURL() string
    GetRequiredConfig() map[string]ConfigRequirement
    FormatMessages(message string) (interface{}, error)
    MakeRequest(ctx, client, message, stream) (string, error)
    GetUsage(data []byte) (string, error)
    BuildHeaders() map[string]string
    ParseResponse(response []byte) (string, error)
}
```

**Adding a provider**: Register in [internal/llm/provider.go](internal/llm/provider.go) init()

**Supported providers**: ai21, azure, bedrock, chatglm, claude, cloudflare, cohere, deepseek, doubao, gemini, groq, hunyuan, kimi, lmstudio, longcat, minimax, mistral, modelscope, ollama, openai, openrouter, qianfan, sambanova, silicon, tongyi, vertex, xai

## Development

### Tools

| Task   | Go                                   | Python                   |
| ------ | ------------------------------------ | ------------------------ |
| Format | `go fmt ./...` + `goimports -w .`    | `ruff format py/`        |
| Lint   | `go vet ./...` + `staticcheck ./...` | `ruff check py/`         |
| Test   | `go test ./...`                      | `pytest tests/py_tests/` |

**Command runner**: Use `just` (see `just --list` for all commands)

- `just install` - Install dependencies
- `just test` / `just test-py` - Run tests
- `just build` / `just build-py` - Build
- `just ci-check` - Run all checks

### Configuration

**Config file**: `~/.config/gptcomet/gptcomet.yaml`

**Default values**: [pkg/config/defaults/defaults.go](pkg/config/defaults/defaults.go)

- Default provider: `openai`
- Default model: `gpt-6.1-sol`
- Default API base: `https://api.openai.com/v1`
- Prompts: `brief_commit_message`, `rich_commit_message`, `translation`, `review`

**Supported keys**: Run `gmsg config keys`

### Key Patterns

**Provider selection**: [internal/client/client.go:New()](internal/client/client.go)

```go
switch config.Provider {
case "openai": provider = llm.NewOpenAILLM(config)
case "claude": provider = llm.NewClaudeLLM(config)
// ... 20+ cases
default: provider = llm.NewDefaultLLM(config)
}
```

**Error handling**: Structured `GPTCometError` with type, title, message, cause, suggestions. Six types: config, network, git, api, validation, unknown. Template functions in [internal/errors/templates.go](internal/errors/templates.go)

**Retry logic**: Exponential backoff with jitter (base 500ms, `config.Retries` attempts)

**Proxy support**: HTTP, HTTPS, SOCKS5 (with auth)

## Testing

**Go tests**: Table-driven tests, use mocks from `internal/testutils/`
**Python tests**: pytest, fixtures in `tests/py_tests/conftest.py`

## Documentation

- [README.md](README.md) - User guide, installation, configuration
- [docs/provider-models.md](docs/provider-models.md) - Default models per provider and refresh checklist
- [docs/deprecations.md](docs/deprecations.md) - Deprecation policy and removal schedule
- [pyproject.toml](pyproject.toml) - Python project config, tool settings

## Notes

- Python wrapper is only a distribution shell; all core logic is in Go
- Each provider has `*_test.go` with unit tests
- Error messages use templates from `internal/errors/templates.go`
- Streaming uses SSE (Server-Sent Events) parsing
- Complex tasks or plans should be handled using sequential thinking if available.

<!-- TRELLIS:START -->
# Trellis Instructions

These instructions are for AI assistants working in this project.

This project is managed by Trellis. The working knowledge you need lives under `.trellis/`:

- `.trellis/workflow.md` — development phases, when to create tasks, skill routing
- `.trellis/spec/` — package- and layer-scoped coding guidelines (read before writing code in a given layer)
- `.trellis/workspace/` — per-developer journals and session traces
- `.trellis/tasks/` — active and archived tasks (PRDs, research, jsonl context)

If a Trellis command is available on your platform (e.g. `/trellis:finish-work`, `/trellis:continue`), prefer it over manual steps. Not every platform exposes every command.

If you're using Codex or another agent-capable tool, additional project-scoped helpers may live in:
- `.agents/skills/` — reusable Trellis skills
- `.codex/agents/` — optional custom subagents

Managed by Trellis. Edits outside this block are preserved; edits inside may be overwritten by a future `trellis update`.

<!-- TRELLIS:END -->
