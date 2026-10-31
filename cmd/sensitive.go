package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/belingud/gptcomet/internal/config"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/secretscan"
	"github.com/belingud/gptcomet/pkg/config/defaults"
)

const (
	sensitiveModeKey  = "sensitive.mode"
	sensitiveAllowKey = "sensitive.allow"
)

// resolveSensitiveOptions reads the sensitive.* configuration and applies the
// command-line override. A nil manager (services built directly in tests) uses
// the built-in default so scanning never silently disappears.
func resolveSensitiveOptions(cfgManager config.ManagerInterface, override string) (secretscan.Options, error) {
	modeValue := strings.TrimSpace(override)
	if modeValue == "" {
		modeValue = defaults.DefaultSensitiveMode
		if cfgManager != nil {
			raw := cfgManager.GetWithDefault(sensitiveModeKey, defaults.DefaultSensitiveMode)
			value, ok := raw.(string)
			if !ok {
				return secretscan.Options{}, gptcometerrors.InvalidSensitiveModeError(fmt.Sprintf("%v", raw))
			}
			modeValue = value
		}
	}

	mode, err := secretscan.ParseMode(modeValue)
	if err != nil {
		return secretscan.Options{}, gptcometerrors.InvalidSensitiveModeError(modeValue)
	}

	var allowPatterns []string
	if cfgManager != nil {
		allowPatterns = configStringSlice(cfgManager.GetWithDefault(sensitiveAllowKey, []string{}))
	}

	allow, err := secretscan.CompileAllow(allowPatterns)
	if err != nil {
		// CompileAllow reports the first bad pattern inside its message; resolve
		// the exact pattern so the template error can quote it.
		for _, pattern := range allowPatterns {
			if _, compileErr := secretscan.CompileAllow([]string{pattern}); compileErr != nil {
				return secretscan.Options{}, gptcometerrors.InvalidSensitiveAllowError(pattern, compileErr)
			}
		}
		return secretscan.Options{}, gptcometerrors.InvalidSensitiveAllowError(strings.Join(allowPatterns, ", "), err)
	}

	return secretscan.Options{Mode: mode, Allow: allow}, nil
}

// printSensitiveFindings writes the masked warning block to stderr. It is
// shared by the commit, generate, hook, and review paths so the format stays
// consistent. The writer may be nil in tests, in which case warnings are
// discarded.
func printSensitiveFindings(w io.Writer, findings []secretscan.Finding, mode secretscan.Mode) {
	if w == nil || len(findings) == 0 {
		return
	}

	fmt.Fprintf(w, "\n⚠ Sensitive data detected (%d finding(s), mode: %s):\n", len(findings), mode)
	limit := len(findings)
	const maxPrinted = 10
	if limit > maxPrinted {
		limit = maxPrinted
	}
	for _, finding := range findings[:limit] {
		location := finding.Path
		if finding.Line > 0 {
			location = fmt.Sprintf("%s:%d", finding.Path, finding.Line)
		}
		if finding.FileLevel {
			fmt.Fprintf(w, "  • %s  [%s]  file-level\n", location, finding.RuleID)
			continue
		}
		fmt.Fprintf(w, "  • %s  [%s]  %s\n", location, finding.RuleID, finding.Preview())
	}
	if len(findings) > limit {
		fmt.Fprintf(w, "  … and %d more\n", len(findings)-limit)
	}

	switch mode {
	case secretscan.ModeRedact:
		fmt.Fprintln(w, "  Findings were redacted before the request; your staged diff is unchanged.")
	case secretscan.ModeBlock:
		fmt.Fprintln(w, "  Request blocked. Unstage the file or adjust sensitive.mode to proceed.")
	default:
		fmt.Fprintln(w, "  Sending with redaction disabled; set sensitive.mode=redact or block to protect the request.")
	}
	fmt.Fprintln(w)
}

// orStderr returns w when set, otherwise os.Stderr. Services built directly in
// tests leave the writer nil.
func orStderr(w io.Writer) io.Writer {
	if w != nil {
		return w
	}
	return os.Stderr
}

func configStringSlice(raw interface{}) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []interface{}:
		result := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	case nil:
		return nil
	default:
		return nil
	}
}
