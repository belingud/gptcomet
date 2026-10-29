package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "simple", value: "gptcomet", want: "'gptcomet'"},
		{name: "spaces", value: "/tmp/my config.yaml", want: "'/tmp/my config.yaml'"},
		{name: "single quote", value: "it's", want: `'it'\''s'`},
		{name: "empty", value: "", want: "''"},
		{name: "shell variable", value: "$HOME", want: "'$HOME'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellQuote(tt.value))
		})
	}
}

func TestBuildScript(t *testing.T) {
	script := buildScript("/tmp/my config.yaml")
	assert.Contains(t, script, ManagedMarker)
	assert.Contains(t, script, "config='/tmp/my config.yaml'")
	assert.Contains(t, script, "hook run \"$@\"")
	assert.Contains(t, script, "GMSG_BIN")
	assert.Contains(t, script, "GMSG_CONFIG")

	empty := buildScript("")
	assert.Contains(t, empty, "config=''")
}

func TestInstalledScriptInvokesHookRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell hook execution is not portable to Windows")
	}
	shPath, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hookPath, err := NewManager(nil).Install(InstallOptions{RepoPath: repo})
	require.NoError(t, err)

	binDir := t.TempDir()
	argsFile := filepath.Join(t.TempDir(), "args")
	fakeGmsg := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GMSG_TEST_ARGS\"\nexit 0\n"
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "gmsg"), []byte(fakeGmsg), 0o755))

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GMSG_TEST_ARGS", argsFile)
	t.Setenv("GMSG_BIN", "")
	t.Setenv("GMSG_CONFIG", "")
	t.Setenv("GMSG_HOOK_DEBUG", "")

	messageFile := filepath.Join(repo, "COMMIT_EDITMSG")
	require.NoError(t, os.WriteFile(messageFile, []byte(""), 0o644))

	cmd := exec.Command(shPath, hookPath, messageFile, "message")
	cmd.Dir = repo
	require.NoError(t, cmd.Run())

	args, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(args)), "\n")
	assert.Equal(t, []string{"hook", "run", messageFile, "message"}, lines)
}

func TestInstalledScriptFailsOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell hook execution is not portable to Windows")
	}
	shPath, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hookPath, err := NewManager(nil).Install(InstallOptions{RepoPath: repo})
	require.NoError(t, err)

	binDir := t.TempDir()
	fakeGmsg := "#!/bin/sh\nexit 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(binDir, "gmsg"), []byte(fakeGmsg), 0o755))

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GMSG_BIN", "")
	t.Setenv("GMSG_CONFIG", "")
	t.Setenv("GMSG_HOOK_DEBUG", "")

	messageFile := filepath.Join(repo, "COMMIT_EDITMSG")
	require.NoError(t, os.WriteFile(messageFile, []byte(""), 0o644))

	cmd := exec.Command(shPath, hookPath, messageFile)
	cmd.Dir = repo
	require.NoError(t, cmd.Run())
}

func TestInstalledScriptMissingBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell hook execution is not portable to Windows")
	}
	shPath, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not available")
	}

	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	hookPath, err := NewManager(nil).Install(InstallOptions{RepoPath: repo})
	require.NoError(t, err)

	t.Setenv("PATH", t.TempDir())
	t.Setenv("GMSG_BIN", "")
	t.Setenv("GMSG_CONFIG", "")
	t.Setenv("GMSG_HOOK_DEBUG", "")

	cmd := exec.Command(shPath, hookPath, filepath.Join(repo, "COMMIT_EDITMSG"))
	cmd.Dir = repo
	require.NoError(t, cmd.Run())
}
