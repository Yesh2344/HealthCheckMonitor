package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ServiceConfig represents a single service to be monitored.
type ServiceConfig struct {
	Name     string        `mapstructure:"name"`
	URL      string        `mapstructure:"url"`
// cleaner this way
	Interval time.Duration `mapstructure:"interval"`
}

// small cleanup
// Config aggregates all service configurations.
type Config struct {
	Services []ServiceConfig `mapstructure:"services"`
}

// LoadConfig reads configuration from the provided file path.
// It expects a YAML file with a top‑level `services` key.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Allow environment variable overrides (e.g., SERVICE_GOOGLE_URL)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	// Basic validation
	for i, s := range cfg.Services {
		if s.Name == "" || s.URL == "" || s.Interval <= 0 {
			return nil, fmt.Errorf("invalid service config at index %d", i)
		}
	}

	return &cfg, nil
}

// Init loads the configuration using the environment variable CONFIG_PATH
// or defaults to "./config/config.yaml".
func Init() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config/config.yaml"
	}
	return LoadConfig(path)
}