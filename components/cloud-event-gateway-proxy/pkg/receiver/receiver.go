package receiver

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
)

const (
	DefaultShutdownTimeout = time.Minute * 1
)

type HttpMessageReceiver struct {
	port        int
	metricsPort int
	handler     nethttp.Handler
	server      *nethttp.Server
	listener    net.Listener
}

func NewHttpMessageReceiver(port int) *HttpMessageReceiver {
	return &HttpMessageReceiver{
		port: port,
	}
}

// Blocking
func (recv *HttpMessageReceiver) StartListen(ctx context.Context, handler nethttp.Handler) error {
	var err error
	if recv.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", recv.port)); err != nil {
		return err
	}

	recv.handler = CreateHandler(handler)

	recv.server = &nethttp.Server{
		Addr:    recv.listener.Addr().String(),
		Handler: recv.handler,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- recv.server.Serve(recv.listener)
	}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), getShutdownTimeout(ctx))
		defer cancel()
		err := recv.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

type shutdownTimeoutKey struct{}

func getShutdownTimeout(ctx context.Context) time.Duration {
	v := ctx.Value(shutdownTimeoutKey{})
	if v == nil {
		return DefaultShutdownTimeout
	}
	return v.(time.Duration)
}

func CreateHandler(handler nethttp.Handler) nethttp.Handler {
	return &ochttp.Handler{
		//Propagation: tracecontextb3.TraceContextEgress,
		Handler: handler,
	}
}
