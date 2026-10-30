package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"
	"github.com/belingud/gptcomet/internal/llm"
	"github.com/belingud/gptcomet/internal/ui"
	"github.com/belingud/gptcomet/pkg/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// providerSetupDeprecationNotice is printed by the legacy newprovider command
// on every invocation. The command keeps working until it is removed in a
// future release in favor of `gmsg provider new`.
const providerSetupDeprecationNotice = "⚠ 'gmsg newprovider' is deprecated and will be removed in a future release. Use 'gmsg provider new' instead."

// NewProviderSetupCmd returns the legacy `newprovider` command. It stays
// visible and functional for now, but prints a deprecation hint pointing at
// `gmsg provider new`.
func NewProviderSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "newprovider",
		Short: "Configure a new provider interactively (deprecated)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), formatRemindMessage(providerSetupDeprecationNotice))
			return runProviderSetup(cmd)
		},
	}
}

// newProviderNewCmd creates the `provider new` subcommand.
func newProviderNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Configure a new provider interactively",
		Long: "Configure a new provider interactively: choose a provider, fill in its " +
			"settings, and save them to gptcomet.yaml.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProviderSetup(cmd)
		},
	}
}

// runProviderSetup is the interactive provider configuration flow shared by
// `gmsg provider new` and the legacy `gmsg newprovider`.
func runProviderSetup(cmd *cobra.Command) error {
	// get providers list
	providers := llm.GetProviders()

	// In test environment, skip interactive selection
	if os.Getenv("GPTCOMET_TEST") == "1" {
		// Just list providers
		fmt.Fprintln(cmd.OutOrStdout(), "Available providers:")
		for _, p := range providers {
			fmt.Fprintln(cmd.OutOrStdout(), "-", p)
		}
		return nil
	}

	// Create and run provider selector
	selector := ui.NewProviderSelector(providers)
	p := tea.NewProgram(selector)
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run provider selector: %w", err)
	}

	providerName := m.(*ui.ProviderSelector).Selected()
	if providerName == "" {
		return nil
	}

	// Create provider instance with config
	provider, err := llm.NewProvider(providerName, &types.ClientConfig{})
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Get required config with default values
	requiredConfig := provider.GetRequiredConfig()
	debug.Printf("Required config: %v", requiredConfig)

	// Create and run config input
	configInput := ui.NewConfigInput(requiredConfig)
	p = tea.NewProgram(configInput)
	m, err = p.Run()
	if err != nil {
		return fmt.Errorf("failed to run config input: %w", err)
	}

	model2 := m.(*ui.ConfigInput)
	if !model2.Done() {
		return fmt.Errorf("configuration cancelled")
	}

	// Get the config values
	configs := model2.GetConfigs()
	debug.Printf("Config values: %v", configs)

	// Get config path from root command
	configPath, err := cmd.Root().PersistentFlags().GetString("config")
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create config manager
	cfgManager, err := config.New(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Check if provider config already exists
	existingConfig, _ := cfgManager.Get(providerName)
	if existingConfig != nil {
		fmt.Printf("Provider %s already has a configuration. Do you want to overwrite it? (y/N): ", providerName)

		// Use term package to read input
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to set terminal to raw mode: %w", err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		reader := term.NewTerminal(os.Stdin, "")
		response, err := reader.ReadLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		response = strings.TrimSpace(response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Configuration cancelled.")
			return nil
		}
	}

	// Update config
	if err := cfgManager.UpdateProviderConfig(providerName, configs); err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}

	// Update new provider
	cfgManager.Set("provider", providerName)

	fmt.Printf("\nProvider %s configured successfully!\n", providerName)
	return nil
}
