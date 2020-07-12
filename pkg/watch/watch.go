package watch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/radovskyb/watcher"
)

const polingCycle = 100 * time.Millisecond

type Watcher struct {
	watcher     *watcher.Watcher
	HandleEvent func()
}

func New() Watcher {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create)
	return Watcher{
		watcher: w,
	}
}

func (watcher Watcher) Start(ctx context.Context, src, dest string) error {
	go func() {
		for {
			select {
			case <-watcher.watcher.Event:
				watcher.HandleEvent()
			case err := <-watcher.watcher.Error:
				log.Println("error occurs on watching "+src, err)
			case <-watcher.watcher.Closed:
				log.Println("watcher is closed")
				return
			case <-ctx.Done():
				watcher.watcher.Close()
			}
		}
	}()

	if err := watcher.watcher.Add(src); err != nil {
		return err
	}

	log.Println("start to watch " + src)
	if err := watcher.watcher.Start(polingCycle); err != nil {
		return fmt.Errorf("failed to start watching: %w", err)
	}

	return nil
}
