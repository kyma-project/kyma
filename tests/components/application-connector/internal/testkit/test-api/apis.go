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

func SetupRoutes(logOut io.Writer, basicAuthCredentials BasicAuthCredentials, oauthCred OAuthCredentials) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", alwaysOk).Methods("GET")
	api := router.PathPrefix("/v1/api").Subrouter()
	api.Use(Logger(logOut, logger.DevLoggerType))

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
		oauth := NewOAuth(oauthCred.ClientID, oauthCred.ClientSecret)
		api.HandleFunc("/mtlsoauth/token", oauth.MTLSToken).Methods(http.MethodPost)

		r := api.PathPrefix("/mtlsoauth").Subrouter()
		r.Use(oauth.Middleware())
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
