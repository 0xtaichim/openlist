package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"openlist/client"
	"openlist/config"

	"github.com/spf13/cobra"
)

// Execute runs the CLI with ctx for cancellation (SIGINT/SIGTERM).
func Execute(ctx context.Context) error {
	root := NewRootCmd()
	err := root.ExecuteContext(ctx)
	if err != nil {
		_ = writeJSON(root.OutOrStdout(), map[string]string{"error": err.Error()})
	}
	return err
}

// NewRootCmd constructs the openlist command tree.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "openlist",
		Short: "A CLI tool for OpenList",
		Long: `OpenList CLI is a command line interface for managing files
and directories on your OpenList server.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if changed, value := persistentFlag(cmd, "url"); changed {
				cfg.SetURL(value)
			}
			if changed, value := persistentFlag(cmd, "token"); changed {
				cfg.SetToken(value)
			}
			cmd.SetContext(config.WithContext(cmd.Context(), cfg))
			return nil
		},
	}

	root.PersistentFlags().StringP("url", "u", config.DefaultURL, "OpenList server URL")
	root.PersistentFlags().StringP("token", "t", "", "OpenList API token")

	root.AddCommand(newConfigCmd())
	root.AddCommand(newFSCmd())
	root.AddCommand(newStrmCmd())
	return root
}

func persistentFlag(cmd *cobra.Command, name string) (bool, string) {
	if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
		return true, f.Value.String()
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil && f.Changed {
		return true, f.Value.String()
	}
	if root := cmd.Root(); root != nil {
		if f := root.PersistentFlags().Lookup(name); f != nil && f.Changed {
			return true, f.Value.String()
		}
	}
	return false, ""
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printJSON(cmd *cobra.Command, v any) error {
	if err := writeJSON(cmd.OutOrStdout(), v); err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}
	return nil
}

func getClient(cmd *cobra.Command) (*client.Client, error) {
	cfg := config.FromContext(cmd.Context())
	if cfg == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}
	if cfg.Token == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: No token provided. Operations might fail if auth is required.")
	}
	return client.NewClient(cfg.URL, cfg.Token), nil
}

func mustGetString(cmd *cobra.Command, name string) (string, error) {
	return cmd.Flags().GetString(name)
}

func mustGetBool(cmd *cobra.Command, name string) (bool, error) {
	return cmd.Flags().GetBool(name)
}

func mustGetInt(cmd *cobra.Command, name string) (int, error) {
	return cmd.Flags().GetInt(name)
}

func mustGetStringSlice(cmd *cobra.Command, name string) ([]string, error) {
	return cmd.Flags().GetStringSlice(name)
}
