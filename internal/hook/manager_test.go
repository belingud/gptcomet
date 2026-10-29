package hook

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/belingud/gptcomet/internal/git"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerInstall(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	configPath := filepath.Join(t.TempDir(), "my config.yaml")
	manager := NewManager(nil)

	hookPath, err := manager.Install(InstallOptions{RepoPath: repo, ConfigPath: configPath})
	require.NoError(t, err)
	assert.Equal(t, HookName, filepath.Base(hookPath))

	script, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Contains(t, string(script), ManagedMarker)
	assert.Contains(t, string(script), "config="+shellQuote(configPath))
	assert.Contains(t, string(script), "hook run \"$@\"")

	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(hookPath)
		require.NoError(t, statErr)
		assert.NotZero(t, info.Mode().Perm()&0o111)
	}

	firstScript := string(script)
	hookPathAgain, err := manager.Install(InstallOptions{RepoPath: repo, ConfigPath: configPath})
	require.NoError(t, err)
	assert.Equal(t, hookPath, hookPathAgain)
	scriptAgain, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, firstScript, string(scriptAgain))

	status, err := manager.Status(repo)
	require.NoError(t, err)
	assert.Equal(t, StateInstalled, status.State)
	assert.Equal(t, hookPath, status.HookPath)
	assert.Empty(t, status.BackupPath)
}

func TestManagerInstallForeignHook(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hooksDir, err := git.GetHooksDir(repo)
	require.NoError(t, err)
	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	foreignScript := "#!/bin/sh\necho foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(foreignScript), 0o755))

	manager := NewManager(nil)

	_, err = manager.Install(InstallOptions{RepoPath: repo})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	unchanged, readErr := os.ReadFile(hookPath)
	require.NoError(t, readErr)
	assert.Equal(t, foreignScript, string(unchanged))

	_, err = manager.Install(InstallOptions{RepoPath: repo, Force: true})
	require.NoError(t, err)

	backup, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, foreignScript, string(backup))

	status, err := manager.Status(repo)
	require.NoError(t, err)
	assert.Equal(t, StateInstalled, status.State)
	assert.Equal(t, backupPath, status.BackupPath)

	result, err := manager.Uninstall(repo)
	require.NoError(t, err)
	assert.True(t, result.Removed)
	assert.True(t, result.Restored)

	restored, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, foreignScript, string(restored))
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err))
}

func TestManagerInstallBackupAlreadyExists(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hooksDir, err := git.GetHooksDir(repo)
	require.NoError(t, err)
	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))
	require.NoError(t, os.WriteFile(backupPath, []byte("old backup\n"), 0o644))

	manager := NewManager(nil)
	_, err = manager.Install(InstallOptions{RepoPath: repo, Force: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "backup")

	current, readErr := os.ReadFile(hookPath)
	require.NoError(t, readErr)
	assert.Equal(t, "#!/bin/sh\necho foreign\n", string(current))
}

func TestManagerUninstallForeignHook(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hooksDir, err := git.GetHooksDir(repo)
	require.NoError(t, err)
	hookPath := filepath.Join(hooksDir, HookName)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	result, err := NewManager(nil).Uninstall(repo)
	require.NoError(t, err)
	assert.False(t, result.Removed)
	assert.False(t, result.Restored)

	current, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/sh\necho foreign\n", string(current))
}

func TestManagerUninstallMissingHook(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	result, err := NewManager(nil).Uninstall(repo)
	require.NoError(t, err)
	assert.False(t, result.Removed)
	assert.False(t, result.Restored)
}

func TestManagerUninstallRestoresBackupWithoutHook(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hooksDir, err := git.GetHooksDir(repo)
	require.NoError(t, err)
	hookPath := filepath.Join(hooksDir, HookName)
	backupPath := hookPath + BackupSuffix
	backupContents := "#!/bin/sh\necho backup\n"
	require.NoError(t, os.WriteFile(backupPath, []byte(backupContents), 0o755))

	result, err := NewManager(nil).Uninstall(repo)
	require.NoError(t, err)
	assert.False(t, result.Removed)
	assert.True(t, result.Restored)

	restored, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, backupContents, string(restored))
}

func TestManagerStatusStates(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	manager := NewManager(nil)
	status, err := manager.Status(repo)
	require.NoError(t, err)
	assert.Equal(t, StateMissing, status.State)

	hooksDir, err := git.GetHooksDir(repo)
	require.NoError(t, err)
	hookPath := filepath.Join(hooksDir, HookName)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	status, err = manager.Status(repo)
	require.NoError(t, err)
	assert.Equal(t, StateForeign, status.State)

	require.NoError(t, os.Remove(hookPath))
	_, err = manager.Install(InstallOptions{RepoPath: repo})
	require.NoError(t, err)

	status, err = manager.Status(repo)
	require.NoError(t, err)
	assert.Equal(t, StateInstalled, status.State)
}
