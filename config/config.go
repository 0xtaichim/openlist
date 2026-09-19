// Package config handles OpenList CLI configuration.
//
// Lookup order (highest priority last applied by the CLI):
//  1. Default values
//  2. Config file ($XDG_CONFIG_HOME/openlist/config.yaml)
//  3. Environment variables (OPENLIST_URL, OPENLIST_TOKEN)
//  4. Command-line flags
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	// AppName is the application name used for XDG directories.
	AppName = "openlist"

	// DefaultURL is the default OpenList server URL.
	DefaultURL = "http://localhost:5244"

	// ConfigFileName is the config file base name (without extension).
	ConfigFileName = "config"

	// ConfigFileType is the config file extension.
	ConfigFileType = "yaml"
)

type ctxKey struct{}

// Config holds the application configuration.
type Config struct {
	URL   string `mapstructure:"url" yaml:"url"`
	Token string `mapstructure:"token" yaml:"token"`
}

// Load reads configuration from defaults, the config file, and the environment.
// Callers that need flag overrides should apply them afterwards.
func Load() (*Config, error) {
	v := newViper()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.SetURL(cfg.URL)
	return &cfg, nil
}

// WithContext stores cfg in ctx.
func WithContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxKey{}, cfg)
}

// FromContext returns the Config stored in ctx, or nil.
func FromContext(ctx context.Context) *Config {
	cfg, _ := ctx.Value(ctxKey{}).(*Config)
	return cfg
}

// Dir returns the XDG config directory for this application.
func Dir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, AppName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", AppName)
}

// FilePath returns the full path of the config file.
func FilePath() string {
	return filepath.Join(Dir(), ConfigFileName+"."+ConfigFileType)
}

// Save writes the configuration to the config file with restrictive permissions.
func (c *Config) Save() error {
	configDir := Dir()
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.Set("url", c.URL)
	v.Set("token", c.Token)

	configPath := FilePath()
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	if err := os.Chmod(configPath, 0o600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}
	return nil
}

// SetURL sets the URL after trimming space and trailing slashes.
func (c *Config) SetURL(raw string) {
	c.URL = normalizeURL(strings.TrimSpace(raw))
}

// SetToken sets the API token.
func (c *Config) SetToken(token string) {
	c.Token = strings.TrimSpace(token)
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("url", DefaultURL)
	v.SetDefault("token", "")

	v.SetEnvPrefix("OPENLIST")
	_ = v.BindEnv("url")
	_ = v.BindEnv("token")

	v.AddConfigPath(Dir())
	v.SetConfigName(ConfigFileName)
	v.SetConfigType(ConfigFileType)
	return v
}

// normalizeURL trims trailing slashes without turning "https://" into "https:".
func normalizeURL(u string) string {
	for strings.HasSuffix(u, "/") {
		trimmed := strings.TrimSuffix(u, "/")
		if !strings.Contains(trimmed, "://") {
			break
		}
		u = trimmed
	}
	return u
}
