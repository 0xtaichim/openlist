package cmd

import (
	"fmt"
	"os"

	"openlist/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "openlist",
	Short: "A CLI tool for OpenList",
	Long: `OpenList CLI is a command line interface for managing files 
and directories on your OpenList server.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override with command-line flags if explicitly set
		if flagChanged(cmd, "url") {
			cfg.URL = viper.GetString("url")
		}
		if flagChanged(cmd, "token") {
			cfg.Token = viper.GetString("token")
		}

		config.GlobalConfig = cfg
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		printErrorJSON(err)
		os.Exit(1)
	}
}

func init() {
	// Define persistent flags
	RootCmd.PersistentFlags().StringP("url", "u", config.DefaultURL, "OpenList server URL")
	RootCmd.PersistentFlags().StringP("token", "t", "", "OpenList API token")

	// Bind flags to viper
	viper.BindPFlag("url", RootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
}

func printErrorJSON(err error) {
	if err == nil {
		return
	}
	payload := map[string]string{
		"error": err.Error(),
	}
	printJSON(payload)
}

func flagChanged(cmd *cobra.Command, name string) bool {
	if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
		return true
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil && f.Changed {
		return true
	}
	if f := cmd.PersistentFlags().Lookup(name); f != nil && f.Changed {
		return true
	}
	if root := cmd.Root(); root != nil {
		if f := root.PersistentFlags().Lookup(name); f != nil && f.Changed {
			return true
		}
	}
	return false
}
