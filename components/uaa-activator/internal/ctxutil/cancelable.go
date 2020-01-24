package ctxutil

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// CancelableContext returns context that is canceled when
// application shutdown is requested.
func CancelableContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
			<-c
			os.Exit(1) // second signal. Exit directly.
		}
	}()

	return ctx, cancel
}
