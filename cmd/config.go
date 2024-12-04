package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/belingud/gptcomet/internal/config"
	"github.com/belingud/gptcomet/internal/debug"

	"github.com/spf13/cobra"
)

// NewConfigCmd creates a new config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	// get command
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config value")

			cfgManager, err := config.New()
			if err != nil {
				return err
			}

			value, exists := cfgManager.Get(args[0])
			if !exists {
				fmt.Printf("Key '%s' not found in configuration\n", args[0])
				return nil
			}

			// If the value is a map, check for api_key and mask it
			if m, ok := value.(map[string]interface{}); ok {
				config.MaskConfigAPIKeys(m)
			}
			// If the value is a string and the key is api_key, mask it
			if s, ok := value.(string); ok && args[0] == "api_key" {
				value = config.MaskAPIKey(s, 3)
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

			cfgManager, err := config.New()
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

	// reset command
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset config",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting reset config")

			cfgManager, err := config.New()
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
	resetCmd.Flags().Bool("prompt", false, "Reset only prompt configuration")

	// set command
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting set config value")

			cfgManager, err := config.New()
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

	// path command
	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Get config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug.Println("Starting get config file path")

			cfgManager, err := config.New()
			if err != nil {
				return err
			}
			debug.Println("Created config manager")

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

			cfgManager, err := config.New()
			if err != nil {
				return err
			}
			debug.Println("Created config manager")

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

			cfgManager, err := config.New()
			if err != nil {
				return err
			}
			debug.Println("Created config manager")

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

			// Get config manager
			cfgManager, err := config.New()
			if err != nil {
				return err
			}
			debug.Println("Created config manager")

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
