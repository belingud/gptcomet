# LLM Provider Guidelines

`internal/llm/` is where the project grows. 27 providers are registered in
`internal/llm/provider.go` `init()`; each one is a struct that embeds `*BaseLLM` and overrides
only the parts of the HTTP contract that differ.

Read this file completely before touching `internal/llm/`.

## The Contract

```go
// internal/llm/llm.go
type LLM interface {
    Name() string
    BuildURL() string
    GetRequiredConfig() map[string]config.ConfigRequirement
    FormatMessages(message string) (interface{}, error)
    MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error)
    GetUsage(data []byte) (string, error)
    BuildHeaders() map[string]string
    ParseResponse(response []byte) (string, error)
}
```

`BaseLLM` (`llm.go`) implements every method, driven by `types.ClientConfig`:

| Method | Default behaviour | Override only when |
| --- | --- | --- |
| `Name()` | — | always: return the registry key (`"openai"`, `"claude"`, …) |
| `BuildURL()` | `APIBase` + `/` + `*CompletionPath` | auth in the query string (gemini `?key=`), model in the path (vertex) |
| `BuildHeaders()` | `Content-Type` + `Authorization: Bearer <api_key>` + `ExtraHeaders` | header-based auth (`x-api-key`, `anthropic-version`) — see `claude.go` |
| `FormatMessages()` | OpenAI `{model, messages[], max_completion_tokens, …}` | non-OpenAI payloads (Anthropic `system` + `messages`, Gemini `contents`) |
| `ParseResponse()` | `gjson` at `AnswerPath`, strips surrounding code fences | nested/multi-choice responses |
| `GetUsage()` | `usage.prompt_tokens/completion_tokens/total_tokens` | different usage key names |
| `MakeRequest()` | full pipeline: URL → headers → payload → POST → usage → parse | essentially never; providers delegate (see below) |

## Anatomy of a Provider

Default-shaped providers (`openai.go`, `deepseek.go`, `groq.go`, `xai.go`, …) are ~110 lines and
follow exactly this shape:

```go
// internal/llm/openai.go
const (
    DefaultOpenAIAPIBase = "https://api.openai.com/v1"
    DefaultOpenAIModel   = "gpt-6.1-sol"
)

type OpenAILLM struct {
    *BaseLLM
}

func NewOpenAILLM(config *types.ClientConfig) *OpenAILLM {
    BuildStandardConfigSimple(config, DefaultOpenAIAPIBase, DefaultOpenAIModel)
    return &OpenAILLM{BaseLLM: NewBaseLLM(config)}
}

func (o *OpenAILLM) Name() string { return "openai" }

func (o *OpenAILLM) GetRequiredConfig() map[string]config.ConfigRequirement {
    return map[string]config.ConfigRequirement{
        "api_base":   {DefaultValue: DefaultOpenAIAPIBase, PromptMessage: "Enter OpenAI API base URL"},
        "model":      {DefaultValue: DefaultOpenAIModel, PromptMessage: "Enter model name"},
        "api_key":    {DefaultValue: "", PromptMessage: "Enter API key"},
        "max_tokens": {DefaultValue: "1024", PromptMessage: "Enter max tokens"},
    }
}

func (o *OpenAILLM) MakeRequest(ctx context.Context, client *http.Client, message string, stream bool) (string, error) {
    return o.BaseLLM.MakeRequest(ctx, client, o, message, stream)
}
```

The `MakeRequest` one-liner is **not** redundant: `BaseLLM` needs the concrete provider so
virtual dispatch reaches the overridden `BuildURL`/`ParseResponse`. Do not delete it, and do not
implement request logic per provider.

Providers with a non-OpenAI wire format (e.g. `claude.go`, `gemini.go`, `cohere.go`,
`bedrock.go`) override `FormatMessages`, `BuildURL`, `BuildHeaders` and/or `ParseResponse`
instead. `gemini.go` additionally has `BuildStreamURL()` for its streaming endpoint.

Builders (`internal/llm/builder.go`):

- `BuildStandardConfigSimple(config, apiBase, model)` — OpenAI-compatible paths.
- `BuildStandardConfig(config, apiBase, model, completionPath, answerPath)` — when the provider
  uses different paths; empty strings fall back to `chat/completions` and
  `choices.0.message.content`.
- `SetDefault*` helpers only write when the user has not set the value — never overwrite an
  explicit user config.

## Adding a Provider — Checklist

1. Create `internal/llm/<name>.go` with the `Default<Name>APIBase` / `Default<Name>Model`
   constants, constructor, `Name()`, `GetRequiredConfig()`, `MakeRequest()` delegation, plus any
   overrides the protocol needs.
2. Register it in `internal/llm/provider.go` `init()` — same loop shape as the existing entries:

   ```go
   RegisterProvider("<name>", func(config *types.ClientConfig) LLM {
       return New<Name>LLM(config)
   })
   ```

3. Add `internal/llm/<name>_test.go`: a table-driven `TestNew<Name>LLM` covering the default
   config and a fully custom config. Copy the structure of `openai_test.go` (it asserts
   `api_base`, `model`, `completionPath`, `answerPath`, and uses the package-level `customPath`
   variable).
4. If the provider needs env-var API keys, confirm `errors.GetAPIKeyEnvVar` returns the right
   variable name (`internal/errors/templates.go`); unknown providers fall back to
   `<UPPER_NAME>_API_KEY`.
5. If it is a local/servers model without auth (like `ollama`, `lmstudio`), add it to
   `providerAllowsEmptyAPIKey` in `internal/config/config.go` and give it a local `api_base`
   special case in `GetClientConfig`.
6. Update the default-model table and follow the refresh checklist in
   [docs/provider-models.md](../../../docs/provider-models.md) — model constants, defaults,
   README/README.zh examples, `CLAUDE.md`.
7. Run `go test ./internal/llm/ ./internal/client/ ./internal/config/` and `just check`.

## Rules

- **`Name()` is the registry key and the YAML section name.** `provider: <name>` in the config
  must resolve to the same string; `GetClientConfig` looks up `m.config[provider]`.
- **No `internal/client` import.** The HTTP client is a parameter, not a dependency.
- **Never log secrets.** URL-sanitising lives in `internal/client/client.go`
  (`sanitizeURLForLogging` strips query params for gemini/vertex). When logging a request URL
  from a provider, reuse that pattern.
- **Streaming is central.** `BaseLLM.MakeRequest` injects `"stream": true` when the `stream`
  argument is set; SSE parsing lives in `client.Stream`/`streamOnce`. Providers only declare
  `StreamAnswerPath` (default `choices.0.delta.content`) when the delta shape differs.
- **Defaults live in two places** — the constructor (`BuildStandardConfig*`) and
  `Manager.GetClientConfig` special cases. Change both or neither.
- **One provider per file, one test file per provider.** No shared mutable globals; the only
  global is the registry.

## Anti-Patterns

- Adding a `RegisterProvider` entry without `internal/llm/<name>.go` and its test.
- Hard-coding model names in tests instead of `Default<Name>Model` (the model refresh checklist
  depends on the constant being the single source).
- Returning `BaseLLM` from a provider constructor — the concrete type is required for dispatch.
- Building request payloads in `cmd/` or `internal/client/` — payload shaping belongs to the
  provider.
- Adding a new `fmt.Printf` for progress inside the provider; use `internal/logger` or the
  `ui.Progress` stages owned by the command layer.
