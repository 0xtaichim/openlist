// Package config handles OpenList CLI configuration management.
// It follows XDG Base Directory Specification and supports:
//   - Command line flags (highest priority)
//   - Environment variables
//   - Config file (~/.config/openlist/config.yaml)
//   - Default values (lowest priority)
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// AppName is the application name for XDG directories
	AppName = "openlist"

	// DefaultURL is the default OpenList server URL
	DefaultURL = "http://localhost:5244"

	// ConfigFileName is the name of the config file
	ConfigFileName = "config"

	// ConfigFileType is the config file extension
	ConfigFileType = "yaml"
)

// Config holds the application configuration
type Config struct {
	URL   string `mapstructure:"url"`
	Token string `mapstructure:"token"`
}

// GlobalConfig is the loaded configuration instance
var GlobalConfig *Config

// Load initializes the configuration with the following priority:
// 1. Command line flags
// 2. Environment variables (OPENLIST_URL, OPENLIST_TOKEN)
// 3. Config file (~/.config/openlist/config.yaml)
// 4. Default values
func Load() (*Config, error) {
	// Set default values
	viper.SetDefault("url", DefaultURL)
	viper.SetDefault("token", "")

	// Set up environment variables
	viper.SetEnvPrefix("OPENLIST")
	viper.AutomaticEnv()
	_ = viper.BindEnv("url")
	_ = viper.BindEnv("token")

	// Set up config file
	configDir := getConfigDir()
	viper.AddConfigPath(configDir)
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)

	// Try to read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		// Only return error if it's not a "config file not found" error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

// GetConfigDir returns the XDG config directory for this application
func getConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, AppName)
	}

	// Fall back to ~/.config
	home, err := os.UserHomeDir()
	if err != nil {
		// Last resort: current directory
		return "."
	}
	return filepath.Join(home, ".config", AppName)
}

// Save saves the current configuration to the config file
func (c *Config) Save() error {
	configDir := getConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set values in viper
	viper.Set("url", c.URL)
	viper.Set("token", c.Token)

	// Write config file
	configPath := filepath.Join(configDir, ConfigFileName+"."+ConfigFileType)
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetURL returns the configured URL
func (c *Config) GetURL() string {
	return c.URL
}

// GetToken returns the configured token
func (c *Config) GetToken() string {
	return c.Token
}

// SetURL sets the URL (does not save to file automatically)
func (c *Config) SetURL(url string) {
	c.URL = url
	viper.Set("url", url)
}

// SetToken sets the token (does not save to file automatically)
func (c *Config) SetToken(token string) {
	c.Token = token
	viper.Set("token", token)
}
