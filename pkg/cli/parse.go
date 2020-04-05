package cli

import (
	"errors"
	"fmt"
)

func (parser *ConfigParser) SetArtifacts(cfg map[string]interface{}, artifacts []interface{}) error {
	b, ok := cfg["build"]
	if !ok {
		b = map[string]interface{}{}
	}
	build, ok := b.(map[string]interface{})
	if !ok {
		return errors.New("the configuration 'build' should be map")
	}
	build["artifacts"] = artifacts
	cfg["build"] = build
	return nil
}

func (parser *ConfigParser) SetManifests(cfg map[string]interface{}, manifests []string) error {
	d, ok := cfg["deploy"]
	if !ok {
		d = map[string]interface{}{}
	}
	deploy, ok := d.(map[string]interface{})
	if !ok {
		return errors.New("the configuration 'deploy' should be map")
	}
	k, ok := deploy["kubectl"]
	if !ok {
		k = map[string]interface{}{}
	}
	kubectl, ok := k.(map[string]interface{})
	if !ok {
		return errors.New("the configuration 'deploy.kubectl' should be map")
	}
	kubectl["manifests"] = manifests
	deploy["kubectl"] = kubectl

	cfg["deploy"] = deploy
	return nil
}

func (parser *ConfigParser) Parse(cfg *Config, targets map[string]struct{}) (map[string]interface{}, error) {
	artifacts := []interface{}{}
	manifestsMap := map[string]struct{}{}
	for _, service := range cfg.Services {
		if len(targets) != 0 {
			if _, ok := targets[service.Name]; !ok {
				continue
			}
		}
		artifacts = append(artifacts, service.Artifacts...)
		for _, m := range service.Manifests {
			manifestsMap[m] = struct{}{}
		}
	}
	manifests := make([]string, len(manifestsMap))
	i := 0
	for m := range manifestsMap {
		manifests[i] = m
		i++
	}

	if cfg.Base == nil {
		cfg.Base = map[string]interface{}{}
	}

	if err := parser.SetArtifacts(cfg.Base, artifacts); err != nil {
		return nil, fmt.Errorf("failed to update artifacts: %w", err)
	}
	if err := parser.SetManifests(cfg.Base, manifests); err != nil {
		return nil, fmt.Errorf("failed to update manifests: %w", err)
	}
	return cfg.Base, nil
}
