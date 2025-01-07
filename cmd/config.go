package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/belingud/go-gptcomet/internal/config"
	"github.com/belingud/go-gptcomet/internal/debug"

	"github.com/spf13/cobra"
)

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	Get(key string) (interface{}, bool)
	List() (string, error)
	Reset(promptOnly bool) error
	Set(key string, value interface{}) error
	GetPath() string
	Remove(key string, value string) error
	Append(key string, value interface{}) error
	GetSupportedKeys() []string
}

// configKey is used as the key type for context
type configKey struct{}

// setConfigManager sets the ConfigManager for the given command.
// The ConfigManager is stored in the command's context and can be retrieved
// using getConfigManager.
func setConfigManager(cmd *cobra.Command, manager ConfigManager) {
	cmd.SetContext(context.WithValue(cmd.Context(), configKey{}, manager))
}

// getConfigManager traverses up the cobra command hierarchy to find and return a ConfigManager
// from the command's context. It starts from the given command and checks each parent command
// until it finds a ConfigManager or reaches the root.
//
// Parameters:
//   - cmd: A pointer to the cobra.Command from which to start the search
//
// Returns:
//   - ConfigManager: The found configuration manager instance
//   - error: An error if no ConfigManager is found in the command hierarchy
func getConfigManager(cmd *cobra.Command) (ConfigManager, error) {
	current := cmd
	for current != nil {
		if manager, ok := current.Context().Value(configKey{}).(ConfigManager); ok {
			return manager, nil
		}
		current = current.Parent()
	}
	return nil, fmt.Errorf("config manager not found in context")
}

// newGetConfigCmd creates and returns a new cobra.Command for getting configuration values.
// It allows users to retrieve values from the configuration by specifying a key.
//
// The command expects exactly one argument which is the configuration key to look up.
// For security purposes, it masks API keys by showing only the last 4 characters.
// If the value is a map, it masks any API keys contained within it.
//
// Usage:
//   - get [key]
//
// Example:
//   - get openai.api_key      // Returns masked API key
//   - get openai              // Returns masked configuration map
//
// Returns an error if:
//   - The configuration manager cannot be initialized
//   - The specified key is not found in the configuration
//   - The value cannot be marshaled to JSON
func newGetConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config value")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			value, ok := cfgManager.Get(args[0])
			if !ok {
				return fmt.Errorf("config key not found: %s", args[0])
			}

			if m, ok := value.(map[string]interface{}); ok {
				config.MaskConfigAPIKeys(m)
			}

			if strings.HasSuffix(args[0], ".api_key") {
				if strValue, ok := value.(string); ok {
					value = config.MaskAPIKey(strValue, 4)
				}
			}

			data, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf("Value for key '%s':\n%s\n", args[0], string(data))
			return nil
		},
	}
}

// newListConfigCmd creates and returns a new Cobra command for listing configuration content.
// The command allows users to view the current configuration settings without prompt.
// If any errors occur during the process, they are returned with appropriate context.
//
// Returns:
//   - *cobra.Command: A command object that handles the 'list' subcommand functionality
func newListConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List config content",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting list config content")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			configStr, err := cfgManager.List()
			if err != nil {
				return fmt.Errorf("failed to list config: %w", err)
			}
			fmt.Print(configStr)
			return nil
		},
	}
}

// newResetConfigCmd creates and returns a new Cobra command for resetting configuration content.
// The command allows users to reset the current configuration settings or only the prompt configuration.
// If any errors occur during the process, they are returned with appropriate context.
//
// Flags:
//
//   - --prompt: Reset only prompt configuration
//
// Returns:
//   - *cobra.Command: A command object that handles the 'reset' subcommand functionality
func newResetConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset config",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting reset config")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			promptOnly, _ := cmd.Flags().GetBool("prompt")
			if err := cfgManager.Reset(promptOnly); err != nil {
				return err
			}

			if promptOnly {
				fmt.Println("Prompt configuration has been reset to default values")
			} else {
				fmt.Println("Configuration has been reset to default values")
			}
			return nil
		},
	}
	cmd.Flags().Bool("prompt", false, "Reset only prompt configuration")
	return cmd
}

// newSetConfigCmd creates a new Cobra command for setting configuration values.
// It allows users to set key-value pairs in the configuration where the value
// can be either a JSON-formatted string or a plain string.
//
// The command takes exactly two arguments:
//   - key: The configuration key to set
//   - value: The value to set (can be JSON or plain string)
//
// Example usage:
//
//	gptcomet config set api.key "your-api-key"
//	gptcomet config set limits.max 100
//	gptcomet config set options {"debug":true,"timeout":30}
//
// Returns a pointer to cobra.Command configured for the 'set' operation.
func newSetConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting set config value")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			var value interface{}
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1]
			}

			if err := cfgManager.Set(args[0], value); err != nil {
				return err
			}

			fmt.Printf("Successfully set '%s' to: %v\n", args[0], args[1])
			return nil
		},
	}
}

// newPathConfigCmd creates and returns a new cobra.Command for the 'path' subcommand.
// This command displays the full path to the configuration file being used by the application.
// It uses the config manager to retrieve the path and outputs it to stdout.
// If there's an error getting the config manager, it will return the error.
func newPathConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Get config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config file path")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Configuration file path: %s\n", cfgManager.GetPath())
			return nil
		},
	}
}

// newRemoveConfigCmd creates a new cobra.Command for removing configuration values.
// It supports two modes of operation:
// 1. Remove an entire configuration key: `remove [key]`
// 2. Remove a specific value from a list: `remove [key] [value]`
//
// The command requires at least one argument (key) and accepts an optional second argument (value).
// If only the key is provided, the entire configuration entry is removed.
// If both key and value are provided, only the specified value is removed from the list associated with the key.
//
// Returns:
//   - *cobra.Command: A configured cobra command ready to handle config removal operations
func newRemoveConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [key] [value]",
		Short: "Remove config value or a value from a list",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting remove config value")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			_, exists := cfgManager.Get(args[0])
			if !exists {
				fmt.Printf("Key '%s' not found in configuration\n", args[0])
				return nil
			}

			value := ""
			if len(args) > 1 {
				value = args[1]
			}

			if err := cfgManager.Remove(args[0], value); err != nil {
				return err
			}

			if value == "" {
				fmt.Printf("Successfully removed key '%s'\n", args[0])
			} else {
				fmt.Printf("Successfully removed '%s' from '%s'\n", value, args[0])
			}
			return nil
		},
	}
}

// newAppendConfigCmd creates a new cobra.Command for appending values to a list configuration.
// The command takes two arguments: a key and a value to append.
// The value can be a JSON-formatted string which will be unmarshaled, or a plain string.
// If the specified key doesn't exist, it will create a new list.
// If the key exists but is not a list, it will be converted to a list.
// Usage:
//   - append [key] [value]
//
// Example:
//   - append "file_ignore" "go.sum"
//   - append "numbers" "[1,2,3]"
//
// Returns a cobra.Command instance configured for the append operation.
func newAppendConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "append [key] [value]",
		Short: "Append value to a list config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting append config value")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			var value interface{}
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1]
			}

			current, exists := cfgManager.Get(args[0])
			if exists {
				if _, ok := current.([]interface{}); !ok {
					return fmt.Errorf("key '%s' exists but is not a list", args[0])
				}
				// check if value already exists in the list
				for _, v := range current.([]interface{}) {
					if v == value {
						fmt.Printf("Value '%v' already exists in the list for key '%s'\n", value, args[0])
						return nil
					}
				}
			}

			if err := cfgManager.Append(args[0], value); err != nil {
				return err
			}

			fmt.Printf("Successfully appended '%v' to '%s'\n", args[1], args[0])
			return nil
		},
	}
}

// newKeysConfigCmd creates and returns a new cobra.Command that lists all supported configuration keys.
// The command:
//   - Has the name "keys"
//   - Lists all configuration keys supported by the application
//   - Retrieves the configuration manager and prints each supported key
//   - Returns an error if the configuration manager cannot be retrieved
//
// Returns:
//   - *cobra.Command: A command instance that handles the 'keys' subcommand
func newKeysConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keys",
		Short: "List all supported configuration keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting list supported config keys")

			cfgManager, err := getConfigManager(cmd)
			if err != nil {
				return err
			}

			keys := cfgManager.GetSupportedKeys()

			fmt.Println("Supported configuration keys:")
			for _, key := range keys {
				fmt.Printf("  %s\n", key)
			}
			return nil
		},
	}
}

// NewConfigCmd creates and returns a new cobra command for managing configuration.
// This command serves as the parent command for all configuration-related subcommands.
// It initializes the configuration manager during pre-run and stores it in the command context.
//
// The command includes the following subcommands:
// - get: Get configuration values
// - list: List configuration entries
// - reset: Reset configuration to default values
// - set: Set configuration values
// - path: Show configuration file path
// - remove: Remove configuration entries
// - append: Append values to configuration arrays
// - keys: List available configuration keys
//
// The configuration manager is initialized using the config path provided via the global --config flag.
// If initialization fails, an error is returned and the command execution is stopped.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			cfgManager, err := config.New(configPath)
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Store config manager in the config command's context
			setConfigManager(cmd, cfgManager)
			return nil
		},
	}

	cmd.AddCommand(
		newGetConfigCmd(),
		newListConfigCmd(),
		newResetConfigCmd(),
		newSetConfigCmd(),
		newPathConfigCmd(),
		newRemoveConfigCmd(),
		newAppendConfigCmd(),
		newKeysConfigCmd(),
	)
	return cmd
}
