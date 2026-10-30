package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/provider"
	"github.com/spf13/cobra"
)

// NewProviderCmd creates the provider command group. It exposes the provider
// inventory, switches the default provider, and probes a provider with one
// minimal request. Interactive configuration stays in `gmsg newprovider`.
func NewProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Inspect, switch, and test LLM providers",
		Long: "Inspect the active LLM provider, switch the default provider, and test " +
			"connectivity with a minimal request.",
	}

	cmd.AddCommand(
		newProviderListCmd(),
		newProviderUseCmd(),
		newProviderTestCmd(),
	)

	return cmd
}

// providerConfigManager returns the config manager from the Cobra context when
// a command (or a test) provided one, otherwise it builds one from the root
// --config flag.
func providerConfigManager(cmd *cobra.Command) (config.ManagerInterface, error) {
	if manager, err := getConfigManager(cmd); err == nil {
		if full, ok := manager.(config.ManagerInterface); ok {
			return full, nil
		}
	}

	configPath := ""
	if root := cmd.Root(); root != nil {
		if path, err := root.PersistentFlags().GetString("config"); err == nil {
			configPath = path
		}
	}

	manager, err := config.New(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}
	return manager, nil
}

// newProviderListCmd creates the `provider list` command.
func newProviderListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List providers with their model and API key source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := providerConfigManager(cmd)
			if err != nil {
				return err
			}

			renderProviderList(cmd.OutOrStdout(), provider.NewService(cfgManager))
			return nil
		},
	}
}

// newProviderUseCmd creates the `provider use` command.
func newProviderUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch the default provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgManager, err := providerConfigManager(cmd)
			if err != nil {
				return err
			}

			result, err := provider.NewService(cfgManager).Use(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if result.AlreadyActive {
				fmt.Fprintf(out, "Provider '%s' is already active.\n", result.Provider)
			} else {
				fmt.Fprintln(out, formatBoxedMessage(fmt.Sprintf("Active provider: %s", result.Provider)))
			}

			if result.Seeded {
				fmt.Fprintf(out, "Created default configuration for '%s'.\n", result.Provider)
			}
			if !result.Registered {
				fmt.Fprintln(out, formatRemindMessage(fmt.Sprintf(
					"⚠ '%s' is a custom provider; gmsg uses the OpenAI-compatible fallback implementation.",
					result.Provider)))
			}
			if result.KeySource == config.APIKeySourceMissing {
				fmt.Fprintln(out, formatRemindMessage(fmt.Sprintf(
					"⚠ API key for '%s' is not configured. Set it with 'gmsg config set %s.api_key <key>' or export %s.",
					result.Provider, result.Provider, result.KeyEnvVar)))
			}
			if len(result.MissingFields) > 0 {
				fmt.Fprintln(out, formatRemindMessage(fmt.Sprintf(
					"⚠ Provider '%s' still needs: %s. Configure it with 'gmsg newprovider' or 'gmsg config set %s.<field> <value>'.",
					result.Provider, strings.Join(result.MissingFields, ", "), result.Provider)))
			}

			return nil
		},
	}
}

// renderProviderList writes the human-readable provider inventory.
func renderProviderList(w io.Writer, svc *provider.Service) {
	infos := svc.List()

	currentName := svc.CurrentName()
	if currentName == "" {
		fmt.Fprintln(w, "Current provider: (not set)")
	} else {
		fmt.Fprintf(w, "Current provider: %s\n", currentName)
		for _, info := range infos {
			if !info.Active {
				continue
			}
			fmt.Fprintf(w, "  model:    %s\n", formatModel(info))
			fmt.Fprintf(w, "  api base: %s\n", formatAPIBase(info))
			fmt.Fprintf(w, "  api key:  %s\n", formatKeySource(info))
			break
		}
	}
	fmt.Fprintln(w)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PROVIDER\tMODEL\tAPI KEY")
	for _, info := range infos {
		marker := " "
		if info.Active {
			marker = "*"
		}
		model := info.Model
		if model == "" {
			model = "—"
		}
		fmt.Fprintf(tw, "%s %s\t%s\t%s\n", marker, info.Name, model, formatKeySource(info))
	}
	_ = tw.Flush()
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Switch providers with 'gmsg provider use <name>'.")
	fmt.Fprintln(w, "Set a missing API key with 'gmsg config set <provider>.api_key <key>' or the provider's environment variable (e.g. OPENAI_API_KEY).")
	fmt.Fprintln(w, "Configure a provider interactively with 'gmsg newprovider'.")
}

// formatKeySource renders where a provider's API key comes from.
func formatKeySource(info provider.Info) string {
	switch info.KeySource {
	case config.APIKeySourceConfig:
		if info.KeyMasked == "" {
			return "config"
		}
		return fmt.Sprintf("config (%s)", info.KeyMasked)
	case config.APIKeySourceEnv:
		return fmt.Sprintf("env %s", info.KeyEnvVar)
	case config.APIKeySourceNotRequired:
		return "not required"
	default:
		return "missing"
	}
}

// formatModel renders the effective model and whether it came from the config.
func formatModel(info provider.Info) string {
	if info.Model == "" {
		return "—"
	}
	if info.ModelFromConfig {
		return fmt.Sprintf("%s (config)", info.Model)
	}
	return fmt.Sprintf("%s (default)", info.Model)
}

// formatAPIBase renders the effective API base and whether it came from the config.
func formatAPIBase(info provider.Info) string {
	if info.APIBase == "" {
		return "—"
	}
	if info.APIBaseFromConfig {
		return fmt.Sprintf("%s (config)", info.APIBase)
	}
	return fmt.Sprintf("%s (default)", info.APIBase)
}
