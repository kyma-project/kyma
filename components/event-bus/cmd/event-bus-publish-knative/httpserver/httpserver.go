package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HTTPServer encapsulates an http.Server.
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer creates a new HTTPServer instance.
func NewHTTPServer(port *int, handler *http.Handler) *HTTPServer {
	if port == nil {
		log.Println("cannot create HTTP server the port is missing")
		return nil
	}
	if handler == nil {
		log.Println("cannot create HTTP server the HTTP handler function is missing")
		return nil
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: *handler,
	}

	httpServer := HTTPServer{server: server}
	return &httpServer
}

// Start HTTPServer
func (h *HTTPServer) Start() {
	if h.server == nil {
		log.Println("cannot start a nil HTTP server")
		return
	}

	if err := h.server.ListenAndServe(); err != nil {
		log.Printf("HTTP server ListenAndServe error: %v", err)
	}
}

// Shutdown HTTPServer
func (h *HTTPServer) Shutdown(timeout time.Duration) {
	if h.server == nil {
		log.Println("cannot shutdown a nil HTTP server")
		return
	}

	shutdownSignal := make(chan os.Signal, 1)
	defer close(shutdownSignal)

	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-shutdownSignal

	log.Printf("HTTP server shutdown with timeout: %s\n", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := h.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	} else {
		log.Println("HTTP server shutdown successful")
	}
}
