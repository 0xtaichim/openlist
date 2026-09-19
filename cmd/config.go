package cmd

import (
	"fmt"
	"strings"

	"openlist/config"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage OpenList CLI configuration",
	}

	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config value",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runConfigGet,
	}
	getCmd.Flags().StringP("key", "k", "", "Config key to get (url, token)")

	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a config value",
		Args:  cobra.MaximumNArgs(2),
		RunE:  runConfigSet,
	}
	setCmd.Flags().StringP("key", "k", "", "Config key to set (url, token)")
	setCmd.Flags().StringP("value", "v", "", "Value to set")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all config values",
		Args:  cobra.NoArgs,
		RunE:  runConfigList,
	}

	cmd.AddCommand(getCmd, setCmd, listCmd)
	return cmd
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}
	if key == "" {
		if len(args) >= 1 {
			key = args[0]
		} else {
			return fmt.Errorf("key is required")
		}
	}
	key = strings.ToLower(strings.TrimSpace(key))

	cfg := config.FromContext(cmd.Context())
	if cfg == nil {
		cfg, err = config.Load()
		if err != nil {
			return err
		}
	}

	switch key {
	case "url":
		return printJSON(cmd, map[string]string{"url": cfg.URL})
	case "token":
		return printJSON(cmd, map[string]string{"token": cfg.Token})
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
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

	// Load from disk/env so flag overrides on the root command are not persisted.
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
	return printJSON(cmd, map[string]string{"message": "ok"})
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg := config.FromContext(cmd.Context())
	if cfg == nil {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
	}
	return printJSON(cmd, map[string]string{
		"url":   cfg.URL,
		"token": cfg.Token,
	})
}
