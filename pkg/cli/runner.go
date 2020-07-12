package cli

import (
	"context"
	"io"
	"log"

	"github.com/suzuki-shunsuke/skaffold-generator/pkg/config"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/constant"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/controller"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/watch"
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

func (runner Runner) action(c *cli.Context) error {
	// targets is a list of service names.
	// targets and services which targets depend on are launched.
	targets := make(map[string]struct{}, c.Args().Len())
	for _, arg := range c.Args().Slice() {
		targets[arg] = struct{}{}
	}

	src := c.String("src")
	dest := c.String("dest")

	w := watch.New()

	ctrl := controller.Controller{
		ConfigParser: config.Parser{Path: src},
		ConfigWriter: config.Writer{Path: dest},
	}

	w.HandleEvent = func() {
		log.Println("detect the update of " + src)
		if err := ctrl.Generate(targets); err != nil {
			log.Println("failed to update "+dest, err)
		}
	}

	if err := ctrl.Generate(targets); err != nil {
		log.Println("failed to update "+dest, err)
	}

	return w.Start(c.Context, src, dest)
}
