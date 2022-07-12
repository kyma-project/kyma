package test_api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func SetupRoutes(logOut io.Writer, basicAuthCredentials BasicAuthCredentials) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", alwaysOk).Methods("GET")
	api := router.PathPrefix("/v1/api").Subrouter()
	api.Use(Logger(logOut, logger.DevLoggerType))

	csrf := NewCSRF()

	{
		api.HandleFunc("/csrf/token", csrf.Token).Methods(http.MethodGet)
		api.HandleFunc("/csrf/bad-token", csrf.BadToken).Methods(http.MethodGet)
	}

	{
		r := api.PathPrefix("/unsecure").Subrouter()
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}
	{
		r := api.PathPrefix("/basic").Subrouter()
		r.Use(BasicAuth(basicAuthCredentials.User, basicAuthCredentials.Password))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}
	{
		r := api.PathPrefix("/csrf-basic").Subrouter()
		r.Use(csrf.Middleware())
		r.Use(BasicAuth(basicAuthCredentials.User, basicAuthCredentials.Password))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}

	return router
}

type BasicAuthCredentials struct {
	User     string
	Password string
}

func handleError(w http.ResponseWriter, code int, format string, a ...interface{}) {
	err := errors.New(fmt.Sprintf(format, a...))
	log.Error(err)
	w.WriteHeader(code)
}
