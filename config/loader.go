package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// LoadConfig loads configuration from the specified YAML file.
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "./config/config.local.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	data = []byte(expandEnvVars(string(data)))

	config := &Config{}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true) // Enable strict mode to catch unknown fields

	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	applyDefaults(config)

	return config, nil
}

// expandEnvVars expands environment variables in the given string.
func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}

// validateConfig validates the configuration.
func validateConfig(cfg *Config) error {
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	return nil
}

// applyDefaults applies default values to the configuration.
func applyDefaults(cfg *Config) {
	// Logger
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Logger.Encoding == "" {
		cfg.Logger.Encoding = "json"
	}
	if len(cfg.Logger.OutputPaths) == 0 {
		cfg.Logger.OutputPaths = []string{"stdout"}
	}
}
