package cmd

import (
	"bytes"
	"testing"

	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCommitMessageGenerator is a commitMessageGenerator test double.
type fakeCommitMessageGenerator struct {
	message string
	err     error
	calls   int
}

func (f *fakeCommitMessageGenerator) GenerateMessage() (string, error) {
	f.calls++
	return f.message, f.err
}

func TestNewGenerateCmd(t *testing.T) {
	cmd := NewGenerateCmd()
	assert.Equal(t, "generate", cmd.Use)

	for _, name := range []string{"repo", "rich", "secrets", "model", "api-base", "timeout"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag %q should be registered", name)
	}
}

// newTestRoot wires the generate command under a root that owns the persistent
// --config flag, mirroring main.go closely enough for RunE.
func newTestRoot(t *testing.T, generator commitMessageGenerator) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	original := newGenerateService
	newGenerateService = func(CommitOptions) (commitMessageGenerator, error) {
		return generator, nil
	}
	t.Cleanup(func() { newGenerateService = original })

	root := &cobra.Command{Use: "gmsg", SilenceUsage: true}
	root.PersistentFlags().String("config", "", "")
	root.AddCommand(NewGenerateCmd())

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"generate"})
	return root, &stdout, &stderr
}

func TestGenerateCmdWritesOnlyMessageToStdout(t *testing.T) {
	fake := &fakeCommitMessageGenerator{message: "feat: add programmable output"}
	root, stdout, _ := newTestRoot(t, fake)

	require.NoError(t, root.Execute())
	assert.Equal(t, "feat: add programmable output\n", stdout.String())
	assert.Equal(t, 1, fake.calls)
}

func TestGenerateCmdPropagatesClassifiedError(t *testing.T) {
	fake := &fakeCommitMessageGenerator{
		err: gptcometerrors.NewConfigError("Configuration Error", "provider not set", nil, nil),
	}
	root, stdout, stderr := newTestRoot(t, fake)

	err := root.Execute()
	require.Error(t, err)
	assert.Empty(t, stdout.String())
	assert.Equal(t, gptcometerrors.ExitConfig, gptcometerrors.ExitCode(err))
	// Cobra prints the error to stderr but never to stdout.
	assert.Contains(t, stderr.String(), "Error:")
}
