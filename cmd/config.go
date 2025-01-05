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

// configKey is a custom type for the context key to avoid collisions
type configKey struct{}

// NewConfigCmd creates a new config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

			// Store config manager in context
			cmd.SetContext(context.WithValue(cmd.Context(), configKey{}, cfgManager))
			return nil
		},
	}

	// get command
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config value")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			// Get config value
			value, ok := cfgManager.Get(args[0])
			if !ok {
				return fmt.Errorf("config key not found: %s", args[0])
			}
			// If the value is a map, check for api_key and mask it
			if m, ok := value.(map[string]interface{}); ok {
				config.MaskConfigAPIKeys(m)
			}

			// Check if the value is an API key
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

	// list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List config content",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting list config content")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			configStr, err := cfgManager.List()
			if err != nil {
				return fmt.Errorf("failed to list config: %w", err)
			}
			fmt.Print(configStr)
			return nil
		},
	}

	// reset command
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset config",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting reset config")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
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
	resetCmd.Flags().Bool("prompt", false, "Reset only prompt configuration")

	// set command
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting set config value")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
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

	// path command
	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Get config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config file path")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			fmt.Printf("Configuration file path: %s\n", cfgManager.GetPath())
			return nil
		},
	}

	// remove command
	removeCmd := &cobra.Command{
		Use:   "remove [key] [value]",
		Short: "Remove config value or a value from a list",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting remove config value")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			// Check if key exists before removing
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

	// append command
	appendCmd := &cobra.Command{
		Use:   "append [key] [value]",
		Short: "Append value to a list config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting append config value")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			var value interface{}
			if err := json.Unmarshal([]byte(args[1]), &value); err != nil {
				value = args[1]
			}

			// Check if key exists and is a list
			current, exists := cfgManager.Get(args[0])
			if exists {
				if _, ok := current.([]interface{}); !ok {
					fmt.Printf("Warning: Key '%s' exists but is not a list. It will be converted to a list.\n", args[0])
				}
			}

			if err := cfgManager.Append(args[0], value); err != nil {
				return err
			}

			fmt.Printf("Successfully appended '%v' to '%s'\n", args[1], args[0])
			return nil
		},
	}

	// keys command
	keysCmd := &cobra.Command{
		Use:   "keys",
		Short: "List all supported configuration keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting list supported config keys")

			// Get config manager from context
			cfgManager, ok := cmd.Context().Value(configKey{}).(*config.Manager)
			if !ok {
				return fmt.Errorf("config manager not found in context")
			}

			// Get supported keys
			keys := cfgManager.GetSupportedKeys()

			// Print keys
			fmt.Println("Supported configuration keys:")
			for _, key := range keys {
				fmt.Printf("  %s\n", key)
			}
			return nil
		},
	}

	cmd.AddCommand(getCmd, listCmd, resetCmd, setCmd, pathCmd, removeCmd, appendCmd, keysCmd)
	return cmd
}
