package receiver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"

	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
)

const (
	// defaultShutdownTimeout is the default timeout for the receiver to shutdown.
	defaultShutdownTimeout = time.Minute * 1

	receiverName = "receiver"

	readHeaderTimeout = time.Second * 5
)

// HTTPMessageReceiver is responsible for receiving messages over HTTP.
type HTTPMessageReceiver struct {
	port     int
	handler  http.Handler
	server   *http.Server
	listener net.Listener
}

// NewHTTPMessageReceiver returns a new NewHTTPMessageReceiver instance with the given port.
func NewHTTPMessageReceiver(port int) *HTTPMessageReceiver {
	return &HTTPMessageReceiver{port: port}
}

// StartListen starts the HTTP message receiver and blocks until it receives a shutdown signal.
func (recv *HTTPMessageReceiver) StartListen(ctx context.Context, handler http.Handler, logger *kymalogger.Logger) error {
	var err error
	if recv.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", recv.port)); err != nil {
		return err
	}

	recv.handler = createHandler(handler)
	recv.server = &http.Server{
		ReadHeaderTimeout: readHeaderTimeout,
		Addr:              recv.listener.Addr().String(),
		Handler:           recv.handler,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- recv.server.Serve(recv.listener)
	}()

	// init the contexted logger
	namedLogger := logger.WithContext().Named(receiverName)

	namedLogger.Info("Event Publisher Receiver has started.")

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
