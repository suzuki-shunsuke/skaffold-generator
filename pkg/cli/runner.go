package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type Runner struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (runner *Runner) Run(ctx context.Context, args ...string) error {
	app := &cli.App{
		Name:  "skaffold-generator",
		Usage: "generate skaffold.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "src",
				Aliases: []string{"s"},
				Usage:   "configuration file path (skaffold-generator.yaml)",
			},
			&cli.StringFlag{
				Name:    "dest",
				Aliases: []string{"d"},
				Usage:   "generated configuration file path (skaffold.yaml)",
			},
		},
		Action: runner.action,
	}

	return app.RunContext(ctx, args)
}

func (runner *Runner) action(c *cli.Context) error {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create)
	targets := make(map[string]struct{}, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		targets[arg] = struct{}{}
	}

	src := "skaffold-generator.yaml"
	dest := "skaffold.yaml"

	parser := &ConfigParser{path: src}
	writer := &ConfigWriter{path: dest}

	if err := runner.Generate(parser, writer, targets); err != nil {
		log.Println("failed to update skaffold.yaml", err)
	}

	go func() {
		for {
			select {
			case <-w.Event:
				log.Println("detect the update of skaffold-generator.yaml")
				if err := runner.Generate(parser, writer, targets); err != nil {
					log.Println("failed to update skaffold.yaml", err)
				}
			case err := <-w.Error:
				log.Println("error occurs on watching skaffold-generator.yaml", err)
			case <-w.Closed:
				log.Println("watcher is closed")
				return
			}
		}
	}()

	if err := w.Add(src); err != nil {
		return err
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	log.Println("start to watch skaffold-generator.yaml")
	<-c.Done()
	return nil
}

func (runner *Runner) Generate(parser *ConfigParser, writer *ConfigWriter, targets map[string]struct{}) error {
	cfg, err := parser.Read()
	if err != nil {
		return fmt.Errorf("failed to read skaffold-generator.yaml: %w", err)
	}
	cfgMap, err := parser.Parse(cfg, targets)
	if err != nil {
		return fmt.Errorf("failed to parse skaffold-generator.yaml: %w", err)
	}
	if err := writer.Write(cfgMap); err != nil {
		return fmt.Errorf("failed to update skaffold.yaml: %w", err)
	}
	return nil
}

type ConfigWriter struct {
	path string
}

func (writer *ConfigWriter) Write(cfg interface{}) error {
	dest := writer.path

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

type Config struct {
	Services []ServiceConfig
	Base     map[string]interface{}
}

type ServiceConfig struct {
	Name      string
	Artifacts []interface{}
	Manifests []string
	DependsOn []string `yaml:"depends_on"`
}

type ConfigParser struct {
	path string
}

func (parser *ConfigParser) Read() (*Config, error) {
	f, err := os.Open(parser.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open skaffold-generator.yaml: %w", err)
	}
	defer f.Close()
	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to read skaffold-generator.yaml as YAML: %w", err)
	}
	return cfg, nil
}

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
