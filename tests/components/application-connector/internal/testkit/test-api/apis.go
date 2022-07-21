package test_api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func SetupRoutes(logOut io.Writer, basicAuthCredentials BasicAuthCredentials, oAuthCredentials OAuthCredentials) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", alwaysOk).Methods("GET")
	api := router.PathPrefix("/v1/api").Subrouter()
	api.Use(Logger(logOut, logger.DevLoggerType))

	oauth := NewOAuth(oAuthCredentials.ClientID, oAuthCredentials.ClientSecret)
	csrf := NewCSRF()

	{
		api.HandleFunc("/oauth/token", oauth.Token).Methods(http.MethodPost)
		api.HandleFunc("/oauth/bad-token", oauth.BadToken).Methods(http.MethodPost)
		api.HandleFunc("/csrf/token", csrf.Token).Methods(http.MethodGet)
		api.HandleFunc("/csrf/bad-token", csrf.BadToken).Methods(http.MethodGet)
	}

	{
		r := api.PathPrefix("/unsecure").Subrouter()
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
		r.HandleFunc("/code/{code:[0-9]+}", resCode)
		r.HandleFunc("/timeout", timeout)
	}
	{
		r := api.PathPrefix("/basic").Subrouter()
		r.Use(BasicAuth(basicAuthCredentials.User, basicAuthCredentials.Password))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}
	{
		r := api.PathPrefix("/oauth").Subrouter()
		r.Use(oauth.Middleware())
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
	err := fmt.Errorf(format, a...)
	log.Error(err)
	w.WriteHeader(code)
}
