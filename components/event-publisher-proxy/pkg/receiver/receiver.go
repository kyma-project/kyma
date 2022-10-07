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
	// defaultShutdownTimeout is the default timeout for the receiver to shut down.
	defaultShutdownTimeout = 1 * time.Minute
	readHeaderTimeout      = 5 * time.Second

	receiverName = "receiver"
)

// HTTPMessageReceiver is responsible for receiving messages over HTTP.
type HTTPMessageReceiver struct {
	Host     string
	Port     int
	handler  http.Handler
	server   *http.Server
	listener net.Listener
}

// NewHTTPMessageReceiver returns a new NewHTTPMessageReceiver instance with the given Port.
func NewHTTPMessageReceiver(port int) *HTTPMessageReceiver {
	return &HTTPMessageReceiver{Port: port}
}

// StartListen starts the HTTP message receiver and blocks until it receives a shutdown signal.
func (r *HTTPMessageReceiver) StartListen(ctx context.Context, handler http.Handler, logger *kymalogger.Logger) error {
	var err error
	if r.listener, err = net.Listen("tcp", fmt.Sprintf("%v:%d", r.Host, r.Port)); err != nil {
		return err
	}

	r.handler = createHandler(handler)
	r.server = &http.Server{
		Addr:              r.listener.Addr().String(),
		Handler:           r.handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- r.server.Serve(r.listener)
	}()

	// init the contexted logger
	namedLogger := logger.WithContext().Named(receiverName)

	namedLogger.Info("Event Publisher Receiver has started.")

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		logger.WithContext().Info("shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()
		err := r.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		logger.WithContext().With(err).Error(err)
		return err
	}
}

// createHandler returns a new opentracing HTTP handler wrapper for the given HTTP handler.
func createHandler(handler http.Handler) http.Handler {
	return &ochttp.Handler{Handler: handler}
}

func (r *HTTPMessageReceiver) BaseURL() string {
	return fmt.Sprintf("http://%s", r.listener.Addr())
}
