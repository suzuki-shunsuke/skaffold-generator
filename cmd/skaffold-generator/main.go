package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/suzuki-shunsuke/skaffold-generator/pkg/cli"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	exitChan := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)
	sentSignals := map[os.Signal]struct{}{}

	runner := cli.Runner{}
	ctx := context.Background()

	go func() {
		exitChan <- runner.Run(ctx, os.Args...)
	}()

	for {
		select {
		case err := <-exitChan:
			return err
		case sig := <-signalChan:
			if _, ok := sentSignals[sig]; ok {
				continue
			}
			sentSignals[sig] = struct{}{}
			exitChan <- nil
		}
	}
}
