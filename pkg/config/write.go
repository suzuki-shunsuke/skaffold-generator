package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Writer struct {
	Path string
}

func (writer Writer) Write(cfg interface{}) error {
	dest := writer.Path

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to open skaffold.yaml: %w", err)
	}
	defer f.Close()
	if err := yaml.NewEncoder(f).Encode(&cfg); err != nil {
		return fmt.Errorf("failed to write YAML in skaffold.yaml: %w", err)
	}
	return nil
}
