# Go Architecture

Module: `github.com/belingud/gptcomet` · Go `1.25.6` (`go.mod`) · entry point `main.go`.

## Layering

```
main.go
  └── cmd/*                       Cobra commands, flags, terminal styling, prompts
        ├── internal/factory      composes VCS + config + ClientConfig + API client
        │     ├── internal/git    VCS interface (Git / SVN), file_ignore matching
        │     ├── internal/config Manager over ~/.config/gptcomet/gptcomet.yaml
        │     └── internal/client HTTP + retry + proxy + SSE streaming
        │           └── internal/llm  LLM interface, BaseLLM, provider registry
        └── internal/ui           progress stages, provider selector (bubbletea)
```

Leaf packages with no internal dependencies: `internal/errors`, `internal/logger`,
`internal/debug`, `internal/constants`, `internal/prompt`, `pkg/types`, `pkg/config`.

## Dependency Rules

These are enforced by the current code — keep them that way:

1. **`pkg/` never imports `internal/`.** `pkg/types` and `pkg/config` are the shared contract;
   they must stay importable from outside the module.
2. **`internal/llm` never imports `internal/client`.** The HTTP `*http.Client` is passed into
   `MakeRequest`; the dependency arrow points `client → llm`, not backwards.
3. **Providers do not import each other.** Shared behaviour belongs in `BaseLLM`
   (`internal/llm/llm.go`) or the builders in `internal/llm/builder.go`.
4. **Only `internal/git` shells out to VCS binaries** (plus test helpers in `internal/testutils`).
   `cmd/` and `internal/client/` never call `git`/`svn` directly.
5. **`internal/hook` never imports `internal/client` or `internal/config`.** `cmd` injects a
   `GenerateFunc`; hook rules remain testable with a stub and network-free.
5. **Only `internal/factory` composes cross-layer dependencies.** If a command needs VCS +
   config + client, it calls a factory function instead of building the pieces itself.

## Where New Code Goes

| Change | Location | Also update |
| --- | --- | --- |
| New command / subcommand | `cmd/<name>.go` with `New<Name>Cmd() *cobra.Command` | register in `main.go` `rootCmd.AddCommand(...)` |
| New Git hook lifecycle | `internal/hook/` | `cmd/hook.go`, `internal/git/hooks.go` (path resolution), README hook section |
| Provider inventory/switch rules | `internal/provider/` | `cmd/provider.go`, README provider section, error templates |
| New secret-scanning rule data | `internal/secretscan/rules/` | `scripts/gen_secret_rules.py`, `rules_gen.go`, `rules/UPSTREAM.md`, `just sync-secret-rules` |
| New flag | `cmd/flags.go` helpers (`AddAdvancedAPIFlags`, `AddGeneralFlags`, `AddConfigFlag`) | command doc string + README flag table |
| New LLM provider | `internal/llm/<provider>.go` | `internal/llm/provider.go` `init()`, `<provider>_test.go`, `docs/provider-models.md` |
| New config key | `pkg/config/defaults/defaults.go` (default) | `internal/config/accessor.go` `GetSupportedKeys`, `internal/config/client_config` mapping, README config section |
| New user-facing error | `internal/errors/templates.go` (scenario) / `errors.go` (type helper) | tests asserting the message |
| New HTTP behaviour (retry, timeout, proxy, streaming) | `internal/client/client.go` | `internal/constants/constants.go` for tunables |
| New VCS | `internal/git/<vcs>.go` implementing `VCS` | `internal/git/vcs.go` `NewVCS` switch, `cmd/flags.go` |
| New interactive UI | `internal/ui/` (bubbletea model) | keep `cmd/` thin |
| New shared type | `pkg/types/types.go` | — |

## Recurring Patterns

- **Interface + mock triad.** `llm.LLM`, `git.VCS` and `config.ManagerInterface` each have a
  testify mock in `internal/testutils/` (`MockLLM`, `MockVCS`, `MockConfigManager`). Interfaces
  exist so commands and the client can be tested without network or a real repo — keep them small.
- **Embedding for defaults.** Providers embed `*BaseLLM` and override only the methods that
  differ. See [provider-guidelines.md](./provider-guidelines.md).
- **Constructor injects defaults.** `New<X>LLM(config *types.ClientConfig)` mutates the config
  (API base, model, completion path, answer path) and then builds the struct — callers never
  pre-fill defaults.
- **Registry via `init()`.** `internal/llm/provider.go` maps provider name → constructor;
  `llm.CreateProvider` falls back to `openai` when the name is unknown.
- **Dependencies travel in the Cobra context.** `cmd/config.go` stores the config manager with
  `setConfigManager(cmd, mgr)` and reads it with `getConfigManager(cmd)`; tests use the same
  seam (`cmd/config_test.go` `setupTest`).
- **Structured errors.** Anything that reaches the terminal is a `*errors.GPTCometError` built
  from a template, rendered with suggestions. See [error-handling.md](./error-handling.md).

## Known Debt (do not spread)

- `internal/llm/llm.go` (`BaseLLM.MakeRequest`) and `internal/client/client.go` print directly
  with `fmt.Printf` (e.g. `📤 Sending request to %s...`, token usage). New code must not add
  prints inside `internal/`: use `internal/logger` for diagnostics, `internal/ui` for progress,
  or print from `cmd/`.
- `llm.NewProvider` returns an empty `&DefaultLLM{}` when a provider is unregistered, while
  `llm.CreateProvider` falls back to `openai`. Production paths go through `client.New` →
  `CreateProvider`; do not add new callers of `NewProvider`.
- `internal/config` spreads `Manager` methods over `config.go`, `accessor.go` and
  `client_config.go`; add a new accessor next to the most similar existing one instead of
  creating a new file.
- Provider `api_base` defaults are duplicated between the provider constructor
  (`BuildStandardConfig*`) and `Manager.GetClientConfig` (special cases for ollama, lmstudio,
  bedrock, cloudflare, qianfan, doubao, gemini, vertex). When touching one, check the other.
