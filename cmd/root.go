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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override with command-line flags if explicitly set
		if cmd.Flags().Changed("url") {
			cfg.URL = viper.GetString("url")
		}
		if cmd.Flags().Changed("token") {
			cfg.Token = viper.GetString("token")
		}

		config.GlobalConfig = cfg
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
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
