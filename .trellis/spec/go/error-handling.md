# Error Handling

Every failure a user can see must arrive as a `*errors.GPTCometError` built from a template, not
as a bare `fmt.Errorf`. The rendered message is the product's error UX — it carries an icon, a
title, an explanation, numbered fixes, and optionally a docs link.

## The Type

```go
// internal/errors/errors.go
type GPTCometError struct {
    Type        ErrorType   // network | config | git | api | validation | unknown
    Title       string
    Message     string
    Details     string      // sanitized excerpt of Cause rendered to the user
    Cause       error       // exposed through Unwrap()
    Suggestions []string
    DocsURL     string
}
```

`Error()` renders:

```
🌐 Network Error

Failed to connect to https://api.openai.com/v1

Details: dial tcp: lookup api.openai.com: no such host

What to do:
  • Check your internet connection
  • Verify the API base URL in your config

Docs: https://github.com/belingud/gptcomet#configuration
```

Icons come from `getErrorIcon` (`errors.go:140`): network 🌐, config ⚙️, git 📦, api 🔑,
validation ⚠️, unknown ❓. Suggestion bullets are `•`. Shared titles live in
`internal/errors/constants.go`; each constructor sets its own `DocsURL` (config errors point at
the README configuration section).

## Two Layers of Constructors

1. **Typed helpers** — `NewConfigError`, `NewNetworkError`, `NewGitError`, `NewAPIError`,
   `NewValidationError`, plus `WrapError(err, title, message)` for unknown causes:

   ```go
   return gptErrors.NewValidationError(
       "Invalid Configuration",
       "Configuration object is nil",
       nil,
       []string{"Ensure valid configuration is provided"},
   )
   ```

2. **Scenario templates** (`internal/errors/templates.go`) — one function per real user
   situation, so the wording and fix instructions live in one place:
   `ConfigFileNotFoundError`, `APIKeyNotSetError`, `NoStagedChangesError`,
   `NetworkConnectionError`, `APIRateLimitError`, `APIAuthenticationError`,
   `InvalidConfigValueError`, `GitRepositoryNotFoundError(path)`, `GitExecutableNotFoundError`,
   `GitCommandFailedError`, `ModelNotFoundError`, `ProviderCreationError`, `VCSCreationError`, `ProxyConfigurationError`,
   `RequestCreationError`, `RequestExecutionError`, `ResponseParsingError`,
   `DependencyCreationError`, `InternalError`, `RequestRetryError`, `MessageFormattingError`,
   `RequestMarshalingError`, `APIStatusError`, `CallbackError`, `UnsupportedProxySchemeError`,
   `ProxyURLParseError`.
   `GetAPIKeyEnvVar(provider)` also lives here — the env-var naming contract for API keys.

Prefer adding a template over assembling a `GPTCometError` inline: inline construction is how the
same problem ends up with three different messages.

## Where Errors Are Raised

| Layer | Style |
| --- | --- |
| `internal/llm` | plain `fmt.Errorf("failed to ...: %w", err)` for protocol issues; the client wraps them |
| `internal/client` | templates for network/API/retry (`NetworkConnectionError`, `APIRateLimitError`, `RequestRetryError`, …) |
| `internal/config` | templates for missing/invalid config (`ConfigFileNotFoundError`, `APIKeyNotSetError`, `InvalidConfigValueError`) |
| `internal/git` | templates for repo/command failures (`GitRepositoryNotFoundError`, `GitCommandFailedError`) |
| `internal/factory` | `DependencyCreationError("config manager", err)` etc. — names the dependency that failed |
| `cmd/` | mostly pass-through: `RunE` returns the error |

## Rules

1. **Wrap with `%w`**, never `%v`, so `errors.Is` / `errors.As` keep working
   (`GPTCometError.Unwrap()` returns `Cause`).
2. **Return, do not print.** `main.go` sets `SilenceUsage: true` and Cobra prints `Error: <msg>`
   to stderr. A `RunE` that both prints and returns produces duplicated output.
3. **Give the user an action.** Any error that is not a programming bug needs a `Suggestions`
   slice with concrete commands (`gptcomet config set <provider>.api_key <key>`) or checks.
4. **No secrets in messages.** Do not interpolate API keys or full URLs with query strings;
   `sanitizeURLForLogging` in `internal/client/client.go` and
   `sanitizeRequestURLForLogging` in `internal/llm/llm.go` exist for this.
5. **Keep the type consistent with the cause.** A 401 is `api`/`NetworkConnectionError`-class, a
   missing YAML key is `config`, a missing `.git` is `git`.
6. **Never swallow.** If a fallback succeeds (e.g. provider falls back to `openai`), log at debug
   level; do not silently drop the original error when the fallback also fails.
7. **Exit codes follow the type.** `errors.ExitCode` (`internal/errors/exit.go`) maps
   `GPTCometError.Type` to the process exit code (config 3, network 4, git 5, api 6, validation 7,
   unknown 1) and `main.go` exits with it; a flag-parsing error is a `*errors.UsageError` → 2.
   Keep the type accurate — scripts branch on the exit code, not on the message. See
   [cli-and-output.md](./cli-and-output.md#output-contract).
8. **Surface the cause in `Details`.** When the fix is "look at the output" (git stderr, exec
   failure), put a sanitized excerpt in `Details` via `sanitizeErrorDetail` (single line, capped
   at 280 runes, credential query values redacted). `GitCommandFailedError` and
   `GitExecutableNotFoundError` already do this; never append raw multi-line output to `Message`
   and never rely on the user finding the `Cause` (it is not rendered).
9. **Classify git failures by exit status, never by localized stderr.** Git translates its
   messages (this machine runs zh_CN), and since Git 2.55 `git diff --staged` outside a
   repository falls back to `--no-index` mode and fails with an unknown option (exit 129)
   instead of "not a git repository". `internal/git.gitCommandError` centralizes the decision:
   `*exec.Error` → `GitExecutableNotFoundError`, exit failure with a failed
   `git rev-parse --is-inside-work-tree` probe → `GitRepositoryNotFoundError(path)`, otherwise
   `GitCommandFailedError` with the joined command label (`git diff --staged -U2`). Do not add
   stderr string matching or a probe on the success path.

## Scenario: Git Failure Classification and Details

### 1. Scope / Trigger

Every user-visible git failure goes through `internal/git` (the only layer that shells out). v2.11.1
changed this contract: failures outside a repository were misclassified as generic command
failures, git's real stderr was hidden, and the exit code leaked into the quoted command. The
fix is a code-spec because error class + rendered output + exit code are cross-layer contracts.

### 2. Signatures

```go
// internal/errors
func GitRepositoryNotFoundError(repoPath string) *GPTCometError // type git, exit 5
func GitExecutableNotFoundError(command string, cause error) *GPTCometError // type git, exit 5
func GitCommandFailedError(command string, cause error) *GPTCometError // type git, exit 5

// internal/git (unexported, single decision point)
func gitCommandError(cmd *exec.Cmd, err error, stderr string) error
func isGitWorkingTree(repoPath string) bool // probes rev-parse --is-inside-work-tree
```

### 3. Contracts

- `Details` is a sanitized single-line excerpt produced by `sanitizeErrorDetail` (≤280 runes,
  credential query values redacted). It renders between `Message` and `What to do:`; empty
  Details render nothing. Never put raw multi-line output in `Message`.
- Command label is `strings.Join(cmd.Args, " ")` → `git diff --staged -U2`; no `[...]` and no
  exit code inside the quoted command.
- Classification inputs are the Go error class and the repository probe result — never git's
  stderr text (localized, version-dependent).

### 4. Validation & Error Matrix

| Condition | Error | Details | Exit |
| --- | --- | --- | --- |
| `git` cannot be executed (`*exec.Error`) | `Git Not Installed` | exec error | 5 |
| exit failure + probe not `true` | `Not a Git Repository(<abs path>)` | — | 5 |
| exit failure inside a work tree | `Git Command Failed` | exit status + stderr | 5 |
| command succeeds | — | — | 0 |

### 5. Good / Base / Bad Cases

- Good: non-repo `gmsg commit` → `Not a Git Repository`, absolute path, `git init` suggestion,
  empty stdout, exit 5. Bad: `Git Command Failed` + "check if git is installed" for the same case.
- Base: in-repo staged diff succeeds with no extra probe subprocess (probe is failure-path only).
- Bad: matching `"not a git repository"` in stderr; `git [diff --staged -U2]` labels; printing
  `Details` to stdout.

### 6. Tests Required

- Unit, `internal/git`: `TestGitVCSOutsideRepository` (all exported commands → Not a Git
  Repository + path), `TestGitCommandFailureIncludesDetails` (in-repo failure keeps Git Command
  Failed, Details has `exit status 128`), `TestGitCommandNotInstalled` (PATH without git →
  Git Not Installed), `TestRepoDisplayPath`.
- Unit, `internal/errors`: template tests assert type, title, path/command, Details excerpt, and
  suggestion text; one `Error()` case asserts `Details:` renders.
- E2E: non-repo `commit`/`generate`/`review`/`hook status` → exit 5, empty stdout, stderr contains
  `Not a Git Repository` + `git init`, no install blame; wrapper-forced in-repo failure → Details
  with git stderr; PATH without git → `Git Not Installed`.

### 7. Wrong vs Correct

```go
// Wrong: loses the real class, hides git output, exit code inside the quoted command.
cmdStr := fmt.Sprintf("git diff --staged --name-only (exit code %d)", exitErr.ExitCode())
return gptcometerrors.GitCommandFailedError(cmdStr, fmt.Errorf("%w\nOutput: %s", err, stderr))

// Correct: one classifier probes only on failure and keeps stderr in Details.
return gitCommandError(cmd, err, stderr.String())
```

## Testing

- Table tests assert on substrings of the rendered message, not on the Go error type:
  `errContains: "Failed to load config"` (`internal/config/config_test.go`).
- For template functions, test the suggestion text and docs URL, since those drive the UX
  (`internal/errors/errors_test.go` follows this).
- When a test exercises a code path that must fail, assert the *type* as well
  (`errors.As(err, &gptErr)`) if the caller branches on `gptErr.Type`.
