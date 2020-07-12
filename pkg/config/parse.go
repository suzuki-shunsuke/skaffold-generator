package config

import (
	"errors"
	"fmt"
)

type Parser struct {
	Path string
}

// Parse returns data of skaffold.yaml.
// targets is a list of service names.
// targets and services which targets depend on are launched.
// If targets is empty, all services are launched.
func (parser Parser) Parse(cfg Config, targets map[string]struct{}) (map[string]interface{}, error) {
	// service name -> service names which the service depends on
	dependencyMap := make(map[string]map[string]struct{}, len(cfg.Services))
	setDependencyMap(cfg.Services, dependencyMap)
	services := make(map[string]struct{}, len(targets))
	if err := calcTargets(dependencyMap, targets, services); err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}

	// artifacts is `skaffold.yaml`'s `build.artifacts`.
	// Docker images which are built.
	artifacts := []interface{}{}

	// manifetsMap is a list of manifest file paths.
	// `deploy.kubectl.manifests` of `skaffold.yaml`.
	manifestsMap := map[string]struct{}{}

	isServicesNotEmpty := len(services) != 0

	for _, service := range cfg.Services {
		if isServicesNotEmpty {
			// If services are specified, the service which isn't in services is ignored.
			if _, ok := services[service.Name]; !ok {
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

	if err := setArtifacts(cfg.Base, artifacts); err != nil {
		return nil, fmt.Errorf("failed to update artifacts: %w", err)
	}
	if err := setManifests(cfg.Base, manifests); err != nil {
		return nil, fmt.Errorf("failed to update manifests: %w", err)
	}
	return cfg.Base, nil
}

func setDependencyMap(services []ServiceConfig, dependencyMap map[string]map[string]struct{}) {
	for _, service := range services {
		dependencies := make(map[string]struct{}, len(service.DependsOn))
		for _, d := range service.DependsOn {
			dependencies[d] = struct{}{}
		}
		dependencyMap[service.Name] = dependencies
	}
}

// setArtifacts sets `build.artifacts` of `skaffold.yaml`.
// cfg is data of `skaffold.yaml` and is updated in this method.
func setArtifacts(cfg map[string]interface{}, artifacts []interface{}) error {
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

func setManifests(cfg map[string]interface{}, manifests []string) error {
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

// calcTargets adds targets and services which target depends on to `services`,
// which means `services` is updated.
func calcTargets(
	dependencyMap map[string]map[string]struct{}, targets, services map[string]struct{},
) error {
	for target := range targets {
		if err := calcTargetsRecursively(dependencyMap, target, services); err != nil {
			return err
		}
	}

	return nil
}

// calcTargetsRecursively adds target and services which target depends on to targets,
// which means targets is updated in this function.
// If target is already included in targets, do nothing.
// dependencyMap is a map of service and services which the service depends on.
func calcTargetsRecursively(
	dependencyMap map[string]map[string]struct{}, target string, targets map[string]struct{},
) error {
	if _, ok := targets[target]; ok {
		// if target is already included in targets, do nothing.
		return nil
	}

	// dependencies are services which target depends on
	dependencies, ok := dependencyMap[target]
	if !ok {
		return errors.New("undefined service: " + target)
	}
	targets[target] = struct{}{}

	for dependency := range dependencies {
		if err := calcTargetsRecursively(dependencyMap, dependency, targets); err != nil {
			return err
		}
	}
	return nil
}
