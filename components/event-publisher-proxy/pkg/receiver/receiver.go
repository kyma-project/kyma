package receiver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"go.opencensus.io/plugin/ochttp"
)

const (
	// defaultShutdownTimeout is the default timeout for the receiver to shutdown.
	defaultShutdownTimeout = time.Minute * 1
)

// HttpMessageReceiver is responsible for receiving messages over HTTP.
type HttpMessageReceiver struct {
	port     int
	handler  http.Handler
	server   *http.Server
	listener net.Listener
}

// NewHttpMessageReceiver returns a new NewHttpMessageReceiver instance with the given port.
func NewHttpMessageReceiver(port int) *HttpMessageReceiver {
	return &HttpMessageReceiver{port: port}
}

// StartListen starts the HTTP message receiver and blocks until it receives a shutdown signal.
func (recv *HttpMessageReceiver) StartListen(ctx context.Context, handler http.Handler) error {
	var err error
	if recv.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", recv.port)); err != nil {
		return err
	}

	recv.handler = createHandler(handler)
	recv.server = &http.Server{
		Addr:    recv.listener.Addr().String(),
		Handler: recv.handler,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- recv.server.Serve(recv.listener)
	}()
	logrus.Info("Event Publisher Receiver has started.")

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		err := recv.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

// createHandler returns a new opentracing HTTP handler wrapper for the given HTTP handler.
func createHandler(handler http.Handler) http.Handler {
	return &ochttp.Handler{Handler: handler}
}
