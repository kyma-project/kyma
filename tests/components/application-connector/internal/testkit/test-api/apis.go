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

	oauth := NewOAuth(oauthCred.ClientID, oauthCred.ClientSecret)
	csrf := NewCSRF()

	{
		api.HandleFunc("/oauth/token", oauth.Token).Methods(http.MethodPost)
		api.HandleFunc("/mtlsoauth/token", oauth.MTLSToken).Methods(http.MethodPost)
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
		r := api.PathPrefix("/oauth").Subrouter()
		r.Use(oauth.Middleware())
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}
	{
		r := api.PathPrefix("/mtlsoauth").Subrouter()
		r.Use(oauth.Middleware())
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo)
	}
	{
		csrfBasicRouter := api.PathPrefix("/csrf-basic").Subrouter()
		csrfBasicRouter.Use(csrf.Middleware())
		csrfBasicRouter.Use(BasicAuth(basicAuthCredentials.User, basicAuthCredentials.Password))
		csrfBasicRouter.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		csrfBasicRouter.HandleFunc("/echo", echo)

		csrfOAuthRouter := api.PathPrefix("/csrf-oauth").Subrouter()
		csrfOAuthRouter.Use(oauth.Middleware())
		csrfOAuthRouter.Use(csrf.Middleware())
		csrfOAuthRouter.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		csrfOAuthRouter.HandleFunc("/echo", echo)

		csrfMTLSOAuthRouter := api.PathPrefix("/csrf-mtlsoauth").Subrouter()
		csrfMTLSOAuthRouter.Use(oauth.Middleware())
		csrfMTLSOAuthRouter.Use(csrf.Middleware())
		csrfMTLSOAuthRouter.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		csrfMTLSOAuthRouter.HandleFunc("/echo", echo)
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
