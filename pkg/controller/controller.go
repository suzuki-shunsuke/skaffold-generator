package controller

import (
	"fmt"

	"github.com/suzuki-shunsuke/skaffold-generator/pkg/config"
)

type ConfigParser interface {
	Read() (config.Config, error)
	Parse(cfg config.Config, targets map[string]struct{}) (map[string]interface{}, error)
}

type ConfigWriter interface {
	Write(cfg interface{}) error
}

type Controller struct {
	ConfigParser ConfigParser
	ConfigWriter ConfigWriter
}

// Generate reads a configuration file `skaffold-generator.yaml` and generates `skaffold.yaml`.
// targets is a list of service names.
// targets and services which targets depend on are launched.
func (ctrl Controller) Generate(targets map[string]struct{}) error {
	cfg, err := ctrl.ConfigParser.Read()
	if err != nil {
		return fmt.Errorf("failed to read skaffold-generator.yaml: %w", err)
	}
	cfgMap, err := ctrl.ConfigParser.Parse(cfg, targets)
	if err != nil {
		return fmt.Errorf("failed to parse skaffold-generator.yaml: %w", err)
	}
	if err := ctrl.ConfigWriter.Write(cfgMap); err != nil {
		return fmt.Errorf("failed to update skaffold.yaml: %w", err)
	}
	return nil
}
