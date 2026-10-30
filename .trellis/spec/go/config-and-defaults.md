# Config & Defaults

All user configuration lives in one YAML file; every default lives in
`pkg/config/defaults/defaults.go`. Nothing else may invent a default value.

## File & Resolution

- Default path: `~/.config/gptcomet/gptcomet.yaml` (`getConfigDir` in
  `internal/config/config.go` = `os.UserHomeDir()` + `.config/gptcomet`; there is **no**
  `XDG_CONFIG_HOME` support).
- Override with `--config/-c <path>` (root persistent flag) or `gmsg config path`.
- An explicitly passed path must exist, otherwise `errors.ConfigFileNotFoundError` is returned
  (`config.New`). A missing default path is created with `defaults.DefaultConfig`.
- The `prompt` section is excluded from `gmsg config list`; API keys are masked there.

Precedence: **CLI flag override → config file value → `defaults.DefaultConfig` → hard-coded
fallback**. Flags are applied after `Manager.GetClientConfig`, never before.

## Manager API (`internal/config`)

`ManagerInterface` (`config.go:33`) is the seam used by commands and mocks:

| Method | Use |
| --- | --- |
| `Get` / `GetWithDefault` / `Set` / `Remove` / `Append` | dot-key access, e.g. `output.lang`, `openai.model` |
| `GetNestedValue` / `SetNestedValue` | explicit key-slice access |
| `List` / `ListWithoutPrompt` | YAML rendering (secrets masked, prompt excluded) |
| `Save` / `Load` / `Reset` | persistence; `Reset(true)` only resets prompts |
| `GetClientConfig(provider)` | the only supported way to build `types.ClientConfig` |
| `ResolveAPIKey(provider)` | where the key resolves from (`config` / `env` / `not-required` / `missing`); the single source of the config→env fallback, used by `GetClientConfig` and `gmsg provider list` |
| `GetSupportedKeys` | `gmsg config keys`, used for validation and help |
| `GetPrompt(isRich)` / `GetReviewPrompt()` / `GetTranslationPrompt()` | prompt lookup with default fallback |
| `GetFileIgnore()` / `GetOutputTranslateTitle()` | convenience getters |

`GetClientConfig` is more than a getter — it also assembles defaults and provider special cases
(see `internal/config/config.go:124`):

- provider comes from the `provider` key or the `--provider` override; a missing section is a
  `NewConfigError`;
- API key falls back to `errors.GetAPIKeyEnvVar(provider)` **only when** the config value is
  empty (config wins over env, per the comment in `client_config.go`);
- `ollama` / `lmstudio` may run without an API key (`providerAllowsEmptyAPIKey`);
- provider-specific `api_base` fallbacks: `ollama` → `http://localhost:11434/api`,
  `lmstudio` → `http://localhost:1234/v1`, `bedrock` → region endpoint (`us-east-1` default),
  `cloudflare` → account-id endpoint, `qianfan`, `doubao`, `gemini`, `vertex`;
- retries/timeout default from `defaults.DefaultRetries` and `constants.DefaultTimeoutSeconds`.

## Key Families

`GetSupportedKeys` (`internal/config/accessor.go:122`) returns these (provider name replaced by
`<provider>`):

| Family | Keys |
| --- | --- |
| Root | `provider`, `file_ignore` |
| Output | `output.lang`, `output.rich_template`, `output.translate_title`, `output.review_lang`, `output.markdown_theme` |
| Console | `console.verbose` |
| Provider | `<provider>.api_base`, `.api_key`, `.model`, `.retries`, `.timeout`, `.proxy`, `.max_tokens`, `.top_p`, `.temperature`, `.frequency_penalty`, `.extra_headers`, `.extra_body`, `.completion_path`, `.answer_path`, `.stream_answer_path`, `.region`, `.account_id` |
| Prompt | `prompt.brief_commit_message`, `prompt.rich_commit_message`, `prompt.translation`, `prompt.review` |

Known gap: `GetClientConfig` also reads `<provider>.api_version`, `.deployment_name`,
`.project_id`, `.location`, `.anthropic_version`, but `GetSupportedKeys` does not list them.
When adding a key that `GetClientConfig` reads, add it to both places.

## Defaults (`pkg/config/defaults`)

- Constants: `DefaultAPIBase = "https://api.openai.com/v1"`, `DefaultModel = "gpt-6.1-sol"`,
  `DefaultRetries = 3`, `DefaultMaxTokens = 1024`, `DefaultTemperature = 0.3`, `DefaultTopP = 1.0`,
  `DefaultFrequencyPenalty = 0.0`, `DefaultRichTemplate = "<title>:<summary>\n\n<detail>"`.
- `DefaultConfig` (`defaultConfig()`): `provider: openai`, `file_ignore` (lockfiles, `go.sum`,
  `uv.lock`, `*.py[cod]`, …), `output`, `console`, and per-provider blocks seeded with the same
  keys as the provider table. Per-provider blocks keep their own `retries` / `max_tokens` /
  `top_p` / `temperature` values — they are intentionally not identical across providers.
- `PromptDefaults`: `brief_commit_message`, `rich_commit_message`, `translation`, `review`.
  Prompt strings may embed `{{ output.lang }}`, `{{ output.rich_template }}` placeholders that are
  filled by `internal/prompt.Render` (`internal/prompt/render.go`).
- The file header records when model defaults were last verified — update it when refreshing
  models (see [docs/provider-models.md](../../../docs/provider-models.md)).

## Adding a Config Key — Checklist

1. Add the default to `pkg/config/defaults/defaults.go` (`DefaultConfig` and/or a `Default*`
   constant). No literal defaults anywhere else.
2. If the key affects the runtime client, map it in `Manager.GetClientConfig` and in
   `types.ClientConfig` (`pkg/types/types.go`) with the matching JSON tag.
3. Add the key string to `GetSupportedKeys` in `internal/config/accessor.go` (grouped by family).
4. Add it to the `gmsg config` docs: README config section + `README.zh.md` + `CLAUDE.md`
   "Supported keys" if user-visible.
5. Extend tests: `pkg/config/defaults/defaults_test.go` for the default,
   `internal/config/config_test.go` for set/get/list behaviour, and a flag test in
   `cmd/flags_test.go` if it gets a CLI override.

## Rules

- Never read `os.Getenv` for configuration outside `GetClientConfig`'s API-key fallback.
- Never hard-code `~/.config` or `gptcomet.yaml` outside `getConfigDir` / `errors` templates.
- List/`gmsg config list` output must stay masked — do not add secrets to `List` paths.
- Prompt text belongs in `PromptDefaults` (or the user's config), never inlined in commands.
