package mock

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type AppMockServer struct {
	http.Server
	port int32
}

func NewAppMockServer(port int32) *AppMockServer {
	router := mux.NewRouter()

	router.Path("/").HandlerFunc(Success)

	basicAuth := NewBasicAuthHandler()
	router.Path("/auth/basic/{username}/{password}").HandlerFunc(basicAuth.BasicAuth)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &AppMockServer{
		Server: server,
		port:   port,
	}
}

func (ams *AppMockServer) Start() {
	log.Infof("Starting test server on port: %d", ams.port)

	go func() {
		log.Info(ams.Server.ListenAndServe())
	}()
}

func (ams *AppMockServer) Kill() error {
	return ams.Server.Shutdown(context.Background())
}

func Success(w http.ResponseWriter, r *http.Request) {
	successResponse(w)
}

func successResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
