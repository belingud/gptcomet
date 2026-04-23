## Why

Homebrew-managed installations should be updated by Homebrew so that the package manager keeps its receipts, symlinks, and installed version consistent. The current `gmsg update` path can replace the binary directly, which bypasses Homebrew and may leave the installation in an inconsistent state.

## What Changes

- Detect when the running `gmsg`/`gptcomet` binary is managed by Homebrew.
- Prevent `gmsg update` from performing the built-in GitHub release update flow for Homebrew-managed installations.
- Show a clear error message that tells the user to update with Homebrew instead.
- Keep the existing built-in update flow unchanged for non-Homebrew installations.
- Add tests that prove Homebrew-managed installs are blocked before any GitHub release request or binary replacement starts.

## Capabilities

### New Capabilities

- `self-update-policy`: Defines when the CLI may use its built-in update command and when it must defer to the installation manager.

### Modified Capabilities

None.

## Impact

- Affected CLI command: `gmsg update`.
- Affected Go code: update command construction and update preflight checks in `cmd/update.go`.
- Affected tests: update command tests in `cmd/update_test.go`, with possible command-level coverage for the new guard.
- Affected release packaging: Homebrew builds need a reliable marker that lets the binary identify a Homebrew-managed installation.
