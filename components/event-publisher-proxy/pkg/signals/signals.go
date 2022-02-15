package signals

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// onlyOneSignalHandler to make sure that only one signal handler is registered.
	onlyOneSignalHandler = make(chan struct{})

	// shutdownSignals array of system signals to cause shutdown.
	shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}
)

// SetupSignalHandler registered for SIGTERM and SIGINT. A stop channel is returned
// which is closed on one of these signals. If a second signal is caught, the program
// is terminated with exit code 1.
func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	osSignal := make(chan os.Signal, 2)
	signal.Notify(osSignal, shutdownSignals...)
	go func() {
		<-osSignal
		close(stop)
		<-osSignal
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

// signalContext represents a signal context.
type signalContext struct {
	stopCh <-chan struct{}
}

// NewContext creates a new context with SetupSignalHandler()
// as our Done() channel.
func NewContext() context.Context {
	return &signalContext{stopCh: SetupSignalHandler()}
}

// Deadline implements context.Context.
func (scc *signalContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements context.Context.
func (scc *signalContext) Done() <-chan struct{} {
	return scc.stopCh
}

// Err implements context.Context.
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

// Value implements context.Context.
func (scc *signalContext) Value(interface{}) interface{} {
	return nil
}
