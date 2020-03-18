package mock

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type AppMockServer struct {
	*http.Server
	port int32
}

func NewAppMockServer(port int32) *AppMockServer {
	router := mux.NewRouter()

	router.Path("/status/ok").HandlerFunc(StatusOK)

	basicAuth := NewBasicAuthHandler()
	router.Path("/auth/basic/{username}/{password}").HandlerFunc(basicAuth.BasicAuth)

	oAuth := NewOauthHandler()
	router.Path("/auth/oauth/token/{clientid}/{clientsecret}").HandlerFunc(oAuth.OAuthTokenHandler)
	router.Path("/auth/oauth/token/{clientid}/{clientsecret}/headers/{header}/{value}").HandlerFunc(oAuth.OAuthTokenHeadersHandler)
	router.Path("/auth/oauth/token/{clientid}/{clientsecret}/queryparams/{param}/{value}").HandlerFunc(oAuth.OAuthTokenQueryParamsHandler)
	router.Path("/target/auth/oauth").HandlerFunc(oAuth.OAuthTargetHandler)

	headers := NewHeadersHandler()
	router.Path("/headers/{header}/{value}").HandlerFunc(headers.HeadersHandler)

	queryParams := NewQueryParamsHandler()
	router.Path("/queryparams/{param}/{value}").HandlerFunc(queryParams.QueryParamsHandler)

	csrf := NewCsrfHandler()
	router.Path("/csrftoken").HandlerFunc(csrf.CsrfToken)
	router.Path("/target").HandlerFunc(csrf.Target)

	router.NotFoundHandler = NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(405, "Method not allowed.")

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return &AppMockServer{
		Server: &server,
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

func StatusOK(w http.ResponseWriter, r *http.Request) {
	successResponse(w)
	w.Write([]byte("Ok"))
}

func successResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
