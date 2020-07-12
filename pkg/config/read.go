package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func (parser Parser) Read() (Config, error) {
	f, err := os.Open(parser.Path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open skaffold-generator.yaml: %w", err)
	}
	defer f.Close()
	cfg := Config{}
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to read skaffold-generator.yaml as YAML: %w", err)
	}
	return cfg, nil
}
