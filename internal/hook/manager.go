// Package hook manages gptcomet's Git prepare-commit-msg integration.
package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	gptErrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/git"
)

// GenerateFunc produces a commit message for the given repository and config.
// It is injected by cmd so this package does not depend on the HTTP client.
type GenerateFunc func(ctx context.Context, repoPath, configPath string) (string, error)

// Manager handles installing, inspecting, and running the managed hook.
type Manager struct {
	Generate GenerateFunc
}

// NewManager creates a hook manager.
func NewManager(generate GenerateFunc) *Manager {
	return &Manager{Generate: generate}
}

// InstallOptions controls hook installation.
type InstallOptions struct {
	RepoPath   string
	ConfigPath string
	Force      bool
}

// HookState describes the current state of the hook file.
type HookState string

const (
	StateMissing   HookState = "missing"
	StateInstalled HookState = "installed"
	StateForeign   HookState = "foreign"
)

// Status reports the resolved paths and current hook state.
type Status struct {
	HooksDir   string
	HookPath   string
	State      HookState
	BackupPath string
}

// UninstallResult reports what uninstall changed.
type UninstallResult struct {
	HookPath string
	Removed  bool
	Restored bool
}

// Install writes an executable prepare-commit-msg hook. It is idempotent for a
// hook already managed by gptcomet, and refuses to overwrite a foreign hook
// unless Force is set.
func (m *Manager) Install(opts InstallOptions) (string, error) {
	configPath := strings.TrimSpace(opts.ConfigPath)
	if configPath != "" {
		abs, absErr := filepath.Abs(configPath)
		if absErr != nil {
			return "", gptErrors.HookInstallError(configPath, absErr)
		}
		configPath = abs
	}

	hooksDir, err := git.GetHooksDir(opts.RepoPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return "", gptErrors.HookInstallError(hooksDir, err)
	}

	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	backupCreated := false

	if info, err := os.Stat(hookPath); err == nil {
		data, readErr := os.ReadFile(hookPath)
		if readErr != nil {
			return "", gptErrors.HookInstallError(hookPath, readErr)
		}
		if !isManagedHook(string(data)) {
			if !opts.Force {
				return "", gptErrors.HookAlreadyExistsError(hookPath, backupPath)
			}
			if _, backupErr := os.Stat(backupPath); backupErr == nil {
				return "", gptErrors.HookBackupExistsError(backupPath)
			} else if !os.IsNotExist(backupErr) {
				return "", gptErrors.HookInstallError(backupPath, backupErr)
			}
			if backupErr := os.WriteFile(backupPath, data, info.Mode().Perm()); backupErr != nil {
				return "", gptErrors.HookInstallError(backupPath, backupErr)
			}
			backupCreated = true
		}
	} else if !os.IsNotExist(err) {
		return "", gptErrors.HookInstallError(hookPath, err)
	}

	if err := os.WriteFile(hookPath, []byte(buildScript(configPath)), 0o755); err != nil {
		if backupCreated {
			_ = os.Remove(backupPath)
		}
		return "", gptErrors.HookInstallError(hookPath, err)
	}
	// WriteFile applies the mode only when creating the file; make sure an
	// existing managed hook is executable too.
	if err := os.Chmod(hookPath, 0o755); err != nil {
		return "", gptErrors.HookInstallError(hookPath, err)
	}

	return hookPath, nil
}

// Uninstall removes a gptcomet-managed hook and restores a --force backup.
// Foreign hooks are never removed.
func (m *Manager) Uninstall(repoPath string) (UninstallResult, error) {
	hooksDir, err := git.GetHooksDir(repoPath)
	if err != nil {
		return UninstallResult{}, err
	}

	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	result := UninstallResult{HookPath: hookPath}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return result, gptErrors.HookUninstallError(hookPath, err)
		}
		return m.restoreBackup(result, backupPath)
	}

	if !isManagedHook(string(data)) {
		return result, nil
	}

	if err := os.Remove(hookPath); err != nil {
		return result, gptErrors.HookUninstallError(hookPath, err)
	}
	result.Removed = true

	return m.restoreBackup(result, backupPath)
}

func (m *Manager) restoreBackup(result UninstallResult, backupPath string) (UninstallResult, error) {
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Rename(backupPath, result.HookPath); err != nil {
			return result, gptErrors.HookUninstallError(result.HookPath, err)
		}
		result.Restored = true
	} else if !os.IsNotExist(err) {
		return result, gptErrors.HookUninstallError(backupPath, err)
	}
	return result, nil
}

// Status inspects the hook file without modifying it.
func (m *Manager) Status(repoPath string) (Status, error) {
	hooksDir, err := git.GetHooksDir(repoPath)
	if err != nil {
		return Status{}, err
	}

	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	status := Status{
		HooksDir: hooksDir,
		HookPath: hookPath,
		State:    StateMissing,
	}

	if _, err := os.Stat(backupPath); err == nil {
		status.BackupPath = backupPath
	} else if !os.IsNotExist(err) {
		return status, gptErrors.HookStatusError(hookPath, err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return status, gptErrors.HookStatusError(hookPath, err)
	}

	if isManagedHook(string(data)) {
		status.State = StateInstalled
	} else {
		status.State = StateForeign
	}

	return status, nil
}
