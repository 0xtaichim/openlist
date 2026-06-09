package cmd

import (
	"fmt"
	"os"
	"strings"

	"openlist/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage OpenList CLI configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a config value",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfigGet(cmd, args); err != nil {
			printErrorJSON(err)
			os.Exit(1)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfigSet(cmd, args); err != nil {
			printErrorJSON(err)
			os.Exit(1)
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runConfigList(); err != nil {
			printErrorJSON(err)
			os.Exit(1)
		}
	},
}

func init() {
	configGetCmd.Flags().StringP("key", "k", "", "Config key to get (url, token)")

	configSetCmd.Flags().StringP("key", "k", "", "Config key to set (url, token)")
	configSetCmd.Flags().StringP("value", "v", "", "Value to set")

	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	RootCmd.AddCommand(configCmd)
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		if len(args) >= 1 {
			key = args[0]
		} else {
			return fmt.Errorf("key is required")
		}
	}
	key = strings.ToLower(strings.TrimSpace(key))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "url":
		printJSON(map[string]string{"url": cfg.URL})
	case "token":
		printJSON(map[string]string{"token": cfg.Token})
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	value, _ := cmd.Flags().GetString("value")
	if key == "" && len(args) >= 1 {
		key = args[0]
	}
	if value == "" && len(args) >= 2 {
		value = args[1]
	}
	if key == "" || value == "" {
		return fmt.Errorf("key and value are required")
	}
	key = strings.ToLower(strings.TrimSpace(key))

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch key {
	case "url":
		cfg.SetURL(value)
	case "token":
		cfg.SetToken(value)
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return err
	}
	printJSON(map[string]string{"message": "ok"})
	return nil
}

func runConfigList() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	printJSON(map[string]string{
		"url":   cfg.URL,
		"token": cfg.Token,
	})
	return nil
}
