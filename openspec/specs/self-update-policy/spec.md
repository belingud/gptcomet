## Purpose

Define when GPTComet may use its built-in update command and when it must defer to the installation manager.

## Requirements

### Requirement: Built-in update is blocked for Homebrew-managed installs

The system SHALL reject the built-in `gmsg update` flow when the running installation is managed by Homebrew.

#### Scenario: Homebrew install runs update

- **WHEN** a user runs `gmsg update` from a Homebrew-managed installation
- **THEN** the command exits with an error before checking GitHub releases
- **AND** the message tells the user to update with `brew upgrade gptcomet`

#### Scenario: Standalone install runs update

- **WHEN** a user runs `gmsg update` from a non-Homebrew installation
- **THEN** the command follows the existing GitHub release update flow

### Requirement: Homebrew-managed installs are detected deterministically

The system SHALL identify Homebrew-managed installations using an explicit installation-source marker, with resolved executable path detection as a fallback for Homebrew Cellar layouts.

#### Scenario: Homebrew marker is present

- **WHEN** the binary reports its installation source as Homebrew
- **THEN** the update guard treats the installation as Homebrew-managed

#### Scenario: Executable resolves inside Homebrew Cellar

- **WHEN** the running executable resolves to a project-specific Homebrew Cellar path
- **THEN** the update guard treats the installation as Homebrew-managed

### Requirement: Rejected Homebrew updates have no update side effects

The system MUST NOT perform update side effects after it determines the installation is Homebrew-managed.

#### Scenario: Homebrew update guard stops early

- **WHEN** the update guard rejects a Homebrew-managed installation
- **THEN** no GitHub release request is sent
- **AND** no archive is downloaded, extracted, copied, renamed, or symlinked
