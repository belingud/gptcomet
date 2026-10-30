# Design — Provider management commands

## Boundaries

| Layer | Files | Responsibility |
| --- | --- | --- |
| CLI | `cmd/provider.go`, `cmd/provider_probe.go`, `cmd/newprovider.go` | Cobra shape, flags, human output; no inventory/config logic |
| Provider domain | `internal/provider/` (new) | known-provider inventory, effective model/API-base/key-source display data, switch + seeding rules |
| Config | `internal/config/config.go`, `accessor.go` | `ResolveAPIKey` + `APIKeySource`; `GetClientConfig` reuses it |
| Errors | `internal/errors/constants.go`, `templates.go` | unknown-provider, no-provider-selected, provider-test-failure templates |
| Wiring | `main.go` | register `provider` group; keep `newprovider` |
| Tests | next to each file | table-driven, temp config, fake prober; no network |

`internal/provider` imports `internal/config` (interface only), `internal/llm`, `internal/errors`,
and `pkg/types`. It performs **no** HTTP and **no** config writes except through
`config.ManagerInterface` (`Use`). `cmd` composes the real API client via
`factory.NewAPIClient`; the provider package never imports `internal/client` or `cmd`.

## Config API (`internal/config`)

```go
// APIKeySource identifies where a provider's API key resolves from.
type APIKeySource string

const (
    APIKeySourceConfig      APIKeySource = "config"
    APIKeySourceEnv         APIKeySource = "env"
    APIKeySourceNotRequired APIKeySource = "not-required"
    APIKeySourceMissing     APIKeySource = "missing"
)

// ResolveAPIKey returns the API key that would be used for provider and where
// it came from. It never errors.
func (m *Manager) ResolveAPIKey(provider string) (key string, source APIKeySource)
```

- `providerAllowsEmptyAPIKey(provider)` (`ollama`, `lmstudio`) → `("", NotRequired)`.
- Non-empty `<provider>.api_key` → `(key, Config)`.
- Else non-empty `os.Getenv(errors.GetAPIKeyEnvVar(provider))` → `(key, Env)`.
- Else `("", Missing)`.
- `GetClientConfig` replaces its inline key block with:

```go
apiKey, source := m.ResolveAPIKey(provider)
if source == APIKeySourceMissing {
    return nil, gptcometerrors.APIKeyNotSetError(provider)
}
```

Behavior is identical to today (config wins, env fallback, keyless providers allowed). Add the
method to `ManagerInterface` and `testutils.MockConfigManager` (return the two values through
`m.Called(provider)`).

## Provider domain (`internal/provider`)

```go
type Info struct {
    Name              string
    Active            bool
    Registered        bool                    // in llm registry
    Configured        bool                    // has a map section in the config
    Model             string
    ModelFromConfig   bool
    APIBase           string
    APIBaseFromConfig bool
    KeySource         config.APIKeySource
    KeyMasked         string                  // masked key, only for KeySourceConfig
    KeyEnvVar         string                  // errors.GetAPIKeyEnvVar(name)
}

type Service struct{ cfg config.ManagerInterface }

func NewService(cfg config.ManagerInterface) *Service

func (s *Service) CurrentName() string                 // "" when unset
func (s *Service) NormalizeName(name string) string    // trim; lowercase when that matches
func (s *Service) List() []Info                        // sorted by name
func (s *Service) Get(name string) Info
func (s *Service) Use(name string) (*SwitchResult, error)

type SwitchResult struct {
    Provider      string
    AlreadyActive bool
    Registered    bool
    Seeded        bool
    KeySource     config.APIKeySource
    KeyEnvVar     string
    MissingFields []string // provider-specific required config still empty
}
```

### Known-provider inventory

1. Start with the set of `llm.GetProviders()`.
2. Add every key of `cfg.ListWithoutPrompt()` whose value is `map[string]interface{}` and which
   is not a non-provider root section: `provider`, `file_ignore`, `output`, `console`, `prompt`.
3. Sort alphabetically.

### Effective model / API base (display only)

`Get(name)` takes the config section (if any) and:

- model: non-empty `<name>.model` → value, `ModelFromConfig = true`; else, for registered
  providers, `GetRequiredConfig()["model"].DefaultValue` (this is the same source
  `gmsg newprovider` pre-fills); else `""` (rendered as `—`).
- API base: non-empty `<name>.api_base` → value, `APIBaseFromConfig = true`; else
  `GetRequiredConfig()["api_base"].DefaultValue` when non-empty (dynamic-endpoint providers such
  as `azure`/`bedrock`/`cloudflare` leave it empty); else `""`.
- key: `cfg.ResolveAPIKey(name)`; mask with `config.MaskAPIKey(key, 4)` when the source is config.

These values are explanatory, not authoritative. `provider test` uses
`Manager.GetClientConfig(name)`, which remains the single runtime resolution path.

### `Use` rules

1. `NormalizeName`: trim spaces, then lowercase when the lowercased name is a known
   registered/configured provider (built-in names are lowercase; custom names keep their spelling).
2. Neither registered nor configured → `errors.ProviderNotFoundError(name)`; nothing is written.
3. If registered but unconfigured → seed a map from `GetRequiredConfig()`:
   every requirement whose `DefaultValue != ""` becomes `{field: DefaultValue}`, plus
   `api_key: ""` so the config documents the field. Write it with
   `cfg.SetNestedValue([]string{name}, section)`, set `Seeded = true`. An existing section is
   never modified.
4. Set the default provider with `cfg.Set("provider", name)` — `Set` saves the whole config in one
   write, including the seeded block.
5. Compute `KeySource`/`KeyEnvVar` via `ResolveAPIKey`.
6. Compute `MissingFields`: for registered providers, every `GetRequiredConfig()` key other than
   `api_key` whose `DefaultValue == ""` and whose value in the (possibly seeded) section is
   empty/missing. Sort for stable output.
7. Unregistered-but-configured providers are returned with `Registered = false`; the command layer
   prints the OpenAI-compatible-fallback warning.

## Probe (`cmd/provider_probe.go`)

```go
const providerProbePrompt = "Reply with the single word: pong"

type providerProber interface {
    Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error)
}

// package-level seam; tests replace it
var newProviderProber = func(cfg *types.ClientConfig) (providerProber, error) {
    return factory.NewAPIClient(cfg)
}

type providerTestResult struct {
    Info         provider.Info
    ClientConfig *types.ClientConfig
    Latency      time.Duration
    Response     string
}

func runProviderTest(ctx context.Context, cfgManager config.ManagerInterface,
    svc *provider.Service, name string) (*providerTestResult, error)
```

- The provider name is normalized with `svc.NormalizeName` first.
- `svc.Get(name)` must report a configured provider: an unknown name returns
  `ProviderNotFoundError`, a registered-but-unconfigured provider returns
  `ProviderNotConfiguredError` (suggesting `gmsg provider use <name>`).
- `cfgManager.GetClientConfig(name)` then resolves the saved config; its typed errors (missing key)
  are returned unchanged.
- `Retries = 1` on the local copy, so `client.Chat` performs one attempt instead of the configured
  retry loop.
- Time only the `Chat` call. On failure return the partial result (latency) **and**
  `errors.ProviderTestError(name, latency, err)`.
- Success response is trimmed, newlines collapsed, and truncated for display by the command.
- No VCS is created or touched, so the command works outside a Git repo.

## CLI wiring (`cmd/provider.go`)

- `git mv cmd/provider.go cmd/newprovider.go`; rename `NewProviderCmd` → `NewProviderSetupCmd`
  (command name `newprovider` unchanged). `cmd/provider_test.go` becomes
  `cmd/newprovider_test.go`.
- New `NewProviderCmd()` parent, `Use: "provider"`, children `list`, `use`, `test`:
  - `list`: no args, no flags.
  - `use <name>`: `cobra.ExactArgs(1)`.
  - `test [name]`: `cobra.RangeArgs(0, 1)`.
- Shared helpers:
  - `providerConfigManager(cmd)`: return `getConfigManager(cmd)` when the Cobra context already
    has one (test seam), else `config.New(<root --config value>)`.
  - `formatKeySource(Info) string`: `config (sk-ab***)` / `env OPENAI_API_KEY` /
    `not required` / `missing`.
  - `formatModel(info)`, `formatAPIBase(info)`: value + ` (config)` / ` (default)` / `—`.
- All command output goes through `cmd.OutOrStdout()`; warnings use `formatRemindMessage`, the
  switch confirmation uses `formatBoxedMessage`; colors stay on the ANSI palette.
- `list` format (tabwriter, two spaces between columns):

```
Current provider: claude
  model:    claude-sonnet-5-5 (config)
  api base: https://api.anthropic.com/v1 (default)
  api key:  env ANTHROPIC_API_KEY

PROVIDER    MODEL                       API KEY
  ai21      jamba-large-1.7            missing
* claude    claude-sonnet-5-5          env ANTHROPIC_API_KEY
  ollama    llama2                     not required
  ...

Switch providers with 'gmsg provider use <name>'.
Set a missing API key with 'gmsg config set <provider>.api_key <key>' or the provider's env var.
Configure a provider interactively with 'gmsg newprovider'.
```

- `use` success prints a boxed `Active provider: <name>`, plus the warnings from `SwitchResult`
  (missing key, missing required fields, custom fallback) as yellow reminder lines.
- `test` prints the resolved model/base/key source first, then on success:
  `✔ Connected in 842ms` and `response: "pong"` (truncated). Failure returns the typed error and
  lets Cobra print it.
- `main.go`: register `NewProviderCmd()` with comment `// provider`, keep
  `NewProviderSetupCmd()` with comment `// newprovider`.

## Error templates (`internal/errors`)

Add constants:

- `ErrTitleUnknownProvider`, `ErrMsgUnknownProvider`
- `ErrTitleProviderNotConfig`, `ErrMsgProviderNotConfig`
- `ErrTitleNoProviderSelected`, `ErrMsgNoProviderSelected`
- `ErrTitleProviderTest`, `ErrMsgProviderTestFailed`
- `SuggListProviders`, `SuggConfigureProvider`, `SuggTestProvider`, `SuggCheckProviderBaseURL`,
  `SuggUseProvider`

Templates:

```go
func ProviderNotFoundError(name string) *GPTCometError
func ProviderNotConfiguredError(name string) *GPTCometError
func NoProviderSelectedError() *GPTCometError
func ProviderTestError(provider string, latency time.Duration, cause error) *GPTCometError
```

`ProviderTestError`:

- message: `Provider '<p>' did not respond successfully after <latency>.` plus
  `Details: <sanitized deepest cause>` (max 280 chars).
- Cause chain is preserved via the error's `Cause`.
- Sanitizer rewrites URL query values for `key`, `api_key`, `api-key`, `token`, `access_token`,
  `secret` to `***`, then collapses whitespace and truncates. This matters because Gemini puts
  the API key in the URL query.
- Suggestions by failure class, checked against the sanitized text (lowercased):
  - `status 401|403|unauthorized|authentication` → verify key, set new key, check permissions.
  - `status 429|rate limit|quota` → wait, check quota.
  - `status 5xx|500|502|503|504|service unavailable` → provider status page, retry later.
  - network words (`dial tcp`, `no such host`, `connection refused`, `x509`, `tls`,
    `timeout`, `deadline exceeded`) → internet, proxy, `--proxy`.
  - fallback → internet, provider status, inspect config, re-run `gmsg provider test <p>`.

## Testing

- `internal/config/config_test.go`: `ResolveAPIKey` matrix (config wins over env; env fallback;
  ollama/lmstudio not-required even with config key; missing), plus one `GetClientConfig` test that
  still succeeds from env and still errors with `APIKeyNotSetError` when neither exists.
- `internal/provider/provider_test.go` (temp config via `testutils.TestConfig` + `config.New`):
  - inventory includes registered providers + configured custom key, excludes `output`/`console`,
    sorted, active marker correct;
  - model/base from config vs provider default;
  - key source config/env/not-required/missing (`t.Setenv` for determinism);
  - `Use` seeds a missing registered block (then `GetClientConfig` succeeds), rejects unknown,
    accepts configured custom, never rewrites an existing block, reports azure-style missing
    fields, reports already-active.
- `cmd/newprovider_test.go`: existing tests, renamed only.
- `cmd/provider_test.go`: command shape; list output (masked key present, full key absent, custom
  provider row, hint lines); use success persists config; use unknown fails; warning text for
  missing key; test uses injected fake prober (success + failure + named provider + no-VCS).
- `cmd/provider_probe_test.go`: `runProviderTest` with a fake `providerProber` (success, error,
  retries=1 passed to config, latency field set, unknown/unconfigured provider, name normalization,
  no-repo path) and preview helpers.
- `internal/errors/errors_test.go`: new templates' titles, suggestions, latency substring, secret
  sanitization, `errors.As` typed checks.
- Manual: temp config, `gmsg provider list`, `use ollama`, `use deepseek`, `use nope`,
  `provider test` with a fake key (expect auth error rendering), then `gmsg config path`.

## Compatibility & rollback

- No existing command, flag, config key, or LLM behavior changes. `newprovider`, `config set
  provider`, and `GetClientConfig` keep their contracts (verified by existing tests).
- New config keys are not introduced; `provider use` only creates a provider section that
  `newprovider` would have created anyway.
- Rollback is a code revert; no migrations, no persisted schema change beyond the ordinary
  `provider`/provider-section writes the user explicitly requested.
- `provider test` consumes a real minimal request by design; it is opt-in and never runs as part
  of another command.

## Trade-offs

- A new `internal/provider` package is justified by the inventory/switch rules and their tests; a
  command-only implementation would either duplicate config/env logic or be untestable.
- Display defaults come from `GetRequiredConfig()` instead of `GetClientConfig` because the latter
  requires a key and errors for unconfigured providers; the test command deliberately uses
  `GetClientConfig` so the probe matches runtime behavior.
- Seeding on `use` slightly changes state as a side effect of switching. The alternative (switch
  then fail with "Provider config not found") is worse UX, and the seeded values are exactly the
  defaults `newprovider` shows.
