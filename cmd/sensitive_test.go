package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/belingud/gptcomet/internal/config"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/secretscan"
	"github.com/belingud/gptcomet/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testGitHubPAT = "ghp_" + "5xJ8kLm2Np4Qr6St8Uv0Wx2Yz4Ab6Cd8Ef0G"

func sensitiveTestDiff(path string, lines ...string) string {
	var builder strings.Builder
	builder.WriteString("diff --git a/" + path + " b/" + path + "\n")
	builder.WriteString("--- a/" + path + "\n")
	builder.WriteString("+++ b/" + path + "\n")
	builder.WriteString("@@ -0,0 +1," + strconv.Itoa(len(lines)) + " @@\n")
	for _, line := range lines {
		builder.WriteString("+" + line + "\n")
	}
	return builder.String()
}

func TestResolveSensitiveOptions(t *testing.T) {
	t.Run("default warn from config", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()
		cfg, err := config.New(configPath)
		require.NoError(t, err)

		options, err := resolveSensitiveOptions(cfg, "")
		require.NoError(t, err)
		assert.Equal(t, secretscan.ModeWarn, options.Mode)
		assert.Empty(t, options.Allow)
	})

	t.Run("configured block mode", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()
		cfg, err := config.New(configPath)
		require.NoError(t, err)
		require.NoError(t, cfg.Set("sensitive.mode", "block"))

		options, err := resolveSensitiveOptions(cfg, "")
		require.NoError(t, err)
		assert.Equal(t, secretscan.ModeBlock, options.Mode)
	})

	t.Run("flag override wins over config", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()
		cfg, err := config.New(configPath)
		require.NoError(t, err)
		require.NoError(t, cfg.Set("sensitive.mode", "block"))

		options, err := resolveSensitiveOptions(cfg, "redact")
		require.NoError(t, err)
		assert.Equal(t, secretscan.ModeRedact, options.Mode)
	})

	t.Run("nil config manager defaults to warn", func(t *testing.T) {
		options, err := resolveSensitiveOptions(nil, "")
		require.NoError(t, err)
		assert.Equal(t, secretscan.ModeWarn, options.Mode)
	})

	t.Run("invalid config mode returns template error", func(t *testing.T) {
		configPath, cleanup := setupTempConfig(t)
		defer cleanup()
		cfg, err := config.New(configPath)
		require.NoError(t, err)
		require.NoError(t, cfg.Set("sensitive.mode", "maybe"))

		_, err = resolveSensitiveOptions(cfg, "")
		require.Error(t, err)
		var gptErr *gptcometerrors.GPTCometError
		require.True(t, errors.As(err, &gptErr))
		assert.Equal(t, gptcometerrors.ErrTitleInvalidSensitiveMode, gptErr.Title)
	})

	t.Run("invalid flag mode returns template error", func(t *testing.T) {
		_, err := resolveSensitiveOptions(nil, "loud")
		require.Error(t, err)
		var gptErr *gptcometerrors.GPTCometError
		require.True(t, errors.As(err, &gptErr))
		assert.Equal(t, gptcometerrors.ErrTitleInvalidSensitiveMode, gptErr.Title)
	})

	t.Run("allow patterns are compiled", func(t *testing.T) {
		mockCfg := new(testutils.MockConfigManager)
		mockCfg.On("GetWithDefault", sensitiveModeKey, mock.Anything).Return("warn")
		mockCfg.On("GetWithDefault", sensitiveAllowKey, mock.Anything).Return([]interface{}{`^dummy-`, "EXAMPLE"})

		options, err := resolveSensitiveOptions(mockCfg, "")
		require.NoError(t, err)
		assert.Len(t, options.Allow, 2)
		mockCfg.AssertExpectations(t)
	})

	t.Run("invalid allow pattern returns template error", func(t *testing.T) {
		mockCfg := new(testutils.MockConfigManager)
		mockCfg.On("GetWithDefault", sensitiveModeKey, mock.Anything).Return("warn")
		mockCfg.On("GetWithDefault", sensitiveAllowKey, mock.Anything).Return([]string{"("})

		_, err := resolveSensitiveOptions(mockCfg, "")
		require.Error(t, err)
		var gptErr *gptcometerrors.GPTCometError
		require.True(t, errors.As(err, &gptErr))
		assert.Equal(t, gptcometerrors.ErrTitleInvalidSensitiveAllow, gptErr.Title)
	})
}

func TestPrintSensitiveFindings(t *testing.T) {
	findings := []secretscan.Finding{
		{RuleID: "github-pat", Path: "cmd/deploy.go", Line: 3, Secret: testGitHubPAT},
		{RuleID: "sensitive-env-file", Path: ".env", Line: 1, FileLevel: true},
	}

	tests := []struct {
		name     string
		mode     secretscan.Mode
		contains []string
	}{
		{
			name: "warn",
			mode: secretscan.ModeWarn,
			contains: []string{
				"Sensitive data detected (2 finding(s), mode: warn)",
				"cmd/deploy.go:3",
				"github-pat",
				"sensitive-env-file",
				"file-level",
				"redaction disabled",
			},
		},
		{
			name: "redact",
			mode: secretscan.ModeRedact,
			contains: []string{
				"mode: redact",
				"redacted before the request",
			},
		},
		{
			name:     "block",
			mode:     secretscan.ModeBlock,
			contains: []string{"mode: block", "Request blocked"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			printSensitiveFindings(&buffer, findings, tt.mode)
			output := buffer.String()
			for _, want := range tt.contains {
				assert.Contains(t, output, want)
			}
			assert.NotContains(t, output, testGitHubPAT, "warning must not print the full secret")
		})
	}

	t.Run("no findings", func(t *testing.T) {
		var buffer bytes.Buffer
		printSensitiveFindings(&buffer, nil, secretscan.ModeWarn)
		assert.Empty(t, buffer.String())
	})
}

func TestCommitServiceScansStagedDiff(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		override     string
		wantErr      bool
		wantRedacted bool
		wantWarning  bool
	}{
		{name: "warn keeps diff", mode: "warn", wantWarning: true},
		{name: "redact rewrites diff", mode: "redact", wantRedacted: true, wantWarning: true},
		{name: "block returns error", mode: "block", wantErr: true, wantWarning: true},
		{name: "off disables scan", mode: "off"},
		{name: "flag override redact", mode: "off", override: "redact", wantRedacted: true, wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, cleanup := setupTempConfig(t)
			defer cleanup()
			cfg, err := config.New(configPath)
			require.NoError(t, err)
			require.NoError(t, cfg.Set("sensitive.mode", tt.mode))

			diff := sensitiveTestDiff("cmd/deploy.go", testGitHubPAT)
			mockVCS := new(MockVCS)
			mockVCS.On("HasStagedChanges", mock.Anything).Return(true, nil)
			mockVCS.On("GetStagedDiffFiltered", mock.Anything, cfg).Return(diff, nil)

			var stderr bytes.Buffer
			service := &CommitService{
				vcs:        mockVCS,
				cfgManager: cfg,
				options:    CommitOptions{RepoPath: ".", SecretsMode: tt.override},
				stderr:     &stderr,
			}

			got, err := service.getStagedDiff()

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				var gptErr *gptcometerrors.GPTCometError
				require.True(t, errors.As(err, &gptErr))
				assert.Contains(t, err.Error(), "Sensitive Data Detected")
			} else {
				require.NoError(t, err)
				if tt.wantRedacted {
					assert.NotContains(t, got, testGitHubPAT)
					assert.Contains(t, got, "[REDACTED:")
				} else {
					assert.Equal(t, diff, got)
				}
			}

			if tt.wantWarning {
				assert.Contains(t, stderr.String(), "Sensitive data detected")
				assert.NotContains(t, stderr.String(), testGitHubPAT)
			} else {
				assert.Empty(t, stderr.String())
			}
			mockVCS.AssertExpectations(t)
		})
	}
}

func TestReviewServiceScansPipedDiff(t *testing.T) {
	configPath, cleanup := setupTempConfig(t)
	defer cleanup()
	cfg, err := config.New(configPath)
	require.NoError(t, err)
	require.NoError(t, cfg.Set("sensitive.mode", "redact"))

	service := &ReviewService{
		cfgManager: cfg,
		options:    ReviewOptions{RepoPath: "test-repo"},
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer reader.Close()

	var stderr bytes.Buffer
	service.stderr = &stderr

	go func() {
		_, _ = io.WriteString(writer, "token="+testGitHubPAT+"\n")
		_ = writer.Close()
	}()
	os.Stdin = reader

	diff, err := service.getDiff()
	require.NoError(t, err)
	assert.NotContains(t, diff, testGitHubPAT)
	assert.Contains(t, diff, "[REDACTED:")
	assert.NotContains(t, stderr.String(), testGitHubPAT)
}

func TestSensitiveFlagRegistered(t *testing.T) {
	commands := map[string]*cobra.Command{
		"commit":   NewCommitCmd(),
		"generate": NewGenerateCmd(),
		"review":   NewReviewCmd(),
	}
	for name, command := range commands {
		if command.Flags().Lookup("secrets") == nil {
			t.Errorf("command %q must register --secrets", name)
		}
	}
}
