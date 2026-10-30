# Add provider management commands

## Goal

Make the 27+ configured LLM providers discoverable and switchable from one command group.
`gmsg provider list` shows the active provider and every known provider with its effective
model and API-key source; `gmsg provider use <name>` switches the default provider (seeding a
working config block when needed); `gmsg provider test [name]` sends one minimal request and
reports connectivity, authentication, latency, and a short response. `gmsg newprovider` keeps
working as the interactive setup entry point.

## Background

- Provider selection today is only possible through `gmsg config set provider <name>`, which
  does not validate the name, does not create the provider's config block, and gives no hint
  about which key (if any) the provider will use.
- There are 28 registered providers plus user-defined custom providers. Only `openai` and
  `claude` blocks are seeded in a fresh config, so switching to any other provider without
  `newprovider` leaves commands failing with "Provider config not found".
- `internal/config.GetClientConfig` already resolves the API key as "config value → environment
  variable fallback" and special-cases `ollama` / `lmstudio` (no key needed), but the source is
  not observable from outside the package.
- `provider test` must work outside a Git repository and must not create a commit or touch a repo.
- Repository convention: commands are thin; config/env resolution belongs in `internal/config`;
  user-visible failures use `internal/errors` templates; new domain logic gets its own internal
  package with tests.

## In scope

- New `gmsg provider` command group registered in `main.go`, with `list`, `use`, and `test`
  subcommands.
- `gmsg provider list`: active provider detail block (model, API base, key source with masked
  key or env-var name) plus a table of all known providers (registered providers and custom
  providers that have a config block) showing effective/default model and key source.
- `gmsg provider use <name>`: validate the name, seed a default config block for a registered
  provider that has none (so subsequent commands can run), persist `provider`, and warn about a
  missing API key or missing provider-specific required fields.
- `gmsg provider test [name]`: resolve the saved configuration, send exactly one minimal
  non-streaming request (no retries), print model / base URL / key source / latency / response
  preview, exit non-zero with a typed error on failure.
- `internal/config.ResolveAPIKey` + `APIKeySource` type, used by `GetClientConfig` so runtime and
  listing agree on "config vs env vs not required vs missing"; `ManagerInterface` and
  `testutils.MockConfigManager` updated.
- New `internal/provider` package holding provider inventory, model/API-base fallback display
  logic, and switch/seed logic; it depends on `config.ManagerInterface`, `internal/llm`, and
  `internal/errors`, and never performs network I/O.
- New error templates: unknown provider, provider test failure (sanitized cause excerpt, latency,
  status-code-aware suggestions), no provider selected.
- Tests for config key resolution, provider inventory/switch, probe measurement with a fake
  prober, command construction/output, and error templates.
- Documentation: `README.md`, `README.zh.md`, `CLAUDE.md`; update Trellis spec docs where the
  package map and Manager API change.

## Out of scope

- Interactive selection or editing inside `gmsg provider` (that stays `gmsg newprovider`).
- Provider fallback chains, per-command `--provider` behavior changes, or provider health caching.
- Changing LLM request/response behavior or adding new providers.
- Changing existing `config set provider` semantics.
- JSON/scriptable output for provider list/test (a separate "output programmability" item).
- Version bump or release work; `FUTURE_FEATURES.md` stays untracked/unchanged.

## Requirements

- **R1**: `gmsg provider list` prints the active provider, its effective model, API base, and API
  key source (config with masked key, env var name, not required, or missing), then all known
  providers in sorted order with an active marker, their model and key source. Full API keys are
  never printed.
- **R2**: Known providers are the registered providers from `llm.GetProviders()` plus any
  custom providers that have a map section in the config; non-provider root keys
  (`provider`, `file_ignore`, `output`, `console`, `prompt`) are excluded.
- **R3**: `gmsg provider use <name>` persists `provider: <name>`. A registered provider without a
  config block gets a block seeded from its `GetRequiredConfig()` non-empty defaults before the
  switch, so `GetClientConfig` and later commands work.
- **R4**: `gmsg provider use` rejects a name that is neither registered nor configured with a
  validation error listing `gmsg provider list` / `gmsg newprovider` as next steps. A configured
  but unregistered (custom) provider is accepted with a warning that it uses the built-in
  OpenAI-compatible fallback.
- **R5**: `gmsg provider use` reports, without failing, when the provider's API key is missing
  (naming the config key and environment variable) and when provider-specific required fields
  (e.g. `azure.deployment_name`, `cloudflare.account_id`) still have no value.
- **R6**: `gmsg provider test [name]` uses the saved configuration for the named provider, or the
  current provider when omitted. It performs exactly one non-streaming request (no retries),
  reports latency and a short response preview on success, and exits non-zero on failure. Names
  are trimmed, and lowercased when the lowercase form is a known provider ("DeepSeek" works). An
  unknown name fails with the unknown-provider error; a registered provider without a config
  block fails with a "Provider Not Configured" error that suggests `gmsg provider use <name>`.
- **R7**: A failed provider test returns a `*GPTCometError` whose message includes the provider
  and elapsed latency, whose cause excerpt is sanitized (query-string secrets removed, length
  capped) and whose suggestions depend on the failure class (401/403 auth, 429 quota, 5xx
  provider outage, network/proxy, otherwise generic).
- **R8**: `gmsg provider test` must not require a Git/SVN repository and must not touch the
  repository.
- **R9**: `gmsg newprovider` keeps its current command name, flags, interactive behavior, and
  `GPTCOMET_TEST=1` non-interactive escape hatch.
- **R10**: All new commands honor the root `--config/-c` flag.
- **R11**: `ResolveAPIKey` keeps `GetClientConfig` behavior identical: config value wins over
  environment, `ollama`/`lmstudio` need no key, otherwise an empty result returns
  `APIKeyNotSetError`.
- **R12**: New user-visible failures use `internal/errors` templates with `%w` causes and
  actionable suggestions.

## Acceptance Criteria

- [x] `gmsg provider list` on a temp config shows the active provider's model/base/key source and
      a table containing registered providers plus a custom configured provider; a configured key
      appears masked, an env-only key shows the env var name, and `ollama` shows "not required".
- [x] `gmsg provider list` never prints a full API key (verified with a distinctive test key).
- [x] `gmsg provider use deepseek` on a config without a `deepseek` block creates the block with
      provider defaults, saves `provider: deepseek`, and later `GetClientConfig("deepseek")`
      succeeds for base/model (key still resolved from env or reported missing).
- [x] `gmsg provider use nonexistent-provider` fails with an unknown-provider error and leaves
      `provider` unchanged.
- [x] `gmsg provider use <custom configured provider>` succeeds with the OpenAI-compatible
      fallback warning.
- [x] `gmsg provider use` on a provider with no key prints a warning naming
      `gmsg config set <p>.api_key` and the env var.
- [x] `gmsg provider test` with a fake prober reports latency, response preview, and exits zero;
      with a fake failing prober it exits non-zero and the rendered error contains the provider
      name, latency, and class-specific suggestion.
- [x] `gmsg provider test` from a non-Git temp directory reaches the prober (no repo required).
- [x] `internal/config` tests cover the four `APIKeySource` cases and preserve `GetClientConfig`
      behavior; existing config tests stay green.
- [x] `gmsg newprovider` tests still pass under the unchanged command name.
- [x] `go test ./...`, `go test -race ./...`, `go vet ./...`, `staticcheck ./...`, and `gofmt`
      are clean.
- [x] `README.md`, `README.zh.md`, `CLAUDE.md`, and relevant `.trellis/spec/go` docs are updated;
      `FUTURE_FEATURES.md` remains unchanged and untracked.

## Decisions

- Keep `gmsg newprovider` as-is and add a separate `provider` group instead of renaming or hiding
  the existing command; changing an established command name is not needed for discoverability.
- `provider use` seeds a provider block from `GetRequiredConfig()` defaults. Without this, the
  switch would persist a provider that every later command rejects with "Provider config not
  found". Seeding is additive only: an existing block is never rewritten.
- API-key source detection moves into `internal/config` behind `ResolveAPIKey`; no new
  `os.Getenv` reads elsewhere, and the runtime path reuses exactly the same function.
- `provider test` reuses `client.Chat` with retries forced to one attempt, so no new HTTP layer is
  introduced; latency is measured around that single call. Provider status lines printed directly
  to stdout by the known debt are acceptable for this human-facing diagnostic command.
- The command group renames the Go constructor for the legacy command to
  `NewProviderSetupCmd`; the user-facing command name `newprovider` is unchanged.
- Custom (unregistered) providers stay usable through the existing OpenAI-compatible fallback;
  `provider use` only requires that they already have a config block.

## Risks & limitations

- Seeding defaults for providers with dynamic endpoints (`azure`, `vertex`, `bedrock`,
  `cloudflare`) cannot fully configure them; `provider use` reports the still-missing required
  fields instead of pretending the provider is ready.
- `provider test` consumes one real (small) API request; it uses the configured timeout and the
  configured `max_tokens`, so a provider with a very large max token setting may still cost more
  than a trivial ping. Documented.
- `provider test` output includes the provider's own `📤 Sending request…` / usage lines because
  those are printed inside `internal/llm` (known debt); the command's own output stays on stdout
  and remains human-readable.
- A custom provider whose model/API-base cannot be resolved is listed with "—" rather than
  guessed.
- The command group reads the config manager through the existing Cobra-context seam
  (`setConfigManager` / `getConfigManager`) and falls back to creating one from the root
  `--config/-c` flag, so root `--debug` keeps working and tests can inject a manager.
