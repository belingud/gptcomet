# Go Quality Guidelines

The bar is: `go vet` + `staticcheck` clean, `go test ./...` green, race-clean under
`go test -race ./...`, and every new behavior covered by a table-driven test next to the code.

## Commands

| Need | Command |
| --- | --- |
| Format | `just format` (`go fmt ./...`, `goimports -w .`, `ruff format py/`) |
| Lint | `just check` (`go vet ./...`, `staticcheck ./...`) |
| Unit tests | `just test` (`go test ./...`) |
| Coverage | `just test-coverage` / `just test-cover-func` |
| Full gate (Go + Python) | `just ci-check` (`go vet`, `staticcheck`, `ruff check py/`) |

CI (`.github/workflows/quality.yml`) runs `go test -v -race ./...`, `go vet ./...`,
`staticcheck` (2026.2.2), and uploads coverage to Codecov. Tests must therefore be deterministic
and race-free; there is no `-short` escape hatch and no build tags in this repo.

## Test Layout

- Unit tests sit beside the code in the same package: `internal/llm/openai.go` →
  `internal/llm/openai_test.go`. 62 test files follow this rule; keep it.
- Cross-package flows live in `tests/integration/` (package `integration`) and use the
  `internal/testutils` helpers.
- Python tests are separate — see [../py/index.md](../py/index.md).

## Table-Driven Pattern

Every non-trivial test uses a `tests` slice plus `t.Run` (this exact shape appears from
`internal/llm/openai_test.go` to `internal/config/config_test.go`):

```go
func TestNewConfig(t *testing.T) {
    tests := []struct {
        name        string
        configPath  string
        configData  string
        wantErr     bool
        errContains string
    }{
        {name: "Success with empty config", configData: ""},
        {name: "Invalid YAML", configData: "provider: [invalid", wantErr: true, errContains: "Failed to load config"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // arrange with tt.*, act, assert
        })
    }
}
```

- Field names: `name`, `wantErr`, `errContains`, `want`; subtest names are lowercase sentences.
- `require` for preconditions that make the rest meaningless, `assert` for the actual assertions.
- Use `t.TempDir()` for anything that touches disk and `t.Setenv` for env-var paths; never write
  into the repo working tree.

## Mocks & Fakes

`internal/testutils` provides the only sanctioned test doubles:

| Double | Implements | Use |
| --- | --- | --- |
| `MockLLM` | `llm.LLM` | provider call assertions (`BuildURL`, `ParseResponse`, …) |
| `MockConfigManager` | `config.ManagerInterface` | command tests; `mock.On("Get", "x").Return(v, true)` |
| `MockVCS` | `git.VCS` | stub diffs/branches without a repo |
| `TestGitRepo`, `TestConfig`, `TestFile`, `TestDir`, `InitGitRepo`, `StageFile` | helpers | real temp repo/file/config fixtures |

Mock conventions:

```go
mock := &testutils.MockConfigManager{}
mock.On("Get", "test.key").Return("test-value", true)
...
mock.AssertExpectations(t)
```

- Always call `AssertExpectations(t)` so unused or missing calls fail the test.
- Prefer a real temp repo (`testutils.TestGitRepo(t)`) over `MockVCS` when the behavior under test
  is git integration itself.
- `httptest.NewServer` is the pattern for HTTP tests (`internal/client/client_test.go`,
  `internal/git/*_test.go`); never hit the real network.

## Non-Interactive Guarantees

Commands that open a bubbletea/TTY UI must short-circuit in tests: `cmd/newprovider.go` checks
`os.Getenv("GPTCOMET_TEST") == "1"` and prints the provider list instead of running the selector.
Follow that convention when adding interactive commands, otherwise `go test` hangs in CI.

Timing: `internal/ui/progress_test.go` uses small `time.Sleep` calls for render timing. Do not
copy that pattern into new tests — assert on state transitions instead.

## Review Checklist

Before reporting a Go change complete:

- [ ] `just format` applied (no `gofmt` diff).
- [ ] `just check` clean; no new `staticcheck` suppressions (`//lint:ignore` needs a reason).
- [ ] `just test` green, including the package you touched and its dependents
      (`go test ./internal/... ./cmd/...`).
- [ ] New behavior has a table-driven test; new provider has `<provider>_test.go`.
- [ ] Errors use templates from `internal/errors`, wrapped with `%w`.
- [ ] No new `fmt.Print*` inside `internal/` (see [cli-and-output.md](./cli-and-output.md)).
- [ ] Config keys/defaults added in both `pkg/config/defaults` and `GetSupportedKeys`.
- [ ] Docs touched when behavior is user-visible (`README.md`, `README.zh.md`, `CLAUDE.md`,
      `docs/provider-models.md`).
