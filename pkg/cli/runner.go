package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/config"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/constant"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/controller"
	"github.com/urfave/cli/v2"
)

type Runner struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (runner Runner) Run(ctx context.Context, args ...string) error {
	app := cli.App{
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
	// targets is a list of service names.
	// targets and services which targets depend on are launched.
	targets := make(map[string]struct{}, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		targets[arg] = struct{}{}
	}

	src := c.String("src")
	dest := c.String("dest")

	ctrl := controller.Controller{
		ConfigParser: config.Parser{Path: src},
		ConfigWriter: config.Writer{Path: dest},
	}

	if err := ctrl.Generate(targets); err != nil {
		log.Println("failed to update "+dest, err)
	}

	go func() {
		for {
			select {
			case <-w.Event:
				log.Println("detect the update of " + src)
				if err := ctrl.Generate(targets); err != nil {
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
