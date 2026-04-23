## 1. Tests

- [x] 1.1 Add unit tests for Homebrew installation detection via explicit installation-source marker.
- [x] 1.2 Add unit tests for Homebrew installation detection via resolved executable paths under a project-specific Homebrew Cellar directory.
- [x] 1.3 Add an update command test proving Homebrew-managed installs return an error before any GitHub release request is made.
- [x] 1.4 Add a regression test proving non-Homebrew installs still use the existing GitHub release update flow.

## 2. Implementation

- [x] 2.1 Add a small installation-source marker in the Go update command package, with a standalone default and a Homebrew value supported by linker flags.
- [x] 2.2 Add a testable Homebrew detection helper that checks the marker first and resolved executable path second.
- [x] 2.3 Wire the Homebrew guard into `gmsg update` before the command calls the GitHub release API.
- [x] 2.4 Return an actionable error message that tells Homebrew users to run `brew upgrade gptcomet`.

## 3. Verification

- [x] 3.1 Run Go formatting for the touched Go files.
- [x] 3.2 Run the update command test suite.
- [x] 3.3 Run the full Go test suite if the focused update tests pass.
