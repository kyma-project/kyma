package receiver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
)

const (
	DefaultShutdownTimeout = time.Minute * 1
)

type HttpMessageReceiver struct {
	port     int
	handler  http.Handler
	server   *http.Server
	listener net.Listener
}

func NewHttpMessageReceiver(port int) *HttpMessageReceiver {
	return &HttpMessageReceiver{
		port: port,
	}
}

// Blocking
func (recv *HttpMessageReceiver) StartListen(ctx context.Context, handler http.Handler) error {
	var err error
	if recv.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", recv.port)); err != nil {
		return err
	}

	recv.handler = CreateHandler(handler)

	recv.server = &http.Server{
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
		ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
		defer cancel()
		err := recv.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

func CreateHandler(handler http.Handler) http.Handler {
	return &ochttp.Handler{Handler: handler}
}
