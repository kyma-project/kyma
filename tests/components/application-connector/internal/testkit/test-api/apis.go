package test_api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-http-utils/logger"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func SetupRoutes(logOut io.Writer, basicAuthCredentials BasicAuthCredentials, oAuthCredentials OAuthCredentials, expectedRequestParameters ExpectedRequestParameters, oauthTokens map[string]OAuthToken, csrfTokens CSRFTokens) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", alwaysOk).Methods("GET")
	api := router.PathPrefix("/v1/api").Subrouter()
	api.Use(Logger(logOut, logger.DevLoggerType))

	oauth := NewOAuth(oAuthCredentials.ClientID, oAuthCredentials.ClientSecret, oauthTokens)
	csrf := NewCSRF(csrfTokens)

	{
		api.HandleFunc("/oauth/token", oauth.Token).Methods(http.MethodPost)
		api.HandleFunc("/oauth/bad-token", oauth.BadToken).Methods(http.MethodPost)
		api.HandleFunc("/csrf/token", csrf.Token).Methods(http.MethodGet)
		api.HandleFunc("/csrf/bad-token", csrf.BadToken).Methods(http.MethodGet)
	}

	{
		r := api.PathPrefix("/unsecure").Subrouter()
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
		r.HandleFunc("/echo", echo).Methods(http.MethodPut, http.MethodPost, http.MethodDelete)
		r.HandleFunc("/code/{code:[0-9]+}", resCode).Methods(http.MethodGet)
		r.HandleFunc("/timeout", timeout).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/basic").Subrouter()
		r.Use(BasicAuth(basicAuthCredentials))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/oauth").Subrouter()
		r.Use(oauth.Middleware())
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/csrf-basic").Subrouter()
		r.Use(csrf.Middleware())
		r.Use(BasicAuth(basicAuthCredentials))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/csrf-oauth").Subrouter()
		r.Use(csrf.Middleware())
		r.Use(oauth.Middleware())
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/request-parameters-basic").Subrouter()
		r.Use(RequestParameters(expectedRequestParameters))
		r.Use(BasicAuth(basicAuthCredentials))
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/redirect").Subrouter()

		r.HandleFunc("/ok/target", alwaysOk).Methods(http.MethodGet)

		r.Handle("/ok", http.RedirectHandler("/v1/api/redirect/ok/target", http.StatusTemporaryRedirect))

		ba := BasicAuth(basicAuthCredentials)
		ok := http.HandlerFunc(alwaysOk)
		r.Handle("/basic/target", ba(ok)).Methods(http.MethodGet)
		r.Handle("/basic", http.RedirectHandler("/v1/api/redirect/basic/target", http.StatusTemporaryRedirect))

		r.Handle("/external", http.RedirectHandler("http://central-application-gateway.kyma-system:8081/v1/health", http.StatusTemporaryRedirect))
	}

	return router
}

func SetupMTLSRoutes(logOut io.Writer, oAuthCredentials OAuthCredentials, oauthTokens map[string]OAuthToken, csrfTokens CSRFTokens) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/v1/health", alwaysOk).Methods("GET")
	api := router.PathPrefix("/v1/api").Subrouter()
	api.Use(Logger(logOut, logger.DevLoggerType))

	oauth := NewOAuth(oAuthCredentials.ClientID, oAuthCredentials.ClientSecret, oauthTokens)
	csrf := NewCSRF(csrfTokens)

	{
		r := api.PathPrefix("/mtls").Subrouter()
		r.Use(oauth.Middleware())
		api.HandleFunc("/mtls-oauth/token", oauth.MTLSToken).Methods(http.MethodPost)
	}

	{
		r := api.PathPrefix("/mtls").Subrouter()
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}
	{
		r := api.PathPrefix("/csrf-mtls").Subrouter()
		r.Use(csrf.Middleware())
		r.HandleFunc("/ok", alwaysOk).Methods(http.MethodGet)
	}

	return router
}

func Logger(out io.Writer, t logger.Type) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return logger.Handler(next, out, t)
	}
}

func handleError(w http.ResponseWriter, code int, format string, a ...interface{}) {
	err := fmt.Errorf(format, a...)
	log.Error(err)
	w.WriteHeader(code)
}
