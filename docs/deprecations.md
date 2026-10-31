# Deprecations & Removal Schedule

This page tracks deprecated user-facing behavior and when it is scheduled to be
removed. The policy is deliberately conservative: a deprecated command keeps
working for **five minor releases after the release that deprecates it**, so
scripts, docs, and muscle memory have time to migrate.

## Policy

- The deprecating release announces the replacement in the CLI (a `⚠` line
  printed by the command itself), in the READMEs, and in `CHANGELOG.md`.
- The replacement ships in the same release as the deprecation, never later.
- Removal happens **no earlier than** the announced target version. If a release
  is skipped, the target moves to the matching later minor release — it never
  moves earlier.
- A removal release deletes the command, its tests, its docs, and any stale
  suggestion text, and records the removal in `CHANGELOG.md`.

## Schedule

| Deprecated | Deprecated in | Removal target | Replacement | Status |
| --- | --- | --- | --- | --- |
| `gmsg newprovider` | `2.9.0` | `2.14.0` | `gmsg provider new` | Deprecated — prints a warning, stays visible in `--help` |

### `gmsg newprovider`

- **Deprecated in**: `2.9.0` (the first release whose CLI prints the warning)
- **Removal target**: `2.14.0` (five minor releases after the deprecation:
  `2.10.0`, `2.11.0`, `2.12.0`, `2.13.0`, `2.14.0`)
- **Replacement**: `gmsg provider new` — same interactive flow, same prompts,
  same result; it just lives under the `provider` command group.
- **Why**: provider management is grouped under `gmsg provider`. The legacy
  top-level command duplicates `provider new` and splits the discoverable
  surface.
- **Migration**: replace `gmsg newprovider` with `gmsg provider new` in scripts
  and documentation. `gmsg provider list`, `gmsg provider use <name>`, and
  `gmsg provider test` are unchanged.
- **Current UX**: the command stays runnable and visible in `gmsg --help`,
  prints
  `⚠ 'gmsg newprovider' is deprecated and will be removed in v2.14.0. Use 'gmsg provider new' instead.`
  and adds `(deprecated)` to its short description. It deliberately does **not**
  use Cobra's `Deprecated` field, because Cobra would hide it from help before
  the announced target.

#### History

- `2.9.0` release notes mention a `2.11.0` removal target. The target was
  extended to `2.14.0` after the release; the policy allows moving a removal
  later, never earlier.

#### Removal checklist (`2.14.0`)

- [ ] Delete `NewProviderSetupCmd` and `providerSetupDeprecationNotice` from
      `cmd/newprovider.go`; keep `runProviderSetup` and `newProviderNewCmd`
      (consider renaming the file to `cmd/provider_setup.go`).
- [ ] Remove `rootCmd.AddCommand(cmd.NewProviderSetupCmd())` and its doc-comment
      line from `main.go`.
- [ ] Delete `cmd/newprovider_test.go`; keep the non-interactive coverage for
      `provider new` in `cmd/provider_test.go`.
- [ ] `grep -rn "gmsg newprovider"` and drop the remaining README / `CLAUDE.md` /
      `.trellis/spec` mentions. `SuggConfigureProvider` already points at
      `gmsg provider new`.
- [ ] Move this row to a `Removed` table, keep the history, and add the removal
      note to `CHANGELOG.md` (the changelog is generated from commits, so the
      removal commit subject must mention the removal).
- [ ] Run `just format`, `just test`, `just check` (and `go test -race ./...` in
      CI).
