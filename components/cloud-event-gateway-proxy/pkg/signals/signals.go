package signals

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var onlyOneSignalHandler = make(chan struct{})

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

// NewContext creates a new context with SetupSignalHandler()
// as our Done() channel.
func NewContext() context.Context {
	return &signalContext{stopCh: SetupSignalHandler()}
}

type signalContext struct {
	stopCh <-chan struct{}
}

// Deadline implements context.Context
func (scc *signalContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements context.Context
func (scc *signalContext) Done() <-chan struct{} {
	return scc.stopCh
}

// Err implements context.Context
func (scc *signalContext) Err() error {
	select {
	case _, ok := <-scc.Done():
		if !ok {
			return errors.New("received a termination signal")
		}
	default:
	}
	return nil
}

// Value implements context.Context
func (scc *signalContext) Value(key interface{}) interface{} {
	return nil
}
