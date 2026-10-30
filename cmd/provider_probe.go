package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/belingud/gptcomet/internal/config"
	gptcometerrors "github.com/belingud/gptcomet/internal/errors"
	"github.com/belingud/gptcomet/internal/factory"
	"github.com/belingud/gptcomet/internal/provider"
	"github.com/belingud/gptcomet/pkg/types"
	"github.com/spf13/cobra"
)

const (
	// providerProbePrompt is intentionally tiny so a probe costs as little as
	// possible while still exercising request formatting and response parsing.
	providerProbePrompt = "Reply with the single word: pong"
	// providerProbePreviewLimit caps the response preview shown on success.
	providerProbePreviewLimit = 160
)

// providerProber is the minimal client surface needed to probe a provider.
type providerProber interface {
	Chat(ctx context.Context, message string, history []types.Message) (*types.CompletionResponse, error)
}

// newProviderProber creates the API client used by `provider test`. It is a
// package variable so tests can inject a fake prober instead of hitting the
// network.
var newProviderProber = func(cfg *types.ClientConfig) (providerProber, error) {
	return factory.NewAPIClient(cfg)
}

// providerTestResult is the diagnostic output of one provider probe.
type providerTestResult struct {
	Info         provider.Info
	ClientConfig *types.ClientConfig
	Latency      time.Duration
	Response     string
}

// probeProvider sends one minimal request and measures its latency.
func probeProvider(ctx context.Context, prober providerProber) (time.Duration, string, error) {
	start := time.Now()
	response, err := prober.Chat(ctx, providerProbePrompt, nil)
	latency := time.Since(start)

	content := ""
	if response != nil {
		content = response.Content
	}
	return latency, content, err
}

// runProviderTest resolves the named provider's saved configuration, sends one
// request without retries, and returns the partial result even on failure so
// callers can report elapsed latency.
func runProviderTest(
	ctx context.Context,
	cfgManager config.ManagerInterface,
	svc *provider.Service,
	name string,
	onResolved func(*providerTestResult),
) (*providerTestResult, error) {
	info := svc.Get(name)
	if !info.Configured {
		if !info.Registered {
			return nil, gptcometerrors.ProviderNotFoundError(name)
		}
		return nil, gptcometerrors.ProviderNotConfiguredError(name)
	}

	clientConfig, err := cfgManager.GetClientConfig(name)
	if err != nil {
		return nil, err
	}

	// A diagnostic probe should fail fast instead of retrying with backoff.
	clientConfig.Retries = 1

	prober, err := newProviderProber(clientConfig)
	if err != nil {
		return nil, err
	}

	result := &providerTestResult{
		Info:         info,
		ClientConfig: clientConfig,
	}
	if onResolved != nil {
		onResolved(result)
	}

	latency, response, probeErr := probeProvider(ctx, prober)
	result.Latency = latency
	result.Response = response
	if probeErr != nil {
		return result, gptcometerrors.ProviderTestError(name, latency, probeErr)
	}

	return result, nil
}

// newProviderTestCmd creates the `provider test` command.
func newProviderTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test [provider]",
		Short: "Send one minimal request to check connectivity, auth, and latency",
		Long: "Send one minimal request to the named provider (or the active provider) " +
			"using the saved configuration, then report latency and the response. " +
			"The probe does not retry and does not require a Git repository.",
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := providerConfigManager(cmd)
			if err != nil {
				return err
			}

			svc := provider.NewService(cfgManager)
			name := svc.CurrentName()
			if len(args) > 0 {
				name = strings.TrimSpace(args[0])
			}
			name = svc.NormalizeName(name)
			if name == "" {
				return gptcometerrors.NoProviderSelectedError()
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Testing provider '%s'...\n", name)

			result, err := runProviderTest(cmd.Context(), cfgManager, svc, name, func(result *providerTestResult) {
				printProviderTestHeader(out, result)
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "%s Connected in %s\n", successStyle.Render("✔"), result.Latency.Round(time.Millisecond))
			if preview := formatProbeResponse(result.Response); preview != "" {
				fmt.Fprintf(out, "  response: %s\n", preview)
			}
			return nil
		},
	}
}

// printProviderTestHeader shows the resolved settings that were probed. Model
// and API base come from the resolved client config (the runtime path), while
// the key source is the descriptive provider info.
func printProviderTestHeader(out io.Writer, result *providerTestResult) {
	model := result.ClientConfig.Model
	if model == "" {
		model = "—"
	}
	apiBase := result.ClientConfig.APIBase
	if apiBase == "" {
		apiBase = "—"
	}

	fmt.Fprintf(out, "  model:    %s\n", model)
	fmt.Fprintf(out, "  api base: %s\n", apiBase)
	fmt.Fprintf(out, "  api key:  %s\n", formatKeySource(result.Info))
}

// formatProbeResponse collapses whitespace and truncates the response for a
// single-line preview. Empty responses produce an empty string.
func formatProbeResponse(response string) string {
	trimmed := strings.Join(strings.Fields(response), " ")
	if trimmed == "" {
		return ""
	}

	runes := []rune(trimmed)
	if len(runes) > providerProbePreviewLimit {
		return fmt.Sprintf("%q", string(runes[:providerProbePreviewLimit])+"…")
	}
	return fmt.Sprintf("%q", trimmed)
}
