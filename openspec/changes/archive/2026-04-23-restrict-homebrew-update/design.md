## Context

`gmsg update` is implemented in Go in `cmd/update.go`. It checks GitHub releases, downloads the matching archive, replaces the installed binary under the user install directory, and recreates the `gmsg` symlink on Unix-like systems.

The Python wrapper already blocks `update` for PyPI-managed installs. Homebrew needs the same ownership boundary: once Homebrew installs the binary, Homebrew must be the only update mechanism for that installation.

## Goals / Non-Goals

**Goals:**

- Detect Homebrew-managed installations before the update command contacts GitHub or modifies files.
- Return a clear, actionable error that points users to `brew upgrade gptcomet`.
- Preserve the existing built-in update flow for standalone installations.
- Keep the detection logic small and testable.

**Non-Goals:**

- Do not add a generic package-manager framework.
- Do not change the release download, archive extraction, or binary replacement flow for standalone installs.
- Do not implement Homebrew formula changes in this repository.

## Decisions

### Use an installation-source guard before update checks

`NewUpdateCmd` should run a preflight guard before calling the existing update flow. If the guard identifies a Homebrew-managed install, it returns an error and stops. This keeps the network check and file replacement code unchanged for all other installs.

Alternative considered: add the guard inside `InstallUpdate`. That is too late because the command would still contact GitHub and present an update as available before rejecting it.

### Prefer an explicit build-time marker, with a Homebrew path fallback

The Go binary should expose a small package variable for installation source, defaulting to the standalone value. Homebrew packaging can set it with Go linker flags when it builds from source.

Because some Homebrew formulas install prebuilt release assets, the guard should also resolve `os.Executable()` through symlinks and recognize paths under a Homebrew Cellar for this project. The explicit marker remains the primary signal; the path fallback protects formulas that cannot inject build metadata.

Alternative considered: call `brew` at runtime to inspect ownership. That adds process execution, depends on `brew` being on `PATH`, and makes tests slower and more brittle.

### Keep user guidance specific

The rejection message should state that Homebrew-installed GPTComet cannot be updated with `gmsg update` and should be updated with `brew upgrade gptcomet`. The command should not suggest manual binary replacement for Homebrew installs.

Alternative considered: use a generic message like "update by the way you installed it." That matches the PyPI wrapper but gives Homebrew users less useful guidance.

## Risks / Trade-offs

- Homebrew path fallback could miss unusual formula layouts -> the explicit build-time marker remains the supported signal for the tap repository.
- Homebrew path fallback could reject a manually copied binary that happens to live inside a Cellar-like path -> the path check should require a project-specific Cellar segment such as `/Cellar/gptcomet/`.
- Linker marker package path can change during refactors -> tests should cover the default value and the Homebrew value so release failures are easier to find.

## Migration Plan

- Add the guard and tests in this repository.
- Update the Homebrew tap/formula repository to set the installation source marker when building from source, or rely on the Cellar path fallback when installing release assets.
- No data migration is required.

## Open Questions

None.
