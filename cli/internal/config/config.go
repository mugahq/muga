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

func configPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, configDir)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", configDir)
}
