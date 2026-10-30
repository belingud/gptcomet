# Implementation — Provider management commands

## Preconditions

- PRD + design reviewed; work starts only after `task.py start`.
- Before editing code, the Go guidelines are already loaded: `cli-and-output.md`,
  `config-and-defaults.md`, `error-handling.md`, `quality-guidelines.md`,
  `architecture.md`.
- Work on `master` or a feature branch; do not touch unrelated untracked Trellis/agent
  scaffolding. `FUTURE_FEATURES.md` stays uncommitted.

## Ordered checklist

### 1. Config key resolution

- [ ] `internal/config/config.go`: add `APIKeySource` type + four constants and
      `(*Manager).ResolveAPIKey(provider) (string, APIKeySource)` next to
      `providerAllowsEmptyAPIKey`.
- [ ] Add `ResolveAPIKey` to `ManagerInterface`; replace the inline API-key block in
      `GetClientConfig` with the resolver, keeping `APIKeyNotSetError` as the missing case.
- [ ] `internal/testutils/mock_config.go`: implement `ResolveAPIKey` through `m.Called(provider)`.
- [ ] `internal/config/config_test.go`: table tests for config/env/not-required/missing and two
      `GetClientConfig` regression cases (env fallback succeeds, neither source errors).
- [ ] Run `go test ./internal/config/`.

### 2. Provider domain package

- [ ] Add `internal/provider/provider.go`:
  - `Info`, `SwitchResult`, `Service`, `NewService(cfg config.ManagerInterface)`.
  - `CurrentName`, `NormalizeName`, `List`, `Get`, `Use` exactly as in `design.md`.
  - reserved root keys constant (`provider`, `file_ignore`, `output`, `console`, `prompt`).
  - provider default lookup via `llm.GetProviderConstructor` + `GetRequiredConfig()`.
- [ ] Add `internal/provider/provider_test.go`:
  - name normalization (built-in lowercase, trimmed unknown);
  - inventory (registered + custom configured, reserved keys excluded, sorted, active);
  - model/base sources (config vs default);
  - key-source matrix with `t.Setenv`;
  - `Use`: seeds missing registered block and then `GetClientConfig` works; unknown → typed error
    and provider unchanged; configured custom accepted; existing block not rewritten; missing
    fields for `azure`; already-active; saved file reload check.
- [ ] Run `go test ./internal/provider/`.

### 3. Error templates

- [ ] `internal/errors/constants.go`: titles/messages/suggestions listed in `design.md`.
- [ ] `internal/errors/templates.go` (new file `provider_errors.go` is not needed; keep in
      `templates.go`): `ProviderNotFoundError`, `ProviderNotConfiguredError`,
      `NoProviderSelectedError`, `ProviderTestError` plus unexported `errorDetailText`,
      `sanitizeErrorDetail`, `providerTestSuggestions`.
- [ ] `internal/errors/errors_test.go`: titles/suggestions per class (401, 429, 5xx, network,
      generic), latency substring, secret redaction (`?key=SECRET`), `errors.As` type checks.
- [ ] Run `go test ./internal/errors/`.

### 4. CLI command group

- [ ] `git mv cmd/provider.go cmd/newprovider.go`; rename constructor to `NewProviderSetupCmd`
      (command `Use: "newprovider"` unchanged).
- [ ] `git mv cmd/provider_test.go cmd/newprovider_test.go`; update constructor references/names.
- [ ] Add `cmd/provider.go`: `NewProviderCmd()` parent (`Use: "provider"`), `newProviderListCmd`,
      `newProviderUseCmd`, `providerConfigManager`, `formatKeySource`, `formatModel`,
      `formatAPIBase`, `renderProviderList`.
- [ ] Add `cmd/provider_probe.go`: `providerProber`, `newProviderProber` seam,
      `providerProbePrompt`, `providerTestResult`, `runProviderTest`, `newProviderTestCmd`,
      `formatProbeResponse` (trim/collapse/truncate).
- [ ] `main.go`: register `cmd.NewProviderCmd()` + `cmd.NewProviderSetupCmd()`; update the
      doc-comment command list.
- [ ] Add `cmd/provider_test.go` and `cmd/provider_probe_test.go` covering every acceptance
      criterion that can be unit-tested (fake prober; `t.Chdir` for the no-repo case).
- [ ] Run `go test ./cmd/`.

### 5. Documentation

- [ ] `README.md`: add `gmsg provider` to the command list and a short "Manage providers"
      subsection under Provider Setup Guide with `list` / `use` / `test` examples and key-source
      meaning; link it from the table of contents if the file has anchors.
- [ ] `README.zh.md`: full translation of the same content (not a stub).
- [ ] `CLAUDE.md`: add `gmsg provider` to Core Commands and `internal/provider` to the
      architecture tree.
- [ ] `docs/provider-models.md`: one line pointing at `gmsg provider list` for effective
      models/key sources.
- [ ] `.trellis/spec/go/architecture.md`: Package Map row for `internal/provider`.
- [ ] `.trellis/spec/go/config-and-defaults.md`: Manager API table row for `ResolveAPIKey`
      (and note it is the single source of the config→env fallback).

### 6. Quality gate

- [ ] `just format` (`gofmt`/`goimports` clean; no unintended `ruff` churn).
- [ ] `just test` green.
- [ ] `go test -race ./...` green.
- [ ] `just check` (`go vet`, `staticcheck`) green.
- [ ] `just ci-check` green.
- [ ] Manual smoke in a temp dir/config:
  - `gmsg provider list` (masked key, active marker, ollama not-required);
  - `gmsg provider use ollama` then `gmsg provider use openai`;
  - `gmsg provider use does-not-exist` (typed error, provider unchanged);
  - `gmsg provider test --help`, `gmsg provider list` outside a Git repo;
  - `gmsg newprovider` still starts (or prints the list under `GPTCOMET_TEST=1`).

### 7. Commit / finish

- [ ] Commit 1: `feat(config): expose API key source resolution`.
- [ ] Commit 2: `feat(provider): add provider list/use/test commands` (provider package, errors,
      CLI, tests, docs, spec updates).
- [ ] Keep `FUTURE_FEATURES.md` untracked; do not `git add .trellis` wholesale (only this task's
      artifacts if the developer agrees).
- [ ] Update the task artifacts/acceptance checkboxes, run `trellis-check`, record the session in
      the journal, and archive the task.

## Risk / rollback points

| Risk | Mitigation / rollback |
| --- | --- |
| Renaming `NewProviderCmd` breaks other references | `grep -rn NewProviderCmd` before/after; tests compile the package |
| `GetClientConfig` behavior drifts | resolver refactor is mechanical; regression tests for env fallback + missing key; existing `internal/config` and `internal/factory` tests |
| `provider use` writes surprising config | seed only when the block is absent; never modify existing keys; tests assert the existing block is byte-identical after use |
| Secrets leak through probe errors | sanitizer + test with `?key=SECRET`; masked keys in list; no raw key in messages |
| `provider test` hangs/retries | `Retries = 1`, configured timeout authority, single `Chat` call |
| Provider package becomes a second config layer | imports only `config.ManagerInterface`/`llm`/`errors`; runtime resolution still `GetClientConfig` |
| Rediscovery tests flip with ambient env vars | `t.Setenv` for every asserted provider env var |
| `cmd/provider_test.go` name collision with the `test` subcommand | probe implementation/tests live in `cmd/provider_probe*.go` |

## Definition of done

- All acceptance criteria in `prd.md` checked.
- `just format`, `just test`, `go test -race ./...`, `just check`, `just ci-check` pass.
- README/README.zh/CLAUDE/spec docs updated and consistent with actual output.
- `newprovider` behavior unchanged; `FUTURE_FEATURES.md` untouched.
- Task artifacts updated and archived.
