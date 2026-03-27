package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultAPIURL = "https://api.muga.sh"
	configDir     = "muga"
	configFile    = "config"
	configType    = "toml"
)

// Config holds the CLI configuration.
type Config struct {
	APIURL  string `mapstructure:"api_url"`
	Project string `mapstructure:"project"`
}

// Load reads configuration from the config file and environment.
// Precedence: flags > env > config file > defaults.
func Load() (*Config, error) {
	viper.SetDefault("api_url", defaultAPIURL)
	viper.SetDefault("project", "")

	_ = viper.BindEnv("api_url", "MUGA_API_URL")
	_ = viper.BindEnv("project", "MUGA_PROJECT")

	viper.SetConfigName(configFile)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to the config file.
func Save(cfg *Config) error {
	dir := configPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	viper.Set("api_url", cfg.APIURL)
	viper.Set("project", cfg.Project)

	path := filepath.Join(dir, configFile+"."+configType)
	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

// SetProject updates the active project in the config file.
func SetProject(slug string) error {
	cfg, err := Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	cfg.Project = slug
	return Save(cfg)
}

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, configDir)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", configDir)
}
