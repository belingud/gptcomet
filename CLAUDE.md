# GPTComet - AI-Powered Git Commit Message Generator

## Project Overview

GPTComet generates Git commit messages and reviews code using LLMs. Supports 23+ providers (OpenAI, Claude, Gemini, etc.) and both Git/SVN.

**Core Commands**:

- `gmsg commit` - Generate commit message from diff
- `gmsg review` - AI code review (supports streaming with `--stream`)
- `gmsg config` - Manage configuration (get/set/list/reset/path/keys/append/remove)
- `gmsg newprovider` - Interactive provider setup

## Architecture

```
Go Core                    Python Wrapper
────────                  ────────────────
cmd/           → CLI commands (commit, review, config, provider)
internal/
├── client/    → HTTP client with retry, proxy (HTTP/SOCKS5), streaming
├── config/    → YAML config manager (~/.config/gptcomet/gptcomet.yaml)
├── factory/   → Dependency injection (VCS, config, client creation)
├── git/       → VCS abstraction (git, svn)
├── llm/       → Provider interface + 23 implementations
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

**Supported providers**: ai21, azure, chatglm, claude, cohere, deepseek, gemini, groq, hunyuan, kimi, longcat, minimax, mistral, modelscope, ollama, openai, openrouter, sambanova, silicon, tongyi, vertex, xai, yi

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
- Default model: `gpt-4o`
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
- [pyproject.toml](pyproject.toml) - Python project config, tool settings

## Notes

- Python wrapper is only a distribution shell; all core logic is in Go
- Each provider has `*_test.go` with unit tests
- Error messages use templates from `internal/errors/templates.go`
- Streaming uses SSE (Server-Sent Events) parsing
- Complex tasks or plans should be handled using sequential thinking if available.
