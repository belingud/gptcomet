package cmd

import (
	"bufio"
	"fmt"
	"gptcomet/internal/config"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func readMaskedInput(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // add newline
	return string(bytePassword), nil
}

// NewProviderCmd creates a new provider command
func NewProviderCmd() *cobra.Command {
	const (
		defaultProvider  = "openai"
		defaultAPIBase   = "https://api.openai.com/v1"
		defaultMaxTokens = 1024
		defaultModel     = "gpt-4"
	)

	cmd := &cobra.Command{
		Use:   "newprovider",
		Short: "Add a new API provider interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a reader for user input
			reader := bufio.NewReader(os.Stdin)

			// Get provider name
			fmt.Printf("Enter provider name [%s]: ", defaultProvider)
			provider, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read provider name: %w", err)
			}
			provider = strings.TrimSpace(provider)
			if provider == "" {
				provider = defaultProvider
			}

			// Get API base
			fmt.Printf("Enter API base URL [%s]: ", defaultAPIBase)
			apiBase, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read API base: %w", err)
			}
			apiBase = strings.TrimSpace(apiBase)
			if apiBase == "" {
				apiBase = defaultAPIBase
			}

			// Get API key (with masked input)
			apiKey, err := readMaskedInput("Enter API key: ")
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Get model
			fmt.Printf("Enter model name [%s]: ", defaultModel)
			model, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read model: %w", err)
			}
			model = strings.TrimSpace(model)
			if model == "" {
				model = defaultModel
			}

			// Get model max tokens
			fmt.Printf("Enter model max tokens [%d]: ", defaultMaxTokens)
			maxTokensStr, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read max tokens: %w", err)
			}
			maxTokensStr = strings.TrimSpace(maxTokensStr)
			maxTokens := defaultMaxTokens
			if maxTokensStr != "" {
				maxTokens, err = strconv.Atoi(maxTokensStr)
				if err != nil {
					return fmt.Errorf("invalid max tokens value: %w", err)
				}
			}

			// Create config manager
			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			// Check if provider already exists
			if _, exists := cfgManager.Get(provider); exists {
				fmt.Printf("Provider '%s' already exists. Do you want to overwrite it? [y/N]: ", provider)
				answer, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read answer: %w", err)
				}
				answer = strings.ToLower(strings.TrimSpace(answer))
				if answer != "y" && answer != "yes" {
					fmt.Println("Operation cancelled")
					return nil
				}
			}

			// Set the provider configuration
			providerConfig := map[string]interface{}{
				"api_key":          apiKey,
				"api_base":         apiBase,
				"model":            model,
				"model_max_tokens": maxTokens,
			}
			if err := cfgManager.Set(provider, providerConfig); err != nil {
				return fmt.Errorf("failed to set provider config: %w", err)
			}

			fmt.Printf("\nProvider configuration saved:\n")
			fmt.Printf("  Provider: %s\n", provider)
			fmt.Printf("  API Base: %s\n", apiBase)
			fmt.Printf("  Model: %s\n", model)
			fmt.Printf("  Max Tokens: %d\n", maxTokens)
			return nil
		},
	}

	return cmd
}
