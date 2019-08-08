package mock

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type Path string

func (p Path) String() string {
	return string(p)
}

const (
	StatusOk   Path = "/v1/status/ok"
	BasicAuth  Path = "/v1/auth/basic"
	OAuthToken Path = "/v1/auth/oauth/token"
	OAuth      Path = "/v1/auth/oauth"
)

type AppMockServer struct {
	http.Server
	port int32
}

func NewAppMockServer(port int32) *AppMockServer {
	router := mux.NewRouter()

	router.Path(StatusOk.String()).HandlerFunc(StatusOK)

	basicAuth := NewBasicAuthHandler()
	router.Path(BasicAuth.String() + "/{username}/{password}").HandlerFunc(basicAuth.BasicAuth)

	oAuth := NewOauthHandler()
	router.Path(OAuthToken.String() + "/{clientid}/{clientsecret}").HandlerFunc(oAuth.OAuthTokenHandler)
	router.Path(OAuth.String() + "/{clientid}/{clientsecret}").HandlerFunc(oAuth.OAuthHandler)

	// TODO - the rest of stuff
	headers := NewHeadersHandler()
	router.Path("/headers/{header}/{value}").HandlerFunc(headers.HeadersHandler)

	queryParams := NewQueryParamsHandler()
	router.Path("/queryparams/{param}/{value}").HandlerFunc(queryParams.QueryParamsHandler)

	csrf := NewCsrfHandler()
	router.Path("/csrftoken").HandlerFunc(csrf.CsrfToken)
	router.Path("/target").HandlerFunc(csrf.Target)

	router.NotFoundHandler = NewErrorHandler(http.StatusNotFound, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = NewErrorHandler(http.StatusMethodNotAllowed, "Method not allowed.")

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

func StatusOK(w http.ResponseWriter, r *http.Request) {
	successResponse(w)
	w.Write([]byte("Ok"))
}

func successResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
