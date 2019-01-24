package externalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/errorhandler"
)

type Config struct {
	Middlewares      []mux.MiddlewareFunc
	TokenCreator     tokens.Creator
	APIUrlsGenerator APIUrlsGenerator
	ContextExtractor httpcontext.SerializerExtractor
	Host             string
	Subject          certificates.CSRSubject
}

type SignatureHandler interface {
	SignCSR(w http.ResponseWriter, r *http.Request)
}

type InfoHandler interface {
	GetCSRInfo(w http.ResponseWriter, r *http.Request)
}

const apiSpecPath = "connectorapi.yaml"

func NewHandler(appHandlerCfg, runtimeHandlerCfg Config, globalMiddlewares []mux.MiddlewareFunc) http.Handler {
	router := mux.NewRouter()

	httphelpers.WithMiddlewares(router, globalMiddlewares)

	// TODO - api spec
	router.Path("/v1").Handler(http.RedirectHandler("/v1/api.yaml", http.StatusMovedPermanently)).Methods(http.MethodGet)
	router.Path("/v1/api.yaml").Handler(NewStaticFileHandler(apiSpecPath)).Methods(http.MethodGet)

	applicationInfoHandler := NewCSRInfoHandler(appHandlerCfg.TokenCreator, appHandlerCfg.ContextExtractor, appHandlerCfg.APIUrlsGenerator, appHandlerCfg.Host, appHandlerCfg.Subject)

	applicationRouter := router.PathPrefix("/v1/applications").Subrouter()
	httphelpers.WithMiddlewares(applicationRouter, appHandlerCfg.Middlewares)
	applicationRouter.HandleFunc("/csr/info", applicationInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	//applicationRouter.HandleFunc("/{appName}/client-certs", sHandler.SignCSR).Methods(http.MethodPost)

	runtimeInfoHandler := NewCSRInfoHandler(runtimeHandlerCfg.TokenCreator, runtimeHandlerCfg.ContextExtractor, runtimeHandlerCfg.APIUrlsGenerator, runtimeHandlerCfg.Host, runtimeHandlerCfg.Subject)

	runtimesRouter := router.PathPrefix("/v1/runtimes").Subrouter()
	httphelpers.WithMiddlewares(runtimesRouter, runtimeHandlerCfg.Middlewares)
	applicationRouter.HandleFunc("/csr/info", runtimeInfoHandler.GetCSRInfo).Methods(http.MethodGet)
	//applicationRouter.HandleFunc("/{appName}/client-certs", sHandler.SignCSR).Methods(http.MethodPost)

	router.NotFoundHandler = errorhandler.NewErrorHandler(404, "Requested resource could not be found.")
	router.MethodNotAllowedHandler = errorhandler.NewErrorHandler(405, "Method not allowed.")

	return router
}
