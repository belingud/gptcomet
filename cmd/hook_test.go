package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHookCmd(t *testing.T) {
	cmd := NewHookCmd()
	assert.Equal(t, "hook", cmd.Use)

	names := make(map[string]bool)
	for _, child := range cmd.Commands() {
		names[child.Name()] = true
	}
	for _, want := range []string{"install", "uninstall", "status", "run"} {
		assert.True(t, names[want], "subcommand %q should exist", want)
	}

	run, _, err := cmd.Find([]string{"run"})
	require.NoError(t, err)
	assert.True(t, run.Hidden)

	install, _, err := cmd.Find([]string{"install"})
	require.NoError(t, err)
	assert.NotNil(t, install.Flags().Lookup("force"))
	assert.NotNil(t, install.Flags().Lookup("repo"))
}

func TestHookCommandsLifecycle(t *testing.T) {
	repo, cleanup := testutils.TestGitRepo(t)
	defer cleanup()

	root := &cobra.Command{Use: "gmsg"}
	root.PersistentFlags().String("config", "", "Config file path")

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("provider: openai\nopenai:\n  api_key: test-key\n"), 0o600))

	hookCmd := NewHookCmd()
	var output bytes.Buffer
	hookCmd.SetOut(&output)
	root.AddCommand(hookCmd)

	root.SetArgs([]string{"--config", configPath, "hook", "install", "--repo", repo})
	require.NoError(t, root.Execute())
	assert.Contains(t, output.String(), "Installed prepare-commit-msg hook")

	output.Reset()
	root.SetArgs([]string{"--config", configPath, "hook", "status", "--repo", repo})
	require.NoError(t, root.Execute())
	assert.Contains(t, output.String(), "installed")
	assert.Contains(t, output.String(), "Backup: none")

	messageFile := filepath.Join(repo, "COMMIT_EDITMSG")
	require.NoError(t, os.WriteFile(messageFile, nil, 0o644))
	output.Reset()
	root.SetArgs([]string{"--config", configPath, "hook", "run", messageFile, "--repo", repo})
	require.NoError(t, root.Execute())
	message, err := os.ReadFile(messageFile)
	require.NoError(t, err)
	assert.Empty(t, string(message))

	output.Reset()
	root.SetArgs([]string{"--config", configPath, "hook", "uninstall", "--repo", repo})
	require.NoError(t, root.Execute())
	assert.Contains(t, output.String(), "Removed managed hook")
}
