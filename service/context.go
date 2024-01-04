package service

import (
	"context"
	"os"
	"os/signal"
)

func LoadContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
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
	}
}
