package httpserver

import (
	"fmt"
	"log"
	"net/http"
)

func NewHttpServer(port *int, handler *http.Handler) *http.Server {
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
	return server
}

func Close(server *http.Server) {
	if server == nil {
		log.Println("cannot close a nil HTTP server")
		return
	}
	if err := server.Close(); err != nil {
		log.Printf("failed to stop the http server: %v", err)
	}
}
