package service

import (
	"context"
	"os"
	"os/signal"

	"github.com/samber/lo"
)

var StopCh = &[]chan os.Signal{}

func StopSignalFIFO() {
	if len(*StopCh) > 0 {
		(*StopCh)[0] <- os.Interrupt
	}
}

func LoadContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	*StopCh = append(*StopCh, c)
	signal.Notify(c, os.Interrupt)
	go func() {
		_, ok := <-c
		if ok {
			cancel()
		}
	}()
	return ctx, func() {
		signal.Stop(c)
		close(c)
		*StopCh = lo.Filter[chan os.Signal](*StopCh, func(item chan os.Signal, index int) bool {
			return item != c
		})
	}
}
