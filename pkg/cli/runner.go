package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/constant"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner Runner) Run(ctx context.Context, args ...string) error {
	app := &cli.App{
		Name:            "skaffold-generator",
		Usage:           "generate skaffold.yaml",
		Version:         constant.Version,
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "src",
				Aliases: []string{"s"},
				Usage:   "configuration file path (skaffold-generator.yaml)",
				Value:   "skaffold-generator.yaml",
			},
			&cli.StringFlag{
				Name:    "dest",
				Aliases: []string{"d"},
				Usage:   "generated configuration file path (skaffold.yaml)",
				Value:   "skaffold.yaml",
			},
		},
		Action: runner.action,
	}

	return app.RunContext(ctx, args)
}

const polingCycle = 100 * time.Millisecond

func (runner Runner) action(c *cli.Context) error {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create)
	targets := make(map[string]struct{}, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		targets[arg] = struct{}{}
	}

	src := c.String("src")
	dest := c.String("dest")

	parser := ConfigParser{path: src}
	writer := ConfigWriter{path: dest}

	if err := runner.Generate(parser, writer, targets); err != nil {
		log.Println("failed to update "+dest, err)
	}

	go func() {
		for {
			select {
			case <-w.Event:
				log.Println("detect the update of " + src)
				if err := runner.Generate(parser, writer, targets); err != nil {
					log.Println("failed to update "+dest, err)
				}
			case err := <-w.Error:
				log.Println("error occurs on watching "+src, err)
			case <-w.Closed:
				log.Println("watcher is closed")
				return
			case <-c.Done():
				w.Close()
			}
		}
	}()

	if err := w.Add(src); err != nil {
		return err
	}

	log.Println("start to watch " + src)
	if err := w.Start(polingCycle); err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	return nil
}

func (runner Runner) Generate(parser ConfigParser, writer ConfigWriter, targets map[string]struct{}) error {
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

func (writer ConfigWriter) Write(cfg interface{}) error {
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

func (parser ConfigParser) Read() (Config, error) {
	f, err := os.Open(parser.path)
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
