package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveContentError(t *testing.T) {
	fullSecret := "ghp_" + "5xJ8kLm2Np4Qr6St8Uv0Wx2Yz4Ab6Cd8Ef0G"
	findings := []SensitiveFinding{
		{Path: ".env", Line: 3, RuleID: "sensitive-env-file", Preview: "(file-level)"},
		{Path: "cmd/deploy.go", Line: 12, RuleID: "github-pat", Preview: "ghp_…Ef0G"},
	}

	err := SensitiveContentError(findings)

	assert.Equal(t, ErrTypeValidation, err.Type)
	assert.Equal(t, ErrTitleSensitiveData, err.Title)
	assert.Contains(t, err.Message, ".env:3")
	assert.Contains(t, err.Message, "cmd/deploy.go:12")
	assert.Contains(t, err.Message, "github-pat")
	assert.Contains(t, err.Message, "2 potential secret(s)")
	assert.Equal(t, SensitiveScanDocsURL, err.DocsURL)

	rendered := err.Error()
	assert.Contains(t, rendered, "Sensitive Data Detected")
	assert.Contains(t, rendered, "--secrets off")
	assert.Contains(t, rendered, "git restore --staged")
	assert.Contains(t, rendered, "gmsg config set sensitive.mode redact")
	assert.NotContains(t, rendered, fullSecret)
}

func TestSensitiveContentErrorCapsFindingList(t *testing.T) {
	findings := make([]SensitiveFinding, 12)
	for i := range findings {
		findings[i] = SensitiveFinding{Path: "file.go", Line: i + 1, RuleID: "generic-api-key", Preview: "***"}
	}

	err := SensitiveContentError(findings)
	assert.Contains(t, err.Message, "… and 2 more")
	assert.Equal(t, 10, strings.Count(err.Message, "•"))
}

func TestInvalidSensitiveModeError(t *testing.T) {
	err := InvalidSensitiveModeError("loud")

	assert.Equal(t, ErrTypeValidation, err.Type)
	assert.Equal(t, ErrTitleInvalidSensitiveMode, err.Title)
	assert.Contains(t, err.Message, "loud")
	assert.Contains(t, err.Message, "off, warn, redact or block")
	assert.Equal(t, SensitiveScanDocsURL, err.DocsURL)
	assert.Contains(t, err.Error(), "gptcomet config set sensitive.mode")
}

func TestInvalidSensitiveAllowError(t *testing.T) {
	cause := errors.New("error parsing regexp: missing closing )")
	err := InvalidSensitiveAllowError("(", cause)

	assert.Equal(t, ErrTypeValidation, err.Type)
	assert.Equal(t, ErrTitleInvalidSensitiveAllow, err.Title)
	assert.Contains(t, err.Message, "(")
	require.ErrorIs(t, err, cause)
	assert.Equal(t, SensitiveScanDocsURL, err.DocsURL)
}
