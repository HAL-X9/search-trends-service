package config

import "fmt"

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	cfg, err := ReadAndDecodeYaml(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load app configuration from YAML: %w", err)
	}

	if err = Validate(cfg); err != nil {
		return nil, fmt.Errorf("failed to validate app configuration: %w", err)
	}

	return cfg, nil
}
