package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadAndDecodeYaml(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var out *Config

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err = dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode config %s: %w", path, err)
	}

	return out, nil
}
